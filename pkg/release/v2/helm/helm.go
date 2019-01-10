package helm

import (
	hapirelease "k8s.io/helm/pkg/proto/hapi/release"
	"time"
	"k8s.io/helm/pkg/helm"
	"github.com/sirupsen/logrus"
	"fmt"
	helmv1 "walm/pkg/release/manager/helm"
	"gopkg.in/yaml.v2"
	"strings"
	"walm/pkg/k8s/handler"
	"walm/pkg/k8s/adaptor"
	releasev2 "walm/pkg/release/v2"
	walmerr "walm/pkg/util/error"
	"walm/pkg/release"
	"sync"
	"errors"
	"mime/multipart"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"walm/pkg/hook"
)

type HelmClientV2 struct {
	helmv1.HelmClient
	releaseConfigHandler    *handler.ReleaseConfigHandler
}

var helmClient *HelmClientV2

func GetDefaultHelmClientV2() *HelmClientV2 {
	if helmClient == nil {
		helmClient = &HelmClientV2{
			HelmClient: *helmv1.GetDefaultHelmClient(),
			releaseConfigHandler:    handler.GetDefaultHandlerSet().GetReleaseConfigHandler(),
		}
	}
	return helmClient
}

// reload dependencies config values, if changes, upgrade release
func (hc *HelmClientV2) GetReleaseV2(namespace, name string) (releaseV2 *releasev2.ReleaseInfoV2, err error) {
	releaseCache, err := hc.GetHelmCache().GetReleaseCache(namespace, name)
	if err != nil {
		logrus.Errorf("failed to get release cache of %s/%s : %s", namespace, name, err.Error())
		return nil, err
	}

	releaseV2, err = hc.buildReleaseInfoV2(releaseCache)
	if err != nil {
		logrus.Errorf("failed to build v2 release info : %s", err.Error())
		return nil, err
	}

	return
}

func  (hc *HelmClientV2)buildReleaseInfoV2(releaseCache *release.ReleaseCache) (*releasev2.ReleaseInfoV2, error) {
	releaseV1, err := helmv1.BuildReleaseInfo(releaseCache)
	if err != nil {
		logrus.Errorf("failed to build release info: %s", err.Error())
		return nil, err
	}
	releaseV2 := &releasev2.ReleaseInfoV2{ReleaseInfo: *releaseV1}
	releaseConfig, err := hc.releaseConfigHandler.GetReleaseConfig(releaseCache.Namespace, releaseCache.Name)
	if err != nil {
		if adaptor.IsNotFoundErr(err) {
			// Compatible
			releaseV2.DependenciesConfigValues = map[string]interface{}{}
			releaseV2.OutputConfigValues = map[string]interface{}{}
		} else {
			logrus.Errorf("failed to get release config : %s", err.Error())
			return nil, err
		}
	} else {
		releaseV2.ConfigValues = releaseConfig.Spec.ConfigValues
		releaseV2.Dependencies = releaseConfig.Spec.Dependencies
		releaseV2.DependenciesConfigValues = releaseConfig.Spec.DependenciesConfigValues
		releaseV2.OutputConfigValues = releaseConfig.Spec.OutputConfig
	}
	releaseV2.ComputedValues = releaseCache.ComputedValues
	return releaseV2, nil
}

func (hc *HelmClientV2) ListReleasesV2(namespace, filter string) ([]*releasev2.ReleaseInfoV2, error) {
	logrus.Debugf("Enter ListReleasesV2 namespace=%s filter=%s\n", namespace, filter)
	releaseCaches, err := hc.GetHelmCache().GetReleaseCaches(namespace, filter, 0)
	if err != nil {
		logrus.Errorf("failed to get release caches with namespace=%s filter=%s : %s", namespace, filter, err.Error())
		return nil, err
	}

	releaseInfos := []*releasev2.ReleaseInfoV2{}
	mux := &sync.Mutex{}
	var wg sync.WaitGroup
	for _, releaseCache := range releaseCaches {
		wg.Add(1)
		go func(releaseCache *release.ReleaseCache) {
			defer wg.Done()
			info, err1 := hc.buildReleaseInfoV2(releaseCache)
			if err1 != nil {
				err = errors.New(fmt.Sprintf("failed to build release info: %s", err1.Error()))
				logrus.Error(err.Error())
				return
			}
			mux.Lock()
			releaseInfos = append(releaseInfos, info)
			mux.Unlock()
		}(releaseCache)
	}
	wg.Wait()
	if err != nil {
		return nil, err
	}
	return releaseInfos, nil
}

