package instance

import (
	"bytes"
	"net/http"
	"strings"
	"time"

	"walm/models"
	helm "walm/pkg/helm"
	. "walm/pkg/util/log"
	"walm/router/ex"

	"github.com/ghodss/yaml"
	"github.com/gin-gonic/gin"
	"github.com/gosuri/uitable"
)

type WalmInterface interface {
	Detele(args, flags []string) error
	Rollback(args, flags []string) error
	DeplyApplications(args, flags []string) error
	UpdateApplications(args, flags []string) error
	StatusApplications(args, flags []string) (string, error)
	ListApplications(args, flags []string) (*bytes.Buffer, error)
	MakeValueFile(data []byte) (string, error) //if or not delete file after install or update
}

var walmInst WalmInterface

func init() {
	SetWalmInst(helm.Helm)
}

func SetWalmInst(inter WalmInterface) {
	walmInst = inter
}

// DeleteApplication godoc
// @Description Delete an Appliation
// @OperationId DeleteApplication
// @Accept  json
// @Produce  json
// @Param   appName     path    string     true        "identifier of the application"
// @Success 200 {object} ex.ApiResponse "OK"
// @Failure 400 {object} ex.ApiResponse "Invalid Name supplied!"
// @Failure 404 {object} ex.ApiResponse "Application not found"
// @Failure 500 {object} ex.ApiResponse "Server Error"
// @Router /application/{appName} [delete]
func DeleteApplication(c *gin.Context) {

	var args []string
	var flags []string
	name := c.Param("appname")
	if len(name) == 0 {
		c.JSON(ex.ReturnBadRequest())
		return
	} else {
		args = append(args, name)
	}
	namespace := c.Param("namespace")
	if len(namespace) == 0 {
		c.JSON(ex.ReturnBadRequest())
		return
	}

	Log.Infof("begin to delete instance:%s; namespace:%s. \n", name, namespace)
	defer Log.Infof("finish delete instance:%s; namespace:%s. \n", name, namespace)
	if err := walmInst.Detele(args, flags); err != nil {
		c.JSON(ex.ReturnInternalServerError(err))
		return
	}

	if err := models.DeleteAppInst(name); err != nil {
		Log.Errorf("occu error when delete AppInst: %s \n", err)
		c.JSON(ex.ReturnInternalServerError(err))
		return
	}

	c.JSON(ex.ReturnOK())

}

