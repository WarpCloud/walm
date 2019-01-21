package helm

import (
	"sync"
	"errors"
	"time"
	"fmt"
	"io/ioutil"
	"strings"
	"walm/pkg/hook"

	"walm/pkg/setting"
	"walm/pkg/release"

	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/storage/driver"
	"walm/pkg/release/manager/helm/cache"
	"walm/pkg/redis"
	"walm/pkg/k8s/client"
	"k8s.io/apimachinery/pkg/util/wait"
	walmerr "walm/pkg/util/error"
	"walm/pkg/k8s/handler"
	"walm/pkg/k8s/adaptor"
	"mime/multipart"
)

const (
	helmCacheDefaultResyncInterval time.Duration = 5 * time.Minute
	multiTenantClientsMaxSize int = 128
)

type ChartRepository struct {
	Name     string
	URL      string
	Username string
	Password string
}

type HelmClient struct {
	systemClient            *helm.Client
	multiTenantClients      *cache.MultiTenantClientsCache
	chartRepoMap            map[string]*ChartRepository
	dryRun                  bool
	helmCache               *cache.HelmCache
	helmCacheResyncInterval time.Duration
}

var helmClient *HelmClient

func GetDefaultHelmClient() *HelmClient {
	if helmClient == nil {
		tillerHost := setting.Config.SysHelm.TillerHost
		client1 := helm.NewClient(helm.Host(tillerHost))
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

		multiTenantClients := cache.NewMultiTenantClientsCache(multiTenantClientsMaxSize)
		helmCache := cache.NewHelmCache(redis.GetDefaultRedisClient(), client1, multiTenantClients, client.GetKubeClient())

		helmClient = &HelmClient{
			systemClient:            client1,
			multiTenantClients:      multiTenantClients,
			chartRepoMap:            chartRepoMap,
			dryRun:                  false,
			helmCache:               helmCache,
			helmCacheResyncInterval: helmCacheDefaultResyncInterval,
		}
	}
	return helmClient
}

func InitHelmByParams(tillerHost string, chartRepoMap map[string]*ChartRepository, dryRun bool) {
	client := helm.NewClient(helm.Host(tillerHost))

	helmClient = &HelmClient{
		systemClient: client,
		chartRepoMap: chartRepoMap,
		dryRun:       dryRun,
	}
}

func (client *HelmClient) GetDryRun() bool {
	return client.dryRun
}

//Deprecated
func (client *HelmClient) ListReleases(namespace, filter string) ([]*release.ReleaseInfo, error) {
	logrus.Debugf("Enter ListReleases namespace=%s filter=%s\n", namespace, filter)
	releaseCaches, err := client.helmCache.GetReleaseCaches(namespace, filter, 0)
	if err != nil {
		logrus.Errorf("failed to get release caches with namespace=%s filter=%s : %s", namespace, filter, err.Error())
		return nil, err
	}

	releaseInfos := []*release.ReleaseInfo{}
	mux := &sync.Mutex{}
	var wg sync.WaitGroup
	for _, releaseCache := range releaseCaches {
		wg.Add(1)
		go func(releaseCache *release.ReleaseCache) {
			defer wg.Done()
			info, err1 := BuildReleaseInfo(releaseCache)
			if err1 != nil {
				err = errors.New(fmt.Sprintf("failed to build release info: %s", err1.Error()))
				logrus.Error(err.Error())
				return
			}
			mux.Lock()
			releaseInfos = append(releaseInfos, info)
			mux.Unlock()
		}(releaseCache)
	}
	wg.Wait()
	if err != nil {
		return nil, err
	}
	return releaseInfos, nil
}

