package main

import (
	"WarpCloud/walm/cmd/walmctl/util/walmctlclient"
	k8sclient "WarpCloud/walm/pkg/k8s/client"
	k8sModel "WarpCloud/walm/pkg/models/k8s"
	"encoding/json"
	"fmt"
	"github.com/migration/pkg/apis/tos/v1beta1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var longMigrateHelp = `
To ensure migrate node work, we need set the kubeconfig.
In Cluster: 
	We use the env $KUBECONFIG as default kubeconfig, if not work, you may need define yourself.
Out of Cluster:
	You must set the --kubeconfig or export ${KUBECONFIG}.
`

const (
	MigNodeType string = "node"
	MigPodType  string = "pod"
)
const appTypeKey = "transwarp.name"

type migrateOptions struct {
	client     *walmctlclient.WalmctlClient
	kind       string
	name       string
	kubeconfig string
	destNode   string
	out        io.Writer
}

func newMigrationCmd(out io.Writer) *cobra.Command {
	migrate := &migrateOptions{out: out}

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "migrate node, pod",
		Long:  longMigrateHelp,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("Arguments error, migrate node/pod nodeName/podName")
			}
			if args[0] != MigPodType && args[0] != MigNodeType {
				return errors.Errorf("Unsupport kind: %s, pod or node support only", args[0])
			}

			migrate.kind = args[0]
			migrate.name = args[1]
			return migrate.run()
		},
	}
	cmd.PersistentFlags().StringVar(&migrate.destNode, "destNode", "", "dest node to migrate")
	cmd.PersistentFlags().StringVar(&migrate.kubeconfig, "kubeconfig", os.Getenv("KUBECONFIG"), "k8s cluster config")

	return cmd
}

func (migrate *migrateOptions) run() error {
	var err error
	if walmserver == "" {
		return errServerRequired
	}

	client, err := walmctlclient.CreateNewClient(walmserver, enableTLS, rootCA)
	if err != nil {
		klog.Errorf("failed to create walmctl client: %s", err.Error())
		return err
	}
	if err = client.ValidateHostConnect(walmserver); err != nil {
		return err
	}
	migrate.client = client

	switch migrate.kind {
	case MigPodType:
		err = migrate.createPodMigTask()
	case MigNodeType:
		err = migrate.createNodeMigTask()
	}

	return err
}

func (migrate *migrateOptions) createPodMigTask() error {
	err := migratePodPreCheck(migrate.client, namespace, migrate.name, migrate.destNode)
	if err != nil {
		return err
	}

	_, err = migrate.client.MigratePod(namespace, &k8sModel.PodMigRequest{
		PodName:  migrate.name,
		DestNode: migrate.destNode,
	})
	if err != nil {
		return errors.Errorf("Failed to migrate pod: %s", err.Error())
	}

	fmt.Printf("Create pod migrate task succeed.\n")
	return nil
}

func (migrate *migrateOptions) createNodeMigTask() error {
	var err error

	migrate.kubeconfig, err = filepath.Abs(migrate.kubeconfig)
	if err != nil {
		return err
	}

	k8sClient, err := k8sclient.NewClient("", migrate.kubeconfig)
	if err != nil {
		return err
	}
	if err = envPreCheck(migrate.client, k8sClient, migrate.name, migrate.destNode); err != nil {
		return err
	}
	if err = cordonNode(k8sClient, migrate.name); err != nil {
		return err
	}

	podList, err := getSupportedPodListFromNode(k8sClient, migrate.name)
	if err != nil {
		return err
	}

	/* check all pods ready to be migrated */
	for _, pod := range podList {
		err = migratePodPreCheck(migrate.client, pod.Namespace, pod.Name, migrate.destNode)
		if err != nil {
			return err
		}
	}

	for _, pod := range podList {
		_, err = migrate.client.MigratePod(pod.Namespace, &k8sModel.PodMigRequest{
			PodName:  pod.Name,
			DestNode: migrate.destNode,
			Labels:   map[string]string{"migType": "node", "srcNode": migrate.name},
		})
		if err != nil {
			klog.Errorf("Send migrate pod request failed: %s", err.Error())
			return err
		}
	}
	time.Sleep(3 * time.Second)
	fmt.Printf("Create node migrate task succeed. use `walmctl get migration node nodeName` for later information\n")
	return nil
}

func getSupportedPodListFromNode(k8sClient *kubernetes.Clientset, srcHost string) ([]corev1.Pod, error) {
	blackAppList := []string{"txsql", "shivatabletserver", "shivamaster"}

	podList := &corev1.PodList{
		Items: []corev1.Pod{},
	}
	pods, err := k8sClient.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to list pods: %s", err.Error())
		return nil, err
	}

	for _, pod := range pods.Items {
		if pod.Spec.NodeName == srcHost {
			for _, ownerReference := range pod.OwnerReferences {
				if ownerReference.Kind == "StatefulSet" {
					if appType, ok := pod.Labels[appTypeKey]; ok {
						isInBlackList := false
						for _, blackApp := range blackAppList {
							if appType == blackApp {
								isInBlackList = true
							}
						}
						if pod.Namespace == "kube-system" {
							isInBlackList = true
						}
						if isInBlackList {
							klog.Warningf("ignore %s because applicaion type is in blacklist", appType)
							continue
						}
					}
					podList.Items = append(podList.Items, pod)
				}
			}
		}
	}
	return podList.Items, nil
}

func envPreCheck(client *walmctlclient.WalmctlClient, k8sClient *kubernetes.Clientset, srcHost string, destHost string) error {
	nodeList, err := k8sClient.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to get nodes: %s", err.Error())
		return err
	}
	if len(nodeList.Items) < 2 {
		return errors.Errorf("Only one node, migration make no sense")
	}

	if destHost != "" {
		destNode, err := k8sClient.CoreV1().Nodes().Get(destHost, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Failed to get node %s: %s", destHost, err.Error())
			return err
		}
		if destNode.Spec.Unschedulable {
			return errors.Errorf("Dest node is Unschedulable, run `kubectl uncordon ...`")
		}
	}

	for _, node := range nodeList.Items {
		migStatus, errMsgs, err := getMigDetails(client, node.Name)
		if err != nil {
			return err
		}
		if migStatus.Total == 0 {
			continue
		}
		if migStatus.Succeed+len(errMsgs) < migStatus.Total {
			return errors.Errorf("Node %s is in migration progress, you must not migrate two node at one time", node.Name)
		}
	}

	return nil
}

func cordonNode(k8sClient *kubernetes.Clientset, srcHost string) error {
	srcNode, err := k8sClient.CoreV1().Nodes().Get(srcHost, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("Failed to get node %s: %s", srcHost, err.Error())
		return err
	}

	/* cordon node */
	if srcNode.Spec.Unschedulable == false {
		oldData, err := json.Marshal(srcNode)
		if err != nil {
			return err
		}

		srcNode.Spec.Unschedulable = true
		newData, err := json.Marshal(srcNode)
		if err != nil {
			return err
		}
		patchBytes, patchErr := strategicpatch.CreateTwoWayMergePatch(oldData, newData, srcNode)
		if patchErr == nil {
			_, err = k8sClient.CoreV1().Nodes().Patch(srcNode.Name, types.StrategicMergePatchType, patchBytes)

		} else {
			_, err = k8sClient.CoreV1().Nodes().Update(srcNode)
		}
		if err != nil {
			return errors.Errorf("Unable to cordon node %q: %v\n", srcNode.Name, err)
		}
	} else {
		klog.Infof("Node %s is unschedulable now", srcNode.Name)
	}

	return nil
}

func migratePodPreCheck(client *walmctlclient.WalmctlClient, namespace string, name string, destNode string) error {
	if namespace == "" {
		return errNamespaceRequired
	}
	resp, err := client.GetPodMigration(namespace, name)
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			return errors.Errorf("Failed to get pod migration: %s", err.Error())
		}
	}
	var podMig k8sModel.Mig
	if resp != nil {
		err = json.Unmarshal(resp.Body(), &podMig)
		if err != nil {
			klog.Errorf("Failed to unmarshal response body to pod migration status")
			return err
		}
		switch podMig.State.Status {
		case v1beta1.MIG_CREATED, v1beta1.MIG_IN_PROGRESS, "":
			return errors.Errorf("Pod %s/%s migration in progress, please wait for the last process end.", podMig.Spec.Namespace, podMig.Spec.PodName)
		case v1beta1.MIG_FAILED:
			_, err = client.DeletePodMigration(podMig.Spec.Namespace, podMig.Spec.PodName)
			if err != nil {
				klog.Errorf("Failed to delete last failed pod migration %s: %s", podMig.Name, err.Error())
				return err
			}
			return errors.Errorf("Last migration for pod failed: %s.\nWe already help you delete the failed pod migration, please fix and try!", podMig.State.Message)
		case v1beta1.MIG_FINISH:
			_, err = client.DeletePodMigration(podMig.Spec.Namespace, podMig.Spec.PodName)
			if err != nil {
				klog.Errorf("Failed to delete pod migration: %s", err.Error())
				return err
			}
		}
	}

	return nil
}
