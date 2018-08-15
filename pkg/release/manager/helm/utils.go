package helm

import (
	"walm/pkg/release"
	"k8s.io/helm/pkg/helm"
)

func BuildReleaseListOptions(option *release.ReleaseListOption) (options []helm.ReleaseListOption) {
	if option == nil {
		return
	}
	if option.Namespace != nil {
		options = append(options, helm.ReleaseListNamespace(*option.Namespace))
	}
	if option.Filter != nil {
		options = append(options, helm.ReleaseListFilter(*option.Filter))
	}
	if option.Limit != nil {
		options = append(options, helm.ReleaseListLimit(*option.Limit))
	}
	if option.Offset != nil {
		options = append(options, helm.ReleaseListOffset(*option.Offset))
	}
	if option.Order != nil {
		options = append(options, helm.ReleaseListOrder(*option.Order))
	}
	if option.Sort != nil {
		options = append(options, helm.ReleaseListSort(*option.Sort))
	}
	if len(option.Statuses) > 0 {
		options = append(options, helm.ReleaseListStatuses(option.Statuses))
	}
	return
}
