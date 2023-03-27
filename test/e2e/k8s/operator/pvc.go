package operator

import (
	. "github.com/onsi/gomega"
	. "github.com/onsi/ginkgo"
	"WarpCloud/walm/test/e2e/framework"
	"WarpCloud/walm/pkg/k8s/operator"
	"WarpCloud/walm/pkg/models/k8s"
	"strings"
	"WarpCloud/walm/pkg/k8s/converter"
	"WarpCloud/walm/pkg/k8s/cache/informer"
	"time"
)

var _ = Describe("K8sOperatorPvc", func() {

	var (
		namespace   string
		k8sOperator *operator.Operator
		err         error
		stopChan    chan struct{}
	)

	BeforeEach(func() {
		By("create namespace")
		namespace, err = framework.CreateRandomNamespace("k8sOperatorPvcTest", nil)
		Expect(err).NotTo(HaveOccurred())
		stopChan = make(chan struct{})
		k8sCache := informer.NewInformer(framework.GetK8sClient(), framework.GetK8sReleaseConfigClient(), framework.GetK8sInstanceClient(), nil, nil, nil, 0, stopChan)
		k8sOperator = operator.NewOperator(framework.GetK8sClient(), k8sCache, nil, nil)
	})

	AfterEach(func() {
		By("delete namespace")
		close(stopChan)
		err = framework.DeleteNamespace(namespace)
		Expect(err).NotTo(HaveOccurred())
	})

	It("test pvc: delete pvc & delete pvcs & delete statefulSet pvcs", func() {

		By("delete pvc")
		pvc, err := framework.CreatePvc(namespace, "test-pvc", nil)
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(time.Millisecond * 500)

		err = k8sOperator.DeletePvc(namespace, pvc.Name)
		Expect(err).NotTo(HaveOccurred())

		err = k8sOperator.DeletePvc(namespace, "not-existed")
		Expect(err).NotTo(HaveOccurred())

		By("delete pvcs")
		pvc1, err := framework.CreatePvc(namespace, "test-pvc1", map[string]string{"test": "true"})
		Expect(err).NotTo(HaveOccurred())
		pvc2, err := framework.CreatePvc(namespace, "test-pvc2", map[string]string{"test": "false"})
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(time.Millisecond * 500)

		err = k8sOperator.DeletePvcs(namespace, "test=true")
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(time.Millisecond * 500)

		_, err = framework.GetPvc(namespace, pvc1.Name)
		Expect(err).To(HaveOccurred())
		Expect(strings.ToLower(err.Error())).To(ContainSubstring("not found"))

		_, err = framework.GetPvc(namespace, pvc2.Name)
		Expect(err).NotTo(HaveOccurred())

		By("delete statefulSet pvcs")
		statefulSet, err := framework.CreateStatefulSet(namespace, "test-sts", nil)
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(time.Second * 1)

		pvcName := statefulSet.Spec.VolumeClaimTemplates[0].Name + "-" + statefulSet.Name + "-0"
		_, err = framework.GetPvc(namespace, pvcName)
		Expect(err).NotTo(HaveOccurred())

		walmStatefulSet, err := converter.ConvertStatefulSetFromK8s(statefulSet, nil)
		Expect(err).NotTo(HaveOccurred())

		err = k8sOperator.DeletePvcs(walmStatefulSet.Namespace, walmStatefulSet.Selector)
		Expect(err).To(HaveOccurred())

		err = framework.DeleteStatefulSet(namespace, "test-sts")
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(time.Millisecond * 500)

		_, err = framework.GetPvc(namespace, pvcName)
		Expect(err).NotTo(HaveOccurred())

		err = k8sOperator.DeleteStatefulSetPvcs([]*k8s.StatefulSet{walmStatefulSet})
		Expect(err).NotTo(HaveOccurred())
	})
})
