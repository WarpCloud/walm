package cluster

import (
	"walm/pkg/release"
)

type Cluster struct {
	ConfigValues map[string]interface{} `json:"configvalues" description:"extra values added to the chart"`
	Apps         []release.ReleaseRequest  `json:"apps" type:"array" ref:"helm.ReleaseRequest"  description:"list of application of the cluster"`
}

type ReleaseList struct {
	Apps []release.ReleaseRequest `json:"apps" type:"array" ref:"helm.ReleaseRequest"  description:"list of application of the cluster"`
}
