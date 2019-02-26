package helm

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"time"
	"walm/pkg/util"
	"walm/pkg/util/transwarpjsonnet"

	"github.com/hashicorp/golang-lru"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/helm/pkg/chart"
	"k8s.io/helm/pkg/chart/loader"
	hapirelease "k8s.io/helm/pkg/hapi/release"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/storage/driver"
	"walm/pkg/hook"
	"walm/pkg/k8s/adaptor"
	"walm/pkg/k8s/client"
	"walm/pkg/k8s/handler"
	"walm/pkg/redis"
	"walm/pkg/release"
	"walm/pkg/release/manager/helm/cache"
	"walm/pkg/setting"
	"walm/pkg/task"
	walmerr "walm/pkg/util/error"

	"k8s.io/helm/pkg/walm"
	"k8s.io/helm/pkg/walm/plugins"
)

const (
	defaultTimeoutSec              int64         = 60 * 5
	defaultSleepTimeSecond         time.Duration = 1 * time.Second
	helmCacheDefaultResyncInterval time.Duration = 5 * time.Minute
)

type ChartRepository struct {
	Name     string
	URL      string
	Username string
	Password string
}

type HelmClient struct {
	chartRepoMap            map[string]*ChartRepository
	dryRun                  bool
	helmCache               *cache.HelmCache
	helmCacheResyncInterval time.Duration
	releaseConfigHandler    *handler.ReleaseConfigHandler
	helmClients             *lru.Cache
}

var helmClient *HelmClient

func GetDefaultHelmClient() *HelmClient {
	if helmClient == nil {
		chartRepoMap := make(map[string]*ChartRepository)

		for _, chartRepo := range setting.Config.RepoList {
			chartRepository := ChartRepository{
				Name:     chartRepo.Name,
				URL:      chartRepo.URL,
				Username: "",
				Password: "",
			}
			chartRepoMap[chartRepo.Name] = &chartRepository
		}

		helmCache := cache.NewHelmCache(redis.GetDefaultRedisClient())

		helmClients, _ := lru.New(100)

		helmClient = &HelmClient{
			chartRepoMap:            chartRepoMap,
			dryRun:                  false,
			helmCache:               helmCache,
			helmCacheResyncInterval: helmCacheDefaultResyncInterval,
			releaseConfigHandler:    handler.GetDefaultHandlerSet().GetReleaseConfigHandler(),
			helmClients:             helmClients,
		}
	}
	return helmClient
}

func (hc *HelmClient) GetHelmCache() *cache.HelmCache {
	return hc.helmCache
}

func (hc *HelmClient) GetDependencies(repoName, chartName, chartVersion string) (subChartNames []string, err error) {
	logrus.Debugf("Enter GetDependencies %s %s\n", chartName, chartVersion)

	subChartNames = []string{}
	detailChartInfo, err := GetDetailChartInfo(repoName, chartName, chartVersion)
	if err != nil {
		return nil, err
	}
	if detailChartInfo.Metainfo != nil && detailChartInfo.Metainfo.ChartDependenciesInfo != nil {
		for _, dependency := range detailChartInfo.Metainfo.ChartDependenciesInfo {
			subChartNames = append(subChartNames, dependency.Name)
		}
	}

	return subChartNames, nil
}

func (hc *HelmClient) GetCurrentHelmClient(namespace string) (*helm.Client, error) {
	if c, ok := hc.helmClients.Get(namespace); ok {
		return c.(*helm.Client), nil
	} else {
		kc := client.GetKubeClient(namespace)
		clientset, err := kc.KubernetesClientSet()
		if err != nil {
			return nil, err
		}

		d := driver.NewConfigMaps(clientset.CoreV1().ConfigMaps(namespace))
		c = helm.NewClient(
			helm.KubeClient(kc),
			helm.Driver(d),
			helm.Discovery(clientset.Discovery()),
		)
		hc.helmClients.Add(namespace, c)
		return c.(*helm.Client), nil
	}
}

func (hc *HelmClient) downloadChart(repoName, charName, version string) (string, error) {
	if repoName == "" {
		repoName = "stable"
	}
	repo, ok := hc.chartRepoMap[repoName]
	if !ok {
		return "", fmt.Errorf("can not find repo name: %s", repoName)
	}
	chartURL, httpGetter, err := FindChartInChartMuseumRepoURL(repo.URL, "", "", charName, version)
	if err != nil {
		return "", err
	}

	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}
	filename, err := ChartMuseumDownloadTo(chartURL, tmpDir, httpGetter)
	if err != nil {
		logrus.Printf("DownloadTo err %v", err)
		return "", err
	}

	return filename, nil
}

