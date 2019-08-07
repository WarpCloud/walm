package helm

import (
	"testing"
	"WarpCloud/walm/pkg/models/release"
	"errors"
	"github.com/stretchr/testify/assert"
	helmMocks "WarpCloud/walm/pkg/helm/mocks"
	k8sMocks "WarpCloud/walm/pkg/k8s/mocks"
	taskMocks "WarpCloud/walm/pkg/task/mocks"
	"github.com/stretchr/testify/mock"
	"WarpCloud/walm/pkg/release/mocks"
	"WarpCloud/walm/pkg/models/task"
)

func TestHelm_sendReleaseTask(t *testing.T) {
	var mockReleaseCache *mocks.Cache
	var mockHelm *helmMocks.Helm
	var mockK8sOperator *k8sMocks.Operator
	var mockK8sCache *k8sMocks.Cache
	var mockTask *taskMocks.Task
	var mockReleaseManager *Helm

	var mockTaskState *taskMocks.TaskState

	refreshMocks := func() {
		mockReleaseCache = &mocks.Cache{}
		mockHelm = &helmMocks.Helm{}
		mockK8sOperator = &k8sMocks.Operator{}
		mockK8sCache = &k8sMocks.Cache{}
		mockTask = &taskMocks.Task{}

		mockTaskState = &taskMocks.TaskState{}

		mockTask.On("RegisterTask", mock.Anything, mock.Anything).Return(nil)

		var err error
		mockReleaseManager, err = NewHelm(mockReleaseCache, mockHelm, mockK8sCache, mockK8sOperator, mockTask)
		assert.IsType(t, err, nil)
	}

	tests := []struct {
		initMock       func()
		oldTask        *release.ReleaseTask
		async          bool
		err            error
	}{
		{
			initMock: func() {
				refreshMocks()
				mockTask.On("SendTask", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New(""))
			},
			err:            errors.New("failed"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockTask.On("SendTask", mock.Anything, mock.Anything, mock.Anything).Return(&task.TaskSig{}, nil)
				mockReleaseCache.On("CreateOrUpdateReleaseTask", mock.Anything).Return(errors.New(""))
			},
			err:            errors.New("failed"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockTask.On("SendTask", mock.Anything, mock.Anything, mock.Anything).Return(&task.TaskSig{}, nil)
				mockReleaseCache.On("CreateOrUpdateReleaseTask", mock.Anything).Return(nil)
			},
			async: true,
			err: nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockTask.On("SendTask", mock.Anything, mock.Anything, mock.Anything).Return(&task.TaskSig{}, nil)
				mockReleaseCache.On("CreateOrUpdateReleaseTask", mock.Anything).Return(nil)
				mockTask.On("PurgeTaskState", mock.Anything).Return(nil)
			},
			oldTask: &release.ReleaseTask{LatestReleaseTaskSig: &task.TaskSig{}},
			async: true,
			err: nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockTask.On("SendTask", mock.Anything, mock.Anything, mock.Anything).Return(&task.TaskSig{}, nil)
				mockReleaseCache.On("CreateOrUpdateReleaseTask", mock.Anything).Return(nil)
				mockTask.On("PurgeTaskState", mock.Anything).Return(errors.New(""))
			},
			oldTask: &release.ReleaseTask{LatestReleaseTaskSig: &task.TaskSig{}},
			async: true,
			err: nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockTask.On("SendTask", mock.Anything, mock.Anything, mock.Anything).Return(&task.TaskSig{}, nil)
				mockReleaseCache.On("CreateOrUpdateReleaseTask", mock.Anything).Return(nil)
				mockTask.On("TouchTask", mock.Anything, mock.Anything).Return(errors.New(""))
			},
			err: errors.New(""),
		},
		{
			initMock: func() {
				refreshMocks()
				mockTask.On("SendTask", mock.Anything, mock.Anything, mock.Anything).Return(&task.TaskSig{}, nil)
				mockReleaseCache.On("CreateOrUpdateReleaseTask", mock.Anything).Return(nil)
				mockTask.On("TouchTask", mock.Anything, mock.Anything).Return(nil)
			},
			err: nil,
		},
	}

	for _, test := range tests {
		test.initMock()
		err := mockReleaseManager.sendReleaseTask("test-ns", "test", "test", nil, test.oldTask, 0, test.async)
		assert.IsType(t, test.err, err)

		mockReleaseCache.AssertExpectations(t)
		mockHelm.AssertExpectations(t)
		mockK8sOperator.AssertExpectations(t)
		mockK8sCache.AssertExpectations(t)
		mockTask.AssertExpectations(t)

		mockTaskState.AssertExpectations(t)
	}

}
