package helm

import (
	. "walm/pkg/util/log"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/chart"
	rls "k8s.io/helm/pkg/proto/hapi/services"
)

type Interface struct {
	helmClient *helm.Client
}

var Helm *Interface

func init() {
	client := helm.NewClient(helm.Host("172.26.0.5:31221"))
	Helm = &Interface{helmClient: client}
}

func ListReleases() ([]ReleaseInfo, error) {
	res, err := Helm.helmClient.ListReleases()
	if err != nil {
		return nil, err
	}

	if len(res.GetReleases()) == 0 {
		return nil, nil
	}

	releases := covertRelease(res)
	return releases, nil
}

func GetReleaseInfo(namespace, releasename string) (ReleaseInfo, error) {

	return ReleaseInfo{}, nil
}

func InstallUpgradeRealese(releaserequest ReleaseRequest) error {
	//releaseHistory, err := Helm.helmClient.ReleaseHistory(u.release, helm.WithMaxHistory(1))
	//
	//if err == nil {
	//	if u.namespace == "" {
	//		u.namespace = defaultNamespace()
	//	}
	//	previousReleaseNamespace := releaseHistory.Releases[0].Namespace
	//	if previousReleaseNamespace != u.namespace {
	//		fmt.Fprintf(u.out,
	//			"WARNING: Namespace %q doesn't match with previous. Release will be deployed to %s\n",
	//			u.namespace, previousReleaseNamespace,
	//		)
	//	}
	//}
	return nil
}

func RollbackRealese(namespace, releasename, version string) error {

	return nil
}

func PatchUpgradeRealese(releaserequest ReleaseRequest) error {
	return nil
}


func DeleteRealese(namespace, name string) error {
	return nil
}

func ListReleasesFastPath() error {
	return nil
}

func covertRelease(helmListReleaseResponse *rls.ListReleasesResponse) []ReleaseInfo {
	var releaseInfos []ReleaseInfo
	for _, helmRelease := range helmListReleaseResponse.GetReleases() {
		release := ReleaseInfo{}
		emptyChart := chart.Chart{}

		release.Name = helmRelease.Name
		release.Namespace = helmRelease.Namespace
		release.Version = helmRelease.Version
		cvals, err := chartutil.CoalesceValues(&emptyChart, helmRelease.Config)
		if err != nil {
			Log.Errorf("parse raw values error %s\n", helmRelease.Config.Raw)
			continue
		}
		release.ConfigValues = cvals

		releaseInfos = append(releaseInfos, release)
	}

	return releaseInfos
}
