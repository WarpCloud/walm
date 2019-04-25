package helm

import (
	"github.com/sirupsen/logrus"
	"walm/pkg/release/manager/helm/cache"
	"walm/pkg/task"
	"time"
	walmerr "walm/pkg/util/error"
	"walm/pkg/k8s/adaptor"
	"transwarp/release-config/pkg/apis/transwarp/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"walm/pkg/release"
	"encoding/json"
)

func (hc *HelmClient) PauseRelease(namespace, releaseName string, isSystem bool, async bool, timeoutSec int64) error {
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

	releaseTaskArgs := &PauseReleaseTaskArgs{
		Namespace:   namespace,
		ReleaseName: releaseName,
	}
	taskSig, err := SendReleaseTask(releaseTaskArgs)
	if err != nil {
		logrus.Errorf("failed to send %s : %s", releaseTaskArgs.GetTaskName(), err.Error())
		return err
	}
	taskSig.TimeoutSec = timeoutSec

	releaseTask := &cache.ReleaseTask{
		Namespace:            namespace,
		Name:                 releaseName,
		LatestReleaseTaskSig: taskSig,
	}

	err = hc.helmCache.CreateOrUpdateReleaseTask(releaseTask)
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
			logrus.Errorf("failed to pause release  %s/%s: %s", namespace, releaseName, err.Error())
			return err
		}
	}
	logrus.Infof("succeed to call pause release %s/%s api", namespace, releaseName)
	return nil
}

func (hc *HelmClient) doPauseRelease(namespace, releaseName string) error {
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

	if releaseInfo.Paused {
		logrus.Warnf("release %s/%s has already been paused", namespace, releaseName)
	} else {
		if releaseInfo.Status != nil {
			err = releaseInfo.Status.Pause()
			if err != nil {
				logrus.Errorf("failed to pause release resources : %s", err.Error())
				return err
			}
		} else {
			logrus.Warnf("release %s/%s has no k8s resources", namespace, releaseName)
		}

		releasePauseInfoAnnos, err := buildReleasePauseInfoAnnos(releaseInfo)
		if err != nil {
			logrus.Errorf("failed to build release pause info annotations : %s", err.Error())
			return err
		}

		_, err = hc.releaseConfigHandler.GetReleaseConfig(namespace, releaseName)
		if err != nil {
			if adaptor.IsNotFoundErr(err) {
				// maybe this release is not created by WALM
				releaseConfig := buildNewReleaseConfig(namespace, releaseName, releasePauseInfoAnnos, releaseInfo)
				_, err := hc.releaseConfigHandler.CreateReleaseConfig(namespace, releaseConfig)
				if err != nil {
					logrus.Errorf("failed to create release config : %s", err.Error())
					return err
				}
			} else {
				logrus.Errorf("failed to get release config : %s", err.Error())
				return err
			}
		} else {
			err = hc.releaseConfigHandler.AnnotateReleaseConfig(namespace, releaseName, releasePauseInfoAnnos, nil)
			if err != nil {
				logrus.Errorf("failed to annotate release config : %s", err.Error())
				return err
			}
		}

		logrus.Infof("succeed to pause release %s/%s", namespace, releaseName)
	}

	return nil
}

func buildNewReleaseConfig(namespace string, releaseName string, releasePauseInfoAnnos map[string]string, releaseInfo *release.ReleaseInfoV2) *v1beta1.ReleaseConfig {
	releaseConfig := &v1beta1.ReleaseConfig{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   namespace,
			Name:        releaseName,
			Annotations: releasePauseInfoAnnos,
		},
		Spec: v1beta1.ReleaseConfigSpec{
			ConfigValues:    releaseInfo.ConfigValues,
			ChartName:       releaseInfo.ChartName,
			ChartAppVersion: releaseInfo.ChartAppVersion,
			ChartVersion:    releaseInfo.ChartVersion,
		},
	}
	return releaseConfig
}

func buildReleasePauseInfoAnnos(releaseInfo *release.ReleaseInfoV2) (releasePauseInfoAnnos map[string]string, err error){
	releasePauseInfoAnnos = map[string]string{}
	releasePauseInfoAnnos[release.ReleasePausedKey] = release.ReleasePausedValue
	releasePauseInfo := buildReleasePauseInfoByReleaseStatus(releaseInfo.Status)
	if releasePauseInfo != nil {
		releasePauseInfoBytes, err := json.Marshal(releasePauseInfo)
		if err != nil {
			logrus.Errorf("failed to marshal release pause info : %s", err.Error())
			return nil, err
		}
		releasePauseInfoAnnos[release.ReleasePauseInfoKey] = string(releasePauseInfoBytes)
	}
	return
}

func buildReleasePauseInfoByReleaseStatus(releaseStatus *adaptor.WalmResourceSet) (releasePauseInfo *release.ReleasePauseInfo) {
	if releaseStatus != nil {
		releasePauseInfo = &release.ReleasePauseInfo{}
		for _, deployment := range releaseStatus.Deployments {
			releasePauseInfo.Deployments = append(releasePauseInfo.Deployments, release.PauseInfo{
				Namespace:        deployment.Namespace,
				Name:             deployment.Name,
				PreviousReplicas: deployment.ExpectedReplicas,
			})
		}
		for _, statefulSet := range releaseStatus.StatefulSets {
			releasePauseInfo.StatefulSets = append(releasePauseInfo.StatefulSets, release.PauseInfo{
				Namespace:        statefulSet.Namespace,
				Name:             statefulSet.Name,
				PreviousReplicas: statefulSet.ExpectedReplicas,
			})
		}
	}
	return
}
