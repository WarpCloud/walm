package helm

import (
	"github.com/sirupsen/logrus"
)

func (helm *Helm) RecoverRelease(namespace, releaseName string, async bool, timeoutSec int64) error {
	releaseInfo, err := helm.GetRelease(namespace, releaseName)
	if err != nil {
		logrus.Errorf("failed to get release %s/%s : %s", namespace, releaseName, err.Error())
		return err
	}

	if !releaseInfo.Paused {
		logrus.Warnf("release %s/%s is not paused", namespace, releaseName)
		return nil
	}

	releaseRequest := releaseInfo.BuildReleaseRequestV2()
	paused := false
	err = helm.InstallUpgradeRelease(namespace, releaseRequest, nil, async, timeoutSec, &paused)
	if err != nil {
		logrus.Errorf("failed to upgrade release %s/%s : %s", namespace, releaseName, err.Error())
		return err
	}
	return nil
}

