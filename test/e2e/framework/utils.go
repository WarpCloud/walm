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
)

var k8sClient *kubernetes.Clientset
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

func GetLimitRange(namespace, name string) (*v1.LimitRange, error) {
	return k8sClient.CoreV1().LimitRanges(namespace).Get(name, metav1.GetOptions{})
}

func CreatePod(namespace, name string) (*v1.Pod, error) {
	pod := &v1.Pod{}
	pod.Name = name
	pod.Spec.Containers = append(pod.Spec.Containers, v1.Container{
		Name: "test-container",
		Image: "busyBox",
	})
	return k8sClient.CoreV1().Pods(namespace).Create(pod)
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

	kubeClients = clienthelm.NewHelmKubeClient(kubeConfig)
	return nil
}
