package cluster

import (
	"github.com/gin-gonic/gin"
)

func DeployCluster(c *gin.Context) {
}

func ListCluster(c *gin.Context) {
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
