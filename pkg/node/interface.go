package node

import (
	"fmt"
	"walm/pkg/k8s/handler"
)


func GetNode() ([]NodeInfo, error){

	var nodeInfos []NodeInfo

	nodeHandler := handler.GetDefaultHandlerSet().GetNodeHandler()

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

	return nodeInfos, nil
}

func GetNodeLabels(nodeName string) (NodeLabelsInfo, error) {

	nodeHandler := handler.GetDefaultHandlerSet().GetNodeHandler()

	nodeInfo, err := nodeHandler.GetNode(nodeName)
	if err != nil {
		fmt.Println(err)
	}

	nodeLabelsInfo := NodeLabelsInfo{
		NodeLabels: nodeInfo.ObjectMeta.Labels,
	}

	return nodeLabelsInfo, nil
}

func UpdateNodeLabels(nodeName string, newLabels map[string]string) error {

	if newLabels == nil || len(newLabels) == 0 {
		return nil
	}

	nodeHandler := handler.GetDefaultHandlerSet().GetNodeHandler()

	nodeInfo, err := nodeHandler.GetNode(nodeName)
	if err != nil {
		return err
	}

	nodeLables := nodeInfo.ObjectMeta.Labels
	if nodeLables == nil || len(nodeLables) == 0 {
		nodeLables = newLabels
	} else {

		for key, value := range newLabels {
			nodeLables[key] = value
		}

	}
	_, err = nodeHandler.LabelNode(nodeName, nodeLables)
	if err != nil {
		return err
	}

	return nil
}

func AddNodeLabels(nodeName string, newLabels map[string]string) error {

	return UpdateNodeLabels(nodeName, newLabels)
}

func DelNodeLabels(nodeName string, newLabels map[string]string) error {

	if newLabels == nil || len(newLabels) == 0 {
		return nil
	}

	nodeHandler := handler.GetDefaultHandlerSet().GetNodeHandler()

	nodeInfo, err := nodeHandler.GetNode(nodeName)
	if err != nil {
		return err
	}

	nodeLables := nodeInfo.ObjectMeta.Labels
	if nodeLables == nil || len(nodeLables) == 0 {
		nodeLables = newLabels
	} else {

		for key, _ := range newLabels {
			delete(nodeLables, key)
		}

	}
	_, err = nodeHandler.LabelNode(nodeName, nodeLables)
	if err != nil {
		return err
	}

	return nil
}
