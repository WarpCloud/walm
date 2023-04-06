package helm

import (
	"WarpCloud/walm/pkg/models/release"
	"github.com/sirupsen/logrus"
	"WarpCloud/walm/pkg/models/common"
)

func (helm *Helm) DryRunRelease(namespace string, releaseRequest *release.ReleaseRequestV2, chartFiles []*common.BufferedFile) ([]map[string]interface{}, error) {
	releaseCache, err := helm.doInstallUpgradeRelease(namespace, releaseRequest, chartFiles, true, nil)
	if err != nil {
		logrus.Errorf("failed to dry run install release : %s", err.Error())
		return nil, err
	}
	logrus.Debugf("release manifest : %s", releaseCache.Manifest)
	resources, err := helm.k8sOperator.BuildManifestObjects(namespace, releaseCache.Manifest)
	if err != nil {
		logrus.Errorf("failed to build unstructured : %s", err.Error())
		return nil, err
	}

	return resources, nil
}

func (helm *Helm) ComputeResourcesByDryRunRelease(namespace string, releaseRequest *release.ReleaseRequestV2, chartFiles []*common.BufferedFile) (*release.ReleaseResources, error) {
	r, err := helm.doInstallUpgradeRelease(namespace, releaseRequest, chartFiles, true, nil)
	if err != nil {
		logrus.Errorf("failed to dry run install release : %s", err.Error())
		return nil, err
	}
	logrus.Debugf("release manifest : %s", r.Manifest)
	resources, err := helm.k8sOperator.ComputeReleaseResourcesByManifest(namespace, r.Manifest)
	if err != nil {
		logrus.Errorf("failed to compute release resources by manifest : %s", err.Error())
		return nil, err
	}
	return resources, nil
}

func (helm *Helm) DryRunUpdateRelease(namespace string, releaseRequest *release.ReleaseRequestV2, chartFiles []*common.BufferedFile) ([]*k8sModel.ReleaseConfig, error) {
	releaseCache, err := helm.doInstallUpgradeRelease(namespace, releaseRequest, chartFiles, true,nil)
	if err != nil {
		klog.Errorf("failed to dry run install release : %s", err.Error())
		return nil, err
	}
	releaseInfo, err := helm.buildReleaseInfoV2(releaseCache)
	if err != nil {
		klog.Errorf("failed to build releaseInfo: %s", err.Error())
		klog.Errorf("failed to build unstructured : %s", err.Error())
		return nil, err
	}
	oldReleaseInfo, err := helm.GetRelease(releaseCache.Namespace, releaseCache.Name)
	if err != nil {
		return nil, err
	}
	var dependedReleases []*k8sModel.ReleaseConfig
	if utils.ConfigValuesDiff(oldReleaseInfo.OutputConfigValues, releaseInfo.OutputConfigValues) {
		releaseConfigs, err := helm.k8sCache.ListReleaseConfigs("", "")
		if err != nil {
			klog.Errorf("failed to list releaseconfigs: %s", err.Error())
			return nil, err
		}
		for _, releaseConfig := range releaseConfigs {
			for _, dependedRelease := range releaseConfig.Dependencies {
				dependedReleaseNamespace, dependedReleaseName, err := utils.ParseDependedRelease(releaseConfig.Namespace, dependedRelease)
				if err != nil {
					continue
				}
				if dependedReleaseNamespace == releaseInfo.Namespace && dependedReleaseName == releaseInfo.Name {
					if dependedReleaseNamespace == releaseCache.Namespace && dependedReleaseName == releaseCache.Name {
						dependedReleases = append(dependedReleases, releaseConfig)
					}
				}
			}
		}
	}
	return dependedReleases,err
}

