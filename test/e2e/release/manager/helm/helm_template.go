package helm

import (
	. "github.com/onsi/gomega"
	. "github.com/onsi/ginkgo"
	"WarpCloud/walm/test/e2e/framework"
	"WarpCloud/walm/pkg/release/manager/helm"
	"WarpCloud/walm/pkg/release"
	"path/filepath"
	"fmt"
	"k8s.io/helm/pkg/chart/loader"
	"k8s.io/helm/pkg/walm"
	"k8s.io/helm/pkg/walm/plugins"
	"encoding/json"
)

var _ = Describe("ReleaseDryRun", func() {

	var (
		namespace  string
		helmClient *helm.HelmClient
		chartPath  string
		chartFiles []*loader.BufferedFile
		releaseRequest *release.ReleaseRequestV2
		err        error
	)

	BeforeEach(func() {
		By("create namespace")
		namespace, err = framework.CreateRandomNamespace("releaseDryRunTest")
		Expect(err).NotTo(HaveOccurred())
		helmClient = helm.GetDefaultHelmClient()
		Expect(helmClient).NotTo(BeNil())
		currentFilePath, err := framework.GetCurrentFilePath()
		Expect(err).NotTo(HaveOccurred())
		chartPath = filepath.Join(filepath.Dir(currentFilePath), "../../../../resources/release/manager/helm/tomcat-0.2.0.tgz")
		chartFiles, err = framework.LoadChartArchive(chartPath)
		Expect(err).NotTo(HaveOccurred())
		releaseRequest = &release.ReleaseRequestV2{
			ReleaseRequest: release.ReleaseRequest{
				Name: "tomcat-test",
			},
		}
	})

	AfterEach(func() {
		By("delete namespace")
		err = framework.DeleteNamespace(namespace, false)
		Expect(err).NotTo(HaveOccurred())
	})

	It("test release dry run", func() {

		By("dry run creating release")
		manifest, err := helmClient.DryRunRelease(namespace, releaseRequest, false, chartFiles)
		Expect(err).NotTo(HaveOccurred())
		manifestStr, err := convertManifestArrayToStr(manifest)
		Expect(err).NotTo(HaveOccurred())
		expectedManifestStr := fmt.Sprintf("[{\"apiVersion\":\"v1\",\"kind\":\"Service\",\"metadata\":{\"labels\":{\"app\":\"tomcat\",\"chart\":\"tomcat-0.2.0\",\"heritage\":\"Helm\",\"release\":\"tomcat-test\"},\"name\":\"tomcat-test\",\"namespace\":\"%s\"},\"spec\":{\"ports\":[{\"name\":\"http\",\"port\":80,\"protocol\":\"TCP\",\"targetPort\":8080}],\"selector\":{\"app\":\"tomcat\",\"release\":\"tomcat-test\"},\"type\":\"NodePort\"}},{\"apiVersion\":\"apps/v1beta2\",\"kind\":\"Deployment\",\"metadata\":{\"labels\":{\"app\":\"tomcat\",\"chart\":\"tomcat-0.2.0\",\"heritage\":\"Helm\",\"release\":\"tomcat-test\"},\"name\":\"tomcat-test\",\"namespace\":\"%s\"},\"spec\":{\"replicas\":1,\"selector\":{\"matchLabels\":{\"app\":\"tomcat\",\"release\":\"tomcat-test\"}},\"template\":{\"metadata\":{\"labels\":{\"app\":\"tomcat\",\"release\":\"tomcat-test\"}},\"spec\":{\"containers\":[{\"image\":\"tomcat:7.0\",\"imagePullPolicy\":\"Always\",\"livenessProbe\":{\"httpGet\":{\"path\":\"/sample\",\"port\":8080},\"initialDelaySeconds\":60,\"periodSeconds\":30},\"name\":\"tomcat\",\"ports\":[{\"containerPort\":8080,\"hostPort\":8009}],\"readinessProbe\":{\"failureThreshold\":6,\"httpGet\":{\"path\":\"/sample\",\"port\":8080},\"initialDelaySeconds\":60,\"periodSeconds\":30},\"resources\":{\"limits\":{\"cpu\":4,\"memory\":\"4Gi\"},\"requests\":{\"cpu\":2,\"memory\":\"2Gi\"}},\"volumeMounts\":[{\"mountPath\":\"/usr/local/tomcat/webapps\",\"name\":\"app-volume\"}]}],\"initContainers\":[{\"command\":[\"sh\",\"-c\",\"cp /*.war /app\"],\"image\":\"ananwaresystems/webarchive:1.0\",\"imagePullPolicy\":\"Always\",\"name\":\"war\",\"volumeMounts\":[{\"mountPath\":\"/app\",\"name\":\"app-volume\"}]}],\"volumes\":[{\"emptyDir\":{},\"name\":\"app-volume\"}]}}}},{\"apiVersion\":\"apiextensions.transwarp.io/v1beta1\",\"kind\":\"ReleaseConfig\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"auto-gen\":\"true\"},\"name\":\"tomcat-test\",\"namespace\":\"%s\"},\"spec\":{\"chartAppVersion\":\"7\",\"chartImage\":\"\",\"chartName\":\"tomcat\",\"chartVersion\":\"0.2.0\",\"configValues\":{},\"dependencies\":{},\"dependenciesConfigValues\":{},\"outputConfig\":{},\"repo\":\"\"},\"status\":{}}]", namespace, namespace, namespace)
		Expect(manifestStr).To(Equal(expectedManifestStr))

	})

	It("test release dry run: plugins", func() {

		By("dry run creating release with plugins")
		lablePodArgs := &plugins.LabelPodArgs{
			LabelsToAdd: map[string]string{
				"test_key": "test_value",
			},
		}
		labelPodArgsBytes, err := json.Marshal(lablePodArgs)
		Expect(err).NotTo(HaveOccurred())
		releaseRequest.Plugins = append(releaseRequest.Plugins, &walm.WalmPlugin{
			Name: plugins.LabelPodPluginName,
			Args: string(labelPodArgsBytes),
		})
		manifest, err := helmClient.DryRunRelease(namespace, releaseRequest, false, chartFiles)
		Expect(err).NotTo(HaveOccurred())
		manifestStr, err := convertManifestArrayToStr(manifest)
		Expect(err).NotTo(HaveOccurred())
		expectedManifestStr := fmt.Sprintf("[{\"apiVersion\":\"v1\",\"kind\":\"Service\",\"metadata\":{\"labels\":{\"app\":\"tomcat\",\"chart\":\"tomcat-0.2.0\",\"heritage\":\"Helm\",\"release\":\"tomcat-test\"},\"name\":\"tomcat-test\",\"namespace\":\"%s\"},\"spec\":{\"ports\":[{\"name\":\"http\",\"port\":80,\"protocol\":\"TCP\",\"targetPort\":8080}],\"selector\":{\"app\":\"tomcat\",\"release\":\"tomcat-test\"},\"type\":\"NodePort\"}},{\"apiVersion\":\"extensions/v1beta1\",\"kind\":\"Deployment\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"tomcat\",\"chart\":\"tomcat-0.2.0\",\"heritage\":\"Helm\",\"release\":\"tomcat-test\"},\"name\":\"tomcat-test\",\"namespace\":\"%s\"},\"spec\":{\"replicas\":1,\"selector\":{\"matchLabels\":{\"app\":\"tomcat\",\"release\":\"tomcat-test\"}},\"strategy\":{},\"template\":{\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"tomcat\",\"release\":\"tomcat-test\",\"test_key\":\"test_value\"}},\"spec\":{\"containers\":[{\"image\":\"tomcat:7.0\",\"imagePullPolicy\":\"Always\",\"livenessProbe\":{\"httpGet\":{\"path\":\"/sample\",\"port\":8080},\"initialDelaySeconds\":60,\"periodSeconds\":30},\"name\":\"tomcat\",\"ports\":[{\"containerPort\":8080,\"hostPort\":8009}],\"readinessProbe\":{\"failureThreshold\":6,\"httpGet\":{\"path\":\"/sample\",\"port\":8080},\"initialDelaySeconds\":60,\"periodSeconds\":30},\"resources\":{\"limits\":{\"cpu\":\"4\",\"memory\":\"4Gi\"},\"requests\":{\"cpu\":\"2\",\"memory\":\"2Gi\"}},\"volumeMounts\":[{\"mountPath\":\"/usr/local/tomcat/webapps\",\"name\":\"app-volume\"}]}],\"initContainers\":[{\"command\":[\"sh\",\"-c\",\"cp /*.war /app\"],\"image\":\"ananwaresystems/webarchive:1.0\",\"imagePullPolicy\":\"Always\",\"name\":\"war\",\"resources\":{},\"volumeMounts\":[{\"mountPath\":\"/app\",\"name\":\"app-volume\"}]}],\"volumes\":[{\"emptyDir\":{},\"name\":\"app-volume\"}]}}},\"status\":{}},{\"apiVersion\":\"apiextensions.transwarp.io/v1beta1\",\"kind\":\"ReleaseConfig\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"auto-gen\":\"true\"},\"name\":\"tomcat-test\",\"namespace\":\"%s\"},\"spec\":{\"chartAppVersion\":\"7\",\"chartImage\":\"\",\"chartName\":\"tomcat\",\"chartVersion\":\"0.2.0\",\"configValues\":{},\"dependencies\":{},\"dependenciesConfigValues\":{},\"outputConfig\":{},\"repo\":\"\"},\"status\":{}}]", namespace, namespace, namespace)
		Expect(manifestStr).To(Equal(expectedManifestStr))

	})
})

func convertManifestArrayToStr(manifests []map[string]interface{}) (string, error) {
	bytes, err := json.Marshal(manifests)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

