package main

import (
	"io"
	"github.com/spf13/cobra"
	"github.com/go-resty/resty"
	"WarpCloud/walm/cmd/walmctl/util/walmctlclient"
	"fmt"
	"encoding/json"
	"WarpCloud/walm/pkg/project"
	"WarpCloud/walm/pkg/release"
	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
	"github.com/bitly/go-simplejson"
	"github.com/pkg/errors"
)

const listDesc = `
This command shows walm releases,projects or releases in a project under namespace.
Examples:
  # List all releases in ps output format.
  walmctl list release
  # List all projects in ps output format.
  walmctl list project
  # List all releases of a specific project in ps output format
  walmctl list release -p projectName
`

type listCmd struct {
	labelSelector string
	projectName   string
	sourceType    string
	colWidth      uint
	output        string
	short         bool
	out           io.Writer
}

type listRelease struct {
	Name         string
	Ready        bool
	ChartName    string
	ChartVersion string
}

type listProject struct {
	Name      string
	Ready     bool
	CreatedAt string
	Message   string
}

func newListCmd(out io.Writer) *cobra.Command {
	lc := listCmd{out: out}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "show release/project under specific namespace",
		Long:  listDesc,
		RunE: func(cmd *cobra.Command, args []string) error {

			if walmserver == "" {
				return errServerRequired
			}
			if namespace == "" {
				return errNamespaceRequired
			}
			if len(args) != 1 {
				return errors.New("arguments error, list release/project")
			}
			err := checkResourceType(args[0])
			if err != nil {
				return err
			}
			lc.sourceType = args[0]
			return lc.run()
		},
	}

	cmd.Flags().StringVarP(&lc.output, "output", "o", "", "output the specified format (json or yaml)")
	cmd.Flags().StringVarP(&lc.projectName, "project", "p", "", "operate resources of the project")
	cmd.Flags().StringVar(&lc.labelSelector, "labelSelector", "", "match values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2")
	return cmd
}

func (lc *listCmd) run() error {

	var err error
	var output string
	var resp *resty.Response
	var projectInfo project.ProjectInfo
	var releases []*release.ReleaseInfoV2
	var projects []*project.ProjectInfo

	err = checkResourceType(lc.sourceType)
	if err != nil {
		return err
	}

	client := walmctlclient.CreateNewClient(walmserver)

	if lc.sourceType == "project" {
		resp, err = client.ListProject(namespace)
		respJson, _ := simplejson.NewJson(resp.Body())
		respByte, _ := respJson.Get("items").MarshalJSON()
		err = json.Unmarshal(respByte, &projects)
		if err != nil {
			return err
		}

		result := lc.getProjectResult(projects)
		output, err = formatProjectResult(lc.output, result)
		if err != nil {
			return err
		}

	} else {
		if lc.projectName == "" {
			resp, err = client.ListRelease(namespace, lc.labelSelector)
			respJson, _ := simplejson.NewJson(resp.Body())
			respByte, _ := respJson.Get("items").MarshalJSON()
			err = json.Unmarshal(respByte, &releases)

		} else {
			resp, err = client.GetProject(namespace, lc.projectName)
			if err != nil {
				return err
			}
			err = json.Unmarshal(resp.Body(), &projectInfo)
			if err != nil {
				return err
			}

			releases = projectInfo.Releases
		}
		result := lc.getListResult(releases)
		output, err = formatReleaseResult(lc.output, result)
		if err != nil {
			return err
		}
	}

	fmt.Fprintln(lc.out, output)
	return nil
}

func (lc *listCmd) getListResult(releases []*release.ReleaseInfoV2) []listRelease {

	var listReleases []listRelease

	for _, release := range releases {

		lr := listRelease{
			Name:         release.Name,
			Ready:        release.Ready,
			ChartName:    release.ChartName,
			ChartVersion: release.ChartVersion,
		}

		listReleases = append(listReleases, lr)
	}

	return listReleases

}

func (lc *listCmd) getProjectResult(projects []*project.ProjectInfo) []listProject {
	var listProjects []listProject

	for _, project := range projects {
		lp := listProject{
			Name:      project.Name,
			Ready:     project.Ready,
			CreatedAt: project.LatestTaskState.CreatedAt.Format("2006-01-02 15:04:05"),
			Message:   project.Message,
		}

		listProjects = append(listProjects, lp)
	}
	return listProjects
}

func formatReleaseResult(format string, result []listRelease) (string, error) {

	var err error
	var output string
	var finalResult interface{}

	finalResult = result

	switch format {
	case "":
		output = formatReleaseText(result)

	case "json":
		o, e := json.Marshal(finalResult)
		if e != nil {
			err = errors.Errorf("Failed to Marshal JSON. output:\n%s", e)
		} else {
			output = string(o)
		}
	case "yaml":
		o, e := yaml.Marshal(finalResult)
		if e != nil {
			err = errors.Errorf("Failed to Marshal YAML. output:\n%s", e)
		} else {
			output = string(o)
		}
	default:
		err = errors.Errorf("Unknown output format.\n%s", format)
	}

	return output, err

}

func formatProjectResult(format string, result []listProject) (string, error) {
	var output string
	var err error
	var finalResult interface{}

	finalResult = result

	switch format {
	case "":
		output = formatProjectText(result)

	case "json":
		o, e := json.Marshal(finalResult)
		if e != nil {
			err = errors.Errorf("Failed to Marshal JSON. output:\n%s", e)
		} else {
			output = string(o)
		}
	case "yaml":
		o, e := yaml.Marshal(finalResult)
		if e != nil {
			err = errors.Errorf("Failed to Marshal YAML. output:\n%s", e)
		} else {
			output = string(o)
		}
	default:
		err = errors.Errorf("Unknown output format.\n%s", format)
	}

	return output, err
}

func formatProjectText(result []listProject) string {

	table := uitable.New()
	table.MaxColWidth = 60
	table.AddRow("Name", "Ready", "CreateAt", "Message")

	for _, project := range result {
		table.AddRow(project.Name, project.Ready, project.CreatedAt, project.Message)
	}
	return fmt.Sprintf("%s", table.String())
}

func formatReleaseText(result []listRelease) string {

	table := uitable.New()
	table.MaxColWidth = 60
	table.AddRow("NAME", "Ready", "ChartName", "ChartVersion")
	for _, release := range result {
		table.AddRow(release.Name, release.Ready, release.ChartName, release.ChartVersion)
	}

	return fmt.Sprintf("%s", table.String())
}
