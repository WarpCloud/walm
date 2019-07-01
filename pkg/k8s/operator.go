package k8s

import (
	"WarpCloud/walm/pkg/models/k8s"
	"WarpCloud/walm/pkg/models/release"
)

type Operator interface {
	DeleteStatefulSetPvcs(statefulSets []*k8s.StatefulSet) error
	DeletePod(namespace string, name string) error

	BuildManifestObjects(namespace string, manifest string) ([]map[string]interface{}, error)
	ComputeReleaseResourcesByManifest(namespace string, manifest string) (*release.ReleaseResources, error)
}
