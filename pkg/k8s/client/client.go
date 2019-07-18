package client

import (
	"time"
	. "WarpCloud/walm/pkg/util/log"

	"k8s.io/apimachinery/pkg/util/wait"
	discovery "k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/helm/pkg/kube"
	releaseconfigclientset "transwarp/release-config/pkg/client/clientset/versioned"
	"github.com/sirupsen/logrus"
	restclient "k8s.io/client-go/rest"
	"github.com/hashicorp/golang-lru"
	"WarpCloud/walm/pkg/setting"
)

const (
	// High enough QPS to fit all expected use cases. QPS=0 is not set here, because
	// client code is overriding it.
	defaultQPS = 1e6
	// High enough Burst to fit all expected use cases. Burst=0 is not set here, because
	// client code is overriding it.
	defaultBurst = 1e6
)

var defaultApiserverClient *kubernetes.Clientset
var defaultRestConfig *restclient.Config
var defaultKubeClient *lru.Cache
var defaultReleaseConfigClient *releaseconfigclientset.Clientset

func GetDefaultClient() *kubernetes.Clientset {
	var err error
	if defaultApiserverClient == nil {
		defaultApiserverClient, err = createApiserverClient("", setting.Config.KubeConfig.Config)
	}
	if err != nil {
		logrus.Fatalf("create apiserver client failed:%v", err)
	}
	return defaultApiserverClient
}

func GetDefaultReleaseConfigClient() *releaseconfigclientset.Clientset {
	if defaultReleaseConfigClient == nil {
		var err error
		defaultReleaseConfigClient, err = createReleaseConfigClient("", setting.Config.KubeConfig.Config)
		if err != nil {
			logrus.Fatalf("create release config client failed:%v", err)
		}
	}

	return defaultReleaseConfigClient
}

//func GetDefaultRestConfig() *restclient.Config {
//	var err error
//	if defaultRestConfig == nil {
//		defaultRestConfig, err = clientcmd.BuildConfigFromFlags("", setting.Config.KubeConfig.Config)
//	}
//	if err != nil {
//		logrus.Fatalf("get default rest config= failed:%v", err)
//	}
//	return defaultRestConfig
//}

func GetKubeClient(namespace string) *kube.Client {
	if defaultKubeClient == nil {
		defaultKubeClient, _ = lru.New(100)
	}

	if kubeClient, ok := defaultKubeClient.Get(namespace); ok {
		return kubeClient.(*kube.Client)
	} else {
		kubeClient = createKubeClient(setting.Config.KubeConfig.Config, namespace)
		defaultKubeClient.Add(namespace, kubeClient)
		return kubeClient.(*kube.Client)
	}
}

// createApiserverClient creates new Kubernetes Apiserver client. When kubeconfig or apiserverHost param is empty
// the function assumes that it is running inside a Kubernetes cluster and attempts to
// discover the Apiserver. Otherwise, it connects to the Apiserver specified.
//
// apiserverHost param is in the format of protocol://address:port/pathPrefix, e.g.http://localhost:8001.
// kubeConfig location of kubeconfig file
func createApiserverClient(apiserverHost string, kubeConfig string) (*kubernetes.Clientset, error) {
	cfg, err := clientcmd.BuildConfigFromFlags(apiserverHost, kubeConfig)
	if err != nil {
		return nil, err
	}

	cfg.QPS = defaultQPS
	cfg.Burst = defaultBurst
	cfg.ContentType = "application/vnd.kubernetes.protobuf"

	Log.Infof("Creating API client for %s", cfg.Host)

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	var v *discovery.Info

	// In some environments is possible the client cannot connect the API server in the first request
	// https://github.com/kubernetes/ingress-nginx/issues/1968
	defaultRetry := wait.Backoff{
		Steps:    10,
		Duration: 1 * time.Second,
		Factor:   1.5,
		Jitter:   0.1,
	}

	var lastErr error
	retries := 0
	Log.Info("trying to discover Kubernetes version")
	err = wait.ExponentialBackoff(defaultRetry, func() (bool, error) {
		v, err = client.Discovery().ServerVersion()

		if err == nil {
			return true, nil
		}

		lastErr = err
		Log.Infof("unexpected error discovering Kubernetes version (attempt %v): %v", err, retries)
		retries++
		return false, nil
	})

	// err is not null only if there was a timeout in the exponential backoff (ErrWaitTimeout)
	if err != nil {
		return nil, lastErr
	}

	// this should not happen, warn the user
	if retries > 0 {
		Log.Warnf("it was required to retry %v times before reaching the API server", retries)
	}

	Log.Infof("Running in Kubernetes Cluster version v%v.%v (%v) - git (%v) commit %v - platform %v",
		v.Major, v.Minor, v.GitVersion, v.GitTreeState, v.GitCommit, v.Platform)

	return client, nil
}

// k8s client to deal with release config, only for k8s 1.9+
func createReleaseConfigClient(apiserverHost string, kubeConfig string) (*releaseconfigclientset.Clientset, error) {
	cfg, err := clientcmd.BuildConfigFromFlags(apiserverHost, kubeConfig)
	if err != nil {
		return nil, err
	}

	cfg.QPS = defaultQPS
	cfg.Burst = defaultBurst
	//TODO to investigate protobuf
	//cfg.ContentType = "application/vnd.kubernetes.protobuf"

	Log.Infof("Creating API release config client for %s", cfg.Host)

	client, err := releaseconfigclientset.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// for test
func CreateFakeApiserverClient(apiserverHost string, kubeConfig string) (*kubernetes.Clientset, error) {
	return createApiserverClient(apiserverHost, kubeConfig)
}

// for test
//func CreateFakeKubeClient(apiserverHost string, kubeConfig string) (*kube.Client) {
//	return createKubeClient(apiserverHost, kubeConfig)
//}

func createKubeClient(kubeConfig string, namespace string) (*kube.Client) {
	cfg := kube.GetConfig(kubeConfig, "", namespace)
	client := kube.New(cfg)

	return client
}
