package transwarp

import (
	"fmt"
	"github.com/ghodss/yaml"
	yaml2 "gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"os"
	"strings"
	"transwarp/application-instance/pkg/apis/transwarp/v1beta1"
	clientsetex "transwarp/application-instance/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
)



func CheckDepencies(chartRequested *chart.Chart, depLinks map[string]interface{}) (error) {

	app := &AppDependency{}
	dep := make([]string, 0)
	appDepencies := make(map[string]string)

	for _, file := range chartRequested.Files {
		if file.TypeUrl == "transwarp-app-yaml" {
			err := yaml.Unmarshal(file.Value, &app)
			if err != nil {
				return err
			}
			for _, dependency := range app.Dependencies {
				dep = append(dep, dependency.Name)

				for key := range dependency.Requires {
					appDepencies[dependency.Name] = key
				}

			}
		}
	}

	if dep == nil || len(dep) == 0 {
		fmt.Fprintf(os.Stdout, "WARNING: Cannot found depencies in app.yaml")
		return nil
	}

	if appDepencies == nil || len(appDepencies) == 0 {
		fmt.Fprintf(os.Stdout, "WARNING: Cannot found depencies requires in app.yaml")
		return nil
	}

	found := false
	for k, v := range depLinks {

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

	}

	if found == false {
		fmt.Fprintf(os.Stdout, "WARN: All links is not  defined in app.yaml. skip...\n")
		return nil
	}

    return nil

}


func GetDepenciesConfig(k8sTranswarpClient *clientsetex.Clientset, k8sClient *kubernetes.Clientset, namespace string, depLinks map[string]interface{}) (map[string]interface{}, error) {

	annotations := make(map[string]interface{})

	for _, v := range depLinks {

		inst, err := k8sTranswarpClient.TranswarpV1beta1().ApplicationInstances(namespace).Get(v.(string), v1.GetOptions{})
		if err != nil {
			return nil, err
		}

		if inst == nil {
			fmt.Fprintf(os.Stdout, "WARN: cannot found instance %s in %s skip...\n", v.(string), namespace)
			return nil, nil
		}

		serviceName := getDepenciesServiceName(inst)
		if serviceName == "" {
			fmt.Fprintf(os.Stdout, "WARN: cannot found service which startwith %s skip...\n", string("app-dummy-"))
			return nil, nil
		}

		service, err := k8sClient.CoreV1().Services(namespace).Get(serviceName, v1.GetOptions{})
		if err != nil {
			return nil, nil
		}

		serviceAnnotations := service.ObjectMeta.Annotations
		resolveServiceAnnotations(serviceAnnotations, &annotations)

	}

	return annotations, nil

}

func resolveServiceAnnotations(serviceAnnotations map[string]string, annotations *map[string]interface{}) error {

	depenciesConfig := make(map[string]map[string]map[string]interface{})
	for _, val := range serviceAnnotations {

		err := yaml2.Unmarshal([]byte(val), &depenciesConfig)
		if err != nil {
			return err
		}

		for _, val1 := range depenciesConfig {

			for k2, val2 := range val1 {

				for _, val3 := range val2 {

					(*annotations)[k2] = val3

				}
			}
		}

	}

	return nil
}

func getDepenciesServiceName(inst *v1beta1.ApplicationInstance) string {

	modules := inst.Status.Modules
	if modules == nil || len(modules) == 0 {
		return ""
	}

	for _, module := range modules {

		resourceRef := module.ResourceRef
		if resourceRef.Kind == "Service" && strings.HasPrefix(string(resourceRef.Name), "app-dummy-") {
			return string(resourceRef.Name)
		}
	}

	return ""

}
