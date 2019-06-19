package plugins

import (
	"k8s.io/helm/pkg/walm"
	"fmt"
	"transwarp/release-config/pkg/apis/transwarp/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
			converted, err := convertUnstructured(resource.(*unstructured.Unstructured))
			if err != nil {
				context.Log("failed to convert unstructured : %s", err.Error())
				return err
			}
			rc := converted.(*v1beta1.ReleaseConfig)
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
		if len(autoGenReleaseConfig.Labels) > 0 {
			delete(autoGenReleaseConfig.Labels, AutoGenLabelKey)
		}
		if releaseConfig == nil {
			newResource = append(newResource, autoGenReleaseConfig)
		}else {
			autoGenReleaseConfig.Spec.OutputConfig = releaseConfig.Spec.OutputConfig
			if autoGenReleaseConfig.Labels == nil {
				autoGenReleaseConfig.Labels = map[string]string{}
			}

			for k, v := range releaseConfig.Labels {
				if _, ok := autoGenReleaseConfig.Labels[k]; !ok {
					autoGenReleaseConfig.Labels[k] = v
				}
			}

			newResource = append(newResource, autoGenReleaseConfig)
		}
	}

	context.Resources = newResource
	return nil
}


