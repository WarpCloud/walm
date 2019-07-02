package main

import (
	"io"
	"github.com/spf13/cobra"
	"github.com/pkg/errors"
	"WarpCloud/walm/cmd/walmctl/util/walmctlclient"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/go-resty/resty"
)

const getDesc = `
Get a walm release or project detail info.
Options:
use --output/-o to print with json/yaml format.
`

type getCmd struct {
	sourceType  string
	sourceName  string
	output 		string
	out    		io.Writer
}


func newGetCmd(out io.Writer) *cobra.Command {
	gc := getCmd{out:out}

	cmd := &cobra.Command{
		Use: "get release/project releaseName/projectName",
		DisableFlagsInUseLine: true,
		Short: "get a release/project info",
		Long: getDesc,
		RunE: func(cmd *cobra.Command, args []string) error {

			if walmserver == "" {
				return errServerRequired
			}
			if namespace == "" {
				return errNamespaceRequired
			}

			if len(args) != 2 {
				return errors.New("arguments error, get release/project releaseName/projectName")
			}

			gc.sourceType = args[0]
			gc.sourceName = args[1]
			return gc.run()
		},
	}

	cmd.Flags().StringVarP(&gc.output, "output", "o", "json", "-o, --output='': Output format for detail description. Support: json, yaml")
	return cmd
}

func (gc *getCmd) run() error {

	var resp *resty.Response
	var err error

	err = checkResourceType(gc.sourceType)
	if err != nil {
		return err
	}
	if gc.sourceType == "release" {
		resp, err = walmctlclient.CreateNewClient(walmserver).GetRelease(namespace, gc.sourceName)
	} else {
		resp, err = walmctlclient.CreateNewClient(walmserver).GetProject(namespace, gc.sourceName)
	}

	if err != nil {
		return err
	}

	if gc.output == "yaml" {
		respByte, err := yaml.JSONToYAML(resp.Body())
		if err != nil {
			return errors.New(err.Error())
		}
		fmt.Printf(string(respByte))

	} else if gc.output == "json" {
		fmt.Println(resp)
	} else {
		return errors.Errorf("output format %s not recognized, only support yaml, json", gc.output)
	}

	return nil
}