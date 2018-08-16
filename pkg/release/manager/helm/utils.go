package helm

import (
	"walm/pkg/release"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/chartutil"
	"github.com/sirupsen/logrus"
	"fmt"
	hapiRelease "k8s.io/helm/pkg/proto/hapi/release"
	"github.com/ghodss/yaml"
	"walm/pkg/k8s/client"
	"bytes"
	"walm/pkg/k8s/adaptor"
)

func BuildReleaseListOptions(option *release.ReleaseListOption) (options []helm.ReleaseListOption) {
	if option == nil {
		return
	}
	if option.Namespace != "" {
		options = append(options, helm.ReleaseListNamespace(option.Namespace))
	}
	if option.Filter != "" {
		options = append(options, helm.ReleaseListFilter(option.Filter))
	}
	if option.Limit != 0 {
		options = append(options, helm.ReleaseListLimit(option.Limit))
	}
	if option.Offset != "" {
		options = append(options, helm.ReleaseListOffset(option.Offset))
	}
	if option.Order != 0 {
		options = append(options, helm.ReleaseListOrder(option.Order))
	}
	if option.Sort != 0 {
		options = append(options, helm.ReleaseListSort(option.Sort))
	}
	if len(option.Statuses) > 0 {
		options = append(options, helm.ReleaseListStatuses(option.Statuses))
	}
	return
}

func buildReleaseInfo(helmRelease *hapiRelease.Release) (releaseInfo *release.ReleaseInfo, err error) {
	emptyChart := chart.Chart{}
	depLinks := make(map[string]string)
	releaseInfo = &release.ReleaseInfo{}
	releaseInfo.Name = helmRelease.Name
	releaseInfo.Namespace = helmRelease.Namespace
	releaseInfo.Version = helmRelease.Version
	releaseInfo.ChartVersion = helmRelease.Chart.Metadata.Version
	releaseInfo.ChartName = helmRelease.Chart.Metadata.Name
	releaseInfo.ChartAppVersion = helmRelease.Chart.Metadata.AppVersion
	cvals, err := chartutil.CoalesceValues(&emptyChart, helmRelease.Config)
	if err != nil {
		logrus.Errorf("parse raw values error %s\n", helmRelease.Config.Raw)
		return
	}
	releaseInfo.ConfigValues = cvals
	depValue, ok := helmRelease.Config.Values["dependencies"]
	if ok {
		yaml.Unmarshal([]byte(depValue.Value), &depLinks)
		releaseInfo.Dependencies = depLinks
	}

	if helmRelease.Info.Status.Code == hapiRelease.Status_DEPLOYED {
		releaseInfo.Status, err = buildReleaseStatus(helmRelease)
		if err != nil {
			logrus.Errorf(fmt.Sprintf("Failed to build the status of releaseInfo: %s", releaseInfo.Name))
			return
		}
		releaseInfo.Ready = isReleaseReady(releaseInfo.Status)
	}

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

func buildReleaseStatus(helmRelease *hapiRelease.Release) (*release.ReleaseStatus, error) {
	status := &release.ReleaseStatus{[]release.ReleaseResource{}}
	resourceMetas, err := getReleaseResourceMetas(helmRelease)
	if err != nil {
		return status, err
	}
	for _, resourceMeta := range resourceMetas {
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

func getReleaseResourceMetas(helmRelease *hapiRelease.Release) (resources []release.ReleaseResourceMeta, err error) {
	resources = []release.ReleaseResourceMeta{}
	results, err := client.GetKubeClient().BuildUnstructured(helmRelease.Namespace, bytes.NewBufferString(helmRelease.Manifest))
	if err != nil {
		return resources, err
	}
	for _, result := range results {
		resource := release.ReleaseResourceMeta{
			Kind:      result.Object.GetObjectKind().GroupVersionKind().Kind,
			Namespace: result.Namespace,
			Name:      result.Name,
		}
		resources = append(resources, resource)
	}
	return
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
