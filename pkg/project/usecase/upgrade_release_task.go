package usecase

import (
	"github.com/sirupsen/logrus"
	"encoding/json"
	"WarpCloud/walm/pkg/models/release"
	"WarpCloud/walm/pkg/models/project"
)

const (
	upgradeReleaseTaskName = "Upgrade-Release-Task"
)

type UpgradeReleaseTaskArgs struct {
	Namespace     string
	ProjectName   string
	ReleaseParams *release.ReleaseRequestV2
}

func (projectImpl *Project) registerUpgradeReleaseTask() {
	projectImpl.task.RegisterTask(upgradeReleaseTaskName, projectImpl.UpgradeReleaseTask)
}

func (projectImpl *Project) UpgradeReleaseTask(upgradeReleaseTaskArgsStr string) error {
	upgradeReleaseTaskArgs := &UpgradeReleaseTaskArgs{}
	err := json.Unmarshal([]byte(upgradeReleaseTaskArgsStr), upgradeReleaseTaskArgs)
	if err != nil {
		logrus.Errorf("upgrade release task arg is not valid : %s", err.Error())
		return err
	}
	return projectImpl.upgradeRelease(upgradeReleaseTaskArgs.Namespace, upgradeReleaseTaskArgs.ProjectName, upgradeReleaseTaskArgs.ReleaseParams)
}

func (projectImpl *Project) upgradeRelease(namespace, projectName string, releaseParams *release.ReleaseRequestV2) (err error) {
	if releaseParams.ReleaseLabels == nil {
		releaseParams.ReleaseLabels = map[string]string{}
	}
	releaseParams.ReleaseLabels[project.ProjectNameLabelKey] = projectName

	err = projectImpl.releaseUseCase.InstallUpgradeReleaseWithRetry(namespace, releaseParams,  nil, false, 0, nil)
	if err != nil {
		logrus.Errorf("failed to upgrade release %s in project %s/%s : %s", releaseParams.Name, namespace, projectName, err.Error())
		return
	}
	return
}
