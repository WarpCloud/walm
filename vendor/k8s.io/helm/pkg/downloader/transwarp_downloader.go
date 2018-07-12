package downloader

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	"net/url"

	"github.com/Masterminds/semver"
	"github.com/levigross/grequests"
	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/provenance"
)

type ChartInfo struct {
	Name        string `json:"name"`
	ChartId     int    `json:"chart_id"`
	ChartName   string `json:"chart_name"`
	ProjectId   int    `json:"project_id"`
	ProjectName string `json:"project_name"`
	Type        int    `json:"type"`
	ShortDesc   string `json:"short_desc"`
	Description string `json:"description"`
	Role        int    `json:"role"`
}

type ChartEntityInfo struct {
	Id      int    `json:"id"`
	Chart   int    `json:"chart"`
	Version string `json:"version"`
	File    string `json:"file"`
}

type TranswarpDownloader struct {
	// Out is the location to write warning and info messages.
	Out io.Writer
	// Verify indicates what verification strategy to use.
	Verify VerificationStrategy

	Username string
	Password string

	// HelmHome is the $HELM_HOME.
	HelmHome helmpath.Home
	// Getter collection for the operation
	Getters getter.Providers
}

func (c *TranswarpDownloader) DownloadTo(ref, version, dest string) (string, *provenance.Verification, error) {
	// Do some validate and clean work
	refSplit := strings.SplitN(ref, "/", 2)
	address := refSplit[0]
	if !strings.HasPrefix(address, "http://") || !strings.HasPrefix(address, "https://") {
		address = fmt.Sprintf("http://%s", address)
	}
	if !strings.HasSuffix(address, "/") {
		address = fmt.Sprintf("%s/", address)
	}
	_, err := url.Parse(address)
	if err != nil {
		return "", nil, err
	}

	// get the logged in session
	session, err := login(address, c.Username, c.Password)
	if err != nil {
		return "", nil, err
	}

	// get all chart info
	chartName := refSplit[1]
	listUrl := fmt.Sprintf("%sapi/chart/", address)
	resp, err := session.Get(listUrl, nil)
	if err != nil {
		return "", nil, err
	}
	if !resp.Ok {
		return "", nil, resp.Error
	}
	var chartList []*ChartInfo
	err = resp.JSON(&chartList)
	if err != nil {
		return "", nil, err
	}
	found := false
	var chartId int
	for _, ch := range chartList {
		if ch.Name == chartName {
			chartId = ch.ChartId
			found = true
			break
		}
	}
	if !found {
		return "", nil, errors.New("chart not found in chart repo or you don't have access to it")
	}

	// get all chart entity info
	chartEntityListUrl := fmt.Sprintf("%sapi/chart/%d/entity/", address, chartId)
	chartEntityListResp, err := session.Get(chartEntityListUrl, nil)
	if err != nil {
		return "", nil, err
	}
	if !chartEntityListResp.Ok {
		return "", nil, chartEntityListResp.Error
	}

	var chartEntityList []*ChartEntityInfo
	err = chartEntityListResp.JSON(&chartEntityList)
	if err != nil {
		return "", nil, err
	}
	var downloadUrl string
	var chartVersion string
	for _, ch := range chartEntityList {
		// if specify version
		if version != "" {
			if ch.Version == version {
				downloadUrl = ch.File
				chartVersion = version
				break
			}
		} else {
			// get the latest version
			if chartVersion == "" {
				chartVersion = ch.Version
				downloadUrl = ch.File
			} else {
				chVersion, err := semver.NewVersion(ch.Version)
				if err != nil {
					return "", nil, err
				}
				currentVersion, err := semver.NewVersion(chartVersion)
				if err != nil {
					return "", nil, err
				}
				if chVersion.GreaterThan(currentVersion) {
					chartVersion = ch.Version
					downloadUrl = ch.File
				}
			}
		}
	}
	if downloadUrl == "" {
		return "", nil, errors.New("required chart not found")
	}

	// download the chart
	downloadResp, err := session.Get(downloadUrl, nil)
	if err != nil {
		return "", nil, err
	}
	if !downloadResp.Ok {
		return "", nil, downloadResp.Error
	}

	name := filepath.Base(downloadUrl)
	destfile := filepath.Join(dest, name)
	if err := ioutil.WriteFile(destfile, downloadResp.Bytes(), 0644); err != nil {
		return destfile, nil, err
	}
	return destfile, nil, nil
}

func login(address, username, password string) (*grequests.Session, error) {
	loginUrl := fmt.Sprintf("%sapi/auth/login/", address)
	session := grequests.NewSession(nil)
	if username != "" && password != "" {
		data := make(map[string]string)
		data["username"] = username
		data["password"] = password
		resp, err := session.Post(loginUrl, &grequests.RequestOptions{Data: data})
		if err != nil {
			return nil, err
		}
		if !resp.Ok {
			return nil, errors.New(resp.String())
		}
	}
	return session, nil
}
