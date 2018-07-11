package cluster

import (
	"github.com/gin-gonic/gin"

	"walm/router/api/v1/instance"
	"walm/router/ex"
)

// DeployInstanceInCluster godoc
// @Tags Cluster
// @Description Deploy an Instance into Cluster
// @OperationId DeployInstanceInCluster
// @Accept  json
// @Produce  json
// @Param   namespace     path    string     true        "identifier of the namespace"
// @Param   name     path    string     true        "the name of cluster"
// @Param   apps     body   instance.Application    true    "Apps of Cluster"
// @Success 200 {object} ex.ApiResponse "OK"
// @Failure 400 {object} ex.ApiResponse "Invalid Name supplied!"
// @Failure 404 {object} ex.ApiResponse "namespace not found"
// @Failure 500 {object} ex.ApiResponse "Server Error"
// @Router /cluster/namespace/{namespace}/name/{name}/instance [post]
func DeployInstanceInCluster(c *gin.Context) {
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
	var postdata instance.Application
	if err := c.BindJSON(&postdata); err != nil {
		c.JSON(ex.ReturnBadRequest())
		return
	} else {

		//if err, cluster := GetClusterInfo(name); err != nil {
		//	c.JSON(ex.ReturnClusterNotExistError())
		//	return
		//} else {
		//
		//	if err, releaseMap := getReleasMap(name); err != nil {
		//		c.JSON(ex.ReturnInternalServerError(err))
		//		return
		//	} else {
		//		if err, apps := getGraghForInstance(cluster.ClusterId, releaseMap, &postdata); err != nil {
		//			c.JSON(ex.ReturnInternalServerError(err))
		//			return
		//		} else {
		//			for _, app := range apps {
		//				if err := deployInstance(cluster, app); err != nil {
		//					c.JSON(ex.ReturnInternalServerError(err))
		//				}
		//			}
		//		}
		//	}
		//
		//}
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
// @Param   apps     body   cluster.Cluster    true    "Apps of Cluster"
// @Success 200 {object} ex.ApiResponse "OK"
// @Failure 400 {object} ex.ApiResponse "Invalid Name supplied!"
// @Failure 404 {object} ex.ApiResponse "namespace not found"
// @Failure 500 {object} ex.ApiResponse "Server Error"
// @Router /cluster/namespace/{namespace}/name/{name} [post]
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

		//if len(postdata.Apps) > 0 {
		//	if err, apps := getGragh(cluster.ClusterId, name, namespace, &postdata); err != nil {
		//		c.JSON(ex.ReturnInternalServerError(err))
		//		return
		//	} else {
		//		for _, app := range apps {
		//			if err := deployInstance(cluster, app); err != nil {
		//				c.JSON(ex.ReturnInternalServerError(err))
		//			}
		//		}
		//	}
		//} else {
		//	c.JSON(ex.ReturnBadRequest())
		//	return
		//}
	}

}

