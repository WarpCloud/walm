package middleware

import (
	"github.com/emicklei/go-restful"
	"github.com/sirupsen/logrus"
)

func BasicAuthenticate(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	u, p, _ := req.Request.BasicAuth()
	logrus.Infof("user %s password %s", u, p)
	chain.ProcessFilter(req, resp)
}
