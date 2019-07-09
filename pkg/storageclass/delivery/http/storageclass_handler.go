package http

import (
	"WarpCloud/walm/pkg/k8s"
	"github.com/emicklei/go-restful"
	"WarpCloud/walm/pkg/models/http"
	k8sModel "WarpCloud/walm/pkg/models/k8s"
	"github.com/emicklei/go-restful-openapi"
	"fmt"
	httpUtils "WarpCloud/walm/pkg/util/http"
	errorModel "WarpCloud/walm/pkg/models/error"
)

type StorageClassHandler struct {
	k8sCache k8s.Cache
}

func RegisterStorageClassHandler(k8sCache k8s.Cache) *restful.WebService {
	handler := &StorageClassHandler{
		k8sCache: k8sCache,
	}

	ws := new(restful.WebService)

	ws.Path(http.ApiV1 + "/storageclass").
		Doc("Kubernetes StorageClass相关操作").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	tags := []string{"storageclass"}

	ws.Route(ws.GET("/").To(handler.GetStorageClasses).
		Doc("获取StorageClass列表").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(k8sModel.StorageClassList{}).
		Returns(200, "OK", k8sModel.StorageClassList{}).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.GET("/{name}").To(handler.GetStorageClass).
		Doc("获取StorageClass详细信息").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("name", "StorageClass名字").DataType("string")).
		Writes(k8sModel.StorageClass{}).
		Returns(200, "OK", k8sModel.StorageClass{}).
		Returns(404, "Not Found", http.ErrorMessageResponse{}).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	return ws
}

func (handler *StorageClassHandler) GetStorageClasses(request *restful.Request, response *restful.Response) {
	storageClasses, err := handler.k8sCache.ListStorageClasses("", "")
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to get storageClasses: %s", err.Error()))
		return
	}
	response.WriteEntity(k8sModel.StorageClassList{len(storageClasses), storageClasses})
}

func (handler *StorageClassHandler) GetStorageClass(request *restful.Request, response *restful.Response) {
	storageClassName := request.PathParameter("name")
	storageClass, err := handler.k8sCache.GetResource(k8sModel.StorageClassKind, "", storageClassName)
	if err != nil {
		if errorModel.IsNotFoundError(err) {
			httpUtils.WriteNotFoundResponse(response, -1, fmt.Sprintf("storageClass %s is not found", storageClassName))
			return
		}
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to get storageClass: %s", err.Error()))
		return
	}

	response.WriteEntity(storageClass.(*k8sModel.StorageClass))
}
