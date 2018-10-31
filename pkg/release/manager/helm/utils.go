package helm

import (
	"walm/pkg/release"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"github.com/sirupsen/logrus"
	"fmt"
	"github.com/ghodss/yaml"
	"walm/pkg/k8s/adaptor"
	"k8s.io/helm/pkg/transwarp"
)

func buildReleaseInfo(releaseCache *release.ReleaseCache) (releaseInfo *release.ReleaseInfo, err error) {
	releaseInfo = &release.ReleaseInfo{}
	releaseInfo.ReleaseSpec = releaseCache.ReleaseSpec

	releaseInfo.Status, err = buildReleaseStatus(releaseCache.ReleaseResourceMetas)
	if err != nil {
		logrus.Errorf(fmt.Sprintf("Failed to build the status of releaseInfo: %s", releaseInfo.Name))
		return
	}
	releaseInfo.Ready = releaseInfo.Status.IsReady()

	return
}

func buildReleaseStatus(releaseResourceMetas []release.ReleaseResourceMeta) (resourceSet *adaptor.WalmResourceSet,err error) {
	resourceSet = adaptor.NewWalmResourceSet()
	for _, resourceMeta := range releaseResourceMetas {
		resource, err := adaptor.GetDefaultAdaptorSet().GetAdaptor(resourceMeta.Kind).GetResource(resourceMeta.Namespace, resourceMeta.Name)
		if err != nil {
			return nil, err
		}
		resource.AddToWalmResourceSet(resourceSet)
	}
	return
}

func parseChartDependencies(chart *chart.Chart) ([]string, error) {
	var dependencies []string

	for _, chartFile := range chart.Files {
		if chartFile.TypeUrl == "transwarp-app-yaml" {
			app := &transwarp.AppDependency{}
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
