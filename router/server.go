package router

import (
	"fmt"
	"net/http"
	"time"
	"walm/pkg/setting"
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
	RunMode                 bool
	ZipkinUrl               string
	server                  *http.Server
}

func NewServer(errch chan error) *Server {
	conf := setting.Config
	return &Server{
		ApiErrCh: errch,

		OauthEnable: conf.Auth.Enable,
		TlsEnable:   conf.Secret.Tls,
		TlsCertFile: conf.Secret.TlsCert,
		TlsKeyFile:  conf.Secret.TlsKey,

		RunMode:   conf.Debug,
		ZipkinUrl: conf.Trace.ZipkinUrl,
		server: &http.Server{
			Addr:           fmt.Sprintf(":%d", conf.Http.HTTPPort),
			ReadTimeout:    conf.Http.ReadTimeout,
			WriteTimeout:   conf.Http.WriteTimeout,
			MaxHeaderBytes: 1 << 20,
		},
	}
}

func (server *Server) StartServer() error {
	go func() {

		if !server.RunMode {
			//EndTrac will be called when close the server
			//so the init need be placed here
			middleware.InitTracer(server.ZipkinUrl, server.Port)
			defer middleware.EndTrace()
		}

		router := InitRouter(server.OauthEnable, server.RunMode)

		server.server.Handler = router
		//walm_api.AddPrometheusHandler(restful.DefaultContainer)

		if server.TlsEnable {
			if err := server.server.ListenAndServeTLS(server.TlsCertFile, server.TlsKeyFile); err != nil {
				server.ApiErrCh <- err
			}
		} else {
			if err := server.server.ListenAndServe(); err != nil {
				server.ApiErrCh <- err
			}
		}

	}()

	return nil

}
