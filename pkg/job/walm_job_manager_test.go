package job

import (
	"github.com/sirupsen/logrus"
	"testing"
	"walm/pkg/redis"
	"time"
	"sync"
	"k8s.io/apimachinery/pkg/util/wait"
	"fmt"
)

func TestWalmJobManager(t *testing.T) {
	redisClient := redis.CreateFakeRedisClient()
	manager := &WalmJobManager{redisClient: redisClient, collectInterval: 1*time.Second, mutex: &sync.Mutex{}, runningWalmJobs: map[string]*WalmJob{}}
	manager.Start(wait.NeverStop)

	id, err := manager.CreateWalmJob("","fake", &FakeJob{"test1"})
	if err != nil {
		logrus.Error(err.Error())
	}

	walmJob, err := manager.GetWalmJob(id)
	if err != nil {
		logrus.Error(err.Error())
	}
	fmt.Println(walmJob.Status)

	time.Sleep(2 * time.Second)
	walmJob, err = manager.GetWalmJob(id)
	if err != nil {
		logrus.Error(err.Error())
	}
	fmt.Println(walmJob.Status)

	_, err = manager.CreateWalmJob("", "fake", &FakeJob{"test2"})
	if err != nil {
		logrus.Error(err.Error())
		t.Fail()
	}
	time.Sleep(4 * time.Second)

	walmJob, err = manager.GetWalmJob(id)
	if err != nil {
		logrus.Error(err.Error())
	}
}