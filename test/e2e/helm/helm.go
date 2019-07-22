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
		helm      *impl.Helm
		err       error
		stopChan  chan struct{}
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

	It("test install release", func() {
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
		assertReleaseCacheBasic(releaseCache, namespace, "tomcat-test", "", "tomcat",
			"0.2.0", "7", 1)

	})

})

func assertReleaseCacheBasic(cache *release.ReleaseCache, namespace, name, repo, chartName, chartVersion,
chartAppVersion string, version int32) {

	Expect(cache.Name).To(Equal(name))
	Expect(cache.Namespace).To(Equal(namespace))
	Expect(cache.RepoName).To(Equal(repo))
	Expect(cache.ChartName).To(Equal(chartName))
	Expect(cache.ChartVersion).To(Equal(chartVersion))
	Expect(cache.ChartAppVersion).To(Equal(chartAppVersion))
	Expect(cache.Version).To(Equal(version))
}
