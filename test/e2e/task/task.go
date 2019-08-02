package task

import (
	. "github.com/onsi/gomega"
	. "github.com/onsi/ginkgo"
	"WarpCloud/walm/pkg/task/machinery"
	"WarpCloud/walm/pkg/setting"
	"WarpCloud/walm/test/e2e/framework"
	machinerylib "github.com/RichardKnop/machinery/v1"
	"WarpCloud/walm/pkg/redis/impl"
	"github.com/sirupsen/logrus"
	"time"
	errorModel "WarpCloud/walm/pkg/models/error"
	"fmt"
)

var _ = Describe("Task", func() {

	var (
		taskImpl  *machinery.Task
		testQueue string
	)

	BeforeEach(func() {
		Expect(setting.Config.TaskConfig).NotTo(BeNil())

		testQueue = framework.GenerateRandomName("task-test")
		taskConfig := setting.Config.TaskConfig
		taskConfig.DefaultQueue = testQueue
		taskConfig.ResultsExpireIn = 60
		var err error
		taskImpl, err = machinery.NewTask(taskConfig)
		Expect(err).NotTo(HaveOccurred())
		taskImpl.StartWorker()
	})

	AfterEach(func() {
		if setting.Config.TaskConfig != nil {
			if taskImpl != nil {
				taskImpl.StopWorker(5)
			}
			host, pwd, db, err := machinerylib.ParseRedisURL(setting.Config.TaskConfig.Broker)
			Expect(err).NotTo(HaveOccurred())

			redisConfig := &setting.RedisConfig{
				Addr:     host,
				Password: pwd,
				DB:       db,
			}
			redisClient := impl.NewRedisClient(redisConfig)
			By(fmt.Sprintf("delete redis list %s", testQueue))
			_, err = redisClient.Del(testQueue).Result()
			Expect(err).NotTo(HaveOccurred())
		}
	})

	It("test task lifycycle", func() {
		By("register task")
		testTask := func(string) error {
			logrus.Infof("testing")
			time.Sleep(time.Second * 2)
			return nil
		}
		err := taskImpl.RegisterTask("test-task", testTask)
		Expect(err).NotTo(HaveOccurred())

		taskSig, err := taskImpl.SendTask("test-task", "", 5)
		Expect(err).NotTo(HaveOccurred())
		Expect(taskSig.Name).To(Equal("test-task"))
		Expect(taskSig.Arg).To(Equal(""))
		Expect(taskSig.TimeoutSec).To(Equal(int64(5)))

		time.Sleep(time.Millisecond * 500)

		taskState, err := taskImpl.GetTaskState(taskSig)
		Expect(err).NotTo(HaveOccurred())
		Expect(taskState.IsFinished()).To(BeFalse())
		Expect(taskState.IsSuccess()).To(BeFalse())
		Expect(taskState.IsTimeout()).To(BeFalse())
		Expect(taskState.GetErrorMsg()).To(Equal(""))

		taskState, err = taskImpl.GetTaskState(nil)
		Expect(err).To(Equal(errorModel.NotFoundError{}))

		err = taskImpl.TouchTask(taskSig, 1)
		Expect(err).NotTo(HaveOccurred())

		err = taskImpl.PurgeTaskState(taskSig)
		Expect(err).NotTo(HaveOccurred())

		taskState, err = taskImpl.GetTaskState(taskSig)
		Expect(err).To(Equal(errorModel.NotFoundError{}))
	})

})
