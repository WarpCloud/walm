package api

import "k8s.io/api/core/v1"

type ErrorMessageResponse struct {
	ErrCode    int    `json:"errCode"`
	ErrMessage string `json:"errMessage"`
}

type LabelNodeRequestBody struct {
	AddLabels    map[string]string `json:"addLabels"`
	RemoveLabels []string          `json:"removeLabels"`
}

type AnnotateNodeRequestBody struct {
	AddAnnotations    map[string]string `json:"addAnnotations"`
	RemoveAnnotations []string          `json:"removeAnnotations"`
}

type CreateSecretRequestBody struct {
	Data map[string]string `json:"data" description:"secret data"`
	Type v1.SecretType     `json:"type" description:"secret type"`
	Name string            `json:"name" description:"resource name"`
}
