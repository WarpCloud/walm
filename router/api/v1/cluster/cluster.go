package cluster

import (
	helm "walm/pkg/helm"
)

type Cluster struct {
	ConfigValues map[string]interface{} `json:"configvalues" description:"extra values added to the chart"`
	Apps         []helm.ReleaseRequest  `json:"apps" description:"list of application of the cluster"`
}
