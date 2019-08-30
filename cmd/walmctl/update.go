package main

import (
	"WarpCloud/walm/cmd/walmctl/util/walmctlclient"

	"github.com/go-resty/resty"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"fmt"
	"encoding/json"
	"WarpCloud/walm/pkg/models/release"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"io"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
)

const updateDesc = `This command update an existing release,
update release with --set or --withchart, update release support only
currently
`

type updateCmd struct {
	out           io.Writer
	sourceType    string
	sourceName    string
	withchart     string
	setproperties string
	file          string
	timeoutSec    int64
	async         bool
}

// Todo:// update project

func newUpdateCmd(out io.Writer) *cobra.Command {
	uc := updateCmd{out: out}

	cmd := &cobra.Command{
		Use:   "update",
		Short: "update an existing release, update project will be support in the future",
		Long:  updateDesc,
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
	cmd.Flags().StringVarP(&uc.file, "file", "f", "", "absolutely or relative path to source file")
	return cmd
}

func (uc *updateCmd) run() error {
	var (
		err            error
		resp           *resty.Response
		releaseInfo    release.ReleaseInfoV2
		releaseRequest *release.ReleaseRequestV2
	)
	client := walmctlclient.CreateNewClient(walmserver)
	if err = client.ValidateHostConnect(); err != nil {
		return err
	}

	// file exists => build releaseRequest from file => set values => sendRequest
	// file not exist => fetch releaseInfo to parse releaseRequest => set values => sendRequest
	if uc.file != "" {
		fileName, err := filepath.Abs(uc.file)
		if err != nil {
			return errors.Errorf("%s not exists.\n", uc.file)
		}
		fileByte, err := ioutil.ReadFile(fileName)
		if err != nil {
			return err
		}
		err = json.Unmarshal(fileByte, &releaseRequest)
		if err != nil {
			return err
		}
	} else {
		resp, err = client.GetRelease(namespace, uc.sourceName)
		if err != nil {
			return err
		}
		if resp.StatusCode() == 404 {
			return errors.Errorf("%s %s is not found.\n", uc.sourceType, uc.sourceName)
		}
		err = json.Unmarshal(resp.Body(), &releaseInfo)
		releaseRequest = releaseInfo.BuildReleaseRequestV2()
		if err != nil {
			return err
		}
	}

	configValuesByte, err := json.Marshal(releaseRequest.ConfigValues)
	if err != nil {
		return err
	}
	configValuesStr := string(configValuesByte)

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

			result := gjson.Get(configValuesStr, propertyKey)

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
			case "Null":
				destVal = propertyVal
			default:

			}

			if err != nil {
				return err
			}

			configValuesStr, err = sjson.Set(configValuesStr, propertyKey, destVal)
			if err != nil {
				return err
			}
		}
		err = json.Unmarshal([]byte(configValuesStr), &releaseRequest.ConfigValues)
		if err != nil {
			return err
		}
	}

	if uc.sourceType == "release" {
		releaseRequestByte, err := json.Marshal(releaseRequest)
		if err != nil {
			return err
		}
		uc.withchart, err = filepath.Abs(uc.withchart)
		if err != nil {
			return err
		}
		_, err = ioutil.ReadFile(uc.withchart)
		if err != nil {
			return err
		}
		if uc.withchart == "" {
			resp, err = client.UpdateRelease(namespace, string(releaseRequestByte), uc.async, uc.timeoutSec)
		} else {
			resp, err = client.UpdateReleaseWithChart(namespace, uc.sourceName, uc.withchart, string(releaseRequestByte))
		}
		if err != nil {
			return errors.Errorf("update release failed, %s", err.Error())
		}
	}

	fmt.Printf("update %s %s succeed\n", uc.sourceType, uc.sourceName)
	return nil
}
