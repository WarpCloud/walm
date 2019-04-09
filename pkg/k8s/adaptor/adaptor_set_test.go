package adaptor

import (
	"testing"
	"walm/pkg/k8s/handler"
	"github.com/stretchr/testify/assert"
	"walm/pkg/k8s/informer"
)

func Test(t *testing.T) {
	tests := []struct{
		kind string
		adaptor ResourceAdaptor
	} {
		{kind: "Deployment", adaptor: &WalmDeploymentAdaptor{}},
		{kind: "Service", adaptor: &WalmServiceAdaptor{}},
		{kind: "StatefulSet", adaptor: &WalmStatefulSetAdaptor{}},
		{kind: "DaemonSet", adaptor: &WalmDaemonSetAdaptor{}},
		{kind: "Job", adaptor: &WalmJobAdaptor{}},
		{kind: "ConfigMap", adaptor: &WalmConfigMapAdaptor{}},
		{kind: "Ingress", adaptor: &WalmIngressAdaptor{}},
		{kind: "Secret", adaptor: &WalmSecretAdaptor{}},
		{kind: "Pod", adaptor: &WalmPodAdaptor{}},
		{kind: "Node", adaptor: &WalmNodeAdaptor{}},
		{kind: "ResourceQuota", adaptor: &WalmResourceQuotaAdaptor{}},
		{kind: "PersistentVolumeClaim", adaptor: &WalmPersistentVolumeClaimAdaptor{}},
		{kind: "StorageClass", adaptor: &WalmStorageClassAdaptor{}},
		{kind: "Namespace", adaptor: &WalmNamespaceAdaptor{}},
		{kind: "UnKnown", adaptor: &WalmDefaultAdaptor{}},
	}

	adaptorSet := AdaptorSet{handlerSet: handler.NewFakeHandlerSet(nil, nil, informer.NewFakeInformerFactory(nil, nil, 0))}

	for _, test := range tests {
		adaptor := adaptorSet.GetAdaptor(test.kind)
		assert.IsType(t, test.adaptor, adaptor)
	}
}
