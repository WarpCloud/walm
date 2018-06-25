package ex

var MsgFlags = map[int]string{
	SUCCESS:        "ok",
	INTERNAL_ERROR: "Internal Server error",
	INVALID_PARAMS: "Invalid Name supplied!",

	ERROR_AUTH_CHECK_TOKEN_FAIL:    "Token鉴权失败",
	ERROR_AUTH_CHECK_TOKEN_TIMEOUT: "Token已超时",
	ERROR_AUTH_TOKEN:               "Token生成失败",
	ERROR_AUTH:                     "Token错误",
}

func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}

	return MsgFlags[INTERNAL_ERROR]
}
