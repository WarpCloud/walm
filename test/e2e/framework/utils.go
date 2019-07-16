package framework

import (
	"strings"
	"fmt"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	"os"
	"WarpCloud/walm/pkg/util/transwarpjsonnet"
	"k8s.io/helm/pkg/chart/loader"
	"errors"
	"WarpCloud/walm/pkg/setting"
	"github.com/sirupsen/logrus"
	"WarpCloud/walm/pkg/k8s/client"
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clienthelm "WarpCloud/walm/pkg/k8s/client/helm"
	appsv1 "k8s.io/api/apps/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	releaseconfigclientset "transwarp/release-config/pkg/client/clientset/versioned"
	"transwarp/release-config/pkg/apis/transwarp/v1beta1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var k8sClient *kubernetes.Clientset
var k8sReleaseConfigClient *releaseconfigclientset.Clientset
var kubeClients *clienthelm.Client

const (
	maxNameLength                = 62
	randomLength                 = 5
	maxGeneratedRandomNameLength = maxNameLength - randomLength
)

func GetKubeClient() *clienthelm.Client {
	return kubeClients
}

func GetK8sClient() *kubernetes.Clientset {
	return k8sClient
}

func GetK8sReleaseConfigClient() *releaseconfigclientset.Clientset {
	return k8sReleaseConfigClient
}

func GenerateRandomName(base string) string {
	if len(base) > maxGeneratedRandomNameLength {
		base = base[:maxGeneratedRandomNameLength]
	}
	return fmt.Sprintf("%s-%s", strings.ToLower(base), utilrand.String(randomLength))
}

func CreateRandomNamespace(base string) (string, error) {
	namespace := GenerateRandomName(base)
	ns := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := k8sClient.CoreV1().Namespaces().Create(&ns)
	return namespace, err
}

func DeleteNamespace(namespace string) (error) {
	return k8sClient.CoreV1().Namespaces().Delete(namespace, &metav1.DeleteOptions{})
}

func GetNamespace(namespace string) (*v1.Namespace, error) {
	return k8sClient.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
}

func GetLimitRange(namespace, name string) (*v1.LimitRange, error) {
	return k8sClient.CoreV1().LimitRanges(namespace).Get(name, metav1.GetOptions{})
}

func GetResourceQuota(namespace, name string) (*v1.ResourceQuota, error) {
	return k8sClient.CoreV1().ResourceQuotas(namespace).Get(name, metav1.GetOptions{})
}

func GetTestNode() (*v1.Node, error) {
	nodeList, err := k8sClient.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	if len(nodeList.Items) == 0 {
		return nil, errors.New("there is none node")
	}
	return &nodeList.Items[0], nil
}

func GetNode(name string) (*v1.Node, error) {
	return k8sClient.CoreV1().Nodes().Get(name, metav1.GetOptions{})
}

func GetSecret(namespace, name string) (*v1.Secret, error) {
	return k8sClient.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
}

func CreatePod(namespace, name string) (*v1.Pod, error) {
	pod := &v1.Pod{}
	pod.Name = name
	pod.Spec.Containers = append(pod.Spec.Containers, v1.Container{
		Name:  "test-container",
		Image: "busyBox",
	})
	return k8sClient.CoreV1().Pods(namespace).Create(pod)
}

func CreatePvc(namespace, name string, labels map[string]string) (*v1.PersistentVolumeClaim, error) {
	pvc := &v1.PersistentVolumeClaim{}
	pvc.Name = name
	pvc.Labels = labels
	pvc.Spec.AccessModes = []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce}
	storageClass := "test-storageclass"
	pvc.Spec.StorageClassName = &storageClass
	pvc.Spec.Resources = v1.ResourceRequirements{
		Requests: v1.ResourceList{v1.ResourceStorage: resource.MustParse("10Gi")},
	}
	return k8sClient.CoreV1().PersistentVolumeClaims(namespace).Create(pvc)
}

func GetPvc(namespace, name string) (*v1.PersistentVolumeClaim, error) {
	return k8sClient.CoreV1().PersistentVolumeClaims(namespace).Get(name, metav1.GetOptions{})
}

func DeleteStatefulSet(namespace, name string) (error) {
	return k8sClient.AppsV1beta1().StatefulSets(namespace).Delete(name, &metav1.DeleteOptions{})
}

func CreateReleaseConfig(namespace, name string, labels map[string]string) (*v1beta1.ReleaseConfig, error) {
	releaseConfig := &v1beta1.ReleaseConfig{}
	releaseConfig.Name = name
	releaseConfig.Labels = labels
	return k8sReleaseConfigClient.TranswarpV1beta1().ReleaseConfigs(namespace).Create(releaseConfig)
}

func UpdateReleaseConfig(releaseConfig *v1beta1.ReleaseConfig) (*v1beta1.ReleaseConfig, error) {
	return k8sReleaseConfigClient.TranswarpV1beta1().ReleaseConfigs(releaseConfig.Namespace).Update(releaseConfig)
}

func DeleteReleaseConfig(namespace, name string) (error) {
	return k8sReleaseConfigClient.TranswarpV1beta1().ReleaseConfigs(namespace).Delete(name, &metav1.DeleteOptions{})
}

