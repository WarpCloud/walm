package handler

import (
	"testing"
	k8sclient "walm/pkg/k8s/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"fmt"
)

func Test(t *testing.T) {
	client, err := k8sclient.CreateFakeApiserverClient("", "C:/kubernetes/0.5/kubeconfig")
	if err != nil {
		println(err.Error())
		return
	}

	_, err = client.CoreV1().Nodes().Get("172.26.0.5", metav1.GetOptions{})
	if err != nil {
		fmt.Println(err.Error())
	}

	//clientEx, err := k8sclient.CreateFakeApiserverClientEx("", "C:/kubernetes/0.5/kubeconfig")
	//if err != nil {
	//	println(err.Error())
	//	return
	//}
	//
	//factory := informer.NewFakeInformerFactory(client, clientEx, 0)
	//factory.Start(wait.NeverStop)
	//factory.WaitForCacheSync(wait.NeverStop)
	//
	//nodeHandler := NodeHandler{client, factory.NodeLister}
	//
	//nodeList, err := nodeHandler.ListNodes(nil)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	//e, err := json.Marshal(nodeList)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//fmt.Println(string(e))
	//
	//node, err := nodeHandler.GetNode("172.16.1.175")
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	//e, err = json.Marshal(node)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//fmt.Println(string(e))
	//
	//podList, err := nodeHandler.GetPodsOnNode("172.16.1.175", nil)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	//fmt.Println(len(podList.Items))
	//e, err = json.Marshal(podList)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//fmt.Println(string(e))

}