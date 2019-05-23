package handler

import (
	"k8s.io/client-go/kubernetes"
	"WarpCloud/walm/pkg/k8s/informer"
	releaseconfigclientset "transwarp/release-config/pkg/client/clientset/versioned"
)

type HandlerSet struct {
	client *kubernetes.Clientset
	releaseConfigClient *releaseconfigclientset.Clientset
	factory *informer.InformerFactory
	configMapHandler *ConfigMapHandler
	daemonSetHandler *DaemonSetHandler
	deploymentHandler *DeploymentHandler
	eventHandler     *EventHandler
	ingressHandler *IngressHandler
	jobHandler *JobHandler
	namespaceHandler *NamespaceHandler
	nodeHandler *NodeHandler
	podHandler *PodHandler
	secretHandler *SecretHandler
	serviceHandler *ServiceHandler
	statefulSetHandler *StatefulSetHandler
	resourceQuotaHandler *ResourceQuotaHandler
	persistentVolumeClaimHandler *PersistentVolumeClaimHandler
	storageClassHandler *StorageClassHandler
	releaseConfigHandler *ReleaseConfigHandler
	endpointsHandler *EndpointsHandler
	limitRangeHandler *LimitRangeHandler
}

func (set *HandlerSet)GetConfigMapHandler() *ConfigMapHandler {
	if set.configMapHandler == nil {
		set.configMapHandler = &ConfigMapHandler{client: set.client, lister: set.factory.ConfigMapLister}
	}
	return set.configMapHandler
}

func (set *HandlerSet)GetDaemonSetHandler() *DaemonSetHandler {
	if set.daemonSetHandler == nil {
		set.daemonSetHandler = &DaemonSetHandler{client: set.client, lister: set.factory.DaemonSetLister}
	}
	return set.daemonSetHandler
}

func (set *HandlerSet)GetDeploymentHandler() *DeploymentHandler {
	if set.deploymentHandler == nil {
		set.deploymentHandler = &DeploymentHandler{client: set.client, lister: set.factory.DeploymentLister}
	}
	return set.deploymentHandler
}

func (set *HandlerSet)GetEventHandler() *EventHandler {
	if set.eventHandler == nil {
		set.eventHandler = &EventHandler{client: set.client}
	}
	return set.eventHandler
}

func (set *HandlerSet)GetIngressHandler() *IngressHandler {
	if set.ingressHandler == nil {
		set.ingressHandler = &IngressHandler{client: set.client, lister: set.factory.IngressLister}
	}
	return set.ingressHandler
}

func (set *HandlerSet)GetJobHandler() *JobHandler {
	if set.jobHandler == nil {
		set.jobHandler = &JobHandler{client: set.client, lister: set.factory.JobLister}
	}
	return set.jobHandler
}

func (set *HandlerSet)GetNamespaceHandler() *NamespaceHandler {
	if set.namespaceHandler == nil {
		set.namespaceHandler = &NamespaceHandler{client: set.client, lister: set.factory.NamespaceLister}
	}
	return set.namespaceHandler
}

func (set *HandlerSet)GetNodeHandler() *NodeHandler {
	if set.nodeHandler == nil {
		set.nodeHandler = &NodeHandler{client: set.client, lister: set.factory.NodeLister}
	}
	return set.nodeHandler
}

func (set *HandlerSet)GetPodHandler() *PodHandler {
	if set.podHandler == nil {
		set.podHandler = &PodHandler{client: set.client, lister: set.factory.PodLister}
	}
	return set.podHandler
}

func (set *HandlerSet)GetSecretHandler() *SecretHandler {
	if set.secretHandler == nil {
		set.secretHandler = &SecretHandler{client: set.client, lister: set.factory.SecretLister}
	}
	return set.secretHandler
}

func (set *HandlerSet)GetServiceHandler() *ServiceHandler {
	if set.serviceHandler == nil {
		set.serviceHandler = &ServiceHandler{client: set.client, lister: set.factory.ServiceLister}
	}
	return set.serviceHandler
}

func (set *HandlerSet)GetStatefulSetHandler() *StatefulSetHandler {
	if set.statefulSetHandler == nil {
		set.statefulSetHandler = &StatefulSetHandler{client: set.client, lister: set.factory.StatefulSetLister}
	}
	return set.statefulSetHandler
}

func (set *HandlerSet)GetResourceQuotaHandler() *ResourceQuotaHandler {
	if set.resourceQuotaHandler == nil {
		set.resourceQuotaHandler = &ResourceQuotaHandler{client: set.client, lister: set.factory.ResourceQuotaLister}
	}
	return set.resourceQuotaHandler
}

func (set *HandlerSet)GetPersistentVolumeClaimHandler() *PersistentVolumeClaimHandler {
	if set.persistentVolumeClaimHandler == nil {
		set.persistentVolumeClaimHandler = &PersistentVolumeClaimHandler{client: set.client, lister: set.factory.PersistentVolumeClaimLister}
	}
	return set.persistentVolumeClaimHandler
}

func (set *HandlerSet)GetStorageClassHandler() *StorageClassHandler {
	if set.storageClassHandler == nil {
		set.storageClassHandler = &StorageClassHandler{client: set.client, lister: set.factory.StorageClassLister}
	}
	return set.storageClassHandler
}

func (set *HandlerSet)GetReleaseConfigHandler() *ReleaseConfigHandler {
	if set.releaseConfigHandler == nil {
		set.releaseConfigHandler = &ReleaseConfigHandler{client: set.releaseConfigClient, lister: set.factory.ReleaseConfigLister}
	}
	return set.releaseConfigHandler
}

func (set *HandlerSet)GetEndpointsHandler() *EndpointsHandler {
	if set.endpointsHandler == nil {
		set.endpointsHandler = &EndpointsHandler{client: set.client, lister: set.factory.EndpointsLister}
	}
	return set.endpointsHandler
}

func (set *HandlerSet)GetLimitRangeHandler() *LimitRangeHandler {
	if set.limitRangeHandler == nil {
		set.limitRangeHandler = &LimitRangeHandler{client: set.client, lister: set.factory.LimitRangeLister}
	}
	return set.limitRangeHandler
}