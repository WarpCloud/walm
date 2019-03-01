package main

import (
	"io"
	"github.com/spf13/cobra"
	"errors"
	"github.com/go-resty/resty"
	"walm/cmd/walmctl/walmctlclient"
	"fmt"
	"encoding/json"
	"walm/pkg/project"
	"walm/pkg/release"
	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
	"github.com/bitly/go-simplejson"
)

const listDesc = `
This command shows walm releases,projects or releases in a project under namespace.
list release support only currently.
`

type listCmd struct {
	projectName string
	sourceType 	string
	colWidth    uint
	output  	string
	short 		bool
	out    		io.Writer
}

type listRelease struct {
	Name string
	Ready bool
	ChartName string
	ChartVersion string
}


func newListCmd(out io.Writer) *cobra.Command {
	lc := listCmd{out:out}

	cmd := &cobra.Command{
		Use: "list",
		Short: "show release/project under specific namespace",
		Long: listDesc,
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) != 1 {
				return errors.New("arguments release/project required after command list")
			}
			lc.sourceType = args[0]
			if lc.sourceType != "release" && lc.sourceType != "project" {
				return errors.New("arguments error, release/project accept only")
			}

			return lc.run()
		},
	}

	cmd.Flags().StringVar(&lc.output, "output", "", "output the specified format (json or yaml)")
	cmd.Flags().StringVarP(&lc.projectName, "project", "p", "", "operate resources of the project")

	return cmd
}

// Todo: currently support list release only
func (c *listCmd) run() error {

	var resp *resty.Response
	var err error

	client := walmctlclient.CreateNewClient(walmserver)
	projectInfo := project.ProjectInfo{}
	var releases []*release.ReleaseInfoV2

	if c.sourceType == "project" {
		resp, err = client.ListProject(namespace)
		if err != nil {
			return err
		}
	} else {
		if c.projectName == "" {
			resp, err = client.ListRelease(namespace)
			respJson, _ := simplejson.NewJson(resp.Body())
			respByte, _ := respJson.Get("items").MarshalJSON()
			err = json.Unmarshal(respByte, &releases)

		} else {
			resp, err = client.GetSource(namespace, c.projectName, "project")
			if err != nil {
				return err
			}
			err = json.Unmarshal(resp.Body(), &projectInfo)
			if err != nil {
				return err
			}

			releases = projectInfo.Releases
		}
	}


	result := c.getListResult(releases)
	output, err := formatResult(c.output, result)

	fmt.Fprintln(c.out, output)

	//Todo: list release -p projectName
	//Todo: response optimization, To be like kubectl get pod, kubectl get instance

	return nil
}


func (c *listCmd) getListResult(releases []*release.ReleaseInfoV2) []listRelease {

	var listReleases []listRelease

	for _, release := range releases {

		lr := listRelease{
			Name: release.Name,
			Ready: release.Ready,
			ChartName: release.ChartName,
			ChartVersion: release.ChartVersion,
		}

		listReleases = append(listReleases, lr)
	}

	return listReleases

}

func formatResult(format string, result []listRelease) (string, error) {

	var output string
	var err error
	var finalResult interface{}

	finalResult = result

	switch format {
	case "":
		output = formatText(result)

	case "json":
		o, e := json.Marshal(finalResult)
		if e != nil {
			err = fmt.Errorf("Failed to Marshal JSON output: %s", e)
		} else {
			output = string(o)
		}
	case "yaml":
		o, e := yaml.Marshal(finalResult)
		if e != nil {
			err = fmt.Errorf("Failed to Marshal YAML output: %s", e)
		} else {
			output = string(o)
		}
	default:
		err = fmt.Errorf("Unknown output format \"%s\"", format)
	}

	return output, err

}


func formatText(result []listRelease) string {

	table := uitable.New()
	table.MaxColWidth = 60
	table.AddRow("NAME", "Ready", "ChartName", "ChartVersion")
	for _, release := range result {
		table.AddRow(release.Name, release.Ready, release.ChartName, release.ChartVersion)
	}

	return fmt.Sprintf("%s", table.String())
}
