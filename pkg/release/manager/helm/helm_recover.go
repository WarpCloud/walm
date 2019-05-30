package helm

import (
	"github.com/sirupsen/logrus"
	"k8s.io/helm/pkg/walm"
	"k8s.io/helm/pkg/walm/plugins"
)

func (hc *HelmClient) RecoverRelease(namespace, releaseName string, isSystem bool, async bool, timeoutSec int64) error {
	releaseInfo, err := hc.GetRelease(namespace, releaseName)
	if err != nil {
		logrus.Errorf("failed to get release %s/%s : %s", namespace, releaseName, err.Error())
		return err
	}

	if !releaseInfo.Paused {
		logrus.Warnf("release %s/%s is not paused", namespace, releaseName)
		return nil
	}

	releaseRequest := releaseInfo.BuildReleaseRequestV2()
	releaseRequest.Plugins, err = mergeWalmPlugins([]*walm.WalmPlugin{
		{
			Name: plugins.PauseReleasePluginName,
			Version: "1.0",
			Disable: true,
		},
	}, releaseRequest.Plugins)
	if err != nil {
		logrus.Errorf("failed to merge walm plugins : %s", err.Error())
		return err
	}

	err = hc.InstallUpgradeRelease(namespace, releaseRequest, isSystem, nil, async, timeoutSec)
	if err != nil {
		logrus.Errorf("failed to upgrade release %s/%s : %s", namespace, releaseName, err.Error())
		return err
	}
	return nil
}

