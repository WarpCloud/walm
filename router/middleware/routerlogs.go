package middleware

import (
	"github.com/emicklei/go-restful"
	"github.com/sirupsen/logrus"
	"time"
)

func RouteLogging(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	now := time.Now()
	chain.ProcessFilter(req, resp)
	logrus.Infof("[route-filter (logger)] OP %s URI %s COST %v RESP %d", req.Request.Method, req.Request.URL, time.Now().Sub(now), resp.StatusCode())
}
