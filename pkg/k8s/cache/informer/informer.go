package informer

import (
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"time"
	listv1beta1 "k8s.io/client-go/listers/extensions/v1beta1"
	"k8s.io/client-go/listers/core/v1"
	batchv1 "k8s.io/client-go/listers/batch/v1"
	"k8s.io/client-go/listers/apps/v1beta1"
	storagev1 "k8s.io/client-go/listers/storage/v1"
	releaseconfigexternalversions "transwarp/release-config/pkg/client/informers/externalversions"
	releaseconfigv1beta1 "transwarp/release-config/pkg/client/listers/transwarp/v1beta1"
	releaseconfigclientset "transwarp/release-config/pkg/client/clientset/versioned"
	"WarpCloud/walm/pkg/models/k8s"
	"github.com/sirupsen/logrus"
	"WarpCloud/walm/pkg/k8s/converter"
	errorModel "WarpCloud/walm/pkg/models/error"
	"WarpCloud/walm/pkg/models/release"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	"sort"
	"WarpCloud/walm/pkg/k8s/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sync"
	"errors"
)

type Informer struct {
	client                      *kubernetes.Clientset
	factory                     informers.SharedInformerFactory
	deploymentLister            listv1beta1.DeploymentLister
	configMapLister             v1.ConfigMapLister
	daemonSetLister             listv1beta1.DaemonSetLister
	ingressLister               listv1beta1.IngressLister
	jobLister                   batchv1.JobLister
	podLister                   v1.PodLister
	secretLister                v1.SecretLister
	serviceLister               v1.ServiceLister
	statefulSetLister           v1beta1.StatefulSetLister
	nodeLister                  v1.NodeLister
	namespaceLister             v1.NamespaceLister
	resourceQuotaLister         v1.ResourceQuotaLister
	persistentVolumeClaimLister v1.PersistentVolumeClaimLister
	storageClassLister          storagev1.StorageClassLister
	endpointsLister             v1.EndpointsLister
	limitRangeLister            v1.LimitRangeLister

	releaseConifgFactory releaseconfigexternalversions.SharedInformerFactory
	releaseConfigLister  releaseconfigv1beta1.ReleaseConfigLister
}

func (informer *Informer)ListStorageClasses(namespace string, labelSelectorStr string) ([]*k8s.StorageClass, error) {
	selector, err := labels.Parse(labelSelectorStr)
	if err != nil {
		logrus.Errorf("failed to parse label string %s : %s", labelSelectorStr, err.Error())
		return nil, err
	}

	resources, err := informer.storageClassLister.List(selector)
	if err != nil {
		logrus.Errorf("failed to list storage classes in namespace %s : %s", namespace, err.Error())
		return nil, err
	}

	storageClasses := []*k8s.StorageClass{}
	for _, resource := range resources {
		storageClass, err := converter.ConvertStorageClassFromK8s(resource)
		if err != nil {
			logrus.Errorf("failed to convert storageClass %s/%s: %s", resource.Namespace, resource.Name, err.Error())
			return nil, err
		}
		storageClasses = append(storageClasses, storageClass)
	}
	return storageClasses, nil
}

func (informer *Informer) GetPodLogs(namespace string, podName string, containerName string, tailLines int64) (string, error) {
	podLogOptions := &corev1.PodLogOptions{}
	if containerName != "" {
		podLogOptions.Container = containerName
	}
	if tailLines != 0 {
		podLogOptions.TailLines = &tailLines
	}
	logs, err := informer.client.CoreV1().Pods(namespace).GetLogs(podName, podLogOptions).Do().Raw()
	if err != nil {
		logrus.Errorf("failed to get pod logs : %s", err.Error())
		return "", err
	}
	return string(logs), nil
}

