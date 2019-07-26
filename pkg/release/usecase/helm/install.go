package helm

import (
	"WarpCloud/walm/pkg/models/release"
	"github.com/sirupsen/logrus"
	"fmt"
	"time"
	"strings"
	"WarpCloud/walm/pkg/models/common"
	errorModel "WarpCloud/walm/pkg/models/error"
)

const (
	defaultTimeoutSec int64 = 60 * 5
)

func (helm *Helm) InstallUpgradeReleaseWithRetry(namespace string, releaseRequest *release.ReleaseRequestV2, chartFiles []*common.BufferedFile, async bool, timeoutSec int64, paused *bool) error {
	retryTimes := 5
	for {
		err := helm.InstallUpgradeRelease(namespace, releaseRequest, chartFiles, async, timeoutSec, paused)
		if err != nil {
			if strings.Contains(err.Error(), waitReleaseTaskMsgPrefix) && retryTimes > 0 {
				logrus.Warnf("retry to install or upgrade release %s/%s after 2 second", namespace, releaseRequest.Name)
				retryTimes --
				time.Sleep(time.Second * 2)
				continue
			}
		}
		return err
	}
}

func (helm *Helm) InstallUpgradeRelease(namespace string, releaseRequest *release.ReleaseRequestV2, chartFiles []*common.BufferedFile, async bool, timeoutSec int64, paused *bool) error {
	err := validateParams(releaseRequest, chartFiles)
	if err != nil {
		logrus.Errorf("failed to validate params : %s", err.Error())
		return err
	}

	if timeoutSec == 0 {
		timeoutSec = defaultTimeoutSec
	}

	oldReleaseTask, err := helm.validateReleaseTask(namespace, releaseRequest.Name, true)
	if err != nil {
		return err
	}

	releaseTaskArgs := &CreateReleaseTaskArgs{
		Namespace:      namespace,
		ReleaseRequest: releaseRequest,
		ChartFiles:     chartFiles,
		Paused:         paused,
	}

	err = helm.sendReleaseTask(namespace, releaseRequest.Name, createReleaseTaskName, releaseTaskArgs, oldReleaseTask, timeoutSec, async)
	if err != nil {
		logrus.Errorf("async=%t, failed to send %s of %s/%s: %s", async, createReleaseTaskName, namespace, releaseRequest.Name, err.Error())
		return err
	}
	logrus.Infof("succeed to call create or update release %s/%s api", namespace, releaseRequest.Name)
	return nil
}

func validateParams(releaseRequest *release.ReleaseRequestV2, chartFiles []*common.BufferedFile) error {
	if releaseRequest.Name == "" {
		return fmt.Errorf("release name can not be empty")
	}

	if releaseRequest.ChartName == "" && releaseRequest.ChartImage == "" && len(chartFiles) == 0 {
		return fmt.Errorf("at lease one of chart name or chart image or chart files should be supported")
	}

	return nil
}

func (helm *Helm) doInstallUpgradeRelease(namespace string, releaseRequest *release.ReleaseRequestV2, chartFiles []*common.BufferedFile, dryRun bool, paused *bool) (*release.ReleaseCache, error) {
	update := true
	oldReleaseCache, err := helm.releaseCache.GetReleaseCache(namespace, releaseRequest.Name)
	if err != nil {
		if errorModel.IsNotFoundError(err) {
			update = false
		} else {
			logrus.Errorf("failed to get release cache of %s/%s : %s", namespace, releaseRequest.Name, err.Error())
			return nil, err
		}
	}

	var oldReleaseInfo *release.ReleaseInfoV2
	if oldReleaseCache != nil {
		oldReleaseInfo, err = helm.buildReleaseInfoV2(oldReleaseCache)
		if err != nil {
			logrus.Errorf("failed to build release info of %s/%s: %s", namespace, releaseRequest.Name, err.Error())
			return nil, err
		}
	}

	preProcessRequest(releaseRequest)

	releaseCache, err := helm.helm.InstallOrCreateRelease(namespace, releaseRequest, chartFiles, dryRun, update, oldReleaseInfo, paused)
	if err != nil {
		logrus.Errorf("failed to install or create release %s/%s : %s", namespace, releaseRequest.Name, err.Error())
		return nil, err
	}
	if !dryRun {
		err = helm.releaseCache.CreateOrUpdateReleaseCache(releaseCache)
		if err != nil {
			logrus.Errorf("failed to create of update release cache of %s/%s : %s", namespace, releaseRequest.Name, err.Error())
			return nil, err
		}
		logrus.Infof("succeed to create or update release %s/%s", namespace, releaseRequest.Name)
	} else {
		logrus.Infof("succeed to dry run create or update release %s/%s", namespace, releaseRequest.Name)
	}

	return releaseCache, nil
}

func preProcessRequest(releaseRequest *release.ReleaseRequestV2) {
	if releaseRequest.ConfigValues == nil {
		releaseRequest.ConfigValues = map[string]interface{}{}
	}
	if releaseRequest.Dependencies == nil {
		releaseRequest.Dependencies = map[string]string{}
	}
	if releaseRequest.ReleaseLabels == nil {
		releaseRequest.ReleaseLabels = map[string]string{}
	}
}
