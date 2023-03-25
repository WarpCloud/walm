package walmctlclient

import (
	errModels "WarpCloud/walm/pkg/models/error"
	k8sModel "WarpCloud/walm/pkg/models/k8s"
	"WarpCloud/walm/pkg/util"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty"
	"github.com/pkg/errors"
	"io/ioutil"
	"k8s.io/klog"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)

type WalmctlClient struct {
	client  *resty.Client
	baseURL    string
}

var walmctlClient *WalmctlClient
var NotFoundError errModels.NotFoundError

func CreateNewClient(hostURL string, enableTLS bool, rootCA string) (*WalmctlClient, error) {
	var client *resty.Client
	protocol := "https://"
	apiVersion := "/api/v1"
	if !enableTLS {
		protocol = "http://"
		client = resty.New()
	} else {
		// rootCA.crt required
		if rootCA == "" {
			return nil, errors.Errorf("rootCA(CA root certificate, public key) can not be empty")
		}
		rootCA, err := filepath.Abs(rootCA)
		if err != nil {
			return nil, err
		}
		cert, err := ioutil.ReadFile(rootCA)
		if err != nil {
			return nil, err
		}
		certPool, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
	}
		certPool.AppendCertsFromPEM(cert)

		tlsConf := &tls.Config{
			RootCAs:            certPool,
			InsecureSkipVerify: true,
		}
		client = resty.NewWithClient(&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConf,
			},
		})
	}
	return &WalmctlClient{
		client:  client,
		baseURL: protocol + hostURL + apiVersion,
	}, nil
}

func (c *WalmctlClient) ValidateHostConnect(hostURL string) error {
	timeout := time.Duration(5 * time.Second)
	_, err := net.DialTimeout("tcp", hostURL, timeout)
	if err != nil {
		return errors.Errorf("WalmServer %s unreachable, error: %s", hostURL, err.Error())
	}
	return nil
}

func (c *WalmctlClient) CreateTenantIfNotExist(namespace string) error {
	fullUrl := c.baseURL + "/tenant/" + namespace

	_, _ = c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody("{}").
		Post(fullUrl)

	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		Get(fullUrl)

	if err != nil || !resp.IsSuccess() {
		return errors.Errorf("create Tenant Error %v", err)
	}
	return nil
}

func (c *WalmctlClient) CreateSecret(namespace, secretName string, secretData map[string]string) error {
	_ = c.CreateTenantIfNotExist(namespace)
	secretFullUrl := c.baseURL + "/secret/" + namespace

	secretReq := k8sModel.CreateSecretRequestBody{
		Data: secretData,
		Type: "Opaque",
		Name: secretName,
	}
	resp, err := c.client.R().SetHeader("Content-Type", "application/json").
		SetBody(secretReq).
		Post(secretFullUrl)
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return errors.New(resp.String())
	}

	return nil
}

func (c *WalmctlClient) DeleteSecret(namespace, secretName string) error {
	_ = c.CreateTenantIfNotExist(namespace)
	secretFullUrl := c.baseURL + "/secret/" + namespace + "/name/" + secretName

	resp, err := c.client.R().SetHeader("Content-Type", "application/json").
		Delete(secretFullUrl)
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return errors.New(resp.String())
	}

	return nil
}

func (c *WalmctlClient) DryRunCreateRelease(
	namespace, chart string, releaseName string,
	configValues map[string]interface{},
) (*resty.Response, error) {
	fullUrl := c.baseURL + "/release/" + namespace + "/dryrun"

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
		chartFullUrl := c.baseURL + "/release/" + namespace + "/dryrun/withchart"
		resp, err = c.client.R().
			SetHeader("Content-Type", "multipart/form-data").
			SetFile("chart", chart).
			SetFormData(map[string]string{
				"namespace": namespace,
				"release":   releaseName,
				"body":      string(filestr[:]),
			}).
			Post(chartFullUrl)
	} else {
		resp, err = c.client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(filestr).
			Post(fullUrl)
	}
	if resp == nil || resp.StatusCode() != 200 {
		return nil, errors.New(fmt.Sprintf("error response %v %v", err, resp))
	}
	return resp, err
}

