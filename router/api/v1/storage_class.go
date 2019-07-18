package v1

import (
	"github.com/emicklei/go-restful"
	"fmt"
	"WarpCloud/walm/pkg/k8s/adaptor"
	"WarpCloud/walm/router/api"
)

func GetStorageClasses(request *restful.Request, response *restful.Response) {
	storageClasses, err := adaptor.GetDefaultAdaptorSet().GetAdaptor("StorageClass").(*adaptor.WalmStorageClassAdaptor).GetWalmStorageClasses("", nil)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to get storageClasses: %s", err.Error()))
		return
	}
	response.WriteEntity(adaptor.WalmStorageClassList{len(storageClasses), storageClasses})
}

func GetStorageClass(request *restful.Request, response *restful.Response) {
	storageClassName := request.PathParameter("name")
	storageClass, err := adaptor.GetDefaultAdaptorSet().GetAdaptor("StorageClass").(*adaptor.WalmStorageClassAdaptor).GetResource("", storageClassName)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to get storageClass: %s", err.Error()))
		return
	}
	if storageClass.GetState().Status == "NotFound" {
		api.WriteNotFoundResponse(response, -1, fmt.Sprintf("storageClass %s is not found", storageClassName))
		return
	}
	response.WriteEntity(storageClass)
}


