package transwarp

import (
	"errors"
	"fmt"
	"os"
	"hash/adler32"
	"time"

	"github.com/ghodss/yaml"

	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/chart"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rls "k8s.io/helm/pkg/proto/hapi/services"
	"k8s.io/api/core/v1"
	v1beta1 "transwarp/application-instance/pkg/apis/transwarp/v1beta1"
)


type DependencyDeclare struct {
	// name of dependency declaration
	Name string `json:"name,omitempty"`
	// dependency variable mappings
	Requires map[string]string `json:"requires,omitempty"`
}

type AppDependency struct {
	Name string `json:"name,omitempty"`
	Dependencies []*DependencyDeclare `json:"dependencies"`
}

type HelmNativeValues struct {
	ChartName string `json:"chartName"`
	ChartVersion string `json:"chartVersion"`
	AppVersion string `json:"appVersion"`
	ReleaseName string `json:"releaseName"`
	ReleaseNamespace string `json:"releaseNamespace"`
}

type AppHelmValues struct {
	Dependencies []*DependencyDeclare `json:"dependencies"`
	NativeValues HelmNativeValues `json:"HelmNativeValues"`
}

func ProcessAppCharts(client helm.Interface, chartRequested *chart.Chart, name, namespace, config string, depLinks map[string]interface{}) error {
	var dependencies []v1beta1.Dependency
	var appConfigMapName string

	appManagerType := false
	annotations := make(map[string]string)
	labels := make(map[string]string)
	app := &AppDependency{}
	dep := make([]string, 0)
	depValuesLinks := make(map[string]string)
	removeIdx := -1

	for idx, file := range chartRequested.Files {
		if file.TypeUrl == "transwarp-configmap-reserved" {
			appChart := string(file.Value[:])
			calc := adler32.Checksum([]byte(appChart))
			appConfigMapName = fmt.Sprintf("appmanager.%s.%d", name, uint64(calc))

			appChartConfigmap := &v1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind: "ConfigMap",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: appConfigMapName,
					Namespace: namespace,
				},
				Data: map[string]string{"release": appChart},
			}
			appChartData, err := yaml.Marshal(appChartConfigmap)
			if err != nil {
				return err
			}
			chartRequested.Templates = append(chartRequested.Templates,
				&chart.Template{Name: "templates/transwarp-configmap-reserved.yaml", Data: appChartData})
			removeIdx = idx
			appManagerType = true
		} else if file.TypeUrl == "transwarp-app-yaml" {
			err := yaml.Unmarshal(file.Value, &app)
			if err != nil {
				return err
			}
			for _, dependency := range app.Dependencies {
				dep = append(dep, dependency.Name)
			}
		}
	}
	if appManagerType == false {
		return nil
	}

	// Merge Values.yaml and rawVals and helm Native Charts Values
	rawValsBase := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(config), &rawValsBase); err != nil {
		return fmt.Errorf("failed to parse rawValues: %s", err)
	}
	chartRawBase := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(chartRequested.Values.Raw), &chartRawBase); err != nil {
		return fmt.Errorf("failed to parse rawValues: %s", err)
	}
	helmVals := AppHelmValues{}
	helmVals.Dependencies = app.Dependencies
	helmVals.NativeValues.ChartVersion = chartRequested.Metadata.Name
	helmVals.NativeValues.ChartVersion = chartRequested.Metadata.Version
	helmVals.NativeValues.AppVersion = chartRequested.Metadata.AppVersion
	helmVals.NativeValues.ReleaseName = name
	helmVals.NativeValues.ReleaseNamespace = namespace
	chartRawBase["HelmAdditionalValues"] = &helmVals

	rawValsBase = mergeValues(chartRawBase, rawValsBase)

	chartRequested.Files[len(chartRequested.Files)-1], chartRequested.Files[removeIdx] = chartRequested.Files[removeIdx], chartRequested.Files[len(chartRequested.Files)-1]
	chartRequested.Files = chartRequested.Files[:len(chartRequested.Files)-1]

	for k, v := range depLinks {
		found := false
		for _, depName := range dep {
			if depName == k {
				found = true
				break
			}
		}
		if found == false {
			fmt.Fprintf(os.Stdout, "WARNING: links %s=%s not defined in app.yaml. skip...\n", k, v)
			continue
		}
		dependency := v1beta1.Dependency{}
		dependency.Name = k
		depInstanceName, depInstanceNamespace, err := findAppInstanceDependency(client, v.(string))
		if err != nil {
			return err
		}
		dependency.DependencyRef = v1.ObjectReference{
			Kind: "ApplicationInstance",
			Namespace: depInstanceNamespace,
			Name: depInstanceName,
			APIVersion: "apiextensions.transwarp.io/v1beta1",
		}
		dependencies = append(dependencies, dependency)
	}

	annotations["helm.sh/storage"] = "ConfigMap"
	annotations["helm.sh/storageName"] = appConfigMapName
	annotations["helm.sh/namespace"] = fmt.Sprintf("%s", namespace)

	labels["transwarp.install"] = name
	labels["transwarp.app"] = chartRequested.GetMetadata().GetName()
	labels["transwarp.name"] = name

	instance := &v1beta1.ApplicationInstance{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ApplicationInstance",
			APIVersion: "apiextensions.transwarp.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: annotations,
			Labels:      labels,
		},
		Spec: v1beta1.ApplicationInstanceSpec{
			ApplicationRef: v1beta1.ApplicationReference{
				Name:    chartRequested.Metadata.GetName(),
				Version: chartRequested.Metadata.GetAppVersion(),
			},
			Configs:      rawValsBase,
			Dependencies: dependencies,
		},
	}
	instanceData, err := yaml.Marshal(instance)
	if err != nil {
		return err
	}
	chartRequested.Templates = append(chartRequested.Templates,
		&chart.Template{Name: "templates/instance-crd.yaml", Data: instanceData})

	depValuesData, err := yaml.Marshal(depValuesLinks)
	if err != nil {
		return err
	}
	depValues := chart.Value{
		Value: string(depValuesData[:]),
	}
	chartRequested.Values.Values = make(map[string]*chart.Value)
	chartRequested.Values.Values["dependencies"] = &depValues

	return nil
}

func findAppInstanceDependency(client helm.Interface, releaseName string) (string, string, error) {
	var err error

	var releaseHistory *rls.GetHistoryResponse
	retry := 20

	for i:=0; i < retry; i++ {
		releaseHistory, err = client.ReleaseHistory(releaseName, helm.WithMaxHistory(1))
		time.Sleep(500 * time.Millisecond)
		if err == nil {
			break
		}
	}
	if err != nil {
		return "", "", err
	}
	if len(releaseHistory.Releases) == 0 {
		return "", "", errors.New(fmt.Sprintf("cannot found helm release %s", releaseName))
	}
	depRelease := releaseHistory.GetReleases()[0]

	return depRelease.Name, depRelease.Namespace, nil
}
