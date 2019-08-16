package http

import (
	"testing"
	"net/http/httptest"
	"github.com/emicklei/go-restful"
	"net/http"
	"WarpCloud/walm/pkg/project/mocks"
	"github.com/stretchr/testify/assert"
	"errors"
	errorModel "WarpCloud/walm/pkg/models/error"
	"encoding/json"
	"bytes"
	"WarpCloud/walm/pkg/models/project"
	"WarpCloud/walm/pkg/models/release"
	"github.com/stretchr/testify/mock"
)

func TestProjectHandler_ListProject(t *testing.T) {
	var mockUseCase *mocks.UseCase
	var mockProjectHandler ProjectHandler

	container := restful.NewContainer()
	container.Add(RegisterProjectHandler(&mockProjectHandler))

	refreshMockUseCase := func() {
		mockUseCase = &mocks.UseCase{}
		mockProjectHandler.usecase = mockUseCase
	}
	tests := []struct {
		initMock   func()
		statusCode int
	}{
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("ListProjects", "").Return(nil, nil)
			},
			statusCode: 200,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("ListProjects", "").Return(nil, errors.New(""))
			},
			statusCode: 500,
		},
	}

	for _, test := range tests {
		test.initMock()
		url := projectRootPath + "/"

		httpRequest, _ := http.NewRequest("GET", url, nil)
		httpWriter := httptest.NewRecorder()
		container.ServeHTTP(httpWriter, httpRequest)
		assert.Equal(t, httpWriter.Code, test.statusCode)
	}
}

func TestProjectHandler_ListProjectByNamespace(t *testing.T) {
	var mockUseCase *mocks.UseCase
	var mockProjectHandler ProjectHandler

	container := restful.NewContainer()
	container.Add(RegisterProjectHandler(&mockProjectHandler))

	refreshMockUseCase := func() {
		mockUseCase = &mocks.UseCase{}
		mockProjectHandler.usecase = mockUseCase
	}
	tests := []struct {
		initMock   func()
		statusCode int
	}{
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("ListProjects", "testns").Return(nil, nil)
			},
			statusCode: 200,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("ListProjects", "testns").Return(nil, errors.New(""))
			},
			statusCode: 500,
		},
	}

	for _, test := range tests {
		test.initMock()
		url := projectRootPath + "/testns"

		httpRequest, _ := http.NewRequest("GET", url, nil)
		httpWriter := httptest.NewRecorder()
		container.ServeHTTP(httpWriter, httpRequest)
		assert.Equal(t, httpWriter.Code, test.statusCode)
	}
}

func TestProjectHandler_GetProjectInfo(t *testing.T) {
	var mockUseCase *mocks.UseCase
	var mockProjectHandler ProjectHandler

	container := restful.NewContainer()
	container.Add(RegisterProjectHandler(&mockProjectHandler))

	refreshMockUseCase := func() {
		mockUseCase = &mocks.UseCase{}
		mockProjectHandler.usecase = mockUseCase
	}
	tests := []struct {
		initMock   func()
		statusCode int
	}{
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("GetProjectInfo", "testns", "testnm").Return(nil, nil)
			},
			statusCode: 200,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("GetProjectInfo", "testns", "testnm").Return(nil, errors.New(""))
			},
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("GetProjectInfo", "testns", "testnm").Return(nil, errorModel.NotFoundError{})
			},
			statusCode: 404,
		},
	}

	for _, test := range tests {
		test.initMock()
		url := projectRootPath + "/testns/name/testnm"

		httpRequest, _ := http.NewRequest("GET", url, nil)
		httpWriter := httptest.NewRecorder()
		container.ServeHTTP(httpWriter, httpRequest)
		assert.Equal(t, httpWriter.Code, test.statusCode)
	}
}

