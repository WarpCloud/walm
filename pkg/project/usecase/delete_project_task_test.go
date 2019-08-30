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

func TestProject_doDeleteProject(t *testing.T) {

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
			},
			err: nil,
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

				mockReleaseUseCase.On("DeleteReleaseWithRetry", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New(""))
			},
			err: errors.New(""),
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

				mockReleaseUseCase.On("DeleteReleaseWithRetry", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockProjectCache.On("DeleteProjectTask", mock.Anything, mock.Anything).Return(nil)
			},
			err: nil,
		},
	}

	for _, test := range tests {
		test.initMock()
		err := mockProjectManager.doDeleteProject("test-ns", "test-name", false)
		assert.IsType(t, test.err, err)

		mockProjectCache.AssertExpectations(t)
		mockHelm.AssertExpectations(t)
		mockTask.AssertExpectations(t)
		mockReleaseUseCase.AssertExpectations(t)

		mockTaskState.AssertExpectations(t)
	}
}
