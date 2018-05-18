package cluster

import (
	"walm/pkg/helm"
)

func DeleteCluster(name, namespace string) error {
	var apps []string
	for _, app := range apps {
		if err := helm.Helm.DeplyApplications([]string{app}, []string{}); err != nil {
			return err
		}
	}
	return nil
}

func StatusCluster(name, namespace string) error {
	var apps []string
	for _, app := range apps {
		if err := helm.Helm.DeplyApplications([]string{app}, []string{"--all", "--namespace", namespace}); err != nil {
			return err
		}
	}
	return nil
}
