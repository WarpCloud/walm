package project

import (
	"WarpCloud/walm/pkg/models/project"
	"WarpCloud/walm/pkg/models/release"
)

type UseCase interface {
	ListProjects(namespace string) (*project.ProjectInfoList, error)
	GetProjectInfo(namespace, projectName string) (*project.ProjectInfo, error)
	CreateProject(namespace string, project string, projectParams *project.ProjectParams, async bool, timeoutSec int64) error
	DryRunProject(namespace, projectName string, projectParams *project.ProjectParams) ([]map[string]interface{}, error)
	ComputeResourcesByDryRunProject(namespace, projectName string, projectParams *project.ProjectParams) ([]*release.ReleaseResources, error)
	DeleteProject(namespace string, project string, async bool, timeoutSec int64, deletePvcs bool) error
	AddReleasesInProject(namespace string, projectName string, projectParams *project.ProjectParams, async bool, timeoutSec int64) error
	UpgradeReleaseInProject(namespace string, projectName string, releaseParams *release.ReleaseRequestV2, async bool, timeoutSec int64) error
	RemoveReleaseInProject(namespace, projectName, releaseName string, async bool, timeoutSec int64, deletePvcs bool) error
}
