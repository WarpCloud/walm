package e2e_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"
	"go/build"
	"walm/pkg/setting"
	"walm/pkg/k8s/informer"
	"walm/pkg/task"

	// tests to run
	_ "walm/test/e2e/project"
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
	setting.Config.KubeConfig = &setting.KubeConfig{
		Config: gopath + "/src/walm/test/k8sconfig/kubeconfig",
	}

	setting.InitConfig(gopath + "/src/walm/walm.yaml")
	setting.Config.KubeConfig.Config = gopath + "/src/walm/test/k8sconfig/kubeconfig"

	informer.StartInformer(stopChan)
	task.GetDefaultTaskManager().StartWorker()
})

var _ = AfterSuite(func() {
	task.GetDefaultTaskManager().StopWorker()
	close(stopChan)
})