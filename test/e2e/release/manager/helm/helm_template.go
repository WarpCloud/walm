package helm

import (
	. "github.com/onsi/gomega"
	. "github.com/onsi/ginkgo"
	"walm/test/e2e/framework"
	"walm/pkg/release/manager/helm"
	"walm/pkg/release"
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
		manifestStr, err := helmClient.DryRunRelease(namespace, releaseRequest, false, chartFiles)
		Expect(err).NotTo(HaveOccurred())
		expectedManifestStr := fmt.Sprintf("\n---\napiVersion: v1\nkind: Service\nmetadata:\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    heritage: Helm\n    release: tomcat-test\n  name: tomcat-test\n  namespace: %s\nspec:\n  ports:\n  - name: http\n    port: 80\n    protocol: TCP\n    targetPort: 8080\n  selector:\n    app: tomcat\n    release: tomcat-test\n  type: NodePort\n\n---\napiVersion: apps/v1beta2\nkind: Deployment\nmetadata:\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    heritage: Helm\n    release: tomcat-test\n  name: tomcat-test\n  namespace: %s\nspec:\n  replicas: 1\n  selector:\n    matchLabels:\n      app: tomcat\n      release: tomcat-test\n  template:\n    metadata:\n      labels:\n        app: tomcat\n        release: tomcat-test\n    spec:\n      containers:\n      - image: tomcat:7.0\n        imagePullPolicy: Always\n        livenessProbe:\n          httpGet:\n            path: /sample\n            port: 8080\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        name: tomcat\n        ports:\n        - containerPort: 8080\n          hostPort: 8009\n        readinessProbe:\n          failureThreshold: 6\n          httpGet:\n            path: /sample\n            port: 8080\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        resources:\n          limits:\n            cpu: 4\n            memory: 4Gi\n          requests:\n            cpu: 2\n            memory: 2Gi\n        volumeMounts:\n        - mountPath: /usr/local/tomcat/webapps\n          name: app-volume\n      initContainers:\n      - command:\n        - sh\n        - -c\n        - cp /*.war /app\n        image: ananwaresystems/webarchive:1.0\n        imagePullPolicy: Always\n        name: war\n        volumeMounts:\n        - mountPath: /app\n          name: app-volume\n      volumes:\n      - emptyDir: {}\n        name: app-volume\n\n---\napiVersion: apiextensions.transwarp.io/v1beta1\nkind: ReleaseConfig\nmetadata:\n  creationTimestamp: null\n  labels:\n    auto-gen: \"true\"\n  name: tomcat-test\n  namespace: %s\nspec:\n  chartAppVersion: \"7\"\n  chartName: tomcat\n  chartVersion: 0.2.0\n  configValues: {}\n  dependencies: {}\n  dependenciesConfigValues: {}\n  outputConfig: {}\n  repo: \"\"\nstatus: {}\n", namespace, namespace, namespace)
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
		manifestStr, err := helmClient.DryRunRelease(namespace, releaseRequest, false, chartFiles)
		Expect(err).NotTo(HaveOccurred())
		expectedManifestStr := fmt.Sprintf("\n---\napiVersion: v1\nkind: Service\nmetadata:\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    heritage: Helm\n    release: tomcat-test\n  name: tomcat-test\n  namespace: %s\nspec:\n  ports:\n  - name: http\n    port: 80\n    protocol: TCP\n    targetPort: 8080\n  selector:\n    app: tomcat\n    release: tomcat-test\n  type: NodePort\n\n---\napiVersion: extensions/v1beta1\nkind: Deployment\nmetadata:\n  creationTimestamp: null\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    heritage: Helm\n    release: tomcat-test\n  name: tomcat-test\n  namespace: %s\nspec:\n  replicas: 1\n  selector:\n    matchLabels:\n      app: tomcat\n      release: tomcat-test\n  strategy: {}\n  template:\n    metadata:\n      creationTimestamp: null\n      labels:\n        app: tomcat\n        release: tomcat-test\n        test_key: test_value\n    spec:\n      containers:\n      - image: tomcat:7.0\n        imagePullPolicy: Always\n        livenessProbe:\n          httpGet:\n            path: /sample\n            port: 8080\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        name: tomcat\n        ports:\n        - containerPort: 8080\n          hostPort: 8009\n        readinessProbe:\n          failureThreshold: 6\n          httpGet:\n            path: /sample\n            port: 8080\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        resources:\n          limits:\n            cpu: \"4\"\n            memory: 4Gi\n          requests:\n            cpu: \"2\"\n            memory: 2Gi\n        volumeMounts:\n        - mountPath: /usr/local/tomcat/webapps\n          name: app-volume\n      initContainers:\n      - command:\n        - sh\n        - -c\n        - cp /*.war /app\n        image: ananwaresystems/webarchive:1.0\n        imagePullPolicy: Always\n        name: war\n        resources: {}\n        volumeMounts:\n        - mountPath: /app\n          name: app-volume\n      volumes:\n      - emptyDir: {}\n        name: app-volume\nstatus: {}\n\n---\napiVersion: apiextensions.transwarp.io/v1beta1\nkind: ReleaseConfig\nmetadata:\n  creationTimestamp: null\n  labels:\n    auto-gen: \"true\"\n  name: tomcat-test\n  namespace: %s\nspec:\n  chartAppVersion: \"7\"\n  chartName: tomcat\n  chartVersion: 0.2.0\n  configValues: {}\n  dependencies: {}\n  dependenciesConfigValues: {}\n  outputConfig: {}\n  repo: \"\"\nstatus: {}\n", namespace, namespace, namespace)
		Expect(manifestStr).To(Equal(expectedManifestStr))

	})
})


