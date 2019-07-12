package converter

import (
	corev1 "k8s.io/api/core/v1"
	"WarpCloud/walm/pkg/models/k8s"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/resource"
	"encoding/json"
	"WarpCloud/walm/pkg/k8s/utils"
)

func ConvertNodeFromK8s(oriNode *corev1.Node, podsOnNode *corev1.PodList) (walmNode *k8s.Node, err error) {
	if oriNode == nil {
		return
	}
	node := oriNode.DeepCopy()

	walmNode = &k8s.Node{
		Meta:                 k8s.NewMeta(k8s.NodeKind, node.Namespace, node.Name, buildNodeState(node)),
		NodeIp:               buildNodeIp(node),
		Labels:               node.Labels,
		Annotations:          node.Annotations,
		Capacity:             convertResourceListToMap(node.Status.Capacity),
		Allocatable:          convertResourceListToMap(node.Status.Allocatable),
		WarpDriveStorageList: []k8s.WarpDriveStorage{},
	}

	requestsAllocated, limitsAllocated := getTotalRequestsAndLimits(podsOnNode)
	walmNode.RequestsAllocated = convertResourceListToMap(requestsAllocated)
	walmNode.LimitsAllocated = convertResourceListToMap(limitsAllocated)
	walmNode.UnifyUnitResourceInfo = buildUnifyUnitResourceInfo(walmNode)

	if len(node.Annotations) > 0 {
		poolResourceListStr := node.Annotations["ResourceVolumePoolList"]
		if poolResourceListStr != "" {
			poolResourceList := []k8s.PoolResource{}
			err = json.Unmarshal([]byte(poolResourceListStr), &poolResourceList)
			if err != nil {
				logrus.Warnf("failed to unmarshal pool resource list str %s : %s", poolResourceListStr, err.Error())
			} else {
				for _, poolResource := range poolResourceList {
					warpDriveStorage := k8s.WarpDriveStorage{
						PoolName: poolResource.PoolName,
					}
					for _, subPool := range poolResource.SubPools {
						warpDriveStorage.StorageLeft += subPool.Size - subPool.UsedSize
						warpDriveStorage.StorageTotal += subPool.Size
					}
					walmNode.WarpDriveStorageList = append(walmNode.WarpDriveStorageList, warpDriveStorage)
				}
			}
		}
	}
	return
}

func convertResourceListToMap(resourceList corev1.ResourceList) map[string]string {
	mapResource := make(map[string]string, 0)

	for k, v := range resourceList {
		mapResource[k.String()] = v.String()
	}

	return mapResource
}

func buildUnifyUnitResourceInfo(node *k8s.Node) k8s.UnifyUnitNodeResourceInfo {
	return k8s.UnifyUnitNodeResourceInfo{
		Capacity:          buildNodeResourceInfo(node.Capacity),
		Allocatable:       buildNodeResourceInfo(node.Allocatable),
		LimitsAllocated:   buildNodeResourceInfo(node.LimitsAllocated),
		RequestsAllocated: buildNodeResourceInfo(node.RequestsAllocated),
	}
}

func buildNodeResourceInfo(resourceList map[string]string) k8s.NodeResourceInfo {
	nodeResourceInfo := k8s.NodeResourceInfo{}
	for resourceName, resourceValue := range resourceList {
		if resourceName == corev1.ResourceCPU.String() {
			nodeResourceInfo.Cpu = utils.ParseK8sResourceCpu(resourceValue)
		}
		if resourceName == corev1.ResourceMemory.String() {
			nodeResourceInfo.Memory = utils.ParseK8sResourceMemory(resourceValue)
		}
	}
	return nodeResourceInfo
}

func getTotalRequestsAndLimits(podList *corev1.PodList) (reqs corev1.ResourceList, limits corev1.ResourceList) {
	reqs, limits = map[corev1.ResourceName]resource.Quantity{}, map[corev1.ResourceName]resource.Quantity{}
	for _, pod := range podList.Items {
		podReqs, podLimits := utils.GetPodRequestsAndLimits(pod.Spec)
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

func buildNodeIp(node *corev1.Node) string {
	for _, address := range node.Status.Addresses {
		if address.Type == corev1.NodeInternalIP {
			return address.Address
		}
	}
	return ""
}

func buildNodeState(node *corev1.Node) k8s.State {
	podState := k8s.NewState("NotReady", "Unknown", "")
	for _, condition := range node.Status.Conditions {
		if condition.Type == "Ready" {
			if condition.Status == corev1.ConditionTrue {
				podState = k8s.NewState("Ready", "", "")
			} else {
				podState = k8s.NewState("NotReady", condition.Reason, condition.Message)
			}
			break
		}
	}
	return podState
}
