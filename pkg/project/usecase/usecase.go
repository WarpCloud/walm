package usecase

import (
	projectModel "WarpCloud/walm/pkg/models/project"
	releaseModel "WarpCloud/walm/pkg/models/release"
	"github.com/sirupsen/logrus"
	"WarpCloud/walm/pkg/project"
	"WarpCloud/walm/pkg/task"
	"WarpCloud/walm/pkg/release"
	errorModel "WarpCloud/walm/pkg/models/error"
	"fmt"
	"sync"
	"errors"
	"encoding/json"
	"WarpCloud/walm/pkg/util/dag"
	"WarpCloud/walm/pkg/helm"
)

const (
	defaultSleepTimeSecond int64 = 1
	defaultTimeoutSec      int64 = 60 * 10
)

type Project struct {
	cache          project.Cache
	task           task.Task
	releaseUseCase release.UseCase
	helm           helm.Helm
}

func (projectImpl *Project) ListProjects(namespace string) (*projectModel.ProjectInfoList, error) {
	projectTasks, err := projectImpl.cache.GetProjectTasks(namespace)
	if err != nil {
		logrus.Errorf("failed to get project tasks in namespace %s : %s", namespace, err.Error())
		return nil, err
	}

	projectInfoList := &projectModel.ProjectInfoList{
		Items: []*projectModel.ProjectInfo{},
	}

	mux := &sync.Mutex{}
	var wg sync.WaitGroup
	for _, projectTask := range projectTasks {
		wg.Add(1)
		go func(projectTask *projectModel.ProjectTask) {
			defer wg.Done()
			projectInfo, err1 := projectImpl.buildProjectInfo(projectTask)
			if err1 != nil {
				logrus.Errorf("failed to build project info from project cache of %s/%s : %s", projectTask.Namespace, projectTask.Name, err1.Error())
				err = errors.New(err1.Error())
				return
			}
			mux.Lock()
			projectInfoList.Items = append(projectInfoList.Items, projectInfo)
			mux.Unlock()
		}(projectTask)
	}

	wg.Wait()
	if err != nil {
		logrus.Errorf("failed to build project infos : %s", err.Error())
		return nil, err
	}

	projectInfoList.Num = len(projectInfoList.Items)
	return projectInfoList, nil
}

func (projectImpl *Project) GetProjectInfo(namespace, projectName string) (*projectModel.ProjectInfo, error) {
	projectTask, err := projectImpl.cache.GetProjectTask(namespace, projectName)
	if err != nil {
		logrus.Errorf("failed to get project task of %s/%s : %s", namespace, projectName, err.Error())
		return nil, err
	}

	return projectImpl.buildProjectInfo(projectTask)
}

func (projectImpl *Project) CreateProject(namespace string, project string, projectParams *projectModel.ProjectParams, async bool, timeoutSec int64) error {
	if len(projectParams.Releases) == 0 {
		return errors.New("project releases can not be empty")
	}

	if timeoutSec == 0 {
		timeoutSec = defaultTimeoutSec
	}

	oldProjectTask, err := projectImpl.validateProjectTask(namespace, project, true)
	if err != nil {
		logrus.Errorf("failed to validate project task : %s", err.Error())
		return err
	}

	createProjectTaskArgs := &CreateProjectTaskArgs{
		Name:          project,
		Namespace:     namespace,
		ProjectParams: projectParams,
	}
	err = projectImpl.sendProjectTask(namespace, project, createProjectTaskName, createProjectTaskArgs, oldProjectTask, timeoutSec, async)
	if err != nil {
		logrus.Errorf("failed to send project task %s of %s/%s : %s", createProjectTaskName, namespace, project, err.Error())
		return err
	}
	logrus.Infof("succeed to create project %s/%s", namespace, project)
	return nil
}

func (projectImpl *Project) validateProjectTask(namespace, name string, allowProjectNotExist bool) (projectTask *projectModel.ProjectTask, err error) {
	projectTask, err = projectImpl.cache.GetProjectTask(namespace, name)
	if err != nil {
		if !errorModel.IsNotFoundError(err) {
			logrus.Errorf("failed to get project task : %s", err.Error())
			return
		} else if !allowProjectNotExist {
			return
		} else {
			err = nil
		}
	} else {
		taskState, err := projectImpl.task.GetTaskState(projectTask.LatestTaskSignature)
		if err != nil {
			if errorModel.IsNotFoundError(err) {
				err = nil
				return projectTask, err
			} else {
				logrus.Errorf("failed to get the last project task state : %s", err.Error())
				return projectTask, err
			}
		}

		if !(taskState.IsFinished() || taskState.IsTimeout()) {
			err = fmt.Errorf("please wait for the last project task %s-%s finished or timeout", projectTask.LatestTaskSignature.Name, projectTask.LatestTaskSignature.UUID)
			logrus.Warn(err.Error())
			return projectTask, err
		}
	}
	return
}

