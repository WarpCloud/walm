package http

import (
	"testing"
	"net/http/httptest"
	"github.com/emicklei/go-restful"
	"net/http"
	"WarpCloud/walm/pkg/release/mocks"
	"github.com/stretchr/testify/assert"
	"errors"
	"encoding/json"
	"bytes"
	"WarpCloud/walm/pkg/models/release"
	"WarpCloud/walm/pkg/models/common"
	errorModel "WarpCloud/walm/pkg/models/error"
	"WarpCloud/walm/test/e2e/framework"
	"mime/multipart"
	"io/ioutil"
	"path/filepath"
	"github.com/stretchr/testify/mock"
	neturl "net/url"
	"os"
)

var mockUseCase *mocks.UseCase
var mockReleaseHandler ReleaseHandler

func TestReleaseHandler_DeleteRelease(t *testing.T) {
	tests := []struct {
		initMock   func()
		queryUrl   string
		statusCode int
	}{
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("DeleteRelease", "testns", "testname", false, false, int64(0)).Return(nil)
			},
			statusCode: 200,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("DeleteRelease", "testns", "testname", false, false, int64(0)).Return(errors.New(""))
			},
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("DeleteRelease", "testns", "testname", true, true, int64(60)).Return(nil)
			},
			queryUrl:   "?deletePvcs=true&async=true&timeoutSec=60",
			statusCode: 200,
		},
		{
			initMock: func() {
				refreshMockUseCase()
			},
			queryUrl:   "?deletePvcs=notvalid&async=true&timeoutSec=60",
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
			},
			queryUrl:   "?deletePvcs=true&async=notvalid&timeoutSec=60",
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
			},
			queryUrl:   "?deletePvcs=true&async=true&timeoutSec=notvalid",
			statusCode: 500,
		},
	}

	for _, test := range tests {
		test.initMock()
		url := releaseRootPath + "/testns/name/testname" + test.queryUrl

		httpRequest, _ := http.NewRequest("DELETE", url, nil)
		httpWriter := httptest.NewRecorder()
		restful.DefaultContainer.ServeHTTP(httpWriter, httpRequest)
		assert.Equal(t, httpWriter.Code, test.statusCode)
	}
}

func TestReleaseHandler_InstallRelease(t *testing.T) {
	tests := []struct {
		initMock   func()
		queryUrl   string
		body       interface{}
		statusCode int
	}{
		{
			initMock: func() {
				refreshMockUseCase()
			},
			body:       "notvalid",
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("InstallUpgradeRelease", "testns", &release.ReleaseRequestV2{}, ([]*common.BufferedFile)(nil), false, int64(0), (*bool)(nil)).Return(nil)
			},
			body:       release.ReleaseRequestV2{},
			statusCode: 200,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("InstallUpgradeRelease", "testns", &release.ReleaseRequestV2{}, ([]*common.BufferedFile)(nil), false, int64(0), (*bool)(nil)).Return(errors.New(""))
			},
			body:       release.ReleaseRequestV2{},
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("InstallUpgradeRelease", "testns", &release.ReleaseRequestV2{}, ([]*common.BufferedFile)(nil), true, int64(60), (*bool)(nil)).Return(nil)
			},
			queryUrl:   "?async=true&timeoutSec=60",
			body:       release.ReleaseRequestV2{},
			statusCode: 200,
		},
		{
			initMock: func() {
				refreshMockUseCase()
			},
			queryUrl:   "?async=notvalid&timeoutSec=60",
			body:       release.ReleaseRequestV2{},
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
			},
			queryUrl:   "?async=true&timeoutSec=notvalid",
			body:       release.ReleaseRequestV2{},
			statusCode: 500,
		},
	}

	for _, test := range tests {
		test.initMock()
		url := releaseRootPath + "/testns" + test.queryUrl

		bodyBytes, err := json.Marshal(test.body)
		assert.IsType(t, nil, err)

		httpRequest, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
		httpRequest.Header.Set("Content-Type", restful.MIME_JSON)
		httpWriter := httptest.NewRecorder()
		restful.DefaultContainer.ServeHTTP(httpWriter, httpRequest)
		assert.Equal(t, test.statusCode, httpWriter.Code)
	}
}