func (informer *Informer) GetPodEventList(namespace string, name string) (*k8s.EventList, error) {
	pod, err := informer.podLister.Pods(namespace).Get(name)
	if err != nil {
		logrus.Errorf("failed to get pod : %s", err.Error())
		return nil, err
	}

	ref := &corev1.ObjectReference{
		Namespace:       pod.Namespace,
		Name:            pod.Name,
		Kind:            pod.Kind,
		ResourceVersion: pod.ResourceVersion,
		UID:             pod.UID,
		APIVersion:      pod.APIVersion,
	}

	podEvents, err := informer.searchEvents(pod.Namespace, ref)
	if err != nil {
		logrus.Errorf("failed to get Events : %s", err.Error())
		return nil, err
	}
	sort.Sort(utils.SortableEvents(podEvents.Items))

	walmEvents := []k8s.Event{}
	for _, event := range podEvents.Items {
		walmEvent := k8s.Event{
			Type:           event.Type,
			Reason:         event.Reason,
			Message:        event.Message,
			Count:          event.Count,
			FirstTimestamp: event.FirstTimestamp.String(),
			LastTimestamp:  event.LastTimestamp.String(),
			From:           utils.FormatEventSource(event.Source),
		}
		walmEvents = append(walmEvents, walmEvent)
	}
	return &k8s.EventList{Events: walmEvents}, nil
}

func (informer *Informer) ListSecrets(namespace string, labelSelectorStr string) (*k8s.SecretList, error) {
	selector, err := labels.Parse(labelSelectorStr)
	if err != nil {
		logrus.Errorf("failed to parse label string %s : %s", labelSelectorStr, err.Error())
		return nil, err
	}

	resources, err := informer.secretLister.Secrets(namespace).List(selector)
	if err != nil {
		logrus.Errorf("failed to list secrets in namespace %s : %s", namespace, err.Error())
		return nil, err
	}

	secrets := []*k8s.Secret{}
	for _, resource := range resources {
		secret, err := converter.ConvertSecretFromK8s(resource)
		if err != nil {
			logrus.Errorf("failed to convert secret %s/%s: %s", resource.Namespace, resource.Name, err.Error())
			return nil, err
		}
		secrets = append(secrets, secret)
	}
	return &k8s.SecretList{
		Num:   len(secrets),
		Items: secrets,
	}, nil
}

func (informer *Informer) ListStatefulSets(namespace string, labelSelectorStr string) ([]*k8s.StatefulSet, error) {
	selector, err := labels.Parse(labelSelectorStr)
	if err != nil {
		logrus.Errorf("failed to parse label string %s : %s", labelSelectorStr, err.Error())
		return nil, err
	}
	resources, err := informer.statefulSetLister.StatefulSets(namespace).List(selector)
	if err != nil {
		logrus.Errorf("failed to list stateful sets in namespace %s : %s", namespace, err.Error())
		return nil, err
	}

	statefulSets := []*k8s.StatefulSet{}
	for _, resource := range resources {
		pods, err := informer.listPods(namespace, resource.Spec.Selector)
		if err != nil {
			return nil, err
		}
		statefulSet, err := converter.ConvertStatefulSetFromK8s(resource, pods)
		if err != nil {
			logrus.Errorf("failed to convert stateful set %s/%s: %s", resource.Namespace, resource.Name, err.Error())
			return nil, err
		}
		statefulSets = append(statefulSets, statefulSet)
	}
	return statefulSets, nil
}

func (informer *Informer) GetNodes(labelSelectorStr string) ([]*k8s.Node, error) {
	selector, err := labels.Parse(labelSelectorStr)
	if err != nil {
		logrus.Errorf("failed to parse label string %s : %s", labelSelectorStr, err.Error())
		return nil, err
	}
	nodeList, err := informer.nodeLister.List(selector)
	if err != nil {
		return nil, err
	}

	walmNodes := []*k8s.Node{}
	if nodeList != nil {
		mux := &sync.Mutex{}
		var wg sync.WaitGroup
		for _, node := range nodeList {
			wg.Add(1)
			go func(node *corev1.Node) {
				defer wg.Done()
				podsOnNode, err1 := informer.getNonTermiatedPodsOnNode(node.Name, nil)
				if err1 != nil {
					logrus.Errorf("failed to get pods on node: %s", err1.Error())
					err = errors.New(err1.Error())
					return
				}
				walmNode, err1 := converter.ConvertNodeFromK8s(node, podsOnNode)
				if err1 != nil {
					logrus.Errorf("failed to build walm node : %s", err1.Error())
					err = errors.New(err1.Error())
					return
				}

				mux.Lock()
				walmNodes = append(walmNodes, walmNode)
				mux.Unlock()
			}(node)
		}
		wg.Wait()
		if err != nil {
			logrus.Errorf("failed to build nodes : %s", err.Error())
			return nil, err
		}
	}

	return walmNodes, nil
}

