package cluster

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gosuri/uitable"

	"walm/models"
	helm "walm/pkg/helm"
	. "walm/pkg/util/log"
	"walm/router/api/v1/instance"
	"walm/router/ex"
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

	name := c.Param("name")
	if len(name) == 0 {
		c.JSON(ex.ReturnBadRequest())
		return
	}
	namespace := c.Param("namespace")
	if len(namespace) == 0 {
		c.JSON(ex.ReturnBadRequest())
		return
	}

	var postdata Cluster
	if err := c.Bind(&postdata); err != nil {
		c.JSON(ex.ReturnBadRequest())
		return
	} else {
		if len(postdata.Apps) > 0 {

		} else {
			c.JSON(ex.ReturnBadRequest())
			return
		}
	}

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
	//ListApplications
	name := c.Param("name")
	if len(name) == 0 {
		c.JSON(ex.ReturnBadRequest())
		return
	}
	namespace := c.Param("namespace")
	if len(namespace) == 0 {
		c.JSON(ex.ReturnBadRequest())
		return
	}

	info := Info{
		Name:  name,
		Infos: []instance.Info{},
	}

	if releases, err := models.GetReleasesOfCluster(name); err != nil {
		c.JSON(ex.ReturnInternalServerError(err))
		return
	} else {
		Log.Infof("begin to get status of cluser %s; namespace %s; releaes: %s", name, namespace, strings.Join(releases, ","))
		var errs []error

		for _, release := range releases {
			if table, err := helm.Helm.ListApplications([]string{release}, []string{"--namespace", namespace, "--all"}); err != nil {
				errs = append(errs, err)
			} else {
				for _, line := range strings.Split(table.String(), "/n") {
					values := strings.Split(line, uitable.Separator)
					info.Infos = append(info.Infos, instance.Info{
						Name:      values[0],
						Revision:  values[1],
						Updated:   values[2],
						Status:    values[3],
						Chart:     values[4],
						Namespace: values[5],
					})
				}
			}
		}
		if len(errs) > 0 {
			c.JSON(ex.ReturnInternalServerErrors(errs))
			return
		}
	}
	c.JSON(http.StatusOK, info)
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
	name := c.Param("name")
	if len(name) == 0 {
		c.JSON(ex.ReturnBadRequest())
		return
	}
	namespace := c.Param("namespace")
	if len(namespace) == 0 {
		c.JSON(ex.ReturnBadRequest())
		return
	}
	if releases, err := models.GetReleasesOfCluster(name); err != nil {
		c.JSON(ex.ReturnInternalServerError(err))
		return
	} else {
		Log.Infof("begin to delete cluser %s; namespace %s; releaes: %s", name, namespace, strings.Join(releases, ","))
		var errs []error
		for _, release := range releases {
			if err := helm.Helm.Detele([]string{release}, []string{}); err != nil {
				errs = append(errs, err)
			} else {
				if err := models.DeleteAppInst(release); err != nil {
					errs = append(errs, err)
				}
			}
		}
		if len(errs) > 0 {
			c.JSON(ex.ReturnInternalServerErrors(errs))
			return
		}
	}

	if err := models.DeleteCluster(name); err != nil {
		c.JSON(ex.ReturnInternalServerError(err))
		return
	}

	c.JSON(ex.ReturnOK())

}