func (projectImpl *Project) DeleteProject(namespace string, project string, async bool, timeoutSec int64, deletePvcs bool) error {
	oldProjectTask, err := projectImpl.validateProjectTask(namespace, project, false)
	if err != nil {
		if errorModel.IsNotFoundError(err) {
			logrus.Warnf("project %s/%s is not found", namespace, project)
			return nil
		}
		logrus.Errorf("failed to validate project job : %s", err.Error())
		return err
	}

	if timeoutSec == 0 {
		timeoutSec = defaultTimeoutSec
	}

	deleteProjectTaskArgs := &DeleteProjectTaskArgs{
		Name:       project,
		Namespace:  namespace,
		DeletePvcs: deletePvcs,
	}

	err = projectImpl.sendProjectTask(namespace, project, deleteProjectTaskName, deleteProjectTaskArgs, oldProjectTask, timeoutSec, async)
	if err != nil {
		logrus.Errorf("failed to send project task %s of %s/%s : %s", deleteProjectTaskName, namespace, project, err.Error())
		return err
	}
	logrus.Infof("succeed to delete project %s/%s", namespace, project)

	return nil
}
func (projectImpl *Project) AddReleasesInProject(namespace string, projectName string,
	projectParams *projectModel.ProjectParams, async bool, timeoutSec int64) error {

	if len(projectParams.Releases) == 0 {
		return errors.New("project releases can not be empty")
	}

	oldProjectTask, err := projectImpl.validateProjectTask(namespace, projectName, true)
	if err != nil {
		logrus.Errorf("failed to validate project job : %s", err.Error())
		return err
	}

	if timeoutSec == 0 {
		timeoutSec = defaultTimeoutSec
	}

	taskArgs := &AddReleaseTaskArgs{
		Name:          projectName,
		Namespace:     namespace,
		ProjectParams: projectParams,
	}

	err = projectImpl.sendProjectTask(namespace, projectName, addReleaseTaskName, taskArgs, oldProjectTask, timeoutSec, async)
	if err != nil {
		logrus.Errorf("failed to send project task %s of %s/%s : %s", addReleaseTaskName, namespace, projectName, err.Error())
		return err
	}
	logrus.Infof("succeed to add releases in project %s/%s", namespace, projectName)

	return nil
}

func (projectImpl *Project) UpgradeReleaseInProject(namespace string, projectName string,
	releaseParams *releaseModel.ReleaseRequestV2, async bool, timeoutSec int64) error {
	oldProjectTask, err := projectImpl.validateProjectTask(namespace, projectName, false)
	if err != nil {
		if errorModel.IsNotFoundError(err) {
			logrus.Warnf("project %s/%s is not found", namespace, projectName)
			return nil
		}
		logrus.Errorf("failed to validate project job : %s", err.Error())
		return err
	}

	projectInfo, err := projectImpl.buildProjectInfo(oldProjectTask)
	if err != nil {
		logrus.Errorf("failed to build project info : %s", err.Error())
		return err
	}

	releaseExistsInProject := false
	for _, releaseInfo := range projectInfo.Releases {
		if releaseInfo.Name == releaseParams.Name {
			releaseExistsInProject = true
			break
		}
	}

	if !releaseExistsInProject {
		err = fmt.Errorf("release %s is not found in project %s", releaseParams.Name, projectName)
		logrus.Error(err.Error())
		return err
	}

	if timeoutSec == 0 {
		timeoutSec = defaultTimeoutSec
	}

	taskArgs := &UpgradeReleaseTaskArgs{
		ProjectName:   projectName,
		Namespace:     namespace,
		ReleaseParams: releaseParams,
	}

	err = projectImpl.sendProjectTask(namespace, projectName, upgradeReleaseTaskName, taskArgs, oldProjectTask, timeoutSec, async)
	if err != nil {
		logrus.Errorf("failed to send project task %s of %s/%s : %s", upgradeReleaseTaskName, namespace, projectName, err.Error())
		return err
	}
	logrus.Infof("succeed to upgrade release %s in project %s/%s", releaseParams.Name, namespace, projectName)

	return nil
}

