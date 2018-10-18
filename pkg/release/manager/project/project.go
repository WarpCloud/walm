package project

import (
	"strings"
	"sync"
	"errors"
	"github.com/sirupsen/logrus"

	"walm/pkg/release"
	"walm/pkg/release/manager/helm"
	"walm/pkg/redis"
	"walm/pkg/job"
	"walm/pkg/util/dag"
	walmerr "walm/pkg/util/error"
	"fmt"
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
	projectCaches, err := manager.helmClient.GetHelmCache().GetProjectCaches(namespace)
	if err != nil {
		logrus.Errorf("failed to get project caches in namespace %s : %s", namespace, err.Error())
		return nil, err
	}

	projectInfoList := &release.ProjectInfoList{
		Items: []*release.ProjectInfo{},
	}

	mux := &sync.Mutex{}
	var wg sync.WaitGroup
	for _, projectCache := range projectCaches {
		wg.Add(1)
		go func(projectCache *release.ProjectCache) {
			defer wg.Done()
			projectInfo, err1 := manager.buildProjectInfo(projectCache)
			if err1 != nil {
				logrus.Errorf("failed to build project info from project cache of %s/%s : %s", projectCache.Namespace, projectCache.Name, err1.Error())
				err = errors.New(err1.Error())
				return
			}
			mux.Lock()
			projectInfoList.Items = append(projectInfoList.Items, projectInfo)
			mux.Unlock()
		}(projectCache)
	}

	wg.Wait()
	if err != nil {
		logrus.Errorf("failed to build project infos : %s", err.Error())
		return nil, err
	}

	projectInfoList.Num = len(projectInfoList.Items)
	return projectInfoList, nil
}

func (manager *ProjectManager) GetProjectInfo(namespace, projectName string) (*release.ProjectInfo, error) {
	projectCache, err := manager.helmClient.GetHelmCache().GetProjectCache(namespace, projectName)
	if err != nil {
		logrus.Errorf("failed to get project cache of %s/%s : %s", namespace, projectName, err.Error())
		return nil, err
	}

	return manager.buildProjectInfo(projectCache)
}

