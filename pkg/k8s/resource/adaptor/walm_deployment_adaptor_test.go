package adaptor

import (
	"testing"
	"walm/pkg/k8s/client"
	"fmt"
	"encoding/json"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"walm/pkg/k8s/utils"
)

func TestDeploymentAdaptor(t *testing.T) {
	client, err := client.CreateFakeApiserverClient("", "C:/kubernetes/0.5/kubeconfig")
	if err != nil {
		println(err.Error())
		return
	}

	deployment, err := client.ExtensionsV1beta1().Deployments("default").Get("hello", v1.GetOptions{})
	if err != nil {
		println(err.Error())
		return
	}

	selectorStr, err := utils.ConvertLabelSelectorToStr(deployment.Spec.Selector)
	if err != nil {
		println(err.Error())
		return
	}

	pods, err := client.CoreV1().Pods("default").List(v1.ListOptions{LabelSelector: selectorStr})
	if err != nil {
		println(err.Error())
		return
	}

	walmPods := []*WalmPod{}
	for _, pod := range pods.Items {
		walmPods = append(walmPods, BuildWalmPod(pod))
	}

	e, err := json.Marshal(walmPods)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(e))

	e, err = json.Marshal(BuildWalmDeploymentState(deployment, walmPods))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(e))

}

func TestParsePod(t *testing.T) {
	walmPods := []*WalmPod{buildWalmPod("test1", "Pending"), buildWalmPod("test2", "Terminating")}
	buildWalmPod("test1", "Terminating")

	allPodsTerminating, unknownPod, pendingPod, runningPod := parsePods(walmPods)

	fmt.Println(allPodsTerminating, unknownPod, pendingPod, runningPod)
}

func buildWalmPod(name string, state string) *WalmPod {
	return &WalmPod{
		WalmMeta: WalmMeta{Name: name, State: WalmState{Status: state}},
	}
}
