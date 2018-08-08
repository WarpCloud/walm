package node

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"walm/router/api/util"
	"walm/router/ex"
	"walm/pkg/k8s/adaptor"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"errors"
	"walm/pkg/k8s/handler"
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
	if nodeAdaptor, ok := adaptor.GetDefaultAdaptorSet().GetAdaptor("Node").(*adaptor.WalmNodeAdaptor); ok {
		nodes, err := nodeAdaptor.GetWalmNodes("", &metav1.LabelSelector{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, nodes)
		return
	}
	c.JSON(http.StatusInternalServerError, errors.New("failed to get node adaptor"))
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
		nodeName := values[0]
		if nodeAdaptor, ok := adaptor.GetDefaultAdaptorSet().GetAdaptor("Node").(*adaptor.WalmNodeAdaptor); ok {
			node, err := nodeAdaptor.GetResource("", nodeName)
			if err != nil {
				c.JSON(http.StatusInternalServerError, err)
				return
			}
			if walmNode, ok := node.(*adaptor.WalmNode); ok {
				c.JSON(http.StatusOK, walmNode.Labels)
				return
			}
			c.JSON(http.StatusInternalServerError, errors.New("failed to get walm node"))
			return
		}
		c.JSON(http.StatusInternalServerError, errors.New("failed to get node adaptor"))
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
			if _, err := handler.GetDefaultHandlerSet().GetNodeHandler().LabelNode(nodename, postdata, nil); err != nil {
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
			if _, err := handler.GetDefaultHandlerSet().GetNodeHandler().LabelNode(nodename, postdata, nil); err != nil {
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
// @Param   labels     body   []string    true    "ReleaseRequest of instance"
// @Success 200 {array}  ex.ApiResponse "OK"
// @Failure 400 {object} ex.ApiResponse "Invalid NodeName supplied!"
// @Failure 404 {object} ex.ApiResponse "node not found"
// @Failure 500 {object} ex.ApiResponse "Server Error"
// @Router /node/{nodename}/labels [delete]
func DelNodeLabels(c *gin.Context) {

	if values, err := util.GetPathParams(c, []string{"nodename"}); err != nil {
		c.JSON(ex.ReturnBadRequest())

	} else {
		var postdata []string

		if err := c.Bind(&postdata); err != nil {
			c.JSON(ex.ReturnBadRequest())
		} else {

			nodename := values[0]
			if _, err := handler.GetDefaultHandlerSet().GetNodeHandler().LabelNode(nodename, nil, postdata); err != nil {
				c.JSON(ex.ReturnInternalServerError(err))
			} else {
				c.JSON(ex.ReturnOK())
			}

		}
	}

}
