package pvc

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"WarpCloud/walm/pkg/release"
	"github.com/satori/go.uuid"
	"WarpCloud/walm/pkg/k8s/handler"
	"go/build"
	"io/ioutil"
	"os"
	"encoding/json"
	"WarpCloud/walm/pkg/release/manager/helm"
	"WarpCloud/walm/pkg/k8s/adaptor"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Pvc", func() {

	var (
		namespace      string
		gopath         string
		releaseName    string
		labelSelector  *v1.LabelSelector
		releaseRequest release.ReleaseRequestV2
		err            error
	)

	BeforeEach(func() {
		By("create namespace")
		randomId := uuid.Must(uuid.NewV4()).String()
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
		err = helm.GetDefaultHelmClient().InstallUpgradeRelease(namespace, &releaseRequest, false, nil, false, 0)
		Expect(err).NotTo(HaveOccurred())
		By("get release pvc num")

		labelSelectorStr := ""
		labelSelector, err := metav1.ParseToLabelSelector(labelSelectorStr)

		releasePvcs, err := adaptor.GetDefaultAdaptorSet().GetAdaptor("PersistentVolumeClaim").
		(*adaptor.WalmPersistentVolumeClaimAdaptor).GetWalmPersistentVolumeClaimAdaptors(namespace, labelSelector)

		Expect(err).NotTo(HaveOccurred())
		Expect(len(releasePvcs)).NotTo(Equal(0))

	})

	AfterEach(func() {

		By("delete namespace")
		err = handler.GetDefaultHandlerSet().GetNamespaceHandler().DeleteNamespace(namespace)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("delete pvcs", func() {
		It("delete pvcs success", func() {

			By("delete release")

			err = helm.GetDefaultHelmClient().DeleteRelease(namespace, releaseName, false, true, false, 0)
			Expect(err).NotTo(HaveOccurred())

			By("validate pvcs related release deleted")

			releasePvcs, err := adaptor.GetDefaultAdaptorSet().GetAdaptor("PersistentVolumeClaim").
			(*adaptor.WalmPersistentVolumeClaimAdaptor).GetWalmPersistentVolumeClaimAdaptors(namespace, labelSelector)

			Expect(err).NotTo(HaveOccurred())
			Expect(len(releasePvcs)).To(Equal(0))

		})
	})
})
