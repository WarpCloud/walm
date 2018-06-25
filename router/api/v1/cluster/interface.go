package cluster

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gosuri/uitable"

	"walm/models"
	. "walm/pkg/util/log"
	"walm/router/api/v1/instance"
	"walm/router/ex"
)

// DeployCluster godoc
// @Tags Cluster
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
// @Router /cluster/{namespace}/{name} [post]
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
	if err := c.BindJSON(&postdata); err != nil {
		c.JSON(ex.ReturnBadRequest())
		return
	} else {

		cluster := &models.Cluster{
			Name:       name,
			Namespace:  namespace,
			ConfigTemp: postdata.Conf,
		}

		if err := models.InsertCluster(cluster); err != nil {
			c.JSON(ex.ReturnInternalServerError(err))
			return
		}

		if len(postdata.Apps) > 0 {
			if err, apps := getGragh(cluster.ClusterId, name, namespace, &postdata); err != nil {
				c.JSON(ex.ReturnInternalServerError(err))
				return
			} else {
				for _, app := range apps {
					if err := deployInstance(cluster, app); err != nil {
						c.JSON(ex.ReturnInternalServerError(err))
					}
				}
			}
		} else {
			c.JSON(ex.ReturnBadRequest())
			return
		}
	}

}

func deployInstance(cluster *models.Cluster, app instance.Application) error {
	var args []string
	var flags []string

	if len(app.Name) > 0 {
		args = append(args, app.Name)
		args = append(args, app.Chart)
	} else {
		app.Name = fmt.Sprintf("%s-%s-%d", app.Chart, cluster.Name, cluster.ClusterId)
	}
	if len(app.Namespace) > 0 {
		flags = append(flags, "--namespace")
		flags = append(flags, app.Namespace)
	}
	if len(app.Repo) > 0 {
		flags = append(flags, "--repo")
		flags = append(flags, app.Repo)
	}
	if len(app.Version) > 0 {
		flags = append(flags, "--version")
		flags = append(flags, app.Version)
	}

	if len(app.Links) > 0 {
		for _, v := range app.Links {
			flags = append(flags, "--link")
			flags = append(flags, v)
		}
	}

	appInst := models.AppInst{
		Name:        app.Name,
		Namespace:   app.Namespace,
		AppPkg:      app.Chart,
		Vers:        app.Version,
		ConfigTemp:  app.Value,
		Status:      "deployed",
		ClusterId:   cluster.ClusterId,
		InstallTime: time.Now().Unix(),
	}

	if err := instance.WalmInst.DeplyApplications(args, flags); err != nil {
		return err
	}

	ti := time.Now().Unix()
	appInst.InstalledTime, appInst.LastTime = ti, ti

	if err := models.InsertAppInst(appInst); err != nil {
		Log.Errorf("occu error when insert into AppInst: %s \n", err)

		return err
	}

	return nil
}

// StatusCluster godoc
// @Tags Cluster
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
// @Router /cluster/{namespace}/{name} [get]
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
			if table, err := instance.WalmInst.ListApplications([]string{release}, []string{"--namespace", namespace, "--all"}); err != nil {
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
			c.JSON(ex.ReturnInternalServerError(errs[0]))
			return
		}
	}
	c.JSON(http.StatusOK, info)
}

// DeleteCluster godoc
// @Tags Cluster
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
// @Router /cluster/{namespace}/{name} [delete]
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
			if err := instance.WalmInst.Detele([]string{release}, []string{}); err != nil {
				errs = append(errs, err)
			} else {
				if err := models.DeleteAppInst(release); err != nil {
					errs = append(errs, err)
				}
			}
		}
		if len(errs) > 0 {
			c.JSON(ex.ReturnInternalServerError(errs[0]))
			return
		}
	}

	if err := models.DeleteCluster(name); err != nil {
		c.JSON(ex.ReturnInternalServerError(err))
		return
	}

	c.JSON(ex.ReturnOK())

}
