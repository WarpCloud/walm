package main

import (
	"WarpCloud/walm/cmd/walmctl/util/walmctlclient"
	k8sModel "WarpCloud/walm/pkg/models/k8s"
	"encoding/json"
	"github.com/migration/pkg/apis/tos/v1beta1"
	"k8s.io/klog"
	"strconv"

	"fmt"
	"github.com/ghodss/yaml"
	"github.com/go-resty/resty"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io"
)

const getDesc = `
Get a walm release or project detail info.
Options:
use --output/-o to print with json/yaml format.

You must specify the type of resource to get. Valid resource types include:
  * release
  * project
  * migration

[release]
walmctl get release xxx -n/--namespace xxx

[project]
walmctl get project xxx -n/--namespace xxx

[migration]
walmctl get migration pod xxx -n/--namespace xxx
walmctl get migration node xxx
walmctl get migration node xxx --detail
`

type getCmd struct {
	sourceType string
	sourceName string
	subType    string
	output     string
	detail     bool
	out        io.Writer
}

func newGetCmd(out io.Writer) *cobra.Command {
	gc := getCmd{out: out}

	cmd := &cobra.Command{
		Use:                   "get",
		DisableFlagsInUseLine: true,
		Short:                 "get [release | project | migration]",
		Long:                  getDesc,
		RunE: func(cmd *cobra.Command, args []string) error {

			if walmserver == "" {
				return errServerRequired
			}

			if err := checkResourceType(args[0]); err != nil {
				return err
			}
			gc.sourceType = args[0]
			if gc.sourceType == "migration" {
				if len(args) != 3 {
					return errors.Errorf("arguments error, get migration pod/node xxx")
				}
				if args[1] != "pod" && args[1] != "node" {
					return errors.Errorf("arguments error, invalid migration type: %s", args[1])
				}
				gc.subType = args[1]
				gc.sourceName = args[2]
			} else {
				if len(args) != 2 {
					return errors.Errorf("arguments error, get release/project xxx")
				}
				gc.sourceName = args[1]
			}

			if namespace == "" && gc.subType != "node" {
				return errNamespaceRequired
			}

			return gc.run()
		},
	}

	cmd.PersistentFlags().BoolVar(&gc.detail, "detail", false, "Print detail migration information for each pod")
	cmd.PersistentFlags().StringVarP(&gc.output, "output", "o", "json", "-o, --output='': Output format for detail description. Support: json, yaml")
	return cmd
}

func (gc *getCmd) run() error {

	var resp *resty.Response
	var err error

	client, err := walmctlclient.CreateNewClient(walmserver, enableTLS, rootCA)
	if err != nil {
		klog.Errorf("failed to create walmctl client: %s", err.Error())
		return err
	}
	if err = client.ValidateHostConnect(walmserver); err != nil {
		return err
	}
	if gc.sourceType == "release" {
		resp, err = client.GetRelease(namespace, gc.sourceName)
	} else if gc.sourceType == "project" {
		resp, err = client.GetProject(namespace, gc.sourceName)
	} else if gc.sourceType == "migration" {
		if gc.subType == "node" {
			resp, err = client.GetNodeMigration(gc.sourceName)
			if !gc.detail {
				migStatus, errMsgs, err := getMigDetails(client, gc.sourceName)
				if err != nil {
					klog.Errorf("failed to get node migration response: %s", err.Error())
					return err
				}
				progress := "[" + bar(migStatus.Succeed, migStatus.Total) + "]" + strconv.Itoa(migStatus.Succeed) + " / " + strconv.Itoa(migStatus.Total)
				fmt.Printf("\r%s", progress)
				fmt.Println()
				if len(errMsgs) > 0 {
					fmt.Println("[Error]:")
					for _, errMsg := range errMsgs {
						fmt.Println(errMsg)
					}
				}
				if migStatus.Succeed == migStatus.Total {
					fmt.Println("migration node succeed")
				}
				return nil
			}
		} else {
			resp, err = client.GetPodMigration(namespace, gc.sourceName)
		}
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

func getMigDetails(client *walmctlclient.WalmctlClient, node string) (k8sModel.MigStatus, []string, error) {
	var migStatus k8sModel.MigStatus
	var errMsgs []string
	resp, err := client.GetNodeMigration(node)
	if err != nil {
		klog.Errorf("failed to get node migration: %s", err.Error())
		return migStatus, errMsgs, err
	}
	err = json.Unmarshal(resp.Body(), &migStatus)
	if err != nil {
		klog.Errorf("failed to unmarshal node migrate response: %s", err.Error())
		return migStatus, errMsgs, err
	}
	for _, item := range migStatus.Items {
		if item.State.Status == v1beta1.MIG_FAILED {

			errMsgs = append(errMsgs, "[Pod] "+item.Spec.Namespace+"/"+item.Spec.PodName+": "+item.State.Message)
		}
	}
	return migStatus, errMsgs, nil
}

func bar(count, size int) string {
	str := ""
	for i := 0; i < size; i++ {
		if i < count {
			str += "="
		} else {
			str += " "
		}
	}
	return str
}
