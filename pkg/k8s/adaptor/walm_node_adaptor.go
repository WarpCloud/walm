package adaptor

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"walm/pkg/k8s/handler"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/resource"
	"encoding/json"
	"walm/pkg/util"
)

type WalmNodeAdaptor struct {
	handler *handler.NodeHandler
}

func (adaptor *WalmNodeAdaptor) GetResource(namespace string, name string) (WalmResource, error) {
	node, err := adaptor.handler.GetNode(name)
	if err != nil {
		if IsNotFoundErr(err) {
			return WalmNode{
				WalmMeta: buildNotFoundWalmMeta("Node", namespace, name),
			}, nil
		}
		return WalmNode{}, err
	}

	return adaptor.BuildWalmNode(*node)
}

func (adaptor *WalmNodeAdaptor) GetWalmNodes(namespace string, labelSelector *metav1.LabelSelector) ([]*WalmNode, error) {
	nodeList, err := adaptor.handler.ListNodes(labelSelector)
	if err != nil {
		return nil, err
	}

	walmNodes := []*WalmNode{}
	if nodeList != nil {
		for _, node := range nodeList {
			walmNode, err := adaptor.BuildWalmNode(*node)
			if err != nil {
				logrus.Errorf("failed to build walm node : %s", err.Error())
				return nil, err
			}
			walmNodes = append(walmNodes, walmNode)
		}
	}

	return walmNodes, nil
}

func convertResourceListToMap(resourceList corev1.ResourceList) map[string]string {
	mapResource := make(map[string]string, 0)

	for k, v := range resourceList {
		mapResource[k.String()] = v.String()
	}

	return mapResource
}

func (adaptor *WalmNodeAdaptor) BuildWalmNode(node corev1.Node) (walmNode *WalmNode, err error) {
	walmNode = &WalmNode{
		WalmMeta:             buildWalmMeta("Node", node.Namespace, node.Name, BuildWalmNodeState(node)),
		NodeIp:               BuildNodeIp(node),
		Labels:               node.Labels,
		Annotations:          node.Annotations,
		Capacity:             convertResourceListToMap(node.Status.Capacity),
		Allocatable:          convertResourceListToMap(node.Status.Allocatable),
		WarpDriveStorageList: []WarpDriveStorage{},
	}
	requestsAllocated, limitsAllocated, err := adaptor.buildAllocated(node)
	if err != nil {
		logrus.Errorf("failed to build node allocated resource : %s", err.Error())
		return
	}

	walmNode.RequestsAllocated = convertResourceListToMap(requestsAllocated)
	walmNode.LimitsAllocated = convertResourceListToMap(limitsAllocated)

	walmNode.UnifyUnitResourceInfo = buildUnifyUnitResourceInfo(walmNode)

	if len(node.Annotations) > 0 {
		poolResourceListStr := node.Annotations["ResourceVolumePoolList"]
		if poolResourceListStr != "" {
			poolResourceList := []PoolResource{}
			err = json.Unmarshal([]byte(poolResourceListStr), &poolResourceList)
			if err != nil {
				logrus.Warnf("failed to unmarshal pool resource list str %s : %s", poolResourceListStr, err.Error())
			} else {
				for _, poolResource := range poolResourceList {
					warpDriveStorage := WarpDriveStorage{
						PoolName: poolResource.PoolName,
					}
					for _, subPool := range poolResource.SubPools {
						warpDriveStorage.StorageLeft += subPool.Size - subPool.UsedSize
					}
					walmNode.WarpDriveStorageList = append(walmNode.WarpDriveStorageList, warpDriveStorage)
				}
			}
		}
	}

	return
}

func (adaptor *WalmNodeAdaptor) buildAllocated(node corev1.Node) (requestsAllocated corev1.ResourceList, limitsAllocated corev1.ResourceList, err error) {
	nonTerminatedPods, err := adaptor.handler.GetPodsOnNode(node.Name, nil)
	if err != nil {
		logrus.Errorf("failed to get non terminated pods on this node : %s", err.Error())
		return
	}
	requestsAllocated, limitsAllocated = getPodsTotalRequestsAndLimits(nonTerminatedPods)
	return
}

