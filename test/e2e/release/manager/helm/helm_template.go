package helm

import (
	. "github.com/onsi/gomega"
	. "github.com/onsi/ginkgo"
	"walm/test/e2e/framework"
	"walm/pkg/release/manager/helm"
	"walm/pkg/release"
	"path/filepath"
	"fmt"
)

var _ = Describe("ReleaseDryRun", func() {

	var (
		namespace  string
		helmClient *helm.HelmClient
		chartPath  string
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
	})

	AfterEach(func() {
		By("delete namespace")
		err = framework.DeleteNamespace(namespace, false)
		Expect(err).NotTo(HaveOccurred())
	})

	It("test release dry run", func() {

		By("dry run creating release")
		chartFiles, err := framework.LoadChartArchive(chartPath)
		Expect(err).NotTo(HaveOccurred())

		releaseRequest := &release.ReleaseRequestV2{
			ReleaseRequest: release.ReleaseRequest{
				Name: "tomcat-test",
			},
		}
		manifestStr, err := helmClient.DryRunRelease(namespace, releaseRequest, false, chartFiles)
		Expect(err).NotTo(HaveOccurred())
		expectedManifestStr := fmt.Sprintf("\n---\n# Source: tomcat/templates/appsrv-svc.yaml\napiVersion: v1\nkind: Service\nmetadata:\n  name: tomcat-test\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    release: tomcat-test\n    heritage: Helm\nspec:\n  type: NodePort\n  ports:\n    - port: 80\n      targetPort: 8080\n      protocol: TCP\n      name: http\n  selector:\n    app: tomcat\n    release: tomcat-test\n---\n# Source: tomcat/templates/appsrv.yaml\napiVersion: apps/v1beta2\nkind: Deployment\nmetadata:\n  name: tomcat-test\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    release: tomcat-test\n    heritage: Helm\nspec:\n  replicas: 1\n  selector:\n    matchLabels:\n      app: tomcat\n      release: tomcat-test\n  template:\n    metadata:\n      labels:\n        app: tomcat\n        release: tomcat-test\n    spec:\n      volumes:\n        - name: app-volume\n          emptyDir: {}\n      initContainers:\n        - name: war\n          image: ananwaresystems/webarchive:1.0\n          imagePullPolicy: Always\n          command:\n            - \"sh\"\n            - \"-c\"\n            - \"cp /*.war /app\"\n          volumeMounts:\n            - name: app-volume\n              mountPath: /app\n      containers:\n        - name: tomcat\n          image: tomcat:7.0\n          imagePullPolicy: Always\n          volumeMounts:\n            - name: app-volume\n              mountPath: /usr/local/tomcat/webapps\n          ports:\n            - containerPort: 8080\n              hostPort: 8009\n          livenessProbe:\n            httpGet:\n              path: /sample\n              port: 8080\n            initialDelaySeconds: 60\n            periodSeconds: 30\n          readinessProbe:\n            httpGet:\n              path: /sample\n              port: 8080\n            initialDelaySeconds: 60\n            periodSeconds: 30\n            failureThreshold: 6\n          resources:\n            limits:\n              cpu: 4\n              memory: 4Gi\n            requests:\n              cpu: 2\n              memory: 2Gi\n---\n# Source: tomcat/autogen-releaseconfig.json.transwarp-jsonnet.yaml\napiVersion: apiextensions.transwarp.io/v1beta1\nkind: ReleaseConfig\nmetadata:\n  creationTimestamp: null\n  labels:\n    auto-gen: \"true\"\n  name: tomcat-test\n  namespace: %s\nspec:\n  chartAppVersion: \"7\"\n  chartName: tomcat\n  chartVersion: 0.2.0\n  configValues: {}\n  dependencies: {}\n  dependenciesConfigValues: {}\n  outputConfig: {}\n  repo: \"\"\nstatus: {}", namespace)
		Expect(manifestStr).To(Equal(expectedManifestStr))

	})
})


