package cluster

import (
	"walm/router/api/v1/instance"
)

type Cluster struct {
	Conf string                 `json:"config" description:"product config (json format)"`
	Apps []instance.Application `json:"apps" description:"list of application of the cluster"`
}
