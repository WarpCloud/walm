package project

import "WarpCloud/walm/pkg/models/project"

type Cache interface {
	GetProjectTask(namespace, name string) (*project.ProjectTask, error)
	GetProjectTasks(namespace string) ([]*project.ProjectTask, error)
	CreateOrUpdateProjectTask(projectTask *project.ProjectTask) error
}
