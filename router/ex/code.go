package ex

import "net/http"

const (
	SUCCESS        = 200
	INTERNAL_ERROR = http.StatusInternalServerError
	INVALID_PARAMS = http.StatusBadRequest

	ERROR_AUTH_CHECK_TOKEN_FAIL    = 10001
	ERROR_AUTH_CHECK_TOKEN_TIMEOUT = 10002
	ERROR_AUTH_TOKEN               = 10003
	ERROR_AUTH                     = 10004

	ERROR_CLUSTER_NOT_EXIST = http.StatusNotFound
	ERROR_CLUSTER_EXIST     = http.StatusFound
)
