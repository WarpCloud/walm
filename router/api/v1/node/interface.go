package node

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"walm/pkg/node"
)

// GetEdgeServer godoc
// @Tags Node
// @Description Get Node Info
// @OperationId GetNode
// @Produce  json
// @Success 200 {array} node.NodeInfo "OK"
// @Failure 400 {object} ex.ApiResponse "Invalid Name supplied!"
// @Failure 404 {object} ex.ApiResponse "node not found"
// @Failure 500 {object} ex.ApiResponse "Server Error"
// @Router /node [get]
func GetNode(c *gin.Context) {

	nodes, err := node.GetNode()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}

	c.JSON(http.StatusOK, nodes)

}