package project

import (
	"strings"
	"sync"
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"

	"walm/pkg/release"
	"walm/pkg/release/manager/helm"
	"walm/pkg/redis"
	"walm/pkg/job"
	"walm/pkg/util/dag"
	walmerr "walm/pkg/util/error"
	"fmt"
	"time"
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
		CommonValues:          projectCache.CommonValues,
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
		CommonValues:      projectParams.CommonValues,
		Releases:          []string{},
		InstalledReleases: []string{},
		CreateProjectJobState: release.CreateProjectJobState{
			CreateProjectJobStatus: jobStatus,
		},
	}

	for _, release := range projectParams.Releases {
		projectCache.Releases = append(projectCache.Releases, buildProjectReleaseName(project, release.Name))
	}
	return projectCache
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
	projectInfo, err := manager.GetProjectInfoSync(namespace, projectName)
	if err != nil {
		return err
	}
	releaseParams.Name = buildProjectReleaseName(projectName, releaseParams.Name)

	if projectInfo != nil {
		releaseParams.ConfigValues = mergeValues(releaseParams.ConfigValues, projectInfo.CommonValues)
	}
	if projectInfo != nil {
		affectReleaseRequest, err2 := manager.brainFuckRuntimeDepParse(projectInfo, releaseParams, false)
		if err2 != nil {
			logrus.Errorf("RuntimeDepParse install release %s error %v\n", releaseParams.Name, err)
			return err2
		}
		err = manager.helmClient.InstallUpgradeRealese(namespace, releaseParams)
		if err != nil {
			logrus.Errorf("AddReleaseInProject install release %s error %v\n", releaseParams.Name, err)
			return err
		}
		for _, affectReleaseParams := range affectReleaseRequest {
			logrus.Infof("Update BecauseOf Dependency Modified: %v", *affectReleaseParams)
			err = manager.helmClient.UpgradeRealese(namespace, affectReleaseParams)
			if err != nil {
				logrus.Errorf("AddReleaseInProject Other Affected Release install release %s error %v\n", releaseParams.Name, err)
				return err
			}
		}
	} else {
		err = manager.helmClient.InstallUpgradeRealese(namespace, releaseParams)
		if err != nil {
			logrus.Errorf("AddReleaseInProject install release %s error %v\n", releaseParams.Name, err)
			return err
		}
	}

	return nil
}

func (manager *ProjectManager) RemoveReleaseInProject(namespace string, projectName, releaseName string) error {
	projectInfo, err := manager.GetProjectInfoSync(namespace, projectName)
	if err != nil {
		return err
	}
	releaseProjectName := buildProjectReleaseName(projectName, releaseName)

	err = manager.helmClient.DeleteRelease(namespace, releaseProjectName)
	if err != nil {
		logrus.Errorf("RemoveReleaseInProject install release %s error %v\n", releaseProjectName, err)
		return err
	}
	releaseParams := buildReleaseRequest(projectInfo, releaseName)
	if releaseParams == nil {
		return fmt.Errorf("can not remove %s in %v", releaseName, *projectInfo)
	}
	if projectInfo != nil {
		affectReleaseRequest, err2 := manager.brainFuckRuntimeDepParse(projectInfo, releaseParams, true)
		if err2 != nil {
			logrus.Errorf("RuntimeDepParse install release %s error %v\n", releaseParams.Name, err)
			return err2
		}
		for _, affectReleaseParams := range affectReleaseRequest {
			logrus.Infof("Update BecauseOf Dependency Modified: %v", *affectReleaseParams)
			err = manager.helmClient.UpgradeRealese(namespace, affectReleaseParams)
			if err != nil {
				logrus.Errorf("RemoveReleaseInProject Other Affected Release install release %s error %v\n", releaseParams.Name, err)
				return err
			}
		}
	}
	manager.helmClient.DeleteRelease(namespace, releaseProjectName)

	return nil
}

