package helm

import (
	"walm/pkg/task"
	"github.com/sirupsen/logrus"
	"encoding/json"
	"github.com/RichardKnop/machinery/v1/tasks"
	"walm/pkg/release/manager/helm/cache"
	"time"
)

type ReleaseTaskArgs interface {
	Run() error
	GetTaskName() string
}

func SendReleaseTask(helmCache *cache.HelmCache, namespace, releaseName string, releaseTaskArgs ReleaseTaskArgs, oldReleaseTask *cache.ReleaseTask, timeoutSec int64, async bool) (error) {
	releaseTaskArgsStr, err := json.Marshal(releaseTaskArgs)
	if err != nil {
		logrus.Errorf("failed to marshal %s args : %s", releaseTaskArgs.GetTaskName(), err.Error())
		return err
	}
	releaseTaskSig := &tasks.Signature{
		Name: releaseTaskArgs.GetTaskName(),
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: string(releaseTaskArgsStr),
			},
		},
	}
	err = task.GetDefaultTaskManager().SendTask(releaseTaskSig)
	if err != nil {
		logrus.Errorf("failed to send %s : %s", releaseTaskArgs.GetTaskName(), err.Error())
		return err
	}

	taskSig := &task.WalmTaskSig{
		Name: releaseTaskArgs.GetTaskName(),
		UUID: releaseTaskSig.UUID,
		Arg:  string(releaseTaskArgsStr),
		TimeoutSec: timeoutSec,
	}

	releaseTask := &cache.ReleaseTask{
		Namespace:            namespace,
		Name:                 releaseName,
		LatestReleaseTaskSig: taskSig,
	}

	err = helmCache.CreateOrUpdateReleaseTask(releaseTask)
	if err != nil {
		logrus.Errorf("failed to set release task of %s/%s to redis: %s", namespace, releaseName, err.Error())
		return err
	}

	if oldReleaseTask != nil && oldReleaseTask.LatestReleaseTaskSig != nil {
		err = task.GetDefaultTaskManager().PurgeTaskState(oldReleaseTask.LatestReleaseTaskSig.GetTaskSignature())
		if err != nil {
			logrus.Warnf("failed to purge task state : %s", err.Error())
		}
	}

	if !async {
		asyncResult := taskSig.GetAsyncResult()
		_, err = asyncResult.GetWithTimeout(time.Duration(timeoutSec)*time.Second, defaultSleepTimeSecond)
		if err != nil {
			logrus.Errorf("release task %s of %s/%s is failed or timeout: %s", releaseTaskArgs.GetTaskName(), namespace, releaseName, err.Error())
			return err
		}
	}

	return nil
}