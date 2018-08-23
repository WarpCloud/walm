package job

import (
	"github.com/sirupsen/logrus"
	"testing"
	"walm/pkg/redis"
	"time"
	"sync"
	"k8s.io/apimachinery/pkg/util/wait"
)

func TestWalmJobManager(t *testing.T) {
	redisClient := redis.CreateFakeRedisClient()
	manager := &WalmJobManager{redisClient: redisClient, collectInterval: 1*time.Second, mutex: &sync.Mutex{}, runningWalmJobs: map[string]*WalmJob{}}
	manager.Start(wait.NeverStop)

	err := manager.CreateWalmJob("fake", &FakeJob{"test1"})
	if err != nil {
		logrus.Error(err.Error())
		t.Fail()
	}

	time.Sleep(1 * time.Second)
	err = manager.CreateWalmJob("fake", &FakeJob{"test2"})
	if err != nil {
		logrus.Error(err.Error())
		t.Fail()
	}
	time.Sleep(3 * time.Second)
}