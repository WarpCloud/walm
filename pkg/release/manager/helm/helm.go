package helm

import (
	"sync"
	"errors"
	"time"
	"fmt"
	"io/ioutil"
	"strings"

	"walm/pkg/setting"
	"walm/pkg/release"

	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/storage/driver"
	"walm/pkg/release/manager/helm/cache"
	"walm/pkg/redis"
	"walm/pkg/k8s/client"
	"k8s.io/apimachinery/pkg/util/wait"
	hapiRelease "k8s.io/helm/pkg/proto/hapi/release"
	walmerr "walm/pkg/util/error"
	"k8s.io/helm/pkg/transwarp"
	"walm/pkg/k8s/handler"
	"walm/pkg/k8s/adaptor"
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
			info, err1 := buildReleaseInfo(releaseCache)
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
			info, err1 := buildReleaseInfo(releaseCache)
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

func (client *HelmClient) GetRelease(namespace, releaseName string) (release *release.ReleaseInfo, err error) {
	logrus.Debugf("Enter GetRelease %s %s\n", namespace, releaseName)
	releaseCache, err := client.helmCache.GetReleaseCache(namespace, releaseName)
	if err != nil {
		logrus.Errorf("failed to get release cache of %s : %s", releaseName, err.Error())
		return nil, err
	}
	release, err = buildReleaseInfo(releaseCache)
	if err != nil {
		logrus.Errorf("failed to build release info: %s\n", err.Error())
		return
	}
	return
}

//TODO
func (client *HelmClient) GetReleaseConfigs() (releaseConfigs []*release.ReleaseConfig, err error) {
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

func (client *HelmClient) UpgradeRealese(namespace string, releaseRequest *release.ReleaseRequest) error {
	if releaseRequest.ConfigValues == nil {
		releaseRequest.ConfigValues = map[string]interface{}{}
	}
	if releaseRequest.Dependencies == nil {
		releaseRequest.Dependencies = map[string]string{}
	}
	chartRequested, err := client.getChartRequest(releaseRequest.RepoName, releaseRequest.ChartName, releaseRequest.ChartVersion)
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
	releaseInfo.Dependencies = releaseRequest.Dependencies
	mergeValues(releaseRequest.ConfigValues, releaseInfo.ConfigValues)
	helmRelease, err := client.installChart(releaseRequest.Name, namespace, releaseRequest.ConfigValues, depLinks, chartRequested, false)
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

func (client *HelmClient) InstallUpgradeRealese(namespace string, releaseRequest *release.ReleaseRequest, isSystem bool) error {
	if releaseRequest.ConfigValues == nil {
		releaseRequest.ConfigValues = map[string]interface{}{}
	}
	if releaseRequest.Dependencies == nil {
		releaseRequest.Dependencies = map[string]string{}
	}
	chartRequested, err := client.getChartRequest(releaseRequest.RepoName, releaseRequest.ChartName, releaseRequest.ChartVersion)
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

func (client *HelmClient) DeleteRelease(namespace, releaseName string, isSystem bool) error {
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

	_, err := client.GetRelease(namespace, releaseName)
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

	logrus.Infof("succeed to delete release %s/%s", namespace, releaseName)
	return err
}

func (client *HelmClient) GetDependencies(repoName, chartName, chartVersion string) (subChartNames []string, err error) {
	logrus.Debugf("Enter GetDependencies %s %s\n", chartName, chartVersion)
	chartRequested, err := client.getChartRequest(repoName, chartName, chartVersion)
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

func (client *HelmClient) getChartRequest(repoName, chartName, chartVersion string) (*chart.Chart, error) {
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

func (client *HelmClient) installChart(releaseName, namespace string, configValues map[string]interface{}, depLinks map[string]interface{}, chart *chart.Chart, isSystem bool) (*hapiRelease.Release, error) {
	currentHelmClient := client.systemClient
	configVals, err := yaml.Marshal(configValues)
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
	releaseHistory, err := currentHelmClient.ReleaseHistory(releaseName, helm.WithMaxHistory(1))
	if err == nil {
		previousReleaseNamespace := releaseHistory.Releases[0].Namespace
		//TODO is it reasonable? it is wield, helm should support the same name in different namespace
		if previousReleaseNamespace != namespace {
			logrus.Warnf("namespace %s doesn't match with previous, release will be deployed to %s",
				namespace, previousReleaseNamespace,
			)
		}
		previousValues := map[string]interface{}{}
		if err := yaml.Unmarshal([]byte(releaseHistory.Releases[0].Chart.Values.Raw), &previousValues); err != nil {
			return nil, fmt.Errorf("failed to parse rawValues: %s", err)
		}
		mergedValues := map[string]interface{}{}
		mergedValues = mergeValues(mergedValues, configValues)
		mergedValues = mergeValues(previousValues, mergedValues)
		mergedVals, err := yaml.Marshal(mergedValues)
		if err != nil {
			logrus.Errorf("failed to marshal mergedVals values: %s", err.Error())
			return nil, err
		}

		logrus.Infof("UpdateChart %s DepLinks %+v ConfigValues %+v, previousValues %+v", releaseName, depLinks, mergedValues, previousValues)
		appConfigMapName, transwarpAppType, appDependency, err := transwarp.ProcessTranswarpChartRequested(chart, releaseName, namespace)
		if err != nil {
			return nil, err
		}
		if transwarpAppType == true {
			dependencies, err := transwarp.GetTranswarpInstanceCRDDependency(currentHelmClient, appDependency, depLinks, namespace, false)
			if err != nil {
				return nil, err
			}

			err = transwarp.ProcessTranswarpInstanceCRD(chart, releaseName, namespace, string(mergedVals[:]), appConfigMapName, dependencies)
			if err != nil {
				return nil, err
			}
		}

		resp, err := currentHelmClient.UpdateReleaseFromChart(
			releaseName,
			chart,
			helm.UpdateValueOverrides([]byte(chart.Values.Raw)),
			helm.ReuseValues(true),
			helm.UpgradeDryRun(client.dryRun),
		)
		if err != nil {
			//TODO should rollback to prev version?
			logrus.Errorf("failed to update release %s from chart : %s", releaseName, err.Error())
			return nil, err
		}
		helmRelease = resp.GetRelease()
	} else if strings.Contains(err.Error(), driver.ErrReleaseNotFound(releaseName).Error()) {
		logrus.Infof("InstallChart %s DepLinks %+v ConfigValues %+v", releaseName, depLinks, configValues)
		appConfigMapName, transwarpAppType, appDependency, err := transwarp.ProcessTranswarpChartRequested(chart, releaseName, namespace)
		if err != nil {
			return nil, err
		}
		if transwarpAppType == true {
			dependencies, err := transwarp.GetTranswarpInstanceCRDDependency(currentHelmClient, appDependency, depLinks, namespace, false)
			if err != nil {
				return nil, err
			}

			err = transwarp.ProcessTranswarpInstanceCRD(chart, releaseName, namespace, string(configVals[:]), appConfigMapName, dependencies)
			if err != nil {
				return nil, err
			}
		}

		resp, err := currentHelmClient.InstallReleaseFromChart(
			chart,
			namespace,
			helm.ValueOverrides([]byte(chart.Values.Raw)),
			helm.ReleaseName(releaseName),
			helm.InstallDryRun(client.dryRun),
		)
		if err != nil {
			logrus.Errorf("failed to install release %s from chart : %s", releaseName, err.Error())
			opts := []helm.DeleteOption{
				helm.DeletePurge(true),
			}
			_, err1 := currentHelmClient.DeleteRelease(
				releaseName, opts...,
			)
			if err1 != nil {
				logrus.Errorf("failed to rollback to delete release %s : %s", releaseName, err1.Error())
			}
			return nil, err
		}
		helmRelease = resp.GetRelease()
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

func (client *HelmClient) DeployTillerCharts(namespace string) error {
	tillerRelease := release.ReleaseRequest{}
	tillerRelease.Name = fmt.Sprintf("tenant-tiller-%s", namespace)
	tillerRelease.ChartName = "helm-tiller-tenant"
	tillerRelease.ConfigValues = make(map[string]interface{}, 0)
	tillerRelease.ConfigValues["tiller"] = map[string]string{
		"image": setting.Config.MultiTenantConfig.TillerImage,
	}
	err := client.InstallUpgradeRealese(namespace, &tillerRelease, true)
	logrus.Infof("tenant %s deploy tiller %v\n", namespace, err)

	return err
}

func mergeValues(dest map[string]interface{}, src map[string]interface{}) map[string]interface{} {
	for k, v := range src {
		// If the key doesn't exist already, then just set the key to that value
		if _, exists := dest[k]; !exists {
			dest[k] = v
			continue
		}
		nextMap, ok := v.(map[string]interface{})
		// If it isn't another map, overwrite the value
		if !ok {
			dest[k] = v
			continue
		}
		// Edge case: If the key exists in the destination, but isn't a map
		destMap, isMap := dest[k].(map[string]interface{})
		// If the source map has a map for this key, prefer it
		if !isMap {
			dest[k] = v
			continue
		}
		// If we got to this point, it is a map in both, so merge them
		dest[k] = mergeValues(destMap, nextMap)
	}
	return dest
}
