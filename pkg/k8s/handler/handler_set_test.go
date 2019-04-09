package handler

import (
	"testing"
	"walm/pkg/k8s/informer"
	"github.com/stretchr/testify/assert"
)

func Test_HandlerSet(t *testing.T) {
	handlerSet := NewFakeHandlerSet(nil, nil, informer.NewFakeInformerFactory(nil, nil, 0))
	assert.NotNil(t, handlerSet.GetReleaseConfigHandler())
	assert.NotNil(t, handlerSet.GetEndpointsHandler())
	assert.NotNil(t, handlerSet.GetStorageClassHandler())
	assert.NotNil(t, handlerSet.GetPersistentVolumeClaimHandler())
	assert.NotNil(t, handlerSet.GetStatefulSetHandler())
	assert.NotNil(t, handlerSet.GetResourceQuotaHandler())
	assert.NotNil(t, handlerSet.GetNamespaceHandler())
	assert.NotNil(t, handlerSet.GetServiceHandler())
	assert.NotNil(t, handlerSet.GetPodHandler())
	assert.NotNil(t, handlerSet.GetEventHandler())
	assert.NotNil(t, handlerSet.GetSecretHandler())
	assert.NotNil(t, handlerSet.GetNodeHandler())
	assert.NotNil(t, handlerSet.GetIngressHandler())
	assert.NotNil(t, handlerSet.GetConfigMapHandler())
	assert.NotNil(t, handlerSet.GetJobHandler())
	assert.NotNil(t, handlerSet.GetDaemonSetHandler())
	assert.NotNil(t, handlerSet.GetDeploymentHandler())
}