// release
func (c *WalmctlClient) CreateRelease(
	namespace, chart string, releaseName string,
	async bool, timeoutSec int64,
	configValues map[string]interface{},
) (*resty.Response, error) {
	_ = c.CreateTenantIfNotExist(namespace)
	fullUrl := c.baseURL + "/release/" + namespace + "?async=" + strconv.FormatBool(async) +
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
		chartFullUrl := c.baseURL + "/release/" + namespace + "/withchart"
		resp, err = c.client.R().
			SetHeader("Content-Type", "multipart/form-data").
			SetFile("chart", chart).
			SetFormData(map[string]string{
				"release": releaseName,
				"body":    string(filestr[:]),
			}).
			Post(chartFullUrl)
	} else {
		resp, err = c.client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(filestr).
			Post(fullUrl)
	}
	if resp == nil || resp.StatusCode() != 200 {
		return nil, errors.New(fmt.Sprintf("error response %v %v", err, resp))
	}
	return resp, err
}

func (c *WalmctlClient) GetRelease(namespace string, releaseName string) (resp *resty.Response, err error) {
	fullUrl := c.baseURL + "/release/" + namespace + "/name/" + releaseName

	resp, err = c.client.R().Get(fullUrl)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.String())
	}

	return resp, err
}

func (c *WalmctlClient) UpdateRelease(namespace string, newConfigStr string, async bool, timeoutSec int64) (resp *resty.Response, err error) {
	fullUrl := c.baseURL + "/release/" + namespace + "?async=" + strconv.FormatBool(async) +
		"&timeoutSec=" + strconv.FormatInt(timeoutSec, 10)

	resp, err = c.client.R().SetHeader("Content-Type", "application/json").
		SetBody(newConfigStr).
		Put(fullUrl)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.String())
	}
	return resp, err
}

func (c *WalmctlClient) UpdateReleaseWithChart(namespace string, releaseName string, file string, newConfigStr string) (resp *resty.Response, err error) {
	fullUrl := c.baseURL + "/release/" + namespace + "/withchart"

	resp, err = c.client.R().SetHeader("Content-Type", "multipart/form-data", ).
		SetFile("chart", file).
		SetFormData(map[string]string{"release": releaseName, "body": newConfigStr}).
		Put(fullUrl)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.String())
	}
	return resp, err
}

func (c *WalmctlClient) DeleteRelease(namespace string, releaseName string, async bool, timeoutSec int64, deletePvcs bool) (resp *resty.Response, err error) {
	fullUrl := c.baseURL + "/release/" + namespace + "/name/" + releaseName + "?async=" + strconv.FormatBool(async) +
		"&timeoutSec=" + strconv.FormatInt(timeoutSec, 10) + "&deletePvcs=" + strconv.FormatBool(deletePvcs)

	resp, err = c.client.R().
		Delete(fullUrl)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.String())
	}
	return resp, err

}

func (c *WalmctlClient) ListRelease(namespace string, labelSelector string) (resp *resty.Response, err error) {
	fullUrl := c.baseURL + "/release/" + namespace
	if namespace == "" {
		fullUrl = c.baseURL + "/release"
	}

	resp, err = c.client.R().
		SetHeader("Accept", "application/json").
		Get(fullUrl)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.String())
	}
	return resp, err
}

// project
func (c *WalmctlClient) CreateProject(
	namespace, chartPath, projectName string,
	async bool, timeoutSec int64,
	configValues map[string]interface{},
) (resp *resty.Response, err error) {
	_ = c.CreateTenantIfNotExist(namespace)
	fullUrl := c.baseURL + "/project/" + namespace + "/name/" + projectName + "?async=" + strconv.FormatBool(async) +
		"&timeoutSec=" + strconv.FormatInt(timeoutSec, 10)

	filestr, err := json.Marshal(configValues)
	if err != nil {
		klog.Errorf("marshal to json error %v", err)
	}

	resp, err = c.client.R().SetHeader("Content-Type", "application/json").
		SetBody(filestr).
		Post(fullUrl)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.String())
	}

	return resp, err
}

func (c *WalmctlClient) GetProject(namespace string, projectName string) (resp *resty.Response, err error) {
	fullUrl := c.baseURL + "/project/" + namespace + "/name/" + projectName
	resp, err = c.client.R().Get(fullUrl)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.String())
	}
	return resp, err
}

func (c *WalmctlClient) DeleteProject(namespace string, projectName string, async bool, timeoutSec int64, deletePvcs bool) (resp *resty.Response, err error) {
	fullUrl := c.baseURL + "/project/" + namespace + "/name/" + projectName + "?async=" + strconv.FormatBool(async) +
		"&timeoutSec=" + strconv.FormatInt(timeoutSec, 10) + "&deletePvcs=" + strconv.FormatBool(deletePvcs)

	resp, err = c.client.R().
		Delete(fullUrl)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.String())
	}

	return resp, err
}