func (hc *HelmClient) LoadChart(repoName, chartName, chartVersion string) (rawChart *chart.Chart, err error) {
	chartPath, err := hc.downloadChart(repoName, chartName, chartVersion)
	if err != nil {
		logrus.Errorf("failed to download chart : %s", err.Error())
		return nil, err
	}

	chartLoader, err := loader.Loader(chartPath)
	if err != nil {
		logrus.Errorf("failed to init chartLoader : %s", err.Error())
		return nil, err
	}

	return chartLoader.Load()
}

func (hc *HelmClient) StartResyncReleaseCaches(stopCh <-chan struct{}) {
	logrus.Infof("start to resync release cache every %v", hc.helmCacheResyncInterval)
	// first time should be sync
	hc.helmCache.Resync()
	firstTime := true
	go wait.Until(func() {
		if firstTime {
			time.Sleep(hc.helmCacheResyncInterval)
			firstTime = false
		}
		hc.helmCache.Resync()
	}, hc.helmCacheResyncInterval, stopCh)
}

func (hc *HelmClient) GetRelease(namespace, name string) (releaseV2 *release.ReleaseInfoV2, err error) {
	releaseTask, err := hc.helmCache.GetReleaseTask(namespace, name)
	if err != nil {
		return nil, err
	}

	releaseV2, err = hc.buildReleaseInfoV2ByReleaseTask(releaseTask, nil)
	if err != nil {
		return nil, err
	}

	return
}

// 当task sig为空 或者task state已经ttl：build release by cache
// 当task state没有ttl:
// 1. 当task没完成：返回message
// 2. 当task完成：
//       a. 成功：build release by cache
//       b. 失败：返回message
func (hc *HelmClient) buildReleaseInfoV2ByReleaseTask(releaseTask *cache.ReleaseTask, releaseCache *release.ReleaseCache) (releaseV2 *release.ReleaseInfoV2, err error) {
	releaseV2 = &release.ReleaseInfoV2{
		ReleaseInfo: release.ReleaseInfo{
			ReleaseSpec: release.ReleaseSpec{
				Namespace: releaseTask.Namespace,
				Name:      releaseTask.Name,
			},
		},
	}
	if releaseTask.LatestReleaseTaskSig != nil {
		taskState := releaseTask.LatestReleaseTaskSig.GetTaskState()
		if taskState != nil && taskState.TaskName != "" {
			if taskState.IsSuccess() {
				if taskState.TaskName == deleteReleaseTaskName {
					return nil, walmerr.NotFoundError{}
				}
			} else if taskState.IsFailure() {
				releaseV2.Message = fmt.Sprintf("the release latest task %s-%s failed : %s", releaseTask.LatestReleaseTaskSig.Name, releaseTask.LatestReleaseTaskSig.UUID, taskState.Error)
				return
			} else {
				releaseV2.Message = fmt.Sprintf("please wait for the release latest task %s-%s finished", releaseTask.LatestReleaseTaskSig.Name, releaseTask.LatestReleaseTaskSig.UUID)
				return
			}
		}
	}

	if releaseCache == nil {
		releaseCache, err = hc.helmCache.GetReleaseCache(releaseTask.Namespace, releaseTask.Name)
		if err != nil {
			logrus.Errorf("failed to get release cache of %s/%s : %s", releaseTask.Namespace, releaseTask.Name, err.Error())
			return
		}
	}

	releaseV2, err = hc.buildReleaseInfoV2(releaseCache)
	if err != nil {
		logrus.Errorf("failed to build v2 release info : %s", err.Error())
		return nil, err
	}
	return
}

