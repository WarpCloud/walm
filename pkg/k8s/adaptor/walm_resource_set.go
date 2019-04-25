package adaptor

import (
	"github.com/sirupsen/logrus"
	"walm/pkg/k8s/handler"
)

type WalmResourceSet struct {
	Services     []WalmService     `json:"services" description:"release services"`
	ConfigMaps   []WalmConfigMap   `json:"configmaps" description:"release configmaps"`
	DaemonSets   []WalmDaemonSet   `json:"daemonsets" description:"release daemonsets"`
	Deployments  []WalmDeployment  `json:"deployments" description:"release deployments"`
	Ingresses    []WalmIngress     `json:"ingresses" description:"release ingresses"`
	Jobs         []WalmJob         `json:"jobs" description:"release jobs"`
	Secrets      []WalmSecret      `json:"secrets" description:"release secrets"`
	StatefulSets []WalmStatefulSet `json:"statefulsets" description:"release statefulsets"`
}

func (resourceSet *WalmResourceSet) GetPodsNeedRestart() []*WalmPod {
	walmPods := []*WalmPod{}
	for _, ds := range resourceSet.DaemonSets {
		if len(ds.Pods) > 0 {
			walmPods = append(walmPods, ds.Pods...)
		}
	}
	for _, ss := range resourceSet.StatefulSets {
		if len(ss.Pods) > 0 {
			walmPods = append(walmPods, ss.Pods...)
		}
	}
	for _, dp := range resourceSet.Deployments {
		if len(dp.Pods) > 0 {
			walmPods = append(walmPods, dp.Pods...)
		}
	}
	return walmPods
}

func (resourceSet *WalmResourceSet) IsReady() (bool, WalmResource) {
	for _, secret := range resourceSet.Secrets {
		if secret.State.Status != "Ready" {
			return false, secret
		}
	}

	for _, job := range resourceSet.Jobs {
		if job.State.Status != "Ready" {
			return false, job
		}
	}

	for _, statefulSet := range resourceSet.StatefulSets {
		if statefulSet.State.Status != "Ready" {
			return false, statefulSet
		}
	}

	for _, service := range resourceSet.Services {
		if service.State.Status != "Ready" {
			return false, service
		}
	}

	for _, ingress := range resourceSet.Ingresses {
		if ingress.State.Status != "Ready" {
			return false, ingress
		}
	}

	for _, deployment := range resourceSet.Deployments {
		if deployment.State.Status != "Ready" {
			return false, deployment
		}
	}

	for _, daemonSet := range resourceSet.DaemonSets {
		if daemonSet.State.Status != "Ready" {
			return false, daemonSet
		}
	}

	for _, configMap := range resourceSet.ConfigMaps {
		if configMap.State.Status != "Ready" {
			return false, configMap
		}
	}

	return true, nil
}

func (resourceSet *WalmResourceSet) Pause() (error) {
	for _, deployment := range resourceSet.Deployments {
		err := GetDefaultAdaptorSet().walmDeploymentAdaptor.Scale(deployment.Namespace, deployment.Name, 0)
		if err != nil {
			logrus.Errorf("failed to pause deployment %s/%s : %s", deployment.Namespace, deployment.Name, err.Error())
			return err
		}
	}
	for _, statefulSet := range resourceSet.StatefulSets {
		err := handler.GetDefaultHandlerSet().GetStatefulSetHandler().Scale(statefulSet.Namespace, statefulSet.Name, 0)
		if err != nil {
			logrus.Errorf("failed to pause stateful set %s/%s : %s", statefulSet.Namespace, statefulSet.Name, err.Error())
			return err
		}
	}

	return nil
}

func NewWalmResourceSet() *WalmResourceSet {
	return &WalmResourceSet{
		StatefulSets: []WalmStatefulSet{},
		Services:     []WalmService{},
		Jobs:         []WalmJob{},
		Ingresses:    []WalmIngress{},
		Deployments:  []WalmDeployment{},
		DaemonSets:   []WalmDaemonSet{},
		ConfigMaps:   []WalmConfigMap{},
		Secrets:      []WalmSecret{},
	}
}
