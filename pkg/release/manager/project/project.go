package project

import (
	"github.com/twmb/algoimpl/go/graph"
	"github.com/sirupsen/logrus"

	"walm/pkg/release"
	"walm/pkg/release/manager/helm"
	"walm/pkg/redis"
	"walm/pkg/job"
	"encoding/json"
	"errors"
	walmerr "walm/pkg/util/error"
	"strings"
	"sync"
)

type ProjectManager struct {
	helmClient  *helm.HelmClient
	redisClient *redis.RedisClient
}

var projectManager *ProjectManager

func GetDefaultProjectManager() *ProjectManager {
	return projectManager
}

func InitProject() {
	projectManager = &ProjectManager{
		helmClient:  helm.GetDefaultHelmClient(),
		redisClient: redis.GetDefaultRedisClient(),
	}
}

func (manager *ProjectManager) ListProjects(namespace string) (*release.ProjectInfoList, error) {
	projectCaches, err := getProjectCachesFromRedis(manager.redisClient, namespace)
	if err != nil {
		logrus.Errorf("failed to get project caches in namespace %s : %s", namespace, err.Error())
		return nil, err
	}

	projectInfoList := &release.ProjectInfoList{
		Items: []*release.ProjectInfo{},
	}

	//TODO 多线程
	for _, projectCache := range projectCaches {
		projectInfo, err := manager.buildProjectInfo(projectCache)
		if err != nil {
			logrus.Errorf("failed to build project info from project cache of %s/%s : %s", projectCache.Namespace, projectCache.Name, err.Error())
			return nil, err
		}
		projectInfoList.Items = append(projectInfoList.Items, projectInfo)
	}

	projectInfoList.Num = len(projectInfoList.Items)
	return projectInfoList, nil
}

func (manager *ProjectManager) GetProjectInfo(namespace, projectName string) (*release.ProjectInfo, error) {
	projectCache, err := getProjectCacheFromRedis(manager.redisClient, namespace, projectName)
	if err != nil {
		logrus.Errorf("failed to get project cache of %s/%s : %s", namespace, projectName, err.Error())
		return nil, err
	}

	return manager.buildProjectInfo(projectCache)
}

func (manager *ProjectManager) buildProjectInfo(projectCache *release.ProjectCache) (projectInfo *release.ProjectInfo, err error) {
	projectInfo = &release.ProjectInfo{
		Name:                  projectCache.Name,
		Namespace:             projectCache.Namespace,
		CommonValues:          map[string]interface{}{},
		CreateProjectJobState: projectCache.CreateProjectJobState,
		Releases:              []*release.ReleaseInfo{},
	}

	if len(projectCache.InstalledReleases) > 0 {
		projectInfo.Releases, err = manager.helmClient.GetReleasesByNames(projectCache.Namespace, projectCache.InstalledReleases...)
		if err != nil {
			logrus.Errorf("failed to get project release info of %s/%s : %s", projectCache.Namespace, projectCache.Name, err.Error())
			return
		}
	}

	if projectInfo.CreateProjectJobState.CreateProjectJobStatus == "Succeed" && len(projectCache.Releases) == len(projectInfo.Releases) {
		projectInfo.Ready = true
		for _, release := range projectInfo.Releases {
			if !release.Ready {
				projectInfo.Ready = false
				break
			}
		}
	}
	return
}

func (manager *ProjectManager) CreateProject(namespace string, project string, projectParams *release.ProjectParams) error {
	if len(projectParams.Releases) == 0 {
		return errors.New("project releases can not be empty")
	}

	createProjectJob := &CreateProjectJob{
		Namespace:     namespace,
		Name:          project,
		ProjectParams: projectParams,
	}

	projectCache := buildProjectCache(namespace, project, "Pending", projectParams)
	err := setProjectCacheToRedis(manager.redisClient, projectCache)
	if err != nil {
		logrus.Errorf("failed to set project cache of %s/%s to redis: %s", namespace, project, err.Error())
		return err
	}

	//TODO support both sync & async
	jobId, err := job.GetDefaultWalmJobManager().CreateWalmJob("", "CreateProject", createProjectJob)
	if err != nil {
		logrus.Errorf("failed to create Async CreateProject Job : %s", err.Error())
		//Rollback
		deleteProjectCacheFromRedis(manager.redisClient, namespace, project)
		return err
	}
	logrus.Infof("succeed to create Async CreateProject Job %s", jobId)

	return nil
}