func TestReleaseHandler_InstallReleaseWithChart(t *testing.T) {
	currentFilePath, err := framework.GetCurrentFilePath()
	if err != nil {
		t.Fatal(err.Error())
	}

	tests := []struct {
		initMock    func()
		chartPath   string
		body        string
		releaseName string
		statusCode  int
	}{
		{
			initMock: func() {
				refreshMockUseCase()
			},
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
			},
			chartPath:  currentFilePath,
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
			},
			chartPath:  filepath.Join(filepath.Dir(currentFilePath), "../../../../test/resources/helm/tomcat-0.2.0.tgz"),
			body: "notvalid",
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("InstallUpgradeRelease", "testns",
					&release.ReleaseRequestV2{ReleaseRequest: release.ReleaseRequest{Name: "testname"}},
					mock.Anything, false, int64(0), (*bool)(nil)).Return(nil)
			},
			chartPath:   filepath.Join(filepath.Dir(currentFilePath), "../../../../test/resources/helm/tomcat-0.2.0.tgz"),
			body:        "{}",
			releaseName: "testname",
			statusCode:  200,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("InstallUpgradeRelease", "testns",
					&release.ReleaseRequestV2{ReleaseRequest: release.ReleaseRequest{Name: "testname"}},
					mock.Anything, false, int64(0), (*bool)(nil)).Return(errors.New(""))
			},
			chartPath:   filepath.Join(filepath.Dir(currentFilePath), "../../../../test/resources/helm/tomcat-0.2.0.tgz"),
			body:        "{}",
			releaseName: "testname",
			statusCode:  500,
		},
	}

	for _, test := range tests {
		test.initMock()
		url := releaseRootPath + "/testns/withchart"

		httpRequest, _ := http.NewRequest("POST", url, nil)
		httpRequest.Header.Set("Content-Type", "multipart/form-data")

		chartBytes := []byte{}
		var err error
		if test.chartPath != "" {
			chartBytes, err = ioutil.ReadFile(test.chartPath)
			if err != nil {
				t.Fatal(err.Error())
			}
		}

		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		{

			part, err := w.CreateFormFile("chart", "my-chart.tgz")
			if err != nil {
				t.Fatalf("CreateFormFile: %v", err)
			}
			part.Write(chartBytes)

			err = w.WriteField("release", test.releaseName)
			if err != nil {
				t.Fatalf("WriteField: %v", err)
			}
			part.Write([]byte(test.releaseName))

			err = w.WriteField("body", test.body)
			if err != nil {
				t.Fatalf("WriteField: %v", err)
			}
			part.Write([]byte(test.body))

			err = w.Close()
			if err != nil {
				t.Fatalf("Close: %v", err)
			}
		}

		r := multipart.NewReader(&b, w.Boundary())
		httpRequest.MultipartForm, err = r.ReadForm(0)
		if err != nil {
			t.Fatal(err.Error())
		}

		httpRequest.Form = neturl.Values(map[string][]string{"body": {test.body}, "release": {test.releaseName}})

		httpWriter := httptest.NewRecorder()
		restful.DefaultContainer.ServeHTTP(httpWriter, httpRequest)
		assert.Equal(t, test.statusCode, httpWriter.Code)

		mockUseCase.AssertExpectations(t)
	}
}

