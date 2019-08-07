package impl

const (
	KeyNotFoundErrMsg = "redis: nil"
)

func isKeyNotFoundError(err error) bool {
	if err.Error() == KeyNotFoundErrMsg {
		return true
	}
	return false
}

func buildHScanFilter(namespace string) string {
	return namespace + "/*"
}
