package helm

import (
	"walm/pkg/release"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"github.com/sirupsen/logrus"
	"fmt"
	"github.com/ghodss/yaml"
	"walm/pkg/k8s/adaptor"
)

func buildReleaseInfo(releaseCache *release.ReleaseCache) (releaseInfo *release.ReleaseInfo, err error) {
	releaseInfo = &release.ReleaseInfo{}
	releaseInfo.ReleaseSpec = releaseCache.ReleaseSpec

	releaseInfo.Status, err = buildReleaseStatus(releaseCache.ReleaseResourceMetas)
	if err != nil {
		logrus.Errorf(fmt.Sprintf("Failed to build the status of releaseInfo: %s", releaseInfo.Name))
		return
	}
	releaseInfo.Ready = isReleaseReady(releaseInfo.Status)

	return
}
func isReleaseReady(status *release.ReleaseStatus) bool {
	ready := true
	for _, resource := range status.Resources {
		if resource.Resource.GetState().Status != "Ready" {
			ready = false
			break
		}
	}
	return ready
}

func buildReleaseStatus(releaseResourceMetas []release.ReleaseResourceMeta) (*release.ReleaseStatus, error) {
	status := &release.ReleaseStatus{[]release.ReleaseResource{}}
	for _, resourceMeta := range releaseResourceMetas {
		resource, err := adaptor.GetDefaultAdaptorSet().GetAdaptor(resourceMeta.Kind).GetResource(resourceMeta.Namespace, resourceMeta.Name)
		if err != nil {
			return status, err
		}
		if resource.GetState().Status == "Unknown" && resource.GetState().Reason == "NotSupportedKind" {
			continue
		}
		status.Resources = append(status.Resources, release.ReleaseResource{Kind: resource.GetKind(), Resource: resource})
	}
	return status, nil
}

func parseChartDependencies(chart *chart.Chart) ([]string, error) {
	var dependencies []string

	for _, chartFile := range chart.Files {
		logrus.Printf("Chartfile %s \n", chartFile.TypeUrl)
		if chartFile.TypeUrl == "transwarp-app-yaml" {
			app := &release.AppDependency{}
			err := yaml.Unmarshal(chartFile.Value, &app)
			if err != nil {
				return dependencies, err
			}
			for _, dependency := range app.Dependencies {
				dependencies = append(dependencies, dependency.Name)
			}
		}
	}

	return dependencies, nil
}
