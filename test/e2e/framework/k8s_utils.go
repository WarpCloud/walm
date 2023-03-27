package framework

import(
	"errors"
	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clienthelm "WarpCloud/walm/pkg/k8s/client/helm"
	transwarpv1beta1 "transwarp/application-instance/pkg/apis/transwarp/v1beta1"
	tosv1beta1 "github.com/migration/pkg/apis/tos/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog"
	"path/filepath"
	applicationinstanceclientset "transwarp/application-instance/pkg/client/clientset/versioned"
	monitorclientset "transwarp/monitor-crd-informer/pkg/client/versioned"
	releaseconfigclientset "transwarp/release-config/pkg/client/clientset/versioned"
	"transwarp/release-config/pkg/apis/transwarp/v1beta1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"WarpCloud/walm/pkg/k8s/operator"
	"time"
	migrationclientset "github.com/migration/pkg/client/clientset/versioned"

)

func GetKubeClient() *clienthelm.Client {
	return kubeClients
}

func GetK8sInstanceClient() *applicationinstanceclientset.Clientset {
	return k8sInstanceClient
}

func GetK8sMonitorClient() *monitorclientset.Clientset {
	return k8sMonitoreClient
}

func GetK8sMigrationClient() *migrationclientset.Clientset {
	return k8sMigrationClient
}

func GetK8sClient() *kubernetes.Clientset {
	return k8sClient
}

func GetK8sReleaseConfigClient() *releaseconfigclientset.Clientset {
	return k8sReleaseConfigClient
}

func CreateRandomNamespace(base string, labels map[string]string) (string, error) {
	namespace := GenerateRandomName(base)
	ns := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   namespace,
			Labels: labels,
		},
	}
	_, err := k8sClient.CoreV1().Namespaces().Create(&ns)
	return namespace, err
}

func CreateNamespace(name string, labels map[string]string) (error) {
	ns := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}
	_, err := k8sClient.CoreV1().Namespaces().Create(&ns)
	return err
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

func CreateResourceQuota(namespace, name string) (*v1.ResourceQuota, error) {
	resourceQuota := &v1.ResourceQuota{}
	resourceQuota.Name = name
	resourceQuota.Spec.Hard = v1.ResourceList{
		v1.ResourceMemory: resource.MustParse("1Gi"),
	}
	return k8sClient.CoreV1().ResourceQuotas(namespace).Create(resourceQuota)
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

func LabelNode(name string, labelsToAdd map[string]string, labelsToRemove []string) error {
	k8sOperator := operator.NewOperator(k8sClient, nil, nil, nil)
	return k8sOperator.LabelNode(name, labelsToAdd, labelsToRemove)
}

func ListNodes() (*v1.NodeList, error) {
	return k8sClient.CoreV1().Nodes().List(metav1.ListOptions{})
}

func GetSecret(namespace, name string) (*v1.Secret, error) {
	return k8sClient.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
}

func CreatePod(namespace, name string) (*v1.Pod, error) {
	pod := &v1.Pod{}
	pod.Name = name
	pod.Spec.Containers = append(pod.Spec.Containers, v1.Container{
		Name:            "test-container",
		Image:           "nginx",
		ImagePullPolicy: v1.PullIfNotPresent,
	})
	return k8sClient.CoreV1().Pods(namespace).Create(pod)
}

func WaitPodRunning(namespace, name string) (error) {
	waitTimes := 10
	waitInterval := time.Second * 20
	for {
		pod, err := k8sClient.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if pod.Status.Phase == "Running" {
			return nil
		}
		if waitTimes <= 0 {
			return errors.New("timeout waiting pod running")
		}
		waitTimes --
		logrus.Infof("waiting for pod %s/%s running...", namespace, name)
		time.Sleep(waitInterval)
	}
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

func DeleteConfigMap(namespace, name string) (error) {
	return k8sClient.CoreV1().ConfigMaps(namespace).Delete(name, &metav1.DeleteOptions{})
}

func CreateCustomConfigMap(namespace, configPath string) (*v1.ConfigMap, error) {
	path, err := filepath.Abs(configPath)
	if err != nil {
		klog.Errorf("get configmap path err: %s", err.Error())
		return nil, err
	}
	configMapByte, err := ioutil.ReadFile(path)
	if err != nil {
		klog.Errorf("read configmap err: %s", err.Error())
	}

	configMap := &v1.ConfigMap{}
	err = yaml.Unmarshal(configMapByte, &configMap)
	if err != nil {
		klog.Errorf("unmarshal to configMap err: %s", err.Error())
	}
	return k8sClient.CoreV1().ConfigMaps(namespace).Create(configMap)
}

func CreateDaemonSet(namespace, name string) (*extv1beta1.DaemonSet, error) {
	resource := &extv1beta1.DaemonSet{}
	resource.Name = name
	resource.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{"app": "ds-" + name},
	}
	resource.Spec.Template = v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{"app": "ds-" + name},
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
		MatchLabels: map[string]string{"app": "dp-" + name},
	}
	resource.Spec.Template = v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{"app": "dp-" + name},
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
		Port:       80,
		TargetPort: intstr.FromInt(80),
	}}
	return k8sClient.CoreV1().Services(namespace).Create(resource)
}

