package helm

import (
	"github.com/sirupsen/logrus"
	walmerr "WarpCloud/walm/pkg/util/error"
	"WarpCloud/walm/pkg/release"
)

func (hc *HelmClient) RecoverRelease(namespace, releaseName string, isSystem bool, async bool, timeoutSec int64) error {
	if timeoutSec == 0 {
		timeoutSec = defaultTimeoutSec
	}

	oldReleaseTask, err := hc.validateReleaseTask(namespace, releaseName, false)
	if err != nil {
		if walmerr.IsNotFoundError(err) {
			logrus.Warnf("release task %s/%s is not found", namespace, releaseName)
			return err
		}
		logrus.Errorf("failed to validate release task : %s", err.Error())
		return err
	}

	releaseTaskArgs := &RecoverReleaseTaskArgs{
		Namespace:   namespace,
		ReleaseName: releaseName,
	}
	err = SendReleaseTask(hc.helmCache, namespace, releaseName, releaseTaskArgs, oldReleaseTask, timeoutSec, async)
	if err != nil {
		logrus.Errorf("async=%t, failed to send %s of %s/%s: %s", async, releaseTaskArgs.GetTaskName(), namespace, releaseName, err.Error())
		return err
	}
	logrus.Infof("succeed to call recover release %s/%s api", namespace, releaseName)
	return nil
}

func (hc *HelmClient) doRecoverRelease(namespace, releaseName string) error {
	releaseCache, err := hc.helmCache.GetReleaseCache(namespace, releaseName)
	if err != nil {
		logrus.Errorf("failed to get release cache %s : %s", releaseName, err.Error())
		return err
	}
	releaseInfo, err := hc.buildReleaseInfoV2(releaseCache)
	if err != nil {
		logrus.Errorf("failed to build release info : %s", err.Error())
		return err
	}

	if !releaseInfo.Paused {
		logrus.Warnf("release %s/%s is not paused", namespace, releaseName)
	} else {
		if releaseInfo.PauseInfo != nil {
			err = releaseInfo.PauseInfo.Recover()
			if err != nil {
				logrus.Errorf("failed to recover release resources : %s", err.Error())
				return err
			}
		} else {
			logrus.Warnf("release %s/%s has no k8s resources to recover", namespace, releaseName)
		}

		err = hc.releaseConfigHandler.AnnotateReleaseConfig(namespace, releaseName, nil, []string{release.ReleasePauseInfoKey, release.ReleasePausedKey})
		if err != nil {
			logrus.Errorf("failed to annotate release config : %s", err.Error())
			return err
		}

		logrus.Infof("succeed to recover release %s/%s", namespace, releaseName)
	}

	return nil
}
