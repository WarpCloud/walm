package helm

import (
	hapirelease "k8s.io/helm/pkg/hapi/release"
	"time"
	"k8s.io/helm/pkg/helm"
	"github.com/sirupsen/logrus"
	"fmt"
	"strings"
	"walm/pkg/k8s/handler"
	"walm/pkg/k8s/adaptor"
	walmerr "walm/pkg/util/error"
	"walm/pkg/release"
	"sync"
	"errors"
	"mime/multipart"

	"k8s.io/helm/pkg/chart"
	"walm/pkg/hook"
	"walm/pkg/task"
	"walm/pkg/release/manager/helm/cache"
	"walm/pkg/redis"
	"walm/pkg/setting"
	"io/ioutil"
	"walm/pkg/k8s/client"
	"k8s.io/helm/pkg/storage/driver"
	"k8s.io/apimachinery/pkg/util/wait"
	"github.com/hashicorp/golang-lru"
	"github.com/ghodss/yaml"
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
	isJsonnetChart, nativeChart, jsonnetChart, err := hc.LoadChart(repoName, chartName, chartVersion)
	if err != nil {
		return nil, err
	}

	subChartNames = []string{}
	if isJsonnetChart {
		appYamlPath := fmt.Sprintf("templates/%s/%s/app.yaml", nativeChart.Metadata.Name, nativeChart.Metadata.AppVersion)
		for _, file := range jsonnetChart.Templates {
			if file.Name == appYamlPath {
				appDependency := &release.AppDependency{}
				err := yaml.Unmarshal(file.Data, &appDependency)
				if err != nil {
					return nil, err
				}
				for _, dependency := range appDependency.Dependencies {
					subChartNames = append(subChartNames, dependency.Name)
				}
				break
			}
		}
	}
	return
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

		d := driver.NewSecrets(clientset.CoreV1().Secrets(namespace))
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

func (hc *HelmClient) LoadChart(repoName, chartName, chartVersion string) (isJsonnetChart bool, nativeChart, jsonnetChart *chart.Chart,err error) {
	chartPath, err := hc.downloadChart(repoName, chartName, chartVersion)
	if err != nil {
		logrus.Errorf("failed to download chart : %s", err.Error())
		return false,  nil, nil, err
	}

	return loadChartByPath(chartPath)
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
			// Compatible
			releaseV2.DependenciesConfigValues = map[string]interface{}{}
			releaseV2.OutputConfigValues = map[string]interface{}{}
		} else {
			logrus.Errorf("failed to get release config : %s", err.Error())
			return nil, err
		}
	} else {
		releaseV2.ConfigValues = releaseConfig.Spec.ConfigValues
		releaseV2.Dependencies = releaseConfig.Spec.Dependencies
		releaseV2.DependenciesConfigValues = releaseConfig.Spec.DependenciesConfigValues
		releaseV2.OutputConfigValues = releaseConfig.Spec.OutputConfig
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

	releaseCacheMap := map[string]*release.ReleaseCache{}
	for _, releaseCache := range releaseCaches {
		releaseCacheMap[releaseCache.Namespace+"/"+releaseCache.Name] = releaseCache
	}

	releaseInfos := []*release.ReleaseInfoV2{}
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
		return nil, err
	}
	return releaseInfos, nil
}

// reload dependencies config values, if changes, upgrade release
func (hc *HelmClient) ReloadRelease(namespace, name string, isSystem bool) error {
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
			logrus.Error(err.Error())
			return
		}
	}
	return
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

