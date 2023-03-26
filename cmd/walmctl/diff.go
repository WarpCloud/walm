package main

import (
	"WarpCloud/walm/cmd/walmctl/util"
	diffUtil "WarpCloud/walm/cmd/walmctl/util/diff"
	"WarpCloud/walm/cmd/walmctl/util/walmctlclient"
	k8sclient "WarpCloud/walm/pkg/k8s/client"
	"WarpCloud/walm/pkg/models/k8s"
	"WarpCloud/walm/pkg/models/release"
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/kubectl/util/i18n"
	"k8s.io/kubernetes/pkg/kubectl/util/templates"
	"os"
	"path/filepath"
	"strings"
)

var (
	diffLong = templates.LongDesc(i18n.T(`
		Compares resources between runtime release and chart dry run release.
		Default, to better compare, the namespace && name of dry run release
		are same with runtime release.

		The flag --withchart can be used to specify the local chart to dry run.
		The flag --file/-f is filepath to custom the dry-run release, the value will
		be merged into chart values.

		The flag --kubeconfig are used to interact with k8s resources, you need to know:
		1. In cluster: you need not use --kubeconfig, we already help you set the env KUBECONFIG.
		2. Out of cluster: you must set --kubeconfig

		The flag --resource stands for the specific resource you would like to compare.
		If you just want compare the StatefulSet(eg: ConfigMap, Deployment, Secret ...)between runtime release
		and dry run release, just set --resource statefulset. If --resource not set or be set to all, we compare 
		all of the resource type, invalid k8s resource type caused error.
		
		For the results output, the "+" stands for which the dry run resource have while the runtime resource missing.
		the "-" stands for which the dry run resource missing.`))

	diffExample = templates.Examples(i18n.T(`
		# Compare the service between runtime release between dry run release
		walmctl -n xxx diff release zookeeper-dzy --withchart /Users/corndai/Desktop/zookeeper-6.2.0.tgz --resource service

		# Compare  all resources between runtime release between dry run release
        walmctl -n xxx diff release zookeeper-dzy --withchart /Users/corndai/Desktop/zookeeper-6.2.0.tgz --resource all
        walmctl -n xxx diff release zookeeper-dzy --withchart /Users/corndai/Desktop/zookeeper-6.2.0.tgz`))
)

type diffCmd struct {
	release    string
	name       string
	kubeconfig string
	withchart  string
	file       string
	resource   string
	out        io.Writer
}

func newDiffCmd(out io.Writer) *cobra.Command {
	diff := &diffCmd{out: out}

	cmd := &cobra.Command{
		Use:                   "diff",
		DisableFlagsInUseLine: false,
		Short:                 i18n.T("diff resources between runtime release and chart dry-run release"),
		Long:                  diffLong,
		Example:               fmt.Sprintf(diffExample),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.Errorf("Arguments invalid, format like `diff release zookeeper-test` instead")
			}
			if args[0] != "release" {
				return errors.Errorf("Unsupported object type, release only")
			}
			if namespace == "" {
				return errNamespaceRequired
			}
			diff.release = args[1]
			return diff.run()
		},
	}
	cmd.PersistentFlags().StringVar(&diff.withchart, "withchart", "", "compare release with local chart , absolutely or relative path to source file")
	cmd.PersistentFlags().StringVarP(&diff.file, "file", "f", "", "filepath to custom dry-run release, optional")
	cmd.Flags().StringVar(&diff.kubeconfig, "kubeconfig", os.Getenv("KUBECONFIG"), "k8s cluster config")
	cmd.PersistentFlags().StringVar(&diff.resource, "resource", "all", "compare custom resource, default all")

	return cmd
}

