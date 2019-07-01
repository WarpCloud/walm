package http

const (
	ApiV1 = "/api/v1"
	ApiV2 = "/api/v2"
)

type ErrorMessageResponse struct {
	ErrCode    int    `json:"errCode"`
	ErrMessage string `json:"errMessage"`
}