// Deploy godoc
// @Description deploy an application with the givin data
// @OperationId Deploy
// @Accept  json
// @Produce  json
// @Param   chart     path    string     true      "identifier of the chart"
// @Param   application     body   instance.Application    true    "Update application"
// @Success 200 {object} ex.ApiResponse "OK"
// @Failure 400 {object} ex.ApiResponse "Invalid Name supplied!"
// @Failure 405 {object} ex.ApiResponse "Invalid input"
// @Failure 500 {object} ex.ApiResponse "Server Error"
// @Router /application/{chart} [post]
func DeployApplication(c *gin.Context) {

	var args []string
	var flags []string
	chart := c.Param("chart")
	if len(chart) == 0 {
		c.JSON(ex.ReturnBadRequest())
		return
	}
	var postdata Application
	if err := c.Bind(&postdata); err != nil {
		c.JSON(ex.ReturnBadRequest())
		return
	} else {
		if len(postdata.Name) > 0 {
			/*
			 flags = append(flags,"--name")
			 flags = append(flags,postdata.Name)
			*/
			args = append(args, postdata.Name)
			args = append(args, chart)
		} else {
			c.JSON(ex.ReturnBadRequest())
			return
		}
		if len(postdata.Namespace) > 0 {
			flags = append(flags, "--namespace")
			flags = append(flags, postdata.Namespace)
		}
		if len(postdata.Repo) > 0 {
			flags = append(flags, "--repo")
			flags = append(flags, postdata.Repo)
		}
		if len(postdata.Version) > 0 {
			flags = append(flags, "--version")
			flags = append(flags, postdata.Version)
		}

		if len(postdata.Links) > 0 {
			for _, v := range postdata.Links {
				flags = append(flags, "--link")
				flags = append(flags, v)
			}
		}

		if len(postdata.Value) > 0 {
			flags = append(flags, "--values")
			if data, err := yaml.JSONToYAML([]byte(postdata.Value)); err != nil {
				c.JSON(ex.ReturnInternalServerError(err))
			} else {
				if name, err := walmInst.MakeValueFile(data); err != nil {
					c.JSON(ex.ReturnInternalServerError(err))
					return
				} else {
					flags = append(flags, name)
				}
			}
		}
	}

	app := models.AppInst{
		Name:        postdata.Name,
		Namespace:   postdata.Namespace,
		AppPkg:      chart,
		Vers:        postdata.Version,
		ConfigTemp:  postdata.Value,
		Status:      "deployed",
		InstallTime: time.Now().Unix(),
	}

	Log.Infof("begin to install instance:%s; namespace:%s; chart \n", app.Name, app.Namespace, app.AppPkg)
	defer Log.Infof("finish delete instance:%s; namespace:%s; chart:%s \n", app.Name, app.Namespace, app.AppPkg)
	if err := walmInst.DeplyApplications(args, flags); err != nil {
		c.JSON(ex.ReturnInternalServerError(err))
		return
	}

	ti := time.Now().Unix()
	app.InstalledTime, app.LastTime = ti, ti

	if err := models.InsertAppInst(app); err != nil {
		Log.Errorf("occu error when insert into AppInst: %s \n", err)
		c.JSON(ex.ReturnInternalServerError(err))
		return
	}
	c.JSON(ex.ReturnOK())
}

// FindApplicationsStatus godoc
// @Description find the aim application status
// @OperationId FindApplicationsStatus
// @Accept  json
// @Produce  json
// @Param   namespace     path    string     true      "identifier of the chart"
// @Param   appname     path    string     true      "identifier of the appName"
// @Param   reverse     query    boolean     false      "identifier of the appName"
// @Param   max     query    int     false      "max num to display"
// @Param   offset     query    int     false      "the offset of result"
// @Param   all     query    boolean     false      "if display all of result"
// @Param   deleted     query    boolean     false      "if display deleted release"
// @Param   deleting     query    boolean     false      "if display deleting release"
// @Param   deployed     query    boolean     false      "if display deployed release"
// @Param   failed     query    boolean     false      "if display failed release"
// @Param   pending     query    boolean     false      "if display pending release"
// @Success 200 {object} instance.Info	"ok"
// @Failure 404 {object} ex.ApiResponse "Invalid status not found"
// @Failure 500 {object} ex.ApiResponse "Server Error"
// @Router /application/{namespace}/status/{appname} [get]
func ListApplicationsWithStatus(c *gin.Context) {

	var args []string
	var flags []string

	name := c.Param("appname")
	if len(name) == 0 {
		c.JSON(ex.ReturnBadRequest())
		return
	} else {
		args = append(args, name)
	}

	namespace := c.Param("namespace")
	if len(namespace) > 0 {
		flags = append(flags, "--namespace")
		flags = append(flags, namespace)
	}

	reverse := c.Query("reverse")
	if len(reverse) > 0 {
		flags = append(flags, "--reverse")
	}

	max := c.Query("max")
	if len(max) > 0 {
		flags = append(flags, "--max")
		flags = append(flags, max)
	}

	offset := c.Query("offset")
	if len(offset) > 0 {
		flags = append(flags, "--offset")
		flags = append(flags, offset)
	}

	all := c.Query("all")
	if len(all) > 0 {
		flags = append(flags, "--all")
	}

	deleted := c.Query("deleted")
	if len(deleted) > 0 {
		flags = append(flags, "--deleted")
	}

	deleting := c.Query("deleting")
	if len(deleting) > 0 {
		flags = append(flags, "--deleting")
	}

	deployed := c.Query("deployed")
	if len(deployed) > 0 {
		flags = append(flags, "--deployed")
	}

	failed := c.Query("failed")
	if len(failed) > 0 {
		flags = append(flags, "--failed")
	}

	pending := c.Query("pending")
	if len(pending) > 0 {
		flags = append(flags, "--pending")
	}

	if table, err := walmInst.ListApplications(args, flags); err != nil {
		c.JSON(ex.ReturnInternalServerError(err))
	} else {
		var ia []Info
		for _, line := range strings.Split(table.String(), "/n") {
			values := strings.Split(line, uitable.Separator)
			ia = append(ia, Info{
				Name:      values[0],
				Revision:  values[1],
				Updated:   values[2],
				Status:    values[3],
				Chart:     values[4],
				Namespace: values[5],
			})
		}
		for _, info := range ia {
			if err := models.UpdateAppInst(info.Name, info.Status, "status"); err != nil {
				Log.Errorf("occu error when update  AppInst: %s \n", err)
				c.JSON(ex.ReturnInternalServerError(err))
			}
		}
		c.JSON(http.StatusOK, ia)
	}

}

