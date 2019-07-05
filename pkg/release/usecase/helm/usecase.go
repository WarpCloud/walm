package helm

import (
	"WarpCloud/walm/pkg/k8s"
	"WarpCloud/walm/pkg/helm"
	"WarpCloud/walm/pkg/task"
	"WarpCloud/walm/pkg/release"
	"github.com/sirupsen/logrus"
	"fmt"
	releaseModel "WarpCloud/walm/pkg/models/release"
	k8sModel "WarpCloud/walm/pkg/models/k8s"
	errorModel "WarpCloud/walm/pkg/models/error"
	"WarpCloud/walm/pkg/release/utils"
)

type Helm struct {
	releaseCache release.Cache
	helm         helm.Helm
	k8sCache     k8s.Cache
	k8sOperator  k8s.Operator
	task         task.Task
}

// reload dependencies config values, if changes, upgrade release
func (helm *Helm) ReloadRelease(namespace, name string) error {
	releaseInfo, err := helm.GetRelease(namespace, name)
	if err != nil {
		if errorModel.IsNotFoundError(err) {
			logrus.Warnf("release %s/%s is not found， ignore to reload release", namespace, name)
			return nil
		}
		logrus.Errorf("failed to get release %s/%s : %s", namespace, name, err.Error())
		return err
	}

	chartInfo, err := helm.helm.GetChartDetailInfo(releaseInfo.RepoName, releaseInfo.ChartName, releaseInfo.ChartVersion)
	if err != nil {
		logrus.Errorf("failed to get chart info : %s", err.Error())
		return err
	}

	oldDependenciesConfigValues := releaseInfo.DependenciesConfigValues
	newDependenciesConfigValues, err := helm.getDependencyOutputConfigs(namespace, releaseInfo.Dependencies, chartInfo.MetaInfo)
	if err != nil {
		logrus.Errorf("failed to get dependencies output configs of %s/%s : %s", namespace, name, err.Error())
		return err
	}

	if utils.ConfigValuesDiff(oldDependenciesConfigValues, newDependenciesConfigValues) {
		releaseRequest := releaseInfo.BuildReleaseRequestV2()
		err = helm.InstallUpgradeRelease(namespace, releaseRequest, nil, false, 0, nil)
		if err != nil {
			logrus.Errorf("failed to upgrade release v2 %s/%s : %s", namespace, name, err.Error())
			return err
		}
		logrus.Infof("succeed to reload release %s/%s", namespace, name)
	} else {
		logrus.Infof("ignore reloading release %s/%s : dependencies config value does not change", namespace, name)
	}

	return nil
}

func (helm *Helm) getDependencyOutputConfigs(namespace string, dependencies map[string]string, chartMetaInfo *releaseModel.ChartMetaInfo) (dependencyConfigs map[string]interface{}, err error) {
	dependencyConfigs = map[string]interface{}{}
	if chartMetaInfo == nil {
		return
	}

	chartDependencies := chartMetaInfo.ChartDependenciesInfo
	dependencyAliasConfigVars := map[string]string{}
	for _, chartDependency := range chartDependencies {
		dependencyAliasConfigVars[chartDependency.Name] = chartDependency.AliasConfigVar
	}

	for dependencyKey, dependency := range dependencies {
		dependencyAliasConfigVar, ok := dependencyAliasConfigVars[dependencyKey]
		if !ok {
			err = fmt.Errorf("dependency key %s is not valid, you can see valid keys in chart metainfo", dependencyKey)
			logrus.Errorf(err.Error())
			return
		}

		dependencyNamespace, dependencyName, err := utils.ParseDependedRelease(namespace, dependency)
		if err != nil {
			return nil, err
		}

		dependencyReleaseConfigResource, err := helm.k8sCache.GetResource(k8sModel.ReleaseConfigKind, dependencyNamespace, dependencyName)
		if err != nil {
			if errorModel.IsNotFoundError(err) {
				logrus.Warnf("release config %s/%s is not found", dependencyNamespace, dependencyName)
				continue
			}
			logrus.Errorf("failed to get release config %s/%s : %s", dependencyNamespace, dependencyName, err.Error())
			return nil, err
		}

		dependencyReleaseConfig := dependencyReleaseConfigResource.(*k8sModel.ReleaseConfig)
		if len(dependencyReleaseConfig.OutputConfig) > 0 {
			dependencyConfigs[dependencyAliasConfigVar] = dependencyReleaseConfig.OutputConfig
		}
	}
	return
}

func (helm *Helm) validateReleaseTask(namespace, name string, allowReleaseTaskNotExist bool) (releaseTask *releaseModel.ReleaseTask, err error) {
	releaseTask, err = helm.releaseCache.GetReleaseTask(namespace, name)
	if err != nil {
		if !errorModel.IsNotFoundError(err) {
			logrus.Errorf("failed to get release task : %s", err.Error())
			return
		} else if !allowReleaseTaskNotExist {
			return
		} else {
			err = nil
		}
	} else {
		taskState, err := helm.task.GetTaskState(releaseTask.LatestReleaseTaskSig)
		if err != nil {
			if errorModel.IsNotFoundError(err) {
				err = nil
				return releaseTask, err
			} else {
				logrus.Errorf("failed to get the last release task state : %s", err.Error())
				return releaseTask, err
			}
		}

		if !(taskState.IsFinished() || taskState.IsTimeout()) {
			err = fmt.Errorf("please wait for the last release task %s-%s finished or timeout", releaseTask.LatestReleaseTaskSig.Name, releaseTask.LatestReleaseTaskSig.UUID)
			logrus.Warn(err.Error())
			return
		}
	}
	return
}
