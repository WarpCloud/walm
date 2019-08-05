package converter

import (
	"WarpCloud/walm/pkg/models/k8s"
	"transwarp/release-config/pkg/apis/transwarp/v1beta1"
)

func ConvertReleaseConfigFromK8s(oriReleaseConfig *v1beta1.ReleaseConfig) (*k8s.ReleaseConfig, error) {
	if oriReleaseConfig == nil {
		return nil, nil
	}
	releaseConfig := oriReleaseConfig.DeepCopy()
	return &k8s.ReleaseConfig{
		Meta:                     k8s.NewMeta(k8s.ReleaseConfigKind, releaseConfig.Namespace, releaseConfig.Name, k8s.NewState("Ready", "", "")),
		Labels:                   releaseConfig.Labels,
		OutputConfig:             releaseConfig.Spec.OutputConfig,
		ChartImage:               releaseConfig.Spec.ChartImage,
		ChartName:                releaseConfig.Spec.ChartName,
		ConfigValues:             releaseConfig.Spec.ConfigValues,
		Dependencies:             releaseConfig.Spec.Dependencies,
		ChartVersion:             releaseConfig.Spec.ChartVersion,
		ChartAppVersion:          releaseConfig.Spec.ChartAppVersion,
		Repo:                     releaseConfig.Spec.Repo,
		DependenciesConfigValues: releaseConfig.Spec.DependenciesConfigValues,
	}, nil
}
