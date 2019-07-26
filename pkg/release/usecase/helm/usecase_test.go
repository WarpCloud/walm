package helm

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"WarpCloud/walm/pkg/release/mocks"
	helmMocks "WarpCloud/walm/pkg/helm/mocks"
	k8sMocks "WarpCloud/walm/pkg/k8s/mocks"
	taskMocks "WarpCloud/walm/pkg/task/mocks"
	"github.com/stretchr/testify/mock"
	"errors"
	errorModel "WarpCloud/walm/pkg/models/error"
)

func TestHelm_ReloadRelease(t *testing.T) {

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
		initMock    func()
		err         error
	}{
		{
			initMock: func() {
				refreshMocks()
				mockReleaseCache.On("GetReleaseTask", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
			},
			err:         errors.New("failed"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockReleaseCache.On("GetReleaseTask", mock.Anything, mock.Anything).Return(nil, errorModel.NotFoundError{})
			},
			err:         nil,
		},

		//{
		//	initMock: func() {
		//		refreshMocks()
		//		mockReleaseCache.On("GetReleaseTask", mock.Anything, mock.Anything).Return(&release.ReleaseTask{
		//			Namespace: "test-ns",
		//			Name:      "test-name",
		//			LatestReleaseTaskSig: &task.TaskSig{
		//				Name: "test-name",
		//				UUID: "test-uuid",
		//			},
		//		}, nil)
		//		mockReleaseCache.On("GetReleaseCache", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
		//	},
		//	releaseInfo: &release.ReleaseInfoV2{
		//		ReleaseInfo: release.ReleaseInfo{
		//			ReleaseSpec: release.ReleaseSpec{
		//				Namespace: "test-ns",
		//				Name:      "test-name",
		//			},
		//		},
		//	},
		//	err: errors.New("failed"),
		//},
		//{
		//	initMock: func() {
		//		refreshMocks()
		//		mockReleaseCache.On("GetReleaseTask", mock.Anything, mock.Anything).Return(&release.ReleaseTask{
		//			Namespace: "test-ns",
		//			Name:      "test-name",
		//			LatestReleaseTaskSig: &task.TaskSig{
		//				Name: "test-name",
		//				UUID: "test-uuid",
		//			},
		//		}, nil)
		//		mockReleaseCache.On("GetReleaseCache", mock.Anything, mock.Anything).Return(nil, errorModel.NotFoundError{})
		//		mockTask.On("GetTaskState", &task.TaskSig{
		//			Name: "test-name",
		//			UUID: "test-uuid",
		//		}).Return(nil, errors.New("failed"))
		//	},
		//	releaseInfo: &release.ReleaseInfoV2{
		//		ReleaseInfo: release.ReleaseInfo{
		//			ReleaseSpec: release.ReleaseSpec{
		//				Namespace: "test-ns",
		//				Name:      "test-name",
		//			},
		//		},
		//	},
		//	err: errors.New("failed"),
		//},
		//{
		//	initMock: func() {
		//		refreshMocks()
		//		mockReleaseCache.On("GetReleaseTask", mock.Anything, mock.Anything).Return(&release.ReleaseTask{
		//			Namespace: "test-ns",
		//			Name:      "test-name",
		//			LatestReleaseTaskSig: &task.TaskSig{
		//				Name: "test-name",
		//				UUID: "test-uuid",
		//			},
		//		}, nil)
		//		mockReleaseCache.On("GetReleaseCache", mock.Anything, mock.Anything).Return(nil, errorModel.NotFoundError{})
		//		mockTask.On("GetTaskState", &task.TaskSig{
		//			Name: "test-name",
		//			UUID: "test-uuid",
		//		}).Return(nil, errorModel.NotFoundError{})
		//	},
		//	releaseInfo: &release.ReleaseInfoV2{
		//		ReleaseInfo: release.ReleaseInfo{
		//			ReleaseSpec: release.ReleaseSpec{
		//				Namespace: "test-ns",
		//				Name:      "test-name",
		//			},
		//		},
		//	},
		//	err: nil,
		//},
		//{
		//	initMock: func() {
		//		refreshMocks()
		//		mockReleaseCache.On("GetReleaseTask", mock.Anything, mock.Anything).Return(&release.ReleaseTask{
		//			Namespace: "test-ns",
		//			Name:      "test-name",
		//			LatestReleaseTaskSig: &task.TaskSig{
		//				Name: "test-name",
		//				UUID: "test-uuid",
		//			},
		//		}, nil)
		//		mockReleaseCache.On("GetReleaseCache", mock.Anything, mock.Anything).Return(nil, errorModel.NotFoundError{})
		//		mockTask.On("GetTaskState", &task.TaskSig{
		//			Name: "test-name",
		//			UUID: "test-uuid",
		//		}).Return(mockTaskState, nil)
		//		mockTaskState.On("IsFinished").Return(true)
		//		mockTaskState.On("IsSuccess").Return(true)
		//	},
		//	releaseInfo: &release.ReleaseInfoV2{
		//		ReleaseInfo: release.ReleaseInfo{
		//			ReleaseSpec: release.ReleaseSpec{
		//				Namespace: "test-ns",
		//				Name:      "test-name",
		//			},
		//		},
		//	},
		//	err: nil,
		//},
		//{
		//	initMock: func() {
		//		refreshMocks()
		//		mockReleaseCache.On("GetReleaseTask", mock.Anything, mock.Anything).Return(&release.ReleaseTask{
		//			Namespace: "test-ns",
		//			Name:      "test-name",
		//			LatestReleaseTaskSig: &task.TaskSig{
		//				Name: "test-name",
		//				UUID: "test-uuid",
		//			},
		//		}, nil)
		//		mockReleaseCache.On("GetReleaseCache", mock.Anything, mock.Anything).Return(nil, errorModel.NotFoundError{})
		//		mockTask.On("GetTaskState", &task.TaskSig{
		//			Name: "test-name",
		//			UUID: "test-uuid",
		//		}).Return(mockTaskState, nil)
		//		mockTaskState.On("IsFinished").Return(true)
		//		mockTaskState.On("IsSuccess").Return(false)
		//		mockTaskState.On("GetErrorMsg").Return("test-err")
		//	},
		//	releaseInfo: &release.ReleaseInfoV2{
		//		ReleaseInfo: release.ReleaseInfo{
		//			ReleaseSpec: release.ReleaseSpec{
		//				Namespace: "test-ns",
		//				Name:      "test-name",
		//			},
		//			Message: "the release latest task test-name-test-uuid failed : test-err",
		//		},
		//	},
		//	err: nil,
		//},
		//{
		//	initMock: func() {
		//		refreshMocks()
		//		mockReleaseCache.On("GetReleaseTask", mock.Anything, mock.Anything).Return(&release.ReleaseTask{
		//			Namespace: "test-ns",
		//			Name:      "test-name",
		//			LatestReleaseTaskSig: &task.TaskSig{
		//				Name: "test-name",
		//				UUID: "test-uuid",
		//			},
		//		}, nil)
		//		mockReleaseCache.On("GetReleaseCache", mock.Anything, mock.Anything).Return(nil, errorModel.NotFoundError{})
		//		mockTask.On("GetTaskState", &task.TaskSig{
		//			Name: "test-name",
		//			UUID: "test-uuid",
		//		}).Return(mockTaskState, nil)
		//		mockTaskState.On("IsFinished").Return(false)
		//	},
		//	releaseInfo: &release.ReleaseInfoV2{
		//		ReleaseInfo: release.ReleaseInfo{
		//			ReleaseSpec: release.ReleaseSpec{
		//				Namespace: "test-ns",
		//				Name:      "test-name",
		//			},
		//			Message: "please wait for the release latest task test-name-test-uuid finished",
		//		},
		//	},
		//	err: nil,
		//},
		//{
		//	initMock: func() {
		//		refreshMocks()
		//		mockReleaseCache.On("GetReleaseTask", mock.Anything, mock.Anything).Return(&release.ReleaseTask{
		//			Namespace: "test-ns",
		//			Name:      "test-name",
		//			LatestReleaseTaskSig: &task.TaskSig{
		//				Name: "test-name",
		//				UUID: "test-uuid",
		//			},
		//		}, nil)
		//		mockReleaseCache.On("GetReleaseCache", mock.Anything, mock.Anything).Return(&release.ReleaseCache{
		//			ReleaseSpec: release.ReleaseSpec{
		//				Namespace: "test-ns",
		//				Name:      "test-name",
		//			},
		//		}, nil)
		//		mockTask.On("GetTaskState", &task.TaskSig{
		//			Name: "test-name",
		//			UUID: "test-uuid",
		//		}).Return(mockTaskState, nil)
		//		mockTaskState.On("IsFinished").Return(true)
		//		mockTaskState.On("IsSuccess").Return(true)
		//		mockK8sCache.On("GetResourceSet", ([]release.ReleaseResourceMeta)(nil)).Return(k8s.NewResourceSet(), nil)
		//		mockK8sCache.On("GetResource", k8s.ReleaseConfigKind, "test-ns", "test-name").Return(&k8s.ReleaseConfig{}, nil)
		//	},
		//	releaseInfo: &release.ReleaseInfoV2{
		//		ReleaseInfo: release.ReleaseInfo{
		//			ReleaseSpec: release.ReleaseSpec{
		//				Namespace: "test-ns",
		//				Name:      "test-name",
		//			},
		//			Ready:  true,
		//			Status: k8s.NewResourceSet(),
		//		},
		//		Plugins: []*release.ReleasePlugin{},
		//	},
		//	err: nil,
		//},
	}

	for _, test := range tests {
		test.initMock()
		err := mockReleaseManager.ReloadRelease("test-ns", "test-name")
		assert.IsType(t, test.err, err)

		mockReleaseCache.AssertExpectations(t)
		mockHelm.AssertExpectations(t)
		mockK8sOperator.AssertExpectations(t)
		mockK8sCache.AssertExpectations(t)
		mockTask.AssertExpectations(t)

		mockTaskState.AssertExpectations(t)
	}
}

