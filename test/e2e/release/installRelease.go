package release

import (
	"encoding/json"
	"go/build"
	"io/ioutil"
	"os"
	"time"
	"walm/pkg/k8s/handler"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"walm/pkg/release"
	"walm/pkg/release/manager/helm"
)

var _ = Describe("Release", func() {

	var (
		namespace      string
		gopath         string
		releaseName    string
		releaseRequest release.ReleaseRequestV2
		releaseInfo    *release.ReleaseInfoV2
		err            error
	)

	BeforeEach(func() {

		By("create namespace")
		randomId := uuid.Must(uuid.NewV4(), err).String()
		namespace = "test-" + randomId[:8]

		ns := corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      namespace,
			},
		}
		_, err = handler.GetDefaultHandlerSet().GetNamespaceHandler().CreateNamespace(&ns)
		Expect(err).NotTo(HaveOccurred())

		By("create a release")

		gopath = os.Getenv("GOPATH")
		if gopath == "" {
			gopath = build.Default.GOPATH
		}

		releaseChartByte, err := ioutil.ReadFile(gopath + "/src/walm/test/resources/releases/smartbi.yaml")
		Expect(err).NotTo(HaveOccurred())

		err = json.Unmarshal(releaseChartByte, &releaseRequest)
		Expect(err).NotTo(HaveOccurred())

		releaseRequest.Name = releaseRequest.Name + "-" + randomId[:8]
		releaseName = releaseRequest.Name

	})

	AfterEach(func() {

		By("delete release")
		err := helm.GetDefaultHelmClient().DeleteRelease(namespace, releaseName, false, true, false, 0)
		Expect(err).NotTo(HaveOccurred())

		By("delete namespace")
		err = handler.GetDefaultHandlerSet().GetNamespaceHandler().DeleteNamespace(namespace)
		Expect(err).NotTo(HaveOccurred())

	})

	Describe("install release", func() {
		It("install release success", func() {

			By("start create a release")
			err = helm.GetDefaultHelmClient().InstallUpgradeRelease(namespace, &releaseRequest, false, nil, false, 0)
			Expect(err).NotTo(HaveOccurred())

			releaseInfo, err = helm.GetDefaultHelmClient().GetRelease(namespace, releaseName)
			Expect(releaseInfo.Name).To(Equal(releaseName))

			By("check release status")

			finish := make(chan bool)
			timeout := time.After(time.Second * 720)

			go func() {
				for {
					select {
					case <-timeout:
						Fail("install release timeout, check out please")
					default:
						releaseInfo, err = helm.GetDefaultHelmClient().GetRelease(namespace, releaseName)
						Expect(err).NotTo(HaveOccurred())
						logrus.Infof("install release status: ongoing")
						if releaseInfo.Ready {
							logrus.Infof("install release ready")
							finish <- true
							return
						}
					}
					time.Sleep(time.Second * 20)
				}
			}()

			<-finish

		})
	})
})
