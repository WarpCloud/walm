package framework

import (
	"WarpCloud/walm/pkg/helm/impl"
	"WarpCloud/walm/pkg/k8s/client"
	clienthelm "WarpCloud/walm/pkg/k8s/client/helm"
	"WarpCloud/walm/pkg/setting"
	"errors"
	"fmt"
	migrationclientset "github.com/migration/pkg/client/clientset/versioned"
	"github.com/sirupsen/logrus"
	"helm.sh/helm/pkg/chart/loader"
	"helm.sh/helm/pkg/registry"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes"
	"runtime"
	"strings"
	applicationinstanceclientset "transwarp/application-instance/pkg/client/clientset/versioned"
	monitorclientset "transwarp/monitor-crd-informer/pkg/client/versioned"
	releaseconfigclientset "transwarp/release-config/pkg/client/clientset/versioned"
)

var k8sClient *kubernetes.Clientset
var k8sReleaseConfigClient *releaseconfigclientset.Clientset
var kubeClients *clienthelm.Client
var k8sMigrationClient *migrationclientset.Clientset
var k8sInstanceClient *applicationinstanceclientset.Clientset
var k8sMonitoreClient *monitorclientset.Clientset

const (
	maxNameLength                = 62
	randomLength                 = 5
	maxGeneratedRandomNameLength = maxNameLength - randomLength

	// For helm test
	TestChartRepoName  = "test"
	TomcatChartName    = "tomcat"
	TomcatChartVersion = "0.2.0"

	V1ZookeeperChartName    = "zookeeper"
	V1ZookeeperChartVersion = "5.2.0"

	tomcatChartImageSuffix = "walm-test/tomcat:0.2.0"
)

func GenerateRandomName(base string) string {
	if len(base) > maxGeneratedRandomNameLength {
		base = base[:maxGeneratedRandomNameLength]
	}
	return fmt.Sprintf("%s-%s", strings.ToLower(base), utilrand.String(randomLength))
}

func GetCurrentFilePath() (string, error) {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		return "", errors.New("Can not get current file info")
	}
	return file, nil
}

func InitFramework() error {
	tomcatChartPath, err := GetLocalTomcatChartPath()
	if err != nil {
		logrus.Errorf("failed to get tomcat chart path : %s", err.Error())
		return err
	}

	v1ZookeeperChartPath, err := GetLocalV1ZookeeperChartPath()
	if err != nil {
		logrus.Errorf("failed to get v1 zookeeper chart path : %s", err.Error())
		return err
	}

	foundTestRepo := false
	for _, repo := range setting.Config.RepoList {
		if repo.Name == TestChartRepoName {
			foundTestRepo = true
			err = PushChartToRepo(repo.URL, tomcatChartPath)
			if err != nil {
				logrus.Errorf("failed to push tomcat chart to repo : %s", err.Error())
				return err
			}
			err = PushChartToRepo(repo.URL, v1ZookeeperChartPath)
			if err != nil {
				logrus.Errorf("failed to push v1 zookeeper chart to repo : %s", err.Error())
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

	chartImage := GetTomcatChartImage()
	logrus.Infof("start to push chart image %s to registry", chartImage)
	registryClient, err := impl.NewRegistryClient(setting.Config.ChartImageConfig)
	if err != nil {
		return err
	}

	testChart, err := loader.Load(tomcatChartPath)
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
	kubeContext := ""
	if setting.Config.KubeConfig != nil {
		kubeContext = setting.Config.KubeConfig.Context
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

	k8sMigrationClient, err = client.NewMigrationClient("", kubeConfig)
	if err != nil {
		logrus.Errorf("failed to create k8s crd client : %s", err.Error())
	}

	k8sInstanceClient, err = client.NewInstanceClient("", kubeConfig)
	if err != nil {
		logrus.Errorf("failed to create k8s instance client : %s", err.Error())
	}

	kubeClients = clienthelm.NewHelmKubeClient(kubeConfig, kubeContext, nil)

	return nil
}