//Deprecated
func (client *HelmClient) GetReleasesByNames(namespace string, names ...string) ([]*release.ReleaseInfo, error) {
	releaseCaches, err := client.helmCache.GetReleaseCachesByNames(namespace, names...)
	if err != nil {
		logrus.Errorf("failed to get release caches : %s", err.Error())
		return nil, err
	}

	releaseInfos := []*release.ReleaseInfo{}
	mux := &sync.Mutex{}
	var wg sync.WaitGroup
	for _, releaseCache := range releaseCaches {
		wg.Add(1)
		go func(releaseCache *release.ReleaseCache) {
			defer wg.Done()
			info, err1 := BuildReleaseInfo(releaseCache)
			if err1 != nil {
				err = errors.New(fmt.Sprintf("failed to build release info: %s\n", err1.Error()))
				logrus.Error(err.Error())
				return
			}
			mux.Lock()
			releaseInfos = append(releaseInfos, info)
			mux.Unlock()
		}(releaseCache)
	}
	wg.Wait()
	if err != nil {
		return nil, err
	}
	return releaseInfos, nil
}

//Deprecated
func (client *HelmClient) GetRelease(namespace, releaseName string) (release *release.ReleaseInfo, err error) {
	logrus.Debugf("Enter GetRelease %s %s\n", namespace, releaseName)
	releaseCache, err := client.helmCache.GetReleaseCache(namespace, releaseName)
	if err != nil {
		logrus.Errorf("failed to get release cache of %s : %s", releaseName, err.Error())
		return nil, err
	}
	release, err = BuildReleaseInfo(releaseCache)
	if err != nil {
		logrus.Errorf("failed to build release info: %s\n", err.Error())
		return
	}
	return
}

