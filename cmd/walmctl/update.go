package main

import (
	"io"
	"github.com/spf13/cobra"
	"github.com/pkg/errors"
	"WarpCloud/walm/cmd/walmctl/walmctlclient"
	"strings"
	"fmt"
	"encoding/json"
	"WarpCloud/walm/pkg/release"
	"github.com/tidwall/gjson"
	"strconv"
	"github.com/tidwall/sjson"
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
	setproperties string
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
				return errNamespaceRequired
			}
			if walmserver == "" {
				return errServerRequired
			}
			if len(args) != 2 {
				return errors.New("releaseName/projectName required, use update release [releaseName] or use project [projectName]")
			}

			if args[0] != "release" {
				return errors.New("now support release only")
			}

			uc.sourceType = args[0]
			uc.sourceName = args[1]

			return uc.run()
		},
	}
	cmd.Flags().StringVar(&uc.withchart, "withchart", "", "update release with local chart")
	cmd.Flags().Int64Var(&uc.timeoutSec, "timeoutSec", 0, "timeout, (default 0), available only when update release without local chart.")
	cmd.Flags().BoolVar(&uc.async, "async", true, "whether asynchronous, available only when update release without local chart.")
	cmd.Flags().StringVar(&uc.setproperties, "set-string", "", "set values on the command line (can specify multiple or separate values with commas: pathA=valA,pathB.1=valB,...")
	return cmd
}


func (uc *updateCmd) run() error {

	client := walmctlclient.CreateNewClient(walmserver)
	resp, err := client.GetRelease(namespace, uc.sourceName)
	if err != nil {
		return err
	}

	if resp.StatusCode() == 404 {
		return errors.Errorf("%s %s is not found.", uc.sourceType, uc.sourceName)
	}

	var releaseInfo release.ReleaseInfoV2
	var releaseInfoByte []byte
	var releaseInfoStr string

	err = json.Unmarshal(resp.Body(), &releaseInfo)
	if err != nil {
		return err
	}

	releaseInfoByte, err = json.Marshal(releaseInfo)
	if err != nil {
		return err
	}

	releaseInfoStr = string(releaseInfoByte)

	if uc.sourceType == "release" {

		propertySetArray := strings.Split(uc.setproperties, ",")

		for _, propertySet := range propertySetArray {
			propertySet = strings.TrimSpace(propertySet)
			propertyMap := strings.Split(propertySet, "=")
			if len(propertyMap) != 2 {
				return errors.Errorf("set values error, params should like --set pathA=valueA, pathB=valueB...")
			}
			propertyKey := propertyMap[0]
			propertyVal := propertyMap[1]

			result := gjson.Get(releaseInfoStr, propertyKey)
			if !result.Exists() {
				return errors.Errorf("path error: %s not exist in releaseInfo", propertyKey)
			}

			var destVal interface{}

			switch result.Type.String() {

			case "True", "False":
				destVal, err = strconv.ParseBool(propertyVal)
			case "String":
				destVal = propertyVal
			case "Number":
				destVal, err = strconv.Atoi(propertyVal)
				if err != nil {
					destVal, err = strconv.ParseFloat(propertyVal, 64)
				}
			case "JSON":
				err = json.Unmarshal([]byte(propertyVal), &destVal)
			default:

			}

			if err != nil {
				return err
			}

			releaseInfoStr, err = sjson.Set(releaseInfoStr, propertyKey, destVal)
			if err != nil {
				return err
			}
		}
	}

	if uc.sourceType == "release" {

		if uc.withchart == "" {
			resp, err = client.UpdateRelease(namespace, releaseInfoStr, uc.async, uc.timeoutSec)
		} else {
			resp, err = client.UpdateReleaseWithChart(namespace, uc.sourceName, uc.withchart)
		}

	}

	if err != nil {
		return errors.Errorf("update release with local chart failed")
	}

	fmt.Printf("Update %s %s succeed\n", uc.sourceType, uc.sourceName)

	return nil
}
