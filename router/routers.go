package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"

	trace "github.com/gin-contrib/tracing"
	stdopentracing "github.com/opentracing/opentracing-go"

	_ "walm/docs"
	. "walm/pkg/util/log"
	cluster "walm/router/api/v1/cluster"
	instance "walm/router/api/v1/instance"
	tenant "walm/router/api/v1/tenant"
	"walm/router/ex"
	"walm/router/middleware"
)

// @title Walm
// @version 1.0.0
// @description Warp application lifecycle manager.

// @contact.name bing.han
// @contact.url http://transwarp.io
// @contact.email bing.han@transwarp.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath /walm/api/v1

func InitRouter(oauth, runmode bool) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	//r.Use(gin.RecoveryWithWriter(Log.Out))

	if runmode {
		gin.SetMode(gin.DebugMode)
		Log.SetLevel(logrus.DebugLevel)
		r.Use(gin.LoggerWithWriter(Log.Out))
	} else {
		Log.SetLevel(logrus.InfoLevel)
		//add Prometheus Metric
		p := middleware.NewPrometheus("Walm")
		p.Use(r)
	}

	//enable swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	//add Probe for readiness and liveness
	r.GET("/readiness", readinessProbe)
	r.GET("/liveness", livenessProbe)

	//define api group
	apiv1 := r.Group("/walm/api/v1")
	if oauth {
		apiv1.Use(middleware.JWT())
	}
	if !runmode && middleware.Tracer != nil {
		//add opentracing
		psr := func(spancontext stdopentracing.SpanContext) stdopentracing.StartSpanOption {
			return stdopentracing.ChildOf(spancontext)
		}
		apiv1.Use(trace.SpanFromHeaders(middleware.Tracer, "Walm", psr, false), trace.InjectToHeaders(middleware.Tracer, false))
	}
	{
		//@Tags
		//@Name instance
		//@Description instance lifecycle manager
		instGroup := apiv1.Group("/instance")
		{
			instGroup.DELETE("/namespace/:namespace/name/:appname", instance.DeleteInstance)
			instGroup.POST("/namespace/:namespace/name/:appname", instance.DeployInstance)
			instGroup.GET("/namespace/:namespace/list", instance.ListInstances)
			instGroup.GET("/namespace/:namespace/name/:appname/info", instance.GetInstanceInfo)
			instGroup.GET("/namespace/:namespace/name/:appname/version/:version/rollback", instance.RollBackInstance)
			instGroup.PUT("/namespace/:namespace/name/:appname", instance.UpdateInstance)
		}

		//@Tags
		//@Name cluster
		//@Description cluster lifecycle manager
		clusterGroup := apiv1.Group("/cluster")
		{
			clusterGroup.POST("/namespace/:namespace/name/:name", cluster.DeployCluster)
			clusterGroup.GET("/namespace/:namespace/name/:name", cluster.GetCluster)
			clusterGroup.DELETE("/namespace/:namespace/name/:name", cluster.DeleteCluster)
			clusterGroup.POST("/namespace/:namespace/name/:name/instance", cluster.DeployInstanceInCluster)
			clusterGroup.POST("/namespace/:namespace/name/:name/list", cluster.DeployListInCluster)
		}

		tenantGroup := apiv1.Group("tenant")
		{
			tenantGroup.POST("/", tenant.CreateTenant)
			tenantGroup.GET("/:tenantname", tenant.GetTenant)
			tenantGroup.DELETE("/:tenantname", tenant.DeleteTenant)
			//tenantGroup.GET("/:tenant_name/services_for_tenant", tenant.GetServiceForTenant)
			//tenantGroup.GET("/:tenant_name/services_for_development", tenant.GetServiceForDev)
			//tenantGroup.GET("/:tenant_name/pods/:pod_name/events", tenant.GetEventForPod)
			//tenantGroup.GET("/:tenant_name/pods/:pod_name/log", tenant.GetLogForPod)
			tenantGroup.GET("/<tenantname>/quotas", tenant.GetQuotas)
			tenantGroup.PUT("/<tenantname>/quotas", tenant.UpdateQuotas)
			//tenantGroup.POST("/register_services_to_kong", tenant.RegisterServicesForTenantToKong)

		}

		podGroup := apiv1.Group("pod")
		{
			podGroup.GET("/:namespace/:pod/shell/:container")
		}
	}

	return r
}

func readinessProbe(c *gin.Context) {
	c.JSON(ex.ReturnOK())
}

func livenessProbe(c *gin.Context) {
	c.JSON(ex.ReturnOK())
}