// GetApplicationbyName godoc
// @Description Get an Appliation by name with status
// @OperationId GetApplicationbyName
// @Accept  json
// @Produce  json
// @Param   appName     path    string     true        "identifier of the application"
// @Success 200 {object} instance.Info	"ok"
// @Failure 400 {object} ex.ApiResponse "Invalid Name supplied!"
// @Failure 404 {object} ex.ApiResponse "Application not found"
// @Failure 405 {object} ex.ApiResponse "Invalid input"
// @Failure 500 {object} ex.ApiResponse "Server Error"
// @Router /application/{appName} [get]
func GetApplicationStatusbyName(c *gin.Context) {
	var args []string
	var flags []string

	name := c.Param("appname")
	if len(name) == 0 {
		c.JSON(ex.ReturnBadRequest())
		return
	}
	args = append(args, name)

	flags = append(flags, "--revision")
	flags = append(flags, "--output json")

	if str, err := walmInst.StatusApplications(args, flags); err != nil {
		c.JSON(ex.ReturnInternalServerError(err))
	} else {
		c.JSON(http.StatusOK, str) //need edit
		return
	}
}

// RollBackApplication godoc
// @Description Rollback Application to aim version
// @OperationId RollBackApplication
// @Accept  json
// @Produce  json
// @Param   appName     path    string     true        "identifier of the application"
// @Param   version     path    string     true        "identifier of the version"
// @Param   recreate     query    boolean     false      "if recreate pods"
// @Param   force     query    boolean     false      "if force to update and restart pods"
// @Param   wait     query    boolean     false      "if wait for finish"
// @Success 200 {object} instance.Info	"ok"
// @Failure 400 {object} ex.ApiResponse "Invalid Name supplied!"
// @Failure 404 {object} ex.ApiResponse "Application not found"
// @Failure 405 {object} ex.ApiResponse "Invalid input"
// @Failure 500 {object} ex.ApiResponse "Server Error"
// @Router /application/{appname}/rollback/{version} [get]
func RollBackApplication(c *gin.Context) {
	var args []string
	var flags []string

	name := c.Param("appname")
	if len(name) == 0 {
		c.JSON(ex.ReturnBadRequest())
		return
	}

	namespace := c.Param("namespace")
	if len(namespace) == 0 {
		c.JSON(ex.ReturnBadRequest())
		return
	}

	args = append(args, name)

	version := c.Param("version")
	if len(version) > 0 {
		args = append(args, version)
	}

	recreate := c.Query("recreate")
	if len(recreate) > 0 {
		flags = append(flags, "--recreate-pods")
	}

	force := c.Query("force")
	if len(force) > 0 {
		flags = append(flags, "--force")

	}

	wait := c.Query("wait")
	if len(wait) > 0 {
		flags = append(flags, "--wait")

	}

	Log.Infof("begin to rollback instance %s; namespace %s; version: %s \n", name, namespace, version)
	defer Log.Infof("finish delete instance %s; namespace %s; version: %s \n", name, namespace, version)
	if err := walmInst.Rollback(args, flags); err != nil {
		c.JSON(ex.ReturnInternalServerError(err))
	} else {
		if err := models.UpdateAppInst(name, version, "version"); err != nil {
			Log.Errorf("occu error when update  AppInst: %s \n", err)
			c.JSON(ex.ReturnInternalServerError(err))
		}
		c.JSON(ex.ReturnOK())
	}
}

