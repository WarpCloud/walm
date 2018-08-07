package transwarp

import (
	"k8s.io/apimachinery/pkg/util/wait"
	discovery "k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path"
	"time"
	clientsetex "transwarp/application-instance/pkg/client/clientset/versioned"
)

const (
	// High enough QPS to fit all expected use cases. QPS=0 is not set here, because
	// client code is overriding it.
	defaultQPS = 1e6
	// High enough Burst to fit all expected use cases. Burst=0 is not set here, because
	// client code is overriding it.
	defaultBurst = 1e6

	defaultK8sConfigFileNameSuffix = ".kube/config"

	defaultK8sConfigPathEnvVar = "KUBECONFIG"
)

var (
	RecommendedK8sConfig = path.Join(homedir.HomeDir(), defaultK8sConfigFileNameSuffix)
)

// k8s transwarp client to deal with instance, only for k8s 1.9+
func GetTranswarpKubeClient(kubeConfig string) (*clientsetex.Clientset, error) {

	if kubeConfig == "" {
		envVarFiles := os.Getenv(defaultK8sConfigPathEnvVar)
		if len(envVarFiles) != 0 {
			kubeConfig = envVarFiles
		} else {
			kubeConfig = RecommendedK8sConfig
		}
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return nil, err
	}

	cfg.QPS = defaultQPS
	cfg.Burst = defaultBurst
	cfg.ContentType = "application/vnd.kubernetes.protobuf"

	client, err := clientsetex.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return client, nil

}

func GetK8sKubeClient(kubeConfig string) (*kubernetes.Clientset, error) {

	if kubeConfig == "" {
		envVarFiles := os.Getenv(defaultK8sConfigPathEnvVar)
		if len(envVarFiles) != 0 {
			kubeConfig = envVarFiles
		} else {
			kubeConfig = RecommendedK8sConfig
		}
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return nil, err
	}

	cfg.QPS = defaultQPS
	cfg.Burst = defaultBurst
	cfg.ContentType = "application/vnd.kubernetes.protobuf"

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

	err = wait.ExponentialBackoff(defaultRetry, func() (bool, error) {
		v, err = client.Discovery().ServerVersion()

		if err == nil {
			return true, nil
		}

		lastErr = err
		retries++
		return false, nil
	})

	// err is not null only if there was a timeout in the exponential backoff (ErrWaitTimeout)
	if err != nil {
		return nil, lastErr
	}

	// this should not happen, warn the user
	//if retries > 0 {
	//	Log.Warnf("it was required to retry %v times before reaching the API server", retries)
	//}

	//Log.Infof("Running in Kubernetes Cluster version v%v.%v (%v) - git (%v) commit %v - platform %v",
	//	v.Major, v.Minor, v.GitVersion, v.GitTreeState, v.GitCommit, v.Platform)

	return client, nil

}
