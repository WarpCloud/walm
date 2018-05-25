package router

import (
	"fmt"
	"net/http"
	"time"
	"walm/router/middleware"
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
	ZipkinUrl               string
}

func (server *Server) StartServer() error {
	go func() {
		router := InitRouter(server.OauthEnable, server.RunMode)
		if server.RunMode != "debug" {
			//EndTrac will be called when close the server
			//so the init need be placed here
			middleware.InitTracer(server.ZipkinUrl)
			defer middleware.EndTrace()
		}

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