func (client *HelmClient) RestartRelease(namespace, releaseName string) error {
	logrus.Debugf("Enter RestartRelease %s %s\n", namespace, releaseName)
	releaseInfo, err := client.GetRelease(namespace, releaseName)
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

//Deprecated
func (client *HelmClient) UpgradeRealese(namespace string, releaseRequest *release.ReleaseRequest, chartArchive multipart.File) (err error) {
	if releaseRequest.ConfigValues == nil {
		releaseRequest.ConfigValues = map[string]interface{}{}
	}
	if releaseRequest.Dependencies == nil {
		releaseRequest.Dependencies = map[string]string{}
	}
	hook.ProcessPrettyParams(releaseRequest)
	var chartRequested *chart.Chart
	if chartArchive != nil {
		chartRequested, err = GetChart(chartArchive)
	} else {
		chartRequested, err = client.GetChartRequest(releaseRequest.RepoName, releaseRequest.ChartName, releaseRequest.ChartVersion)
	}
	if err != nil {
		logrus.Errorf("failed to get chart %s/%s:%s", releaseRequest.RepoName, releaseRequest.ChartName, releaseRequest.ChartVersion)
		return err
	}
	depLinks := make(map[string]interface{})
	for k, v := range releaseRequest.Dependencies {
		depLinks[k] = v
	}
	releaseInfo, err := client.GetRelease(namespace, releaseRequest.Name)
	if err != nil {
		return err
	}
	logrus.Infof("releaseInfo.Dependencies OLD(%+v) NEW(%+v) %+v %+v\n", releaseInfo.Dependencies, releaseRequest.Dependencies, releaseInfo.Name, releaseInfo.ConfigValues)
	for k, v := range releaseInfo.Dependencies {
		depLinks[k] = v
	}
	for k, v := range releaseRequest.Dependencies {
		depLinks[k] = v
	}
	//releaseInfo.Dependencies = releaseRequest.Dependencies
	tempConfigValues := make(map[string]interface{}, 0)
	MergeValues(tempConfigValues, releaseInfo.ConfigValues)
	MergeValues(tempConfigValues, releaseRequest.ConfigValues)
	helmRelease, err := client.installChart(releaseRequest.Name, namespace, tempConfigValues, depLinks, chartRequested, false)
	if err != nil {
		logrus.Errorf("failed to install chart : %s", err.Error())
		return err
	}

	err = client.helmCache.CreateOrUpdateReleaseCache(helmRelease)
	if err != nil {
		logrus.Errorf("failed to create of update release cache of %s : %s", helmRelease.Name, err.Error())
		return err
	}

	logrus.Infof("succeed to update release %s", releaseRequest.Name)
	return nil
}

//Deprecated
func (client *HelmClient) InstallUpgradeRealese(namespace string, releaseRequest *release.ReleaseRequest, isSystem bool, chartArchive multipart.File) (err error) {
	if releaseRequest.ConfigValues == nil {
		releaseRequest.ConfigValues = map[string]interface{}{}
	}
	if releaseRequest.Dependencies == nil {
		releaseRequest.Dependencies = map[string]string{}
	}
	hook.ProcessPrettyParams(releaseRequest)
	var chartRequested *chart.Chart
	if chartArchive != nil {
		chartRequested, err = GetChart(chartArchive)
	} else {
		chartRequested, err = client.GetChartRequest(releaseRequest.RepoName, releaseRequest.ChartName, releaseRequest.ChartVersion)
	}
	if err != nil {
		logrus.Errorf("failed to get chart %s/%s:%s", releaseRequest.RepoName, releaseRequest.ChartName, releaseRequest.ChartVersion)
		return err
	}

	depLinks := make(map[string]interface{})
	for k, v := range releaseRequest.Dependencies {
		depLinks[k] = v
	}
	helmRelease, err := client.installChart(releaseRequest.Name, namespace, releaseRequest.ConfigValues, depLinks, chartRequested, isSystem)
	if err != nil {
		logrus.Errorf("failed to install chart : %s", err.Error())
		return err
	}

	err = client.helmCache.CreateOrUpdateReleaseCache(helmRelease)
	if err != nil {
		logrus.Errorf("failed to create of update release cache of %s : %s", helmRelease.Name, err.Error())
		return err
	}

	logrus.Infof("succeed to create or update release %s", releaseRequest.Name)
	return nil
}

func (client *HelmClient) RollbackRealese(namespace, releaseName, version string) error {
	return nil
}

//Deprecated
func (client *HelmClient) DeleteRelease(namespace, releaseName string, isSystem bool, deletePvcs bool) error {
	logrus.Debugf("Enter DeleteRelease %s %s\n", namespace, releaseName)
	currentHelmClient := client.systemClient

	if !isSystem {
		multiTenant, err := cache.IsMultiTenant(namespace)
		if err != nil {
			logrus.Errorf("InstallChart IsMultiTenant error %s\n", err.Error())
		}
		if multiTenant {
			tillerHosts := fmt.Sprintf("tiller-tenant.%s.svc:44134", namespace)
			currentHelmClient = client.multiTenantClients.Get(tillerHosts)
		}
	}

	releaseInfo, err := client.GetRelease(namespace, releaseName)
	if err != nil {
		if walmerr.IsNotFoundError(err) {
			logrus.Warnf("release %s is not found in redis", releaseName)
			return nil
		}
		logrus.Errorf("failed to get release %s : %s", releaseName, err.Error())
		return err
	}

	opts := []helm.DeleteOption{
		helm.DeletePurge(true),
	}
	res, err := currentHelmClient.DeleteRelease(
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

	err = client.helmCache.DeleteReleaseCache(namespace, releaseName)
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
			if instance.Modules != nil && len(instance.Modules.StatefulSets) > 0{
				statefulSets = append(statefulSets, instance.Modules.StatefulSets...)
			}
		}

		for _, statefulSet := range statefulSets {
			if statefulSet.Selector != nil && (len(statefulSet.Selector.MatchLabels) > 0 || len(statefulSet.Selector.MatchExpressions) > 0){
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
	return err
}

func (client *HelmClient) GetDependencies(repoName, chartName, chartVersion string) (subChartNames []string, err error) {
	logrus.Debugf("Enter GetDependencies %s %s\n", chartName, chartVersion)
	chartRequested, err := client.GetChartRequest(repoName, chartName, chartVersion)
	if err != nil {
		return nil, err
	}
	dependencies, err := parseChartDependencies(chartRequested)
	if err != nil {
		return nil, err
	}
	return dependencies, nil
}

func (client *HelmClient) GetHelmCache() *cache.HelmCache {
	return client.helmCache
}

func(client *HelmClient) GetCurrentHelmClient(namespace string, isSystem bool) (*helm.Client, error) {
	currentHelmClient := client.systemClient
	if !isSystem {
		multiTenant, err := cache.IsMultiTenant(namespace)
		if err != nil {
			logrus.Errorf("failed to check whether is multi tenant", err.Error())
			return nil, err
		}
		if multiTenant {
			tillerHosts := fmt.Sprintf("tiller-tenant.%s.svc:44134", namespace)
			currentHelmClient = client.multiTenantClients.Get(tillerHosts)
		}
	}

	//TODO improve
	retry := 20
	for i := 0; i < retry; i++ {
		err := currentHelmClient.PingTiller()
		if err == nil {
			break
		}
		if i == retry-1 {
			return nil, fmt.Errorf("tiller is not ready, PingTiller timeout: %s", err.Error())
		}
		time.Sleep(500 * time.Millisecond)
	}
	return currentHelmClient, nil
}

func (client *HelmClient) downloadChart(repoName, charName, version string) (string, error) {
	if repoName == "" {
		repoName = "stable"
	}
	repo, ok := client.chartRepoMap[repoName]
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

func (client *HelmClient) GetChartRequest(repoName, chartName, chartVersion string) (*chart.Chart, error) {
	chartPath, err := client.downloadChart(repoName, chartName, chartVersion)
	if err != nil {
		logrus.Errorf("failed to download chart : %s", err.Error())
		return nil, err
	}
	chartRequested, err := chartutil.Load(chartPath)
	if err != nil {
		logrus.Errorf("failed to load chart : %s", err.Error())
		return nil, err
	}

	return chartRequested, nil
}

func GetChart(chartArchive multipart.File) (*chart.Chart, error) {
	chartRequested, err := chartutil.LoadArchive(chartArchive)
	if err != nil {
		logrus.Errorf("failed to load chart : %s", err.Error())
		return nil, err
	}

	return chartRequested, nil
}

func (client *HelmClient) installChart(releaseName, namespace string, configValues map[string]interface{}, depLinks map[string]interface{}, chart *chart.Chart, isSystem bool) (*hapiRelease.Release, error) {
	currentHelmClient := client.systemClient
	if err != nil {
		logrus.Errorf("failed to marshal config values: %s", err.Error())
		return nil, err
	}

	if !isSystem {
		multiTenant, err := cache.IsMultiTenant(namespace)
		if err != nil {
			logrus.Errorf("InstallChart IsMultiTenant error %s\n", err.Error())
		}
		if multiTenant {
			tillerHosts := fmt.Sprintf("tiller-tenant.%s.svc:44134", namespace)
			currentHelmClient = client.multiTenantClients.Get(tillerHosts)
		}
	}

	retry := 20
	for i := 0; i < retry; i++ {
		err = currentHelmClient.PingTiller()
		if err == nil {
			break
		}
		if i == retry-1 {
			return nil, fmt.Errorf("tiller is not ready, PingTiller timeout: %s", err.Error())
		}
		time.Sleep(500 * time.Millisecond)
	}

	logrus.Infof("InstallChart Params %s %s %+v %+v", releaseName, namespace, configValues, depLinks)
	helmRelease := &hapiRelease.Release{}
	_, err := currentHelmClient.ReleaseHistory(releaseName, helm.WithMaxHistory(1))
	if err == nil {
		return nil, nil
	} else {
		logrus.Errorf("failed to get release history : %s", err.Error())
		return nil, err
	}

	return helmRelease, nil
}

func (client *HelmClient) StartResyncReleaseCaches(stopCh <-chan struct{}) {
	logrus.Infof("start to resync release cache every %v", client.helmCacheResyncInterval)
	// first time should be sync
	client.helmCache.Resync()
	firstTime := true
	go wait.Until(func() {
		if firstTime {
			time.Sleep(client.helmCacheResyncInterval)
			firstTime = false
		}
		client.helmCache.Resync()
	}, client.helmCacheResyncInterval, stopCh)
}