func (manager *ProjectManager) brainFuckRuntimeDepParse(projectInfo *release.ProjectInfo, releaseParams *release.ReleaseRequest, isRemove bool) ([]*release.ReleaseRequest, error) {
	var g dag.AcyclicGraph
	affectReleases := make([]*release.ReleaseRequest, 0)

	// init node
	for _, helmRelease := range projectInfo.Releases {
		g.Add(helmRelease.Name)
	}

	// init edge
	for _, helmRelease := range projectInfo.Releases {
		for _, v := range helmRelease.Dependencies {
			g.Connect(dag.BasicEdge(helmRelease.Name, v))
		}
	}

	if !isRemove {
		g.Add(releaseParams.Name)
		for _, helmRelease := range projectInfo.Releases {
			subCharts, err := manager.helmClient.GetDependencies(helmRelease.RepoName, helmRelease.ChartName, helmRelease.ChartVersion)
			if err != nil {
				return nil, err
			}
			for _, subChartName := range subCharts {
				_, ok := helmRelease.Dependencies[subChartName]
				if subChartName == releaseParams.ChartName && !ok {
					g.Connect(dag.BasicEdge(helmRelease.Name, releaseParams.Name))
				}
			}
		}
		releaseSubCharts, err := manager.helmClient.GetDependencies(releaseParams.RepoName, releaseParams.ChartName, releaseParams.ChartVersion)
		if err != nil {
			return nil, err
		}
		for _, releaseSubChart := range releaseSubCharts {
			_, ok := releaseParams.Dependencies[releaseSubChart]
			if ok {
				continue
			}
			for _, helmRelease := range projectInfo.Releases {
				if releaseSubChart == helmRelease.ChartName {
					g.Connect(dag.BasicEdge(releaseParams.Name, helmRelease.Name))
				}
			}
		}
		logrus.Infof("add %s Modify UpperStream Release %+v deps %s\n", releaseParams.Name, g.UpEdges(releaseParams.Name).List(), releaseParams.Name)
		for _, upperReleaseName := range g.UpEdges(releaseParams.Name).List() {
			upperRelease := buildReleaseRequest(projectInfo, upperReleaseName.(string))
			if upperRelease == nil {
				continue
			}
			_, ok := upperRelease.Dependencies[releaseParams.ChartName]
			if !ok {
				upperRelease.Dependencies[releaseParams.ChartName] = releaseParams.Name
			}
			affectReleases = append(affectReleases, upperRelease)
		}
		logrus.Infof("add %s release add more %+v deps. current %+v\n",
			releaseParams.Name, g.DownEdges(releaseParams.Name).List(), releaseParams.Dependencies)
		for _, downReleaseName := range g.DownEdges(releaseParams.Name).List() {
			downRelease := buildReleaseRequest(projectInfo, downReleaseName.(string))
			if downRelease == nil {
				continue
			}
			_, ok := releaseParams.Dependencies[downRelease.ChartName]
			if !ok {
				releaseParams.Dependencies[downRelease.ChartName] = downRelease.Name
				logrus.Infof("RuntimeDepParse release %s Dependencies %+v\n", releaseParams.Name, releaseParams.Dependencies)
			}
		}
	} else {
		logrus.Infof("remove %+v\n", g.UpEdges(releaseParams.Name).List())
		for _, upperReleaseName := range g.UpEdges(releaseParams.Name).List() {
			upperRelease := buildReleaseRequest(projectInfo, upperReleaseName.(string))
			if upperRelease == nil {
				continue
			}
			_, ok := upperRelease.Dependencies[releaseParams.ChartName]
			if ok {
				delete(upperRelease.Dependencies, releaseParams.ChartName)
			}
			affectReleases = append(affectReleases, upperRelease)
		}
	}

	return affectReleases, nil
}

func buildReleaseRequest(projectInfo *release.ProjectInfo, releaseName string) *release.ReleaseRequest {
	var releaseRequest release.ReleaseRequest
	found := false
	for _, releaseInfo := range projectInfo.Releases {
		if releaseInfo.Name != releaseName {
			continue
		}
		releaseRequest.ConfigValues = make(map[string]interface{})
		releaseRequest.ConfigValues["UPDATE"] = time.Now().String()
		releaseRequest.Dependencies = make(map[string]string)
		for k, v := range releaseInfo.Dependencies {
			releaseRequest.Dependencies[k] = v
		}
		releaseRequest.Name = buildProjectReleaseName(projectInfo.Name, releaseInfo.Name)
		releaseRequest.ChartName = releaseInfo.ChartName
		releaseRequest.ChartVersion = releaseInfo.ChartVersion
		found = true
		break
	}

	if !found {
		return nil
	}
	return &releaseRequest
}

