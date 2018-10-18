package job

import (
	"sync"
	"github.com/sirupsen/logrus"
	"walm/pkg/redis"
	"github.com/google/uuid"
	"encoding/json"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"
	"fmt"
	walmerr "walm/pkg/util/error"
)

const (
	collectWalmJobsInterval time.Duration = 1 * time.Second
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
		walmJobs, err := manager.GetWalmJobsFromRedis()
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
	walmJob.Status = jobStatusRunning
	err := manager.updateWalmJob(jobId, walmJob)
	if err != nil {
		//TODO
	}
	walmJob.Run()
	manager.mutex.Lock()
	delete(manager.runningWalmJobs, jobId)
	manager.deleteWalmJobFromRedis(jobId)
	manager.mutex.Unlock()
}

func (manager *WalmJobManager) GetWalmJobsFromRedis() (map[string]*WalmJob, error) {
	walmJobStrs, err := manager.redisClient.GetClient().HGetAll(redis.WalmJobsKey).Result()
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
	_, err := manager.redisClient.GetClient().HDel(redis.WalmJobsKey, jobId).Result()
	if err != nil {
		logrus.Errorf("failed to delete walm job %s: %s", jobId, err.Error())
	} else {
		logrus.Infof("succeed to delete walm job %s", jobId)
	}
	return err
}

func (manager *WalmJobManager) updateWalmJob(jobId string, walmJob *WalmJob) (error) {
	walmJobStr, err := json.Marshal(walmJob)
	if err != nil {
		logrus.Errorf("failed to update walm job %s: %s", jobId, err.Error())
		return err
	}

	_, err = manager.redisClient.GetClient().HSet(redis.WalmJobsKey, jobId, walmJobStr).Result()
	if err != nil {
		logrus.Errorf("failed to update walm job %s: %s", walmJobStr, err.Error())
	} else {
		logrus.Infof("succeed to update walm job %s", walmJobStr)
	}
	return err
}

func (manager *WalmJobManager) GetWalmJob(jobId string) (*WalmJob, error) {
	walmJobStr, err := manager.redisClient.GetClient().HGet(redis.WalmJobsKey, jobId).Result()
	if err != nil {
		if err.Error() == redis.KeyNotFoundErrMsg {
			logrus.Errorf("walm job %s is not found in redis", jobId)
			return nil, walmerr.NotFoundError{}
		}
		logrus.Errorf("failed to get walm job %s from redis: %s", jobId, err.Error())
		return nil, err
	}

	return convertJsonStrToWalmJob(walmJobStr)
}

// if param jobId is empty, will use uuid as job id
func (manager *WalmJobManager) CreateWalmJob(jobId string, job Job) (string, error) {
	newJobId := jobId
	if newJobId == "" {
		newJobId = uuid.New().String()
	}

	walmJob := &WalmJob{
		Id:      newJobId,
		Job:     job,
		JobType: job.Type(),
		Status:  jobStatusPending,
	}
	walmJobStr, err := json.Marshal(walmJob)
	if err != nil {
		logrus.Errorf("failed to create walm job: %s", err.Error())
		return "", err
	}

	ok, err := manager.redisClient.GetClient().HSetNX(redis.WalmJobsKey, newJobId, walmJobStr).Result()
	if err != nil {
		logrus.Errorf("failed to create walm job %s : %s", walmJobStr, err.Error())
		return "", err
	}
	if !ok {
		err = fmt.Errorf("failed to create walm job : job id %s exists", newJobId)
		return "", err
	}

	logrus.Infof("succeed to create walm job : %s", walmJobStr)
	return newJobId, nil
}

type FakeJob struct {
	Name string `json:"name" description:"job name"`
}

func (fakeJob *FakeJob) Do() error {
	logrus.Infof("fake job %s is running", fakeJob.Name)
	time.Sleep(2 * time.Second)
	return nil
}
