package helm

import (
	"github.com/sirupsen/logrus"
	"WarpCloud/walm/pkg/models/release"
	"encoding/json"
)

const (
	defaultSleepTimeSecond int64 = 1
)

func (helm *Helm) sendReleaseTask(namespace, releaseName , taskName string, taskArgs interface{}, oldReleaseTask *release.ReleaseTask, timeoutSec int64, async bool) (error) {
	taskArgsStr, err := json.Marshal(taskArgs)
	if err != nil {
		logrus.Errorf("failed to marshal task args : %s", err.Error())
		return err
	}

	taskSig, err := helm.task.SendTask(taskName, string(taskArgsStr), timeoutSec)
	if err != nil {
		logrus.Errorf("failed to send %s : %s", taskName, err.Error())
		return err
	}

	releaseTask := &release.ReleaseTask{
		Namespace:            namespace,
		Name:                 releaseName,
		LatestReleaseTaskSig: taskSig,
	}

	err = helm.releaseCache.CreateOrUpdateReleaseTask(releaseTask)
	if err != nil {
		logrus.Errorf("failed to set release task of %s/%s to redis: %s", namespace, releaseName, err.Error())
		return err
	}

	if oldReleaseTask != nil && oldReleaseTask.LatestReleaseTaskSig != nil {
		_ = helm.task.PurgeTaskState(oldReleaseTask.LatestReleaseTaskSig)
	}

	if !async {
		err = helm.task.TouchTask(taskSig, defaultSleepTimeSecond)
		if err != nil {
			logrus.Errorf("release task %s of %s/%s is failed or timeout: %s", taskName, namespace, releaseName, err.Error())
			return err
		}
	}

	return nil
}
