package release

import (
	"WarpCloud/walm/pkg/models/release"
	"WarpCloud/walm/pkg/models/common"
)

type UseCase interface {
	GetRelease(namespace, name string) (releaseV2 *release.ReleaseInfoV2, err error)
	ListReleases(namespace string) ([]*release.ReleaseInfoV2, error)
	ListReleasesByLabels(namespace string, labelSelectorStr string) ([]*release.ReleaseInfoV2, error)
	DryRunRelease(namespace string, releaseRequest *release.ReleaseRequestV2, chartFiles []*common.BufferedFile) ([]map[string]interface{}, error)
	ComputeResourcesByDryRunRelease(namespace string, releaseRequest *release.ReleaseRequestV2, chartFiles []*common.BufferedFile) (*release.ReleaseResources, error)
	DeleteReleaseWithRetry(namespace, releaseName string, deletePvcs bool, async bool, timeoutSec int64) error
	DeleteRelease(namespace, releaseName string, deletePvcs bool, async bool, timeoutSec int64) error
	// paused :
	// 1. nil: maintain pause state
	// 2. true: make release paused
	// 3. false: make release recovered
	InstallUpgradeReleaseWithRetry(namespace string, releaseRequest *release.ReleaseRequestV2, chartFiles []*common.BufferedFile, async bool, timeoutSec int64, paused *bool) error
	InstallUpgradeRelease(namespace string, releaseRequest *release.ReleaseRequestV2, chartFiles []*common.BufferedFile, async bool, timeoutSec int64, paused *bool) error
	ReloadRelease(namespace, name string) error
	RestartRelease(namespace, releaseName string) error
	RecoverRelease(namespace, releaseName string, async bool, timeoutSec int64) error
	PauseRelease(namespace, releaseName string, async bool, timeoutSec int64) error
}
