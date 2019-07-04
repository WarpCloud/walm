package task

import (
	"github.com/RichardKnop/machinery/v1"
	"github.com/sirupsen/logrus"
	"WarpCloud/walm/pkg/setting"
	"os"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/RichardKnop/machinery/v1/backends/result"
	"github.com/RichardKnop/machinery/v1/log"
	"time"
)

var taskManager *TaskManager
var registeredTasks map[string]interface{}

func GetDefaultTaskManager() *TaskManager {
	if taskManager == nil {
		taskConfig := &config.Config{
			Broker:          setting.Config.TaskConfig.Broker,
			DefaultQueue:    setting.Config.TaskConfig.DefaultQueue,
			ResultBackend:   setting.Config.TaskConfig.ResultBackend,
			ResultsExpireIn: setting.Config.TaskConfig.ResultsExpireIn,
			NoUnixSignals:   true,
			Redis: &config.RedisConfig{
				MaxIdle:                3,
				IdleTimeout:            240,
				ReadTimeout:            15,
				WriteTimeout:           15,
				ConnectTimeout:         15,
				DelayedTasksPollPeriod: 20,
			},
		}
		server, err := machinery.NewServer(taskConfig)
		if err != nil {
			logrus.Fatalf("Failed to init work queue server: %s", err.Error())
		}
		log.Set(logrus.StandardLogger())
		taskManager = &TaskManager{
			server: server,
		}
		taskManager.server.RegisterTasks(registeredTasks)
	}
	return taskManager
}

type TaskManager struct {
	server           *machinery.Server
	worker           *machinery.Worker
}

func (manager *TaskManager) StartWorker() {
	manager.worker = manager.server.NewWorker(os.Getenv("Pod_Name"), 100)
	errorsChan := make(chan error)
	manager.worker.LaunchAsync(errorsChan)
	go func(errChan chan error) {
		if err := <-errChan; err != nil {
			logrus.Error(err.Error())
		}
	}(errorsChan)
	logrus.Info("worker starting to consume tasks")
}

func (manager *TaskManager) StopWorker() {
	quitChan := make(chan struct{})
	go func() {
		manager.worker.Quit()
		close(quitChan)
	}()
	select {
	case <-quitChan:
		logrus.Info("worker stopped consuming tasks successfully")
	case <-time.After(time.Second * 30):
		logrus.Warn("worker stopped consuming tasks failed after 30 seconds")
	}
}

func (manager *TaskManager) SendTask(signature *tasks.Signature) (err error) {
	_, err = manager.server.SendTask(signature)
	if err != nil {
		logrus.Errorf("failed to send task : %s", err.Error())
		return
	}
	logrus.Infof("succeed to send task %s", signature.Name)
	return
}

func (manager *TaskManager) PurgeTaskState(signature *tasks.Signature) (err error) {
	if signature != nil && signature.UUID != ""{
		return manager.server.GetBackend().PurgeState(signature.UUID)
	}
	return nil
}

func (manager *TaskManager) NewAsyncResult(signature *tasks.Signature) (*result.AsyncResult) {
	return result.NewAsyncResult(signature, manager.server.GetBackend())
}

func RegisterTasks(taskName string, task interface{}) {
	if registeredTasks == nil {
		registeredTasks = map[string]interface{}{}
	}
	registeredTasks[taskName] = task
}
