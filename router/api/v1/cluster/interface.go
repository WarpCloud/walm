package cluster

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"walm/router/api/util"

	"walm/router/ex"
	"walm/pkg/release"
	"walm/pkg/release/manager/helm"
)

// DeployInstanceInCluster godoc
// @Tags Cluster
// @Description Deploy an Instance into Cluster
// @OperationId DeployInstanceInCluster
// @Accept  json
// @Produce  json
// @Param   namespace     path    string     true        "identifier of the namespace"
// @Param   name     path    string     true        "the name of cluster"
// @Param   apps     body   helm.ReleaseRequest    true    "Apps of Cluster"
// @Success 200 {object} ex.ApiResponse "OK"
// @Failure 400 {object} ex.ApiResponse "Invalid Name supplied!"
// @Failure 404 {object} ex.ApiResponse "namespace not found"
// @Failure 500 {object} ex.ApiResponse "Server Error"
// @Router /cluster/namespace/{namespace}/name/{name}/instance [post]
func DeployInstanceInCluster(c *gin.Context) {
	var namespace, name string
	if values, err := util.GetPathParams(c, []string{"namespace", "name"}); err != nil {
		return
	} else {
		namespace, name = values[0], values[1]
	}

	var postdata release.ReleaseRequest
	if err := c.BindJSON(&postdata); err != nil {
		c.JSON(ex.ReturnBadRequest())
		return
	} else {

		if err, releaseMap := getReleasMap(namespace, name); err != nil {
			c.JSON(ex.ReturnInternalServerError(err))
			return
		} else {
			if err, apps := getGraghForInstance(namespace, name, releaseMap, &postdata); err != nil {
				c.JSON(ex.ReturnInternalServerError(err))
				return
			} else {
				for _, app := range apps {
					if err := deployInstance(namespace, name, postdata.ConfigValues, app); err != nil {
						c.JSON(ex.ReturnInternalServerError(err))
					}
				}
			}
		}
		c.JSON(ex.ReturnOK())
	}
}

// DeployListInCluster godoc
// @Tags Cluster
// @Description Deploy an Instance list into Cluster
// @OperationId DeployListInCluster
// @Accept  json
// @Produce  json
// @Param   namespace     path    string     true        "identifier of the namespace"
// @Param   name     path    string     true        "the name of cluster"
// @Param   apps     body   cluster.ReleaseList   true    "Apps list of Cluster"
// @Success 200 {object} ex.ApiResponse "OK"
// @Failure 400 {object} ex.ApiResponse "Invalid Name supplied!"
// @Failure 404 {object} ex.ApiResponse "namespace not found"
// @Failure 500 {object} ex.ApiResponse "Server Error"
// @Router /cluster/namespace/{namespace}/name/{name}/list [post]
func DeployListInCluster(c *gin.Context) {
	var namespace, name string
	if values, err := util.GetPathParams(c, []string{"namespace", "name"}); err != nil {
		return
	} else {
		namespace, name = values[0], values[1]
	}

	var postdata ReleaseList
	if err := c.BindJSON(&postdata); err != nil {
		c.JSON(ex.ReturnBadRequest())

	} else {

		if len(postdata.Apps) > 0 {
			for _, app := range postdata.Apps {
				if err := deployInstance(namespace, name, map[string]interface{}{}, app); err != nil {
					c.JSON(ex.ReturnInternalServerError(err))
				}
			}
			c.JSON(ex.ReturnOK())
		} else {
			c.JSON(ex.ReturnBadRequest())

		}
	}

}

