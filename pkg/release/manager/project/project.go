package project

import (
	"walm/pkg/release"
	"walm/pkg/release/manager/helm"
	"github.com/twmb/algoimpl/go/graph"
	"fmt"
	"github.com/sirupsen/logrus"
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

	for _, releaseParams := range projectParams.Releases {
		releaseParams.Name = fmt.Sprintf("%s--%s", project, releaseParams.Name)
		fmt.Printf("%v\n", releaseParams.ConfigValues)
		releaseParams.ConfigValues = mergeValues(releaseParams.ConfigValues, rawValsBase)
	}

	releaseList, err := brainFuckChartDepParse(projectParams)
	if err != nil {
		return err
	}
	for _, releaseParams := range releaseList {
		err = helm.InstallUpgradeRealese(releaseParams)
		if err != nil {
			logrus.Errorf("CreateProject install release %s error %v\n", releaseParams.Name, err)
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
	releaseParsed := make([]*release.ReleaseRequest, 0)

	for _, releaseInfo := range projectParams.Releases {
		projectParamsMap[releaseInfo.ChartName] = &releaseInfo
	}

	// init node
	for _, helmRelease := range projectParams.Releases {
		projectDepGraph[helmRelease.ChartName] = g.MakeNode()
		*projectDepGraph[helmRelease.ChartName].Value = helmRelease
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
		releaseRequest := *(*sortedChartList[i].Value).(*release.ReleaseRequest)
		if len(releaseRequest.Dependencies) == 0 {
			chartsNeighbors := g.Neighbors(sortedChartList[i])
			for _, chartNeighbor := range chartsNeighbors {
				releaseRequest.Dependencies[(*chartNeighbor.Value).(*release.ReleaseRequest).ChartName] =
					releaseRequest.Name
			}
		}
		releaseParsed = append(releaseParsed, &releaseRequest)
	}

	for i, j := 0, len(releaseParsed)-1; i < j; i, j = i+1, j-1 {
		releaseParsed[i], releaseParsed[j] = releaseParsed[j], releaseParsed[i]
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
