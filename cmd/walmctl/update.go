package main

import (
	"io"
	"github.com/spf13/cobra"
	"github.com/pkg/errors"
	"walm/cmd/walmctl/walmctlclient"
	"strings"
	"k8s.io/helm/pkg/strvals"
	"encoding/json"
)

const updateDesc = `This command update an existing release,
update release with --set or --withchart, update release support only
currently
`

type updateCmd struct {
	out    io.Writer
	sourceType string
	sourceName string
	withchart string
	properties string
	timeoutSec  int64
	async       bool
}

// Todo:// update project

func newUpdateCmd(out io.Writer) *cobra.Command {
	uc := updateCmd{out:out}

	cmd := &cobra.Command{
		Use: "update",
		Short: "update an existing release, update project will be support in the future",
		Long: updateDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if namespace == "" {
				return errors.New("flag --namespace/-n required")
			}
			if walmserver == "" {
				return errors.New("flag --server/-s required")

			}
			if len(args) != 2 {
				return errors.New("releaseName/projectName required, use update release [releaseName] or use project [projectName]")
			}

			if args[0] != "release" {
				return errors.New("release/project required, the walm source type you want to update")
			}

			uc.sourceType = args[0]
			uc.sourceName = args[1]

			return uc.run()
		},
	}
	cmd.Flags().StringVar(&uc.withchart, "withchart", "", "update release with local chart")
	cmd.Flags().Int64Var(&uc.timeoutSec, "timeoutSec", 0, "timeout, (default 0), available only when update release without local chart.")
	cmd.Flags().BoolVar(&uc.async, "async", true, "whether asynchronous, available only when update release without local chart.")
	cmd.Flags().StringVar(&uc.properties, "set", "", "set values on the command line (can specify multiple or separate values with commas: key1=val1,dependencies.guardian=...")
	return cmd
}


func (uc *updateCmd) run() error {

	client := walmctlclient.CreateNewClient(walmserver)
	resp, err := client.GetSource(namespace, uc.sourceName, uc.sourceType)
	if err != nil {
		return err
	}

	// 1.check release/project is found

	if resp.StatusCode() == 404 {
		return errors.Errorf("%s %s is not found.", uc.sourceType, uc.sourceName)
	}


	// 2.update config

	baseConfig := map[string]interface{}{}

	err = json.Unmarshal(resp.Body(), &baseConfig)
	if err != nil {
		return errors.Errorf("")
	}


	propertyArray := strings.Split(uc.properties, ",")
	for _, property := range propertyArray {
		property = strings.TrimSpace(property)
		strvals.ParseInto(property, baseConfig)
	}

	newConfig, err := json.Marshal(baseConfig)


	if uc.sourceType == "release" {

		if uc.withchart == "" {
			resp, err = client.UpdateRelease(namespace, string(newConfig), uc.async, uc.timeoutSec)
		} else {
			resp, err = client.UpdateReleaseWithChart(namespace, uc.sourceName, string(newConfig), uc.withchart)
		}

	}

	if err != nil {
		return errors.Errorf("update release with local chart failed")
	}
	return nil
}
