package helm

import (
	"github.com/sirupsen/logrus"
	"strings"
	"time"
	errorModel "WarpCloud/walm/pkg/models/error"
)

func (helm *Helm) DeleteReleaseWithRetry(namespace, releaseName string, deletePvcs bool, async bool, timeoutSec int64) error {
	retryTimes := 5
	for {
		err := helm.DeleteRelease(namespace, releaseName, deletePvcs, async, timeoutSec)
		if err != nil {
			if strings.Contains(err.Error(), "please wait for the release latest task") && retryTimes > 0 {
				logrus.Warnf("retry to delete release %s/%s after 2 second", namespace, releaseName)
				retryTimes --
				time.Sleep(time.Second * 2)
				continue
			}
		}
		return err
	}
}

func (helm *Helm) DeleteRelease(namespace, releaseName string, deletePvcs bool, async bool, timeoutSec int64) error {
	if timeoutSec == 0 {
		timeoutSec = defaultTimeoutSec
	}

	oldReleaseTask, err := helm.validateReleaseTask(namespace, releaseName, false)
	if err != nil {
		if errorModel.IsNotFoundError(err) {
			logrus.Warnf("release task %s/%s is not found", namespace, releaseName)
			return nil
		}
		logrus.Errorf("failed to validate release task : %s", err.Error())
		return err
	}

	releaseTaskArgs := &DeleteReleaseTaskArgs{
		Namespace:   namespace,
		ReleaseName: releaseName,
		DeletePvcs:  deletePvcs,
	}

	err = helm.sendReleaseTask(namespace, releaseName, deleteReleaseTaskName, releaseTaskArgs, oldReleaseTask, timeoutSec, async)
	if err != nil {
		logrus.Errorf("async=%t, failed to send %s of %s/%s: %s", async, deleteReleaseTaskName, namespace, releaseName, err.Error())
		return err
	}
	logrus.Infof("succeed to call delete release %s/%s api", namespace, releaseName)
	return nil
}

func (helm *Helm) doDeleteRelease(namespace, releaseName string, deletePvcs bool) error {
	releaseCache, err := helm.releaseCache.GetReleaseCache(namespace, releaseName)
	if err != nil {
		if errorModel.IsNotFoundError(err) {
			logrus.Warnf("release cache %s is not found in redis", releaseName)
			return nil
		}
		logrus.Errorf("failed to get release cache %s : %s", releaseName, err.Error())
		return err
	}
	releaseInfo, err := helm.buildReleaseInfoV2(releaseCache)
	if err != nil {
		logrus.Errorf("failed to build release info : %s", err.Error())
		return err
	}

	err = helm.helm.DeleteRelease(namespace, releaseName)
	if err != nil {
		logrus.Errorf("failed to delete release %s/%s from helm : %s", namespace, releaseName, err.Error())
		return err
	}

	err = helm.releaseCache.DeleteReleaseCache(namespace, releaseName)
	if err != nil {
		logrus.Errorf("failed to delete release cache of %s : %s", releaseName, err.Error())
		return err
	}

	if deletePvcs {
		err = helm.k8sOperator.DeleteStatefulSetPvcs(releaseInfo.Status.StatefulSets)
		if err != nil {
			logrus.Errorf("failed to delete stateful set pvcs : %s", err.Error())
			return err
		}
	}

	logrus.Infof("succeed to delete release %s/%s", namespace, releaseName)
	return nil
}