func (hc *HelmClient) buildReleaseInfoV2(releaseCache *release.ReleaseCache) (*release.ReleaseInfoV2, error) {
	releaseV1, err := buildReleaseInfo(releaseCache)
	if err != nil {
		logrus.Errorf("failed to build release info: %s", err.Error())
		return nil, err
	}
	releaseV2 := &release.ReleaseInfoV2{ReleaseInfo: *releaseV1}
	releaseConfig, err := hc.releaseConfigHandler.GetReleaseConfig(releaseCache.Namespace, releaseCache.Name)
	if err != nil {
		if adaptor.IsNotFoundErr(err) {
			releaseV2.DependenciesConfigValues = map[string]interface{}{}
			releaseV2.OutputConfigValues = map[string]interface{}{}
			releaseV2.ReleaseLabels = map[string]string{}
		} else {
			logrus.Errorf("failed to get release config : %s", err.Error())
			return nil, err
		}
	} else {
		releaseV2.ConfigValues = releaseConfig.Spec.ConfigValues
		releaseV2.Dependencies = releaseConfig.Spec.Dependencies
		releaseV2.DependenciesConfigValues = releaseConfig.Spec.DependenciesConfigValues
		releaseV2.OutputConfigValues = releaseConfig.Spec.OutputConfig
		releaseV2.ReleaseLabels = releaseConfig.Labels
	}
	releaseV2.ComputedValues = releaseCache.ComputedValues
	return releaseV2, nil
}

func (hc *HelmClient) ListReleases(namespace, filter string) ([]*release.ReleaseInfoV2, error) {
	logrus.Debugf("Enter ListReleases namespace=%s filter=%s\n", namespace, filter)
	releaseTasks, err := hc.helmCache.GetReleaseTasks(namespace, filter, 0)
	if err != nil {
		logrus.Errorf("failed to get release tasks with namespace=%s filter=%s : %s", namespace, filter, err.Error())
		return nil, err
	}

	releaseCaches, err := hc.helmCache.GetReleaseCaches(namespace, filter, 0)
	if err != nil {
		logrus.Errorf("failed to get release caches with namespace=%s filter=%s : %s", namespace, filter, err.Error())
		return nil, err
	}

	return hc.doListReleases(releaseTasks, releaseCaches)
}

func (hc *HelmClient) ListReleasesByLabels(namespace string, labelSelector *v1.LabelSelector) ([]*release.ReleaseInfoV2, error) {
	releaseConfigs, err := hc.releaseConfigHandler.ListReleaseConfigs(namespace, labelSelector)
	if err != nil {
		logrus.Errorf("failed to list release configs : %s", err.Error())
		return nil, err
	}

	releaseNames := []cache.ReleaseFieldName{}
	for _, releaseConfig := range releaseConfigs {
		releaseNames = append(releaseNames, cache.ReleaseFieldName{
			Namespace: releaseConfig.Namespace,
			Name:      releaseConfig.Name,
		})
	}

	releases, err := hc.listReleasesByNames(releaseNames)
	if err != nil {
		return nil, err
	}
	return releases, nil
}

func (hc *HelmClient) listReleasesByNames(names []cache.ReleaseFieldName) ([]*release.ReleaseInfoV2, error) {
	if len(names) == 0 {
		return []*release.ReleaseInfoV2{}, nil
	}
	releaseTasks, err := hc.helmCache.GetReleaseTasksByNames(names)
	if err != nil {
		logrus.Errorf("failed to get release tasks : %s", err.Error())
		return nil, err
	}

	releaseCaches, err := hc.helmCache.GetReleaseCachesByNames(names)
	if err != nil {
		logrus.Errorf("failed to get release caches : %s", err.Error())
		return nil, err
	}

	return hc.doListReleases(releaseTasks, releaseCaches)
}

func (hc *HelmClient) doListReleases(releaseTasks []*cache.ReleaseTask, releaseCaches []*release.ReleaseCache) (releaseInfos []*release.ReleaseInfoV2, err error) {
	releaseCacheMap := map[string]*release.ReleaseCache{}
	for _, releaseCache := range releaseCaches {
		releaseCacheMap[releaseCache.Namespace+"/"+releaseCache.Name] = releaseCache
	}

	releaseInfos = []*release.ReleaseInfoV2{}
	//TODO 限制协程的数量
	mux := &sync.Mutex{}
	var wg sync.WaitGroup
	for _, releaseTask := range releaseTasks {
		wg.Add(1)
		go func(releaseTask *cache.ReleaseTask, releaseCache *release.ReleaseCache) {
			defer wg.Done()
			info, err1 := hc.buildReleaseInfoV2ByReleaseTask(releaseTask, releaseCache)
			if err1 != nil {
				if walmerr.IsNotFoundError(err1) {
					return
				}
				err = errors.New(fmt.Sprintf("failed to build release info: %s", err1.Error()))
				logrus.Error(err.Error())
				return
			}
			mux.Lock()
			releaseInfos = append(releaseInfos, info)
			mux.Unlock()
		}(releaseTask, releaseCacheMap[releaseTask.Namespace+"/"+releaseTask.Name])
	}
	wg.Wait()
	if err != nil {
		return
	}
	return
}

