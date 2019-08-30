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
	errorModel "WarpCloud/walm/pkg/models/error"
	"WarpCloud/walm/pkg/models/release"
	"WarpCloud/walm/pkg/models/task"
)

func TestProject_doAddRelease(t *testing.T) {

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
				mockProjectCache.On("GetProjectTask", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
			},
			err: errors.New("failed"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockProjectCache.On("GetProjectTask", mock.Anything, mock.Anything).Return(nil, errorModel.NotFoundError{})
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
				mockProjectCache.On("GetProjectTask", mock.Anything, mock.Anything).Return(nil, errorModel.NotFoundError{})
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
				mockProjectCache.On("GetProjectTask", mock.Anything, mock.Anything).Return(nil, errorModel.NotFoundError{})
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
		{
			initMock: func() {
				refreshMocks()
				mockProjectCache.On("GetProjectTask", mock.Anything, mock.Anything).Return(&project.ProjectTask{
					Namespace: "test-ns",
					Name:      "test-name",
					LatestTaskSignature: &task.TaskSig{
						Name: "test-name",
						UUID: "test-uuid",
					}}, nil)
				mockReleaseUseCase.On("ListReleasesByLabels", "test-ns", project.ProjectNameLabelKey+"=test-name").Return([]*release.ReleaseInfoV2{
					{
						ReleaseInfo: release.ReleaseInfo{
							ReleaseSpec: release.ReleaseSpec{
								Name: "A",
								ChartName: "chartA",
							},
						},
					},
				}, nil)
				mockTask.On("GetTaskState", &task.TaskSig{
					Name: "test-name",
					UUID: "test-uuid",
				}).Return(mockTaskState, nil)
				mockTaskState.On("IsFinished").Return(true)
				mockTaskState.On("IsSuccess").Return(true)

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
		err := mockProjectManager.doAddRelease("test-ns", "test-name", test.projectParams)
		assert.IsType(t, test.err, err)

		mockProjectCache.AssertExpectations(t)
		mockHelm.AssertExpectations(t)
		mockTask.AssertExpectations(t)
		mockReleaseUseCase.AssertExpectations(t)

		mockTaskState.AssertExpectations(t)
	}
}
