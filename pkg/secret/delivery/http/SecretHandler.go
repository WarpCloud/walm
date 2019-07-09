package http

import (
	"WarpCloud/walm/pkg/k8s"
	"github.com/emicklei/go-restful"
	"WarpCloud/walm/pkg/models/http"
	"github.com/emicklei/go-restful-openapi"
	k8sModel "WarpCloud/walm/pkg/models/k8s"
	httpUtils "WarpCloud/walm/pkg/util/http"
	"fmt"
)

type SecretHandler struct {
	k8sCache    k8s.Cache
	k8sOperator k8s.Operator
}

func RegisterSecretHandler(k8sCache k8s.Cache, k8sOperator k8s.Operator) *restful.WebService {
	handler := &SecretHandler{
		k8sCache:    k8sCache,
		k8sOperator: k8sOperator,
	}

	ws := new(restful.WebService)

	ws.Path(http.ApiV1 + "/secret").
		Doc("Kubernetes Secret相关操作").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	tags := []string{"secret"}

	ws.Route(ws.GET("/{namespace}").To(handler.GetSecrets).
		Doc("获取Namepace下的所有Secret列表").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Writes(k8sModel.SecretList{}).
		Returns(200, "OK", k8sModel.SecretList{}).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.GET("/{namespace}/name/{secretname}").To(handler.GetSecret).
		Doc("获取对应Secret的详细信息").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("secretname", "secret名字").DataType("string")).
		Writes(k8sModel.Secret{}).
		Returns(200, "OK", k8sModel.Secret{}).
		Returns(404, "Not Found", http.ErrorMessageResponse{}).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.DELETE("/{namespace}/name/{secretname}").To(handler.DeleteSecret).
		Doc("删除一个Secret").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("secretname", "Secret名字").DataType("string")).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.POST("/{namespace}").To(handler.CreateSecret).
		Doc("创建一个Secret").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Reads(k8sModel.CreateSecretRequestBody{}).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.PUT("/{namespace}").To(handler.UpdateSecret).
		Doc("更新一个Secret").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Reads(k8sModel.CreateSecretRequestBody{}).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	return ws
}

func (handler *SecretHandler)GetSecret(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("secretname")
	secret, err := handler.k8sCache.GetResource(k8sModel.SecretKind,namespace, name)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to get secret %s/%s: %s", namespace, name, err.Error()))
		return
	}
	if secret.GetState().Status == "NotFound" {
		httpUtils.WriteNotFoundResponse(response, -1, fmt.Sprintf("secret %s/%s is not found",namespace, name))
		return
	}
	response.WriteEntity(secret.(*k8sModel.Secret))
}

func (handler *SecretHandler)GetSecrets(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	secrets, err := handler.k8sCache.ListSecrets(namespace, "")
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to list secrets under %s: %s", namespace, err.Error()))
		return
	}
	response.WriteEntity(secrets)
}

func (handler *SecretHandler)CreateSecret(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	createSecretRequestBody := &k8sModel.CreateSecretRequestBody{}
	err := request.ReadEntity(createSecretRequestBody)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read request body: %s", err.Error()))
		return
	}

	err = handler.k8sOperator.CreateSecret(namespace, createSecretRequestBody)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to create secret : %s", err.Error()))
		return
	}
}

func (handler *SecretHandler)UpdateSecret(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	createSecretRequestBody := &k8sModel.CreateSecretRequestBody{}
	err := request.ReadEntity(createSecretRequestBody)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read request body: %s", err.Error()))
		return
	}

	err =  handler.k8sOperator.UpdateSecret(namespace, createSecretRequestBody)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to update secret : %s", err.Error()))
		return
	}
}

func (handler *SecretHandler)DeleteSecret(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("secretname")
	err := handler.k8sOperator.DeleteSecret(namespace, name)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to delete secret : %s", err.Error()))
		return
	}
}