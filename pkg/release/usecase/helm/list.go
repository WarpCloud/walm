package helm

import (
	releaseModel "WarpCloud/walm/pkg/models/release"
	"github.com/sirupsen/logrus"
	"fmt"
	k8sModel "WarpCloud/walm/pkg/models/k8s"
	errorModel "WarpCloud/walm/pkg/models/error"
	"sync"
	"errors"
	walmHelm "WarpCloud/walm/pkg/helm"
)

func (helm *Helm) GetRelease(namespace, name string) (releaseV2 *releaseModel.ReleaseInfoV2, err error) {
	releaseTask, err := helm.releaseCache.GetReleaseTask(namespace, name)
	if err != nil {
		return nil, err
	}

	return helm.buildReleaseInfoV2ByReleaseTask(releaseTask, nil)
}

func (helm *Helm) buildReleaseInfoV2ByReleaseTask(releaseTask *releaseModel.ReleaseTask, releaseCache *releaseModel.ReleaseCache) (releaseV2 *releaseModel.ReleaseInfoV2, err error) {
	releaseV2 = &releaseModel.ReleaseInfoV2{
		ReleaseInfo: releaseModel.ReleaseInfo{
			ReleaseSpec: releaseModel.ReleaseSpec{
				Namespace: releaseTask.Namespace,
				Name:      releaseTask.Name,
			},
		},
	}

	if releaseCache == nil {
		releaseCache, err = helm.releaseCache.GetReleaseCache(releaseTask.Namespace, releaseTask.Name)
		if err != nil {
			if errorModel.IsNotFoundError(err) {
				logrus.Warnf("release cache %s/%s is not found in redis", releaseTask.Namespace, releaseTask.Name)
				err = nil
			} else {
				logrus.Errorf("failed to get release cache of %s/%s : %s", releaseTask.Namespace, releaseTask.Name, err.Error())
				return
			}
		}
	}

	if releaseCache != nil {
		releaseV2, err = helm.buildReleaseInfoV2(releaseCache)
		if err != nil {
			logrus.Errorf("failed to build v2 release info : %s", err.Error())
			return nil, err
		}
	}

	taskState, err := helm.task.GetTaskState(releaseTask.LatestReleaseTaskSig)
	if err != nil {
		if errorModel.IsNotFoundError(err) {
			err = nil
			return
		} else {
			logrus.Errorf("failed to get task state : %s", err.Error())
			return
		}
	}

	if taskState.IsFinished() {
		if !taskState.IsSuccess() {
			releaseV2.Message = fmt.Sprintf("the release latest task %s-%s failed : %s", releaseTask.LatestReleaseTaskSig.Name, releaseTask.LatestReleaseTaskSig.UUID, taskState.GetErrorMsg())
		}
	} else {
		releaseV2.Message = fmt.Sprintf("please wait for the release latest task %s-%s finished", releaseTask.LatestReleaseTaskSig.Name, releaseTask.LatestReleaseTaskSig.UUID)
	}

	return
}

func (helm *Helm) buildReleaseInfoV2(releaseCache *releaseModel.ReleaseCache) (*releaseModel.ReleaseInfoV2, error) {
	releaseV1, err := helm.buildReleaseInfo(releaseCache)
	if err != nil {
		logrus.Errorf("failed to build release info: %s", err.Error())
		return nil, err
	}

	releaseV2 := &releaseModel.ReleaseInfoV2{ReleaseInfo: *releaseV1}
	releaseConfigResource, err := helm.k8sCache.GetResource(k8sModel.ReleaseConfigKind, releaseCache.Namespace, releaseCache.Name)
	if err != nil {
		if errorModel.IsNotFoundError(err) {
			releaseV2.DependenciesConfigValues = map[string]interface{}{}
			releaseV2.OutputConfigValues = map[string]interface{}{}
			releaseV2.ReleaseLabels = map[string]string{}
		} else {
			logrus.Errorf("failed to get release config : %s", err.Error())
			return nil, err
		}
	} else {
		releaseConfig := releaseConfigResource.(*k8sModel.ReleaseConfig)
		releaseV2.ConfigValues = releaseConfig.ConfigValues
		releaseV2.Dependencies = releaseConfig.Dependencies
		releaseV2.DependenciesConfigValues = releaseConfig.DependenciesConfigValues
		releaseV2.OutputConfigValues = releaseConfig.OutputConfig
		releaseV2.ReleaseLabels = releaseConfig.Labels
		releaseV2.RepoName = releaseConfig.Repo
		releaseV2.ChartImage = releaseConfig.ChartImage
	}
	releaseV2.ComputedValues = releaseCache.ComputedValues
	releaseV2.MetaInfoValues = releaseCache.MetaInfoValues
	releaseV2.Plugins, releaseV2.Paused, err = walmHelm.BuildReleasePluginsByConfigValues(releaseV2.ComputedValues)

	if releaseV2.Paused {
		releaseV2.Ready = false
		releaseV2.Message = "Release is paused now"
	}

	return releaseV2, nil
}

