package v1

import (
	"github.com/emicklei/go-restful"
	"WarpCloud/walm/pkg/k8s/adaptor"
	"fmt"
	"WarpCloud/walm/router/api"
	"WarpCloud/walm/pkg/k8s/handler"
	"github.com/sirupsen/logrus"
)

func GetSecret(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("secretname")
	secret, err := adaptor.GetDefaultAdaptorSet().GetAdaptor("Secret").GetResource(namespace, name)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to get secret %s/%s: %s", namespace, name, err.Error()))
		return
	}
	if secret.GetState().Status == "NotFound" {
		api.WriteNotFoundResponse(response, -1, fmt.Sprintf("secret %s/%s is not found",namespace, name))
		return
	}
	response.WriteEntity(secret)
}

func GetSecrets(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	secrets, err := adaptor.GetDefaultAdaptorSet().GetAdaptor("Secret").(*adaptor.WalmSecretAdaptor).ListSecrets(namespace, nil)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to list secrets under %s: %s", namespace, err.Error()))
		return
	}
	response.WriteEntity(secrets)
}

func CreateSecret(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	createSecretRequestBody := &api.CreateSecretRequestBody{}
	err := request.ReadEntity(createSecretRequestBody)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read request body: %s", err.Error()))
		return
	}
	walmSecret := &adaptor.WalmSecret{
		Type: createSecretRequestBody.Type,
		Data: createSecretRequestBody.Data,
		WalmMeta: adaptor.WalmMeta{
			Namespace: namespace,
			Name: createSecretRequestBody.Name,
		},
	}

	err = adaptor.GetDefaultAdaptorSet().GetAdaptor("Secret").(*adaptor.WalmSecretAdaptor).CreateSecret(walmSecret)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to create secret : %s", err.Error()))
		return
	}
}

func UpdateSecret(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	createSecretRequestBody := &api.CreateSecretRequestBody{}
	err := request.ReadEntity(createSecretRequestBody)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read request body: %s", err.Error()))
		return
	}

	walmSecret := &adaptor.WalmSecret{
		Type: createSecretRequestBody.Type,
		Data: createSecretRequestBody.Data,
		WalmMeta: adaptor.WalmMeta{
			Namespace: namespace,
			Name: createSecretRequestBody.Name,
		},
	}
	err = adaptor.GetDefaultAdaptorSet().GetAdaptor("Secret").(*adaptor.WalmSecretAdaptor).UpdateSecret(walmSecret)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to update secret : %s", err.Error()))
		return
	}
}

func DeleteSecret(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("secretname")
	err := handler.GetDefaultHandlerSet().GetSecretHandler().DeleteSecret(namespace, name)
	if err != nil {
		if adaptor.IsNotFoundErr(err) {
			logrus.Warnf("secret %s/%s is not found", namespace, name)
			return
		}
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to delete secret : %s", err.Error()))
		return
	}
}
