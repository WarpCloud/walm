package adaptor

import (
	"walm/pkg/k8s/handler"
)

var defaultAdaptorSet *AdaptorSet

func GetDefaultAdaptorSet() *AdaptorSet {
	if defaultAdaptorSet == nil {
		defaultAdaptorSet = &AdaptorSet{handlerSet: handler.GetDefaultHandlerSet()}
	}
	return defaultAdaptorSet
}

type AdaptorSet struct {
	handlerSet *handler.HandlerSet
	walmConfigMapAdaptor *WalmConfigMapAdaptor
	walmDaemonSetAdaptor * WalmDaemonSetAdaptor
	walmDeploymentAdaptor *WalmDeploymentAdaptor
	walmIngressAdaptor *WalmIngressAdaptor
	walmInstanceAdaptor *WalmInstanceAdaptor
	walmJobAdaptor *WalmJobAdaptor
	walmPodAdaptor *WalmPodAdaptor
	walmSecretAdaptor *WalmSecretAdaptor
	walmServiceAdaptor *WalmServiceAdaptor
	walmStatefulSetAdaptor *WalmStatefulSetAdaptor
	walmNodeAdaptor *WalmNodeAdaptor
	walmResourceQuotaAdaptor *WalmResourceQuotaAdaptor
}

func(set *AdaptorSet) GetHandlerSet() *handler.HandlerSet{
	return set.handlerSet
}

func(set *AdaptorSet) GetAdaptor(kind string) (resourceAdaptor ResourceAdaptor){

	switch kind {
	case "ApplicationInstance":
		if set.walmInstanceAdaptor == nil {
			set.walmInstanceAdaptor = &WalmInstanceAdaptor{set}
		}
		resourceAdaptor = set.walmInstanceAdaptor
	case "Deployment":
		if set.walmDeploymentAdaptor == nil {
			set.walmDeploymentAdaptor = &WalmDeploymentAdaptor{set.handlerSet.GetDeploymentHandler(), set.GetAdaptor("Pod").(*WalmPodAdaptor)}
		}
		resourceAdaptor = set.walmDeploymentAdaptor
	case "Service":
		if set.walmServiceAdaptor == nil {
			set.walmServiceAdaptor = &WalmServiceAdaptor{set.handlerSet.GetServiceHandler()}
		}
		resourceAdaptor = set.walmServiceAdaptor
	case "StatefulSet":
		if set.walmStatefulSetAdaptor == nil {
			set.walmStatefulSetAdaptor = &WalmStatefulSetAdaptor{set.handlerSet.GetStatefulSetHandler(),set.GetAdaptor("Pod").(*WalmPodAdaptor)}
		}
		resourceAdaptor = set.walmStatefulSetAdaptor
	case "DaemonSet":
		if set.walmDaemonSetAdaptor == nil {
			set.walmDaemonSetAdaptor = &WalmDaemonSetAdaptor{set.handlerSet.GetDaemonSetHandler(), set.GetAdaptor("Pod").(*WalmPodAdaptor)}
		}
		resourceAdaptor = set.walmDaemonSetAdaptor
	case "Job":
		if set.walmJobAdaptor == nil {
			set.walmJobAdaptor = &WalmJobAdaptor{set.handlerSet.GetJobHandler(),set.GetAdaptor("Pod").(*WalmPodAdaptor)}
		}
		resourceAdaptor = set.walmJobAdaptor
	case "ConfigMap":
		if set.walmConfigMapAdaptor == nil {
			set.walmConfigMapAdaptor = &WalmConfigMapAdaptor{set.handlerSet.GetConfigMapHandler()}
		}
		resourceAdaptor = set.walmConfigMapAdaptor
	case "Ingress":
		if set.walmIngressAdaptor == nil {
			set.walmIngressAdaptor = &WalmIngressAdaptor{set.handlerSet.GetIngressHandler()}
		}
		resourceAdaptor = set.walmIngressAdaptor
	case "Secret":
		if set.walmSecretAdaptor == nil {
			set.walmSecretAdaptor = &WalmSecretAdaptor{set.handlerSet.GetSecretHandler()}
		}
		resourceAdaptor = set.walmSecretAdaptor
	case "Pod":
		if set.walmPodAdaptor == nil {
			set.walmPodAdaptor = &WalmPodAdaptor{set.handlerSet.GetPodHandler()}
		}
		resourceAdaptor = set.walmPodAdaptor
	case "Node":
		if set.walmNodeAdaptor == nil {
			set.walmNodeAdaptor = &WalmNodeAdaptor{set.handlerSet.GetNodeHandler()}
		}
		resourceAdaptor = set.walmNodeAdaptor
	case "ResourceQuota":
		if set.walmResourceQuotaAdaptor == nil {
			set.walmResourceQuotaAdaptor = &WalmResourceQuotaAdaptor{set.handlerSet.GetResourceQuotaHandler()}
		}
		resourceAdaptor = set.walmResourceQuotaAdaptor
	default:
		resourceAdaptor = &WalmDefaultAdaptor{kind}
	}
	return

}