//func deployInstance(cluster *models.Cluster, app instance.Application) error {
//	var args []string
//	var flags []string
//
//	if len(app.Name) > 0 {
//		args = append(args, app.Name)
//		args = append(args, app.Chart)
//	} else {
//		app.Name = fmt.Sprintf("%s-%d-%s", cluster.Name, cluster.ClusterId, app.Chart)
//	}
//	if len(app.Namespace) > 0 {
//		flags = append(flags, "--namespace")
//		flags = append(flags, app.Namespace)
//	}
//	if len(app.Repo) > 0 {
//		flags = append(flags, "--repo")
//		flags = append(flags, app.Repo)
//	}
//	if len(app.Version) > 0 {
//		flags = append(flags, "--version")
//		flags = append(flags, app.Version)
//	}
//
//	if len(app.Links) > 0 {
//		for k, v := range app.Links {
//			flags = append(flags, "--link")
//			flags = append(flags, k+"="+v)
//		}
//	}
//
//	if len(cluster.ConfigTemp) > 0 {
//		flags = append(flags, "--values")
//		if data, err := yaml.JSONToYAML([]byte(cluster.ConfigTemp)); err != nil {
//			return err
//		} else {
//			if name, err := file.MakeValueFile(data); err != nil {
//				return err
//			} else {
//				flags = append(flags, name)
//			}
//		}
//	}
//
//	if len(app.Value) > 0 {
//		flags = append(flags, "--values")
//		if data, err := yaml.JSONToYAML([]byte(app.Value)); err != nil {
//			return err
//		} else {
//			if name, err := file.MakeValueFile(data); err != nil {
//				return err
//			} else {
//				flags = append(flags, name)
//			}
//		}
//	}
//
//	if err := helm.Helm.DeplyApplications(args, flags); err != nil {
//		return err
//	}
//
//	return nil
//}
//
//// StatusCluster godoc
//// @Tags Cluster
//// @Description Get states of an Cluster
//// @OperationId StatusCluster
//// @Accept  json
//// @Produce  json
//// @Param   namespace     path    string     true        "identifier of the namespace"
//// @Param   name     path    string     true        "the name of cluster"
//// @Success 200 {object} ex.ApiResponse "OK"
//// @Failure 400 {object} ex.ApiResponse "Invalid Name supplied!"
//// @Failure 404 {object} ex.ApiResponse "cluster not found"
//// @Failure 500 {object} ex.ApiResponse "Server Error"
//// @Router /cluster/namespace/{namespace}/name/{name} [get]
//func StatusCluster(c *gin.Context) {
//	//ListApplications
//	name := c.Param("name")
//	if len(name) == 0 {
//		c.JSON(ex.ReturnBadRequest())
//		return
//	}
//	namespace := c.Param("namespace")
//	if len(namespace) == 0 {
//		c.JSON(ex.ReturnBadRequest())
//		return
//	}
//
//	info := Info{
//		Name:  name,
//		Infos: []instance.Info{},
//	}
//
//	if releases, err := GetReleasesOfCluster(name); err != nil {
//		c.JSON(ex.ReturnInternalServerError(err))
//		return
//	} else {
//		Log.Infof("begin to get status of cluser %s; namespace %s; ", name, namespace)
//		var errs []error
//
//		for _, release := range releases {
//			if table, err := helm.Helm.ListApplications([]string{release.Name}, []string{"--namespace", namespace, "--all"}); err != nil {
//				errs = append(errs, err)
//			} else {
//				for _, line := range strings.Split(table.String(), "/n") {
//					values := strings.Split(line, uitable.Separator)
//					info.Infos = append(info.Infos, instance.Info{
//						Name:      values[0],
//						Revision:  values[1],
//						Updated:   values[2],
//						Status:    values[3],
//						Chart:     values[4],
//						Namespace: values[5],
//					})
//				}
//			}
//		}
//		if len(errs) > 0 {
//			c.JSON(ex.ReturnInternalServerError(errs[0]))
//			return
//		}
//	}
//	c.JSON(http.StatusOK, info)
//}
//
//// DeleteCluster godoc
//// @Tags Cluster
//// @Description Delete an Cluster
//// @OperationId DeleteCluster
//// @Accept  json
//// @Produce  json
//// @Param   namespace     path    string     true        "identifier of the namespace"
//// @Param   name     path    string     true        "the name of cluster"
//// @Success 200 {object} ex.ApiResponse "OK"
//// @Failure 400 {object} ex.ApiResponse "Invalid Name supplied!"
//// @Failure 404 {object} ex.ApiResponse "cluster not found"
//// @Failure 500 {object} ex.ApiResponse "Server Error"
//// @Router /cluster/namespace/{namespace}/name/{name} [delete]
//func DeleteCluster(c *gin.Context) {
//	name := c.Param("name")
//	if len(name) == 0 {
//		c.JSON(ex.ReturnBadRequest())
//		return
//	}
//	namespace := c.Param("namespace")
//	if len(namespace) == 0 {
//		c.JSON(ex.ReturnBadRequest())
//		return
//	}
//	if releases, err := GetReleasesOfCluster(name); err != nil {
//		c.JSON(ex.ReturnInternalServerError(err))
//		return
//	} else {
//		Log.Infof("begin to delete cluser %s; namespace %s;", name, namespace)
//		var errs []error
//		for _, release := range releases {
//			if err := helm.Helm.Detele([]string{release.Name}, []string{}); err != nil {
//				errs = append(errs, err)
//			}
//		}
//		if len(errs) > 0 {
//			c.JSON(ex.ReturnInternalServerError(errs[0]))
//			return
//		}
//	}
//
//	c.JSON(ex.ReturnOK())
//
//}
//
//func getReleasMap(name string) (error, map[string]string) {
//	if releases, err := GetReleasesOfCluster(name); err != nil {
//		return err, nil
//	} else {
//		releaeMap := map[string]string{}
//		for _, release := range releases {
//			releaeMap[release.Release] = release.Name
//		}
//		return nil, releaeMap
//	}
//}
//
//func GetReleasesOfCluster(name string) ([]release, err) {
//	return []release{}, nil
//}
