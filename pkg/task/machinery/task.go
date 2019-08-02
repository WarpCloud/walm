package machinery

import (
	taskModel "WarpCloud/walm/pkg/models/task"
	"WarpCloud/walm/pkg/task"
	errorModel "WarpCloud/walm/pkg/models/error"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/backends/result"
	"github.com/sirupsen/logrus"
	"time"
	"github.com/RichardKnop/machinery/v1/config"
	"WarpCloud/walm/pkg/setting"
	"github.com/RichardKnop/machinery/v1/log"
	"os"
)

type Task struct {
	server *machinery.Server
	worker *machinery.Worker
}

func (task *Task) GetTaskState(sig *taskModel.TaskSig) (state task.TaskState, err error) {
	taskSig := convertTaskSig(sig)
	if taskSig == nil {
		return nil, errorModel.NotFoundError{}
	}
	asyncResult := result.NewAsyncResult(taskSig, task.server.GetBackend())
	taskState := asyncResult.GetState()
	if taskState == nil || taskState.TaskName == "" {
		return nil, errorModel.NotFoundError{}
	}
	state = &TaskStateAdaptor{
		taskState:      taskState,
		taskTimeoutSec: sig.TimeoutSec,
	}
	return
}

func convertTaskSig(sig *taskModel.TaskSig) *tasks.Signature {
	if sig == nil || sig.UUID == "" {
		return nil
	}
	return &tasks.Signature{
		Name: sig.Name,
		UUID: sig.UUID,
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: sig.Arg,
			},
		},
	}
}

func (task *Task) RegisterTask(taskName string, taskRunner func(taskArgs string) error) error{
	err := task.server.RegisterTask(taskName, taskRunner)
	if err != nil {
		logrus.Errorf("failed to register task %s : %s", taskName, err.Error())
		return err
	}
	return nil
}

func (task *Task) SendTask(taskName, taskArgs string, timeoutSec int64) (*taskModel.TaskSig, error){
	taskSig := &tasks.Signature{
		Name: taskName,
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: taskArgs,
			},
		},
	}
	_, err := task.server.SendTask(taskSig)
	if err != nil {
		logrus.Errorf("failed to send %s : %s", taskName, err.Error())
		return nil, err
	}

	sig := &taskModel.TaskSig{
		Name:       taskName,
		UUID:       taskSig.UUID,
		Arg:        taskArgs,
		TimeoutSec: timeoutSec,
	}
	return sig, nil
}

func (task *Task) TouchTask(sig *taskModel.TaskSig, pollingIntervalSec int64) (error){
	taskSig := convertTaskSig(sig)
	if taskSig == nil {
		return errorModel.NotFoundError{}
	}
	asyncResult := result.NewAsyncResult(taskSig, task.server.GetBackend())
	_, err := asyncResult.GetWithTimeout(time.Duration(sig.TimeoutSec)*time.Second, time.Duration(pollingIntervalSec) * time.Second)
	if err != nil {
		logrus.Errorf("touch task %s-%s failed: %s", sig.Name, sig.UUID, err.Error())
		return err
	}
	return nil
}

func (task *Task) PurgeTaskState(sig *taskModel.TaskSig) (error){
	if sig == nil || sig.UUID == ""{
		return nil
	}
	err := task.server.GetBackend().PurgeState(sig.UUID)
	if err != nil {
		logrus.Errorf("failed to purge task state : %s", err.Error())
		return err
	}
	return nil
}

func (task *Task) StartWorker() {
	task.worker = task.server.NewWorker(os.Getenv("Pod_Name"), 100)
	errorsChan := make(chan error)
	task.worker.LaunchAsync(errorsChan)
	go func(errChan chan error) {
		if err := <-errChan; err != nil {
			logrus.Error(err.Error())
		}
	}(errorsChan)
	logrus.Info("worker starting to consume tasks")
}

func (task *Task) StopWorker(timeoutSec int64) {
	quitChan := make(chan struct{})
	go func() {
		task.worker.Quit()
		close(quitChan)
	}()
	select {
	case <-quitChan:
		logrus.Info("worker stopped consuming tasks successfully")
	case <-time.After(time.Second * time.Duration(timeoutSec)):
		logrus.Warn("worker stopped consuming tasks failed after 30 seconds")
	}
}

func NewTask(c *setting.TaskConfig) (*Task, error) {
	taskConfig := &config.Config{
		Broker:          c.Broker,
		DefaultQueue:    c.DefaultQueue,
		ResultBackend:   c.ResultBackend,
		ResultsExpireIn: c.ResultsExpireIn,
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
		logrus.Errorf("Failed to create task server: %s", err.Error())
		return nil, err
	}
	log.Set(logrus.StandardLogger())
	return &Task{
		server: server,
	}, nil
}