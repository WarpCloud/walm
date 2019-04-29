package helm

import (
	"walm/pkg/release"
	"k8s.io/helm/pkg/chart/loader"
	"github.com/sirupsen/logrus"
	"bytes"
	"walm/pkg/k8s/client"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (hc *HelmClient) DryRunRelease(namespace string, releaseRequest *release.ReleaseRequestV2, isSystem bool, chartFiles []*loader.BufferedFile) ([]map[string]interface{}, error) {
	release, err := hc.doInstallUpgradeRelease(namespace, releaseRequest, isSystem, chartFiles, true)
	if err != nil {
		logrus.Errorf("failed to dry run install release : %s", err.Error())
		return nil, err
	}
	logrus.Debugf("release manifest : %s", release.Manifest)
	resources, err := client.GetKubeClient(namespace).BuildUnstructured(namespace, bytes.NewBufferString(release.Manifest))
	if err != nil {
		logrus.Errorf("failed to build unstructured : %s", err.Error())
		return nil, err
	}

	results := []map[string]interface{}{}
	for _, resource := range resources {
		results = append(results, resource.Object.(*unstructured.Unstructured).Object)
	}

	return results, nil
}