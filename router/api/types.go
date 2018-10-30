package api

import "k8s.io/api/core/v1"

type ErrorMessageResponse struct {
	ErrCode    int    `json:"err_code"`
	ErrMessage string `json:"err_message"`
}

type LabelNodeRequestBody struct {
	AddLabels    map[string]string `json:"add_labels"`
	RemoveLabels []string          `json:"remove_labels"`
}

type AnnotateNodeRequestBody struct {
	AddAnnotations    map[string]string `json:"add_annotations"`
	RemoveAnnotations []string          `json:"remove_annotations"`
}

type CreateSecretRequestBody struct {
	Data map[string]string `json:"data" description:"secret data"`
	Type v1.SecretType     `json:"type" description:"secret type"`
	Name string            `json:"name" description:"resource name"`
}
