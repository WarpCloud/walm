package v1

import (
	"github.com/emicklei/go-restful"
	"walm/router/api"
)

func WriteErrorResponse(response *restful.Response, code int, errMsg string) error {
	return response.WriteHeaderAndEntity(500, api.ErrorMessageResponse{code, errMsg})
}