// reload dependencies config values, if changes, upgrade release
func (hc *HelmClient) ReloadRelease(namespace, name string, isSystem bool) error {
	//TODO need to opt: task or cache?
	_, err := hc.helmCache.GetReleaseCache(namespace, name)
	if err != nil {
		logrus.Errorf("failed to get release cache of %s/%s : %s", namespace, name, err.Error())
		return err
	}

	releaseConfig, err := handler.GetDefaultHandlerSet().GetReleaseConfigHandler().GetReleaseConfig(namespace, name)
	if err != nil {
		logrus.Errorf("failed to get release config of %s/%s : %s", namespace, name, err.Error())
		return err
	}

	oldDependenciesConfigValues := releaseConfig.Spec.DependenciesConfigValues
	newDependenciesConfigValues, err := hc.getDependencyOutputConfigs(namespace, releaseConfig.Spec.Dependencies)
	if err != nil {
		logrus.Errorf("failed to get dependencies output configs of %s/%s : %s", namespace, name, err.Error())
		return err
	}

	if ConfigValuesDiff(oldDependenciesConfigValues, newDependenciesConfigValues) {
		releaseRequest := &release.ReleaseRequestV2{
			ReleaseRequest: release.ReleaseRequest{
				Name:         name,
				ChartVersion: releaseConfig.Spec.ChartVersion,
				ChartName:    releaseConfig.Spec.ChartName,
				Dependencies: releaseConfig.Spec.Dependencies,
				ConfigValues: releaseConfig.Spec.ConfigValues,
			},
			ReleaseLabels: releaseConfig.Labels,
		}
		err = hc.InstallUpgradeRelease(namespace, releaseRequest, isSystem, nil, false, 0)
		if err != nil {
			logrus.Errorf("failed to upgrade release v2 %s/%s : %s", namespace, name, err.Error())
			return err
		}
		logrus.Infof("succeed to reload release %s/%s", namespace, name)
	} else {
		logrus.Infof("ignore reloading release %s/%s : dependencies config value does not change", namespace, name)
	}

	return nil
}

func (hc *HelmClient) validateReleaseTask(namespace, name string, allowReleaseTaskNotExist bool) (releaseTask *cache.ReleaseTask, err error) {
	releaseTask, err = hc.helmCache.GetReleaseTask(namespace, name)
	if err != nil {
		if !walmerr.IsNotFoundError(err) {
			logrus.Errorf("failed to get release task : %s", err.Error())
			return
		} else if !allowReleaseTaskNotExist {
			return
		} else {
			err = nil
		}
	} else {
		if releaseTask.LatestReleaseTaskSig != nil && !releaseTask.LatestReleaseTaskSig.IsTaskFinishedOrTimeout() {
			err = fmt.Errorf("please wait for the release latest task %s-%s finished or timeout", releaseTask.LatestReleaseTaskSig.Name, releaseTask.LatestReleaseTaskSig.UUID)
			logrus.Warn(err.Error())
			return
		}
	}
	return
}

func (hc *HelmClient) DeleteReleaseWithRetry(namespace, releaseName string, isSystem bool, deletePvcs bool, async bool, timeoutSec int64) error {
	retryTimes := 5
	for {
		err := hc.DeleteRelease(namespace, releaseName, isSystem, deletePvcs, async, timeoutSec)
		if err != nil {
			if strings.Contains(err.Error(), "please wait for the release latest task") && retryTimes > 0 {
				logrus.Warnf("retry to delete release %s/%s after 2 second", namespace, releaseName)
				retryTimes --
				time.Sleep(time.Second * 2)
				continue
			}
		}
		return err
	}
}

