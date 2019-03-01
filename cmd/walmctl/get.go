package main

import (
	"io"
	"github.com/spf13/cobra"
	"github.com/pkg/errors"
	"walm/cmd/walmctl/walmctlclient"
	"fmt"
	"github.com/ghodss/yaml"
)

const getDesc = `
This command get a walm release or project info under specific namespace
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
		Short: "get a release/project info",
		Long: getDesc,
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) != 2 {
				return errors.New("arguments error, get release [releaseName] or get project [projectName]")
			}
			gc.sourceType = args[0]
			if gc.sourceType != "release" && gc.sourceType != "project" {
				return errors.New("get [args]: first arg must one of: release|project")
			}

			if gc.output != "yaml" && gc.output != "json" && gc.output != "" {
				return errors.New("flag --output/-o needs an argument, yaml/json")
			}
			gc.sourceName = args[1]
			return gc.run()
		},
	}

	cmd.Flags().StringVarP(&gc.output, "output", "o", "", "-o, --output='': Output format for detail description. One of: json|yaml")
	cmd.MarkFlagRequired("output")
	return cmd
}


func (c *getCmd) run() error {

	resp, err := walmctlclient.CreateNewClient(walmserver).GetSource(namespace, c.sourceName, c.sourceType)

	if err != nil {
		return err
	}

	if c.output == "yaml" {
		respByte, err := yaml.JSONToYAML(resp.Body())
		if err != nil {
			return errors.New(err.Error())
		}
		fmt.Printf(string(respByte))

	} else if c.output == "json" {
		fmt.Println(resp)
	} else {
		// Todo: optimization in processing without flag --output|-o, consider add in the future
	}

	return nil
}