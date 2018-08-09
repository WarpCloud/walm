package middleware

import (
	"github.com/emicklei/go-restful"
	"github.com/thoas/stats"
)

var ServerStats = stats.New()

func ServerStatsFilter(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	beginning, recorder := ServerStats.Begin(response)
	chain.ProcessFilter(request, response)
	ServerStats.End(beginning, recorder)
}

func ServerStatsData(request *restful.Request, response *restful.Response) {
	response.WriteEntity(ServerStats.Data())
}
