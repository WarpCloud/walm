package downloader

import "k8s.io/helm/pkg/provenance"

type Downloader interface {
	DownloadTo(ref, version, dest string) (string, *provenance.Verification, error)
}