func TestReleaseHandler_UpgradeRelease(t *testing.T) {
	tests := []struct {
		initMock   func()
		queryUrl   string
		body       interface{}
		statusCode int
	}{
		{
			initMock: func() {
				refreshMockUseCase()
			},
			body:       "notvalid",
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("InstallUpgradeRelease", "testns", &release.ReleaseRequestV2{}, ([]*common.BufferedFile)(nil), false, int64(0), (*bool)(nil)).Return(nil)
			},
			body:       release.ReleaseRequestV2{},
			statusCode: 200,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("InstallUpgradeRelease", "testns", &release.ReleaseRequestV2{}, ([]*common.BufferedFile)(nil), false, int64(0), (*bool)(nil)).Return(errors.New(""))
			},
			body:       release.ReleaseRequestV2{},
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("InstallUpgradeRelease", "testns", &release.ReleaseRequestV2{}, ([]*common.BufferedFile)(nil), true, int64(60), (*bool)(nil)).Return(nil)
			},
			queryUrl:   "?async=true&timeoutSec=60",
			body:       release.ReleaseRequestV2{},
			statusCode: 200,
		},
		{
			initMock: func() {
				refreshMockUseCase()
			},
			queryUrl:   "?async=notvalid&timeoutSec=60",
			body:       release.ReleaseRequestV2{},
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
			},
			queryUrl:   "?async=true&timeoutSec=notvalid",
			body:       release.ReleaseRequestV2{},
			statusCode: 500,
		},
	}

	for _, test := range tests {
		test.initMock()
		url := releaseRootPath + "/testns" + test.queryUrl

		bodyBytes, err := json.Marshal(test.body)
		assert.IsType(t, nil, err)

		httpRequest, _ := http.NewRequest("PUT", url, bytes.NewBuffer(bodyBytes))
		httpRequest.Header.Set("Content-Type", restful.MIME_JSON)
		httpWriter := httptest.NewRecorder()
		restful.DefaultContainer.ServeHTTP(httpWriter, httpRequest)
		assert.Equal(t, test.statusCode, httpWriter.Code)
	}
}

func TestReleaseHandler_UpgradeReleaseWithChart(t *testing.T) {
	currentFilePath, err := framework.GetCurrentFilePath()
	if err != nil {
		t.Fatal(err.Error())
	}

	tests := []struct {
		initMock    func()
		chartPath   string
		body        string
		releaseName string
		statusCode  int
	}{
		{
			initMock: func() {
				refreshMockUseCase()
			},
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
			},
			chartPath:  currentFilePath,
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
			},
			chartPath:  filepath.Join(filepath.Dir(currentFilePath), "../../../../test/resources/helm/tomcat-0.2.0.tgz"),
			body: "notvalid",
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("InstallUpgradeRelease", "testns",
					&release.ReleaseRequestV2{ReleaseRequest: release.ReleaseRequest{Name: "testname"}},
					mock.Anything, false, int64(0), (*bool)(nil)).Return(nil)
			},
			chartPath:   filepath.Join(filepath.Dir(currentFilePath), "../../../../test/resources/helm/tomcat-0.2.0.tgz"),
			body:        "{}",
			releaseName: "testname",
			statusCode:  200,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("InstallUpgradeRelease", "testns",
					&release.ReleaseRequestV2{ReleaseRequest: release.ReleaseRequest{Name: "testname"}},
					mock.Anything, false, int64(0), (*bool)(nil)).Return(errors.New(""))
			},
			chartPath:   filepath.Join(filepath.Dir(currentFilePath), "../../../../test/resources/helm/tomcat-0.2.0.tgz"),
			body:        "{}",
			releaseName: "testname",
			statusCode:  500,
		},
	}

	for _, test := range tests {
		test.initMock()
		url := releaseRootPath + "/testns/withchart"

		httpRequest, _ := http.NewRequest("PUT", url, nil)
		httpRequest.Header.Set("Content-Type", "multipart/form-data")

		chartBytes := []byte{}
		var err error
		if test.chartPath != "" {
			chartBytes, err = ioutil.ReadFile(test.chartPath)
			if err != nil {
				t.Fatal(err.Error())
			}
		}

		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		{

			part, err := w.CreateFormFile("chart", "my-chart.tgz")
			if err != nil {
				t.Fatalf("CreateFormFile: %v", err)
			}
			part.Write(chartBytes)

			err = w.WriteField("release", test.releaseName)
			if err != nil {
				t.Fatalf("WriteField: %v", err)
			}
			part.Write([]byte(test.releaseName))

			err = w.WriteField("body", test.body)
			if err != nil {
				t.Fatalf("WriteField: %v", err)
			}
			part.Write([]byte(test.body))

			err = w.Close()
			if err != nil {
				t.Fatalf("Close: %v", err)
			}
		}

		r := multipart.NewReader(&b, w.Boundary())
		httpRequest.MultipartForm, err = r.ReadForm(0)
		if err != nil {
			t.Fatal(err.Error())
		}

		httpRequest.Form = neturl.Values(map[string][]string{"body": {test.body}, "release": {test.releaseName}})

		httpWriter := httptest.NewRecorder()
		restful.DefaultContainer.ServeHTTP(httpWriter, httpRequest)
		assert.Equal(t, test.statusCode, httpWriter.Code)

		mockUseCase.AssertExpectations(t)
	}
}