func buildProjectCache(namespace, project, jobStatus string, projectParams *release.ProjectParams) (projectCache *release.ProjectCache) {
	projectCache = &release.ProjectCache{
		Namespace:         namespace,
		Name:              project,
		Releases:          []string{},
		InstalledReleases: []string{},
		CreateProjectJobState: release.CreateProjectJobState{
			CreateProjectJobStatus: jobStatus,
		},
	}

	for _, release := range projectParams.Releases {
		projectCache.Releases = append(projectCache.Releases, buildProjectReleaseName(project, release.Name))
	}
	return
}

func setProjectCacheToRedis(redisClient *redis.RedisClient, projectCache *release.ProjectCache) error {
	projectCacheStr, err := json.Marshal(projectCache)
	if err != nil {
		logrus.Errorf("failed to marshal project cache of %s/%s: %s", projectCache.Namespace, projectCache.Name, err.Error())
		return err
	}
	_, err = redisClient.GetClient().HSet(redis.WalmProjectsKey, buildProjectCacheFieldName(projectCache.Namespace, projectCache.Name), projectCacheStr).Result()
	if err != nil {
		logrus.Errorf("failed to set project cache of  %s/%s: %s", projectCache.Namespace, projectCache.Name, err.Error())
		return err
	}
	return nil
}

func getProjectCacheFromRedis(redisClient *redis.RedisClient, namespace, name string) (*release.ProjectCache, error) {
	projectCacheStr, err := redisClient.GetClient().HGet(redis.WalmProjectsKey, buildProjectCacheFieldName(namespace, name)).Result()
	if err != nil {
		if err.Error() == redis.KeyNotFoundErrMsg {
			logrus.Errorf("project cache of %s/%s is not found in redis", namespace, name)
			return nil, walmerr.NotFoundError{}
		}
		logrus.Errorf("failed to get project cache of %s/%s from redis : %s", namespace, name, err.Error())
		return nil, err
	}

	projectCache := &release.ProjectCache{}
	err = json.Unmarshal([]byte(projectCacheStr), projectCache)
	if err != nil {
		logrus.Errorf("failed to unmarshal projectCacheStr %s : %s", projectCacheStr, err.Error())
		return nil, err
	}
	return projectCache, nil
}

func getProjectCachesFromRedis(redisClient *redis.RedisClient, namespace string) ([]*release.ProjectCache, error) {
	filter := namespace + "/*"
	if namespace == "" {
		filter = "*/*"
	}
	scanResult, _, err := redisClient.GetClient().HScan(redis.WalmProjectsKey, 0, filter, 1000).Result()
	if err != nil {
		logrus.Errorf("failed to scan the release caches from redis in namespace %s : %s", namespace, err.Error())
		return nil, err
	}

	projectCacheStrs := []string{}
	for i := 1; i < len(scanResult); i += 2 {
		projectCacheStrs = append(projectCacheStrs, scanResult[i])
	}

	projectCaches := []*release.ProjectCache{}
	for _, projectCacheStr := range projectCacheStrs {
		projectCache := &release.ProjectCache{}
		err = json.Unmarshal([]byte(projectCacheStr), projectCache)
		if err != nil {
			logrus.Errorf("failed to unmarshal projectCacheStr %s : %s", projectCacheStr, err.Error())
			return nil, err
		}
		projectCaches = append(projectCaches, projectCache)
	}

	return projectCaches, nil
}

func deleteProjectCacheFromRedis(redisClient *redis.RedisClient, namespace, name string) (err error) {
	_, err = redisClient.GetClient().HDel(redis.WalmProjectsKey, buildProjectCacheFieldName(namespace, name)).Result()
	if err != nil {
		logrus.Errorf("failed to delete project cache of %s/%s from redis : %s", namespace, name, err.Error())
		return
	}

	return
}

func buildProjectCacheFieldName(namespace, name string) string {
	return namespace + "/" + name
}

