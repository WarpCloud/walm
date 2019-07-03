package helm

import (
	"WarpCloud/walm/pkg/models/release"
	"WarpCloud/walm/pkg/models/common"
	"k8s.io/helm/pkg/walm"
	"k8s.io/helm/pkg/walm/plugins"
)

type Helm interface {
	// paused :
	// 1. nil: nothing to do
	// 2. true: enable pause release plugin
	// 3. false: disable pause release plugin
	InstallOrCreateRelease(namespace string, releaseRequest *release.ReleaseRequestV2, chartFiles []*common.BufferedFile,
		dryRun bool, update bool, oldReleaseInfo *release.ReleaseInfoV2, paused *bool) (*release.ReleaseCache, error)
	DeleteRelease(namespace string, name string) error

	GetChartDetailInfo(repoName, chartName, chartVersion string) (*release.ChartDetailInfo, error)
	GetChartList(repoName string) (*release.ChartInfoList, error)
	GetDetailChartInfoByImage(chartImage string) (*release.ChartDetailInfo, error)
	GetRepoList() *release.RepoInfoList
	GetChartAutoDependencies(repoName, chartName, chartVersion string) (subChartNames []string, err error)
}

func BuildReleasePluginsByConfigValues(configValues map[string]interface{}) (releasePlugins []*release.ReleasePlugin, hasPauseReleasePlugin bool, err error){
	releasePlugins = []*release.ReleasePlugin{}
	if configValues != nil {
		if walmPlugins, ok := configValues[walm.WalmPluginConfigKey]; ok {
			delete(configValues, walm.WalmPluginConfigKey)
			for _, plugin := range walmPlugins.([]interface{}) {
				walmPlugin := plugin.(map[string]interface{})
				if walmPlugin["name"].(string) != plugins.ValidateReleaseConfigPluginName && !walmPlugin["disable"].(bool) {
					if walmPlugin["name"].(string) == plugins.PauseReleasePluginName {
						hasPauseReleasePlugin = true
					}
					releasePlugins = append(releasePlugins, &release.ReleasePlugin{
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