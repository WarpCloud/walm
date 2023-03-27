package redis

import (
	. "github.com/onsi/gomega"
	. "github.com/onsi/ginkgo"
	"WarpCloud/walm/pkg/redis/impl"
	"WarpCloud/walm/pkg/setting"
	"WarpCloud/walm/pkg/models/release"
	"encoding/json"
	"WarpCloud/walm/pkg/redis"
	redislib "github.com/go-redis/redis"
	errorModel "WarpCloud/walm/pkg/models/error"
	"WarpCloud/walm/test/e2e/framework"
)

var _ = Describe("Redis", func() {

	var (
		redisClient *redislib.Client
		redisImpl   *impl.Redis
		testKey string
	)

	BeforeEach(func() {
		Expect(setting.Config.RedisConfig).NotTo(BeNil())

		testKey = framework.GenerateRandomName("redis-test")
		redisClient = impl.NewRedisClient(setting.Config.RedisConfig)
		redisImpl = impl.NewRedis(redisClient)
	})

	AfterEach(func() {
		if redisClient != nil {
			_, err := redisClient.Del(testKey).Result()
			Expect(err).NotTo(HaveOccurred())
		}
	})

	It("test redis lifycycle", func() {
		By("set field to redis")
		releaseCache1 := &release.ReleaseCache{}
		releaseCache1.Namespace = "testns"
		releaseCache1.Name = "testnm1"
		field1 := redis.BuildFieldName(releaseCache1.Namespace, releaseCache1.Name)

		releaseCache1Str, err := json.Marshal(releaseCache1)
		Expect(err).NotTo(HaveOccurred())

		releaseCache2 := &release.ReleaseCache{}
		releaseCache2.Namespace = "testns"
		releaseCache2.Name = "testnm2"
		field2 := redis.BuildFieldName(releaseCache2.Namespace, releaseCache2.Name)

		releaseCache2Str, err := json.Marshal(releaseCache2)
		Expect(err).NotTo(HaveOccurred())

		err = redisImpl.SetFieldValues(testKey, map[string]interface{}{
			field1: releaseCache1,
			field2: releaseCache2,
		})
		Expect(err).NotTo(HaveOccurred())

		By("get field")
		value, err := redisImpl.GetFieldValue(testKey, releaseCache1.Namespace, releaseCache1.Name)
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(Equal(string(releaseCache1Str)))

		_, err = redisImpl.GetFieldValue(testKey, "notexisted", "notexisted")
		Expect(err).To(Equal(errorModel.NotFoundError{}))

		values, err := redisImpl.GetFieldValues(testKey, releaseCache1.Namespace, "")
		Expect(err).NotTo(HaveOccurred())
		Expect(values).To(HaveLen(2))
		Expect(values).To(ConsistOf([]string{string(releaseCache1Str), string(releaseCache2Str)}))

		values, err = redisImpl.GetFieldValues(testKey, "", "")
		Expect(err).NotTo(HaveOccurred())
		Expect(values).To(HaveLen(2))
		Expect(values).To(ConsistOf([]string{string(releaseCache1Str), string(releaseCache2Str)}))

		values, err = redisImpl.GetFieldValuesByNames(testKey,
			redis.BuildFieldName(releaseCache1.Namespace, releaseCache1.Name),
			redis.BuildFieldName(releaseCache2.Namespace, releaseCache2.Name))
		Expect(err).NotTo(HaveOccurred())
		Expect(values).To(Equal([]string{string(releaseCache1Str), string(releaseCache2Str)}))


		By("delete field")
		err = redisImpl.DeleteField(testKey, releaseCache1.Namespace, releaseCache1.Name)
		Expect(err).NotTo(HaveOccurred())

		_, err = redisImpl.GetFieldValue(testKey, releaseCache1.Namespace, releaseCache1.Name)
		Expect(err).To(Equal(errorModel.NotFoundError{}))

		err = redisImpl.DeleteField(testKey, "notexisted", "notexisted")
		Expect(err).NotTo(HaveOccurred())
	})

})
