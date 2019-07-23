package framework

import (
	"strings"
	"fmt"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	"os"
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
	"WarpCloud/walm/pkg/k8s/operator"
	"time"
	"WarpCloud/walm/pkg/models/common"
	"runtime"
	"github.com/go-resty/resty"
	"path/filepath"
	"WarpCloud/walm/pkg/helm/impl"
	"k8s.io/helm/pkg/chart/loader"
	"k8s.io/helm/pkg/registry"
)

var k8sClient *kubernetes.Clientset
var k8sReleaseConfigClient *releaseconfigclientset.Clientset
var kubeClients *clienthelm.Client

const (
	maxNameLength                = 62
	randomLength                 = 5
	maxGeneratedRandomNameLength = maxNameLength - randomLength

	// For helm test
	TestChartRepoName = "test"
	TestChartName = "tomcat"
	TestChartVersion = "0.2.0"

	testChartImageSuffix = "walm-test/chart:0.2.0"
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
	k8sOperator := operator.NewOperator(k8sClient, nil, nil)
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

func CreateStatefulSet(namespace, name string, labels map[string]string) (*appsv1.StatefulSet, error) {
	statefulSet := &appsv1.StatefulSet{}
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

func LoadChartArchive(name string) ([]*common.BufferedFile, error) {
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
	return common.LoadArchive(raw)
}

func GetCurrentFilePath() (string, error) {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		return "", errors.New("Can not get current file info")
	}
	return file, nil
}

func InitFramework() error {
	testChartPath, err := GetTestTomcatChartPath()
	if err != nil {
		logrus.Errorf("failed to get test chart path : %s", err.Error())
		return err
	}

	foundTestRepo := false
	for _, repo := range setting.Config.RepoList {
		if repo.Name == TestChartRepoName{
			foundTestRepo = true
			err = PushChartToRepo(repo.URL, testChartPath)
			if err != nil {
				logrus.Errorf("failed to push test chart to repo : %s", err.Error())
				return err
			}
			break
		}
	}
	if !foundTestRepo {
		return fmt.Errorf("repo %s is not found", TestChartRepoName)
	}

	if setting.Config.ChartImageRegistry == "" {
		return errors.New("chart image registry should not be empty")
	}

	chartImage := GetTestChartImage()
	logrus.Infof("start to push chart image %s to registry", chartImage)
	registryClient := impl.NewRegistryClient(setting.Config.ChartImageConfig)

	testChart, err := loader.Load(testChartPath)
	if err != nil {
		logrus.Errorf("failed to load test chart : %s", err.Error())
		return err
	}

	ref, err := registry.ParseReference(chartImage)
	if err != nil {
		logrus.Errorf("failed to parse chart image %s : %s", chartImage, err.Error())
		return err
	}

	registryClient.SaveChart(testChart, ref)
	err = registryClient.PushChart(ref)
	if err != nil {
		logrus.Errorf("failed to push chart image : %s", err.Error())
		return err
	}

	kubeConfig := ""
	if setting.Config.KubeConfig != nil {
		kubeConfig = setting.Config.KubeConfig.Config
	}

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

func GetTestTomcatChartPath() (string, error) {
	currentFilePath, err := GetCurrentFilePath()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(currentFilePath), "../../resources/helm/tomcat-0.2.0.tgz"), nil
}

func PushChartToRepo(repoBaseUrl, chartPath string) error{
	logrus.Infof("start to push %s to repo %s", chartPath, repoBaseUrl)
	if !strings.HasSuffix(repoBaseUrl, "/") {
		repoBaseUrl += "/"
	}

	fullUrl := repoBaseUrl + "api/charts"

	resp, err := resty.R().SetHeader("Content-Type", "multipart/form-data" ).
		SetFile("chart", chartPath).Post(fullUrl)

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		logrus.Errorf("status code : %d", resp.StatusCode())
		return errors.New(resp.String())
	}
	return nil
}

func GetTestChartImage() string {
	chartImageRegistry := setting.Config.ChartImageRegistry
	if !strings.HasSuffix(chartImageRegistry, "/") {
		chartImageRegistry += "/"
	}
	return chartImageRegistry + testChartImageSuffix
}