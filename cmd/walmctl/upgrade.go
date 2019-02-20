package main

import (
	"io"
	"github.com/spf13/cobra"
	"github.com/pkg/errors"
	"walm/cmd/walmctl/walmctlclient"
	"fmt"
)

const upgradeDesc = `
This command upgrade an existing release 
`

type upgradeCmd struct {
	out    io.Writer
	filename string

}

func newUpgradeCmd(out io.Writer) *cobra.Command {
	uc := upgradeCmd{out:out}

	cmd := &cobra.Command{
		Use: "upgrade",
		Short: "upgrade an existing release",
		Long: upgradeDesc,
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(uc.filename) == 0 {
				return errors.New("please add flag -f/--file after command create, eg: create -f xxx.yaml")
			}

			return uc.run()
		},
	}
	cmd.Flags().StringVarP(&uc.filename, "file", "f", "", "resource file")

	return cmd
}


func (c *upgradeCmd) run() error {


	// Todo:// upgrade specific field
	resp, err := walmctlclient.CreateNewClient(walmserver).UpgradeRelease(namespace, c.filename)
	if err != nil {
		fmt.Printf(err.Error())
	}

	fmt.Println(resp)
	return nil
}