package http

import (
	"WarpCloud/walm/pkg/k8s"
	"github.com/emicklei/go-restful"
	"WarpCloud/walm/pkg/models/http"
	"github.com/emicklei/go-restful-openapi"
	k8sModel "WarpCloud/walm/pkg/models/k8s"
	"fmt"
	httpUtils "WarpCloud/walm/pkg/util/http"
	errorModel "WarpCloud/walm/pkg/models/error"
)

type PvcHandler struct {
	k8sCache    k8s.Cache
	k8sOperator k8s.Operator
}

func RegisterPvcHandler(k8sCache k8s.Cache, k8sOperator k8s.Operator) *restful.WebService {
	handler := &PvcHandler{
		k8sCache:    k8sCache,
		k8sOperator: k8sOperator,
	}

	ws := new(restful.WebService)

	ws.Path(http.ApiV1 + "/pvc").
		Doc("Kubernetes Pvc相关操作").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	tags := []string{"pvc"}

	ws.Route(ws.GET("/{namespace}").To(handler.GetPvcs).
		Doc("获取Namepace下的Pvc列表").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.QueryParameter("labelselector", "节点标签过滤").DataType("string")).
		Writes(k8sModel.PersistentVolumeClaimList{}).
		Returns(200, "OK", k8sModel.PersistentVolumeClaimList{}).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.GET("/{namespace}/name/{pvcname}").To(handler.GetPvc).
		Doc("获取对应Pvc的详细信息").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("pvcname", "pvc名字").DataType("string")).
		Writes(k8sModel.PersistentVolumeClaim{}).
		Returns(200, "OK", k8sModel.PersistentVolumeClaim{}).
		Returns(404, "Not Found", http.ErrorMessageResponse{}).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.DELETE("/{namespace}/name/{pvcname}").To(handler.DeletePvc).
		Doc("删除一个Pvc").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("pvcname", "Pvc名字").DataType("string")).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.DELETE("/{namespace}").To(handler.DeletePvcs).
		Doc("删除namespace下满足labelselector的Pvc列表").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.QueryParameter("labelselector", "pvc标签过滤").Required(true).DataType("string")).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	return ws
}

func (handler *PvcHandler) GetPvcs(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	labelSelectorStr := request.QueryParameter("labelselector")
	pvcs, err := handler.k8sCache.ListPersistentVolumeClaims(namespace, labelSelectorStr)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to get pvcs: %s", err.Error()))
		return
	}
	response.WriteEntity(k8sModel.PersistentVolumeClaimList{len(pvcs), pvcs})
}

func (handler *PvcHandler) GetPvc(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	pvcName := request.PathParameter("pvcname")
	resource, err := handler.k8sCache.GetResource(k8sModel.PersistentVolumeClaimKind, namespace, pvcName)
	if err != nil {
		if errorModel.IsNotFoundError(err) {
			httpUtils.WriteNotFoundResponse(response, -1, fmt.Sprintf("pvc %s is not found", pvcName))
			return
		}
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to get pvc: %s", err.Error()))
		return
	}

	response.WriteEntity(resource.(*k8sModel.PersistentVolumeClaim))
}

func (handler *PvcHandler) DeletePvc(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	pvcName := request.PathParameter("pvcname")
	err := handler.k8sOperator.DeletePvc(namespace, pvcName)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to delete pvc : %s", err.Error()))
		return
	}
}

func (handler *PvcHandler) DeletePvcs(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	labelSelectorStr := request.QueryParameter("labelselector")

	err := handler.k8sOperator.DeletePvcs(namespace, labelSelectorStr)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to delete pvcs: %s",  err.Error()))
		return
	}
}
