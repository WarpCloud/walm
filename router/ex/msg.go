package ex

var MsgFlags = map[int]string{
	SUCCESS:        "ok",
	INTERNAL_ERROR: "Internal Server error",
	INVALID_PARAMS: "Invalid Name supplied!",

	ERROR_AUTH_CHECK_TOKEN_FAIL:    "Token auth failed",
	ERROR_AUTH_CHECK_TOKEN_TIMEOUT: "Token time out",
	ERROR_AUTH_TOKEN:               "Token generate failed",
	ERROR_AUTH:                     "Token error",

	ERROR_CLUSTER_NOT_EXIST: "cluster is not exist",
	ERROR_CLUSTER_EXIST:     "cluster already exist",
}

func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}

	return MsgFlags[INTERNAL_ERROR]
}
