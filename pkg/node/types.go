package node

//Node Info
type NodeInfo struct {
	NodeName         string  `json:"node_name,omitempty" description:"name of the k8s node"`
	NodeIp           string  `json:"node_ip,omitempty" description:"ip of the k8s node"`
	NodeStatus       string  `json:"node_status,omitempty" description:"status of the k8s node"`
	NodeLabels       map[string]string  `json:"node_labels,omitempty" description:"labels of the k8s node"`
}

type NodeLabelsInfo struct {
	NodeLabels       map[string]string  `json:"node_labels,omitempty" description:"labels of the k8s node"`
}