func TestProjectHandler_CreateProject(t *testing.T) {
	var mockUseCase *mocks.UseCase
	var mockProjectHandler ProjectHandler

	container := restful.NewContainer()
	container.Add(RegisterProjectHandler(&mockProjectHandler))

	refreshMockUseCase := func() {
		mockUseCase = &mocks.UseCase{}
		mockProjectHandler.usecase = mockUseCase
	}
	tests := []struct {
		initMock   func()
		body       interface{}
		queryUrl   string
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
			},
			body:       &project.ProjectParams{},
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("CreateProject", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New(""))
			},
			body: &project.ProjectParams{
				Releases: []*release.ReleaseRequestV2{
					{
						ReleaseRequest: release.ReleaseRequest{},
					},
				},
			},
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("CreateProject", "testns", "testnm", mock.Anything, false, int64(0)).Return(nil)
			},
			body: &project.ProjectParams{
				Releases: []*release.ReleaseRequestV2{
					{
						ReleaseRequest: release.ReleaseRequest{},
					},
				},
			},
			statusCode: 200,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("CreateProject", "testns", "testnm", mock.Anything, true, int64(60)).Return(nil)
			},
			body: &project.ProjectParams{
				Releases: []*release.ReleaseRequestV2{
					{
						ReleaseRequest: release.ReleaseRequest{},
					},
				},
			},
			queryUrl:   "?async=true&timeoutSec=60",
			statusCode: 200,
		},
		{
			initMock: func() {
				refreshMockUseCase()
			},
			body: &project.ProjectParams{
				Releases: []*release.ReleaseRequestV2{
					{
						ReleaseRequest: release.ReleaseRequest{},
					},
				},
			},
			queryUrl:   "?async=true&timeoutSec=notvalid",
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
			},
			body: &project.ProjectParams{
				Releases: []*release.ReleaseRequestV2{
					{
						ReleaseRequest: release.ReleaseRequest{},
					},
				},
			},
			queryUrl:   "?async=notvalid&timeoutSec=60",
			statusCode: 500,
		},
	}

	for _, test := range tests {
		test.initMock()
		url := projectRootPath + "/testns/name/testnm" + test.queryUrl

		bodyBytes, err := json.Marshal(test.body)
		assert.IsType(t, nil, err)

		httpRequest, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
		httpRequest.Header.Set("Content-Type", restful.MIME_JSON)
		httpWriter := httptest.NewRecorder()
		container.ServeHTTP(httpWriter, httpRequest)
		assert.Equal(t, httpWriter.Code, test.statusCode)
	}
}

func TestProjectHandler_DeleteProject(t *testing.T) {
	var mockUseCase *mocks.UseCase
	var mockProjectHandler ProjectHandler

	container := restful.NewContainer()
	container.Add(RegisterProjectHandler(&mockProjectHandler))

	refreshMockUseCase := func() {
		mockUseCase = &mocks.UseCase{}
		mockProjectHandler.usecase = mockUseCase
	}
	tests := []struct {
		initMock   func()
		queryUrl   string
		statusCode int
	}{
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("DeleteProject", "testns", "testname", false, int64(0), false).Return(nil)
			},
			statusCode: 200,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("DeleteProject", "testns", "testname", false, int64(0), false).Return(errors.New(""))
			},
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("DeleteProject", "testns", "testname", true, int64(60), true).Return(nil)
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
		url := projectRootPath + "/testns/name/testname" + test.queryUrl

		httpRequest, _ := http.NewRequest("DELETE", url, nil)
		httpWriter := httptest.NewRecorder()
		container.ServeHTTP(httpWriter, httpRequest)
		assert.Equal(t, httpWriter.Code, test.statusCode)
	}
}

func TestProjectHandler_AddReleaseInProject(t *testing.T) {
	var mockUseCase *mocks.UseCase
	var mockProjectHandler ProjectHandler

	container := restful.NewContainer()
	container.Add(RegisterProjectHandler(&mockProjectHandler))

	refreshMockUseCase := func() {
		mockUseCase = &mocks.UseCase{}
		mockProjectHandler.usecase = mockUseCase
	}
	tests := []struct {
		initMock   func()
		body       interface{}
		queryUrl   string
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
				mockUseCase.On("AddReleasesInProject", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New(""))
			},
			body:       &release.ReleaseRequestV2{},
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("AddReleasesInProject", "testns", "testnm", mock.Anything, false, int64(0)).Return(nil)
			},
			body:       &release.ReleaseRequestV2{},
			statusCode: 200,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("AddReleasesInProject", "testns", "testnm", mock.Anything, true, int64(60)).Return(nil)
			},
			body:       &release.ReleaseRequestV2{},
			queryUrl:   "?async=true&timeoutSec=60",
			statusCode: 200,
		},
		{
			initMock: func() {
				refreshMockUseCase()
			},
			body:       &release.ReleaseRequestV2{},
			queryUrl:   "?async=true&timeoutSec=notvalid",
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
			},
			body:       &release.ReleaseRequestV2{},
			queryUrl:   "?async=notvalid&timeoutSec=60",
			statusCode: 500,
		},
	}

	for _, test := range tests {
		test.initMock()
		url := projectRootPath + "/testns/name/testnm/instance" + test.queryUrl

		bodyBytes, err := json.Marshal(test.body)
		assert.IsType(t, nil, err)

		httpRequest, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
		httpRequest.Header.Set("Content-Type", restful.MIME_JSON)
		httpWriter := httptest.NewRecorder()
		container.ServeHTTP(httpWriter, httpRequest)
		assert.Equal(t, httpWriter.Code, test.statusCode)
	}
}

