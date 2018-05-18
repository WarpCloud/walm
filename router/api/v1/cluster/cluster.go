package cluster

import (
	"walm/router/api/v1/instance"
)

type Cluster struct {
	ProdName string                 `json:"prodname" description:"product info"`
	Conf     string                 `json:"config" description:"product config"`
	Apps     []instance.Application `json:"apps" description:"list of application of the cluster"`
}
