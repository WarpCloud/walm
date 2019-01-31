package project

import (
	"github.com/sirupsen/logrus"
	"encoding/json"
	"walm/pkg/task"
	"github.com/RichardKnop/machinery/v1/tasks"
	"walm/pkg/release/manager/helm/cache"
	"walm/pkg/release"
)

const (
	upgradeReleaseTaskName = "Upgrade-Release-Task"
)

func init() {
	task.RegisterTasks(upgradeReleaseTaskName, UpgradeReleaseTask)
}

func UpgradeReleaseTask(upgradeReleaseTaskArgsStr string) error {
	upgradeReleaseTaskArgs := &UpgradeReleaseTaskArgs{}
	err := json.Unmarshal([]byte(upgradeReleaseTaskArgsStr), upgradeReleaseTaskArgs)
	if err != nil {
		logrus.Errorf("upgrade release task arg is not valid : %s", err.Error())
		return err
	}
	return upgradeReleaseTaskArgs.upgradeRelease()
}

func SendUpgradeReleaseTask(upgradeReleaseTaskArgs *UpgradeReleaseTaskArgs) (*cache.ProjectTaskSignature, error) {
	upgradeReleaseTaskArgsStr, err := json.Marshal(upgradeReleaseTaskArgs)
	if err != nil {
		logrus.Errorf("failed to marshal upgrade release task args: %s", err.Error())
		return nil, err
	}
	upgradeReleaseTaskSig := &tasks.Signature{
		Name: upgradeReleaseTaskName,
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: string(upgradeReleaseTaskArgsStr),
			},
		},
	}
	err = task.GetDefaultTaskManager().SendTask(upgradeReleaseTaskSig)
	if err != nil {
		logrus.Errorf("failed to send upgrade release task : %s", err.Error())
		return nil, err
	}
	return &cache.ProjectTaskSignature{
		Name: upgradeReleaseTaskName,
		UUID: upgradeReleaseTaskSig.UUID,
		Arg:  string(upgradeReleaseTaskArgsStr),
	}, nil
}

type UpgradeReleaseTaskArgs struct {
	Namespace     string
	ProjectName          string
	ReleaseParams *release.ReleaseRequestV2
}

func (upgradeReleaseTaskArgs *UpgradeReleaseTaskArgs) upgradeRelease() (err error) {
	if upgradeReleaseTaskArgs.ReleaseParams.ReleaseLabels == nil {
		upgradeReleaseTaskArgs.ReleaseParams.ReleaseLabels = map[string]string{}
	}
	upgradeReleaseTaskArgs.ReleaseParams.ReleaseLabels[cache.ProjectNameLabelKey] = upgradeReleaseTaskArgs.ProjectName

	err = GetDefaultProjectManager().helmClient.InstallUpgradeReleaseWithRetry(upgradeReleaseTaskArgs.Namespace, upgradeReleaseTaskArgs.ReleaseParams, false, nil, false, 0)
	if err != nil {
		logrus.Errorf("failed to upgrade release %s in project %s/%s : %s", upgradeReleaseTaskArgs.ReleaseParams.Name, upgradeReleaseTaskArgs.Namespace, upgradeReleaseTaskArgs.ProjectName)
		return
	}
	return
}