func (manager *ProjectManager) buildProjectInfo(projectCache *release.ProjectCache) (projectInfo *release.ProjectInfo, err error) {
	projectInfo = &release.ProjectInfo{
		ProjectCache: *projectCache,
		Releases:     []*release.ReleaseInfo{},
	}

	if projectInfo.LatestProjectJobState.Status == "Pending" {
		return
	}

	releaseList, err := manager.helmClient.ListReleases(projectCache.Namespace, projectCache.Name+"--*")
	if err != nil {
		return nil, err
	}

	for _, releaseInfo := range releaseList {
		projectNameArray := strings.Split(releaseInfo.Name, "--")
		if len(projectNameArray) == 2 {
			if projectInfo.Name == projectNameArray[0] {
				releaseInfo.Name = projectNameArray[1]
				projectInfo.Releases = append(projectInfo.Releases, releaseInfo)
			}
		}
	}

	if projectInfo.LatestProjectJobState.Status == "Succeed" {
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

func (manager *ProjectManager) validateProjectJob(namespace, name string, allowProjectNotExist bool) (projectCache *release.ProjectCache, err error) {
	projectCache, err = manager.helmClient.GetHelmCache().GetProjectCache(namespace, name)
	if err != nil {
		if !walmerr.IsNotFoundError(err) {
			logrus.Errorf("failed to get project cache : %s", err.Error())
			return
		} else if !allowProjectNotExist {
			return
		} else {
			err = nil
		}
	} else {
		if projectCache.IsProjectJobNotFinished() {
			err = fmt.Errorf("please wait for the project's %s job ending", projectCache.LatestProjectJobState.Type)
			logrus.Error(err.Error())
			return
		}
	}
	return
}

func (manager *ProjectManager) CreateProject(namespace string, project string, projectParams *release.ProjectParams, async bool) error {
	if len(projectParams.Releases) == 0 {
		return errors.New("project releases can not be empty")
	}

	_, err := manager.validateProjectJob(namespace, project, true)
	if err != nil {
		logrus.Errorf("failed to validate project job : %s", err.Error())
		return err
	}

	createProjectJob := &CreateProjectJob{
		Namespace:     namespace,
		Name:          project,
		Async:         async,
		ProjectParams: projectParams,
	}

	projectCache := buildProjectCache(namespace, project, createProjectJob.Type(), "Pending", async)
	err = manager.helmClient.GetHelmCache().CreateOrUpdateProjectCache(projectCache)
	if err != nil {
		logrus.Errorf("failed to set project cache of %s/%s to redis: %s", namespace, project, err.Error())
		return err
	}

	if async {
		jobId, err := job.GetDefaultWalmJobManager().CreateWalmJob("", createProjectJob)
		if err != nil {
			logrus.Errorf("failed to create Async %s Job : %s", createProjectJob.Type(), err.Error())
			return err
		}
		logrus.Infof("succeed to create Async %s Job %s", createProjectJob.Type(), jobId)
	} else {
		err = createProjectJob.Do()
		if err != nil {
			return err
		}
	}

	return nil
}

func (manager *ProjectManager) DeleteProject(namespace string, project string, async bool) error {
	_, err := manager.validateProjectJob(namespace, project, false)
	if err != nil {
		if walmerr.IsNotFoundError(err) {
			logrus.Warnf("project %s/%s is not found", namespace, project)
			return nil
		}
		logrus.Errorf("failed to validate project job : %s", err.Error())
		return err
	}

	deleteProjectJob := &DeleteProjectJob{
		Namespace: namespace,
		Name:      project,
		Async:     async,
	}

	projectCache := buildProjectCache(namespace, project, deleteProjectJob.Type(), "Pending", async)
	err = manager.helmClient.GetHelmCache().CreateOrUpdateProjectCache(projectCache)
	if err != nil {
		logrus.Errorf("failed to set project cache of %s/%s to redis: %s", namespace, project, err.Error())
		return err
	}

	if async {
		jobId, err := job.GetDefaultWalmJobManager().CreateWalmJob("", deleteProjectJob)
		if err != nil {
			logrus.Errorf("failed to create Async %s Job : %s", deleteProjectJob.Type(), err.Error())
			return err
		}
		logrus.Infof("succeed to create Async %s Job %s", deleteProjectJob.Type(), jobId)
	} else {
		err = deleteProjectJob.Do()
		if err != nil {
			return err
		}
	}

	return nil
}

func (manager *ProjectManager) AddReleaseInProject(namespace string, projectName string, releaseParams *release.ReleaseRequest, async bool) error {
	return manager.AddReleasesInProject(namespace, projectName, &release.ProjectParams{Releases: []*release.ReleaseRequest{releaseParams}}, async)
}

func (manager *ProjectManager) RemoveReleaseInProject(namespace, projectName, releaseName string, async bool) error {
	projectCache, err := manager.validateProjectJob(namespace, projectName, false)
	if err != nil {
		if walmerr.IsNotFoundError(err) {
			logrus.Warnf("project %s/%s is not found", namespace, projectName)
			return nil
		}
		logrus.Errorf("failed to validate project job : %s", err.Error())
		return err
	}

	projectInfo, err := manager.buildProjectInfo(projectCache)
	if err != nil {
		logrus.Errorf("failed to build project info : %s", err.Error())
		return err
	}

	releaseExistsInProject := false
	for _, release := range projectInfo.Releases {
		if release.Name == releaseName {
			releaseExistsInProject = true
			break
		}
	}

	if !releaseExistsInProject {
		logrus.Warnf("release %s is not found in project %s", releaseName, projectName)
		return nil
	}

	removeReleaseJob := &RemoveReleaseJob{
		Async:       async,
		Namespace:   namespace,
		Name:        projectName,
		ReleaseName: releaseName,
	}

	projectCache = buildProjectCache(namespace, projectName, removeReleaseJob.Type(), "Pending", async)
	err = manager.helmClient.GetHelmCache().CreateOrUpdateProjectCache(projectCache)
	if err != nil {
		logrus.Errorf("failed to set project cache of %s/%s to redis: %s", namespace, projectName, err.Error())
		return err
	}

	if async {
		jobId, err := job.GetDefaultWalmJobManager().CreateWalmJob("", removeReleaseJob)
		if err != nil {
			logrus.Errorf("failed to create Async %s Job : %s", removeReleaseJob.Type(), err.Error())
			return err
		}
		logrus.Infof("succeed to create Async %s Job %s", removeReleaseJob.Type(), jobId)
	} else {
		err = removeReleaseJob.Do()
		if err != nil {
			return err
		}
	}

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

func (manager *ProjectManager) AddReleasesInProject(namespace string, projectName string, projectParams *release.ProjectParams, async bool) error {
	if len(projectParams.Releases) == 0 {
		return errors.New("project releases can not be empty")
	}

	_, err := manager.validateProjectJob(namespace, projectName, true)
	if err != nil {
		logrus.Errorf("failed to validate project job : %s", err.Error())
		return err
	}

	addReleasesJob := &AddReleasesJob{
		Async:         async,
		Namespace:     namespace,
		Name:          projectName,
		ProjectParams: projectParams,
	}

	projectCache := buildProjectCache(namespace, projectName, addReleasesJob.Type(), "Pending", async)
	err = manager.helmClient.GetHelmCache().CreateOrUpdateProjectCache(projectCache)
	if err != nil {
		logrus.Errorf("failed to set project cache of %s/%s to redis: %s", namespace, projectName, err.Error())
		return err
	}

	if async {
		jobId, err := job.GetDefaultWalmJobManager().CreateWalmJob("", addReleasesJob)
		if err != nil {
			logrus.Errorf("failed to create Async %s Job : %s", addReleasesJob.Type(), err.Error())
			return err
		}
		logrus.Infof("succeed to create Async %s Job %s", addReleasesJob.Type(), jobId)
	} else {
		err = addReleasesJob.Do()
		if err != nil {
			return err
		}
	}

	return nil
}
