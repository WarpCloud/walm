package main

import (
	"walm/models"
	"walm/pkg/helm"
	"walm/pkg/util/oauth"
	"walm/router"

	"github.com/prometheus/common/log"
	"github.com/spf13/cobra"
)

const servereDesc = `
This command enable a WALM Web server.

you can sp  the listen host and pot like :

	$ walm serv -a addr -p port

`

type ServCmd struct {
	port  int
	oauth bool
}

func newServCmd() *cobra.Command {
	inst := &walmCmd{}

	cmd := &cobra.Command{
		Use:   "serv [-a addr] [-p port]",
		Short: "enable a Walm Web Server",
		Long:  servereDesc,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return models.Init(&settings)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if inst.port == 0 {
				inst.port = settings.HTTPPort
			}
			if inst.port == 0 {
				Log.Errorf("start API server failed:%s exiting\n", err)
				return error.New()
			}
			if inst.oauth {
				if len(settings.JwtSecret) > 0 {
					oauth.SetJwtSecret(settings.JwtSecret)
				} else {
					Log.Errorln("If enable oauth ,please set JwtSecret")
					return error.New()
				}

			}
			return inst.run()
		},
		PostRun: func(_ *cobra.Command, _ []string) {
			return models.CloseDB()
		},
	}

	f := cmd.Flags()
	f.Int32VarP(&inst.port, "port", "p", 0, "address to listen on")
	f.BoolVarP(&inst.oauth, "oauth", false, "enable oauth or not")

	return cmd
}

func (sc *ServCmd) run() error {
	apiErrCh := make(chan error)

	server := &router.Server{
		ApiErrCh:    apiErrCh,
		Wi:          helm.Helm,
		Port:        sc.port,
		OauthEnable: sc.oauth,
		TlsEnable:   tlsEnable,
		TlsCertFile: tlsCertFile,
		TlsKeyFile:  tlsKeyFile,
	}

	if err := server.StartServer(&settings); err != nil {
		log.Errorf("start API server failed:%s exiting\n", err)
		return err
	} else {
		select {
		case err := <-apiErrCh:
			log.Errorf("API server error:%s exiting\n", err)
			return err
		}
	}
	return nil
}
