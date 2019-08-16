package usecase

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"WarpCloud/walm/pkg/project/mocks"
	helmMocks "WarpCloud/walm/pkg/helm/mocks"
	taskMocks "WarpCloud/walm/pkg/task/mocks"
	"github.com/stretchr/testify/mock"
	"errors"
	releaseMocks "WarpCloud/walm/pkg/release/mocks"
	"WarpCloud/walm/pkg/models/project"
	"WarpCloud/walm/pkg/models/release"
)

func TestProject_doCreateProject(t *testing.T) {

	var mockProjectCache *mocks.Cache
	var mockHelm *helmMocks.Helm
	var mockTask *taskMocks.Task
	var mockReleaseUseCase *releaseMocks.UseCase

	var mockProjectManager *Project

	var mockTaskState *taskMocks.TaskState

	refreshMocks := func() {
		mockProjectCache = &mocks.Cache{}
		mockHelm = &helmMocks.Helm{}
		mockTask = &taskMocks.Task{}
		mockReleaseUseCase = &releaseMocks.UseCase{}

		mockTaskState = &taskMocks.TaskState{}

		mockTask.On("RegisterTask", mock.Anything, mock.Anything).Return(nil)

		var err error
		mockProjectManager, err = NewProject(mockProjectCache, mockTask, mockReleaseUseCase, mockHelm)
		assert.IsType(t, err, nil)
	}

	tests := []struct {
		initMock      func()
		projectParams *project.ProjectParams
		err           error
	}{
		{
			initMock: func() {
				refreshMocks()
				mockHelm.On("GetChartAutoDependencies", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New(""))
			},
			projectParams: &project.ProjectParams{
				Releases: []*release.ReleaseRequestV2{
					{
						ReleaseRequest: release.ReleaseRequest{
							Name:      "A",
							ChartName: "chartA",
						},
					},
					{
						ReleaseRequest: release.ReleaseRequest{
							Name:      "B",
							ChartName: "chartB",
						},
					},
				},
			},
			err: errors.New("failed"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockHelm.On("GetChartAutoDependencies", mock.Anything, mock.Anything, mock.Anything).Return(func(repo, chart, version string) (result []string) {
					if chart == "chartA" {
						result = append(result, "chartB")
					}
					return
				}, nil)
				mockReleaseUseCase.On("InstallUpgradeReleaseWithRetry", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New(""))
			},
			projectParams: &project.ProjectParams{
				Releases: []*release.ReleaseRequestV2{
					{
						ReleaseRequest: release.ReleaseRequest{
							Name:      "A",
							ChartName: "chartA",
						},
					},
					{
						ReleaseRequest: release.ReleaseRequest{
							Name:      "B",
							ChartName: "chartB",
						},
					},
				},
			},
			err: errors.New("failed"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockHelm.On("GetChartAutoDependencies", mock.Anything, mock.Anything, mock.Anything).Return(func(repo, chart, version string) (result []string) {
					if chart == "chartA" {
						result = append(result, "chartB")
					}
					return
				}, nil)
				mockReleaseUseCase.On("InstallUpgradeReleaseWithRetry", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			projectParams: &project.ProjectParams{
				Releases: []*release.ReleaseRequestV2{
					{
						ReleaseRequest: release.ReleaseRequest{
							Name:      "A",
							ChartName: "chartA",
						},
					},
					{
						ReleaseRequest: release.ReleaseRequest{
							Name:      "B",
							ChartName: "chartB",
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		test.initMock()
		err := mockProjectManager.doCreateProject("test-ns", "test-name", test.projectParams)
		assert.IsType(t, test.err, err)

		mockProjectCache.AssertExpectations(t)
		mockHelm.AssertExpectations(t)
		mockTask.AssertExpectations(t)
		mockReleaseUseCase.AssertExpectations(t)

		mockTaskState.AssertExpectations(t)
	}
}