// DeployCluster godoc
// @Tags Cluster
// @Description Deploy an Cluster
// @OperationId DeployCluster
// @Accept  json
// @Produce  json
// @Param   namespace     path    string     true        "identifier of the namespace"
// @Param   name     path    string     true        "the name of cluster"
// @Param   apps     body    cluster.Cluster    true    "Apps of Cluster"
// @Success 200 {object} ex.ApiResponse "OK"
// @Failure 400 {object} ex.ApiResponse "Invalid Name supplied!"
// @Failure 404 {object} ex.ApiResponse "namespace not found"
// @Failure 500 {object} ex.ApiResponse "Server Error"
// @Router /cluster/namespace/{namespace}/name/{name} [post]
func DeployCluster(c *gin.Context) {

	var namespace, name string
	if values, err := util.GetPathParams(c, []string{"namespace", "name"}); err != nil {
		return
	} else {
		namespace, name = values[0], values[1]
	}

	var postdata Cluster
	if err := c.BindJSON(&postdata); err != nil {
		c.JSON(ex.ReturnBadRequest())
		return
	} else {

		if len(postdata.Apps) > 0 {
			if err, apps := getGragh(name, namespace, &postdata); err != nil {
				c.JSON(ex.ReturnInternalServerError(err))
				return
			} else {
				for _, app := range apps {
					if err := deployInstance(namespace, name, postdata.ConfigValues, app); err != nil {
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

func mergeConf(conf1, conf2 map[string]interface{}) map[string]interface{} {
	if len(conf1) == 0 {
		return conf2
	}
	result := conf2
	for k, v := range conf1 {
		//if value,ok:=conf1[k];ok{}
		result[k] = v
	}
	return result
}

func deployInstance(namespace, name string, conf map[string]interface{}, app release.ReleaseRequest) error {

	/*
		if len(app.Name) == 0 {
			app.Name = fmt.Sprintf("%s-%s-%s", name, app.ChartName, app.ChartVersion)
		}
		app.Namespace = namespace
	*/
	app.ConfigValues = mergeConf(conf, app.ConfigValues)

	if err := helm.InstallUpgradeRealese(app); err != nil {
		return err
	}
	return nil
}

// GetCluster godoc
// @Tags Cluster
// @Description Get states of an Cluster
// @OperationId GetCluster
// @Accept  json
// @Produce  json
// @Param   namespace     path    string     true        "identifier of the namespace"
// @Param   name     path    string     true        "the name of cluster"
// @Success 200 {array} helm.ReleaseInfo "OK"
// @Failure 400 {object} ex.ApiResponse "Invalid Name supplied!"
// @Failure 404 {object} ex.ApiResponse "cluster not found"
// @Failure 500 {object} ex.ApiResponse "Server Error"
// @Router /cluster/namespace/{namespace}/name/{name} [get]
func GetCluster(c *gin.Context) {

	var clusterReleases []release.ReleaseInfo
	var namespace, name string
	if values, err := util.GetPathParams(c, []string{"namespace", "name"}); err != nil {
		return
	} else {
		if releases, err := helm.ListReleases(values[0]); err != nil {
			c.JSON(ex.ReturnInternalServerError(err))
			return
		} else {

			namespace, name = values[0], values[1]
			for _, release := range releases {
				if release.Namespace == namespace && release.Name[0:len(name)] == name {
					clusterReleases = append(clusterReleases, release)
				}
			}
		}
	}
	c.JSON(http.StatusOK, clusterReleases)

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
// @Router /cluster/namespace/{namespace}/name/{name} [delete]
func DeleteCluster(c *gin.Context) {
	var clusterReleases []string
	var namespace, name string
	if values, err := util.GetPathParams(c, []string{"namespace", "name"}); err != nil {
		return
	} else {
		namespace, name = values[0], values[1]
		if releases, err := helm.ListReleases(namespace); err != nil {
			c.JSON(ex.ReturnInternalServerError(err))
			return
		} else {

			for _, release := range releases {
				if release.Namespace == namespace && release.Name[0:len(name)] == name {
					clusterReleases = append(clusterReleases, release.Name)
				}
			}
		}
	}

	for _, clusterRelease := range clusterReleases {
		if err := helm.DeleteRealese(namespace, clusterRelease); err != nil {
			c.JSON(ex.ReturnInternalServerError(err))
			return
		}
	}
	c.JSON(ex.ReturnOK())

}

func getReleasMap(namespace, name string) (error, map[string]string) {

	releaeMap := map[string]string{}

	if releases, err := helm.ListReleases(namespace); err != nil {
		return err, nil
	} else {

		for _, release := range releases {
			begin, end := len(namespace)+1, len(namespace)+1+len(name)
			if release.Namespace == namespace && release.Name[begin:end] == name {
				releaeMap[release.ChartName] = release.Name
			}
		}
	}
	return nil, releaeMap

}
