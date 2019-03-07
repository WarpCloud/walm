package helm

import (
	"walm/pkg/release"
	"walm/pkg/release/manager/helm/cache"
	"fmt"
	"github.com/sirupsen/logrus"
	"walm/pkg/k8s/adaptor"
	"k8s.io/helm/pkg/walm"
	"k8s.io/helm/pkg/walm/plugins"
	"sync"
	walmerr "walm/pkg/util/error"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"errors"
)

func (hc *HelmClient) GetRelease(namespace, name string) (releaseV2 *release.ReleaseInfoV2, err error) {
	releaseTask, err := hc.helmCache.GetReleaseTask(namespace, name)
	if err != nil {
		return nil, err
	}

	releaseV2, err = hc.buildReleaseInfoV2ByReleaseTask(releaseTask, nil)
	if err != nil {
		return nil, err
	}

	return
}

// 当task sig为空 或者task state已经ttl：build release by cache
// 当task state没有ttl:
// 1. 当task没完成：返回message
// 2. 当task完成：
//       a. 成功：build release by cache
//       b. 失败：返回message
func (hc *HelmClient) buildReleaseInfoV2ByReleaseTask(releaseTask *cache.ReleaseTask, releaseCache *release.ReleaseCache) (releaseV2 *release.ReleaseInfoV2, err error) {
	releaseV2 = &release.ReleaseInfoV2{
		ReleaseInfo: release.ReleaseInfo{
			ReleaseSpec: release.ReleaseSpec{
				Namespace: releaseTask.Namespace,
				Name:      releaseTask.Name,
			},
		},
	}
	if releaseTask.LatestReleaseTaskSig != nil {
		taskState := releaseTask.LatestReleaseTaskSig.GetTaskState()
		if taskState != nil && taskState.TaskName != "" {
			if taskState.IsSuccess() {
				if taskState.TaskName == deleteReleaseTaskName {
					return nil, walmerr.NotFoundError{}
				}
			} else if taskState.IsFailure() {
				releaseV2.Message = fmt.Sprintf("the release latest task %s-%s failed : %s", releaseTask.LatestReleaseTaskSig.Name, releaseTask.LatestReleaseTaskSig.UUID, taskState.Error)
				return
			} else {
				releaseV2.Message = fmt.Sprintf("please wait for the release latest task %s-%s finished", releaseTask.LatestReleaseTaskSig.Name, releaseTask.LatestReleaseTaskSig.UUID)
				return
			}
		}
	}

	if releaseCache == nil {
		releaseCache, err = hc.helmCache.GetReleaseCache(releaseTask.Namespace, releaseTask.Name)
		if err != nil {
			logrus.Errorf("failed to get release cache of %s/%s : %s", releaseTask.Namespace, releaseTask.Name, err.Error())
			return
		}
	}

	releaseV2, err = hc.buildReleaseInfoV2(releaseCache)
	if err != nil {
		logrus.Errorf("failed to build v2 release info : %s", err.Error())
		return nil, err
	}
	return
}

func (hc *HelmClient) buildReleaseInfoV2(releaseCache *release.ReleaseCache) (*release.ReleaseInfoV2, error) {
	releaseV1, err := buildReleaseInfo(releaseCache)
	if err != nil {
		logrus.Errorf("failed to build release info: %s", err.Error())
		return nil, err
	}
	releaseV2 := &release.ReleaseInfoV2{ReleaseInfo: *releaseV1}
	releaseConfig, err := hc.releaseConfigHandler.GetReleaseConfig(releaseCache.Namespace, releaseCache.Name)
	releaseConfig.DeepCopy()
	if err != nil {
		if adaptor.IsNotFoundErr(err) {
			releaseV2.DependenciesConfigValues = map[string]interface{}{}
			releaseV2.OutputConfigValues = map[string]interface{}{}
			releaseV2.ReleaseLabels = map[string]string{}
		} else {
			logrus.Errorf("failed to get release config : %s", err.Error())
			return nil, err
		}
	} else {
		// if release config is not deeply copied, release config cache in memory would be changed when
		// the var below who is referred is changed(such as dependencies, releaseLabels)
		releaseConfigCopy := releaseConfig.DeepCopy()
		releaseV2.ConfigValues = releaseConfigCopy.Spec.ConfigValues
		releaseV2.Dependencies = releaseConfigCopy.Spec.Dependencies
		releaseV2.DependenciesConfigValues = releaseConfigCopy.Spec.DependenciesConfigValues
		releaseV2.OutputConfigValues = releaseConfigCopy.Spec.OutputConfig
		releaseV2.ReleaseLabels = releaseConfigCopy.Labels
		releaseV2.RepoName = releaseConfig.Spec.Repo
	}
	releaseV2.ComputedValues = releaseCache.ComputedValues
	releaseV2.Plugins = []*walm.WalmPlugin{}
	if releaseV2.ComputedValues != nil {
		if walmPlugins, ok := releaseV2.ComputedValues[walm.WalmPluginConfigKey]; ok {
			delete(releaseV2.ComputedValues, walm.WalmPluginConfigKey)
			for _, plugin := range walmPlugins.([]interface{}) {
				walmPlugin := plugin.(map[string]interface{})
				if walmPlugin["name"].(string) != plugins.ValidateReleaseConfigPluginName {
					releaseV2.Plugins = append(releaseV2.Plugins, &walm.WalmPlugin{
						Name:    walmPlugin["name"].(string),
						Args:    walmPlugin["args"].(string),
						Version: walmPlugin["version"].(string),
						Disable: walmPlugin["disable"].(bool),
					})
				}
			}
		}
	}
	return releaseV2, nil
}