func buildUnifyUnitResourceInfo(node *WalmNode) UnifyUnitNodeResourceInfo {
	return UnifyUnitNodeResourceInfo{
		Capacity:          buildNodeResourceInfo(node.Capacity),
		Allocatable:       buildNodeResourceInfo(node.Allocatable),
		LimitsAllocated:   buildNodeResourceInfo(node.LimitsAllocated),
		RequestsAllocated: buildNodeResourceInfo(node.RequestsAllocated),
	}
}

func buildNodeResourceInfo(resourceList map[string]string) NodeResourceInfo {
	nodeResourceInfo := NodeResourceInfo{}
	for resourceName, resourceValue := range resourceList {
		if resourceName == corev1.ResourceCPU.String() {
			quantity, err := resource.ParseQuantity(resourceValue)
			if err != nil {
				logrus.Warnf("failed to parse quantity %s : %s", resourceValue, err.Error())
				continue
			}
			nodeResourceInfo.Cpu = float64(quantity.MilliValue()) / util.K8sResourceCpuScale
		}
		if resourceName == corev1.ResourceMemory.String() {
			quantity, err := resource.ParseQuantity(resourceValue)
			if err != nil {
				logrus.Warnf("failed to parse quantity %s : %s", resourceValue, err.Error())
				continue
			}
			nodeResourceInfo.Memory = quantity.Value() / util.K8sResourceMemoryScale
		}
	}
	return nodeResourceInfo
}

func getPodsTotalRequestsAndLimits(podList *corev1.PodList) (reqs corev1.ResourceList, limits corev1.ResourceList) {
	reqs, limits = map[corev1.ResourceName]resource.Quantity{}, map[corev1.ResourceName]resource.Quantity{}
	for _, pod := range podList.Items {
		podReqs, podLimits := getPodRequestsAndLimits(&pod)
		for podReqName, podReqValue := range podReqs {
			if value, ok := reqs[podReqName]; !ok {
				reqs[podReqName] = *podReqValue.Copy()
			} else {
				value.Add(podReqValue)
				reqs[podReqName] = value
			}
		}
		for podLimitName, podLimitValue := range podLimits {
			if value, ok := limits[podLimitName]; !ok {
				limits[podLimitName] = *podLimitValue.Copy()
			} else {
				value.Add(podLimitValue)
				limits[podLimitName] = value
			}
		}
	}
	return
}

func getPodRequestsAndLimits(pod *corev1.Pod) (reqs map[corev1.ResourceName]resource.Quantity, limits map[corev1.ResourceName]resource.Quantity) {
	reqs, limits = map[corev1.ResourceName]resource.Quantity{}, map[corev1.ResourceName]resource.Quantity{}
	for _, container := range pod.Spec.Containers {
		for name, quantity := range container.Resources.Requests {
			if value, ok := reqs[name]; !ok {
				reqs[name] = *quantity.Copy()
			} else {
				value.Add(quantity)
				reqs[name] = value
			}
		}
		for name, quantity := range container.Resources.Limits {
			if value, ok := limits[name]; !ok {
				limits[name] = *quantity.Copy()
			} else {
				value.Add(quantity)
				limits[name] = value
			}
		}
	}
	// init containers define the minimum of any resource
	for _, container := range pod.Spec.InitContainers {
		for name, quantity := range container.Resources.Requests {
			value, ok := reqs[name]
			if !ok {
				reqs[name] = *quantity.Copy()
				continue
			}
			if quantity.Cmp(value) > 0 {
				reqs[name] = *quantity.Copy()
			}
		}
		for name, quantity := range container.Resources.Limits {
			value, ok := limits[name]
			if !ok {
				limits[name] = *quantity.Copy()
				continue
			}
			if quantity.Cmp(value) > 0 {
				limits[name] = *quantity.Copy()
			}
		}
	}
	return
}

func BuildNodeIp(node corev1.Node) string {
	for _, address := range node.Status.Addresses {
		if address.Type == corev1.NodeInternalIP {
			return address.Address
		}
	}
	return ""
}

func BuildWalmNodeState(node corev1.Node) WalmState {
	podState := buildWalmState("NotReady", "Unknown", "")
	for _, condition := range node.Status.Conditions {
		if condition.Type == "Ready" {
			if condition.Status == corev1.ConditionTrue {
				podState = buildWalmState("Ready", "", "")
			} else {
				podState = buildWalmState("NotReady", condition.Reason, condition.Message)
			}
			break
		}
	}
	return podState
}
