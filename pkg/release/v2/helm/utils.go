package helm

import (
	"reflect"
	"github.com/sirupsen/logrus"
	"walm/pkg/release/v2"
	"walm/pkg/release"
)

func ConfigValuesDiff(configValue1 map[string]interface{}, configValue2 map[string]interface{}) bool {
	if len(configValue1) == 0 && len(configValue2) == 0 {
		return false
	}
	return !reflect.DeepEqual(configValue1, configValue2)
}

func MigrateV1Releases(namespace string) error {
	releases, err := GetDefaultHelmClientV2().ListReleasesV2(namespace, "")
	if err != nil {
		logrus.Errorf("failed to list releases : %s", err.Error())
		return err
	}

	for _, releaseInfo := range releases {
		if len(releaseInfo.Status.Instances) > 0 {
			releaseRequest := &v2.ReleaseRequestV2{
				ReleaseRequest: release.ReleaseRequest{
					Name: releaseInfo.Name,
					Dependencies: releaseInfo.Dependencies,
					ChartName: releaseInfo.ChartName,
					ChartVersion: releaseInfo.ChartVersion,
				},
			}
			err = GetDefaultHelmClientV2().InstallUpgradeReleaseV2(namespace, releaseRequest, false, nil)
			if err != nil {
				logrus.Errorf("failed to migrate release %s : %s", releaseInfo.Name, err.Error())
				return err
			}
		}
	}
	logrus.Infof("succeed to migrate v1 releases in %s", namespace)
	return nil
}