func (hc *HelmClient) DeleteRelease(namespace, releaseName string, isSystem bool, deletePvcs bool, async bool, timeoutSec int64) error {
	if timeoutSec == 0 {
		timeoutSec = defaultTimeoutSec
	}

	oldReleaseTask, err := hc.validateReleaseTask(namespace, releaseName, false)
	if err != nil {
		if walmerr.IsNotFoundError(err) {
			logrus.Warnf("release task %s/%s is not found", namespace, releaseName)
			return nil
		}
		logrus.Errorf("failed to validate release task : %s", err.Error())
		return err
	}

	releaseTaskArgs := &DeleteReleaseTaskArgs{
		Namespace:   namespace,
		ReleaseName: releaseName,
		IsSystem:    isSystem,
		DeletePvcs:  deletePvcs,
	}
	taskSig, err := SendReleaseTask(releaseTaskArgs)
	if err != nil {
		logrus.Errorf("failed to send %s : %s", releaseTaskArgs.GetTaskName(), err.Error())
		return err
	}
	taskSig.TimeoutSec = timeoutSec

	releaseTask := &cache.ReleaseTask{
		Namespace:            namespace,
		Name:                 releaseName,
		LatestReleaseTaskSig: taskSig,
	}

	err = hc.helmCache.CreateOrUpdateReleaseTask(releaseTask)
	if err != nil {
		logrus.Errorf("failed to set release task of %s/%s to redis: %s", namespace, releaseName, err.Error())
		return err
	}

	if oldReleaseTask != nil && oldReleaseTask.LatestReleaseTaskSig != nil {
		err = task.GetDefaultTaskManager().PurgeTaskState(oldReleaseTask.LatestReleaseTaskSig.GetTaskSignature())
		if err != nil {
			logrus.Warnf("failed to purge task state : %s", err.Error())
		}
	}

	if !async {
		asyncResult := taskSig.GetAsyncResult()
		_, err = asyncResult.GetWithTimeout(time.Duration(timeoutSec)*time.Second, defaultSleepTimeSecond)
		if err != nil {
			logrus.Errorf("failed to delete release  %s/%s: %s", namespace, releaseName, err.Error())
			return err
		}
	}
	logrus.Infof("succeed to call delete release %s/%s api", namespace, releaseName)
	return nil
}

func (hc *HelmClient) doDeleteRelease(namespace, releaseName string, isSystem bool, deletePvcs bool) error {
	currentHelmClient, err := hc.GetCurrentHelmClient(namespace)
	if err != nil {
		logrus.Errorf("failed to get current helm client : %s", err.Error())
		return err
	}

	releaseCache, err := hc.helmCache.GetReleaseCache(namespace, releaseName)
	if err != nil {
		if walmerr.IsNotFoundError(err) {
			logrus.Warnf("release cache %s is not found in redis", releaseName)
			return nil
		}
		logrus.Errorf("failed to get release cache %s : %s", releaseName, err.Error())
		return err
	}
	releaseInfo, err := hc.buildReleaseInfoV2(releaseCache)
	if err != nil {
		logrus.Errorf("failed to build release info : %s", err.Error())
		return err
	}

	opts := []helm.UninstallOption{
		helm.UninstallPurge(true),
	}
	res, err := currentHelmClient.UninstallRelease(
		releaseName, opts...,
	)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logrus.Warnf("release %s is not found in tiller", releaseName)
		} else {
			logrus.Errorf("failed to delete release : %s", err.Error())
			return err
		}
	}
	if res != nil && res.Info != "" {
		logrus.Println(res.Info)
	}

	err = hc.helmCache.DeleteReleaseCache(namespace, releaseName)
	if err != nil {
		logrus.Errorf("failed to delete release cache of %s : %s", releaseName, err.Error())
		return err
	}

	if deletePvcs {
		statefulSets := []adaptor.WalmStatefulSet{}
		if len(releaseInfo.Status.StatefulSets) > 0 {
			statefulSets = append(statefulSets, releaseInfo.Status.StatefulSets...)
		}

		for _, instance := range releaseInfo.Status.Instances {
			if instance.Modules != nil && len(instance.Modules.StatefulSets) > 0 {
				statefulSets = append(statefulSets, instance.Modules.StatefulSets...)
			}
		}

		for _, statefulSet := range statefulSets {
			if statefulSet.Selector != nil && (len(statefulSet.Selector.MatchLabels) > 0 || len(statefulSet.Selector.MatchExpressions) > 0) {
				pvcs, err := handler.GetDefaultHandlerSet().GetPersistentVolumeClaimHandler().ListPersistentVolumeClaims(statefulSet.Namespace, statefulSet.Selector)
				if err != nil {
					logrus.Errorf("failed to list pvcs ralated to stateful set %s/%s : %s", statefulSet.Namespace, statefulSet.Name, err.Error())
					return err
				}

				for _, pvc := range pvcs {
					err = handler.GetDefaultHandlerSet().GetPersistentVolumeClaimHandler().DeletePersistentVolumeClaim(pvc.Namespace, pvc.Name)
					if err != nil {
						if adaptor.IsNotFoundErr(err) {
							logrus.Warnf("pvc %s/%s related to stateful set %s/%s is not found", pvc.Namespace, pvc.Name, statefulSet.Namespace, statefulSet.Name)
							continue
						}
						logrus.Errorf("failed to delete pvc %s/%s related to stateful set %s/%s : %s", pvc.Namespace, pvc.Name, statefulSet.Namespace, statefulSet.Name, err.Error())
						return err
					}
					logrus.Infof("succeed to delete pvc %s/%s related to stateful set %s/%s", pvc.Namespace, pvc.Name, statefulSet.Namespace, statefulSet.Name)
				}
			}
		}
	}

	logrus.Infof("succeed to delete release %s/%s", namespace, releaseName)
	return nil
}

