package config

import (
	"k8s.io/client-go/tools/cache"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/util/workqueue"
	"time"
	"k8s.io/apimachinery/pkg/util/wait"
	"transwarp/release-config/pkg/apis/transwarp/v1beta1"
	"strings"
	"reflect"
	releaseModel "WarpCloud/walm/pkg/models/release"
	"WarpCloud/walm/pkg/kafka"
	"encoding/json"
	"WarpCloud/walm/pkg/k8s"
	"WarpCloud/walm/pkg/release/utils"
	k8sModel "WarpCloud/walm/pkg/models/k8s"
	"WarpCloud/walm/pkg/release"
	errorModel "WarpCloud/walm/pkg/models/error"
)

// 动态依赖管理核心需求：
// 1. 保存release的依赖关系， 当被依赖的release的输出配置改变时， 依赖者可以自动更新。
// 2. 保存release的输出配置， 当安装release时可以注入依赖的输出配置。
// 3. 保存release的输入配置， 可以实时上报release 输入配置和输出配置到配置中心， 输入配置和输出配置要保持一致性
// 4. 用户可以获取release依赖关系， 输出配置， 输入配置， 当前release状态， 依赖这个release更新的状态。

const (
	defaultWorkers                       = 1
	defaultReloadDependingReleaseWorkers = 10
	defaultKafkaWorkers                  = 2

	defaultRetryReloadDelayTimeSecond = 5
)

type ReleaseConfigController struct {
	workingQueue                       workqueue.DelayingInterface
	workers                            int
	reloadDependingReleaseWorkingQueue workqueue.DelayingInterface
	reloadDependingReleaseWorkers      int
	kafkaWorkingQueue                  workqueue.DelayingInterface
	kafkaWorkers                       int
	k8sCache                           k8s.Cache
	releaseUseCase                     release.UseCase
	kafka                              kafka.Kafka
	retryReloadDelayTimeSecond         int64
}

func NewReleaseConfigController(k8sCache k8s.Cache, releaseUseCase release.UseCase, kafka kafka.Kafka, retryReloadDelayTimeSecond int64) *ReleaseConfigController {
	controller := &ReleaseConfigController{
		workingQueue:                       workqueue.NewNamedDelayingQueue("release-config"),
		workers:                            defaultWorkers,
		reloadDependingReleaseWorkingQueue: workqueue.NewNamedDelayingQueue("reload-depending-release"),
		reloadDependingReleaseWorkers:      defaultReloadDependingReleaseWorkers,
		kafkaWorkingQueue:                  workqueue.NewNamedDelayingQueue("kafka"),
		kafkaWorkers:                       defaultKafkaWorkers,
		k8sCache:                           k8sCache,
		releaseUseCase:                     releaseUseCase,
		kafka:                              kafka,
		retryReloadDelayTimeSecond:         retryReloadDelayTimeSecond,
	}

	if controller.retryReloadDelayTimeSecond == 0 {
		controller.retryReloadDelayTimeSecond = defaultRetryReloadDelayTimeSecond
	}

	return controller
}

func (controller *ReleaseConfigController) Start(stopChan <-chan struct{}) {
	defer func() {
		logrus.Info("v2 release config controller stopped")
	}()
	logrus.Info("v2 release config controller started")

	defer controller.workingQueue.ShutDown()
	for i := 0; i < controller.workers; i++ {
		go wait.Until(controller.worker, time.Second, stopChan)
	}

	defer controller.kafkaWorkingQueue.ShutDown()
	for i := 0; i < controller.kafkaWorkers; i++ {
		go wait.Until(controller.kafkaWorker, time.Second, stopChan)
	}

	defer controller.reloadDependingReleaseWorkingQueue.ShutDown()
	for i := 0; i < controller.reloadDependingReleaseWorkers; i++ {
		go wait.Until(controller.reloadDependingReleaseWorker, time.Second, stopChan)
	}

	AddFunc := func(obj interface{}) {
		controller.enqueueReleaseConfig(obj)
		controller.enqueueKafka(obj)
	}
	UpdateFunc := func(old, cur interface{}) {
		oldReleaseConfig, ok := old.(*v1beta1.ReleaseConfig)
		if !ok {
			logrus.Error("old object is not release config")
			return
		}
		curReleaseConfig, ok := cur.(*v1beta1.ReleaseConfig)
		if !ok {
			logrus.Error("cur object is not release config")
			return
		}
		if needsEnqueueUpdatedReleaseConfig(oldReleaseConfig, curReleaseConfig) {
			controller.enqueueReleaseConfig(cur)
		}
		if !reflect.DeepEqual(oldReleaseConfig.Spec, curReleaseConfig.Spec) {
			controller.enqueueKafka(cur)
		}
	}
	DeleteFunc := func(obj interface{}) {
		controller.enqueueKafka(obj)
	}
	controller.k8sCache.AddReleaseConfigHandler(AddFunc, UpdateFunc, DeleteFunc)

	<-stopChan
}

func (controller *ReleaseConfigController) enqueueReleaseConfig(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		logrus.Errorf("Couldn't get key for object %#v: %v", obj, err)
		return
	}
	controller.workingQueue.Add(key)
}

func (controller *ReleaseConfigController) enqueueKafka(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		logrus.Errorf("Couldn't get key for object %#v: %v", obj, err)
		return
	}
	controller.kafkaWorkingQueue.Add(key)
}

