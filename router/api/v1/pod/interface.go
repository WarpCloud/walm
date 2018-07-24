package pod

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"k8s.io/client-go/tools/remotecommand"

	"walm/pkg/k8s/client"
	"walm/router/api/util"
	"walm/router/ex"
)

// ExecShell godoc
// @Tags tenant
// @Description exec shell to prompt with container
// @OperationId ExecShell
// @Accept  json
// @Produce  json
// @Param   namespace     path    string     true      "namespace of the pod"
// @Param   pod     path    string     true      "name of the pod"
// @Param   container     path    string     true      "container of the pod"
// @Param   shell     query    string     true      "shell type to exec"
// @Success 200 {object} TerminalResponse	"ok"
// @Failure 400 {object} ex.ApiResponse "Invalid Name supplied!"
// @Failure 404 {object} ex.ApiResponse "Instance not found"
// @Failure 405 {object} ex.ApiResponse "Invalid input"
// @Failure 500 {object} ex.ApiResponse "Server Error"
// @Router /pod/{namespace}/{pod}/shell/{container} [get]
func ExecShell(c *gin.Context) {
	if values, err := util.GetPathParams(c, []string{"namespace", "pod", "container"}); err != nil {
		c.JSON(ex.ReturnBadRequest())
	} else {
		namespace, podName, containerName := values[0], values[1], values[2]
		shell := c.Query("shell")

		request := map[string]string{
			"namespace": values[0],
			"pod":       values[1],
			"container": values[2],
			"shell":     shell,
		}

		sessionId, err := genTerminalSessionId()
		if err != nil {
			c.JSON(ex.ReturnInternalServerError(err))
			return
		} else {
			terminalSessions[sessionId] = TerminalSession{
				id:       sessionId,
				bound:    make(chan error),
				sizeChan: make(chan remotecommand.TerminalSize),
			}

			go WaitForTerminal(client.GetDefaultClient(), client.GetDefaultRestConfig, request, sessionId)
			//return success
			c.JSON(http.StatusOK, TerminalResponse{Id: sessionId})
		}
	}

}