func (c *WalmctlClient) ListProject(namespace string) (resp *resty.Response, err error) {
	fullUrl := c.baseURL + "/project/" + namespace
	if namespace == "" {
		fullUrl = c.baseURL + "/project"
	}

	resp, err = c.client.R().
		SetHeader("Accept", "application/json").
		Get(fullUrl)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.String())
	}

	return resp, err
}

func (c *WalmctlClient) AddReleaseInProject(namespace string, releaseName string, projectName string, async bool, timeoutSec int64, configValues map[string]interface{}) (resp *resty.Response, err error) {
	if releaseName != "" {
		releaseNameConfigs := make(map[string]interface{}, 0)
		releaseNameConfigs["name"] = releaseName
		util.MergeValues(configValues, releaseNameConfigs, false)
	}
	fileStr, err := json.Marshal(configValues)
	if err != nil {
		klog.Errorf("marshal to json error %v", err)
	}

	fullUrl := c.baseURL + "/project/" + namespace + "/name/" + projectName + "/instance?async=" + strconv.FormatBool(async) + "&timeoutSec=" + strconv.FormatInt(timeoutSec, 10)
	resp, err = c.client.R().SetHeader("Content-Type", "application/json").
		SetBody(fileStr).
		Post(fullUrl)

	return resp, err
}

func (c *WalmctlClient) DeleteReleaseInProject(namespace string, projectName string, releaseName string, async bool, timeoutSec int64, deletePvcs bool) (resp *resty.Response, err error) {
	fullUrl := c.baseURL + "/project/" + namespace + "/name/" + projectName + "/instance/" + releaseName + "?async=" + strconv.FormatBool(async) +
		"&timeoutSec=" + strconv.FormatInt(timeoutSec, 10) + "&deletePvcs=" + strconv.FormatBool(deletePvcs)

	resp, err = c.client.R().
		Delete(fullUrl)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.String())
	}

	return resp, err
}

func (c *WalmctlClient) DeleteTenant(namespace string) (resp *resty.Response, err error) {
	fullUrl := c.baseURL + "/tenant/" + namespace

	resp, err = c.client.R().
		Delete(fullUrl)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.String())
	}
	return resp, err
}

func (c *WalmctlClient) MigratePod(namespace string, podMig *k8sModel.PodMigRequest) (resp *resty.Response, err error) {

	podMigByte, err := json.Marshal(podMig)
	if err != nil {
		return nil, err
	}

	fullUrl := c.baseURL + "/crd/migration/pod/" + namespace

	resp, err = c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(string(podMigByte)).
		Post(fullUrl)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.String())
	}
	return resp, err
}

func (c *WalmctlClient) MigrateNode(nodeMig *k8sModel.NodeMigRequest) (resp *resty.Response, err error) {
	nodeMigByte, err := json.Marshal(nodeMig)
	if err != nil {
		return nil, err
	}

	fullUrl := c.baseURL + "/crd/migration/node"

	resp, err = c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(string(nodeMigByte)).
		Post(fullUrl)

	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.String())
	}
	return resp, err
}

func (c *WalmctlClient) GetPodMigration(namespace string, name string) (resp *resty.Response, err error) {
	fullUrl := c.baseURL + "/crd/migration/pod/" + namespace + "/name/" + name
	resp, err = c.client.R().
		Get(fullUrl)

	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.String())
	}
	return resp, err
}

func (c *WalmctlClient) DeletePodMigration(namespace string, name string) (resp *resty.Response, err error) {
	fullUrl := c.baseURL + "/crd/migration/pod/" + namespace + "/name/" + name
	resp, err = c.client.R().
		Delete(fullUrl)

	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, errors.Errorf(resp.String())
	}
	return resp, err
}

func (c *WalmctlClient) GetNodeMigration(name string) (resp *resty.Response, err error) {
	fullUrl := c.baseURL + "/crd/migration/node/" + name
	resp, err = c.client.R().
		Get(fullUrl)

	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, errors.Errorf(resp.String())
	}

	return resp, err
}

func (c *WalmctlClient) GetRepoList() (resp *resty.Response, err error) {
	//api/v1/chart/repolist
	fullUrl := c.baseURL + "/chart/repolist"
	resp, err = c.client.R().
		Get(fullUrl)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, errors.Errorf(resp.String())
	}
	return resp, err
}
