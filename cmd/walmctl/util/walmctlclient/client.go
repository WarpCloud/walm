package walmctlclient

import (
	"io/ioutil"
	"strconv"
	"github.com/go-resty/resty"
	"github.com/bitly/go-simplejson"
	"github.com/pkg/errors"
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

// release

func (c *WalmctlClient) CreateRelease(namespace string, releaseName string, async bool, timeoutSec int64, file string) (resp *resty.Response, err error) {

	fullUrl := walmctlClient.baseURL + "/release/" + namespace + "?async=" + strconv.FormatBool(async) +
		"&timeoutSec="+ strconv.FormatInt(timeoutSec, 10)
	fileByte, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	fileJson, err := simplejson.NewJson(fileByte)
	if err != nil {
		return nil, err
	}

	if len(releaseName) > 0 {
		fileJson.Set("name", releaseName)
		fileByte, err = fileJson.MarshalJSON()
		if err != nil {
			return nil, err
		}
	}
	filestr := string(fileByte[:])

	resp, err = resty.R().SetHeader("Content-Type", "application/json").
		SetBody(filestr).
		Post(fullUrl)

	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.String())
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

	resp, err = resty.R().SetHeader("Content-Type", "application/json").
		SetBody(newConfigStr).
		Put(fullUrl)
	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.String())
	}
	return resp, err
}

func (c *WalmctlClient) UpdateReleaseWithChart(namespace string, releaseName string, file string) (resp *resty.Response, err error) {

	fullUrl := walmctlClient.baseURL + "/release/" + namespace + "/withchart"

	resp, err = resty.R().SetHeader("Content-Type", "multipart/form-data", ).
	   SetFile("chart", file).
	   SetFormData(map[string]string{
	   	"release": releaseName,
	}).
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

	resp, err = resty.R().
		SetHeader("Accept", "application/json").
		Get(fullUrl)

	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.String())
	}
	return resp, err
}

// project

func (c *WalmctlClient) CreateProject(namespace string, projectName string, async bool, timeoutSec int64, file string) (resp *resty.Response, err error) {

	fullUrl := walmctlClient.baseURL + "/project/" + namespace + "/name/" + projectName + "?async=" + strconv.FormatBool(async) +
		"&timeoutSec=" + strconv.FormatInt(timeoutSec, 10)

	fileByte, err := ioutil.ReadFile(file)

	if err != nil {
		return nil, err
	}

	filestr := string(fileByte[:])
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
