package router

import (
	"fmt"
	"net/http"

	"walm/pkg/setting"
	"walm/router/api/v1"
)

type Server struct {
	ApiErrCh                chan error
	Wi                      v1.WalmInterface
	Port                    int
	TlsEnable               bool
	TlsCertFile, TlsKeyFile string
	OauthEnable             bool
}

var conf *setting.Config

func (server *Server) StartServer(sc *setting.Config) error {
	conf = sc
	go func() {
		router := InitRouter(server.OauthEnable)

		s := &http.Server{
			Addr:           fmt.Sprintf(":%d", server.Port),
			Handler:        router,
			ReadTimeout:    conf.ReadTimeout,
			WriteTimeout:   conf.WriteTimeout,
			MaxHeaderBytes: 1 << 20,
		}
		//walm_api.AddPrometheusHandler(restful.DefaultContainer)

		if server.TlsEnable {
			if err := s.ListenAndServeTLS(server.TlsCertFile, server.TlsKeyFile); err != nil {
				server.ApiErrCh <- err
			}
		} else {
			if err := s.ListenAndServe(); err != nil {
				server.ApiErrCh <- err
			}
		}

	}()

	return nil

}
