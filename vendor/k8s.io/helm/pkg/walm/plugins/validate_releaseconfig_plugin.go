package plugins

import (
	"k8s.io/helm/pkg/hapi/release"
	"k8s.io/helm/pkg/walm"
	"bytes"
	"k8s.io/helm/pkg/tiller/environment"
	"k8s.io/helm/pkg/kube"
	"k8s.io/cli-runtime/pkg/genericclioptions/resource"
	"fmt"
	"transwarp/release-config/pkg/apis/transwarp/v1beta1"
)

// ValidateReleaseConfig plugin is used to make sure:
// 1. release have and only have one ReleaseConfig
// 2. ReleaseConfig has the same namespace and name with the release

func init() {
	walm.Register("ValidateReleaseConfig", &walm.WalmPluginRunner{
		Run:  ValidateReleaseConfig,
		Type: walm.Pre_Install,
	})
}


func ValidateReleaseConfig(context *walm.WalmPluginManagerContext, args string) (err error) {
	resources, err := buildResources(context.KubeClient, context.R)
	if err != nil {
		return err
	}
	releaseConfigResources := []*resource.Info{}
	for _, resource := range resources {
		if resource.Object.GetObjectKind().GroupVersionKind().Kind == "ReleaseConfig" {
			if resource.Name != context.R.Name {
				continue
			}
			releaseConfigResources = append(releaseConfigResources, resource)
		}
	}

	releaseConfigNum := len(releaseConfigResources)
	if releaseConfigNum == 0 {
		return fmt.Errorf("release must have one ReleaseConfig resource")
	} else if releaseConfigNum == 2 {
		releaseConfigs := []*v1beta1.ReleaseConfig{}
		for _, releaseConfigResource := range releaseConfigResources {
			releaseConfigs = append(releaseConfigs, releaseConfigResource.Object.(*v1beta1.ReleaseConfig))
		}
		//TODO
	} else if releaseConfigNum > 2 {
		return fmt.Errorf("release can not have more than one ReleaseConfig resource")
	}

	return
}

func buildResources(kubeClient environment.KubeClient, r *release.Release) (kube.Result, error){
	return kubeClient.BuildUnstructured(r.Namespace, bytes.NewBufferString(r.Manifest))
}

