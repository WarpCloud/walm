package helm

import (
	"testing"
	"WarpCloud/walm/pkg/models/release"
	"WarpCloud/walm/pkg/models/common"
	"errors"
	"github.com/stretchr/testify/assert"
	helmMocks "WarpCloud/walm/pkg/helm/mocks"
	k8sMocks "WarpCloud/walm/pkg/k8s/mocks"
	taskMocks "WarpCloud/walm/pkg/task/mocks"
	"github.com/stretchr/testify/mock"
	"WarpCloud/walm/pkg/release/mocks"
	errorModel "WarpCloud/walm/pkg/models/error"
	"WarpCloud/walm/pkg/models/task"
)

func TestHelm_InstallUpgradeReleaseWithRetry(t *testing.T) {
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
		initMock       func()
		releaseRequest *release.ReleaseRequestV2
		err            error
	}{
		{
			initMock: func() {
				refreshMocks()
				mockReleaseCache.On("GetReleaseTask", mock.Anything, mock.Anything).Return(nil, errorModel.NotFoundError{})
				mockTask.On("SendTask", mock.Anything, mock.Anything, mock.Anything).Return(&task.TaskSig{}, nil)
				mockReleaseCache.On("CreateOrUpdateReleaseTask", mock.Anything).Return(nil)
				mockTask.On("TouchTask", mock.Anything, mock.Anything).Return(nil)
			},
			releaseRequest: &release.ReleaseRequestV2{
				ReleaseRequest: release.ReleaseRequest{
					Name:      "test",
					ChartName: "test",
				},
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
			releaseRequest: &release.ReleaseRequestV2{
				ReleaseRequest: release.ReleaseRequest{
					Name:      "test",
					ChartName: "test",
				},
			},
			err: nil,
		},
	}

	for _, test := range tests {
		test.initMock()
		err := mockReleaseManager.InstallUpgradeReleaseWithRetry("test-ns", test.releaseRequest, nil, false, 0, nil)
		assert.IsType(t, test.err, err)

		mockReleaseCache.AssertExpectations(t)
		mockHelm.AssertExpectations(t)
		mockK8sOperator.AssertExpectations(t)
		mockK8sCache.AssertExpectations(t)
		mockTask.AssertExpectations(t)

		mockTaskState.AssertExpectations(t)
	}

}

func TestHelm_InstallUpgradeRelease(t *testing.T) {
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
		releaseRequest *release.ReleaseRequestV2
		err            error
	}{
		{
			initMock: func() {
				refreshMocks()
			},
			releaseRequest: &release.ReleaseRequestV2{},
			err:            errors.New("failed"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockReleaseCache.On("GetReleaseTask", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
			},
			releaseRequest: &release.ReleaseRequestV2{
				ReleaseRequest: release.ReleaseRequest{
					Name:      "test",
					ChartName: "test",
				},
			},
			err: errors.New("failed"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockReleaseCache.On("GetReleaseTask", mock.Anything, mock.Anything).Return(nil, errorModel.NotFoundError{})
				mockTask.On("SendTask", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New(""))
			},
			releaseRequest: &release.ReleaseRequestV2{
				ReleaseRequest: release.ReleaseRequest{
					Name:      "test",
					ChartName: "test",
				},
			},
			err: errors.New("failed"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockReleaseCache.On("GetReleaseTask", mock.Anything, mock.Anything).Return(nil, errorModel.NotFoundError{})
				mockTask.On("SendTask", mock.Anything, mock.Anything, mock.Anything).Return(&task.TaskSig{}, nil)
				mockReleaseCache.On("CreateOrUpdateReleaseTask", mock.Anything).Return(nil)
				mockTask.On("TouchTask", mock.Anything, mock.Anything).Return(nil)
			},
			releaseRequest: &release.ReleaseRequestV2{
				ReleaseRequest: release.ReleaseRequest{
					Name:      "test",
					ChartName: "test",
				},
			},
			err: nil,
		},
	}

	for _, test := range tests {
		test.initMock()
		err := mockReleaseManager.InstallUpgradeRelease("test-ns", test.releaseRequest, nil, false, 0, nil)
		assert.IsType(t, test.err, err)

		mockReleaseCache.AssertExpectations(t)
		mockHelm.AssertExpectations(t)
		mockK8sOperator.AssertExpectations(t)
		mockK8sCache.AssertExpectations(t)
		mockTask.AssertExpectations(t)

		mockTaskState.AssertExpectations(t)
	}

}