func (hc *HelmClient) RestartRelease(namespace, releaseName string) error {
	logrus.Debugf("Enter RestartRelease %s %s\n", namespace, releaseName)
	releaseInfo, err := hc.GetRelease(namespace, releaseName)
	if err != nil {
		logrus.Errorf("failed to get release info : %s", err.Error())
		return err
	}

	podsToRestart := releaseInfo.Status.GetPodsNeedRestart()
	podsRestartFailed := []string{}
	mux := &sync.Mutex{}
	var wg sync.WaitGroup
	for _, podToRestart := range podsToRestart {
		wg.Add(1)
		go func(podToRestart *adaptor.WalmPod) {
			defer wg.Done()
			err1 := handler.GetDefaultHandlerSet().GetPodHandler().DeletePod(podToRestart.Namespace, podToRestart.Name)
			if err1 != nil {
				logrus.Errorf("failed to restart pod %s/%s : %s", podToRestart.Namespace, podToRestart.Name, err1.Error())
				mux.Lock()
				podsRestartFailed = append(podsRestartFailed, podToRestart.Namespace+"/"+podToRestart.Name)
				mux.Unlock()
				return
			}
		}(podToRestart)
	}

	wg.Wait()
	if len(podsRestartFailed) > 0 {
		err = fmt.Errorf("failed to restart pods : %v", podsRestartFailed)
		logrus.Errorf("failed to restart release : %s", err.Error())
		return err
	}

	logrus.Infof("succeed to restart release %s", releaseName)
	return nil
}

func (hc *HelmClient) InstallUpgradeReleaseWithRetry(namespace string, releaseRequest *release.ReleaseRequestV2, isSystem bool, chartFiles []*loader.BufferedFile, async bool, timeoutSec int64) error {
	retryTimes := 5
	for {
		err := hc.InstallUpgradeRelease(namespace, releaseRequest, isSystem, chartFiles, async, timeoutSec)
		if err != nil {
			if strings.Contains(err.Error(), "please wait for the release latest task") && retryTimes > 0 {
				logrus.Warnf("retry to install or upgrade release %s/%s after 2 second", namespace, releaseRequest.Name)
				retryTimes --
				time.Sleep(time.Second * 2)
				continue
			}
		}
		return err
	}
}

