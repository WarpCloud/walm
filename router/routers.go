package router

import (
	"github.com/gin-gonic/gin"
	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"

	_ "walm/docs"
	. "walm/pkg/util/log"
	"walm/router/api/v1"
	"walm/router/ex"
	"walm/router/middleware"
)

func InitRouter(oauth bool, runmode string) *gin.Engine {
	r := gin.New()

	r.Use(gin.RecoveryWithWriter(Log.Out))

	//enable swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	gin.SetMode(runmode)
	if runmode == "debug" {
		r.Use(gin.LoggerWithWriter(Log.Out))
	} else {
		//add Prometheus Metric
		p := middleware.NewPrometheus("Walm-gin")
		p.Use(r)
		//add open tracing

	}

	//add Probe for readiness and liveness
	r.GET("/readiness", readinessProbe)
	r.GET("/liveness", livenessProbe)

	//define api group
	apiv1 := r.Group("/walm/api/v1")
	if oauth {
		apiv1.Use(middleware.JWT())
	}
	{
		apiv1.DELETE("/application/{appName}", v1.DeleteApplication)
		apiv1.POST("/application/{chart}", v1.DeployApplication)
		apiv1.GET("/application/{namespace}/status/{appname}", v1.FindApplicationsStatus)
		apiv1.GET("/application/{appname}", v1.GetApplicationbyName)
		apiv1.GET("/application/{appname}/rollback/{version}", v1.RollBackApplication)
		apiv1.PUT("/application/{chart}", v1.UpdateApplication)
	}

	return r
}

func readinessProbe(c *gin.Context) {
	c.JSON(ex.ReturnOK())
}

func livenessProbe(c *gin.Context) {
	c.JSON(ex.ReturnOK())
}
