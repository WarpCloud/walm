package node

import (
	"walm/pkg/k8s/client"
	"fmt"
	"walm/pkg/k8s/handler"
	"walm/pkg/k8s/informer"
)


func GetNode() ([]NodeInfo, error){

	var nodeInfos []NodeInfo

	client := client.GetDefaultClient()

	nodeHandler := handler.NewNodeHandler(client, informer.Factory.NodeLister)

	items, err := nodeHandler.ListNodes(nil)
	if err != nil {
		fmt.Println(err)
	}

	if items == nil || len(items) == 0 {
		return nodeInfos, nil
	}

	for _, item := range items {

		conditions := item.Status.Conditions
		if conditions == nil || len(conditions) == 0 {
			return nodeInfos, nil
		}

		var nodeStatus = "NotReady"
		for _, condition := range conditions {
			if condition.Type == "Ready" &&  condition.Status == "True" {
				nodeStatus = "Ready"
				break
			}
			continue
		}

		nodeInfo := NodeInfo{
			NodeName: item.ObjectMeta.Name,
			NodeIp: item.Status.Addresses[0].Address,
			NodeLabels: item.ObjectMeta.Labels,
			NodeStatus: nodeStatus,
		}
		nodeInfos = append(nodeInfos, nodeInfo)

	}

	return  nodeInfos, nil
}