func (helm *Helm) buildReleaseInfo(releaseCache *releaseModel.ReleaseCache) (releaseInfo *releaseModel.ReleaseInfo, err error) {
	releaseInfo = &releaseModel.ReleaseInfo{}
	releaseInfo.ReleaseSpec = releaseCache.ReleaseSpec

	releaseInfo.Status, err = helm.k8sCache.GetResourceSet(releaseCache.ReleaseResourceMetas)
	if err != nil {
		logrus.Errorf(fmt.Sprintf("Failed to build the status of releaseInfo: %s", releaseInfo.Name))
		return
	}
	ready, notReadyResource := releaseInfo.Status.IsReady()
	if ready {
		releaseInfo.Ready = true
	} else {
		releaseInfo.Message = fmt.Sprintf("%s %s/%s is in state %s", notReadyResource.GetKind(), notReadyResource.GetNamespace(), notReadyResource.GetName(), notReadyResource.GetState().Status)
	}

	return
}

func (helm *Helm) ListReleases(namespace string) ([]*releaseModel.ReleaseInfoV2, error) {
	releaseTasks, err := helm.releaseCache.GetReleaseTasks(namespace)
	if err != nil {
		logrus.Errorf("failed to get release tasks with namespace=%s : %s", namespace, err.Error())
		return nil, err
	}

	releaseCaches, err := helm.releaseCache.GetReleaseCaches(namespace)
	if err != nil {
		logrus.Errorf("failed to get release caches with namespace=%s : %s", namespace, err.Error())
		return nil, err
	}

	return helm.doListReleases(releaseTasks, releaseCaches)
}

func (helm *Helm) ListReleasesByLabels(namespace string, labelSelectorStr string) ([]*releaseModel.ReleaseInfoV2, error) {
	releaseConfigs, err := helm.k8sCache.ListReleaseConfigs(namespace, labelSelectorStr)
	if err != nil {
		logrus.Errorf("failed to list release configs : %s", err.Error())
		return nil, err
	}

	return helm.listReleasesByReleaseConfigs(releaseConfigs)
}

func (helm *Helm) listReleasesByReleaseConfigs(releaseConfigs []*k8sModel.ReleaseConfig) ([]*releaseModel.ReleaseInfoV2, error) {
	if len(releaseConfigs) == 0 {
		return []*releaseModel.ReleaseInfoV2{}, nil
	}
	releaseTasks, err := helm.releaseCache.GetReleaseTasksByReleaseConfigs(releaseConfigs)
	if err != nil {
		logrus.Errorf("failed to get release tasks : %s", err.Error())
		return nil, err
	}

	releaseCaches, err := helm.releaseCache.GetReleaseCachesByReleaseConfigs(releaseConfigs)
	if err != nil {
		logrus.Errorf("failed to get release caches : %s", err.Error())
		return nil, err
	}

	return helm.doListReleases(releaseTasks, releaseCaches)
}

func (helm *Helm) doListReleases(releaseTasks []*releaseModel.ReleaseTask, releaseCaches []*releaseModel.ReleaseCache) (releaseInfos []*releaseModel.ReleaseInfoV2, err error) {
	releaseCacheMap := map[string]*releaseModel.ReleaseCache{}
	for _, releaseCache := range releaseCaches {
		releaseCacheMap[releaseCache.Namespace+"/"+releaseCache.Name] = releaseCache
	}

	releaseInfos = []*releaseModel.ReleaseInfoV2{}
	//TODO 限制协程的数量
	mux := &sync.Mutex{}
	var wg sync.WaitGroup
	for _, releaseTask := range releaseTasks {
		wg.Add(1)
		go func(releaseTask *releaseModel.ReleaseTask, releaseCache *releaseModel.ReleaseCache) {
			defer wg.Done()
			info, err1 := helm.buildReleaseInfoV2ByReleaseTask(releaseTask, releaseCache)
			if err1 != nil {
				err = errors.New(fmt.Sprintf("failed to build release info: %s", err1.Error()))
				logrus.Error(err.Error())
				return
			}
			mux.Lock()
			releaseInfos = append(releaseInfos, info)
			mux.Unlock()
		}(releaseTask, releaseCacheMap[releaseTask.Namespace+"/"+releaseTask.Name])
	}
	wg.Wait()
	if err != nil {
		return
	}
	return
}
