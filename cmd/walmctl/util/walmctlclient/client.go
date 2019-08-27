package walmctlclient

import (
	"WarpCloud/walm/pkg/util"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty"
	"github.com/pkg/errors"
	"k8s.io/klog"
	"net"
	"strconv"
	"time"
)

type WalmctlClient struct {
	protocol   string
	hostURL    string
	apiVersion string
	baseURL    string
}

var walmctlClient *WalmctlClient

func CreateNewClient(hostURL string) *WalmctlClient {
	if walmctlClient == nil {
		walmctlClient = &WalmctlClient{
			protocol:   "http://",
			hostURL:    hostURL,
			apiVersion: "/api/v1",
		}
		walmctlClient.baseURL = walmctlClient.protocol + walmctlClient.hostURL + walmctlClient.apiVersion
	}
	return walmctlClient
}

func (c *WalmctlClient) ValidateHostConnect() error {
	timeout := time.Duration(1 * time.Second)
	_, err := net.DialTimeout("tcp", walmctlClient.hostURL, timeout)
	if err != nil {
		return errors.Errorf("WalmServer unreachable, error: %s", err.Error())
	}
	return nil
}

// release
func (c *WalmctlClient) CreateRelease(namespace, chart string, releaseName string, async bool, timeoutSec int64, configValues map[string]interface{}) (*resty.Response, error) {
	fullUrl := walmctlClient.baseURL + "/release/" + namespace + "?async=" + strconv.FormatBool(async) +
		"&timeoutSec=" + strconv.FormatInt(timeoutSec, 10)

	if releaseName != "" {
		releaseNameConfigs := make(map[string]interface{}, 0)
		releaseNameConfigs["name"] = releaseName
		util.MergeValues(configValues, releaseNameConfigs, false)
	}
	filestr, err := json.Marshal(configValues)
	if err != nil {
		klog.Errorf("marshal to json error %v", err)
	}

	resp := &resty.Response{}
	if chart != "" {
		chartFullUrl := walmctlClient.baseURL + "/release/" + namespace + "/withchart"
		resp, err = resty.R().
			SetHeader("Content-Type", "multipart/form-data").
			SetFile("chart", chart).
			SetFormData(map[string]string{
				"release": releaseName,
				"body":    string(filestr[:]),
			}).
			Post(chartFullUrl)
	} else {
		resp, err = resty.R().
			SetHeader("Accept", "application/json").
			SetBody(filestr).
			Post(fullUrl)
	}
	if resp == nil || resp.StatusCode() != 200 {
		return nil, errors.New(fmt.Sprintf("error response %v %v", err, resp))
	}
	return resp, err
}

func (c *WalmctlClient) GetRelease(namespace string, releaseName string) (resp *resty.Response, err error) {
	fullUrl := walmctlClient.baseURL + "/release/" + namespace + "/name/" + releaseName

	resp, err = resty.R().Get(fullUrl)
	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.String())
	}

	return resp, err
}

func (c *WalmctlClient) UpdateRelease(namespace string, newConfigStr string, async bool, timeoutSec int64) (resp *resty.Response, err error) {
	fullUrl := walmctlClient.baseURL + "/release/" + namespace + "?async=" + strconv.FormatBool(async) +
		"&timeoutSec=" + strconv.FormatInt(timeoutSec, 10)

	resp, err = resty.R().SetHeader("Accept", "application/json").
		SetBody(newConfigStr).
		Put(fullUrl)
	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.String())
	}
	return resp, err
}

func (c *WalmctlClient) UpdateReleaseWithChart(namespace string, releaseName string, file string, newConfigStr string) (resp *resty.Response, err error) {
	fullUrl := walmctlClient.baseURL + "/release/" + namespace + "/withchart"

	resp, err = resty.R().SetHeader("Content-Type", "multipart/form-data", ).
		SetFile("chart", file).
		SetFormData(map[string]string{"body": newConfigStr}).
		Put(fullUrl)

	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.String())
	}
	return resp, err
}

func (c *WalmctlClient) DeleteRelease(namespace string, releaseName string, async bool, timeoutSec int64, deletePvcs bool) (resp *resty.Response, err error) {
	fullUrl := walmctlClient.baseURL + "/release/" + namespace + "/name/" + releaseName + "?async=" + strconv.FormatBool(async) +
		"&timeoutSec=" + strconv.FormatInt(timeoutSec, 10) + "&deletePvcs=" + strconv.FormatBool(deletePvcs)

	resp, err = resty.R().
		Delete(fullUrl)

	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.String())
	}
	return resp, err

}

func (c *WalmctlClient) ListRelease(namespace string, labelSelector string) (resp *resty.Response, err error) {
	fullUrl := walmctlClient.baseURL + "/release/" + namespace
	if namespace == "" {
		fullUrl = walmctlClient.baseURL + "/release"
	}

	resp, err = resty.R().
		SetHeader("Accept", "application/json").
		Get(fullUrl)

	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.String())
	}
	return resp, err
}

// project
func (c *WalmctlClient) CreateProject(namespace, chartPath, projectName string, async bool, timeoutSec int64, configValues map[string]interface{}) (resp *resty.Response, err error) {
	fullUrl := walmctlClient.baseURL + "/project/" + namespace + "/name/" + projectName + "?async=" + strconv.FormatBool(async) +
		"&timeoutSec=" + strconv.FormatInt(timeoutSec, 10)

	filestr, err := json.Marshal(configValues)
	if err != nil {
		klog.Errorf("marshal to json error %v", err)
	}

	resp, err = resty.R().SetHeader("Content-Type", "application/json").
		SetBody(filestr).
		Post(fullUrl)

	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.String())
	}

	return resp, err
}

func (c *WalmctlClient) GetProject(namespace string, projectName string) (resp *resty.Response, err error) {
	fullUrl := walmctlClient.baseURL + "/project/" + namespace + "/name/" + projectName
	resp, err = resty.R().Get(fullUrl)
	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.String())
	}
	return resp, err
}

func (c *WalmctlClient) DeleteProject(namespace string, projectName string, async bool, timeoutSec int64, deletePvcs bool) (resp *resty.Response, err error) {
	fullUrl := walmctlClient.baseURL + "/project/" + namespace + "/name/" + projectName + "?async=" + strconv.FormatBool(async) +
		"&timeoutSec=" + strconv.FormatInt(timeoutSec, 10) + "&deletePvcs=" + strconv.FormatBool(deletePvcs)

	resp, err = resty.R().
		Delete(fullUrl)

	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.String())
	}

	return resp, err
}

func (c *WalmctlClient) ListProject(namespace string) (resp *resty.Response, err error) {
	fullUrl := walmctlClient.baseURL + "/project/" + namespace
	if namespace == "" {
		fullUrl = walmctlClient.baseURL + "/project"
	}

	resp, err = resty.R().
		SetHeader("Accept", "application/json").
		Get(fullUrl)

	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.String())
	}

	return resp, err
}

func (c *WalmctlClient) DeleteReleaseInProject(namespace string, projectName string, releaseName string, async bool, timeoutSec int64, deletePvcs bool) (resp *resty.Response, err error) {
	fullUrl := walmctlClient.baseURL + "/project/" + namespace + "/name/" + projectName + "/instance/" + releaseName + "?async=" + strconv.FormatBool(async) +
		"&timeoutSec=" + strconv.FormatInt(timeoutSec, 10) + "&deletePvcs=" + strconv.FormatBool(deletePvcs)

	resp, err = resty.R().
		Delete(fullUrl)

	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.String())
	}

	return resp, err
}