func (manager *ProjectManager) DeleteProject(namespace string, project string) error {
	projectCache, err := getProjectCacheFromRedis(manager.redisClient, namespace, project)
	if err != nil {
		if walmerr.IsNotFoundError(err) {
			logrus.Warnf("project %s/%s is not found in redis", namespace, project)
			return nil
		}
		logrus.Errorf("failed to get project cache of %s/%s : %s", namespace, project, err.Error())
		return err
	}

	releaseLeft := []string{}
	mux := &sync.Mutex{}
	var wg sync.WaitGroup
	for _, releaseName := range projectCache.InstalledReleases {
		wg.Add(1)
		go func(releaseName string) {
			defer wg.Done()
			err1 := manager.helmClient.DeleteRelease(namespace, releaseName)
			if err1 != nil {
				if strings.Contains(err.Error(), "not found") {
					return
				}
				logrus.Errorf("failed to delete project release %s : %s", releaseName, err.Error())
				err = errors.New(err1.Error())
				mux.Lock()
				releaseLeft = append(releaseLeft, releaseName)
				mux.Unlock()
			}
		}(releaseName)
	}
	wg.Wait()

	if err != nil {
		logrus.Errorf("failed to delete project releases of %s/%s : %s", namespace, project, err.Error())
		projectCache.InstalledReleases = releaseLeft
		err1 := setProjectCacheToRedis(manager.redisClient, projectCache)
		if err1 != nil {
			logrus.Errorf("failed to update project cache of %s/%s to installed releases v% : %s", namespace, project, projectCache.InstalledReleases, err1.Error())
		}
		return err
	}

	err = deleteProjectCacheFromRedis(manager.redisClient, namespace, project)
	if err != nil {
		logrus.Errorf("failed to delete project cache of %s/%s : %s", namespace, project, err.Error())
		return err
	}
	logrus.Infof("succeed to delete project %s/%s", namespace, project)
	return nil
}

func (manager *ProjectManager) AddReleaseInProject(namespace string, projectName string, releaseParams *release.ReleaseRequest) error {
	projectInfo, err := manager.GetProjectInfo(namespace, projectName)
	if err != nil {
		return err
	}
	if projectInfo == nil {
		err = manager.helmClient.InstallUpgradeRealese(namespace, releaseParams)
		if err != nil {
			logrus.Errorf("AddReleaseInProject install release %s error %v\n", releaseParams.Name, err)
			return err
		}
	}
	return nil
}

func (manager *ProjectManager) RemoveReleaseInProject(namespace string, projectName string, releaseParams *release.ReleaseRequest) error {
	return nil
}

func brainFuckRuntimeDepParse(projectInfo *release.ProjectInfo, releaseParams *release.ReleaseRequest) ([]*release.ReleaseRequest, error) {
	//subCharts, err := helm.GetDependencies(releaseParams.RepoName, releaseParams.ChartName, releaseParams.ChartVersion)
	//if err != nil {
	//	return nil, err
	//}

	// Find Upstream Release
	//for _, chartName := range subCharts {
	//	for _, releaseInfo := range projectInfo.Releases {
	//		releaseSubCharts, err := helm.GetDependencies(releaseInfo.ChartName, releaseInfo.ChartVersion)
	//		if err != nil {
	//			return nil, err
	//		}
	//		logrus.Infof("%s %v", chartName, releaseSubCharts)
	//	}
	//}
	//projectParams := {
	//}
	// Find Downstream Release
	//for _, chartName := range subCharts {
	//}

	return nil, nil
}

func (manager *ProjectManager) brainFuckChartDepParse(projectParams *release.ProjectParams) ([]*release.ReleaseRequest, error) {
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
		subCharts, err := manager.helmClient.GetDependencies(helmRelease.RepoName, helmRelease.ChartName, helmRelease.ChartVersion)
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
		logrus.Debugf("DEBUG: %v", releaseRequest.Dependencies)
		chartsNeighbors := g.Neighbors(sortedChartList[i])
		for _, chartNeighbor := range chartsNeighbors {
			chartName := (*chartNeighbor.Value).(*release.ReleaseRequest).ChartName
			_, ok := releaseRequest.Dependencies[chartName]
			if !ok {
				releaseRequest.Dependencies[chartName] = (*chartNeighbor.Value).(*release.ReleaseRequest).Name
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
