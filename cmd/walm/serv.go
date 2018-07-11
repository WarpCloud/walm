package main

import (
	"errors"
	. "walm/pkg/util/log"
	"walm/pkg/util/oauth"
	"walm/router"

	"github.com/prometheus/common/log"
	"github.com/spf13/cobra"
)

const servDesc = `
This command enable a WALM Web server.

you can sp  the listen host and pot like :

	$ walm serv -a addr -p port

`

type ServCmd struct {
	oauth bool
}

func newServCmd() *cobra.Command {
	inst := &ServCmd{}

	cmd := &cobra.Command{
		Use:   "serv [-a addr] [-p port]",
		Short: "enable a Walm Web Server",
		Long:  servDesc,

		RunE: func(cmd *cobra.Command, args []string) error {

			if conf.Http.HTTPPort == 0 {
				Log.Errorln("start API server failed, please spec JwtSecret")
				return errors.New("none port spec")
			}
			if conf.Auth.Enable {
				if len(conf.Auth.JwtSecret) > 0 {
					oauth.SetJwtSecret(conf.Auth.JwtSecret)
				} else {
					Log.Errorln("If enable oauth ,please set JwtSecret")
					return errors.New("none JwtSecret")
				}

			}
			return inst.run()
		},
	}

	return cmd
}

func (sc *ServCmd) run() error {
	apiErrCh := make(chan error)

	server := &router.Server{
		ApiErrCh:     apiErrCh,
		Port:         conf.Http.HTTPPort,
		OauthEnable:  conf.Auth.Enable,
		TlsEnable:    conf.Secret.Tls,
		TlsCertFile:  conf.Secret.TlsCert,
		TlsKeyFile:   conf.Secret.TlsKey,
		ReadTimeout:  conf.Http.ReadTimeout,
		WriteTimeout: conf.Http.WriteTimeout,
		RunMode:      conf.Debug,
		ZipkinUrl:    conf.Trace.ZipkinUrl,
	}

	if err := server.StartServer(); err != nil {
		log.Errorf("start API server failed:%s exiting\n", err)
		return err
	} else {
		select {
		case err := <-apiErrCh:
			log.Errorf("API server error:%s exiting\n", err)
			return err
		}
	}

}
