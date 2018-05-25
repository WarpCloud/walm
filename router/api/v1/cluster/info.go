package cluster

import (
	"walm/router/api/v1/instance"
)

type Info struct {
	Name      string          `json:"name" description:"Name of the cluster"`
	Namespace string          `json:"name" description:"Namespace of the cluster"`
	Infos     []instance.Info `json:"infos" description:"list of info of the releases of cluster"`
}