func (hc *HelmClient) InstallUpgradeRelease(namespace string, releaseRequest *release.ReleaseRequestV2, isSystem bool, chartArchive multipart.File, async bool, timeoutSec int64) error {
	if timeoutSec == 0 {
		timeoutSec = defaultTimeoutSec
	}

	oldReleaseTask, err := hc.validateReleaseTask(namespace, releaseRequest.Name, true)
	if err != nil {
		logrus.Errorf("failed to validate release task : %s", err.Error())
		return err
	}

	releaseTaskArgs := &CreateReleaseTaskArgs{
		Namespace:      namespace,
		ReleaseRequest: releaseRequest,
		IsSystem:       isSystem,
		ChartArchive:   chartArchive,
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

func (hc *HelmClient) doInstallUpgradeRelease(namespace string, releaseRequest *release.ReleaseRequestV2, isSystem bool, chartArchive multipart.File) error {
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

	if releaseRequest.ConfigValues == nil {
		releaseRequest.ConfigValues = map[string]interface{}{}
	}
	if releaseRequest.Dependencies == nil {
		releaseRequest.Dependencies = map[string]string{}
	}

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
		releaseConfig, err := hc.releaseConfigHandler.GetReleaseConfig(namespace, releaseRequest.Name)
		if err != nil {
			if adaptor.IsNotFoundErr(err) {
				logrus.Warnf("release config %s/%s is not found", namespace, releaseRequest.Name)
				releaseInfo, err := hc.buildReleaseInfoV2(releaseCache)
				if err != nil {
					logrus.Errorf("failed to build release info : %s", err.Error())
					return err
				}
				mergeValues(configValues, releaseInfo.ConfigValues)
				if len(releaseInfo.Status.Instances) > 0 {
					err = fmt.Errorf("now v1 release %s/%s with instances is not support to upgrade", namespace, releaseRequest.Name)
					return err
				}
			} else {
				logrus.Errorf("failed to get release config : %s", err.Error())
				return err
			}
		} else {
			mergeValues(configValues, releaseConfig.Spec.ConfigValues)
		}
	}
	mergeValues(configValues, releaseRequest.ConfigValues)

	// if jsonnet chart, add template-jsonnet/, app.yaml to chart.Files
	// app.yaml : used to define chart dependency relations
	var chart, jsonnetChart *chart.Chart
	var isJsonnetChart bool
	if chartArchive != nil {
		isJsonnetChart, chart, jsonnetChart, err = loadChartByArchive(chartArchive)
		if err != nil {
			logrus.Errorf("failed to load chart by archive : %s", err.Error())
			return err
		}
	} else {
		isJsonnetChart, chart, jsonnetChart, err = hc.LoadChart(releaseRequest.RepoName, releaseRequest.ChartName, releaseRequest.ChartVersion)
		if err != nil {
			logrus.Errorf("failed to load chart : %s", err.Error())
			return err
		}
	}
	if err != nil {
		logrus.Errorf("failed to load chart %s-%s from %s : %s", releaseRequest.ChartName, releaseRequest.ChartVersion, releaseRequest.RepoName, err.Error())
		return err
	}

	if isJsonnetChart {
		nativeTemplates := chart.Templates
		chart, err = convertJsonnetChart(namespace, releaseRequest.Name, releaseRequest.Dependencies, jsonnetChart, configValues, dependencyConfigs)
		if err != nil {
			logrus.Errorf("failed to convert jsonnet chart %s-%s from %s : %s", releaseRequest.ChartName, releaseRequest.ChartVersion, releaseRequest.RepoName, err.Error())
			return err
		}
		if len(nativeTemplates) > 0 {
			chart.Templates = append(chart.Templates, nativeTemplates...)
		}
	} else {
		//TODO native helm chart如何处理？
	}

	valueOverride := map[string]interface{}{}
	mergeValues(valueOverride, configValues)

	var release *hapirelease.Release
	currentHelmClient, err := hc.GetCurrentHelmClient(namespace)
	if err != nil {
		logrus.Errorf("failed to get helm client : %s", err.Error())
		return err
	}

	if update {
		release, err = currentHelmClient.UpdateReleaseFromChart(
			releaseRequest.Name,
			chart,
			helm.UpdateValueOverrides(valueOverride),
			helm.UpgradeDryRun(hc.dryRun),
		)
		if err != nil {
			logrus.Errorf("failed to upgrade release %s/%s from chart : %s", namespace, releaseRequest.Name, err.Error())
			return err
		}
	} else {
		release, err = currentHelmClient.InstallReleaseFromChart(
			chart,
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
			return err
		}
	}

	err = hc.helmCache.CreateOrUpdateReleaseCache(release)
	if err != nil {
		logrus.Errorf("failed to create of update release cache of %s/%s : %s", namespace, releaseRequest.Name, err.Error())
		return err
	}

	logrus.Infof("succeed to create or update release %s/%s", namespace, releaseRequest.Name)

	return nil
}

func (hc *HelmClient) getDependencyOutputConfigs(namespace string, dependencies map[string]string) (dependencyConfigs map[string]interface{}, err error) {
	dependencyConfigs = map[string]interface{}{}
	for _, dependency := range dependencies {
		ss := strings.Split(dependency, ".")
		if len(ss) > 2 {
			err = fmt.Errorf("dependency value %s should not contains more than 1 \".\"", dependency)
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
			// compatible
			provideConfigValues, ok := dependencyReleaseConfig.Spec.OutputConfig["provides"].(map[string]interface{})
			if ok {
				valueToMerge := make(map[string]interface{}, len(provideConfigValues))
				for key, value := range provideConfigValues {
					if immediateValue, ok := value.(map[string]interface{}); ok {
						if immediateValue["immediate_value"] != nil {
							valueToMerge[key] = immediateValue["immediate_value"]
						}
					}
				}
				mergeValues(dependencyConfigs, valueToMerge)
			} else {
				mergeValues(dependencyConfigs, dependencyReleaseConfig.Spec.OutputConfig)
			}
		}
	}
	return
}