func (projectImpl *Project) RemoveReleaseInProject(namespace, projectName,
releaseName string, async bool, timeoutSec int64, deletePvcs bool) error {
	oldProjectTask, err := projectImpl.validateProjectTask(namespace, projectName, false)
	if err != nil {
		if errorModel.IsNotFoundError(err) {
			logrus.Warnf("project %s/%s is not found", namespace, projectName)
			return nil
		}
		logrus.Errorf("failed to validate project job : %s", err.Error())
		return err
	}

	projectInfo, err := projectImpl.buildProjectInfo(oldProjectTask)
	if err != nil {
		logrus.Errorf("failed to build project info : %s", err.Error())
		return err
	}

	releaseExistsInProject := false
	for _, releaseInfo := range projectInfo.Releases {
		if releaseInfo.Name == releaseName {
			releaseExistsInProject = true
			break
		}
	}

	if !releaseExistsInProject {
		logrus.Warnf("release %s is not found in project %s", releaseName, projectName)
		return nil
	}

	if timeoutSec == 0 {
		timeoutSec = defaultTimeoutSec
	}

	taskArgs := &RemoveReleaseTaskArgs{
		Name:        projectName,
		Namespace:   namespace,
		ReleaseName: releaseName,
		DeletePvcs:  deletePvcs,
	}

	err = projectImpl.sendProjectTask(namespace, projectName, removeReleaseTaskName, taskArgs, oldProjectTask, timeoutSec, async)
	if err != nil {
		logrus.Errorf("failed to send project task %s of %s/%s : %s", removeReleaseTaskName, namespace, projectName, err.Error())
		return err
	}
	logrus.Infof("succeed to remove release %s in project %s/%s", releaseName, namespace, projectName)

	return nil
}

func (projectImpl *Project) buildProjectInfo(task *projectModel.ProjectTask) (projectInfo *projectModel.ProjectInfo, err error) {
	projectInfo = &projectModel.ProjectInfo{
		Namespace: task.Namespace,
		Name:      task.Name,
		Releases:  []*releaseModel.ReleaseInfoV2{},
	}

	projectInfo.Releases, err = projectImpl.releaseUseCase.ListReleasesByLabels(task.Namespace, projectModel.ProjectNameLabelKey+"="+task.Name)
	if err != nil {
		return nil, err
	}

	taskState, err := projectImpl.task.GetTaskState(task.LatestTaskSignature)
	if err != nil {
		if errorModel.IsNotFoundError(err) {
			err = nil
			projectInfo.Ready, projectInfo.Message = isProjectReadyByReleases(projectInfo.Releases)
			return
		}
		logrus.Errorf("failed to get task state : %s", err.Error())
		return
	}

	if taskState.IsFinished() {
		if taskState.IsSuccess() {
			projectInfo.Ready, projectInfo.Message = isProjectReadyByReleases(projectInfo.Releases)
		} else {
			projectInfo.Message = fmt.Sprintf("the project latest task %s-%s failed : %s", task.LatestTaskSignature.Name, task.LatestTaskSignature.UUID, taskState.GetErrorMsg())
		}
	} else {
		projectInfo.Message = fmt.Sprintf("please wait for the project latest task %s-%s finished", task.LatestTaskSignature.Name, task.LatestTaskSignature.UUID)
	}

	return
}

func isProjectReadyByReleases(releases []*releaseModel.ReleaseInfoV2) (ready bool, message string) {
	if len(releases) > 0 {
		ready = true
		for _, releaseInfo := range releases {
			if !releaseInfo.Ready {
				ready = false
				message = releaseInfo.Message
				break
			}
		}
	} else {
		message = "no release can be found"
	}
	return
}

func (projectImpl *Project) sendProjectTask(namespace, projectName, taskName string, taskArgs interface{},
	oldProjectTask *projectModel.ProjectTask, timeoutSec int64, async bool) (error) {

	taskArgsStr, err := json.Marshal(taskArgs)
	if err != nil {
		logrus.Errorf("failed to marshal task args : %s", err.Error())
		return err
	}

	taskSig, err := projectImpl.task.SendTask(taskName, string(taskArgsStr), timeoutSec)
	if err != nil {
		logrus.Errorf("failed to send %s : %s", taskName, err.Error())
		return err
	}

	projectTask := &projectModel.ProjectTask{
		Namespace:           namespace,
		Name:                projectName,
		LatestTaskSignature: taskSig,
	}

	err = projectImpl.cache.CreateOrUpdateProjectTask(projectTask)
	if err != nil {
		logrus.Errorf("failed to set project task of %s/%s to redis: %s", namespace, projectName, err.Error())
		return err
	}

	if oldProjectTask != nil && oldProjectTask.LatestTaskSignature != nil {
		_ = projectImpl.task.PurgeTaskState(oldProjectTask.LatestTaskSignature)
	}

	if !async {
		err = projectImpl.task.TouchTask(taskSig, defaultSleepTimeSecond)
		if err != nil {
			logrus.Errorf("project task %s of %s/%s is failed or timeout: %s", taskName, namespace, projectName, err.Error())
			return err
		}
	}

	return nil
}

