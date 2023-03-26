package helm

import (
	"WarpCloud/walm/pkg/models/release"
	"WarpCloud/walm/pkg/models/common"
	"WarpCloud/walm/pkg/helm/impl/plugins"
	"WarpCloud/walm/pkg/models/k8s"
)

type Helm interface {
	// paused :
	// 1. nil: nothing to do
	// 2. true: enable pause release plugin
	// 3. false: disable pause release plugin
	InstallOrCreateRelease(namespace string, releaseRequest *release.ReleaseRequestV2, chartFiles []*common.BufferedFile,
		dryRun bool, update bool, oldReleaseInfo *release.ReleaseInfoV2) (*release.ReleaseCache, error)
	InstallOrCreateReleaseWithStrict(namespace string, releaseRequest *release.ReleaseRequestV2, chartFiles []*common.BufferedFile,
		dryRun bool, update bool, oldReleaseInfo *release.ReleaseInfoV2, fullUpdate bool, strict bool) (*release.ReleaseCache, error)
	DeleteRelease(namespace string, name string) error
	PauseOrRecoverRelease(paused bool, oldReleaseInfo *release.ReleaseInfoV2) (*release.ReleaseCache, error)
	ListAllReleases() ([]*release.ReleaseCache, error)
	GetDependencyOutputConfigs(namespace string, dependencies map[string]string, chartInfo *release.ChartDetailInfo, strict bool) (dependencyConfigs map[string]interface{}, err error)

	GetChartDetailInfo(repoName, chartName, chartVersion string) (*release.ChartDetailInfo, error)
	GetChartList(repoName string) (*release.ChartInfoList, error)
	GetDetailChartInfoByImage(chartImage string) (*release.ChartDetailInfo, error)
	GetRepoList() *release.RepoInfoList
	GetChartAutoDependencies(repoName, chartName, chartVersion string) (subChartNames []string, err error)
}

func BuildReleasePluginsByConfigValues(configValues map[string]interface{}) (releasePlugins []*k8s.ReleasePlugin, hasPauseReleasePlugin bool, err error){
	releasePlugins = []*k8s.ReleasePlugin{}
	if configValues != nil {
		if walmPlugins, ok := configValues[plugins.WalmPluginConfigKey]; ok {
			delete(configValues, plugins.WalmPluginConfigKey)
			for _, plugin := range walmPlugins.([]interface{}) {
				walmPlugin := plugin.(map[string]interface{})
				if walmPlugin["name"].(string) != plugins.ValidateReleaseConfigPluginName &&
					walmPlugin["name"].(string) != plugins.IsomateSetConverterPluginName && !walmPlugin["disable"].(bool) {
					if walmPlugin["name"].(string) == plugins.PauseReleasePluginName {
						hasPauseReleasePlugin = true
					}
					releasePlugins = append(releasePlugins, &k8s.ReleasePlugin{
						Name:    walmPlugin["name"].(string),
						Args:    walmPlugin["args"].(string),
						Version: walmPlugin["version"].(string),
						Disable: walmPlugin["disable"].(bool),
					})
				}
			}
		}
	}
	return
}