func (diff *diffCmd) run() error {
	if diff.withchart == "" && diff.file == "" {
		return errors.Errorf("One of flags --withchart and --file must be set.")
	}

	var (
		err              error
		manifest         []map[string]interface{}
		configValues     map[string]interface{}
		destConfigValues map[string]interface{}
		releaseInfo      release.ReleaseInfoV2
	)

	// get original release
	client, err := walmctlclient.CreateNewClient(walmserver, enableTLS, rootCA)
	if err != nil {
		klog.Errorf("failed to create walmctl client: %s", err.Error())
		return err
	}
	if err = client.ValidateHostConnect(walmserver); err != nil {
		return err
	}
	resp, err := client.GetRelease(namespace, diff.release)
	if err != nil {
		return nil
	}
	err = json.Unmarshal(resp.Body(), &releaseInfo)
	if err != nil {
		return err
	}

	// get dry run manifest
	if diff.withchart != "" {
		diff.withchart, err = filepath.Abs(diff.withchart)
		if err != nil {
			return err
		}
	}

	destConfigValues = make(map[string]interface{}, 0)
	if diff.file != "" {
		filePath, err := filepath.Abs(diff.file)
		if err != nil {
			return err
		}
		fileBytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			klog.Errorf("read file %s error %v", diff.file, err)
			return err
		}
		err = yaml.Unmarshal(fileBytes, &configValues)
		if err != nil {
			klog.Errorf("yaml Unmarshal file %s error %v", diff.file, err)
			return err
		}
		destConfigValues, _, _, err = util.SmartConfigValues(configValues)
		if err != nil {
			klog.Errorf("smart yaml Unmarshal file %s error %v", diff.file, err)
			return err
		}
	}

	response, err := client.DryRunCreateRelease(namespace, diff.withchart, diff.release, destConfigValues)
	if err != nil {
		klog.Errorf("dry run release with chart %s failed: %s", diff.withchart, err.Error())
		return err
	}
	err = json.Unmarshal(response.Body(), &manifest)
	if err != nil {
		return err
	}

	// get k8s resources
	diff.kubeconfig, err = filepath.Abs(diff.kubeconfig)
	if err != nil {
		return err
	}
	k8sClient, err := k8sclient.NewClient("", diff.kubeconfig)
	if err != nil {
		return err
	}

	resList := []string{diff.resource}
	if diff.resource == "all" {
		resList = []string{"ConfigMap", "DaemonSet", "Deployment", "Ingress", "Job", "Secret", "StatefulSet", "Service"}
	}

	for _, res := range resList {
		fmt.Println(res)
		k8sResources, err := getRuntimeResource(k8sClient, releaseInfo.Status, res)
		if err != nil {
			return err
		}
		k8sManifest, err := json.MarshalIndent(k8sResources, "", "  ")
		if err != nil {
			return err
		}
		manifest = filterManifest(manifest, res)
		manifestByte, err := json.MarshalIndent(manifest, "", "  ")
		if err != nil {
			return err
		}

		fmt.Println(diffUtil.Diff(fmt.Sprintf("%v", string(manifestByte)), fmt.Sprintf("%v", string(k8sManifest))))
	}

	return nil
}

func filterManifest(manifest []map[string]interface{}, res string) []map[string]interface{} {
	var results []map[string]interface{}
	for _, obj := range manifest {
		if strings.ToLower(fmt.Sprintf("%v", obj["kind"])) == res {
			results = append(results, obj)
		}
	}
	return results
}