func (hc *HelmClient) ListReleases(namespace, filter string) ([]*release.ReleaseInfoV2, error) {
	logrus.Debugf("Enter ListReleases namespace=%s filter=%s\n", namespace, filter)
	releaseTasks, err := hc.helmCache.GetReleaseTasks(namespace, filter, 0)
	if err != nil {
		logrus.Errorf("failed to get release tasks with namespace=%s filter=%s : %s", namespace, filter, err.Error())
		return nil, err
	}

	releaseCaches, err := hc.helmCache.GetReleaseCaches(namespace, filter, 0)
	if err != nil {
		logrus.Errorf("failed to get release caches with namespace=%s filter=%s : %s", namespace, filter, err.Error())
		return nil, err
	}

	return hc.doListReleases(releaseTasks, releaseCaches)
}

func (hc *HelmClient) ListReleasesByLabels(namespace string, labelSelector *v1.LabelSelector) ([]*release.ReleaseInfoV2, error) {
	releaseConfigs, err := hc.releaseConfigHandler.ListReleaseConfigs(namespace, labelSelector)
	if err != nil {
		logrus.Errorf("failed to list release configs : %s", err.Error())
		return nil, err
	}

	releaseNames := []cache.ReleaseFieldName{}
	for _, releaseConfig := range releaseConfigs {
		releaseNames = append(releaseNames, cache.ReleaseFieldName{
			Namespace: releaseConfig.Namespace,
			Name:      releaseConfig.Name,
		})
	}

	releases, err := hc.listReleasesByNames(releaseNames)
	if err != nil {
		return nil, err
	}
	return releases, nil
}

func (hc *HelmClient) listReleasesByNames(names []cache.ReleaseFieldName) ([]*release.ReleaseInfoV2, error) {
	if len(names) == 0 {
		return []*release.ReleaseInfoV2{}, nil
	}
	releaseTasks, err := hc.helmCache.GetReleaseTasksByNames(names)
	if err != nil {
		logrus.Errorf("failed to get release tasks : %s", err.Error())
		return nil, err
	}

	releaseCaches, err := hc.helmCache.GetReleaseCachesByNames(names)
	if err != nil {
		logrus.Errorf("failed to get release caches : %s", err.Error())
		return nil, err
	}

	return hc.doListReleases(releaseTasks, releaseCaches)
}

func (hc *HelmClient) doListReleases(releaseTasks []*cache.ReleaseTask, releaseCaches []*release.ReleaseCache) (releaseInfos []*release.ReleaseInfoV2, err error) {
	releaseCacheMap := map[string]*release.ReleaseCache{}
	for _, releaseCache := range releaseCaches {
		releaseCacheMap[releaseCache.Namespace+"/"+releaseCache.Name] = releaseCache
	}

	releaseInfos = []*release.ReleaseInfoV2{}
	//TODO 限制协程的数量
	mux := &sync.Mutex{}
	var wg sync.WaitGroup
	for _, releaseTask := range releaseTasks {
		wg.Add(1)
		go func(releaseTask *cache.ReleaseTask, releaseCache *release.ReleaseCache) {
			defer wg.Done()
			info, err1 := hc.buildReleaseInfoV2ByReleaseTask(releaseTask, releaseCache)
			if err1 != nil {
				if walmerr.IsNotFoundError(err1) {
					return
				}
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
