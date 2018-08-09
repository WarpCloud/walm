package api

type ErrorMessageResponse struct {
	ErrCode int `json:"errCode"`
	ErrMessage string `json:"errMessage"`
}
