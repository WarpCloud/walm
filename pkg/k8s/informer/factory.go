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
	"github.com/sirupsen/logrus"
	storagev1 "k8s.io/client-go/listers/storage/v1"
	releaseconfigexternalversions "transwarp/release-config/pkg/client/informers/externalversions"
	releaseconfigv1beta1 "transwarp/release-config/pkg/client/listers/transwarp/v1beta1"
	releaseconfigclientset "transwarp/release-config/pkg/client/clientset/versioned"
)

var defaultFactory *InformerFactory

func StartInformer(stopCh <-chan struct{}) {
	defaultFactory = newInformerFactory(client.GetDefaultClient(), client.GetDefaultClientEx(), client.GetDefaultReleaseConfigClient(), 0)
	defaultFactory.Start(stopCh)
	defaultFactory.WaitForCacheSync(stopCh)
	logrus.Info("informer started")
}

func GetDefaultFactory() *InformerFactory {
	return defaultFactory
}

type InformerFactory struct {
	Factory                     informers.SharedInformerFactory
	DeploymentLister            listv1beta1.DeploymentLister
	ConfigMapLister             v1.ConfigMapLister
	DaemonSetLister             listv1beta1.DaemonSetLister
	IngressLister               listv1beta1.IngressLister
	JobLister                   batchv1.JobLister
	PodLister                   v1.PodLister
	SecretLister                v1.SecretLister
	ServiceLister               v1.ServiceLister
	StatefulSetLister           v1beta1.StatefulSetLister
	NodeLister                  v1.NodeLister
	NamespaceLister             v1.NamespaceLister
	ResourceQuotaLister         v1.ResourceQuotaLister
	PersistentVolumeClaimLister v1.PersistentVolumeClaimLister
	StorageClassLister          storagev1.StorageClassLister

	factoryEx      externalversions.SharedInformerFactory
	InstanceLister tranv1beta1.ApplicationInstanceLister

	ReleaseConifgFactory releaseconfigexternalversions.SharedInformerFactory
	ReleaseConfigLister  releaseconfigv1beta1.ReleaseConfigLister
}

func (factory *InformerFactory) Start(stopCh <-chan struct{}) {
	factory.Factory.Start(stopCh)
	factory.factoryEx.Start(stopCh)
	factory.ReleaseConifgFactory.Start(stopCh)
}

func (factory *InformerFactory) WaitForCacheSync(stopCh <-chan struct{}) {
	factory.Factory.WaitForCacheSync(stopCh)
	factory.factoryEx.WaitForCacheSync(stopCh)
	factory.ReleaseConifgFactory.WaitForCacheSync(stopCh)
}

func newInformerFactory(client *kubernetes.Clientset, clientEx *clientsetex.Clientset, releaseConfigClient *releaseconfigclientset.Clientset, resyncPeriod time.Duration) (*InformerFactory) {
	factory := &InformerFactory{}
	factory.Factory = informers.NewSharedInformerFactory(client, resyncPeriod)
	factory.DeploymentLister = factory.Factory.Extensions().V1beta1().Deployments().Lister()
	factory.ConfigMapLister = factory.Factory.Core().V1().ConfigMaps().Lister()
	factory.DaemonSetLister = factory.Factory.Extensions().V1beta1().DaemonSets().Lister()
	factory.IngressLister = factory.Factory.Extensions().V1beta1().Ingresses().Lister()
	factory.JobLister = factory.Factory.Batch().V1().Jobs().Lister()
	factory.PodLister = factory.Factory.Core().V1().Pods().Lister()
	factory.SecretLister = factory.Factory.Core().V1().Secrets().Lister()
	factory.ServiceLister = factory.Factory.Core().V1().Services().Lister()
	factory.StatefulSetLister = factory.Factory.Apps().V1beta1().StatefulSets().Lister()
	factory.NodeLister = factory.Factory.Core().V1().Nodes().Lister()
	factory.NamespaceLister = factory.Factory.Core().V1().Namespaces().Lister()
	factory.ResourceQuotaLister = factory.Factory.Core().V1().ResourceQuotas().Lister()
	factory.PersistentVolumeClaimLister = factory.Factory.Core().V1().PersistentVolumeClaims().Lister()
	factory.StorageClassLister = factory.Factory.Storage().V1().StorageClasses().Lister()

	factory.factoryEx = externalversions.NewSharedInformerFactory(clientEx, resyncPeriod)
	factory.InstanceLister = factory.factoryEx.Transwarp().V1beta1().ApplicationInstances().Lister()

	factory.ReleaseConifgFactory = releaseconfigexternalversions.NewSharedInformerFactory(releaseConfigClient, resyncPeriod)
	factory.ReleaseConfigLister = factory.ReleaseConifgFactory.Transwarp().V1beta1().ReleaseConfigs().Lister()
	return factory
}

// for test
func NewFakeInformerFactory(client *kubernetes.Clientset, clientEx *clientsetex.Clientset, releaseConfigClient *releaseconfigclientset.Clientset, resyncPeriod time.Duration) (*InformerFactory) {
	return newInformerFactory(client, clientEx, releaseConfigClient, resyncPeriod)
}
