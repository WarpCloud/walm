package k8s

type ResourceSet struct {
	Services     []*Service     `json:"services" description:"release services"`
	ConfigMaps   []*ConfigMap   `json:"configmaps" description:"release configmaps"`
	DaemonSets   []*DaemonSet   `json:"daemonsets" description:"release daemonsets"`
	Deployments  []*Deployment  `json:"deployments" description:"release deployments"`
	Ingresses    []*Ingress     `json:"ingresses" description:"release ingresses"`
	Jobs         []*Job         `json:"jobs" description:"release jobs"`
	Secrets      []*Secret      `json:"secrets" description:"release secrets"`
	StatefulSets []*StatefulSet `json:"statefulsets" description:"release statefulsets"`
}

func (resourceSet *ResourceSet) GetPodsNeedRestart() []*Pod {
	pods := []*Pod{}
	for _, ds := range resourceSet.DaemonSets {
		if len(ds.Pods) > 0 {
			pods = append(pods, ds.Pods...)
		}
	}
	for _, ss := range resourceSet.StatefulSets {
		if len(ss.Pods) > 0 {
			pods = append(pods, ss.Pods...)
		}
	}
	for _, dp := range resourceSet.Deployments {
		if len(dp.Pods) > 0 {
			pods = append(pods, dp.Pods...)
		}
	}
	return pods
}

func (resourceSet *ResourceSet) IsReady() (bool, Resource) {
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

func NewResourceSet() *ResourceSet {
	return &ResourceSet{
		StatefulSets: []*StatefulSet{},
		Services:     []*Service{},
		Jobs:         []*Job{},
		Ingresses:    []*Ingress{},
		Deployments:  []*Deployment{},
		DaemonSets:   []*DaemonSet{},
		ConfigMaps:   []*ConfigMap{},
		Secrets:      []*Secret{},
	}
}