func (hc *HelmClient) InstallUpgradeRelease(namespace string, releaseRequest *release.ReleaseRequestV2, isSystem bool, chartFiles []*loader.BufferedFile, async bool, timeoutSec int64) error {
	if timeoutSec == 0 {
		timeoutSec = defaultTimeoutSec
	}

	oldReleaseTask, err := hc.validateReleaseTask(namespace, releaseRequest.Name, true)
	if err != nil {
		return err
	}

	releaseTaskArgs := &CreateReleaseTaskArgs{
		Namespace:      namespace,
		ReleaseRequest: releaseRequest,
		IsSystem:       isSystem,
		ChartFiles:     chartFiles,
	}
	taskSig, err := SendReleaseTask(releaseTaskArgs)
	if err != nil {
		logrus.Errorf("failed to send %s : %s", releaseTaskArgs.GetTaskName(), err.Error())
		return err
	}
	taskSig.TimeoutSec = timeoutSec

	releaseTask := &cache.ReleaseTask{
		Namespace:            namespace,
		Name:                 releaseRequest.Name,
		LatestReleaseTaskSig: taskSig,
	}

	err = hc.helmCache.CreateOrUpdateReleaseTask(releaseTask)
	if err != nil {
		logrus.Errorf("failed to set release task of %s/%s to redis: %s", namespace, releaseRequest.Name, err.Error())
		return err
	}

	if oldReleaseTask != nil && oldReleaseTask.LatestReleaseTaskSig != nil {
		err = task.GetDefaultTaskManager().PurgeTaskState(oldReleaseTask.LatestReleaseTaskSig.GetTaskSignature())
		if err != nil {
			logrus.Warnf("failed to purge task state : %s", err.Error())
		}
	}

	if !async {
		asyncResult := taskSig.GetAsyncResult()
		_, err = asyncResult.GetWithTimeout(time.Duration(timeoutSec)*time.Second, defaultSleepTimeSecond)
		if err != nil {
			logrus.Errorf("failed to create or update release  %s/%s: %s", namespace, releaseRequest.Name, err.Error())
			return err
		}
	}
	logrus.Infof("succeed to call create or update release %s/%s api", namespace, releaseRequest.Name)
	return nil
}

func (hc *HelmClient) doInstallUpgradeRelease(namespace string, releaseRequest *release.ReleaseRequestV2, isSystem bool, chartFiles []*loader.BufferedFile) error {
	update := true
	releaseCache, err := hc.helmCache.GetReleaseCache(namespace, releaseRequest.Name)
	if err != nil {
		if walmerr.IsNotFoundError(err) {
			update = false
		} else {
			logrus.Errorf("failed to get release cache of %s/%s : %s", namespace, releaseRequest.Name, err.Error())
			return err
		}
	}

	preProcessRequest(releaseRequest)

	hook.ProcessPrettyParams(&(releaseRequest.ReleaseRequest))

	// get all the dependency releases' output configs from ReleaseConfig
	dependencyConfigs, err := hc.getDependencyOutputConfigs(namespace, releaseRequest.Dependencies)
	if err != nil {
		logrus.Errorf("failed to get all the dependency releases' output configs : %s", err.Error())
		return err
	}

	// reuse config values
	configValues := map[string]interface{}{}
	if update {
		releaseInfo, err := hc.buildReleaseInfoV2(releaseCache)
		if err != nil {
			logrus.Errorf("failed to build release info : %s", err.Error())
			return err
		}
		util.MergeValues(configValues, releaseInfo.ConfigValues)

		if len(releaseInfo.ReleaseLabels) > 0 {
			oldProjectName, ok1 := releaseInfo.ReleaseLabels[cache.ProjectNameLabelKey]
			_, ok2 := releaseRequest.ReleaseLabels[cache.ProjectNameLabelKey]
			if ok1 && !ok2 {
				releaseRequest.ReleaseLabels[cache.ProjectNameLabelKey] = oldProjectName
			}
		}
	}
	util.MergeValues(configValues, releaseRequest.ConfigValues)

	var rawChart *chart.Chart
	var chartErr error
	if chartFiles != nil {
		rawChart, chartErr = loader.LoadFiles(chartFiles)
	} else {
		rawChart, chartErr = hc.LoadChart(releaseRequest.RepoName, releaseRequest.ChartName, releaseRequest.ChartVersion)
	}
	if chartErr != nil {
		logrus.Errorf("failed to load chart : %s", chartErr.Error())
		return chartErr
	}

	err = transwarpjsonnet.ProcessJsonnetChart(rawChart, namespace, releaseRequest.Name, configValues,
		dependencyConfigs, releaseRequest.Dependencies, releaseRequest.ReleaseLabels)
	if err != nil {
		logrus.Errorf("failed to ProcessJsonnetChart : %s", err.Error())
		return err
	}

	// add default plugin
	releaseRequest.Plugins = append(releaseRequest.Plugins, &walm.WalmPlugin{
		Name: plugins.LabelPodPluginName,
	})

	valueOverride := map[string]interface{}{}
	util.MergeValues(valueOverride, configValues)
	util.MergeValues(valueOverride, dependencyConfigs)
	valueOverride[walm.WalmPluginConfigKey] = releaseRequest.Plugins

	currentHelmClient, err := hc.GetCurrentHelmClient(namespace)
	if err != nil {
		logrus.Errorf("failed to get helm client : %s", err.Error())
		return err
	}

	releaseInfo, err := hc.doInstallUpgradeReleaseFromChart(currentHelmClient, namespace, releaseRequest, rawChart, valueOverride, update)
	if err != nil {
		logrus.Errorf("failed to create or update release from chart : %s", err.Error())
		return err
	}

	err = hc.helmCache.CreateOrUpdateReleaseCache(releaseInfo)
	if err != nil {
		logrus.Errorf("failed to create of update release cache of %s/%s : %s", namespace, releaseRequest.Name, err.Error())
		return err
	}

	logrus.Infof("succeed to create or update release %s/%s", namespace, releaseRequest.Name)

	return nil
}

