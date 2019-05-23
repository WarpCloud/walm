package e2e_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"
	"go/build"
	"WarpCloud/walm/pkg/setting"
	"WarpCloud/walm/pkg/k8s/informer"
	"WarpCloud/walm/pkg/task"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	transwarpscheme "transwarp/release-config/pkg/client/clientset/versioned/scheme"
	// tests to run
	//_ "WarpCloud/walm/test/e2e/pvc"
	//_ "WarpCloud/walm/test/e2e/release"
	_ "WarpCloud/walm/test/e2e/node"
	_ "WarpCloud/walm/test/e2e/secret"
	_ "WarpCloud/walm/test/e2e/tenant"
	//_ "WarpCloud/walm/test/e2e/project"
	_ "WarpCloud/walm/test/e2e/k8s/handler"
	_ "WarpCloud/walm/test/e2e/release/manager/helm"
)

var stopChan = make(chan struct{})

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2e Suite")
}

var _ = BeforeSuite(func() {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}

	setting.InitConfig(gopath + "/src/WarpCloud/walm/walm.yaml")
	setting.Config.KubeConfig.Config = gopath + "/src/WarpCloud/walm/test/k8sconfig/kubeconfig"

	transwarpscheme.AddToScheme(clientsetscheme.Scheme)

	informer.StartInformer(stopChan)
	task.GetDefaultTaskManager().StartWorker()
})

var _ = AfterSuite(func() {
	task.GetDefaultTaskManager().StopWorker()
	close(stopChan)
})
