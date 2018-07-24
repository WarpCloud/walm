package handler

import (
	"testing"
	"walm/pkg/k8s/client"
	"fmt"
	"encoding/json"
)

func Test(t *testing.T) {
	client, err := client.CreateApiserverClient("", "C:/kubernetes/0.5/kubeconfig")
	if err != nil {
		println(err.Error())
		return
	}

	nodeHandler := NewNodeHandler(client)

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