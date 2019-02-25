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

func ValidateReleaseConfig(context *walm.WalmPluginManagerContext, args string) (err error) {
	releaseConfigKey := walm.BuildResourceKey("ReleaseConfig", context.R.Namespace, context.R.Name)
	releaseConfigResources := []runtime.Object{}
	for resourceKey, resource := range context.Resources {
		if resource.GetObjectKind().GroupVersionKind().Kind == "ReleaseConfig" {
			delete(context.Resources, resourceKey)
			if resourceKey != releaseConfigKey {
				continue
			}
			releaseConfigResources = append(releaseConfigResources, resource)
		}
	}

	releaseConfigNum := len(releaseConfigResources)
	if releaseConfigNum == 0 {
		return fmt.Errorf("release must have one ReleaseConfig resource")
	} else if releaseConfigNum == 1 {
		context.Resources[releaseConfigKey] = releaseConfigResources[0]
	} else if releaseConfigNum == 2 {
		var autoGenReleaseConfig, releaseConfig *v1beta1.ReleaseConfig
		for _, releaseConfigResource := range releaseConfigResources {
			rc := releaseConfigResource.(*v1beta1.ReleaseConfig)
			if len(rc.Labels) > 0 && rc.Labels[AutoGenLabelKey] == "true" {
				autoGenReleaseConfig = rc
			} else {
				releaseConfig = rc
			}
		}
		if autoGenReleaseConfig == nil {
			return fmt.Errorf("release can not have more than one ReleaseConfig resource")
		}
		if releaseConfig == nil {
			return fmt.Errorf("release can not have more than one auto gen ReleaseConfig resource")
		}
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
			releaseConfig.Labels[k] = v
		}

		context.Resources[releaseConfigKey] = releaseConfig
	} else if releaseConfigNum > 2 {
		return fmt.Errorf("release can not have more than one ReleaseConfig resource")
	}

	return
}


