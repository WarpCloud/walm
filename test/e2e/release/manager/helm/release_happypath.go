package helm

import (
	. "github.com/onsi/gomega"
	. "github.com/onsi/ginkgo"
	"WarpCloud/walm/test/e2e/framework"
	"WarpCloud/walm/pkg/release/manager/helm"
	"WarpCloud/walm/pkg/release"
	"path/filepath"
)

var _ = Describe("ReleaseHappyPath", func() {

	var (
		namespace  string
		helmClient *helm.HelmClient
		chartPath  string
		err        error
	)

	BeforeEach(func() {
		By("create namespace")
		namespace, err = framework.CreateRandomNamespace("releaseHappyPathTest")
		Expect(err).NotTo(HaveOccurred())
		helmClient = helm.GetDefaultHelmClient()
		Expect(helmClient).NotTo(BeNil())
		currentFilePath, err := framework.GetCurrentFilePath()
		Expect(err).NotTo(HaveOccurred())
		chartPath = filepath.Join(filepath.Dir(currentFilePath), "../../../../resources/release/manager/helm/tomcat-0.2.0.tgz")
	})

	AfterEach(func() {
		By("delete namespace")
		err = framework.DeleteNamespace(namespace, true)
		Expect(err).NotTo(HaveOccurred())
	})

	It("test release happy path: create, update, pause, recover, restart, delete", func() {

		By("create release")
		chartFiles, err := framework.LoadChartArchive(chartPath)
		Expect(err).NotTo(HaveOccurred())

		releaseRequest := &release.ReleaseRequestV2{
			ReleaseRequest: release.ReleaseRequest{
				Name: "tomcat-test",
			},
		}
		err = helmClient.InstallUpgradeRelease(namespace, releaseRequest, false, chartFiles, false, 0)
		Expect(err).NotTo(HaveOccurred())

		By("update release")
		releaseRequest.ConfigValues = map[string]interface{}{
			"replicaCount": 2,
		}

		err = helmClient.InstallUpgradeRelease(namespace, releaseRequest, false, chartFiles, false, 0)
		Expect(err).NotTo(HaveOccurred())

		release, err := helmClient.GetRelease(namespace, releaseRequest.Name)
		Expect(err).NotTo(HaveOccurred())
		Expect(release.ConfigValues["replicaCount"]).To(Equal(int64(2)))

		By("pause release")
		err = helmClient.PauseRelease(namespace, releaseRequest.Name, false, false, 0)
		Expect(err).NotTo(HaveOccurred())

		release, err = helmClient.GetRelease(namespace, releaseRequest.Name)
		Expect(err).NotTo(HaveOccurred())
		Expect(release.Paused).To(BeTrue())
		Expect(release.Status.Deployments[0].ExpectedReplicas).To(Equal(int32(0)))

		By("recover release")
		err = helmClient.RecoverRelease(namespace, releaseRequest.Name, false, false, 0)
		Expect(err).NotTo(HaveOccurred())

		release, err = helmClient.GetRelease(namespace, releaseRequest.Name)
		Expect(err).NotTo(HaveOccurred())
		Expect(release.Paused).NotTo(BeTrue())
		Expect(release.Status.Deployments[0].ExpectedReplicas).To(Equal(int32(2)))

		By("restart release")
		err = helmClient.RestartRelease(namespace, releaseRequest.Name)
		Expect(err).NotTo(HaveOccurred())

		By("delete release")
		err = helmClient.DeleteRelease(namespace, releaseRequest.Name, false, false, false, 0)
		Expect(err).NotTo(HaveOccurred())
	})
})


