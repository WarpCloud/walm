package api

import (
	"github.com/emicklei/go-restful"
	"net/http"
)

func WriteErrorResponse(response *restful.Response, code int, errMsg string) error {
	return response.WriteHeaderAndEntity(http.StatusInternalServerError, ErrorMessageResponse{code, errMsg})
}

func WriteNotFoundResponse(response *restful.Response, code int, errMsg string) error {
	return response.WriteHeaderAndEntity(http.StatusNotFound, ErrorMessageResponse{code, errMsg})
}
