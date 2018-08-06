package handler

import (
	"walm/pkg/k8s/client"
	"walm/pkg/k8s/informer"
	"k8s.io/client-go/kubernetes"
	"transwarp/application-instance/pkg/client/clientset/versioned"
)

var handlerSets *HandlerSet

func GetDefaultHandlerSet() *HandlerSet {
	if handlerSets == nil {
		handlerSets = &HandlerSet{
			client: client.GetDefaultClient(),
			clientEx: client.GetDefaultClientEx(),
			factory: informer.GetDefaultFactory(),
		}
	}
	return handlerSets
}

func NewFakeHandlerSet(client *kubernetes.Clientset, clientEx *versioned.Clientset, factory *informer.InformerFactory) *HandlerSet{
	return &HandlerSet{
		client: client,
		clientEx: clientEx,
		factory: factory,
	}
}
