package helm

import (
	"walm/pkg/release"
	"k8s.io/helm/pkg/chart/loader"
	"github.com/sirupsen/logrus"
)

func (hc *HelmClient) DryRunRelease(namespace string, releaseRequest *release.ReleaseRequestV2, isSystem bool, chartFiles []*loader.BufferedFile) (string, error) {
	release, err := hc.doInstallUpgradeRelease(namespace, releaseRequest, isSystem, chartFiles, true)
	if err != nil {
		logrus.Errorf("failed to dry run install release : %s", err.Error())
		return "", err
	}
	logrus.Debugf("release manifest : %s", release.Manifest)
	return release.Manifest, nil
}