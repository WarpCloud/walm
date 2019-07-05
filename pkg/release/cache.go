package release

import (
	"WarpCloud/walm/pkg/models/k8s"
	"WarpCloud/walm/pkg/models/release"
)

type Cache interface {
	GetReleaseCache(namespace, name string)(*release.ReleaseCache, error)
	GetReleaseCaches(namespace string)([]*release.ReleaseCache, error)
	GetReleaseCachesByReleaseConfigs(releaseConfigs []*k8s.ReleaseConfig) ([]*release.ReleaseCache, error)
	CreateOrUpdateReleaseCache(releaseCache *release.ReleaseCache) error
	DeleteReleaseCache(namespace string, name string) error

	GetReleaseTask(namespace, name string) (*release.ReleaseTask, error)
	GetReleaseTasks(namespace string)([]*release.ReleaseTask, error)
	GetReleaseTasksByReleaseConfigs(releaseConfigs []*k8s.ReleaseConfig) ([]*release.ReleaseTask, error)
	CreateOrUpdateReleaseTask(releaseTask *release.ReleaseTask) error
	DeleteReleaseTask(namespace string, name string) error
}
