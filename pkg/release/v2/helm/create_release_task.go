package helm

import (
	"github.com/sirupsen/logrus"
	"encoding/json"
	"walm/pkg/task"
	"mime/multipart"
	"walm/pkg/release/v2"
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
	ReleaseRequest *v2.ReleaseRequestV2
	IsSystem       bool
	ChartArchive   multipart.File
}

func (createReleaseTaskArgs *CreateReleaseTaskArgs) Run() error {
	return GetDefaultHelmClientV2().doInstallUpgradeReleaseV2(createReleaseTaskArgs.Namespace, createReleaseTaskArgs.ReleaseRequest, createReleaseTaskArgs.IsSystem, createReleaseTaskArgs.ChartArchive)
}

func (createReleaseTaskArgs *CreateReleaseTaskArgs) GetTaskName() string {
	return createReleaseTaskName
}
