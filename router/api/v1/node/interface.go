package node

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"walm/pkg/node"
	"walm/router/api/util"
	"walm/router/ex"
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


// GetCluster godoc
// @Tags Node
// @Description GetNodeLabels
// @OperationId GetNodeLabels
// @Accept  json
// @Produce  json
// @Param   nodename     path    string     true   "the name of node"
// @Success 200 {array} node.NodeLabelsInfo "OK"
// @Failure 400 {object} ex.ApiResponse "Invalid NodeName supplied!"
// @Failure 404 {object} ex.ApiResponse "node not found"
// @Failure 500 {object} ex.ApiResponse "Server Error"
// @Router /node/{nodename}/labels [get]
func GetNodeLabels(c *gin.Context) {

	if values, err := util.GetPathParams(c, []string{"nodename"}); err != nil {
		c.JSON(ex.ReturnBadRequest())
	} else {
		nodename := values[0]
		nodeLabels, err := node.GetNodeLabels(nodename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
		}

		c.JSON(http.StatusOK, nodeLabels)
	}

}


// GetNodeLabels godoc
// @Tags Node
// @Description UpdateNodeLabels
// @OperationId UpdateNodeLabels
// @Accept  json
// @Produce  json
// @Param   nodename     path    string     true   "the name of node"
// @Param   labels     body   map[string]string    true    "ReleaseRequest of instance"
// @Success 200 {array}  ex.ApiResponse "OK"
// @Failure 400 {object} ex.ApiResponse "Invalid NodeName supplied!"
// @Failure 404 {object} ex.ApiResponse "node not found"
// @Failure 500 {object} ex.ApiResponse "Server Error"
// @Router /node/{nodename}/labels [put]
func UpdateNodeLabels(c *gin.Context) {

	if values, err := util.GetPathParams(c, []string{"nodename"}); err != nil {
		c.JSON(ex.ReturnBadRequest())
	} else {
		var postdata map[string]string

		if err := c.Bind(&postdata); err != nil {
			c.JSON(ex.ReturnBadRequest())
		} else {

			nodename := values[0]
			if err := node.UpdateNodeLabels(nodename, postdata); err != nil {
				c.JSON(ex.ReturnInternalServerError(err))
			} else {
				c.JSON(ex.ReturnOK())
			}

		}
	}

}


// GetNodeLabels godoc
// @Tags Node
// @Description AddNodeLabels
// @OperationId AddNodeLabels
// @Accept  json
// @Produce  json
// @Param   nodename     path    string     true   "the name of node"
// @Param   labels     body   map[string]string    true    "ReleaseRequest of instance"
// @Success 200 {array}  ex.ApiResponse "OK"
// @Failure 400 {object} ex.ApiResponse "Invalid NodeName supplied!"
// @Failure 404 {object} ex.ApiResponse "node not found"
// @Failure 500 {object} ex.ApiResponse "Server Error"
// @Router /node/{nodename}/labels [post]
func AddNodeLabels(c *gin.Context) {

	if values, err := util.GetPathParams(c, []string{"nodename"}); err != nil {
		c.JSON(ex.ReturnBadRequest())

	} else {
		var postdata map[string]string

		if err := c.Bind(&postdata); err != nil {
			c.JSON(ex.ReturnBadRequest())
		} else {

			nodename := values[0]
			if err := node.AddNodeLabels(nodename, postdata); err != nil {
				c.JSON(ex.ReturnInternalServerError(err))
			} else {
				c.JSON(ex.ReturnOK())
			}

		}
	}

}


// DelNodeLabels godoc
// @Tags Node
// @Description DelNodeLabels
// @OperationId DelNodeLabels
// @Accept  json
// @Produce  json
// @Param   nodename     path    string     true   "the name of node"
// @Param   labels     body   map[string]string    true    "ReleaseRequest of instance"
// @Success 200 {array}  ex.ApiResponse "OK"
// @Failure 400 {object} ex.ApiResponse "Invalid NodeName supplied!"
// @Failure 404 {object} ex.ApiResponse "node not found"
// @Failure 500 {object} ex.ApiResponse "Server Error"
// @Router /node/{nodename}/labels [delete]
func DelNodeLabels(c *gin.Context) {

	if values, err := util.GetPathParams(c, []string{"nodename"}); err != nil {
		c.JSON(ex.ReturnBadRequest())

	} else {
		var postdata map[string]string

		if err := c.Bind(&postdata); err != nil {
			c.JSON(ex.ReturnBadRequest())
		} else {

			nodename := values[0]
			if err := node.DelNodeLabels(nodename, postdata); err != nil {
				c.JSON(ex.ReturnInternalServerError(err))
			} else {
				c.JSON(ex.ReturnOK())
			}

		}
	}

}