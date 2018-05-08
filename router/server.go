package router

import (
	"fmt"
	"net/http"
	"time"
)

type Server struct {
	ApiErrCh                chan error
	Port                    int
	TlsEnable               bool
	TlsCertFile, TlsKeyFile string
	OauthEnable             bool
	ReadTimeout             time.Duration
	WriteTimeout            time.Duration
	RunMode                 string
}

func (server *Server) StartServer() error {
	go func() {
		router := InitRouter(server.OauthEnable, server.RunMode)

		s := &http.Server{
			Addr:           fmt.Sprintf(":%d", server.Port),
			Handler:        router,
			ReadTimeout:    server.ReadTimeout,
			WriteTimeout:   server.WriteTimeout,
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