// reload dependencies config values, if changes, upgrade release
func (hc *HelmClientV2) ReloadRelease(namespace, name string, isSystem bool) error {
	_, err := hc.GetHelmCache().GetReleaseCache(namespace, name)
	if err != nil {
		logrus.Errorf("failed to get release cache of %s/%s : %s", namespace, name, err.Error())
		return err
	}

	releaseConfig, err := handler.GetDefaultHandlerSet().GetReleaseConfigHandler().GetReleaseConfig(namespace, name)
	if err != nil {
		logrus.Errorf("failed to get release config of %s/%s : %s", namespace, name, err.Error())
		return err
	}

	oldDependenciesConfigValues := releaseConfig.Spec.DependenciesConfigValues
	newDependenciesConfigValues, err := hc.getDependencyOutputConfigs(namespace, releaseConfig.Spec.Dependencies)
	if err != nil {
		logrus.Errorf("failed to get dependencies output configs of %s/%s : %s", namespace, name, err.Error())
		return err
	}

	if ConfigValuesDiff(oldDependenciesConfigValues, newDependenciesConfigValues) {
		//TODO add spec RepoName
		chart, err := hc.GetChartRequest("", releaseConfig.Spec.ChartName, releaseConfig.Spec.ChartVersion)
		if err != nil {
			logrus.Errorf("failed to load chart %s-%s from %s : %s", releaseConfig.Spec.ChartName, releaseConfig.Spec.ChartVersion, "", err.Error())
			return err
		}

		isJsonnetChart, jsonnetChart, _, err := isJsonnetChart(chart)
		if err != nil {
			logrus.Errorf("failed to check whether is jsonnet chart : %s", err.Error())
			return err
		}

		if isJsonnetChart {
			chart, err = convertJsonnetChart(namespace, name, releaseConfig.Spec.Dependencies, jsonnetChart, releaseConfig.Spec.ConfigValues, newDependenciesConfigValues)
			if err != nil {
				logrus.Errorf("failed to convert jsonnet chart %s-%s from %s : %s", releaseConfig.Spec.ChartName, releaseConfig.Spec.ChartVersion, "", err.Error())
				return err
			}
		}

		valueOverride := map[string]interface{}{}
		helmv1.MergeValues(valueOverride, releaseConfig.Spec.ConfigValues)
		valueOverrideBytes, err := yaml.Marshal(valueOverride)

		currentHelmClient, err := hc.GetCurrentHelmClient(namespace, isSystem)
		if err != nil {
			logrus.Errorf("failed to get current helm client : %s", err.Error())
			return err
		}

		resp, err := currentHelmClient.UpdateReleaseFromChart(
			name,
			chart,
			helm.UpdateValueOverrides(valueOverrideBytes),
			helm.UpgradeDryRun(hc.GetDryRun()),
		)
		if err != nil {
			logrus.Errorf("failed to upgrade release %s/%s from chart : %s", namespace, name, err.Error())
			return err
		}
		err = hc.GetHelmCache().CreateOrUpdateReleaseCache(resp.GetRelease())
		if err != nil {
			logrus.Errorf("failed to update release cache of %s/%s : %s", namespace, name, err.Error())
			return err
		}

		logrus.Infof("succeed to reload release %s/%s", namespace, name)
	} else {
		logrus.Infof("ignore reloading release %s/%s : dependencies config value does not change", namespace, name)
	}

	return nil
}

