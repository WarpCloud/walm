package node

import (
	"WarpCloud/walm/pkg/k8s/adaptor"
	. "github.com/onsi/gomega"
	. "github.com/onsi/ginkgo"
	"WarpCloud/walm/pkg/k8s/handler"
	"WarpCloud/walm/router/api"
)

var _ = Describe("Node", func() {

	var (
		nodeName string
	)

	BeforeEach(func() {

		// check node
		nodes, err := handler.GetDefaultHandlerSet().GetNodeHandler().ListNodes(nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(nodes).NotTo(HaveLen(0))

		nodeName = nodes[0].Name

		_, err = adaptor.GetDefaultAdaptorSet().GetAdaptor("Node").(*adaptor.WalmNodeAdaptor).GetResource("", nodeName)
		Expect(err).NotTo(HaveOccurred())

	})

	AfterEach(func() {
		// ....
	})

	Describe("update node annotations && labels", func() {

		It("update annotations", func() {

			annotateNodeRequest := &api.AnnotateNodeRequestBody{}
			addAnnotations := map[string]string{}

			By("add annotations")

			addAnnotations["timestamp"] = "2019-01-01"
			annotateNodeRequest.AddAnnotations = addAnnotations

			nodeInfo, err := handler.GetDefaultHandlerSet().GetNodeHandler().AnnotateNode(nodeName, annotateNodeRequest.AddAnnotations, annotateNodeRequest.RemoveAnnotations)
			Expect(err).NotTo(HaveOccurred())
			Expect(nodeInfo.Annotations["timestamp"]).To(Equal("2019-01-01"))

			By("remove annotations")

			var removeAnnotations []string
			removeAnnotations = append(removeAnnotations, "timestamp")
			annotateNodeRequest.RemoveAnnotations = removeAnnotations
			annotateNodeRequest.AddAnnotations = map[string]string{}

			newNodeInfo, err := handler.GetDefaultHandlerSet().GetNodeHandler().AnnotateNode(nodeName, annotateNodeRequest.AddAnnotations, annotateNodeRequest.RemoveAnnotations)
			Expect(err).NotTo(HaveOccurred())
			if _, ok := newNodeInfo.Annotations["timestamp"]; ok {
				Fail("remove annotations failed")
			}
		})

		It("update labels", func() {

			labelNodeRequest := &api.LabelNodeRequestBody{}
			addLabels := map[string]string{}

			By("add labels")

			addLabels["environment"] = "tos"
			labelNodeRequest.AddLabels = addLabels

			nodeInfo, err := handler.GetDefaultHandlerSet().GetNodeHandler().LabelNode(nodeName, labelNodeRequest.AddLabels, labelNodeRequest.RemoveLabels)
			Expect(err).NotTo(HaveOccurred())
			Expect(nodeInfo.Labels["environment"]).To(Equal("tos"))

			By("remove labels")

			var removeLabels []string
			removeLabels = append(removeLabels, "environment")
			labelNodeRequest.RemoveLabels = removeLabels
			labelNodeRequest.AddLabels = map[string]string{}

			newNodeInfo, err := handler.GetDefaultHandlerSet().GetNodeHandler().LabelNode(nodeName, labelNodeRequest.AddLabels, labelNodeRequest.RemoveLabels)
			Expect(err).NotTo(HaveOccurred())
			if _, ok := newNodeInfo.Labels["environment"]; ok {
				Fail("remove labels failed")
			}

		})

	})
})
