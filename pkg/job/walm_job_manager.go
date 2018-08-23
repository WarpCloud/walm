package job

import (
	"sync"
	"github.com/sirupsen/logrus"
	"walm/pkg/redis"
	"github.com/google/uuid"
	"encoding/json"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"
)

const (
	walmJobsKey                           = "walm-jobs"
	collectWalmJobsInterval time.Duration = 10 * time.Second
)

var walmJobManager *WalmJobManager

func GetDefaultWalmJobManager() *WalmJobManager {
	return walmJobManager
}

func InitWalmJobManager() {
	walmJobManager = &WalmJobManager{redisClient: redis.GetDefaultRedisClient()}
	walmJobManager.collectInterval = collectWalmJobsInterval
	walmJobManager.runningWalmJobs = map[string]*WalmJob{}
	walmJobManager.mutex = &sync.Mutex{}
	logrus.Info("walm job manager inited")
}

type WalmJobManager struct {
	runningWalmJobs map[string]*WalmJob
	mutex           *sync.Mutex
	started         bool
	collectInterval time.Duration
	redisClient     *redis.RedisClient
}

func (manager *WalmJobManager) Start(stopCh <-chan struct{}) {
	if manager.started {
		logrus.Warn("walm job manager has been started before")
		return
	}
	manager.started = true
	logrus.Info("walm job manager started")
	go manager.collectAndRunWalmJobs(stopCh)
}

func (manager *WalmJobManager) Stop() {
	manager.started = false
	logrus.Info("walm job manager stopped")
}

func (manager *WalmJobManager) collectAndRunWalmJobs(stopCh <-chan struct{}) {
	wait.NonSlidingUntil(func() {
		manager.mutex.Lock()
		defer manager.mutex.Unlock()
		walmJobs, err := manager.getWalmJobsFromRedis()
		if err != nil {
			logrus.Errorf("Failed to collect and run walm jobs, will retry after %v", manager.collectInterval)
			return
		}
		for jobId, walmJob := range walmJobs {
			if _, ok := manager.runningWalmJobs[jobId]; !ok {
				manager.runningWalmJobs[jobId] = walmJob
				go manager.runWalmJob(jobId, walmJob)
			}
		}
	}, manager.collectInterval, stopCh)
}

func (manager *WalmJobManager) runWalmJob(jobId string, walmJob *WalmJob) {
	walmJob.Run()
	manager.mutex.Lock()
	delete(manager.runningWalmJobs, jobId)
	manager.deleteWalmJobFromRedis(jobId)
	manager.mutex.Unlock()
}

func (manager *WalmJobManager) getWalmJobsFromRedis() (map[string]*WalmJob, error) {
	walmJobStrs, err := manager.redisClient.GetClient().HGetAll(walmJobsKey).Result()
	if err != nil {
		logrus.Errorf("failed to get walm jobs from redis: %s", err.Error())
		return nil, err
	}

	walmJobs := map[string]*WalmJob{}
	for jobId, walmJobStr := range walmJobStrs {
		walmJob, err := convertJsonStrToWalmJob(walmJobStr)
		if err != nil {
			continue
		}
		walmJobs[jobId] = walmJob
	}
	return walmJobs, nil
}

func convertJsonStrToWalmJob(walmJobStr string) (walmJob *WalmJob, err error) {
	walmJobAdaptor := &WalmJobAdaptor{}
	err = json.Unmarshal([]byte(walmJobStr), walmJobAdaptor)
	if err != nil {
		logrus.Errorf("failed to unmarshal job %s : %s", walmJobStr, err.Error())
		return
	}

	return walmJobAdaptor.GetWalmJob()
}

func (manager *WalmJobManager) deleteWalmJobFromRedis(jobId string) error {
	_, err := manager.redisClient.GetClient().HDel(walmJobsKey, jobId).Result()
	if err != nil {
		logrus.Errorf("failed to delete walm job %s: %s", jobId, err.Error())
	} else {
		logrus.Infof("succeed to delete walm job %s", jobId)
	}
	return err
}

func (manager *WalmJobManager) CreateWalmJob(jobType string, job Job) error {
	newJobId := uuid.New().String()
	walmJob := &WalmJob{
		Id:      newJobId,
		Job:     job,
		JobType: jobType,
	}
	walmJobStr, err := json.Marshal(walmJob)
	if err != nil {
		logrus.Errorf("failed to create walm job: %s", err.Error())
		return err
	}

	for {
		ok, err := manager.redisClient.GetClient().HSetNX(walmJobsKey, newJobId, walmJobStr).Result()
		if err != nil {
			logrus.Errorf("failed to create walm job %s : %s", walmJobStr, err.Error())
			return err
		}
		if !ok {
			logrus.Warn("job id %s exists in redis, should recreate the job id")
			newJobId = uuid.New().String()
		} else {
			break
		}
	}
	logrus.Infof("succeed to create walm job : %s", walmJobStr)
	return nil
}

type FakeJob struct {
	Name string
}

func (fakeJob *FakeJob) Do() error {
	logrus.Infof("fake job %s is running", fakeJob.Name)
	time.Sleep(2 * time.Second)
	return nil
}
