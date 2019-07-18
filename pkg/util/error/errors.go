package error

type NotFoundError struct {}

func (err NotFoundError) Error() string {
	return "not found error"
}

func IsNotFoundError(err error) bool {
	_, ok := err.(NotFoundError)
	return ok
}