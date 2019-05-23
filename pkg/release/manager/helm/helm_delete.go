package helm

import (
	"github.com/sirupsen/logrus"
	"strings"
	"WarpCloud/walm/pkg/k8s/adaptor"
	"WarpCloud/walm/pkg/k8s/handler"
	walmerr "WarpCloud/walm/pkg/util/error"
	"k8s.io/helm/pkg/helm"
	"time"
)

func (hc *HelmClient) DeleteReleaseWithRetry(namespace, releaseName string, isSystem bool, deletePvcs bool, async bool, timeoutSec int64) error {
	retryTimes := 5
	for {
		err := hc.DeleteRelease(namespace, releaseName, isSystem, deletePvcs, async, timeoutSec)
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

func (hc *HelmClient) DeleteRelease(namespace, releaseName string, isSystem bool, deletePvcs bool, async bool, timeoutSec int64) error {
	if timeoutSec == 0 {
		timeoutSec = defaultTimeoutSec
	}

	oldReleaseTask, err := hc.validateReleaseTask(namespace, releaseName, false)
	if err != nil {
		if walmerr.IsNotFoundError(err) {
			logrus.Warnf("release task %s/%s is not found", namespace, releaseName)
			return nil
		}
		logrus.Errorf("failed to validate release task : %s", err.Error())
		return err
	}

	releaseTaskArgs := &DeleteReleaseTaskArgs{
		Namespace:   namespace,
		ReleaseName: releaseName,
		IsSystem:    isSystem,
		DeletePvcs:  deletePvcs,
	}

	err = SendReleaseTask(hc.helmCache, namespace, releaseName, releaseTaskArgs, oldReleaseTask, timeoutSec, async)
	if err != nil {
		logrus.Errorf("async=%t, failed to send %s of %s/%s: %s", async, releaseTaskArgs.GetTaskName(), namespace, releaseName, err.Error())
		return err
	}
	logrus.Infof("succeed to call delete release %s/%s api", namespace, releaseName)
	return nil
}

func (hc *HelmClient) doDeleteRelease(namespace, releaseName string, isSystem bool, deletePvcs bool) error {
	currentHelmClient, err := hc.getCurrentHelmClient(namespace)
	if err != nil {
		logrus.Errorf("failed to get current helm client : %s", err.Error())
		return err
	}

	releaseCache, err := hc.helmCache.GetReleaseCache(namespace, releaseName)
	if err != nil {
		if walmerr.IsNotFoundError(err) {
			logrus.Warnf("release cache %s is not found in redis", releaseName)
			return nil
		}
		logrus.Errorf("failed to get release cache %s : %s", releaseName, err.Error())
		return err
	}
	releaseInfo, err := hc.buildReleaseInfoV2(releaseCache)
	if err != nil {
		logrus.Errorf("failed to build release info : %s", err.Error())
		return err
	}

	opts := []helm.UninstallOption{
		helm.UninstallPurge(true),
	}
	res, err := currentHelmClient.UninstallRelease(
		releaseName, opts...,
	)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logrus.Warnf("release %s is not found from helm", releaseName)
		} else {
			logrus.Errorf("failed to delete release : %s", err.Error())
			return err
		}
	}
	if res != nil && res.Info != "" {
		logrus.Println(res.Info)
	}

	err = hc.helmCache.DeleteReleaseCache(namespace, releaseName)
	if err != nil {
		logrus.Errorf("failed to delete release cache of %s : %s", releaseName, err.Error())
		return err
	}

	if deletePvcs {
		statefulSets := []adaptor.WalmStatefulSet{}
		if len(releaseInfo.Status.StatefulSets) > 0 {
			statefulSets = append(statefulSets, releaseInfo.Status.StatefulSets...)
		}

		for _, statefulSet := range statefulSets {
			if statefulSet.Selector != nil && (len(statefulSet.Selector.MatchLabels) > 0 || len(statefulSet.Selector.MatchExpressions) > 0) {
				pvcs, err := handler.GetDefaultHandlerSet().GetPersistentVolumeClaimHandler().ListPersistentVolumeClaims(statefulSet.Namespace, statefulSet.Selector)
				if err != nil {
					logrus.Errorf("failed to list pvcs ralated to stateful set %s/%s : %s", statefulSet.Namespace, statefulSet.Name, err.Error())
					return err
				}

				for _, pvc := range pvcs {
					err = handler.GetDefaultHandlerSet().GetPersistentVolumeClaimHandler().DeletePersistentVolumeClaim(pvc.Namespace, pvc.Name)
					if err != nil {
						if adaptor.IsNotFoundErr(err) {
							logrus.Warnf("pvc %s/%s related to stateful set %s/%s is not found", pvc.Namespace, pvc.Name, statefulSet.Namespace, statefulSet.Name)
							continue
						}
						logrus.Errorf("failed to delete pvc %s/%s related to stateful set %s/%s : %s", pvc.Namespace, pvc.Name, statefulSet.Namespace, statefulSet.Name, err.Error())
						return err
					}
					logrus.Infof("succeed to delete pvc %s/%s related to stateful set %s/%s", pvc.Namespace, pvc.Name, statefulSet.Namespace, statefulSet.Name)
				}
			}
		}
	}

	logrus.Infof("succeed to delete release %s/%s", namespace, releaseName)
	return nil
}