func (informer *Informer) AddReleaseConfigHandler(OnAdd func(obj interface{}), OnUpdate func(oldObj, newObj interface{}), OnDelete func(obj interface{})) {
	handlerFuncs := &cache.ResourceEventHandlerFuncs{
		AddFunc:    OnAdd,
		UpdateFunc: OnUpdate,
		DeleteFunc: OnDelete,
	}
	informer.releaseConifgFactory.Transwarp().V1beta1().ReleaseConfigs().Informer().AddEventHandler(handlerFuncs)
}

func (informer *Informer) ListPersistentVolumeClaims(namespace string, labelSelectorStr string) ([]*k8s.PersistentVolumeClaim, error) {
	selector, err := labels.Parse(labelSelectorStr)
	if err != nil {
		logrus.Errorf("failed to parse label string %s : %s", labelSelectorStr, err.Error())
		return nil, err
	}
	resources, err := informer.persistentVolumeClaimLister.PersistentVolumeClaims(namespace).List(selector)
	if err != nil {
		logrus.Errorf("failed to list pvcs in namespace %s : %s", namespace, err.Error())
		return nil, err
	}

	pvcs := []*k8s.PersistentVolumeClaim{}
	for _, resource := range resources {
		pvc, err := converter.ConvertPvcFromK8s(resource)
		if err != nil {
			logrus.Errorf("failed to convert release config %s/%s: %s", resource.Namespace, resource.Name, err.Error())
			return nil, err
		}
		pvcs = append(pvcs, pvc)
	}
	return pvcs, nil
}

func (informer *Informer) ListReleaseConfigs(namespace, labelSelectorStr string) ([]*k8s.ReleaseConfig, error) {
	selector, err := labels.Parse(labelSelectorStr)
	if err != nil {
		logrus.Errorf("failed to parse label string %s : %s", labelSelectorStr, err.Error())
		return nil, err
	}
	resources, err := informer.releaseConfigLister.ReleaseConfigs(namespace).List(selector)
	if err != nil {
		logrus.Errorf("failed to list release configs in namespace %s : %s", namespace, err.Error())
		return nil, err
	}

	releaseConfigs := []*k8s.ReleaseConfig{}
	for _, resource := range resources {
		releaseConfig, err := converter.ConvertReleaseConfigFromK8s(resource)
		if err != nil {
			logrus.Errorf("failed to convert release config %s/%s: %s", resource.Namespace, resource.Name, err.Error())
			return nil, err
		}
		releaseConfigs = append(releaseConfigs, releaseConfig)
	}
	return releaseConfigs, nil
}

func (informer *Informer) GetResourceSet(releaseResourceMetas []release.ReleaseResourceMeta) (resourceSet *k8s.ResourceSet, err error) {
	resourceSet = k8s.NewResourceSet()
	for _, resourceMeta := range releaseResourceMetas {
		resource, err := informer.GetResource(resourceMeta.Kind, resourceMeta.Namespace, resourceMeta.Name)
		// if resource is not found , do not return error, add it into resource set, so resource should not be nil
		if err != nil && !errorModel.IsNotFoundError(err) {
			return nil, err
		}
		resource.AddToResourceSet(resourceSet)
	}
	return
}

