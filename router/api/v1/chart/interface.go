package chart

import (
	"walm/router/api/util"
	"walm/router/ex"
	"github.com/gin-gonic/gin"
	"net/http"
	"walm/pkg/release/manager/helm"
	"walm/pkg/release"
)

// Deploy godoc
// @Tags Chart
// @Description ValidateChart
// @OperationId ValidateChart
// @Accept  json
// @Produce  json
// @Param   chartname     path    string     true      "chart name, eg: stable/mysql"
// @Param   chartversion     path    string   false      "chart version, default latest version, eg: 5.2.0"
// @Param   values     body   helm.ReleaseRequest    true    "ReleaseRequest of instance"
// @Success 200 {object} helm.ChartValicationInfo	"OK"
// @Failure 400 {object} ex.ApiResponse "Invalid Name supplied!"
// @Failure 405 {object} ex.ApiResponse "Invalid input"
// @Failure 500 {object} ex.ApiResponse "Server Error"
// @Router /chart/{chartname}/version/{chartversion}/validation [post]
func ValidateChart(c *gin.Context) {

	if _, err := util.GetPathParams(c, []string{"chartname", "chartversion"}); err != nil {
		c.JSON(ex.ReturnBadRequest())
	} else {
		var postdata release.ReleaseRequest
		if err := c.Bind(&postdata); err != nil {
			c.JSON(ex.ReturnBadRequest())

		} else {

			var chartValicationInfo release.ChartValicationInfo

			if chartValicationInfo, err = helm.ValidateChart(postdata); err != nil {
				c.JSON(ex.ReturnInternalServerError(err))
			}
			c.JSON(http.StatusOK, chartValicationInfo)

		}
	}

}