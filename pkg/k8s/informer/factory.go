package informer

import (
	"k8s.io/client-go/informers"
	"transwarp/application-instance/pkg/client/informers/externalversions"
	"k8s.io/client-go/kubernetes"
	clientsetex "transwarp/application-instance/pkg/client/clientset/versioned"
	"time"
	listv1beta1 "k8s.io/client-go/listers/extensions/v1beta1"
	tranv1beta1 "transwarp/application-instance/pkg/client/listers/transwarp/v1beta1"
	"k8s.io/client-go/listers/core/v1"
	batchv1 "k8s.io/client-go/listers/batch/v1"
	"k8s.io/client-go/listers/apps/v1beta1"
	"walm/pkg/k8s/client"
	"k8s.io/apimachinery/pkg/util/wait"
)

var defaultFactory *InformerFactory

func InitInformer() {
	defaultFactory = newInformerFactory(client.GetDefaultClient(), client.GetDefaultClientEx(), 0)
	defaultFactory.Start(wait.NeverStop)
	defaultFactory.WaitForCacheSync(wait.NeverStop)
}

func GetDefaultFactory() *InformerFactory {
	return defaultFactory
}

type InformerFactory struct {
	factory           informers.SharedInformerFactory
	DeploymentLister  listv1beta1.DeploymentLister
	ConfigMapLister   v1.ConfigMapLister
	DaemonSetLister   listv1beta1.DaemonSetLister
	IngressLister     listv1beta1.IngressLister
	JobLister         batchv1.JobLister
	PodLister         v1.PodLister
	SecretLister      v1.SecretLister
	ServiceLister     v1.ServiceLister
	StatefulSetLister v1beta1.StatefulSetLister
	NodeLister        v1.NodeLister
	NamespaceLister   v1.NamespaceLister

	factoryEx      externalversions.SharedInformerFactory
	InstanceLister tranv1beta1.ApplicationInstanceLister
}

func (factory *InformerFactory) Start(stopCh <-chan struct{}) {
	factory.factory.Start(stopCh)
	factory.factoryEx.Start(stopCh)
}

func (factory *InformerFactory) WaitForCacheSync(stopCh <-chan struct{}) {
	factory.factory.WaitForCacheSync(stopCh)
	factory.factoryEx.WaitForCacheSync(stopCh)
}

func newInformerFactory(client *kubernetes.Clientset, clientEx *clientsetex.Clientset, resyncPeriod time.Duration) (*InformerFactory) {
	factory := &InformerFactory{}
	factory.factory =  informers.NewSharedInformerFactory(client, resyncPeriod)
	factory.DeploymentLister = factory.factory.Extensions().V1beta1().Deployments().Lister()
	factory.ConfigMapLister = factory.factory.Core().V1().ConfigMaps().Lister()
	factory.DaemonSetLister = factory.factory.Extensions().V1beta1().DaemonSets().Lister()
	factory.IngressLister = factory.factory.Extensions().V1beta1().Ingresses().Lister()
	factory.JobLister = factory.factory.Batch().V1().Jobs().Lister()
	factory.PodLister = factory.factory.Core().V1().Pods().Lister()
	factory.SecretLister = factory.factory.Core().V1().Secrets().Lister()
	factory.ServiceLister = factory.factory.Core().V1().Services().Lister()
	factory.StatefulSetLister = factory.factory.Apps().V1beta1().StatefulSets().Lister()
	factory.NodeLister = factory.factory.Core().V1().Nodes().Lister()
	factory.NamespaceLister = factory.factory.Core().V1().Namespaces().Lister()

	factory.factoryEx = externalversions.NewSharedInformerFactory(clientEx, resyncPeriod)
	factory.InstanceLister = factory.factoryEx.Transwarp().V1beta1().ApplicationInstances().Lister()
	return factory
}

// for test
func NewFakeInformerFactory(client *kubernetes.Clientset, clientEx *clientsetex.Clientset, resyncPeriod time.Duration) (*InformerFactory) {
	return newInformerFactory(client, clientEx, resyncPeriod)
}