func TestReleaseHandler_ListRelease(t *testing.T) {
	tests := []struct {
		initMock   func()
		queryUrl   string
		statusCode int
	}{
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("ListReleases", "").Return(nil, errors.New(""))
			},
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("ListReleases", "").Return(nil, nil)
			},
			statusCode: 200,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("ListReleasesByLabels", "", "test=true").Return(nil, errors.New(""))
			},
			queryUrl:   "?labelselector=test=true",
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("ListReleasesByLabels", "", "test=true").Return(nil, nil)
			},
			queryUrl:   "?labelselector=test=true",
			statusCode: 200,
		},
	}

	for _, test := range tests {
		test.initMock()
		url := releaseRootPath + "/" + test.queryUrl

		httpRequest, _ := http.NewRequest("GET", url, nil)
		httpWriter := httptest.NewRecorder()
		restful.DefaultContainer.ServeHTTP(httpWriter, httpRequest)
		assert.Equal(t, test.statusCode, httpWriter.Code)
	}
}

func TestReleaseHandler_ListReleaseByNamespace(t *testing.T) {
	tests := []struct {
		initMock   func()
		queryUrl   string
		statusCode int
	}{
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("ListReleases", "testns").Return(nil, errors.New(""))
			},
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("ListReleases", "testns").Return(nil, nil)
			},
			statusCode: 200,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("ListReleasesByLabels", "testns", "test=true").Return(nil, errors.New(""))
			},
			queryUrl:   "?labelselector=test=true",
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("ListReleasesByLabels", "testns", "test=true").Return(nil, nil)
			},
			queryUrl:   "?labelselector=test=true",
			statusCode: 200,
		},
	}

	for _, test := range tests {
		test.initMock()
		url := releaseRootPath + "/testns" + test.queryUrl

		httpRequest, _ := http.NewRequest("GET", url, nil)
		httpWriter := httptest.NewRecorder()
		restful.DefaultContainer.ServeHTTP(httpWriter, httpRequest)
		assert.Equal(t, test.statusCode, httpWriter.Code)
	}
}

func TestReleaseHandler_GetRelease(t *testing.T) {
	tests := []struct {
		initMock   func()
		statusCode int
	}{
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("GetRelease", "testns", "testname").Return(nil, errors.New(""))
			},
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("GetRelease", "testns", "testname").Return(nil, errorModel.NotFoundError{})
			},
			statusCode: 404,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("GetRelease", "testns", "testname").Return(nil, nil)
			},
			statusCode: 200,
		},
	}

	for _, test := range tests {
		test.initMock()
		url := releaseRootPath + "/testns/name/testname"

		httpRequest, _ := http.NewRequest("GET", url, nil)
		httpWriter := httptest.NewRecorder()
		restful.DefaultContainer.ServeHTTP(httpWriter, httpRequest)
		assert.Equal(t, test.statusCode, httpWriter.Code)
	}
}