func CreateStatefulSet(namespace, name string, labels map[string]string) (*appsv1beta1.StatefulSet, error) {
	statefulSet := &appsv1beta1.StatefulSet{}
	statefulSet.Name = name
	statefulSet.Labels = labels
	statefulSet.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{"app": "sts-" + name},
	}
	statefulSet.Spec.Template = v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{"app": "sts-" + name},
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
					Name:    "pi",
					Image:   "perl",
					Command: []string{"perl", "-Mbignum=bpi", "-wle", "print bpi(2000)"},
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

func CreateSecret(namespace, name string, labels map[string]string) (*v1.Secret, error) {
	resource := &v1.Secret{}
	resource.Name = name
	resource.Labels = labels
	return k8sClient.CoreV1().Secrets(namespace).Create(resource)
}

func CreateReplicaSet(namespace, name string, labels map[string]string) (*appsv1.ReplicaSet, error) {
	replicas := int32(1)
	resource := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{"app": "guestbook", "tier": "frontend"},
		},
		Spec:       appsv1.ReplicaSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"tier": "frontend",
				},
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key: "tier",
						Operator: metav1.LabelSelectorOpIn,
						Values: []string{"frontend"},
					},
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "guestbook",
						"tier": "frontend",
					},
				},
				Spec:       v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "redis",
								Image: "docker.io/library/redis",
							},
						},
				},
			},
		},
		Status:     appsv1.ReplicaSetStatus{},
	}
	resource.Name = name
	resource.Labels = labels
	return k8sClient.AppsV1().ReplicaSets(namespace).Create(resource)
}

func CreateInstance(namespace, name string, labels map[string]string) (*transwarpv1beta1.ApplicationInstance, error) {
	resource := &transwarpv1beta1.ApplicationInstance{}
	resource.Name = name
	resource.Labels = labels
	return k8sInstanceClient.TranswarpV1beta1().ApplicationInstances(namespace).Create(resource)
}

func CreateMigration(namespace, name string, labels map[string]string) (*tosv1beta1.Mig, error) {
	resource := &tosv1beta1.Mig{}
	resource.Name = name
	resource.Labels = labels
	return k8sMigrationClient.ApiextensionsV1beta1().Migs(namespace).Create(resource)
}

func CreateStorageClass(namespace, name string, labels map[string]string) (*storagev1.StorageClass, error) {
	resource := &storagev1.StorageClass{}
	resource.Name = name
	resource.Labels = labels
	resource.Provisioner = "test-provisioner"
	return k8sClient.StorageV1().StorageClasses().Create(resource)
}

func DeleteStorageClass(namespace, name string) (error) {
	return k8sClient.StorageV1().StorageClasses().Delete(name, &metav1.DeleteOptions{})
}