package cluster

import (
	"github.com/gin-gonic/gin"
)

// DeployCluster godoc
// @Description Deplou an Cluster
// @OperationId DeployCluster
// @Accept  json
// @Produce  json
// @Param   namespace     path    string     true        "identifier of the namespace"
// @Param   name     path    string     true        "the name of cluster"
// @Param   apps     body   cluster.Cluster    true    "Apps of Cluster"
// @Success 200 {object} ex.ApiResponse "OK"
// @Failure 400 {object} ex.ApiResponse "Invalid Name supplied!"
// @Failure 404 {object} ex.ApiResponse "namespace not found"
// @Failure 500 {object} ex.ApiResponse "Server Error"
// @Router /{namespace}/{name} [post]
func DeployCluster(c *gin.Context) {
}

// StatusCluster godoc
// @Description Get states of an Cluster
// @OperationId StatusCluster
// @Accept  json
// @Produce  json
// @Param   namespace     path    string     true        "identifier of the namespace"
// @Param   name     path    string     true        "the name of cluster"
// @Success 200 {object} ex.ApiResponse "OK"
// @Failure 400 {object} ex.ApiResponse "Invalid Name supplied!"
// @Failure 404 {object} ex.ApiResponse "cluster not found"
// @Failure 500 {object} ex.ApiResponse "Server Error"
// @Router /{namespace}/{name} [get]
func StatusCluster(c *gin.Context) {
}

// DeleteCluster godoc
// @Description Delete an Cluster
// @OperationId DeleteCluster
// @Accept  json
// @Produce  json
// @Param   namespace     path    string     true        "identifier of the namespace"
// @Param   name     path    string     true        "the name of cluster"
// @Success 200 {object} ex.ApiResponse "OK"
// @Failure 400 {object} ex.ApiResponse "Invalid Name supplied!"
// @Failure 404 {object} ex.ApiResponse "cluster not found"
// @Failure 500 {object} ex.ApiResponse "Server Error"
// @Router /{namespace}/{name} [delete]
func DeleteCluster(c *gin.Context) {
}