func TestReleaseHandler_DryRunRelease(t *testing.T) {
	tests := []struct {
		initMock   func()
		body       interface{}
		statusCode int
	}{
		{
			initMock: func() {
				refreshMockUseCase()
			},
			body:       "notvalid",
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("DryRunRelease", "testns", &release.ReleaseRequestV2{}, ([]*common.BufferedFile)(nil)).Return(nil, nil)
			},
			body:       release.ReleaseRequestV2{},
			statusCode: 200,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("DryRunRelease", "testns", &release.ReleaseRequestV2{}, ([]*common.BufferedFile)(nil)).Return(nil, errors.New(""))
			},
			body:       release.ReleaseRequestV2{},
			statusCode: 500,
		},
	}

	for _, test := range tests {
		test.initMock()
		url := releaseRootPath + "/testns/dryrun"

		bodyBytes, err := json.Marshal(test.body)
		assert.IsType(t, nil, err)

		httpRequest, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
		httpRequest.Header.Set("Content-Type", restful.MIME_JSON)
		httpWriter := httptest.NewRecorder()
		restful.DefaultContainer.ServeHTTP(httpWriter, httpRequest)
		assert.Equal(t, test.statusCode, httpWriter.Code)
	}
}

func TestReleaseHandler_DryRunReleaseWithChart(t *testing.T) {
	currentFilePath, err := framework.GetCurrentFilePath()
	if err != nil {
		t.Fatal(err.Error())
	}

	tests := []struct {
		initMock    func()
		chartPath   string
		body        string
		releaseName string
		statusCode  int
	}{
		{
			initMock: func() {
				refreshMockUseCase()
			},
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
			},
			chartPath:  currentFilePath,
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
			},
			chartPath:  filepath.Join(filepath.Dir(currentFilePath), "../../../../test/resources/helm/tomcat-0.2.0.tgz"),
			body: "notvalid",
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("DryRunRelease", "testns",
					&release.ReleaseRequestV2{ReleaseRequest: release.ReleaseRequest{Name: "testname"}},
					mock.Anything).Return(nil,nil)
			},
			chartPath:   filepath.Join(filepath.Dir(currentFilePath), "../../../../test/resources/helm/tomcat-0.2.0.tgz"),
			body:        "{}",
			releaseName: "testname",
			statusCode:  200,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("DryRunRelease", "testns",
					&release.ReleaseRequestV2{ReleaseRequest: release.ReleaseRequest{Name: "testname"}},
					mock.Anything).Return(nil,errors.New(""))
			},
			chartPath:   filepath.Join(filepath.Dir(currentFilePath), "../../../../test/resources/helm/tomcat-0.2.0.tgz"),
			body:        "{}",
			releaseName: "testname",
			statusCode:  500,
		},
	}

	for _, test := range tests {
		test.initMock()
		url := releaseRootPath + "/testns/dryrun/withchart"

		httpRequest, _ := http.NewRequest("POST", url, nil)
		httpRequest.Header.Set("Content-Type", "multipart/form-data")

		chartBytes := []byte{}
		var err error
		if test.chartPath != "" {
			chartBytes, err = ioutil.ReadFile(test.chartPath)
			if err != nil {
				t.Fatal(err.Error())
			}
		}

		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		{

			part, err := w.CreateFormFile("chart", "my-chart.tgz")
			if err != nil {
				t.Fatalf("CreateFormFile: %v", err)
			}
			part.Write(chartBytes)

			err = w.WriteField("release", test.releaseName)
			if err != nil {
				t.Fatalf("WriteField: %v", err)
			}
			part.Write([]byte(test.releaseName))

			err = w.WriteField("body", test.body)
			if err != nil {
				t.Fatalf("WriteField: %v", err)
			}
			part.Write([]byte(test.body))

			err = w.Close()
			if err != nil {
				t.Fatalf("Close: %v", err)
			}
		}

		r := multipart.NewReader(&b, w.Boundary())
		httpRequest.MultipartForm, err = r.ReadForm(0)
		if err != nil {
			t.Fatal(err.Error())
		}

		httpRequest.Form = neturl.Values(map[string][]string{"body": {test.body}, "release": {test.releaseName}})

		httpWriter := httptest.NewRecorder()
		restful.DefaultContainer.ServeHTTP(httpWriter, httpRequest)
		assert.Equal(t, test.statusCode, httpWriter.Code)

		mockUseCase.AssertExpectations(t)
	}
}

