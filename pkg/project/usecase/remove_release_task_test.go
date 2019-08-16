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
	"WarpCloud/walm/pkg/models/task"
	"WarpCloud/walm/pkg/models/release"
)

func TestProject_doRemoveRelease(t *testing.T) {

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
								Dependencies: map[string]string{"chartB": "B"},
							},
						},
					},
					{
						ReleaseInfo: release.ReleaseInfo{
							ReleaseSpec: release.ReleaseSpec{
								Name: "B",
								ChartName: "chartB",
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
				mockReleaseUseCase.On("InstallUpgradeReleaseWithRetry", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockReleaseUseCase.On("DeleteReleaseWithRetry", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			err: nil,
		},
	}

	for _, test := range tests {
		test.initMock()
		err := mockProjectManager.doRemoveRelease("test-ns", "test-name", "B", false)
		assert.IsType(t, test.err, err)

		mockProjectCache.AssertExpectations(t)
		mockHelm.AssertExpectations(t)
		mockTask.AssertExpectations(t)
		mockReleaseUseCase.AssertExpectations(t)

		mockTaskState.AssertExpectations(t)
	}
}
