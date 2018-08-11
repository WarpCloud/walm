package project

import (
	"walm/pkg/release"
	"walm/pkg/release/manager/helm"
	"github.com/twmb/algoimpl/go/graph"
	)

func InitProject() {
	InitRedisClient()
}

func ListProjects(namespace string) (*release.ProjectInfoList, error) {
	return nil, nil
}

func CreateProject(namespace string, project string, projectParams *release.ProjectParams) error {
	helmExtraLabelsBase := map[string]interface{}{}
	helmExtraLabelsVals := release.HelmExtraLabels{}
	helmExtraLabelsVals.ProjectName = project
	helmExtraLabelsBase["HelmExtraLabels"] = &helmExtraLabelsVals

	rawValsBase := map[string]interface{}{}
	rawValsBase = mergeValues(rawValsBase, projectParams.CommonValues)
	rawValsBase = mergeValues(helmExtraLabelsBase, rawValsBase)

	releaseList, err := brainFuckChartDepParse(projectParams)
	if err != nil {
		return err
	}
	for _, releaseParams := range releaseList {
		releaseParams.ConfigValues = mergeValues(releaseParams.ConfigValues, rawValsBase)
		err = helm.InstallUpgradeRealese(releaseParams)
		if err != nil {
			return err
		}
	}
	return nil
}

func DeleteProject() {
}

func AddReleaseInProject(namespace string, project string, releaseParams *release.ReleaseRequest) error {
	return nil
}

func RemoveReleaseInProject(namespace string, project string, releaseParams *release.ReleaseRequest) error {
	return nil
}

func brainFuckChartDepParse(projectParams *release.ProjectParams) ([]*release.ReleaseRequest, error) {
	projectParamsMap := make(map[string]interface{})
	g := graph.New(graph.Directed)
	projectDepGraph := make(map[string]graph.Node, 0)
	releaseParsed := make([]*release.ReleaseRequest, 1)

	for _, releaseInfo := range projectParams.Releases {
		projectParamsMap[releaseInfo.ChartName] = &releaseInfo
	}

	// init node
	for _, helmRelease := range projectParams.Releases {
		projectDepGraph[helmRelease.ChartName] = g.MakeNode()
	}

	// init edge
	for _, helmRelease := range projectParams.Releases {
		subCharts, err := helm.GetDependencies(helmRelease.ChartName, helmRelease.ChartVersion)
		if err != nil {
			return nil, err
		}

		for _, subChartName := range subCharts {
			g.MakeEdge(projectDepGraph[helmRelease.ChartName], projectDepGraph[subChartName])
		}
	}

	sortedChartList := g.TopologicalSort()

	for i := range sortedChartList {
		newReleaseRequest := release.ReleaseRequest{}
		releaseRequest := projectParamsMap[(*sortedChartList[i].Value).(string)]
		newReleaseRequest.Name = releaseRequest.(release.ReleaseInfo).Name
		newReleaseRequest.Namespace = releaseRequest.(release.ReleaseInfo).Namespace
		newReleaseRequest.ChartName = releaseRequest.(release.ReleaseInfo).ChartName
		newReleaseRequest.ChartVersion = releaseRequest.(release.ReleaseInfo).ChartVersion
		newReleaseRequest.ConfigValues = releaseRequest.(release.ReleaseInfo).ConfigValues
		newReleaseRequest.Dependencies = releaseRequest.(release.ReleaseInfo).Dependencies
		if len(newReleaseRequest.Dependencies) == 0 {
			chartsNeighbors := g.Neighbors(sortedChartList[i])
			for _, chartNeighbor := range chartsNeighbors {
				newReleaseRequest.Dependencies[(*chartNeighbor.Value).(string)] = releaseRequest.(release.ReleaseInfo).Name
			}
		}
		releaseParsed = append(releaseParsed, &newReleaseRequest)
	}

	return releaseParsed, nil
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
