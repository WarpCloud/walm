package project

import (
	"fmt"

	"github.com/twmb/algoimpl/go/graph"
	"github.com/sirupsen/logrus"

	"walm/pkg/release"
	"walm/pkg/release/manager/helm"
	"strings"
)

func InitProject() {
	InitRedisClient()
}

func ListProjects(namespace string) (*release.ProjectInfoList, error) {
	projectMap := make(map[string]*release.ProjectInfo)
	projectList := new(release.ProjectInfoList)

	option := &release.ReleaseListOption{}
	releaseList, err := helm.ListReleases(option)
	if err != nil {
		return nil, err
	}
	for _, releaseInfo := range releaseList {
		projectNameArray := strings.Split(releaseInfo.Name, "--")
		if len(projectNameArray) == 2 {
			projectName := projectNameArray[0]
			projectInfo, ok := projectMap[projectName]
			if ok {
				releaseInfo.Name = projectNameArray[1]
				projectInfo.Releases = append(projectInfo.Releases, releaseInfo)
			} else {
				projectMap[projectName] = new(release.ProjectInfo)
				projectMap[projectName].Name = projectName
				projectMap[projectName].Namespace = namespace
				releaseInfo.Name = projectNameArray[1]
				projectMap[projectName].Releases = append(projectMap[projectName].Releases, releaseInfo)
				projectList.Items = append(projectList.Items, projectMap[projectName])
			}
		}
	}
	return projectList, nil
}

func GetProjectInfo(namespace, projectName string) (*release.ProjectInfo, error) {
	found := false
	option := &release.ReleaseListOption{}
	projectInfo := new(release.ProjectInfo)
	releaseList, err := helm.ListReleases(option)
	if err != nil {
		return nil, err
	}
	for _, releaseInfo := range releaseList {
		projectNameArray := strings.Split(releaseInfo.Name, "--")
		if len(projectNameArray) == 2 {
			if projectName == projectNameArray[0] {
				found = true
				projectInfo.Name = projectName
				projectInfo.Namespace = namespace
				releaseInfo.Name = projectNameArray[1]
				projectInfo.Releases = append(projectInfo.Releases, releaseInfo)
			}
		}
	}
	if found {
		return projectInfo, nil
	}
	return nil, nil
}

func CreateProject(namespace string, project string, projectParams *release.ProjectParams) error {
	helmExtraLabelsBase := map[string]interface{}{}
	helmExtraLabelsVals := release.HelmExtraLabels{}
	helmExtraLabelsVals.HelmLabels = make(map[string]interface{})
	helmExtraLabelsVals.HelmLabels["project_name"] = project
	helmExtraLabelsBase["HelmExtraLabels"] = helmExtraLabelsVals

	rawValsBase := map[string]interface{}{}
	rawValsBase = mergeValues(rawValsBase, projectParams.CommonValues)
	rawValsBase = mergeValues(helmExtraLabelsBase, rawValsBase)

	for _, releaseParams := range projectParams.Releases {
		releaseParams.Name = fmt.Sprintf("%s--%s", project, releaseParams.Name)
		releaseParams.Namespace = namespace
		fmt.Printf("%v\n", releaseParams.ConfigValues)
		releaseParams.ConfigValues = mergeValues(releaseParams.ConfigValues, rawValsBase)
	}

	releaseList, err := brainFuckChartDepParse(projectParams)
	if err != nil {
		return err
	}
	for _, releaseParams := range releaseList {
		err = helm.InstallUpgradeRealese(namespace, releaseParams)
		if err != nil {
			logrus.Errorf("CreateProject install release %s error %v\n", releaseParams.Name, err)
			return err
		}
	}
	return nil
}

func DeleteProject(namespace string, project string) error {
	return nil
}

func AddReleaseInProject(namespace string, projectName string, releaseParams *release.ReleaseRequest) error {
	projectInfo, err := GetProjectInfo(namespace, projectName)
	if err != nil {
		return err
	}
	if projectInfo == nil {
		err = helm.InstallUpgradeRealese(releaseParams)
		if err != nil {
			logrus.Errorf("AddReleaseInProject install release %s error %v\n", releaseParams.Name, err)
			return err
		}
	}
	return nil
}

func RemoveReleaseInProject(namespace string, projectName string, releaseParams *release.ReleaseRequest) error {
	return nil
}

func brainFuckRuntimeDepParse(projectInfo *release.ProjectInfo, releaseParams *release.ReleaseRequest) ([]*release.ReleaseRequest, error) {
	subCharts, err := helm.GetDependencies(releaseParams.ChartName, releaseParams.ChartVersion)
	if err != nil {
		return nil, err
	}

	// Find Upstream Release
	for _, chartName := range subCharts {
		for _, releaseInfo := range projectInfo.Releases {
			releaseSubCharts, err := helm.GetDependencies(releaseInfo.ChartName, releaseInfo.ChartVersion)
			if err != nil {
				return nil, err
			}
			logrus.Infof("%s %v", chartName, releaseSubCharts)
		}
	}
	//projectParams := {
		
	//}
	// Find Downstream Release
	//for _, chartName := range subCharts {

	//}

	return nil, nil
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
		subCharts, err := helm.GetDependencies(helmRelease.RepoName, helmRelease.ChartName, helmRelease.ChartVersion)
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
					(*chartNeighbor.Value).(*release.ReleaseRequest).Name
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