func (projectImpl *Project) autoCreateReleaseDependencies(projectParams *projectModel.ProjectParams) ([]*releaseModel.ReleaseRequestV2, error) {
	projectParamsMap := make(map[string]*releaseModel.ReleaseRequestV2)
	releaseParsed := make([]*releaseModel.ReleaseRequestV2, 0)
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
		subCharts, err := projectImpl.helm.GetChartAutoDependencies(helmRelease.RepoName, helmRelease.ChartName, helmRelease.ChartVersion)
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
		releaseRequest := v.(*releaseModel.ReleaseRequestV2)
		for _, dv := range g.DownEdges(releaseRequest).List() {
			releaseInfo := dv.(*releaseModel.ReleaseRequestV2)
			releaseRequest.Dependencies[releaseInfo.ChartName] = releaseInfo.Name
		}
		releaseParsed = append(releaseParsed, releaseRequest)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return releaseParsed, nil
}

func (projectImpl *Project) autoUpdateReleaseDependencies(projectInfo *projectModel.ProjectInfo, releaseParams *releaseModel.ReleaseRequestV2, isRemove bool) ([]*releaseModel.ReleaseRequestV2, error) {
	var g dag.AcyclicGraph
	affectReleases := make([]*releaseModel.ReleaseRequestV2, 0)

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
			subCharts, err := projectImpl.helm.GetChartAutoDependencies(helmRelease.RepoName, helmRelease.ChartName, helmRelease.ChartVersion)
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
		releaseSubCharts, err := projectImpl.helm.GetChartAutoDependencies(releaseParams.RepoName, releaseParams.ChartName, releaseParams.ChartVersion)
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
				if upperRelease.Dependencies == nil {
					upperRelease.Dependencies = map[string]string{}
				}
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
				if releaseParams.Dependencies == nil {
					releaseParams.Dependencies = make(map[string]string)
				}
				releaseParams.Dependencies[downRelease.ChartName] = downRelease.Name
				logrus.Infof("RuntimeDepParse release %s Dependencies %+v\n", releaseParams.Name, releaseParams.Dependencies)
			}
		}
	} else {
		logrus.Infof("%+v are depending on %s", g.UpEdges(releaseParams.Name).List(), releaseParams.Name)
		for _, upperReleaseName := range g.UpEdges(releaseParams.Name).List() {
			upperRelease := buildReleaseRequest(projectInfo, upperReleaseName.(string))
			if upperRelease == nil {
				continue
			}

			deleteReleaseDependency(upperRelease.Dependencies, releaseParams.ChartName)
			affectReleases = append(affectReleases, upperRelease)
		}
	}

	return affectReleases, nil
}

func buildReleaseRequest(projectInfo *projectModel.ProjectInfo, releaseName string) (releaseRequest *releaseModel.ReleaseRequestV2) {
	for _, releaseInfo := range projectInfo.Releases {
		if releaseInfo.Name == releaseName {
			releaseRequest = releaseInfo.BuildReleaseRequestV2()
			break
		}
	}

	return
}

func deleteReleaseDependency(dependencies map[string]string, dependencyKey string) {
	if _, ok := dependencies[dependencyKey]; ok {
		dependencies[dependencyKey] = ""
	}
}

func NewProject(cache project.Cache, task task.Task, releaseUseCase release.UseCase, helm helm.Helm) (*Project, error) {
	p := &Project{
		cache:          cache,
		task:           task,
		releaseUseCase: releaseUseCase,
		helm:           helm,
	}
	err := p.registerAddReleaseTask()
	if err != nil {
		return nil, err
	}
	err = p.registerCreateProjectTask()
	if err != nil {
		return nil, err
	}
	err = p.registerDeleteProjectTask()
	if err != nil {
		return nil, err
	}
	err = p.registerRemoveReleaseTask()
	if err != nil {
		return nil, err
	}
	err = p.registerUpgradeReleaseTask()
	if err != nil {
		return nil, err
	}
	return p, nil
}
