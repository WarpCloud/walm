package main

import (
	. "walm/pkg/util/log"
	"walm/router"

	"github.com/spf13/cobra"
)

const servDesc = `
This command enable a WALM Web server.

$ walm serv 

Before to start serv ,you need to config the conf file 

The file is named conf.yaml and it's path is define by  $WALM_CONF_PATH

and the default path is /etc/walm/conf

`

type ServCmd struct {
	oauth bool
}

func newServCmd() *cobra.Command {
	inst := &ServCmd{}

	cmd := &cobra.Command{
		Use:   "serv",
		Short: "enable a Walm Web Server",
		Long:  servDesc,

		RunE: func(cmd *cobra.Command, args []string) error {
			return inst.run()
		},
	}

	return cmd
}

func (sc *ServCmd) run() error {
	apiErrCh := make(chan error)

	server := router.NewServer(apiErrCh)

	if err := server.StartServer(); err != nil {
		Log.Errorf("start API server failed:%s exiting\n", err)
		return err
	} else {
		select {
		case err := <-apiErrCh:
			Log.Errorf("API server error:%s exiting\n", err)
			return err
		}
	}

}