func getRuntimeResource(client *kubernetes.Clientset, status *k8s.ResourceSet, resource string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	var result map[string]interface{}
	switch strings.ToLower(resource) {
	case "configmap":
		if len(status.ConfigMaps) > 0 {
			for _, configmap := range status.ConfigMaps {
				k8sConfigMap, err := client.CoreV1().ConfigMaps(configmap.Namespace).Get(configmap.Name, metav1.GetOptions{})
				if err != nil {
					return nil, err
				}
				k8sConfigMapByte, err := json.Marshal(k8sConfigMap)
				if err != nil {
					return nil, err
				}
				err = json.Unmarshal(k8sConfigMapByte, &result)
				if err != nil {
					return nil, err
				}
				results = append(results, result)
			}
		}
	case "daemonset":
		if len(status.DaemonSets) > 0 {
			for _, daemonset := range status.DaemonSets {
				k8sDaemonSet, err := client.AppsV1beta1().Deployments(daemonset.Namespace).Get(daemonset.Name, metav1.GetOptions{})
				if err != nil {
					return nil, err
				}
				k8sDaemonSetByte, err := json.Marshal(k8sDaemonSet)
				if err != nil {
					return nil, err
				}
				err = json.Unmarshal(k8sDaemonSetByte, &result)
				if err != nil {
					return nil, err
				}
				results = append(results, result)
			}
		}
	case "deployment":
		if len(status.Deployments) > 0 {
			for _, deployment := range status.Deployments {
				k8sDeployment, err := client.AppsV1beta1().Deployments(deployment.Namespace).Get(deployment.Name, metav1.GetOptions{})
				if err != nil {
					return nil, err
				}
				k8sDeploymentByte, err := json.Marshal(k8sDeployment)
				if err != nil {
					return nil, err
				}
				err = json.Unmarshal(k8sDeploymentByte, &result)
				if err != nil {
					return nil, err
				}
				results = append(results, result)
			}
		}
	case "ingress":
		if len(status.Ingresses) > 0 {
			for _, ingress := range status.Ingresses {
				k8sIngress, err := client.ExtensionsV1beta1().Ingresses(ingress.Namespace).Get(ingress.Name, metav1.GetOptions{})
				if err != nil {
					return nil, err
				}
				k8sIngressByte, err := json.Marshal(k8sIngress)
				if err != nil {
					return nil, err
				}
				err = json.Unmarshal(k8sIngressByte, &result)
				if err != nil {
					return nil, err
				}
				results = append(results, result)
			}
		}
	case "job":
		if len(status.Jobs) > 0 {
			for _, job := range status.Jobs {
				k8sJob, err := client.BatchV1().Jobs(job.Namespace).Get(job.Name, metav1.GetOptions{})
				if err != nil {
					return nil, err
				}
				k8sJobByte, err := json.Marshal(k8sJob)
				if err != nil {
					return nil, err
				}
				err = json.Unmarshal(k8sJobByte, &result)
				if err != nil {
					return nil, err
				}
				results = append(results, result)
			}
		}
	case "secret":
		if len(status.Secrets) > 0 {
			for _, secret := range status.Secrets {
				k8sSecret, err := client.CoreV1().Secrets(secret.Namespace).Get(secret.Name, metav1.GetOptions{})
				if err != nil {
					return nil, err
				}
				k8sSecretByte, err := json.Marshal(k8sSecret)
				if err != nil {
					return nil, err
				}
				err = json.Unmarshal(k8sSecretByte, &result)
				if err != nil {
					return nil, err
				}
				results = append(results, result)
			}
		}
	case "statefulset":
		if len(status.StatefulSets) > 0 {
			for _, statefulset := range status.StatefulSets {
				k8sStatefulSet, err := client.AppsV1beta1().StatefulSets(statefulset.Namespace).Get(statefulset.Name, metav1.GetOptions{})
				if err != nil {
					return nil, err
				}
				k8sStatefulSetByte, err := json.Marshal(k8sStatefulSet)
				if err != nil {
					return nil, err
				}
				err = json.Unmarshal(k8sStatefulSetByte, &result)
				if err != nil {
					return nil, err
				}
				results = append(results, result)
			}
		}
	case "service":
		if len(status.Services) > 0 {
			for _, service := range status.Services {
				k8sService, err := client.CoreV1().Services(service.Namespace).Get(service.Name, metav1.GetOptions{})
				if err != nil {
					return nil, err
				}
				k8sServiceByte, err := json.Marshal(k8sService)
				if err != nil {
					return nil, err
				}
				err = json.Unmarshal(k8sServiceByte, &result)
				if err != nil {
					return nil, err
				}
				results = append(results, result)
			}
		}
	}
	return results, nil
}
