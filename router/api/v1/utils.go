package v1

import (
	"github.com/emicklei/go-restful"
	"walm/router/api"
	"net/http"
)

func WriteErrorResponse(response *restful.Response, code int, errMsg string) error {
	return response.WriteHeaderAndEntity(http.StatusInternalServerError, api.ErrorMessageResponse{code, errMsg})
}

func WriteNotFoundResponse(response *restful.Response, code int, errMsg string) error {
	return response.WriteHeaderAndEntity(http.StatusNotFound, api.ErrorMessageResponse{code, errMsg})
}