func TestHelm_doInstallUpgradeRelease(t *testing.T) {
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
		releaseRequest *release.ReleaseRequestV2
		dryRun         bool
		err            error
	}{
		{
			initMock: func() {
				refreshMocks()
				mockReleaseCache.On("GetReleaseCache", mock.Anything, mock.Anything).Return(nil, errors.New(""))
			},
			releaseRequest: &release.ReleaseRequestV2{},
			err:            errors.New("failed"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockReleaseCache.On("GetReleaseCache", mock.Anything, mock.Anything).Return(nil, errorModel.NotFoundError{})
				mockHelm.On("InstallOrCreateRelease", mock.Anything, mock.Anything, mock.Anything, mock.Anything, false, mock.Anything, mock.Anything).Return(nil, errors.New(""))
			},
			releaseRequest: &release.ReleaseRequestV2{},
			err:            errors.New("failed"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockReleaseCache.On("GetReleaseCache", mock.Anything, mock.Anything).Return(nil, errorModel.NotFoundError{})
				mockHelm.On("InstallOrCreateRelease", mock.Anything, mock.Anything, mock.Anything, mock.Anything, false, mock.Anything, mock.Anything).Return(&release.ReleaseCache{}, nil)
			},
			dryRun: true,
			releaseRequest: &release.ReleaseRequestV2{},
			err:            nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockReleaseCache.On("GetReleaseCache", mock.Anything, mock.Anything).Return(nil, errorModel.NotFoundError{})
				mockHelm.On("InstallOrCreateRelease", mock.Anything, mock.Anything, mock.Anything, mock.Anything, false, mock.Anything, mock.Anything).Return(&release.ReleaseCache{}, nil)
				mockReleaseCache.On("CreateOrUpdateReleaseCache", mock.Anything).Return(errors.New(""))
			},
			dryRun: false,
			releaseRequest: &release.ReleaseRequestV2{},
			err:            errors.New(""),
		},
		{
			initMock: func() {
				refreshMocks()
				mockReleaseCache.On("GetReleaseCache", mock.Anything, mock.Anything).Return(nil, errorModel.NotFoundError{})
				mockHelm.On("InstallOrCreateRelease", mock.Anything, mock.Anything, mock.Anything, mock.Anything, false, mock.Anything, mock.Anything).Return(&release.ReleaseCache{}, nil)
				mockReleaseCache.On("CreateOrUpdateReleaseCache", mock.Anything).Return(nil)
			},
			dryRun: false,
			releaseRequest: &release.ReleaseRequestV2{},
			err:            nil,
		},
	}

	for _, test := range tests {
		test.initMock()
		_, err := mockReleaseManager.doInstallUpgradeRelease("test-ns", test.releaseRequest, nil, test.dryRun, nil)
		assert.IsType(t, test.err, err)

		mockReleaseCache.AssertExpectations(t)
		mockHelm.AssertExpectations(t)
		mockK8sOperator.AssertExpectations(t)
		mockK8sCache.AssertExpectations(t)
		mockTask.AssertExpectations(t)

		mockTaskState.AssertExpectations(t)
	}

}

func Test_validateParams(t *testing.T) {
	tests := []struct {
		releaseRequest *release.ReleaseRequestV2
		chartFiles     []*common.BufferedFile
		err            error
	}{
		{
			releaseRequest: &release.ReleaseRequestV2{
			},
			err: errors.New(""),
		},
		{
			releaseRequest: &release.ReleaseRequestV2{
				ReleaseRequest: release.ReleaseRequest{
					Name: "test",
				},
			},
			err: errors.New(""),
		},
		{
			releaseRequest: &release.ReleaseRequestV2{
				ReleaseRequest: release.ReleaseRequest{
					Name:      "test",
					ChartName: "test",
				},
			},
			err: nil,
		},
		{
			releaseRequest: &release.ReleaseRequestV2{
				ReleaseRequest: release.ReleaseRequest{
					Name: "test",
				},
				ChartImage: "test",
			},
			err: nil,
		},
		{
			releaseRequest: &release.ReleaseRequestV2{
				ReleaseRequest: release.ReleaseRequest{
					Name: "test",
				},
			},
			chartFiles: []*common.BufferedFile{
				{
				},
			},
			err: nil,
		},
	}

	for _, test := range tests {
		err := validateParams(test.releaseRequest, test.chartFiles)
		assert.IsType(t, test.err, err)
	}
}
