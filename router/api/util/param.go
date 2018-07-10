package util

import (
	"errors"
	"walm/router/ex"

	"github.com/gin-gonic/gin"
)

func GetPathParams(c *gin.Context, names []string) (values []string, err error) {
	for _, name := range names {
		values = append(values, c.Param(name))
	}
	for _, value := range values {
		if len(value) == 0 {
			err = errors.New("")
			c.JSON(ex.ReturnBadRequest())
			break
		}
	}
	return
}
