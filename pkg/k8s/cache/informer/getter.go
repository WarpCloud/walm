package informer

import (
	"WarpCloud/walm/pkg/models/k8s"
	"WarpCloud/walm/pkg/k8s/utils"
	"github.com/sirupsen/logrus"
	"WarpCloud/walm/pkg/k8s/converter"
	errorModel "WarpCloud/walm/pkg/models/error"
)

func (informer *Informer) getReleaseConfig(namespace, name string) (k8s.Resource, error) {
	resource, err := informer.releaseConfigLister.ReleaseConfigs(namespace).Get(name)
	if err != nil {
		return convertResourceError(err, &k8s.ReleaseConfig{
			Meta: k8s.NewNotFoundMeta(k8s.ReleaseConfigKind, namespace, name),
		})
	}
	return converter.ConvertReleaseConfigFromK8s(resource)
}

func (informer *Informer) getConfigMap(namespace, name string) (k8s.Resource, error) {
	resource, err := informer.configMapLister.ConfigMaps(namespace).Get(name)
	if err != nil {
		return convertResourceError(err, &k8s.ConfigMap{
			Meta: k8s.NewNotFoundMeta(k8s.ConfigMapKind, namespace, name),
		})
	}
	return converter.ConvertConfigMapFromK8s(resource)
}

func (informer *Informer) getPvc(namespace, name string) (k8s.Resource, error) {
	resource, err := informer.persistentVolumeClaimLister.PersistentVolumeClaims(namespace).Get(name)
	if err != nil {
		return convertResourceError(err, &k8s.PersistentVolumeClaim{
			Meta: k8s.NewNotFoundMeta(k8s.PersistentVolumeClaimKind, namespace, name),
		})
	}
	return converter.ConvertPvcFromK8s(resource)
}

func (informer *Informer) getDaemonSet(namespace, name string) (k8s.Resource, error) {
	resource, err := informer.daemonSetLister.DaemonSets(namespace).Get(name)
	if err != nil {
		return convertResourceError(err, &k8s.DaemonSet{
			Meta: k8s.NewNotFoundMeta(k8s.DaemonSetKind, namespace, name),
		})
	}
	pods, err := informer.listPods(namespace, resource.Spec.Selector)
	if err != nil {
		return nil, err
	}
	return converter.ConvertDaemonSetFromK8s(resource, pods)
}

func (informer *Informer) getDeployment(namespace, name string) (k8s.Resource, error) {
	resource, err := informer.deploymentLister.Deployments(namespace).Get(name)
	if err != nil {
		return convertResourceError(err, &k8s.Deployment{
			Meta: k8s.NewNotFoundMeta(k8s.DeploymentKind, namespace, name),
		})
	}
	pods, err := informer.listPods(namespace, resource.Spec.Selector)
	if err != nil {
		return nil, err
	}
	return converter.ConvertDeploymentFromK8s(resource, pods)
}

func (informer *Informer) getIngress(namespace, name string) (k8s.Resource, error) {
	resource, err := informer.ingressLister.Ingresses(namespace).Get(name)
	if err != nil {
		return convertResourceError(err, &k8s.Ingress{
			Meta: k8s.NewNotFoundMeta(k8s.IngressKind, namespace, name),
		})
	}
	return converter.ConvertIngressFromK8s(resource)
}

func (informer *Informer) getJob(namespace, name string) (k8s.Resource, error) {
	resource, err := informer.jobLister.Jobs(namespace).Get(name)
	if err != nil {
		return convertResourceError(err, &k8s.Job{
			Meta: k8s.NewNotFoundMeta(k8s.JobKind, namespace, name),
		})
	}
	pods, err := informer.listPods(namespace, resource.Spec.Selector)
	if err != nil {
		return nil, err
	}
	return converter.ConvertJobFromK8s(resource, pods)
}

func (informer *Informer) getSecret(namespace, name string) (k8s.Resource, error) {
	resource, err := informer.secretLister.Secrets(namespace).Get(name)
	if err != nil {
		return convertResourceError(err, &k8s.Secret{
			Meta: k8s.NewNotFoundMeta(k8s.SecretKind, namespace, name),
		})
	}
	return converter.ConvertSecretFromK8s(resource)
}

func (informer *Informer) getService(namespace, name string) (k8s.Resource, error) {
	resource, err := informer.serviceLister.Services(namespace).Get(name)
	if err != nil {
		return convertResourceError(err, &k8s.Service{
			Meta: k8s.NewNotFoundMeta(k8s.ServiceKind, namespace, name),
		})
	}

	endpoints, err := informer.getEndpoints(namespace, name)
	if err != nil {
		return nil, err
	}
	return converter.ConvertServiceFromK8s(resource, endpoints)
}

func (informer *Informer) getStatefulSet(namespace, name string) (k8s.Resource, error) {
	resource, err := informer.statefulSetLister.StatefulSets(namespace).Get(name)
	if err != nil {
		return convertResourceError(err, &k8s.StatefulSet{
			Meta: k8s.NewNotFoundMeta(k8s.StatefulSetKind, namespace, name),
		})
	}
	pods, err := informer.listPods(namespace, resource.Spec.Selector)
	if err != nil {
		return nil, err
	}
	return converter.ConvertStatefulSetFromK8s(resource, pods)
}

func convertResourceError(err error, notFoundResource k8s.Resource) (k8s.Resource, error) {
	if utils.IsK8sResourceNotFoundErr(err) {
		logrus.Warnf(" %s %s/%s is not found", notFoundResource.GetKind(), notFoundResource.GetNamespace(), notFoundResource.GetName())
		return notFoundResource, errorModel.NotFoundError{}
	}
	logrus.Errorf("failed to get %s %s/%s : %s", notFoundResource.GetKind(), notFoundResource.GetNamespace(), notFoundResource.GetName(), err.Error())
	return nil, err
}
