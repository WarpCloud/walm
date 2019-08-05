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
	errorModel "WarpCloud/walm/pkg/models/error"
	"WarpCloud/walm/pkg/models/task"
	"WarpCloud/walm/pkg/models/k8s"
)

func TestHelm_DeleteReleaseWithRetry(t *testing.T) {
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

	var retryTimes int

	tests := []struct {
		initMock func()
		err      error
	}{
		{
			initMock: func() {
				refreshMocks()
				mockReleaseCache.On("GetReleaseTask", mock.Anything, mock.Anything).Return(nil, errorModel.NotFoundError{})
			},
			err: nil,
		},
		{
			initMock: func() {
				refreshMocks()
				retryTimes = 0
				mockReleaseCache.On("GetReleaseTask", mock.Anything, mock.Anything).Return(&release.ReleaseTask{
					LatestReleaseTaskSig: &task.TaskSig{
						Name: "test-name",
						UUID: "test-uuid",
					},
				}, nil)
				mockTask.On("GetTaskState", mock.Anything).Return(mockTaskState, nil)

				mockTaskState.On("IsFinished").Return(func() bool {
					if retryTimes <= 0 {
						retryTimes ++
						return false
					} else {
						return true
					}
				}).Twice()
				mockTaskState.On("IsTimeout").Return(false)

				mockTask.On("SendTask", mock.Anything, mock.Anything, mock.Anything).Return(&task.TaskSig{}, nil)
				mockReleaseCache.On("CreateOrUpdateReleaseTask", mock.Anything).Return(nil)
				mockTask.On("TouchTask", mock.Anything, mock.Anything).Return(nil)
				mockTask.On("PurgeTaskState", mock.Anything).Return(errors.New(""))
			},
			err: nil,
		},
	}

	for _, test := range tests {
		test.initMock()
		err := mockReleaseManager.DeleteReleaseWithRetry("test-ns", "test-name", false, false, 0)
		assert.IsType(t, test.err, err)

		mockReleaseCache.AssertExpectations(t)
		mockHelm.AssertExpectations(t)
		mockK8sOperator.AssertExpectations(t)
		mockK8sCache.AssertExpectations(t)
		mockTask.AssertExpectations(t)

		mockTaskState.AssertExpectations(t)
	}

}

func TestHelm_DeleteRelease(t *testing.T) {
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
		initMock func()
		err      error
	}{
		{
			initMock: func() {
				refreshMocks()
				mockReleaseCache.On("GetReleaseTask", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
			},
			err: errors.New("failed"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockReleaseCache.On("GetReleaseTask", mock.Anything, mock.Anything).Return(nil, errorModel.NotFoundError{})
			},
			err: nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockReleaseCache.On("GetReleaseTask", mock.Anything, mock.Anything).Return(&release.ReleaseTask{}, nil)
				mockTask.On("GetTaskState", mock.Anything).Return(nil, errorModel.NotFoundError{})
				mockTask.On("SendTask", mock.Anything, mock.Anything, mock.Anything).Return(&task.TaskSig{}, nil)
				mockReleaseCache.On("CreateOrUpdateReleaseTask", mock.Anything).Return(nil)
				mockTask.On("TouchTask", mock.Anything, mock.Anything).Return(nil)
			},
			err: nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockReleaseCache.On("GetReleaseTask", mock.Anything, mock.Anything).Return(&release.ReleaseTask{}, nil)
				mockTask.On("GetTaskState", mock.Anything).Return(nil, errorModel.NotFoundError{})
				mockTask.On("SendTask", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New(""))
			},
			err: errors.New(""),
		},
	}

	for _, test := range tests {
		test.initMock()
		err := mockReleaseManager.DeleteRelease("test-ns", "test-name", false, false, 0)
		assert.IsType(t, test.err, err)

		mockReleaseCache.AssertExpectations(t)
		mockHelm.AssertExpectations(t)
		mockK8sOperator.AssertExpectations(t)
		mockK8sCache.AssertExpectations(t)
		mockTask.AssertExpectations(t)

		mockTaskState.AssertExpectations(t)
	}

}

