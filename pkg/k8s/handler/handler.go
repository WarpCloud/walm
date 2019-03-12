package handler

import (
	"walm/pkg/k8s/client"
	"walm/pkg/k8s/informer"
	"k8s.io/client-go/kubernetes"
)

var handlerSets *HandlerSet

func GetDefaultHandlerSet() *HandlerSet {
	if handlerSets == nil {
		handlerSets = &HandlerSet{
			client: client.GetDefaultClient(),
			releaseConfigClient: client.GetDefaultReleaseConfigClient(),
			factory: informer.GetDefaultFactory(),
		}
	}
	return handlerSets
}

func NewFakeHandlerSet(client *kubernetes.Clientset, factory *informer.InformerFactory) *HandlerSet{
	return &HandlerSet{
		client: client,
		factory: factory,
	}
}