func (hc *HelmClientV2) InstallUpgradeReleaseV2(namespace string, releaseRequest *releasev2.ReleaseRequestV2, isSystem bool, chartArchive multipart.File) error {
	update := true
	releaseInfo, err := hc.GetReleaseV2(namespace, releaseRequest.Name)
	if err != nil {
		if walmerr.IsNotFoundError(err) {
			update = false
		} else {
			logrus.Errorf("failed to get release cache of %s/%s : %s", namespace, releaseRequest.Name, err.Error())
			return err
		}
	}

	now := time.Now()
	if releaseRequest.ConfigValues == nil {
		releaseRequest.ConfigValues = map[string]interface{}{}
	}
	if releaseRequest.Dependencies == nil {
		releaseRequest.Dependencies = map[string]string{}
	}

	hook.ProcessPrettyParams(&(releaseRequest.ReleaseRequest))

	// if jsonnet chart, add template-jsonnet/, app.yaml to chart.Files
	// app.yaml : used to define chart dependency relations
	var chart *chart.Chart
	if chartArchive != nil {
		chart, err = helmv1.GetChart(chartArchive)
	} else {
		chart, err = hc.GetChartRequest(releaseRequest.RepoName, releaseRequest.ChartName, releaseRequest.ChartVersion)
	}
	if err != nil {
		logrus.Errorf("failed to load chart %s-%s from %s : %s", releaseRequest.ChartName, releaseRequest.ChartVersion, releaseRequest.RepoName, err.Error())
		return err
	}

	// get all the dependency releases' output configs from ReleaseConfig
	dependencyConfigs, err := hc.getDependencyOutputConfigs(namespace, releaseRequest.Dependencies)
	if err != nil {
		logrus.Errorf("failed to get all the dependency releases' output configs : %s", err.Error())
		return err
	}

	// check whether is jsonnet chart
	isJsonnetChart, jsonnetChart, _, err := isJsonnetChart(chart)
	if err != nil {
		logrus.Errorf("failed to check whether is jsonnet chart : %s", err.Error())
		return err
	}

	// reuse config values
	configValues := map[string]interface{}{}
	if update {
		releaseConfig, err := hc.releaseConfigHandler.GetReleaseConfig(namespace, releaseRequest.Name)
		if err != nil {
			if adaptor.IsNotFoundErr(err) {
				logrus.Warnf("release config %s/%s is not found", namespace, releaseRequest.Name)
				helmv1.MergeValues(configValues, releaseInfo.ConfigValues)
				if len(releaseInfo.Status.Instances) > 0 {
					err = fmt.Errorf("now v1 release %s/%s with instances is not support to upgrade", namespace, releaseRequest.Name)
					return err
				}
			} else {
				logrus.Errorf("failed to get release config : %s", err.Error())
				return err
			}
		} else {
			helmv1.MergeValues(configValues, releaseConfig.Spec.ConfigValues)
		}
	}
	helmv1.MergeValues(configValues, releaseRequest.ConfigValues)

	if isJsonnetChart {
		chart, err = convertJsonnetChart(namespace, releaseRequest.Name, releaseRequest.Dependencies, jsonnetChart, configValues, dependencyConfigs)
		if err != nil {
			logrus.Errorf("failed to convert jsonnet chart %s-%s from %s : %s", releaseRequest.ChartName, releaseRequest.ChartVersion, releaseRequest.RepoName, err.Error())
			return err
		}
	} else {
		//TODO native helm chart如何处理？
	}


	valueOverride := map[string]interface{}{}
	helmv1.MergeValues(valueOverride, configValues)
	valueOverrideBytes, err := yaml.Marshal(valueOverride)
	logrus.Infof("convert %s takes %v", releaseRequest.Name, time.Now().Sub(now))

	currentHelmClient, err := hc.GetCurrentHelmClient(namespace, isSystem)
	if err != nil {
		logrus.Errorf("failed to get current helm client : %s", err.Error())
		return err
	}

	var release *hapirelease.Release
	if update {
		resp, err := currentHelmClient.UpdateReleaseFromChart(
			releaseRequest.Name,
			chart,
			helm.UpdateValueOverrides(valueOverrideBytes),
			helm.UpgradeDryRun(hc.GetDryRun()),
		)
		if err != nil {
			logrus.Errorf("failed to upgrade release %s/%s from chart : %s", namespace, releaseRequest.Name, err.Error())
			return err
		}
		release = resp.GetRelease()
	} else {
		resp, err := currentHelmClient.InstallReleaseFromChart(
			chart,
			namespace,
			helm.ValueOverrides(valueOverrideBytes),
			helm.ReleaseName(releaseRequest.Name),
			helm.InstallDryRun(hc.GetDryRun()),
		)
		if err != nil {
			logrus.Errorf("failed to install release %s/%s from chart : %s", namespace, releaseRequest.Name, err.Error())
			opts := []helm.DeleteOption{
				helm.DeletePurge(true),
			}
			_, err1 := currentHelmClient.DeleteRelease(
				releaseRequest.Name, opts...,
			)
			if err1 != nil {
				logrus.Errorf("failed to rollback to delete release %s/%s : %s", namespace, releaseRequest.Name, err1.Error())
			}
			return err
		}
		release = resp.GetRelease()
	}

	err = hc.GetHelmCache().CreateOrUpdateReleaseCache(release)
	if err != nil {
		logrus.Errorf("failed to create of update release cache of %s/%s : %s", namespace, releaseRequest.Name, err.Error())
		return err
	}

	logrus.Infof("succeed to create or update release %s/%s", namespace, releaseRequest.Name)

	return nil
}

func (hc *HelmClientV2) getDependencyOutputConfigs(namespace string, dependencies map[string]string) (dependencyConfigs map[string]interface{}, err error) {
	dependencyConfigs = map[string]interface{}{}
	for _, dependency := range dependencies {
		ss := strings.Split(dependency, ".")
		if len(ss) > 2 {
			err = fmt.Errorf("dependency value %s should not contains more than 1 \".\"", dependency)
			return
		}
		dependencyNamespace, dependencyName := "", ""
		if len(ss) == 2 {
			dependencyNamespace = ss[0]
			dependencyName = ss[1]
		} else {
			dependencyNamespace = namespace
			dependencyName = ss[0]
		}
		dependencyReleaseConfig, err := hc.releaseConfigHandler.GetReleaseConfig(dependencyNamespace, dependencyName)
		if err != nil {
			if adaptor.IsNotFoundErr(err) {
				continue
			}
			logrus.Errorf("failed to get release config %s/%s : %s", dependencyNamespace, dependencyName, err.Error())
			return nil, err
		}

		// TODO how to deal with key conflict?
		if len(dependencyReleaseConfig.Spec.OutputConfig) > 0 {
			// compatible
			provideConfigValues, ok := dependencyReleaseConfig.Spec.OutputConfig["provides"].(map[string]interface{})
			if ok {
				valueToMerge := make(map[string]interface{}, len(provideConfigValues))
				for key, value := range provideConfigValues {
					if immediateValue, ok := value.(map[string]interface{}); ok {
						if immediateValue["immediate_value"] != nil {
							valueToMerge[key] = immediateValue["immediate_value"]
						}
					}
				}
				helmv1.MergeValues(dependencyConfigs, valueToMerge)
			} else {
				helmv1.MergeValues(dependencyConfigs, dependencyReleaseConfig.Spec.OutputConfig)
			}
		}
	}
	return
}