func TestReleaseHandler_ComputeResourcesByDryRunRelease(t *testing.T) {
	tests := []struct {
		initMock   func()
		body       interface{}
		statusCode int
	}{
		{
			initMock: func() {
				refreshMockUseCase()
			},
			body:       "notvalid",
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("ComputeResourcesByDryRunRelease", "testns", &release.ReleaseRequestV2{}, ([]*common.BufferedFile)(nil)).Return(nil, nil)
			},
			body:       release.ReleaseRequestV2{},
			statusCode: 200,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("ComputeResourcesByDryRunRelease", "testns", &release.ReleaseRequestV2{}, ([]*common.BufferedFile)(nil)).Return(nil, errors.New(""))
			},
			body:       release.ReleaseRequestV2{},
			statusCode: 500,
		},
	}

	for _, test := range tests {
		test.initMock()
		url := releaseRootPath + "/testns/dryrun/resources"

		bodyBytes, err := json.Marshal(test.body)
		assert.IsType(t, nil, err)

		httpRequest, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
		httpRequest.Header.Set("Content-Type", restful.MIME_JSON)
		httpWriter := httptest.NewRecorder()
		restful.DefaultContainer.ServeHTTP(httpWriter, httpRequest)
		assert.Equal(t, test.statusCode, httpWriter.Code)
	}
}

func TestReleaseHandler_ComputeResourcesByDryRunReleaseWithChart(t *testing.T) {
	currentFilePath, err := framework.GetCurrentFilePath()
	if err != nil {
		t.Fatal(err.Error())
	}

	tests := []struct {
		initMock    func()
		chartPath   string
		body        string
		releaseName string
		statusCode  int
	}{
		{
			initMock: func() {
				refreshMockUseCase()
			},
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
			},
			chartPath:  currentFilePath,
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
			},
			chartPath:  filepath.Join(filepath.Dir(currentFilePath), "../../../../test/resources/helm/tomcat-0.2.0.tgz"),
			body: "notvalid",
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("ComputeResourcesByDryRunRelease", "testns",
					&release.ReleaseRequestV2{ReleaseRequest: release.ReleaseRequest{Name: "testname"}},
					mock.Anything).Return(nil,nil)
			},
			chartPath:   filepath.Join(filepath.Dir(currentFilePath), "../../../../test/resources/helm/tomcat-0.2.0.tgz"),
			body:        "{}",
			releaseName: "testname",
			statusCode:  200,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("ComputeResourcesByDryRunRelease", "testns",
					&release.ReleaseRequestV2{ReleaseRequest: release.ReleaseRequest{Name: "testname"}},
					mock.Anything).Return(nil,errors.New(""))
			},
			chartPath:   filepath.Join(filepath.Dir(currentFilePath), "../../../../test/resources/helm/tomcat-0.2.0.tgz"),
			body:        "{}",
			releaseName: "testname",
			statusCode:  500,
		},
	}

	for _, test := range tests {
		test.initMock()
		url := releaseRootPath + "/testns/dryrun/withchart/resources"

		httpRequest, _ := http.NewRequest("POST", url, nil)
		httpRequest.Header.Set("Content-Type", "multipart/form-data")

		chartBytes := []byte{}
		var err error
		if test.chartPath != "" {
			chartBytes, err = ioutil.ReadFile(test.chartPath)
			if err != nil {
				t.Fatal(err.Error())
			}
		}

		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		{

			part, err := w.CreateFormFile("chart", "my-chart.tgz")
			if err != nil {
				t.Fatalf("CreateFormFile: %v", err)
			}
			part.Write(chartBytes)

			err = w.WriteField("release", test.releaseName)
			if err != nil {
				t.Fatalf("WriteField: %v", err)
			}
			part.Write([]byte(test.releaseName))

			err = w.WriteField("body", test.body)
			if err != nil {
				t.Fatalf("WriteField: %v", err)
			}
			part.Write([]byte(test.body))

			err = w.Close()
			if err != nil {
				t.Fatalf("Close: %v", err)
			}
		}

		r := multipart.NewReader(&b, w.Boundary())
		httpRequest.MultipartForm, err = r.ReadForm(0)
		if err != nil {
			t.Fatal(err.Error())
		}

		httpRequest.Form = neturl.Values(map[string][]string{"body": {test.body}, "release": {test.releaseName}})

		httpWriter := httptest.NewRecorder()
		restful.DefaultContainer.ServeHTTP(httpWriter, httpRequest)
		assert.Equal(t, test.statusCode, httpWriter.Code)

		mockUseCase.AssertExpectations(t)
	}
}

