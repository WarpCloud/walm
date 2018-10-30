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
	"strings"
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
	Dependencies map[string]string `json:"dependencies"`
	NativeValues HelmNativeValues `json:"HelmNativeValues"`
}

// will change chartRequested values
func ProcessTranswarpChartRequested(chartRequested *chart.Chart, name, namespace string) (string, bool, *AppDependency, error) {
	var appConfigMapName string
	removeIdx := -1
	transwarpAppType := false
	appDependency := AppDependency{}

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
				return "", transwarpAppType, nil, err
			}
			chartRequested.Templates = append(chartRequested.Templates,
				&chart.Template{Name: "templates/transwarp-configmap-reserved.yaml", Data: appChartData})
			removeIdx = idx
			transwarpAppType = true
		} else if file.TypeUrl == "transwarp-app-yaml" {
			err := yaml.Unmarshal(file.Value, &appDependency)
			if err != nil {
				return "", transwarpAppType, nil, err
			}
		}
	}

	if transwarpAppType {
		chartRequested.Files[len(chartRequested.Files)-1], chartRequested.Files[removeIdx] = chartRequested.Files[removeIdx], chartRequested.Files[len(chartRequested.Files)-1]
		chartRequested.Files = chartRequested.Files[:len(chartRequested.Files)-1]
	}

	return appConfigMapName, transwarpAppType, &appDependency, nil
}

func FindAppInstanceDependency(client helm.Interface, releaseName string) (string, string, error) {
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

func GetTranswarpInstanceCRDDependency(client helm.Interface, appDependency *AppDependency, depLinks map[string]interface{}, defaultNamespace string, depMustExist bool) ([]v1beta1.Dependency, error) {
	var dependencies []v1beta1.Dependency
	dep := make([]string, 0)

	for _, dependency := range appDependency.Dependencies {
		dep = append(dep, dependency.Name)
	}
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
		depInstanceName := ""
		depInstanceNamespace := ""
		var err error
		if depMustExist {
			depInstanceName, depInstanceNamespace, err = FindAppInstanceDependency(client, v.(string))
			if err != nil {
				return nil, err
			}
		} else {
			depLinkValues := v.(string)
			depLinks := strings.SplitN(depLinkValues, ".", 2)
			if len(depLinks) > 1 {
				depInstanceNamespace = depLinks[0]
				depInstanceName = depLinks[1]
			} else {
				depInstanceNamespace = defaultNamespace
				depInstanceName = depLinkValues
			}
		}

		dependency.DependencyRef = v1.ObjectReference{
			Kind: "ApplicationInstance",
			Namespace: depInstanceNamespace,
			Name: depInstanceName,
			APIVersion: "apiextensions.transwarp.io/v1beta1",
		}
		dependencies = append(dependencies, dependency)
		//depValuesLinks[k] = v.(string)
	}

	return dependencies, nil
}

func ProcessTranswarpInstanceCRD(chartRequested *chart.Chart, name, namespace, config, appConfigMapName string, dependencies []v1beta1.Dependency) error {
	annotations := make(map[string]string)
	labels := make(map[string]string)
	depValuesLinks := make(map[string]string)

	// Unmarshal User Configs
	rawValsBase := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(config), &rawValsBase); err != nil {
		return fmt.Errorf("failed to parse rawValues: %s", err)
	}
	// add more helm values to instance
	helmVals := AppHelmValues{}
	helmVals.NativeValues.ChartVersion = chartRequested.Metadata.Name
	helmVals.NativeValues.ChartVersion = chartRequested.Metadata.Version
	helmVals.NativeValues.AppVersion = chartRequested.Metadata.AppVersion
	helmVals.NativeValues.ReleaseName = name
	helmVals.NativeValues.ReleaseNamespace = namespace
	chartRawBase := map[string]interface{}{}
	chartRawBase["HelmAdditionalValues"] = &helmVals
	helmVals.Dependencies = depValuesLinks
	rawValsBase = mergeValues(chartRawBase, rawValsBase)
	rawMergeVals, err := yaml.Marshal(rawValsBase)
	if err != nil {
		return err
	}
	chartRequested.Values.Raw = string(rawMergeVals[:])

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

	return nil
}

func ProcessAppCharts(client helm.Interface, chartRequested *chart.Chart, name, namespace, config string, depLinks map[string]interface{}) error {
	appConfigMapName, transwarpAppType, appDependency, err := ProcessTranswarpChartRequested(chartRequested, name, namespace)
	if err != nil {
		return err
	}
	if transwarpAppType == false {
		return nil
	}

	dependencies, err := GetTranswarpInstanceCRDDependency(client, appDependency, depLinks, namespace, true)
	if err != nil {
		return err
	}

	err = ProcessTranswarpInstanceCRD(chartRequested, name, namespace, config, appConfigMapName, dependencies)
	if err != nil {
		return err
	}

	return nil
}