func (manager *ProjectManager) brainFuckChartDepParse(projectParams *release.ProjectParams) ([]*release.ReleaseRequest, error) {
	projectParamsMap := make(map[string]*release.ReleaseRequest)
	releaseParsed := make([]*release.ReleaseRequest, 0)
	var g dag.AcyclicGraph

	for _, releaseInfo := range projectParams.Releases {
		projectParamsMap[releaseInfo.ChartName] = releaseInfo
	}

	// init node
	for _, helmRelease := range projectParams.Releases {
		g.Add(helmRelease)
	}

	// init edge
	for _, helmRelease := range projectParams.Releases {
		subCharts, err := manager.helmClient.GetDependencies(helmRelease.RepoName, helmRelease.ChartName, helmRelease.ChartVersion)
		if err != nil {
			return nil, err
		}

		for _, subChartName := range subCharts {
			_, ok := projectParamsMap[subChartName]
			_, ok2 := helmRelease.Dependencies[subChartName]
			if ok && !ok2 {
				g.Connect(dag.BasicEdge(helmRelease, projectParamsMap[subChartName]))
			}
		}
	}

	_, err := g.Root()
	if err != nil {
		return nil, err
	}

	var lock sync.Mutex
	err = g.Walk(func(v dag.Vertex) error {
		lock.Lock()
		defer lock.Unlock()
		releaseRequest := v.(*release.ReleaseRequest)
		for _, dv := range g.DownEdges(releaseRequest).List() {
			release := dv.(*release.ReleaseRequest)
			releaseRequest.Dependencies[release.ChartName] = release.Name
		}
		releaseParsed = append(releaseParsed, releaseRequest)
		return nil
	})
	if err != nil {
		return nil, err
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

func CreateProjectSync(tenantName, projectName string, projectParams *release.ProjectParams) error {
	helmExtraLabelsBase := map[string]interface{}{}
	helmExtraLabelsVals := release.HelmExtraLabels{}
	helmExtraLabelsVals.HelmLabels = make(map[string]interface{})
	helmExtraLabelsVals.HelmLabels["project_name"] = projectName
	helmExtraLabelsBase["HelmExtraLabels"] = helmExtraLabelsVals

	rawValsBase := map[string]interface{}{}
	rawValsBase = mergeValues(rawValsBase, projectParams.CommonValues)
	rawValsBase = mergeValues(helmExtraLabelsBase, rawValsBase)

	for _, releaseParams := range projectParams.Releases {
		releaseParams.Name = buildProjectReleaseName(projectName, releaseParams.Name)
		releaseParams.ConfigValues = mergeValues(releaseParams.ConfigValues, rawValsBase)
	}

	releaseList, err := GetDefaultProjectManager().brainFuckChartDepParse(projectParams)
	if err != nil {
		logrus.Errorf("failed to parse project charts dependency relation  : %s", err.Error())
		return err
	}
	for _, releaseParams := range releaseList {
		err = GetDefaultProjectManager().helmClient.InstallUpgradeRealese(tenantName, releaseParams)
		if err != nil {
			logrus.Errorf("failed to create project release %s/%s : %s", tenantName, releaseParams.Name, err)
			return err
		}
	}
	return nil
}

func (manager *ProjectManager) ListProjectsSync(namespace string) (*release.ProjectInfoList, error) {
	projectMap := make(map[string]*release.ProjectInfo)
	projectList := new(release.ProjectInfoList)

	releaseList, err := manager.helmClient.ListReleases(namespace, "*--*")
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
				if !releaseInfo.Ready {
					projectInfo.Ready = releaseInfo.Ready
				}
				projectInfo.Releases = append(projectInfo.Releases, releaseInfo)
			} else {
				projectMap[projectName] = new(release.ProjectInfo)
				projectInfo.Ready = true
				projectMap[projectName].Name = projectName
				projectMap[projectName].Namespace = releaseInfo.Namespace
				projectMap[projectName].CommonValues = make(map[string]interface{})
				releaseInfo.Name = projectNameArray[1]
				if !releaseInfo.Ready {
					projectInfo.Ready = releaseInfo.Ready
				}
				projectMap[projectName].Releases = append(projectMap[projectName].Releases, releaseInfo)
				projectList.Items = append(projectList.Items, projectMap[projectName])
			}
		}
	}
	return projectList, nil
}

