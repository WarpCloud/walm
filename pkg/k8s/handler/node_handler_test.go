package handler

import (
	"testing"
	k8sclient "walm/pkg/k8s/client"
	"fmt"
	"encoding/json"
	"walm/pkg/k8s/informer"
	"k8s.io/apimachinery/pkg/util/wait"
)

func Test(t *testing.T) {
	client, err := k8sclient.CreateApiserverClient("", "C:/kubernetes/0.5/kubeconfig")
	if err != nil {
		println(err.Error())
		return
	}

	clientEx, err := k8sclient.CreateApiserverClientEx("", "C:/kubernetes/0.5/kubeconfig")
	if err != nil {
		println(err.Error())
		return
	}

	factory := informer.NewInformerFactory(client, clientEx, 0)
	factory.Start(wait.NeverStop)
	factory.WaitForCacheSync(wait.NeverStop)

	nodeHandler := NewNodeHandler(client, factory.NodeLister)

	nodeList, err := nodeHandler.ListNodes(nil)
	if err != nil {
		fmt.Println(err)
	}

	e, err := json.Marshal(nodeList)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(e))

	node, err := nodeHandler.GetNode("172.16.1.175")
	if err != nil {
		fmt.Println(err)
	}

	e, err = json.Marshal(node)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(e))

	podList, err := nodeHandler.GetPodsOnNode("172.16.1.175", nil)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(len(podList.Items))
	e, err = json.Marshal(podList)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(e))

}