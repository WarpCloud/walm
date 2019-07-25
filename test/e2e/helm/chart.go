package helm

import (
	. "github.com/onsi/gomega"
	. "github.com/onsi/ginkgo"
	"WarpCloud/walm/pkg/helm/impl"
	"WarpCloud/walm/test/e2e/framework"
	"WarpCloud/walm/pkg/k8s/cache/informer"
	"WarpCloud/walm/pkg/setting"
	"WarpCloud/walm/pkg/models/release"
	"path/filepath"
)

var _ = Describe("HelmChart", func() {

	var (
		//namespace string
		helm     *impl.Helm
		err      error
		stopChan chan struct{}
	)

	BeforeEach(func() {
		stopChan = make(chan struct{})
		k8sCache := informer.NewInformer(framework.GetK8sClient(), framework.GetK8sReleaseConfigClient(), 0, stopChan)
		registryClient := impl.NewRegistryClient(setting.Config.ChartImageConfig)

		helm, err = impl.NewHelm(setting.Config.RepoList, registryClient, k8sCache, framework.GetKubeClient())
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		close(stopChan)
	})

	It("test get chart auto dependency", func() {
		currentFilePath, err := framework.GetCurrentFilePath()
		Expect(err).NotTo(HaveOccurred())

		chartPath := filepath.Join(filepath.Dir(currentFilePath), "../../resources/helm/kafka-6.1.0.tgz")
		Expect(err).NotTo(HaveOccurred())

		testRepoUrl := ""
		for _, repo := range helm.GetRepoList().Items {
			if repo.TenantRepoName == framework.TestChartRepoName {
				testRepoUrl = repo.TenantRepoURL
			}
		}
		Expect(testRepoUrl).NotTo(Equal(""))

		err = framework.PushChartToRepo(testRepoUrl, chartPath)
		Expect(err).NotTo(HaveOccurred())

		dependencies, err := helm.GetChartAutoDependencies(framework.TestChartRepoName, "kafka", "6.1.0")
		Expect(dependencies).To(Equal([]string{"zookeeper"}))
	})

	It("test get chart list", func() {
		chartList, err := helm.GetChartList(framework.TestChartRepoName)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(chartList.Items) >= 1).To(BeTrue())

		getChart := func(chartName, chartVersion string) *release.ChartInfo {
			for _, chartInfo := range chartList.Items {
				if chartInfo.ChartName == chartName && chartInfo.ChartVersion == chartVersion{
					return chartInfo
				}
			}
			return nil
		}

		testChart := getChart(framework.TestChartName, framework.TestChartVersion)
		Expect(testChart).NotTo(BeNil())
		Expect(testChart.ChartAppVersion).To(Equal("7"))
		Expect(testChart.ChartDescription).To(Equal("开源的轻量级Web应用服务器，支持HTML、JS等静态资源的处理，可以作为轻量级Web服务器使用。"))
	})

	It("test get chart", func() {
		chartInfo, err := helm.GetChartDetailInfo(framework.TestChartRepoName, framework.TestChartName, framework.TestChartVersion)
		Expect(err).NotTo(HaveOccurred())
		Expect(chartInfo.Icon).NotTo(Equal(""))
		Expect(chartInfo.Advantage).NotTo(Equal(""))
		Expect(chartInfo.Architecture).NotTo(Equal(""))
		Expect(chartInfo.ChartInfo).To(Equal(release.ChartInfo{
			ChartVersion:     framework.TestChartVersion,
			ChartName:        framework.TestChartName,
			ChartAppVersion:  "7",
			DefaultValue:     "{\"affinity\":{},\"app\":{\"path\":\"/sample\"},\"deploy\":{\"directory\":\"/usr/local/tomcat/webapps\"},\"image\":{\"pullPolicy\":\"Always\",\"pullSecrets\":[],\"tomcat\":{\"image\":\"tomcat:7.0\"},\"webarchive\":{\"image\":\"ananwaresystems/webarchive:1.0\"}},\"ingress\":{\"annotations\":{},\"enabled\":false,\"hosts\":[\"chart-example.local\"],\"path\":\"/\",\"tls\":[]},\"livenessProbe\":{\"initialDelaySeconds\":60,\"periodSeconds\":30},\"nodeSelector\":{},\"readinessProbe\":{\"failureThreshold\":6,\"initialDelaySeconds\":60,\"periodSeconds\":30},\"replicaCount\":1,\"resources\":{\"limits\":{\"cpu\":0.2,\"memory\":\"200Mi\"},\"requests\":{\"cpu\":0.1,\"memory\":\"100Mi\"}},\"service\":{\"externalPort\":80,\"internalPort\":8080,\"name\":\"http\",\"type\":\"NodePort\"},\"tolerations\":[]}",
			ChartDescription: "开源的轻量级Web应用服务器，支持HTML、JS等静态资源的处理，可以作为轻量级Web服务器使用。",
			ChartEngine:      "transwarp",
			MetaInfo: &release.ChartMetaInfo{
				FriendlyName: "tomcat",
				Categories:   []string{"tdh", "tdc"},
				ChartRoles: []*release.MetaRoleConfig{
					{
						Name:        "webarchive",
						Description: "webarchive",
						Type:        "initContainer",
						RoleBaseConfig: &release.MetaRoleBaseConfig{
							Image: &release.MetaStringConfig{
								Type:         "string",
								MapKey:       "image.webarchive.image",
								Description:  "镜像",
								DefaultValue: "ananwaresystems/webarchive:1.0",
								Required:     true,
							},
						},
						RoleHealthCheckConfig: &release.MetaHealthCheckConfig{
							LivenessProbe:  &release.MetaHealthProbeConfig{},
							ReadinessProbe: &release.MetaHealthProbeConfig{},
						},
					},
					{
						Name:        "tomcat",
						Description: "tomcat",
						Type:        "container",
						RoleBaseConfig: &release.MetaRoleBaseConfig{
							Image: &release.MetaStringConfig{
								Type:         "string",
								MapKey:       "image.tomcat.image",
								Description:  "镜像",
								DefaultValue: "tomcat:7.0",
								Required:     true,
							},
							Replicas: &release.MetaIntConfig{
								IntConfig: release.IntConfig{
									Type:         "number",
									MapKey:       "replicaCount",
									Description:  "副本个数",
									DefaultValue: 1,
									Required:     true,
								},
							},
							Others: []*release.MetaCommonConfig{
								{
									Name:         "path",
									MapKey:       "app.path",
									Description:  "appPath",
									Type:         "string",
									Required:     true,
									DefaultValue: "\"/sample\"",
								},
							},
						},
						RoleResourceConfig: &release.MetaResourceConfig{
							LimitsMemory: &release.MetaResourceMemoryConfig{
								IntConfig: release.IntConfig{
									MapKey:       "resources.limits.memory",
									DefaultValue: 200,
								},
							},
							RequestsMemory:&release.MetaResourceMemoryConfig{
								IntConfig: release.IntConfig{
									MapKey:       "resources.requests.memory",
									DefaultValue: 100,
								},
							},
							LimitsCpu: &release.MetaResourceCpuConfig{
								FloatConfig: release.FloatConfig{
									MapKey:       "resources.limits.cpu",
									DefaultValue: 0.2,
								},
							},
							RequestsCpu: &release.MetaResourceCpuConfig{
								FloatConfig: release.FloatConfig{
									MapKey:       "resources.requests.cpu",
									DefaultValue: 0.1,
								},
							},
						},
						RoleHealthCheckConfig: &release.MetaHealthCheckConfig{
							LivenessProbe:  &release.MetaHealthProbeConfig{},
							ReadinessProbe: &release.MetaHealthProbeConfig{},
						},
					},
				},
			},
		}))
	})

	It("test get chart by image", func() {
		chartInfo, err := helm.GetDetailChartInfoByImage(framework.GetTestChartImage())
		Expect(err).NotTo(HaveOccurred())
		Expect(chartInfo.Icon).NotTo(Equal(""))
		Expect(chartInfo.Advantage).NotTo(Equal(""))
		Expect(chartInfo.Architecture).NotTo(Equal(""))
		Expect(chartInfo.ChartInfo).To(Equal(release.ChartInfo{
			ChartVersion:     framework.TestChartVersion,
			ChartName:        framework.TestChartName,
			ChartAppVersion:  "7",
			DefaultValue:     "{\"affinity\":{},\"app\":{\"path\":\"/sample\"},\"deploy\":{\"directory\":\"/usr/local/tomcat/webapps\"},\"image\":{\"pullPolicy\":\"Always\",\"pullSecrets\":[],\"tomcat\":{\"image\":\"tomcat:7.0\"},\"webarchive\":{\"image\":\"ananwaresystems/webarchive:1.0\"}},\"ingress\":{\"annotations\":{},\"enabled\":false,\"hosts\":[\"chart-example.local\"],\"path\":\"/\",\"tls\":[]},\"livenessProbe\":{\"initialDelaySeconds\":60,\"periodSeconds\":30},\"nodeSelector\":{},\"readinessProbe\":{\"failureThreshold\":6,\"initialDelaySeconds\":60,\"periodSeconds\":30},\"replicaCount\":1,\"resources\":{\"limits\":{\"cpu\":0.2,\"memory\":\"200Mi\"},\"requests\":{\"cpu\":0.1,\"memory\":\"100Mi\"}},\"service\":{\"externalPort\":80,\"internalPort\":8080,\"name\":\"http\",\"type\":\"NodePort\"},\"tolerations\":[]}",
			ChartDescription: "开源的轻量级Web应用服务器，支持HTML、JS等静态资源的处理，可以作为轻量级Web服务器使用。",
			ChartEngine:      "transwarp",
			MetaInfo: &release.ChartMetaInfo{
				FriendlyName: "tomcat",
				Categories:   []string{"tdh", "tdc"},
				ChartRoles: []*release.MetaRoleConfig{
					{
						Name:        "webarchive",
						Description: "webarchive",
						Type:        "initContainer",
						RoleBaseConfig: &release.MetaRoleBaseConfig{
							Image: &release.MetaStringConfig{
								Type:         "string",
								MapKey:       "image.webarchive.image",
								Description:  "镜像",
								DefaultValue: "ananwaresystems/webarchive:1.0",
								Required:     true,
							},
						},
						RoleHealthCheckConfig: &release.MetaHealthCheckConfig{
							LivenessProbe:  &release.MetaHealthProbeConfig{},
							ReadinessProbe: &release.MetaHealthProbeConfig{},
						},
					},
					{
						Name:        "tomcat",
						Description: "tomcat",
						Type:        "container",
						RoleBaseConfig: &release.MetaRoleBaseConfig{
							Image: &release.MetaStringConfig{
								Type:         "string",
								MapKey:       "image.tomcat.image",
								Description:  "镜像",
								DefaultValue: "tomcat:7.0",
								Required:     true,
							},
							Replicas: &release.MetaIntConfig{
								IntConfig: release.IntConfig{
									Type:         "number",
									MapKey:       "replicaCount",
									Description:  "副本个数",
									DefaultValue: 1,
									Required:     true,
								},
							},
							Others: []*release.MetaCommonConfig{
								{
									Name:         "path",
									MapKey:       "app.path",
									Description:  "appPath",
									Type:         "string",
									Required:     true,
									DefaultValue: "\"/sample\"",
								},
							},
						},
						RoleResourceConfig: &release.MetaResourceConfig{
							LimitsMemory: &release.MetaResourceMemoryConfig{
								IntConfig: release.IntConfig{
									MapKey:       "resources.limits.memory",
									DefaultValue: 200,
								},
							},
							RequestsMemory:&release.MetaResourceMemoryConfig{
								IntConfig: release.IntConfig{
									MapKey:       "resources.requests.memory",
									DefaultValue: 100,
								},
							},
							LimitsCpu: &release.MetaResourceCpuConfig{
								FloatConfig: release.FloatConfig{
									MapKey:       "resources.limits.cpu",
									DefaultValue: 0.2,
								},
							},
							RequestsCpu: &release.MetaResourceCpuConfig{
								FloatConfig: release.FloatConfig{
									MapKey:       "resources.requests.cpu",
									DefaultValue: 0.1,
								},
							},
						},
						RoleHealthCheckConfig: &release.MetaHealthCheckConfig{
							LivenessProbe:  &release.MetaHealthProbeConfig{},
							ReadinessProbe: &release.MetaHealthProbeConfig{},
						},
					},
				},
			},
		}))
	})
})
