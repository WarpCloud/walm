package api

type ErrorMessageResponse struct {
	ErrCode    int    `json:"errCode"`
	ErrMessage string `json:"errMessage"`
}

type LabelNodeRequestBody struct {
	AddLabels    map[string]string `json:"add_labels"`
	RemoveLabels []string `json:"remove_labels"`
}