// UpdateApplication godoc
// @Description Update an Appliation
// @OperationId UpdateApplication
// @Accept  json
// @Produce  json
// @Param   chart     path    string     true      "identifier of the chart"
// @Param   application     body   instance.Application    true    "Update application"
// @Success 200 {object} ex.ApiResponse "OK"
// @Failure 400 {object} ex.ApiResponse "Invalid Name supplied!"
// @Failure 404 {object} ex.ApiResponse "Application not found"
// @Failure 405 {object} ex.ApiResponse "Invalid input"
// @Failure 500 {object} ex.ApiResponse "Server Error"
// @Router /application/{chart} [put]
func UpdateApplication(c *gin.Context) {
	var args []string
	var flags []string
	chart := c.Param("chart")
	if len(chart) == 0 {
		c.JSON(ex.ReturnBadRequest())
		return
	} else {
		args = append(args, chart)
	}
	var postdata Application
	if err := c.Bind(&postdata); err != nil {
		c.JSON(ex.ReturnBadRequest())
		return
	} else {
		if len(postdata.Name) > 0 {
			flags = append(flags, "--name")
			flags = append(flags, postdata.Name)
		}
		if len(postdata.Namespace) > 0 {
			flags = append(flags, "--namespace")
			flags = append(flags, postdata.Namespace)
		}
		if len(postdata.Repo) > 0 {
			flags = append(flags, "--repo")
			flags = append(flags, postdata.Repo)
		}
		if len(postdata.Version) > 0 {
			flags = append(flags, "--version")
			flags = append(flags, postdata.Version)
		}
		if postdata.Install {
			flags = append(flags, "--install")
			//flags = append(flags,postdata.Install)
		}
		if postdata.ResetValue {
			flags = append(flags, "--reset-values")
			//flags = append(flags,postdata.ResetValue)
		}
		if postdata.ReuseValue {
			flags = append(flags, "--reuse-values")
			//flags = append(flags,postdata.ReuseValue)
		}

		if len(postdata.Links) > 0 {
			for _, v := range postdata.Links {
				flags = append(flags, "--link")
				flags = append(flags, v)
			}
		}

		if len(postdata.Value) > 0 {
			flags = append(flags, "--values")
			if data, err := yaml.JSONToYAML([]byte(postdata.Value)); err != nil {
				c.JSON(ex.ReturnInternalServerError(err))
			} else {
				if name, err := walmInst.MakeValueFile(data); err != nil {
					c.JSON(ex.ReturnInternalServerError(err))
					return
				} else {
					flags = append(flags, name)
				}
			}
		}

	}
	Log.Infof("begin to update instance %s; namespace %s; chart: %s", postdata.Name, postdata.Namespace, chart)
	defer Log.Infof("finish update instance %s; namespace %s; chart: %s", postdata.Name, postdata.Namespace, chart)
	if err := walmInst.UpdateApplications(args, flags); err != nil {
		c.JSON(ex.ReturnInternalServerError(err))
		return
	}

	ti := time.Now().Unix()
	app := models.AppInst{
		Name:          postdata.Name,
		Namespace:     postdata.Namespace,
		AppPkg:        chart,
		Vers:          postdata.Version,
		ConfigTemp:    postdata.Value,
		Status:        "deployed",
		InstalledTime: ti,
		LastTime:      ti,
	}

	if err := models.UpdateAppInstByApp(app); err != nil {
		Log.Errorf("occu error when update AppInst: %s", err)
		c.JSON(ex.ReturnInternalServerError(err))
		return
	}
	c.JSON(ex.ReturnOK())
}