func TestProjectHandler_UpgradeReleaseInProject(t *testing.T) {
	var mockUseCase *mocks.UseCase
	var mockProjectHandler ProjectHandler

	container := restful.NewContainer()
	container.Add(RegisterProjectHandler(&mockProjectHandler))

	refreshMockUseCase := func() {
		mockUseCase = &mocks.UseCase{}
		mockProjectHandler.usecase = mockUseCase
	}
	tests := []struct {
		initMock   func()
		body       interface{}
		queryUrl   string
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
				mockUseCase.On("UpgradeReleaseInProject", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New(""))
			},
			body:       &release.ReleaseRequestV2{},
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("UpgradeReleaseInProject", "testns", "testnm", mock.Anything, false, int64(0)).Return(nil)
			},
			body:       &release.ReleaseRequestV2{},
			statusCode: 200,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("UpgradeReleaseInProject", "testns", "testnm", mock.Anything, true, int64(60)).Return(nil)
			},
			body:       &release.ReleaseRequestV2{},
			queryUrl:   "?async=true&timeoutSec=60",
			statusCode: 200,
		},
		{
			initMock: func() {
				refreshMockUseCase()
			},
			body:       &release.ReleaseRequestV2{},
			queryUrl:   "?async=true&timeoutSec=notvalid",
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
			},
			body:       &release.ReleaseRequestV2{},
			queryUrl:   "?async=notvalid&timeoutSec=60",
			statusCode: 500,
		},
	}

	for _, test := range tests {
		test.initMock()
		url := projectRootPath + "/testns/name/testnm/instance" + test.queryUrl

		bodyBytes, err := json.Marshal(test.body)
		assert.IsType(t, nil, err)

		httpRequest, _ := http.NewRequest("PUT", url, bytes.NewBuffer(bodyBytes))
		httpRequest.Header.Set("Content-Type", restful.MIME_JSON)
		httpWriter := httptest.NewRecorder()
		container.ServeHTTP(httpWriter, httpRequest)
		assert.Equal(t, httpWriter.Code, test.statusCode)
	}
}

func TestProjectHandler_AddReleasesInProject(t *testing.T) {
	var mockUseCase *mocks.UseCase
	var mockProjectHandler ProjectHandler

	container := restful.NewContainer()
	container.Add(RegisterProjectHandler(&mockProjectHandler))

	refreshMockUseCase := func() {
		mockUseCase = &mocks.UseCase{}
		mockProjectHandler.usecase = mockUseCase
	}
	tests := []struct {
		initMock   func()
		body       interface{}
		queryUrl   string
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
				mockUseCase.On("AddReleasesInProject", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New(""))
			},
			body:       &project.ProjectParams{},
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("AddReleasesInProject", "testns", "testnm", mock.Anything, false, int64(0)).Return(nil)
			},
			body:       &project.ProjectParams{},
			statusCode: 200,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("AddReleasesInProject", "testns", "testnm", mock.Anything, true, int64(60)).Return(nil)
			},
			body:       &project.ProjectParams{},
			queryUrl:   "?async=true&timeoutSec=60",
			statusCode: 200,
		},
		{
			initMock: func() {
				refreshMockUseCase()
			},
			body:       &project.ProjectParams{},
			queryUrl:   "?async=true&timeoutSec=notvalid",
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
			},
			body:       &project.ProjectParams{},
			queryUrl:   "?async=notvalid&timeoutSec=60",
			statusCode: 500,
		},
	}

	for _, test := range tests {
		test.initMock()
		url := projectRootPath + "/testns/name/testnm/project" + test.queryUrl

		bodyBytes, err := json.Marshal(test.body)
		assert.IsType(t, nil, err)

		httpRequest, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
		httpRequest.Header.Set("Content-Type", restful.MIME_JSON)
		httpWriter := httptest.NewRecorder()
		container.ServeHTTP(httpWriter, httpRequest)
		assert.Equal(t, httpWriter.Code, test.statusCode)
	}
}

func TestProjectHandler_DeleteReleaseInProject(t *testing.T) {
	var mockUseCase *mocks.UseCase
	var mockProjectHandler ProjectHandler

	container := restful.NewContainer()
	container.Add(RegisterProjectHandler(&mockProjectHandler))

	refreshMockUseCase := func() {
		mockUseCase = &mocks.UseCase{}
		mockProjectHandler.usecase = mockUseCase
	}
	tests := []struct {
		initMock   func()
		queryUrl   string
		statusCode int
	}{
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("RemoveReleaseInProject", "testns", "testname", "testrls", false, int64(0), false).Return(nil)
			},
			statusCode: 200,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("RemoveReleaseInProject", "testns", "testname", "testrls", false, int64(0), false).Return(errors.New(""))
			},
			statusCode: 500,
		},
		{
			initMock: func() {
				refreshMockUseCase()
				mockUseCase.On("RemoveReleaseInProject", "testns", "testname", "testrls", true, int64(60), true).Return(nil)
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
		url := projectRootPath + "/testns/name/testname/instance/testrls" + test.queryUrl

		httpRequest, _ := http.NewRequest("DELETE", url, nil)
		httpWriter := httptest.NewRecorder()
		container.ServeHTTP(httpWriter, httpRequest)
		assert.Equal(t, httpWriter.Code, test.statusCode)
	}
}
