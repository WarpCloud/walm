package helm

import (
	. "github.com/onsi/gomega"
	. "github.com/onsi/ginkgo"
	"WarpCloud/walm/test/e2e/framework"
	"WarpCloud/walm/pkg/k8s/cache/informer"
	"WarpCloud/walm/pkg/helm/impl"
	"WarpCloud/walm/pkg/setting"
	"WarpCloud/walm/pkg/models/release"
	"path/filepath"
)

var _ = Describe("HelmRelease", func() {

	var (
		namespace string
		helm  *impl.Helm
		err       error
		stopChan     chan struct{}
	)

	BeforeEach(func() {
		By("create namespace")
		namespace, err = framework.CreateRandomNamespace("helmReleaseTest", nil)
		Expect(err).NotTo(HaveOccurred())
		stopChan = make(chan struct{})
		k8sCache := informer.NewInformer(framework.GetK8sClient(), framework.GetK8sReleaseConfigClient(), 0, stopChan)
		registryClient := impl.NewRegistryClient(setting.Config.ChartImageConfig)

		helm, err = impl.NewHelm(setting.Config.RepoList, registryClient, k8sCache, framework.GetKubeClient())
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		By("delete namespace")
		close(stopChan)
		err = framework.DeleteNamespace(namespace)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("test install release", func() {
		By("install release with local chart")
		releaseRequest := &release.ReleaseRequestV2{
			ReleaseRequest: release.ReleaseRequest{
				Name: "tomcat-test",
			},
		}
		currentFilePath, err := framework.GetCurrentFilePath()
		Expect(err).NotTo(HaveOccurred())
		chartPath := filepath.Join(filepath.Dir(currentFilePath), "../../resources/helm/tomcat-0.2.0.tgz")
		chartFiles, err := framework.LoadChartArchive(chartPath)
		Expect(err).NotTo(HaveOccurred())

		releaseCache, err := helm.InstallOrCreateRelease(namespace, releaseRequest, chartFiles, false, false, nil, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(releaseCache).To(Equal(&release.ReleaseCache{
			ReleaseSpec: release.ReleaseSpec{
				Name: "tomcat-test",
				Dependencies: map[string]string{},
				ConfigValues: map[string]interface{}{},
				Version: 0,
				ChartName: "tomcat",
				ChartVersion: "",
				ChartAppVersion: "",
				Namespace: namespace,
			},
			Manifest: "",
			ComputedValues: map[string]interface{}{},

		}))
	})


})

