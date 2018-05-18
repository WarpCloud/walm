package cluster

import (
	"walm/router/api/v1/instance"
)

type Info struct {
	Status string          `json:"name" description:"status of the cluster"`
	Infos  []instance.Info `json:"infos" description:"list of info of the releases of cluster"`
}
