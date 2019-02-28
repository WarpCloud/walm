package plugins

import (
	"k8s.io/helm/pkg/walm"
	"fmt"
	"transwarp/release-config/pkg/apis/transwarp/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	AutoGenLabelKey = "auto-gen"
	ValidateReleaseConfigPluginName = "ValidateReleaseConfig"
)

// ValidateReleaseConfig plugin is used to make sure:
// 1. release have and only have one ReleaseConfig
// 2. ReleaseConfig has the same namespace and name with the release

func init() {
	walm.Register(ValidateReleaseConfigPluginName, &walm.WalmPluginRunner{
		Run:  ValidateReleaseConfig,
		Type: walm.Pre_Install,
	})
}

func ValidateReleaseConfig(context *walm.WalmPluginManagerContext, args string) error {
	var autoGenReleaseConfig, releaseConfig *v1beta1.ReleaseConfig
	newResource := []runtime.Object{}
	for _, resource := range context.Resources {
		if resource.GetObjectKind().GroupVersionKind().Kind == "ReleaseConfig" {
			rc := resource.(*v1beta1.ReleaseConfig)
			if rc.Name != context.R.Name {
				continue
			}
			if len(rc.Labels) > 0 && rc.Labels[AutoGenLabelKey] == "true" {
				if autoGenReleaseConfig != nil {
					return fmt.Errorf("release can not have more than one auto-gen ReleaseConfig resource")
				} 
				autoGenReleaseConfig = rc
			} else {
				if releaseConfig != nil {
					return fmt.Errorf("release can not have more than one defined ReleaseConfig resource")
				}
				releaseConfig = rc
			}
		} else {
			newResource = append(newResource, resource)
		}
	}

	if autoGenReleaseConfig == nil {
		if releaseConfig == nil {
			return fmt.Errorf("release must have one ReleaseConfig resource")
		} else {
			newResource = append(newResource, releaseConfig)
		}
	} else {
		if releaseConfig == nil {
			newResource = append(newResource, autoGenReleaseConfig)
		}else {
			releaseConfig.Spec.Dependencies = autoGenReleaseConfig.Spec.Dependencies
			releaseConfig.Spec.DependenciesConfigValues = autoGenReleaseConfig.Spec.DependenciesConfigValues
			releaseConfig.Spec.ConfigValues = autoGenReleaseConfig.Spec.ConfigValues
			releaseConfig.Spec.ChartName = autoGenReleaseConfig.Spec.ChartName
			releaseConfig.Spec.ChartVersion = autoGenReleaseConfig.Spec.ChartVersion
			releaseConfig.Spec.ChartAppVersion = autoGenReleaseConfig.Spec.ChartAppVersion
			if releaseConfig.Labels == nil {
				releaseConfig.Labels = map[string]string{}
			}
			for k, v := range autoGenReleaseConfig.Labels {
				if k != AutoGenLabelKey {
					releaseConfig.Labels[k] = v
				}
			}

			newResource = append(newResource, releaseConfig)
		}
	}

	context.Resources = newResource
	return nil
}


