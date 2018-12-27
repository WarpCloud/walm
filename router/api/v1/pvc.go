package v1

import (
	"github.com/emicklei/go-restful"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"fmt"
	"walm/pkg/k8s/adaptor"
	"github.com/sirupsen/logrus"
)

func GetPvcs(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	labelSelectorStr := request.QueryParameter("labelselector")
	labelSelector, err := metav1.ParseToLabelSelector(labelSelectorStr)
	if err != nil {
		WriteErrorResponse(response, -1,  fmt.Sprintf("parse label selector failed: %s", err.Error()))
		return
	}
	pvcs, err := adaptor.GetDefaultAdaptorSet().GetAdaptor("PersistentVolumeClaim").
		(*adaptor.WalmPersistentVolumeClaimAdaptor).GetWalmPersistentVolumeClaimAdaptors(namespace, labelSelector)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to get pvcs: %s", err.Error()))
		return
	}
	response.WriteEntity(adaptor.WalmPersistentVolumeClaimList{len(pvcs),pvcs})
}

func GetPvc(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	pvcName := request.PathParameter("pvcname")
	pvc, err := adaptor.GetDefaultAdaptorSet().GetAdaptor("PersistentVolumeClaim").GetResource(namespace, pvcName)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to get pvc: %s", err.Error()))
		return
	}
	if pvc.GetState().Status == "NotFound" {
		WriteNotFoundResponse(response, -1, fmt.Sprintf("pvc %s is not found", pvcName))
		return
	}
	response.WriteEntity(pvc.(adaptor.WalmPersistentVolumeClaim))
}

func DeletePvc(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	pvcName := request.PathParameter("pvcname")
	err := adaptor.GetDefaultAdaptorSet().GetAdaptor("PersistentVolumeClaim").(*adaptor.WalmPersistentVolumeClaimAdaptor).DeletePvc(namespace, pvcName)
	if err != nil {
		if adaptor.IsNotFoundErr(err) {
			logrus.Warnf("pvc %s/%s is not found", namespace, pvcName)
			return
		}
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to delete pvc : %s", err.Error()))
		return
	}
}

func DeletePvcs(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	labelSelectorStr := request.QueryParameter("labelselector")
	labelSelector, err := metav1.ParseToLabelSelector(labelSelectorStr)
	if err != nil {
		WriteErrorResponse(response, -1,  fmt.Sprintf("parse label selector failed: %s", err.Error()))
		return
	}

	pvcAdaptor := adaptor.GetDefaultAdaptorSet().GetAdaptor("PersistentVolumeClaim").
	(*adaptor.WalmPersistentVolumeClaimAdaptor)

	pvcs, err := pvcAdaptor.GetWalmPersistentVolumeClaimAdaptors(namespace, labelSelector)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to get pvcs: %s", err.Error()))
		return
	}

	for _, pvc := range pvcs {
		err = pvcAdaptor.DeletePvc(namespace, pvc.Name)
		if err != nil {
			if adaptor.IsNotFoundErr(err) {
				logrus.Warnf("pvc %s/%s is not found", namespace, pvc.Name)
				continue
			}
			WriteErrorResponse(response, -1, fmt.Sprintf("failed to delete pvc %s: %s", pvc.Name, err.Error()))
			return
		}
	}

}