func CreateConfigMap(namespace, name string) (*v1.ConfigMap, error) {
	configMap := &v1.ConfigMap{}
	configMap.Name = name
	return k8sClient.CoreV1().ConfigMaps(namespace).Create(configMap)
}

func CreateDaemonSet(namespace, name string) (*extv1beta1.DaemonSet, error) {
	resource := &extv1beta1.DaemonSet{}
	resource.Name = name
	resource.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{"app": "fluentd"},
	}
	resource.Spec.Template = v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{"app": "fluentd"},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "fluentd",
					Image: "test-fluentd",
				},
			},
		},
	}
	return k8sClient.ExtensionsV1beta1().DaemonSets(namespace).Create(resource)
}

func CreateDeployment(namespace, name string) (*extv1beta1.Deployment, error) {
	resource := &extv1beta1.Deployment{}
	resource.Name = name
	resource.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{"app": "deploy-nginx"},
	}
	resource.Spec.Template = v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{"app": "deploy-nginx"},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "nginx",
					Image: "test-nginx",
				},
			},
		},
	}
	return k8sClient.ExtensionsV1beta1().Deployments(namespace).Create(resource)
}

func CreateService(namespace, name string, selector map[string]string) (*v1.Service, error) {
	resource := &v1.Service{}
	resource.Name = name
	resource.Spec.Selector = selector
	resource.Spec.Ports = []v1.ServicePort{{
		Port: 80,
		TargetPort: intstr.FromInt(80),
	}}
	return k8sClient.CoreV1().Services(namespace).Create(resource)
}

func CreateStatefulSet(namespace, name string) (*appsv1.StatefulSet, error) {
	statefulSet := &appsv1.StatefulSet{}
	statefulSet.Name = name
	statefulSet.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{"app": "nginx"},
	}
	statefulSet.Spec.Template = v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{"app": "nginx"},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "nginx",
					Image: "test-nginx",
				},
			},
		},
	}
	testStorageClass := "test-storage-class"
	statefulSet.Spec.VolumeClaimTemplates = []v1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "www-test",
			},
			Spec: v1.PersistentVolumeClaimSpec{
				StorageClassName: &testStorageClass,
				AccessModes:      []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{v1.ResourceStorage: resource.MustParse("10Gi")},
				},
			},
		},
	}
	return k8sClient.AppsV1beta1().StatefulSets(namespace).Create(statefulSet)
}

func CreateJob(namespace, name string) (*batchv1.Job, error) {
	resource := &batchv1.Job{}
	resource.Name = name
	resource.Spec.Template = v1.PodTemplateSpec{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "pi",
					Image: "perl",
					Command: []string{"perl",  "-Mbignum=bpi", "-wle", "print bpi(2000)"},
				},
			},
			RestartPolicy: v1.RestartPolicyNever,
		},
	}

	return k8sClient.BatchV1().Jobs(namespace).Create(resource)
}

func CreateIngress(namespace, name string) (*extv1beta1.Ingress, error) {
	resource := &extv1beta1.Ingress{}
	resource.Name = name
	resource.Spec.Rules = []extv1beta1.IngressRule{{
		Host: "test.cn",
	}}
	return k8sClient.ExtensionsV1beta1().Ingresses(namespace).Create(resource)
}

func CreateSecret(namespace, name string) (*v1.Secret, error) {
	resource := &v1.Secret{}
	resource.Name = name
	return k8sClient.CoreV1().Secrets(namespace).Create(resource)
}

func CreateStorageClass(namespace, name string) (*storagev1.StorageClass, error) {
	resource := &storagev1.StorageClass{}
	resource.Name = name
	resource.Provisioner = "test-provisioner"
	return k8sClient.StorageV1().StorageClasses().Create(resource)
}

func DeleteStorageClass(namespace, name string) (error) {
	return k8sClient.StorageV1().StorageClasses().Delete(name, &metav1.DeleteOptions{})
}

func LoadChartArchive(name string) ([]*loader.BufferedFile, error) {
	if fi, err := os.Stat(name); err != nil {
		return nil, err
	} else if fi.IsDir() {
		return nil, errors.New("cannot load a directory")
	}

	raw, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer raw.Close()
	return transwarpjsonnet.LoadArchive(raw)
}

//func GetCurrentFilePath() (string, error) {
//	_, file, _, ok := runtime.Caller(1)
//	if !ok {
//		return "", errors.New("Can not get current file info")
//	}
//	return file, nil
//}

func InitFramework() error {
	kubeConfig := ""
	if setting.Config.KubeConfig != nil {
		kubeConfig = setting.Config.KubeConfig.Config
	}

	var err error
	k8sClient, err = client.NewClient("", kubeConfig)
	if err != nil {
		logrus.Errorf("failed to create k8s client : %s", err.Error())
		return err
	}

	k8sReleaseConfigClient, err = client.NewReleaseConfigClient("", kubeConfig)
	if err != nil {
		logrus.Errorf("failed to create k8s release config client : %s", err.Error())
		return err
	}

	kubeClients = clienthelm.NewHelmKubeClient(kubeConfig)
	return nil
}
