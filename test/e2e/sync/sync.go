package redis

import (
	. "github.com/onsi/gomega"
	. "github.com/onsi/ginkgo"
	"WarpCloud/walm/pkg/redis/impl"
	"WarpCloud/walm/pkg/setting"
	redislib "github.com/go-redis/redis"
	"WarpCloud/walm/test/e2e/framework"
	"WarpCloud/walm/pkg/sync"
	helmmocks "WarpCloud/walm/pkg/helm/mocks"
	k8smocks "WarpCloud/walm/pkg/k8s/mocks"
	taskmocks "WarpCloud/walm/pkg/task/mocks"
	"WarpCloud/walm/pkg/models/release"
	"WarpCloud/walm/pkg/models/k8s"
	"WarpCloud/walm/pkg/models/project"
	"github.com/stretchr/testify/mock"
	errorModel "WarpCloud/walm/pkg/models/error"
)

var _ = Describe("Sync", func() {

	var (
		redisClient   *redislib.Client
		mockHelm      *helmmocks.Helm
		mockK8sCache  *k8smocks.Cache
		mockTask      *taskmocks.Task
		mockSync      *sync.Sync
		testRcKey     string
		testRtKey     string
		testPtKey     string
		refreshMocks  func()
	)

	BeforeEach(func() {
		Expect(setting.Config.RedisConfig).NotTo(BeNil())

		refreshMocks = func() {
			mockHelm = &helmmocks.Helm{}
			mockK8sCache = &k8smocks.Cache{}
			mockTask = &taskmocks.Task{}

			mockSync = sync.NewSync(redisClient, mockHelm, mockK8sCache, mockTask, testRcKey, testRtKey, testPtKey)
		}
		testRcKey = framework.GenerateRandomName("sync-test-rc")
		testRtKey = framework.GenerateRandomName("sync-test-rt")
		testPtKey = framework.GenerateRandomName("sync-test-pt")

		redisClient = impl.NewRedisClient(setting.Config.RedisConfig)
		refreshMocks()
	})

	AfterEach(func() {
		if redisClient != nil {
			_, err := redisClient.Del(testRcKey, testRtKey, testPtKey).Result()
			Expect(err).NotTo(HaveOccurred())
		}
	})

	It("test sync releases & projects", func() {
		By("a release of a project is installed by helm cli")
		mockHelm.On("ListAllReleases").Return([]*release.ReleaseCache{
			{
				ReleaseSpec: release.ReleaseSpec{
					Namespace: "testns",
					Name: "testnm",
				},
			},
		}, nil)
		mockK8sCache.On("ListReleaseConfigs", "", "").Return([]*k8s.ReleaseConfig{
			{
				Meta: k8s.Meta{
					Namespace: "testns",
					Name: "testnm",
				},
				Labels: map[string]string{project.ProjectNameLabelKey : "testpj"},
			},
		}, nil)
		mockSync.Resync()
		rc, err := redisClient.HGet(testRcKey, "testns/testnm").Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(rc).To(Equal("{\"name\":\"testnm\",\"repoName\":\"\",\"configValues\":null,\"creationTimestamp\":\"\",\"version\":0,\"namespace\":\"testns\",\"dependencies\":null,\"chartName\":\"\",\"chartVersion\":\"\",\"chartAppVersion\":\"\",\"releaseResourceMetas\":null,\"computedValues\":null,\"metaInfoValues\":null,\"manifest\":\"\",\"helmVersion\":\"\",\"releasePrettyParams\":null}"))
		rt, err := redisClient.HGet(testRtKey, "testns/testnm").Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(rt).To(Equal("{\"name\":\"testnm\",\"namespace\":\"testns\",\"latestReleaseTaskSignature\":null}"))
		pt, err := redisClient.HGet(testPtKey, "testns/testpj").Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(pt).To(Equal("{\"name\":\"testpj\",\"namespace\":\"testns\",\"walmVersion\":\"v2\",\"latestTaskSignature\":null,\"latestTaskTimeoutSec\":0}"))

		By("a release of a project is updated by helm cli")
		refreshMocks()
		mockHelm.On("ListAllReleases").Return([]*release.ReleaseCache{
			{
				ReleaseSpec: release.ReleaseSpec{
					Namespace: "testns",
					Name: "testnm",
					ChartName: "testct",
				},
			},
		}, nil)
		mockK8sCache.On("ListReleaseConfigs", "", "").Return([]*k8s.ReleaseConfig{
			{
				Meta: k8s.Meta{
					Namespace: "testns",
					Name: "testnm",
				},
				Labels: map[string]string{project.ProjectNameLabelKey : "testpj"},
			},
		}, nil)
		mockSync.Resync()
		rc, err = redisClient.HGet(testRcKey, "testns/testnm").Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(rc).To(Equal("{\"name\":\"testnm\",\"repoName\":\"\",\"configValues\":null,\"creationTimestamp\":\"\",\"version\":0,\"namespace\":\"testns\",\"dependencies\":null,\"chartName\":\"testct\",\"chartVersion\":\"\",\"chartAppVersion\":\"\",\"releaseResourceMetas\":null,\"computedValues\":null,\"metaInfoValues\":null,\"manifest\":\"\",\"helmVersion\":\"\",\"releasePrettyParams\":null}"))
		rt, err = redisClient.HGet(testRtKey, "testns/testnm").Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(rt).To(Equal("{\"name\":\"testnm\",\"namespace\":\"testns\",\"latestReleaseTaskSignature\":null}"))
		pt, err = redisClient.HGet(testPtKey, "testns/testpj").Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(pt).To(Equal("{\"name\":\"testpj\",\"namespace\":\"testns\",\"walmVersion\":\"v2\",\"latestTaskSignature\":null,\"latestTaskTimeoutSec\":0}"))

		By("a release of a project is deleted by helm cli")
		refreshMocks()
		mockHelm.On("ListAllReleases").Return([]*release.ReleaseCache{}, nil)
		mockK8sCache.On("ListReleaseConfigs", "", "").Return([]*k8s.ReleaseConfig{}, nil)
		mockTask.On("GetTaskState", mock.Anything).Return(nil, errorModel.NotFoundError{})
		mockSync.Resync()
		rc, err = redisClient.HGet(testRcKey, "testns/testnm").Result()
		Expect(err).To(HaveOccurred())
		rt, err = redisClient.HGet(testRtKey, "testns/testnm").Result()
		Expect(err).To(HaveOccurred())
		pt, err = redisClient.HGet(testPtKey, "testns/testpj").Result()
		Expect(err).To(HaveOccurred())
	})

})
