package test

import (
	"testing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"WarpCloud/walm/pkg/setting"
	"flag"

	_ "WarpCloud/walm/test/e2e/k8s/operator"
	_ "WarpCloud/walm/test/e2e/k8s/cache"
	_ "WarpCloud/walm/test/e2e/helm"
	_ "WarpCloud/walm/test/e2e/redis"
	_ "WarpCloud/walm/test/e2e/task"
	_ "WarpCloud/walm/test/e2e/sync"
	"WarpCloud/walm/test/e2e/framework"
)

var stopChan = make(chan struct{})
var configPath string

func init() {
	flag.StringVar(&configPath, "configPath", "e2e_walm.yaml", "configPath is used to init config")
}

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2e Suite")
}

var _ = BeforeSuite(func() {
	setting.InitConfig(configPath)

	err := framework.InitFramework()
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
})
