package guardianclient

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty"
	"github.com/pkg/errors"
)

const APIV1 = "/api/v1"

type GuardianLogin struct {
	Password string `json:"password"`
	Username string `json:"username"`
	IsSystem bool   `json:"isSystem"`
}

type GuardianPrincipals struct {
	Principals []string `json:"principals"`
}

type GuardianClient struct {
	baseURL    string
	Password   string `json:"password"`
	Username   string `json:"username"`
	restClient *resty.Client
}

type GuardianErrorResponse struct {
	ReturnCode    int    `json:"returnCode"`
	ErrorMessage  string `json:"errorMessage"`
	DetailMessage string `json:"detailMessage"`
}

type GuardianUsers struct {
	UserName        string `json:"userName"`
	FullName        string `json:"fullName"`
	UserDept        string `json:"userDept"`
	UserDescription string `json:"userDescription"`
	UserLocked      bool   `json:"userLocked"`
	RememberMe      bool   `json:"rememberMe"`
	GidNumber       string `json:"gidNumber"`
	UidNumber       string `json:"uidNumber"`
}

var guardianClient *GuardianClient

func NewClient(baseURL, username, password string) *GuardianClient {
	if guardianClient == nil {
		guardianClient = &GuardianClient{
			baseURL:  baseURL,
			Username: username,
			Password: password,
		}
	}
	return guardianClient
}

func (c *GuardianClient) Login() error {
	fullUrl := c.baseURL + APIV1 + "/login"
	guardianLogin := GuardianLogin{
		Username: c.Username,
		Password: c.Password,
		IsSystem: false,
	}
	bodyStr, _ := json.Marshal(guardianLogin)

	loginClient := resty.New()
	loginClient.SetTLSClientConfig(&tls.Config{
		InsecureSkipVerify: true,
	})
	resp, err := loginClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(bodyStr).
		Post(fullUrl)
	if err != nil {
		return err
	}
	if !resp.IsSuccess() {
		errResp := GuardianErrorResponse{}
		err = json.Unmarshal(resp.Body(), &errResp)
		return errors.New(fmt.Sprintf("Login %s Error, errResp %v error %v resp %s", fullUrl, errResp, err, resp.Body()))
	}

	client := resty.New()
	client.SetCookies(resp.Cookies())
	client.SetTLSClientConfig(&tls.Config{
		InsecureSkipVerify: true,
	})
	c.restClient = client
	return nil
}

func (c *GuardianClient) GetUsers() ([]GuardianUsers, error) {
	users := make([]GuardianUsers, 0)
	fullUrl := c.baseURL + APIV1 + "/users"

	if c.restClient == nil {
		if err := c.Login(); err != nil {
			return users, err
		}
	}
	resp, err := c.restClient.R().
		SetHeader("Content-Type", "application/json").
		Get(fullUrl)
	if err != nil {
		return users, err
	}
	if !resp.IsSuccess() {
		errResp := GuardianErrorResponse{}
		_ = json.Unmarshal(resp.Body(), &errResp)
		return users, errors.New(fmt.Sprintf("GetUsers Error, errResp %v", errResp))
	}
	err = json.Unmarshal(resp.Body(), &users)
	if err != nil {
		return users, err
	}
	return users, nil
}

func (c *GuardianClient) GetMultipleKeytabs(principals []string) ([]byte, error) {
	fullUrl := c.baseURL + APIV1 + "/users/multiple/keytabs"
	guardianPrincipals := GuardianPrincipals{}
	guardianPrincipals.Principals = principals

	if c.restClient == nil {
		if err := c.Login(); err != nil {
			return []byte{}, err
		}
	}
	resp, err := c.restClient.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/octet-stream").
		SetBody(&guardianPrincipals).
		Post(fullUrl)
	if err != nil {
		return []byte{}, err
	}
	if !resp.IsSuccess() {
		errResp := GuardianErrorResponse{}
		err = json.Unmarshal(resp.Body(), &errResp)
		return []byte{}, errors.New(fmt.Sprintf("GetMultipleKeytabs Error, errResp %v err %v body %s", errResp, err, resp.Body()))
	}
	return resp.Body(), nil
}
