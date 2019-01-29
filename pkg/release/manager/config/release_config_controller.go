package config

import (
	"k8s.io/client-go/tools/cache"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/util/workqueue"
	"time"
	"k8s.io/apimachinery/pkg/util/wait"
	"transwarp/release-config/pkg/apis/transwarp/v1beta1"
	"walm/pkg/k8s/handler"
	"strings"
	"walm/pkg/k8s/informer"
	"walm/pkg/release/manager/helm"
)

// 动态依赖管理核心需求：
// 1. 保存release的依赖关系， 当被依赖的release的输出配置改变时， 依赖者可以自动更新。
// 2. 保存release的输出配置， 当安装release时可以注入依赖的输出配置。
// 3. 保存release的输入配置， 可以实时上报release 输入配置和输出配置到配置中心， 输入配置和输出配置要保持一致性
// 4. 用户可以获取release依赖关系， 输出配置， 输入配置， 当前release状态， 依赖这个release更新的状态。

const (
	defaultWorkers                       = 1
	defaultReloadDependingReleaseWorkers = 10
)

type ReleaseConfigController struct {
	handlerFuncs                       *cache.ResourceEventHandlerFuncs
	workingQueue                       workqueue.DelayingInterface
	workers                            int
	reloadDependingReleaseWorkingQueue workqueue.DelayingInterface
	reloadDependingReleaseWorkers      int
	started                            bool
	releaseConfigHandler               *handler.ReleaseConfigHandler
}

func NewReleaseConfigController() *ReleaseConfigController {
	controller := &ReleaseConfigController{
		workingQueue:                       workqueue.NewNamedDelayingQueue("release-config"),
		workers:                            defaultWorkers,
		reloadDependingReleaseWorkingQueue: workqueue.NewNamedDelayingQueue("reload-depending-release"),
		reloadDependingReleaseWorkers:      defaultReloadDependingReleaseWorkers,
		releaseConfigHandler:               handler.GetDefaultHandlerSet().GetReleaseConfigHandler(),
	}

	return controller
}

func (controller *ReleaseConfigController) Start(stopChan <-chan struct{}) {
	defer func() {
		controller.started = false
		logrus.Info("v2 release config controller stopped")
	}()
	logrus.Info("v2 release config controller started")
	controller.started = true

	defer controller.workingQueue.ShutDown()
	for i := 0; i < controller.workers; i++ {
		go wait.Until(controller.worker, time.Second, stopChan)
	}

	defer controller.reloadDependingReleaseWorkingQueue.ShutDown()
	for i := 0; i < controller.reloadDependingReleaseWorkers; i++ {
		go wait.Until(controller.reloadDependingReleaseWorker, time.Second, stopChan)
	}

	if controller.handlerFuncs == nil {
		controller.handlerFuncs = &cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if !controller.started {
					return
				}
				controller.enqueueReleaseConfig(obj)
			},
			UpdateFunc: func(old, cur interface{}) {
				if !controller.started {
					return
				}
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
				if controller.needsUpdate(oldReleaseConfig, curReleaseConfig) {
					controller.enqueueReleaseConfig(cur)
				}
			},
			DeleteFunc: func(obj interface{}) {
				//if !controller.started {
				//	return
				//}
				//controller.enqueueReleaseConfig(obj)
			},
		}
		informer.GetDefaultFactory().ReleaseConifgFactory.Transwarp().V1beta1().ReleaseConfigs().Informer().AddEventHandler(controller.handlerFuncs)
	}

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
				if strings.Contains(err.Error(), "please wait for the release latest task") {
					logrus.Warnf("depending release %s would be reloaded after 5 second", key.(string))
					controller.reloadDependingReleaseWorkingQueue.AddAfter(key, time.Second * 5)
				} else {
					logrus.Errorf("Error reload depending release %s: %v", key.(string), err)
				}
			}
		}()
	}
}

func (controller *ReleaseConfigController) needsUpdate(old *v1beta1.ReleaseConfig, cur *v1beta1.ReleaseConfig) bool {
	if helm.ConfigValuesDiff(old.Spec.OutputConfig, cur.Spec.OutputConfig) {
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
	err = helm.GetDefaultHelmClient().ReloadRelease(namespace, name, false)
	if err != nil {
		logrus.Errorf("failed to reload release %s/%s : %s", namespace, name, err.Error())
		return err
	}
	return nil
}

//TODO retry?
func (controller *ReleaseConfigController) syncReleaseConfig(releaseConfigKey string) error {
	//TODO 上报配置中心 : get latest release config using release config lister
	namespace, name, err := cache.SplitMetaNamespaceKey(releaseConfigKey)
	if err != nil {
		return err
	}

	releaseConfigs, err := controller.releaseConfigHandler.ListReleaseConfigs("", nil)
	if err != nil {
		logrus.Errorf("failed to list all release configs : %s", err.Error())
		return err
	}
	for _, releaseConfig := range releaseConfigs {
		for _, dependedRelease := range releaseConfig.Spec.Dependencies {
			dependedReleaseNamespace, dependedReleaseName := parseDependedRelease(releaseConfig.Namespace, dependedRelease)
			if dependedReleaseNamespace == namespace && dependedReleaseName == name {
				controller.enqueueDependingRelease(releaseConfig)
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

func parseDependedRelease(dependingReleaseNamespace, dependedRelease string) (namespace, name string) {
	ss := strings.Split(dependedRelease, ".")
	if len(ss) == 2 {
		namespace = ss[0]
		name = ss[1]
	} else if len(ss) == 1 {
		namespace = dependingReleaseNamespace
		name = ss[0]
	} else {
		logrus.Errorf("depended release %s is not valid: only 1 or 0 \".\" is allowed", dependedRelease)
	}
	return
}
