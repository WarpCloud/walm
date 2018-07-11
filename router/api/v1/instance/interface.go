package instance

import (
	"errors"
	"net/http"
	"strconv"

	"walm/pkg/helm"
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

// DeleteInstance godoc
// @Tags instance
// @Description Delete an Appliation
// @OperationId DeleteInstance
// @Accept  json
// @Produce  json
// @Param   namespace     path    string     true      "identifier of the instance"
// @Param   appname     path    string     true        "identifier of the instance"
// @Success 200 {object} ex.ApiResponse "OK"
// @Failure 400 {object} ex.ApiResponse "Invalid Name supplied!"
// @Failure 404 {object} ex.ApiResponse "Instance not found"
// @Failure 500 {object} ex.ApiResponse "Server Error"
// @Router /instance/namespace/{namespace}/name/{appname} [delete]
func DeleteInstance(c *gin.Context) {

	if values, err := GetPathParams(c, []string{"namespace", "appname"}); err != nil {
		return
	} else {
		if err := helm.DeleteRealese(values[0], values[1]); err != nil {
			c.JSON(ex.ReturnInternalServerError(err))
			return
		}
	}
	c.JSON(ex.ReturnOK())

}

// Deploy godoc
// @Tags instance
// @Description deploy an instance with the givin data
// @OperationId Deploy
// @Accept  json
// @Produce  json
// @Param   namespace     path    string     true      "identifier of the instance"
// @Param   instance     path    string     true      "identifier of the instance"
// @Param   instance     body   instance.Instance    true    "Update instance"
// @Success 200 {object} ex.ApiResponse "OK"
// @Failure 400 {object} ex.ApiResponse "Invalid Name supplied!"
// @Failure 405 {object} ex.ApiResponse "Invalid input"
// @Failure 500 {object} ex.ApiResponse "Server Error"
// @Router /instance/namespace/:namespace/name/:appname [post]
func DeployInstance(c *gin.Context) {

	if _, err := GetPathParams(c, []string{"namespace", "appname"}); err != nil {
		return
	}
	var postdata helm.ReleaseRequest
	if err := c.Bind(&postdata); err != nil {
		c.JSON(ex.ReturnBadRequest())
		return
	} else {
		if err := helm.InstallUpgradeRealese(postdata); err != nil {
			c.JSON(ex.ReturnInternalServerError(err))
			return
		} else {
			c.JSON(ex.ReturnOK())
		}
	}
}

// FindInstancesStatus godoc
// @Tags instance
// @Description find the aim instance status
// @OperationId FindInstancesStatus
// @Accept  json
// @Produce  json
// @Param   namespace     path    string     true      "identifier of the instance"
// @Param   appname     path    string     true      "identifier of the appname"
// @Param   max     query    int     false      "max num to display"
// @Param   offset     query    int     false      "the offset of result"
// @Success 200 {object} []helm.ReleaseInfo	"ok"
// @Failure 404 {object} ex.ApiResponse "Invalid status not found"
// @Failure 500 {object} ex.ApiResponse "Server Error"
// @Router /instance/namespace/:namespace/list [get]
func ListInstances(c *gin.Context) {

	namespace := c.Param("namespace")
	if len(namespace) == 0 {
		c.JSON(ex.ReturnBadRequest())
		return
	}

	var imax, ioffset int

	max := c.Query("max")
	if len(max) > 0 {
		var err error
		if imax, err = strconv.Atoi(max); err != nil {
			imax = -1
		}
	}

	offset := c.Query("offset")
	if len(offset) > 0 {
		var err error
		if ioffset, err = strconv.Atoi(offset); err != nil {
			ioffset = 0
		}
	}

	if releases, err := helm.ListReleases(namespace); err != nil {
		c.JSON(ex.ReturnInternalServerError(err))
	} else {
		ilen := imax + ioffset
		if ilen > ioffset && ilen < len(releases) {
			c.JSON(http.StatusOK, releases[ioffset:ilen])
		} else {
			c.JSON(http.StatusOK, releases[ioffset:])
		}
	}

}

// GetInstancebyName godoc
// @Tags instance
// @Description Get an Appliation by name with status
// @OperationId GetInstancebyName
// @Accept  json
// @Produce  json
// @Param   namespace     path    string     true      "identifier of the instance"
// @Param   appname     path    string     true        "identifier of the application"
// @Success 200 {object} helm.ReleaseInfo	"ok"
// @Failure 400 {object} ex.ApiResponse "Invalid Name supplied!"
// @Failure 404 {object} ex.ApiResponse "Instance not found"
// @Failure 405 {object} ex.ApiResponse "Invalid input"
// @Failure 500 {object} ex.ApiResponse "Server Error"
// @Router /instance/namespace/{namespace}/name/{appname}/info [get]
func GetInstanceInfo(c *gin.Context) {

	if values, err := GetPathParams(c, []string{"namespace", "appname"}); err != nil {
		return
	} else {
		if release, err := helm.GetReleaseInfo(values[0], values[1]); err != nil {
			c.JSON(ex.ReturnInternalServerError(err))
		} else {
			c.JSON(http.StatusOK, release)
		}
	}
}

// RollBackInstance godoc
// @Tags instance
// @Description Rollback Instance to aim version
// @OperationId RollBackInstance
// @Accept  json
// @Produce  json
// @Param   namespace     path    string     true      "identifier of the instance"
// @Param   appname     path    string     true        "identifier of the instance"
// @Param   version     path    string     true        "identifier of the version"
// @Success 200 {object} instance.Info	"ok"
// @Failure 400 {object} ex.ApiResponse "Invalid Name supplied!"
// @Failure 404 {object} ex.ApiResponse "Instance not found"
// @Failure 405 {object} ex.ApiResponse "Invalid input"
// @Failure 500 {object} ex.ApiResponse "Server Error"
// @Router /instance/namespace/{namespace}/name/{appname}/version/{version}/rollback [get]
func RollBackInstance(c *gin.Context) {
	if values, err := GetPathParams(c, []string{"namespace", "appname", "version"}); err != nil {
		return
	} else {
		if err := helm.RollbackRealese(values[0], values[1], values[2]); err != nil {
			c.JSON(ex.ReturnInternalServerError(err))
			return
		}
	}
	c.JSON(ex.ReturnOK())
}

// UpdateInstance godoc
// @Tags instance
// @Description Update an Appliation
// @OperationId UpdateInstance
// @Accept  json
// @Produce  json
// @Param   namespace     path    string     true      "namespace of the instance"
// @Param   appname     path    string     true      "identifier of the instance"
// @Param   instance     body   helm.ReleaseRequest    true    "Update instance"
// @Success 200 {object} ex.ApiResponse "OK"
// @Failure 400 {object} ex.ApiResponse "Invalid Name supplied!"
// @Failure 404 {object} ex.ApiResponse "Instance not found"
// @Failure 405 {object} ex.ApiResponse "Invalid input"
// @Failure 500 {object} ex.ApiResponse "Server Error"
// @Router /instance/namespace/:namespace/name/:appname [put]
func UpdateInstance(c *gin.Context) {

	if _, err := GetPathParams(c, []string{"namespace", "appname"}); err != nil {
		return
	}
	var postdata helm.ReleaseRequest
	if err := c.Bind(&postdata); err != nil {
		c.JSON(ex.ReturnBadRequest())
		return
	} else {
		if err := helm.PatchUpgradeRealese(postdata); err != nil {
			c.JSON(ex.ReturnInternalServerError(err))
			return
		} else {
			c.JSON(ex.ReturnOK())
		}
	}

}