func TestHelm_doDeleteRelease(t *testing.T) {
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
		initMock   func()
		deletePvcs bool
		err        error
	}{
		{
			initMock: func() {
				refreshMocks()
				mockReleaseCache.On("GetReleaseCache", mock.Anything, mock.Anything).Return(nil, errors.New(""))
			},
			err: errors.New("failed"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockReleaseCache.On("GetReleaseCache", mock.Anything, mock.Anything).Return(nil, errorModel.NotFoundError{})
			},
			err: nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockReleaseCache.On("GetReleaseCache", mock.Anything, mock.Anything).Return(nil, errorModel.NotFoundError{})
			},
			err: nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockReleaseCache.On("GetReleaseCache", mock.Anything, mock.Anything).Return(&release.ReleaseCache{}, nil)
				mockK8sCache.On("GetResourceSet", mock.Anything).Return(nil, errors.New(""))
			},
			err: errors.New(""),
		},
		{
			initMock: func() {
				refreshMocks()
				mockReleaseCache.On("GetReleaseCache", mock.Anything, mock.Anything).Return(&release.ReleaseCache{}, nil)
				mockK8sCache.On("GetResourceSet", mock.Anything).Return(k8s.NewResourceSet(), nil)
				mockK8sCache.On("GetResource", mock.Anything, mock.Anything, mock.Anything).Return(&k8s.ReleaseConfig{}, nil)
				mockHelm.On("DeleteRelease", mock.Anything, mock.Anything).Return(errors.New(""))
			},
			err: errors.New(""),
		},
		{
			initMock: func() {
				refreshMocks()
				mockReleaseCache.On("GetReleaseCache", mock.Anything, mock.Anything).Return(&release.ReleaseCache{}, nil)
				mockK8sCache.On("GetResourceSet", mock.Anything).Return(k8s.NewResourceSet(), nil)
				mockK8sCache.On("GetResource", mock.Anything, mock.Anything, mock.Anything).Return(&k8s.ReleaseConfig{}, nil)
				mockHelm.On("DeleteRelease", mock.Anything, mock.Anything).Return(nil)
				mockReleaseCache.On("DeleteReleaseCache", mock.Anything, mock.Anything).Return(errors.New(""))
			},
			err: errors.New(""),
		},
		{
			initMock: func() {
				refreshMocks()
				mockReleaseCache.On("GetReleaseCache", mock.Anything, mock.Anything).Return(&release.ReleaseCache{}, nil)
				mockK8sCache.On("GetResourceSet", mock.Anything).Return(k8s.NewResourceSet(), nil)
				mockK8sCache.On("GetResource", mock.Anything, mock.Anything, mock.Anything).Return(&k8s.ReleaseConfig{}, nil)
				mockHelm.On("DeleteRelease", mock.Anything, mock.Anything).Return(nil)
				mockReleaseCache.On("DeleteReleaseCache", mock.Anything, mock.Anything).Return(nil)
				mockK8sOperator.On("DeleteStatefulSetPvcs", mock.Anything).Return(errors.New(""))
			},
			deletePvcs: true,
			err: errors.New(""),
		},
		{
			initMock: func() {
				refreshMocks()
				mockReleaseCache.On("GetReleaseCache", mock.Anything, mock.Anything).Return(&release.ReleaseCache{}, nil)
				mockK8sCache.On("GetResourceSet", mock.Anything).Return(k8s.NewResourceSet(), nil)
				mockK8sCache.On("GetResource", mock.Anything, mock.Anything, mock.Anything).Return(&k8s.ReleaseConfig{}, nil)
				mockHelm.On("DeleteRelease", mock.Anything, mock.Anything).Return(nil)
				mockReleaseCache.On("DeleteReleaseCache", mock.Anything, mock.Anything).Return(nil)
				mockK8sOperator.On("DeleteStatefulSetPvcs", mock.Anything).Return(nil)
			},
			deletePvcs: true,
			err: nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockReleaseCache.On("GetReleaseCache", mock.Anything, mock.Anything).Return(&release.ReleaseCache{}, nil)
				mockK8sCache.On("GetResourceSet", mock.Anything).Return(k8s.NewResourceSet(), nil)
				mockK8sCache.On("GetResource", mock.Anything, mock.Anything, mock.Anything).Return(&k8s.ReleaseConfig{}, nil)
				mockHelm.On("DeleteRelease", mock.Anything, mock.Anything).Return(nil)
				mockReleaseCache.On("DeleteReleaseCache", mock.Anything, mock.Anything).Return(nil)
			},
			deletePvcs: false,
			err: nil,
		},
	}

	for _, test := range tests {
		test.initMock()
		err := mockReleaseManager.doDeleteRelease("test-ns", "test-name", test.deletePvcs)
		assert.IsType(t, test.err, err)

		mockReleaseCache.AssertExpectations(t)
		mockHelm.AssertExpectations(t)
		mockK8sOperator.AssertExpectations(t)
		mockK8sCache.AssertExpectations(t)
		mockTask.AssertExpectations(t)

		mockTaskState.AssertExpectations(t)
	}

}