func (hc *HelmClient) doInstallUpgradeReleaseFromChart(currentHelmClient *helm.Client, namespace string,
	releaseRequest *release.ReleaseRequestV2, rawChart *chart.Chart, valueOverride map[string]interface{},
	update bool) (releaseInfo *hapirelease.Release, err error) {
	if update {
		releaseInfo, err = currentHelmClient.UpdateReleaseFromChart(
			releaseRequest.Name,
			rawChart,
			helm.UpdateValueOverrides(valueOverride),
			helm.UpgradeDryRun(hc.dryRun),
		)
		if err != nil {
			logrus.Errorf("failed to upgrade release %s/%s from chart : %s", namespace, releaseRequest.Name, err.Error())
			return nil, err
		}
	} else {
		releaseInfo, err = currentHelmClient.InstallReleaseFromChart(
			rawChart,
			namespace,
			helm.ValueOverrides(valueOverride),
			helm.ReleaseName(releaseRequest.Name),
			helm.InstallDryRun(hc.dryRun),
		)
		if err != nil {
			logrus.Errorf("failed to install release %s/%s from chart : %s", namespace, releaseRequest.Name, err.Error())
			opts := []helm.UninstallOption{
				helm.UninstallPurge(true),
			}
			_, err1 := currentHelmClient.UninstallRelease(
				releaseRequest.Name, opts...,
			)
			if err1 != nil {
				logrus.Errorf("failed to rollback to delete release %s/%s : %s", namespace, releaseRequest.Name, err1.Error())
			}
			return nil, err
		}
	}
	return
}

func preProcessRequest(releaseRequest *release.ReleaseRequestV2) {
	if releaseRequest.ConfigValues == nil {
		releaseRequest.ConfigValues = map[string]interface{}{}
	}
	if releaseRequest.Dependencies == nil {
		releaseRequest.Dependencies = map[string]string{}
	}
	if releaseRequest.ReleaseLabels == nil {
		releaseRequest.ReleaseLabels = map[string]string{}
	}
}

func (hc *HelmClient) getDependencyOutputConfigs(namespace string, dependencies map[string]string) (dependencyConfigs map[string]interface{}, err error) {
	dependencyConfigs = map[string]interface{}{}
	for _, dependency := range dependencies {
		ss := strings.Split(dependency, "/")
		if len(ss) > 2 {
			err = fmt.Errorf("dependency value %s should not contains more than 1 \"/\"", dependency)
			return
		}
		dependencyNamespace, dependencyName := "", ""
		if len(ss) == 2 {
			dependencyNamespace = ss[0]
			dependencyName = ss[1]
		} else {
			dependencyNamespace = namespace
			dependencyName = ss[0]
		}
		dependencyReleaseConfig, err := hc.releaseConfigHandler.GetReleaseConfig(dependencyNamespace, dependencyName)
		if err != nil {
			if adaptor.IsNotFoundErr(err) {
				logrus.Warnf("release config %s/%s is not found", dependencyNamespace, dependencyName)
				continue
			}
			logrus.Errorf("failed to get release config %s/%s : %s", dependencyNamespace, dependencyName, err.Error())
			return nil, err
		}

		// TODO how to deal with key conflict?
		if len(dependencyReleaseConfig.Spec.OutputConfig) > 0 {
			util.MergeValues(dependencyConfigs, dependencyReleaseConfig.Spec.OutputConfig)
		}
	}
	return
}
