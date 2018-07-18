package tenant

import (
	"net/http"
	"walm/pkg/tenant"
	"walm/router/api/util"
	"walm/router/ex"

	"github.com/gin-gonic/gin"
)

// CreateTenant godoc
// @Tags tenant
// @Description Create an Tenant with spec info
// @OperationId CreateTenant
// @Accept  json
// @Produce  json
// @Param   tenant     body   helm.ReleaseRequest    true    "Request Info of tenant"
// @Success 200 {object} ex.ApiResponse "OK"
// @Failure 400 {object} ex.ApiResponse "Invalid Name supplied!"
// @Failure 404 {object} ex.ApiResponse "Instance not found"
// @Failure 405 {object} ex.ApiResponse "Invalid input"
// @Failure 500 {object} ex.ApiResponse "Server Error"
// @Router /tenant [post]
func CreateTenant(c *gin.Context) {
	/*
		var postdata helm.ReleaseRequest
			if err := c.Bind(&postdata); err != nil {
				c.JSON(ex.ReturnBadRequest())
			} else {
				if err := helm.InstallUpgradeRealese(postdata); err != nil {
					c.JSON(ex.ReturnInternalServerError(err))
				} else {
					c.JSON(ex.ReturnOK())
				}
			}
	*/
}

// GetTenant godoc
// @Tags tenant
// @Description Get an Tenant by name
// @OperationId GetTenant
// @Accept  json
// @Produce  json
// @Param   tenantname     path    string     true      "identifier of the tenant"
// @Success 200 {object} tenant.TenantInfo	"ok"
// @Failure 400 {object} ex.ApiResponse "Invalid Name supplied!"
// @Failure 404 {object} ex.ApiResponse "Instance not found"
// @Failure 405 {object} ex.ApiResponse "Invalid input"
// @Failure 500 {object} ex.ApiResponse "Server Error"
// @Router /tenant/{tenantname} [get]
func GetTenant(c *gin.Context) {
	if values, err := util.GetPathParams(c, []string{"tenantname"}); err != nil {
		c.JSON(ex.ReturnBadRequest())
	} else {
		name := values[0]
		namespace, err := tenant.GetTenantInfo(name)
		if err != nil {
			c.JSON(ex.ReturnInternalServerError(err))
		} else {
			c.JSON(http.StatusOK, namespace)
		}
	}

}

// DeleteTenant godoc
// @Tags tenant
// @Description Delete an Tenant by name
// @OperationId DeleteTenant
// @Accept  json
// @Produce  json
// @Param   tenantname     path    string     true      "identifier of the tenant"
// @Success 200 {object} tenant.TenantInfo	"ok"
// @Failure 400 {object} ex.ApiResponse "Invalid Name supplied!"
// @Failure 404 {object} ex.ApiResponse "Instance not found"
// @Failure 405 {object} ex.ApiResponse "Invalid input"
// @Failure 500 {object} ex.ApiResponse "Server Error"
// @Router /tenant/{tenantname} [delete]
func DeleteTenant(c *gin.Context) {
	if values, err := util.GetPathParams(c, []string{"tenantname"}); err != nil {
		c.JSON(ex.ReturnBadRequest())
	} else {
		name := values[0]
		err := tenant.DeleteTenant(name)
		if err != nil {
			c.JSON(ex.ReturnInternalServerError(err))
		} else {
			c.JSON(ex.ReturnOK())
		}
	}
}

/*
func GetServiceForTenant(c *gin.Context) {

}

func GetServiceForDev(c *gin.Context) {

}

func GetEventForPod(c *gin.Context) {

}

func GetLogForPod(c *gin.Context) {

}
*/

func GetQuotas(c *gin.Context) {

}

func UpdateQuotas(c *gin.Context) {

}

/*
func RegisterServicesForTenantToKong(c *gin.Context) {

}
*/
