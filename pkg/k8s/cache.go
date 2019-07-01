package k8s

import (
	"WarpCloud/walm/pkg/models/k8s"
	"WarpCloud/walm/pkg/models/release"
)

type Cache interface {
	GetResourceSet(releaseResourceMetas []release.ReleaseResourceMeta) (resourceSet *k8s.ResourceSet,err error)
	GetResource(kind k8s.ResourceKind, namespace, name string) (k8s.Resource, error)
	ListReleaseConfigs(namespace, labelSelectorStr string) ([]*k8s.ReleaseConfig, error)
	ListPersistentVolumeClaims(namespace string, labelSelectorStr string) ([]*k8s.PersistentVolumeClaim, error)
}
