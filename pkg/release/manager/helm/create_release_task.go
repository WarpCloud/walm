package helm

import (
	"github.com/sirupsen/logrus"
	"encoding/json"
	"WarpCloud/walm/pkg/task"
	"WarpCloud/walm/pkg/release"
	"k8s.io/helm/pkg/chart/loader"
)

const (
	createReleaseTaskName = "Create-Release-Task"
)

func init() {
	task.RegisterTasks(createReleaseTaskName, createReleaseTask)
}

func createReleaseTask(releaseTaskArgsStr string) error {
	releaseTaskArgs := &CreateReleaseTaskArgs{}
	err := json.Unmarshal([]byte(releaseTaskArgsStr), releaseTaskArgs)
	if err != nil {
		logrus.Errorf("%s args is not valid : %s", releaseTaskArgs.GetTaskName(), err.Error())
		return err
	}
	return releaseTaskArgs.Run()
}

type CreateReleaseTaskArgs struct {
	Namespace      string
	ReleaseRequest *release.ReleaseRequestV2
	IsSystem       bool
	ChartFiles     []*loader.BufferedFile
}

func (createReleaseTaskArgs *CreateReleaseTaskArgs) Run() (err error) {
	_, err = GetDefaultHelmClient().doInstallUpgradeRelease(createReleaseTaskArgs.Namespace,
		createReleaseTaskArgs.ReleaseRequest, createReleaseTaskArgs.IsSystem, createReleaseTaskArgs.ChartFiles, false)
	return
}

func (createReleaseTaskArgs *CreateReleaseTaskArgs) GetTaskName() string {
	return createReleaseTaskName
}