// worker runs a worker thread that just dequeues items, processes them, and marks them done.
// It enforces that the syncHandler is never invoked concurrently with the same key.
func (controller *ReleaseConfigController) worker() {
	for {
		func() {
			key, quit := controller.workingQueue.Get()
			if quit {
				return
			}
			defer controller.workingQueue.Done(key)
			err := controller.syncReleaseConfig(key.(string))
			if err != nil {
				logrus.Errorf("Error syncing release config: %v", err)
			}
		}()
	}
}

func (controller *ReleaseConfigController) kafkaWorker() {
	for {
		func() {
			key, quit := controller.kafkaWorkingQueue.Get()
			if quit {
				return
			}
			defer controller.kafkaWorkingQueue.Done(key)
			err := controller.publishToKafka(key.(string))
			if err != nil {
				logrus.Errorf("failed to publish release config of %s to kafka: %s", key.(string), err.Error())
			}
		}()
	}
}

func (controller *ReleaseConfigController) publishToKafka(releaseKey string) error {
	logrus.Infof("start to publish release config of %s to kafka", releaseKey)
	namespace, name, err := cache.SplitMetaNamespaceKey(releaseKey)
	if err != nil {
		return err
	}

	event := releaseModel.ReleaseConfigDeltaEvent{}

	resource, err := controller.k8sCache.GetResource(k8sModel.ReleaseConfigKind, namespace, name)
	if err != nil {
		if errorModel.IsNotFoundError(err) {
			event.Type = releaseModel.Delete
			event.Data = &k8sModel.ReleaseConfig{
				Meta: k8sModel.Meta{
					Namespace: namespace,
					Name:      name,
				},
			}
		} else {
			logrus.Errorf("failed to get release config of %s", releaseKey)
			return err
		}
	} else {
		_, err = controller.releaseUseCase.GetRelease(namespace, name)
		if err != nil {
			if errorModel.IsNotFoundError(err) {
				logrus.Warnf("release %s is not found， ignore to publish release config to kafka", releaseKey)
				return nil
			}
			logrus.Errorf("failed to get release %s : %s", releaseKey, err.Error())
			return err
		}
		event.Type = releaseModel.CreateOrUpdate
		event.Data = resource.(*k8sModel.ReleaseConfig)
	}

	eventMsg, err := json.Marshal(event)
	if err != nil {
		logrus.Errorf("failed to marshal event : %s", err.Error())
		return err
	}

	err = controller.kafka.SyncSendMessage(kafka.ReleaseConfigTopic, string(eventMsg))
	if err != nil {
		logrus.Errorf("failed to send release config event of %s to kafka : %s", releaseKey, err.Error())
		return err
	}

	return nil
}

// worker runs a worker thread that just dequeues items, processes them, and marks them done.
// It enforces that the syncHandler is never invoked concurrently with the same key.
func (controller *ReleaseConfigController) reloadDependingReleaseWorker() {
	for {
		func() {
			key, quit := controller.reloadDependingReleaseWorkingQueue.Get()
			if quit {
				return
			}
			defer controller.reloadDependingReleaseWorkingQueue.Done(key)
			err := controller.reloadDependingRelease(key.(string))
			if err != nil {
				if strings.Contains(err.Error(), release.WaitReleaseTaskMsgPrefix) {
					logrus.Warnf("depending release %s would be reloaded after %d second", key.(string), controller.retryReloadDelayTimeSecond)
					controller.reloadDependingReleaseWorkingQueue.AddAfter(key, time.Second* time.Duration(controller.retryReloadDelayTimeSecond))
				} else {
					logrus.Errorf("Error reload depending release %s: %v", key.(string), err)
				}
			}
		}()
	}
}

func needsEnqueueUpdatedReleaseConfig(old *v1beta1.ReleaseConfig, cur *v1beta1.ReleaseConfig) bool {
	if utils.ConfigValuesDiff(old.Spec.OutputConfig, cur.Spec.OutputConfig) {
		return true
	}
	return false
}

func (controller *ReleaseConfigController) reloadDependingRelease(releaseKey string) error {
	logrus.Infof("start to reload release %s", releaseKey)
	namespace, name, err := cache.SplitMetaNamespaceKey(releaseKey)
	if err != nil {
		return err
	}
	err = controller.releaseUseCase.ReloadRelease(namespace, name)
	if err != nil {
		logrus.Errorf("failed to reload release %s/%s : %s", namespace, name, err.Error())
		return err
	}
	return nil
}

// 两级work queue设计初衷：利用work queue压缩相同key的功能， 尽可能地减少reload release的次数
// a 有多个依赖 b, c, d...， 当b, c, d... 同时更新了，a最好的情况是只更新一次
func (controller *ReleaseConfigController) syncReleaseConfig(releaseConfigKey string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(releaseConfigKey)
	if err != nil {
		return err
	}

	releaseConfigs, err := controller.k8sCache.ListReleaseConfigs("", "")
	if err != nil {
		logrus.Errorf("failed to list all release configs : %s", err.Error())
		return err
	}
	for _, releaseConfig := range releaseConfigs {
		for _, dependedRelease := range releaseConfig.Dependencies {
			dependedReleaseNamespace, dependedReleaseName, err := utils.ParseDependedRelease(releaseConfig.Namespace, dependedRelease)
			if err != nil {
				continue
			}
			if dependedReleaseNamespace == namespace && dependedReleaseName == name {
				rc := &v1beta1.ReleaseConfig{}
				rc.Namespace = releaseConfig.Namespace
				rc.Name = releaseConfig.Name
				controller.enqueueDependingRelease(rc)
				break
			}
		}
	}

	return nil
}

func (controller *ReleaseConfigController) enqueueDependingRelease(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		logrus.Errorf("Couldn't get key for object %#v: %v", obj, err)
		return
	}
	controller.reloadDependingReleaseWorkingQueue.Add(key)
}