func (manager *ProjectManager) GetProjectInfoSync(namespace, projectName string) (*release.ProjectInfo, error) {
	projectInfo := &release.ProjectInfo{
		Name: projectName,
		Namespace: namespace,
		CommonValues: map[string]interface{}{},
		Ready: true,
	}
	releaseList, err := manager.helmClient.ListReleases(namespace, projectName + "--*")
	if err != nil {
		return nil, err
	}
	for _, releaseInfo := range releaseList {
		projectNameArray := strings.Split(releaseInfo.Name, "--")
		if len(projectNameArray) == 2 {
			if projectName == projectNameArray[0] {
				releaseInfo.Name = projectNameArray[1]
				projectInfo.Releases = append(projectInfo.Releases, releaseInfo)
				if !releaseInfo.Ready {
					projectInfo.Ready = false
				}
			}
		}
	}
	if len(projectInfo.Releases) > 0 {
		return projectInfo, nil
	}
	return nil, nil
}

func (manager *ProjectManager) DeleteProjectSync(namespace string, project string) error {
	projectInfo, err := manager.GetProjectInfoSync(namespace, project)
	if err != nil {
		logrus.Errorf("DeleteProject get project info error %v\n", err)
		return err
	}
	if projectInfo == nil && err == nil {
		logrus.Infof("DeleteProject can't found project %s %s", namespace, project)
	}
	for _, releaseInfo := range projectInfo.Releases {
		releaseName := fmt.Sprintf("%s--%s", projectInfo.Name, releaseInfo.Name)
		err = manager.helmClient.DeleteRelease(namespace, releaseName)
		if err != nil {
			logrus.Errorf("DeleteProject deleteRelease %s info error %v\n", releaseName, err)
		}
	}
	return nil
}

func (manager *ProjectManager) AddProjectInProject(namespace string, projectName string, projectParams *release.ProjectParams) error {
	projectInfo, err := manager.GetProjectInfoSync(namespace, projectName)
	if err != nil {
		return err
	}
	releaseList, err := GetDefaultProjectManager().brainFuckChartDepParse(projectParams)
	if err != nil {
		logrus.Errorf("failed to parse project charts dependency relation  : %s", err.Error())
		return err
	}
	for _, releaseParams := range releaseList {
		releaseParams.Name = buildProjectReleaseName(projectName, releaseParams.Name)
		releaseParams.ConfigValues = mergeValues(releaseParams.ConfigValues, projectParams.CommonValues)

		if projectInfo != nil {
			affectReleaseRequest, err2 := manager.brainFuckRuntimeDepParse(projectInfo, releaseParams, false)
			if err2 != nil {
				logrus.Errorf("RuntimeDepParse install release %s error %v\n", releaseParams.Name, err)
				return err2
			}
			err = manager.helmClient.InstallUpgradeRealese(namespace, releaseParams)
			if err != nil {
				logrus.Errorf("AddReleaseInProject install release %s error %v\n", releaseParams.Name, err)
				return err
			}
			for _, affectReleaseParams := range affectReleaseRequest {
				logrus.Infof("Update BecauseOf Dependency Modified: %v", *affectReleaseParams)
				err = manager.helmClient.UpgradeRealese(namespace, affectReleaseParams)
				if err != nil {
					logrus.Errorf("AddReleaseInProject Other Affected Release install release %s error %v\n", releaseParams.Name, err)
					return err
				}
			}
		} else {
			err = manager.helmClient.InstallUpgradeRealese(namespace, releaseParams)
			if err != nil {
				logrus.Errorf("AddReleaseInProject install release %s error %v\n", releaseParams.Name, err)
				return err
			}
		}
	}

	return nil
}