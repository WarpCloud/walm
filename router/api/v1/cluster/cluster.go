package cluster

import (
	helm "walm/pkg/helm"
)

type Cluster struct {
	ConfigValues map[string]interface{} `json:"configvalues" description:"extra values added to the chart"`
	Apps         []helm.ReleaseRequest  `json:"apps" type:"array" ref:"helm.ReleaseRequest"  description:"list of application of the cluster"`
}

type ReleaseList struct {
	Apps []helm.ReleaseRequest `json:"apps" type:"array" ref:"helm.ReleaseRequest"  description:"list of application of the cluster"`
}