func TestReleaseHandler_PauseRelease(t *testing.T) {
	tests := []struct {
		initMock   func()
		queryUrl   string
		statusCode int
	}{
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("PauseRelease", "testns", "testname", false, int64(0)).Return(nil)
			},
			statusCode: 200,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("PauseRelease", "testns", "testname", false, int64(0)).Return(errors.New(""))
			},
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("PauseRelease", "testns", "testname", true, int64(60)).Return(nil)
			},
			queryUrl:   "?async=true&timeoutSec=60",
			statusCode: 200,
		},
		{
			initMock: func() {
				refreshMockUseCase()
			},
			queryUrl:   "?async=notvalid&timeoutSec=60",
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
			},
			queryUrl:   "?async=true&timeoutSec=notvalid",
			statusCode: 500,
		},
	}

	for _, test := range tests {
		test.initMock()
		url := releaseRootPath + "/testns/name/testname/pause" + test.queryUrl

		httpRequest, _ := http.NewRequest("POST", url, nil)
		httpRequest.Header.Set("Content-Type", restful.MIME_JSON)
		httpWriter := httptest.NewRecorder()
		restful.DefaultContainer.ServeHTTP(httpWriter, httpRequest)
		assert.Equal(t, test.statusCode, httpWriter.Code)
	}
}

func TestReleaseHandler_RecoverRelease(t *testing.T) {
	tests := []struct {
		initMock   func()
		queryUrl   string
		statusCode int
	}{
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("RecoverRelease", "testns", "testname", false, int64(0)).Return(nil)
			},
			statusCode: 200,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("RecoverRelease", "testns", "testname", false, int64(0)).Return(errors.New(""))
			},
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("RecoverRelease", "testns", "testname", true, int64(60)).Return(nil)
			},
			queryUrl:   "?async=true&timeoutSec=60",
			statusCode: 200,
		},
		{
			initMock: func() {
				refreshMockUseCase()
			},
			queryUrl:   "?async=notvalid&timeoutSec=60",
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
			},
			queryUrl:   "?async=true&timeoutSec=notvalid",
			statusCode: 500,
		},
	}

	for _, test := range tests {
		test.initMock()
		url := releaseRootPath + "/testns/name/testname/recover" + test.queryUrl

		httpRequest, _ := http.NewRequest("POST", url, nil)
		httpRequest.Header.Set("Content-Type", restful.MIME_JSON)
		httpWriter := httptest.NewRecorder()
		restful.DefaultContainer.ServeHTTP(httpWriter, httpRequest)
		assert.Equal(t, test.statusCode, httpWriter.Code)
	}
}

func TestReleaseHandler_RestartRelease(t *testing.T) {
	tests := []struct {
		initMock   func()
		queryUrl   string
		statusCode int
	}{
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("RestartRelease", "testns", "testname").Return(nil)
			},
			statusCode: 200,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("RestartRelease", "testns", "testname").Return(errors.New(""))
			},
			statusCode: 500,
		},
	}

	for _, test := range tests {
		test.initMock()
		url := releaseRootPath + "/testns/name/testname/restart" + test.queryUrl

		httpRequest, _ := http.NewRequest("POST", url, nil)
		httpRequest.Header.Set("Content-Type", restful.MIME_JSON)
		httpWriter := httptest.NewRecorder()
		restful.DefaultContainer.ServeHTTP(httpWriter, httpRequest)
		assert.Equal(t, test.statusCode, httpWriter.Code)
	}
}

func TestMain(m *testing.M) {
	restful.Add(RegisterReleaseHandler(&mockReleaseHandler))
	os.Exit(m.Run())
}

func refreshMockUseCase() {
	mockUseCase = &mocks.UseCase{}
	mockReleaseHandler.usecase = mockUseCase
}