func (informer *Informer) GetResource(kind k8s.ResourceKind, namespace, name string) (k8s.Resource, error) {
	switch kind {
	case k8s.ReleaseConfigKind:
		return informer.getReleaseConfig(namespace, name)
	case k8s.ConfigMapKind:
		return informer.getConfigMap(namespace, name)
	case k8s.PersistentVolumeClaimKind:
		return informer.getPvc(namespace, name)
	case k8s.DaemonSetKind:
		return informer.getDaemonSet(namespace, name)
	case k8s.DeploymentKind:
		return informer.getDeployment(namespace, name)
	case k8s.ServiceKind:
		return informer.getService(namespace, name)
	case k8s.StatefulSetKind:
		return informer.getStatefulSet(namespace, name)
	case k8s.JobKind:
		return informer.getJob(namespace, name)
	case k8s.IngressKind:
		return informer.getIngress(namespace, name)
	case k8s.SecretKind:
		return informer.getSecret(namespace, name)
	case k8s.NodeKind:
		return informer.getNode(namespace, name)
	case k8s.StorageClassKind:
		return informer.getStorageClass(namespace, name)
	default:
		return &k8s.DefaultResource{Meta: k8s.NewMeta(kind, namespace, name, k8s.NewState("Unknown", "NotSupportedKind", "Can not get this resource"))}, nil
	}
}

func (informer *Informer) start(stopCh <-chan struct{}) {
	informer.factory.Start(stopCh)
	informer.releaseConifgFactory.Start(stopCh)
}

func (informer *Informer) waitForCacheSync(stopCh <-chan struct{}) {
	informer.factory.WaitForCacheSync(stopCh)
	informer.releaseConifgFactory.WaitForCacheSync(stopCh)
}

func (informer *Informer) searchEvents(namespace string, objOrRef runtime.Object) (*corev1.EventList, error) {
	return informer.client.CoreV1().Events(namespace).Search(runtime.NewScheme(), objOrRef)
}

func NewInformer(client *kubernetes.Clientset, releaseConfigClient *releaseconfigclientset.Clientset, resyncPeriod time.Duration, stopCh <-chan struct{}) (*Informer) {
	informer := &Informer{}
	informer.client = client
	informer.factory = informers.NewSharedInformerFactory(client, resyncPeriod)
	informer.deploymentLister = informer.factory.Extensions().V1beta1().Deployments().Lister()
	informer.configMapLister = informer.factory.Core().V1().ConfigMaps().Lister()
	informer.daemonSetLister = informer.factory.Extensions().V1beta1().DaemonSets().Lister()
	informer.ingressLister = informer.factory.Extensions().V1beta1().Ingresses().Lister()
	informer.jobLister = informer.factory.Batch().V1().Jobs().Lister()
	informer.podLister = informer.factory.Core().V1().Pods().Lister()
	informer.secretLister = informer.factory.Core().V1().Secrets().Lister()
	informer.serviceLister = informer.factory.Core().V1().Services().Lister()
	informer.statefulSetLister = informer.factory.Apps().V1beta1().StatefulSets().Lister()
	informer.nodeLister = informer.factory.Core().V1().Nodes().Lister()
	informer.namespaceLister = informer.factory.Core().V1().Namespaces().Lister()
	informer.resourceQuotaLister = informer.factory.Core().V1().ResourceQuotas().Lister()
	informer.persistentVolumeClaimLister = informer.factory.Core().V1().PersistentVolumeClaims().Lister()
	informer.storageClassLister = informer.factory.Storage().V1().StorageClasses().Lister()
	informer.endpointsLister = informer.factory.Core().V1().Endpoints().Lister()
	informer.limitRangeLister = informer.factory.Core().V1().LimitRanges().Lister()

	informer.releaseConifgFactory = releaseconfigexternalversions.NewSharedInformerFactory(releaseConfigClient, resyncPeriod)
	informer.releaseConfigLister = informer.releaseConifgFactory.Transwarp().V1beta1().ReleaseConfigs().Lister()

	informer.start(stopCh)
	informer.waitForCacheSync(stopCh)
	logrus.Info("k8s cache sync finished")
	return informer
}
