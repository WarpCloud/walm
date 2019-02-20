package walmctlclient

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"github.com/go-resty/resty"
	"github.com/bitly/go-simplejson"
)

type WalmctlClient struct {
	protocol   string
	hostURL    string
	apiversion string
	baseURL    string
}

var walmctlClient *WalmctlClient

func CreateNewClient(hostURL string) *WalmctlClient {

	if walmctlClient == nil {

		walmctlClient = &WalmctlClient{
			protocol:   "http://",
			hostURL:    hostURL,
			apiversion: "/api/v1",
		}
		walmctlClient.baseURL = walmctlClient.protocol + walmctlClient.hostURL + walmctlClient.apiversion
	}
	return walmctlClient
}

func (c *WalmctlClient) CreateRelease(namespace string, releaseName string, file string) (resp *resty.Response, err error) {

	fullUrl := walmctlClient.baseURL + "/release/" + namespace + "?async=false&timeoutSec=0"
	fileByte, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println("ReadFile Error: ", err.Error())
	}

	fileJson, err := simplejson.NewJson(fileByte)
	if err != nil {
		fmt.Printf(err.Error())
	}

	if len(releaseName) > 0 {
		fileJson.Set("name", releaseName)
		fileByte, err = fileJson.MarshalJSON()
	}
	filestr := string(fileByte[:])

	resp, err = resty.R().SetHeader("Content-Type", "application/json").
		SetBody(filestr).
		Post(fullUrl)

	return resp, err
}

func (c *WalmctlClient) UpgradeRelease(namespace string, file string) (resp *resty.Response, err error) {

	fullUrl := walmctlClient.baseURL + "release" + namespace + "?async=false&timeoutSec=0"

	fileByte, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println("ReadFile Error: ", err.Error())
	}

	filestr := string(fileByte[:])

	resp, err = resty.R().SetHeader("Content-Type", "application/json").
		SetBody(filestr).
		Put(fullUrl)

	return resp, err
}

func (c *WalmctlClient) DeleteRelease(namespace string, releaseName string, async bool, timeoutSec int64, deletePvcs bool) (resp *resty.Response, err error) {

	fullUrl := walmctlClient.baseURL + "/release/" + namespace + "/name/" + releaseName + "?async=" + strconv.FormatBool(async) +
		"&timeoutSec=" + strconv.FormatInt(timeoutSec, 10) + "&deletePvcs=" + strconv.FormatBool(deletePvcs)

	resp, err = resty.R().
		Delete(fullUrl)
	return resp, err

}

func (c *WalmctlClient) ListRelease(namespace string) (resp *resty.Response, err error) {

	fullUrl := walmctlClient.baseURL + "/release/" + namespace

	resp, err = resty.R().
		SetHeader("Accept", "application/json").
		Get(fullUrl)

	return resp, err
}

func (c *WalmctlClient) CreateProject(namespace string, projectName string, async bool, timeoutSec int64, file string) (resp *resty.Response, err error) {

	fullUrl := walmctlClient.baseURL + "/project/" + namespace + "/name/" + projectName + "?async=" + strconv.FormatBool(async) +
		"&timeoutSec=" + strconv.FormatInt(timeoutSec, 10)

	fileByte, err := ioutil.ReadFile(file)

	if err != nil {
		fmt.Println("ReadFile Error: ", err.Error())
	}

	filestr := string(fileByte[:])

	resp, err = resty.R().SetHeader("Content-Type", "application/json").
		SetBody(filestr).
		Post(fullUrl)

	return resp, err
}

func (c *WalmctlClient) DeleteProject(namespace string, projectName string, async bool, timeoutSec int64, deletePvcs bool) (resp *resty.Response, err error) {

	fullUrl := walmctlClient.baseURL + "/project/" + namespace + "/name/" + projectName + "?async=" + strconv.FormatBool(async) +
		"&timeoutSec=" + strconv.FormatInt(timeoutSec, 10) + "&deletePvcs=" + strconv.FormatBool(deletePvcs)

	resp, err = resty.R().
		Delete(fullUrl)

	return resp, err
}

func (c *WalmctlClient) ListProject(namespace string) (resp *resty.Response, err error) {

	fullUrl := walmctlClient.baseURL + "/project/" + namespace

	resp, err = resty.R().
		SetHeader("Accept", "application/json").
		Get(fullUrl)

	return resp, err

}

func (c *WalmctlClient) DeleteReleaseInProject(namespace string, projectName string, releaseName string, async bool, timeoutSec int64, deletePvcs bool) (resp *resty.Response, err error) {

	fullUrl := walmctlClient.baseURL + "/project/" + namespace + "/name/" + projectName + "/instance/" + releaseName + "?async=" + strconv.FormatBool(async) +
		"&timeoutSec=" + strconv.FormatInt(timeoutSec, 10) + "&deletePvcs=" + strconv.FormatBool(deletePvcs)

	resp, err = resty.R().
		Delete(fullUrl)

	return resp, err
}

func (c *WalmctlClient) GetSource(namespace string, sourceName string, sourceType string) (resp *resty.Response, err error) {

	fullUrl := walmctlClient.baseURL + "/" + sourceType + "/" + namespace + "/name/" + sourceName

	resp, err = resty.R().Get(fullUrl)

	return resp, err
}
