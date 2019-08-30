package usecase

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"WarpCloud/walm/pkg/project/mocks"
	helmMocks "WarpCloud/walm/pkg/helm/mocks"
	taskMocks "WarpCloud/walm/pkg/task/mocks"
	"github.com/stretchr/testify/mock"
	releaseMocks "WarpCloud/walm/pkg/release/mocks"
	"WarpCloud/walm/pkg/models/release"
	"github.com/pkg/errors"
)

func TestProject_doUpgradeRelease(t *testing.T) {

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
				mockReleaseUseCase.On("InstallUpgradeReleaseWithRetry", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			err: nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockReleaseUseCase.On("InstallUpgradeReleaseWithRetry", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New(""))
			},
			err: errors.New(""),
		},
	}

	for _, test := range tests {
		test.initMock()
		err := mockProjectManager.upgradeRelease("test-ns", "test-name", &release.ReleaseRequestV2{})
		assert.IsType(t, test.err, err)

		mockProjectCache.AssertExpectations(t)
		mockHelm.AssertExpectations(t)
		mockTask.AssertExpectations(t)
		mockReleaseUseCase.AssertExpectations(t)

		mockTaskState.AssertExpectations(t)
	}
}
