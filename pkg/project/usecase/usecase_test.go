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
	errorModel "WarpCloud/walm/pkg/models/error"
	"WarpCloud/walm/pkg/models/release"
)

func TestProject_ListProjects(t *testing.T) {

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
		initMock        func()
		projectInfoList *project.ProjectInfoList
		err             error
	}{
		{
			initMock: func() {
				refreshMocks()
				mockProjectCache.On("GetProjectTasks", mock.Anything).Return(nil, errors.New("failed"))
			},
			projectInfoList: nil,
			err:             errors.New("failed"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockProjectCache.On("GetProjectTasks", mock.Anything).Return([]*project.ProjectTask{{
					Namespace: "test-ns",
					Name:      "test-name",
					LatestTaskSignature: &task.TaskSig{
						Name: "test-name",
						UUID: "test-uuid",
					}}}, nil)
				mockReleaseUseCase.On("ListReleasesByLabels", "test-ns", project.ProjectNameLabelKey+"=test-name").Return(nil, nil)
				mockTask.On("GetTaskState", &task.TaskSig{
					Name: "test-name",
					UUID: "test-uuid",
				}).Return(mockTaskState, nil)
				mockTaskState.On("IsFinished").Return(true)
				mockTaskState.On("IsSuccess").Return(true)
			},
			projectInfoList: &project.ProjectInfoList{
				Items: []*project.ProjectInfo{
					{
						Namespace: "test-ns",
						Name:      "test-name",
						Message:   noReleaseFoundMsg,
					},
				},
				Num: 1,
			},
			err: nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockProjectCache.On("GetProjectTasks", mock.Anything).Return([]*project.ProjectTask{{
					Namespace: "test-ns",
					Name:      "test-name",
					LatestTaskSignature: &task.TaskSig{
						Name: "test-name",
						UUID: "test-uuid",
					}}}, nil)
				mockReleaseUseCase.On("ListReleasesByLabels", "test-ns", project.ProjectNameLabelKey+"=test-name").Return(nil, errors.New(""))
			},
			projectInfoList: nil,
			err:             errors.New(""),
		},
	}

	for _, test := range tests {
		test.initMock()
		projectInfoList, err := mockProjectManager.ListProjects("test-ns")
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.projectInfoList, projectInfoList)

		mockProjectCache.AssertExpectations(t)
		mockHelm.AssertExpectations(t)
		mockTask.AssertExpectations(t)
		mockReleaseUseCase.AssertExpectations(t)

		mockTaskState.AssertExpectations(t)
	}
}

func TestProject_GetProjectInfo(t *testing.T) {

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
		initMock    func()
		projectInfo *project.ProjectInfo
		err         error
	}{
		{
			initMock: func() {
				refreshMocks()
				mockProjectCache.On("GetProjectTask", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
			},
			projectInfo: nil,
			err:         errors.New("failed"),
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
				mockReleaseUseCase.On("ListReleasesByLabels", "test-ns", project.ProjectNameLabelKey+"=test-name").Return(nil, nil)
				mockTask.On("GetTaskState", &task.TaskSig{
					Name: "test-name",
					UUID: "test-uuid",
				}).Return(mockTaskState, nil)
				mockTaskState.On("IsFinished").Return(true)
				mockTaskState.On("IsSuccess").Return(true)
			},
			projectInfo: &project.ProjectInfo{
				Namespace: "test-ns",
				Name:      "test-name",
				Message:   noReleaseFoundMsg,
			},
			err: nil,
		},
	}

	for _, test := range tests {
		test.initMock()
		projectInfo, err := mockProjectManager.GetProjectInfo("test-ns", "test-name")
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.projectInfo, projectInfo)

		mockProjectCache.AssertExpectations(t)
		mockHelm.AssertExpectations(t)
		mockTask.AssertExpectations(t)
		mockReleaseUseCase.AssertExpectations(t)

		mockTaskState.AssertExpectations(t)
	}
}

func TestProject_buildProjectInfo(t *testing.T) {

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
		initMock    func()
		task        *project.ProjectTask
		projectInfo *project.ProjectInfo
		err         error
	}{
		{
			initMock: func() {
				refreshMocks()
				mockReleaseUseCase.On("ListReleasesByLabels", "test-ns", project.ProjectNameLabelKey+"=test-name").Return(nil, errors.New("failed"))
			},
			task: &project.ProjectTask{
				Namespace: "test-ns",
				Name:      "test-name",
				LatestTaskSignature: &task.TaskSig{
					Name: "test-name",
					UUID: "test-uuid",
				},
			},
			projectInfo: nil,
			err:         errors.New("failed"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockReleaseUseCase.On("ListReleasesByLabels", "test-ns", project.ProjectNameLabelKey+"=test-name").Return(nil, nil)
				mockTask.On("GetTaskState", &task.TaskSig{
					Name: "test-name",
					UUID: "test-uuid",
				}).Return(nil, errors.New("failed"))
			},
			task: &project.ProjectTask{
				Namespace: "test-ns",
				Name:      "test-name",
				LatestTaskSignature: &task.TaskSig{
					Name: "test-name",
					UUID: "test-uuid",
				},
			},
			projectInfo: nil,
			err:         errors.New("failed"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockReleaseUseCase.On("ListReleasesByLabels", "test-ns", project.ProjectNameLabelKey+"=test-name").Return(nil, nil)
				mockTask.On("GetTaskState", &task.TaskSig{
					Name: "test-name",
					UUID: "test-uuid",
				}).Return(nil, errorModel.NotFoundError{})
			},
			task: &project.ProjectTask{
				Namespace: "test-ns",
				Name:      "test-name",
				LatestTaskSignature: &task.TaskSig{
					Name: "test-name",
					UUID: "test-uuid",
				},
			},
			projectInfo: &project.ProjectInfo{
				Namespace: "test-ns",
				Name:      "test-name",
				Message:   noReleaseFoundMsg,
			},
			err: nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockReleaseUseCase.On("ListReleasesByLabels", "test-ns", project.ProjectNameLabelKey+"=test-name").Return(nil, nil)
				mockTask.On("GetTaskState", &task.TaskSig{
					Name: "test-name",
					UUID: "test-uuid",
				}).Return(mockTaskState, nil)
				mockTaskState.On("IsFinished").Return(true)
				mockTaskState.On("IsSuccess").Return(true)
			},
			task: &project.ProjectTask{
				Namespace: "test-ns",
				Name:      "test-name",
				LatestTaskSignature: &task.TaskSig{
					Name: "test-name",
					UUID: "test-uuid",
				},
			},
			projectInfo: &project.ProjectInfo{
				Namespace: "test-ns",
				Name:      "test-name",
				Message:   noReleaseFoundMsg,
			},
			err: nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockReleaseUseCase.On("ListReleasesByLabels", "test-ns", project.ProjectNameLabelKey+"=test-name").Return(nil, nil)
				mockTask.On("GetTaskState", &task.TaskSig{
					Name: "test-name",
					UUID: "test-uuid",
				}).Return(mockTaskState, nil)
				mockTaskState.On("IsFinished").Return(true)
				mockTaskState.On("IsSuccess").Return(false)
				mockTaskState.On("GetErrorMsg").Return("test-err")
			},
			task: &project.ProjectTask{
				Namespace: "test-ns",
				Name:      "test-name",
				LatestTaskSignature: &task.TaskSig{
					Name: "test-name",
					UUID: "test-uuid",
				},
			},
			projectInfo: &project.ProjectInfo{
				Namespace: "test-ns",
				Name:      "test-name",
				Message:   "the project latest task test-name-test-uuid failed : test-err",
			},
			err: nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockReleaseUseCase.On("ListReleasesByLabels", "test-ns", project.ProjectNameLabelKey+"=test-name").Return(nil, nil)
				mockTask.On("GetTaskState", &task.TaskSig{
					Name: "test-name",
					UUID: "test-uuid",
				}).Return(mockTaskState, nil)
				mockTaskState.On("IsFinished").Return(false)
			},
			task: &project.ProjectTask{
				Namespace: "test-ns",
				Name:      "test-name",
				LatestTaskSignature: &task.TaskSig{
					Name: "test-name",
					UUID: "test-uuid",
				},
			},
			projectInfo: &project.ProjectInfo{
				Namespace: "test-ns",
				Name:      "test-name",
				Message:   "please wait for the project latest task test-name-test-uuid finished",
			},
			err: nil,
		},
	}

	for _, test := range tests {
		test.initMock()
		projectInfo, err := mockProjectManager.buildProjectInfo(test.task)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.projectInfo, projectInfo)

		mockProjectCache.AssertExpectations(t)
		mockHelm.AssertExpectations(t)
		mockTask.AssertExpectations(t)
		mockReleaseUseCase.AssertExpectations(t)

		mockTaskState.AssertExpectations(t)
	}
}

func Test_isProjectReadyByReleases(t *testing.T) {
	tests := []struct {
		releases []*release.ReleaseInfoV2
		ready    bool
		message  string
	}{
		{
			message: noReleaseFoundMsg,
		},
		{
			releases: []*release.ReleaseInfoV2{
				{
					ReleaseInfo: release.ReleaseInfo{
						Ready: true,
					},
				},
			},
			ready: true,
		},
		{
			releases: []*release.ReleaseInfoV2{
				{
					ReleaseInfo: release.ReleaseInfo{
						Ready: true,
					},
				},
				{
					ReleaseInfo: release.ReleaseInfo{
						Ready:   false,
						Message: "pending",
					},
				},
			},
			message: "pending",
		},
	}

	for _, test := range tests {
		ready, message := isProjectReadyByReleases(test.releases)
		assert.Equal(t, test.ready, ready)
		assert.Equal(t, test.message, message)
	}
}

func TestProject_validateProjectTask(t *testing.T) {

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
		allowNotExist bool
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
			allowNotExist: true,
			err:           nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockProjectCache.On("GetProjectTask", mock.Anything, mock.Anything).Return(nil, errorModel.NotFoundError{})
			},
			allowNotExist: false,
			err:           errorModel.NotFoundError{},
		},
		{
			initMock: func() {
				refreshMocks()
				mockProjectCache.On("GetProjectTask", mock.Anything, mock.Anything).Return(&project.ProjectTask{
					LatestTaskSignature: &task.TaskSig{
						Name: "test-name",
						UUID: "test-uuid",
					},
				}, nil)
				mockTask.On("GetTaskState", mock.Anything).Return(mockTaskState, nil)
				mockTaskState.On("IsFinished").Return(true)
			},
			err: nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockProjectCache.On("GetProjectTask", mock.Anything, mock.Anything).Return(&project.ProjectTask{
					LatestTaskSignature: &task.TaskSig{
						Name: "test-name",
						UUID: "test-uuid",
					},
				}, nil)
				mockTask.On("GetTaskState", mock.Anything).Return(mockTaskState, nil)
				mockTaskState.On("IsFinished").Return(false)
				mockTaskState.On("IsTimeout").Return(false)
			},
			err: errors.New("please wait for the last project task test-name-test-uuid finished or timeout"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockProjectCache.On("GetProjectTask", mock.Anything, mock.Anything).Return(&project.ProjectTask{
					LatestTaskSignature: &task.TaskSig{
						Name: "test-name",
						UUID: "test-uuid",
					},
				}, nil)
				mockTask.On("GetTaskState", mock.Anything).Return(mockTaskState, nil)
				mockTaskState.On("IsFinished").Return(false)
				mockTaskState.On("IsTimeout").Return(true)
			},
			err: nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockProjectCache.On("GetProjectTask", mock.Anything, mock.Anything).Return(&project.ProjectTask{
					LatestTaskSignature: &task.TaskSig{
						Name: "test-name",
						UUID: "test-uuid",
					},
				}, nil)
				mockTask.On("GetTaskState", mock.Anything).Return(nil, errors.New("failed"))
			},
			err: errors.New("failed"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockProjectCache.On("GetProjectTask", mock.Anything, mock.Anything).Return(&project.ProjectTask{
					LatestTaskSignature: &task.TaskSig{
						Name: "test-name",
						UUID: "test-uuid",
					},
				}, nil)
				mockTask.On("GetTaskState", mock.Anything).Return(nil, errorModel.NotFoundError{})
			},
			err: nil,
		},
	}

	for _, test := range tests {
		test.initMock()
		_, err := mockProjectManager.validateProjectTask("test-ns", "test-name", test.allowNotExist)
		assert.IsType(t, test.err, err)

		mockProjectCache.AssertExpectations(t)
		mockHelm.AssertExpectations(t)
		mockTask.AssertExpectations(t)
		mockReleaseUseCase.AssertExpectations(t)

		mockTaskState.AssertExpectations(t)
	}
}

func TestProject_sendProjectTask(t *testing.T) {
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
		initMock func()
		oldTask  *project.ProjectTask
		async    bool
		err      error
	}{
		{
			initMock: func() {
				refreshMocks()
				mockTask.On("SendTask", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New(""))
			},
			err: errors.New("failed"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockTask.On("SendTask", mock.Anything, mock.Anything, mock.Anything).Return(&task.TaskSig{}, nil)
				mockProjectCache.On("CreateOrUpdateProjectTask", mock.Anything).Return(errors.New(""))
			},
			err: errors.New("failed"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockTask.On("SendTask", mock.Anything, mock.Anything, mock.Anything).Return(&task.TaskSig{}, nil)
				mockProjectCache.On("CreateOrUpdateProjectTask", mock.Anything).Return(nil)
			},
			async: true,
			err:   nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockTask.On("SendTask", mock.Anything, mock.Anything, mock.Anything).Return(&task.TaskSig{}, nil)
				mockProjectCache.On("CreateOrUpdateProjectTask", mock.Anything).Return(nil)
				mockTask.On("PurgeTaskState", mock.Anything).Return(nil)
			},
			oldTask: &project.ProjectTask{LatestTaskSignature: &task.TaskSig{}},
			async:   true,
			err:     nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockTask.On("SendTask", mock.Anything, mock.Anything, mock.Anything).Return(&task.TaskSig{}, nil)
				mockProjectCache.On("CreateOrUpdateProjectTask", mock.Anything).Return(nil)
				mockTask.On("PurgeTaskState", mock.Anything).Return(errors.New(""))
			},
			oldTask: &project.ProjectTask{LatestTaskSignature: &task.TaskSig{}},
			async:   true,
			err:     nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockTask.On("SendTask", mock.Anything, mock.Anything, mock.Anything).Return(&task.TaskSig{}, nil)
				mockProjectCache.On("CreateOrUpdateProjectTask", mock.Anything).Return(nil)
				mockTask.On("TouchTask", mock.Anything, mock.Anything).Return(errors.New(""))
			},
			err: errors.New(""),
		},
		{
			initMock: func() {
				refreshMocks()
				mockTask.On("SendTask", mock.Anything, mock.Anything, mock.Anything).Return(&task.TaskSig{}, nil)
				mockProjectCache.On("CreateOrUpdateProjectTask", mock.Anything).Return(nil)
				mockTask.On("TouchTask", mock.Anything, mock.Anything).Return(nil)
			},
			err: nil,
		},
	}

	for _, test := range tests {
		test.initMock()
		err := mockProjectManager.sendProjectTask("test-ns", "test", "test", nil, test.oldTask, 0, test.async)
		assert.IsType(t, test.err, err)

		mockProjectCache.AssertExpectations(t)
		mockHelm.AssertExpectations(t)
		mockTask.AssertExpectations(t)
		mockReleaseUseCase.AssertExpectations(t)

		mockTaskState.AssertExpectations(t)
	}

}

func TestProject_CreateProject(t *testing.T) {
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
			},
			projectParams: &project.ProjectParams{},
			err:           errors.New("failed"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockProjectCache.On("GetProjectTask", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
			},
			projectParams: &project.ProjectParams{
				Releases: []*release.ReleaseRequestV2{{
					ReleaseRequest: release.ReleaseRequest{
						Name:      "test",
						ChartName: "test",
					},
				}},
			},
			err: errors.New("failed"),
		},
		{
			initMock:
			func() {
				refreshMocks()
				mockProjectCache.On("GetProjectTask", mock.Anything, mock.Anything).Return(nil, errorModel.NotFoundError{})
				mockTask.On("SendTask", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New(""))
			},
			projectParams: &project.ProjectParams{
				Releases: []*release.ReleaseRequestV2{{
					ReleaseRequest: release.ReleaseRequest{
						Name:      "test",
						ChartName: "test",
					},
				}},
			},
			err: errors.New("failed"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockProjectCache.On("GetProjectTask", mock.Anything, mock.Anything).Return(nil, errorModel.NotFoundError{})
				mockTask.On("SendTask", mock.Anything, mock.Anything, mock.Anything).Return(&task.TaskSig{}, nil)
				mockProjectCache.On("CreateOrUpdateProjectTask", mock.Anything).Return(nil)
				mockTask.On("TouchTask", mock.Anything, mock.Anything).Return(nil)
			},
			projectParams: &project.ProjectParams{
				Releases: []*release.ReleaseRequestV2{{
					ReleaseRequest: release.ReleaseRequest{
						Name:      "test",
						ChartName: "test",
					},
				}},
			},
			err: nil,
		},
	}

	for _, test := range tests {
		test.initMock()
		err := mockProjectManager.CreateProject("test-ns", "test-nm", test.projectParams, false, 0)
		assert.IsType(t, test.err, err)

		mockProjectCache.AssertExpectations(t)
		mockHelm.AssertExpectations(t)
		mockTask.AssertExpectations(t)
		mockReleaseUseCase.AssertExpectations(t)

		mockTaskState.AssertExpectations(t)
	}

}

func TestProject_DeleteProject(t *testing.T) {
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
		initMock func()
		err      error
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
				mockProjectCache.On("GetProjectTask", mock.Anything, mock.Anything).Return(&project.ProjectTask{}, nil)
				mockTask.On("GetTaskState", mock.Anything).Return(nil, errorModel.NotFoundError{})
				mockTask.On("SendTask", mock.Anything, mock.Anything, mock.Anything).Return(&task.TaskSig{}, nil)
				mockProjectCache.On("CreateOrUpdateProjectTask", mock.Anything).Return(nil)
				mockTask.On("TouchTask", mock.Anything, mock.Anything).Return(nil)
			},
			err: nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockProjectCache.On("GetProjectTask", mock.Anything, mock.Anything).Return(&project.ProjectTask{}, nil)
				mockTask.On("GetTaskState", mock.Anything).Return(nil, errorModel.NotFoundError{})
				mockTask.On("SendTask", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New(""))
			},
			err: errors.New(""),
		},
	}

	for _, test := range tests {
		test.initMock()
		err := mockProjectManager.DeleteProject("test-ns", "test-nm", false, 0, false)
		assert.IsType(t, test.err, err)

		mockProjectCache.AssertExpectations(t)
		mockHelm.AssertExpectations(t)
		mockTask.AssertExpectations(t)
		mockReleaseUseCase.AssertExpectations(t)

		mockTaskState.AssertExpectations(t)
	}

}

func TestProject_AddReleasesInProject(t *testing.T) {
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
			},
			projectParams: &project.ProjectParams{},
			err:           errors.New("failed"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockProjectCache.On("GetProjectTask", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
			},
			projectParams: &project.ProjectParams{
				Releases: []*release.ReleaseRequestV2{{
					ReleaseRequest: release.ReleaseRequest{
						Name:      "test",
						ChartName: "test",
					},
				}},
			},
			err: errors.New("failed"),
		},
		{
			initMock:
			func() {
				refreshMocks()
				mockProjectCache.On("GetProjectTask", mock.Anything, mock.Anything).Return(nil, errorModel.NotFoundError{})
				mockTask.On("SendTask", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New(""))
			},
			projectParams: &project.ProjectParams{
				Releases: []*release.ReleaseRequestV2{{
					ReleaseRequest: release.ReleaseRequest{
						Name:      "test",
						ChartName: "test",
					},
				}},
			},
			err: errors.New("failed"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockProjectCache.On("GetProjectTask", mock.Anything, mock.Anything).Return(nil, errorModel.NotFoundError{})
				mockTask.On("SendTask", mock.Anything, mock.Anything, mock.Anything).Return(&task.TaskSig{}, nil)
				mockProjectCache.On("CreateOrUpdateProjectTask", mock.Anything).Return(nil)
				mockTask.On("TouchTask", mock.Anything, mock.Anything).Return(nil)
			},
			projectParams: &project.ProjectParams{
				Releases: []*release.ReleaseRequestV2{{
					ReleaseRequest: release.ReleaseRequest{
						Name:      "test",
						ChartName: "test",
					},
				}},
			},
			err: nil,
		},
	}

	for _, test := range tests {
		test.initMock()
		err := mockProjectManager.AddReleasesInProject("test-ns", "test-nm", test.projectParams, false, 0)
		assert.IsType(t, test.err, err)

		mockProjectCache.AssertExpectations(t)
		mockHelm.AssertExpectations(t)
		mockTask.AssertExpectations(t)
		mockReleaseUseCase.AssertExpectations(t)

		mockTaskState.AssertExpectations(t)
	}

}

func TestProject_UpgradeReleaseInProject(t *testing.T) {
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
		initMock func()
		err      error
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
				mockProjectCache.On("GetProjectTask", mock.Anything, mock.Anything).Return(&project.ProjectTask{}, nil)
				mockTask.On("GetTaskState", mock.Anything).Return(nil, errorModel.NotFoundError{})
				mockReleaseUseCase.On("ListReleasesByLabels", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
			},
			err: errors.New(""),
		},
		{
			initMock: func() {
				refreshMocks()
				mockProjectCache.On("GetProjectTask", mock.Anything, mock.Anything).Return(&project.ProjectTask{}, nil)
				mockTask.On("GetTaskState", mock.Anything).Return(nil, errorModel.NotFoundError{})
				mockReleaseUseCase.On("ListReleasesByLabels", mock.Anything, mock.Anything).Return(nil, nil)

			},
			err: errors.New(""),
		},
		{
			initMock: func() {
				refreshMocks()
				mockProjectCache.On("GetProjectTask", mock.Anything, mock.Anything).Return(&project.ProjectTask{}, nil)
				mockTask.On("GetTaskState", mock.Anything).Return(nil, errorModel.NotFoundError{})
				mockReleaseUseCase.On("ListReleasesByLabels", mock.Anything, mock.Anything).Return([]*release.ReleaseInfoV2{
					{
						ReleaseInfo: release.ReleaseInfo{
							ReleaseSpec: release.ReleaseSpec{
								Name: "test-name",
							},
						},
					}}, nil)

				mockTask.On("SendTask", mock.Anything, mock.Anything, mock.Anything).Return(&task.TaskSig{}, nil)
				mockProjectCache.On("CreateOrUpdateProjectTask", mock.Anything).Return(nil)
				mockTask.On("TouchTask", mock.Anything, mock.Anything).Return(nil)
			},
			err: nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockProjectCache.On("GetProjectTask", mock.Anything, mock.Anything).Return(&project.ProjectTask{}, nil)
				mockReleaseUseCase.On("ListReleasesByLabels", mock.Anything, mock.Anything).Return([]*release.ReleaseInfoV2{{
					ReleaseInfo: release.ReleaseInfo{
						ReleaseSpec: release.ReleaseSpec{
							Name: "test-name",
						},
					},
				}}, nil)
				mockTask.On("GetTaskState", mock.Anything).Return(nil, errorModel.NotFoundError{})
				mockTask.On("SendTask", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New(""))
			},
			err: errors.New(""),
		},
	}

	for _, test := range tests {
		test.initMock()
		releaseRequest := &release.ReleaseRequestV2{}
		releaseRequest.Name = "test-name"
		err := mockProjectManager.UpgradeReleaseInProject("test-ns", "test-nm", releaseRequest, false, 0)
		assert.IsType(t, test.err, err)

		mockProjectCache.AssertExpectations(t)
		mockHelm.AssertExpectations(t)
		mockTask.AssertExpectations(t)
		mockReleaseUseCase.AssertExpectations(t)

		mockTaskState.AssertExpectations(t)
	}

}

func TestProject_RemoveReleaseInProject(t *testing.T) {
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
		initMock func()
		err      error
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
				mockProjectCache.On("GetProjectTask", mock.Anything, mock.Anything).Return(&project.ProjectTask{}, nil)
				mockTask.On("GetTaskState", mock.Anything).Return(nil, errorModel.NotFoundError{})
				mockReleaseUseCase.On("ListReleasesByLabels", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
			},
			err: errors.New(""),
		},
		{
			initMock: func() {
				refreshMocks()
				mockProjectCache.On("GetProjectTask", mock.Anything, mock.Anything).Return(&project.ProjectTask{}, nil)
				mockTask.On("GetTaskState", mock.Anything).Return(nil, errorModel.NotFoundError{})
				mockReleaseUseCase.On("ListReleasesByLabels", mock.Anything, mock.Anything).Return(nil, nil)

			},
			err: nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockProjectCache.On("GetProjectTask", mock.Anything, mock.Anything).Return(&project.ProjectTask{}, nil)
				mockTask.On("GetTaskState", mock.Anything).Return(nil, errorModel.NotFoundError{})
				mockReleaseUseCase.On("ListReleasesByLabels", mock.Anything, mock.Anything).Return([]*release.ReleaseInfoV2{
					{
						ReleaseInfo: release.ReleaseInfo{
							ReleaseSpec: release.ReleaseSpec{
								Name: "test-name",
							},
						},
					}}, nil)

				mockTask.On("SendTask", mock.Anything, mock.Anything, mock.Anything).Return(&task.TaskSig{}, nil)
				mockProjectCache.On("CreateOrUpdateProjectTask", mock.Anything).Return(nil)
				mockTask.On("TouchTask", mock.Anything, mock.Anything).Return(nil)
			},
			err: nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockProjectCache.On("GetProjectTask", mock.Anything, mock.Anything).Return(&project.ProjectTask{}, nil)
				mockReleaseUseCase.On("ListReleasesByLabels", mock.Anything, mock.Anything).Return([]*release.ReleaseInfoV2{{
					ReleaseInfo: release.ReleaseInfo{
						ReleaseSpec: release.ReleaseSpec{
							Name: "test-name",
						},
					},
				}}, nil)
				mockTask.On("GetTaskState", mock.Anything).Return(nil, errorModel.NotFoundError{})
				mockTask.On("SendTask", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New(""))
			},
			err: errors.New(""),
		},
	}

	for _, test := range tests {
		test.initMock()
		err := mockProjectManager.RemoveReleaseInProject("test-ns", "test-nm", "test-name", false, 0, false)
		assert.IsType(t, test.err, err)

		mockProjectCache.AssertExpectations(t)
		mockHelm.AssertExpectations(t)
		mockTask.AssertExpectations(t)
		mockReleaseUseCase.AssertExpectations(t)

		mockTaskState.AssertExpectations(t)
	}

}

func TestProject_autoCreateReleaseDependencies(t *testing.T) {
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
		initMock        func()
		projectParams   *project.ProjectParams
		releaseRequests []*release.ReleaseRequestV2
		err             error
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
			releaseRequests: []*release.ReleaseRequestV2{
				{
					ReleaseRequest: release.ReleaseRequest{
						Name:         "A",
						ChartName:    "chartA",
						Dependencies: map[string]string{"chartB": "B"},
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
		{
			initMock: func() {
				refreshMocks()
				mockHelm.On("GetChartAutoDependencies", mock.Anything, mock.Anything, mock.Anything).Return(func(repo, chart, version string) (result []string) {
					if chart == "chartA" {
						result = append(result, "chartB")
					}
					return
				}, nil)
			},
			projectParams: &project.ProjectParams{
				Releases: []*release.ReleaseRequestV2{
					{
						ReleaseRequest: release.ReleaseRequest{
							Name:         "A",
							ChartName:    "chartA",
							Dependencies: map[string]string{"chartB": "BB"},
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
			releaseRequests: []*release.ReleaseRequestV2{
				{
					ReleaseRequest: release.ReleaseRequest{
						Name:         "A",
						ChartName:    "chartA",
						Dependencies: map[string]string{"chartB": "BB"},
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
	}

	for _, test := range tests {
		test.initMock()
		releaseRequests, err := mockProjectManager.autoCreateReleaseDependencies(test.projectParams)
		assert.IsType(t, test.err, err)
		assert.ElementsMatch(t, test.releaseRequests, releaseRequests)

		mockProjectCache.AssertExpectations(t)
		mockHelm.AssertExpectations(t)
		mockTask.AssertExpectations(t)
		mockReleaseUseCase.AssertExpectations(t)

		mockTaskState.AssertExpectations(t)
	}

}

func TestProject_autoUpdateReleaseDependencies(t *testing.T) {
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
		projectInfo   *project.ProjectInfo
		releaseParams *release.ReleaseRequestV2
		isRemove      bool

		updatedReleaseParams *release.ReleaseRequestV2
		affectedReleases     []*release.ReleaseRequestV2
		err                  error
	}{
		{
			initMock: func() {
				refreshMocks()
				mockHelm.On("GetChartAutoDependencies", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New(""))
			},
			projectInfo: &project.ProjectInfo{
				Releases: []*release.ReleaseInfoV2{
					{
						ReleaseInfo: release.ReleaseInfo{
							ReleaseSpec: release.ReleaseSpec{
								Name:      "A",
								ChartName: "chartA",
							},
						},
					},
					{
						ReleaseInfo: release.ReleaseInfo{
							ReleaseSpec: release.ReleaseSpec{
								Name:      "B",
								ChartName: "chartB",
							},
						},
					},
				},
			},
			releaseParams: &release.ReleaseRequestV2{
				ReleaseRequest: release.ReleaseRequest{
					Name:      "C",
					ChartName: "chartC",
				},
			},
			updatedReleaseParams: &release.ReleaseRequestV2{
				ReleaseRequest: release.ReleaseRequest{
					Name:      "C",
					ChartName: "chartC",
				},
			},
			err: errors.New("failed"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockHelm.On("GetChartAutoDependencies", mock.Anything, mock.Anything, mock.Anything).Return(func(repo, chart, version string) (result []string) {
					if chart == "chartA" {
						result = append(result, "chartC")
					} else if chart == "chartC" {
						result = append(result, "chartB")
					}
					return
				}, nil)
			},
			projectInfo: &project.ProjectInfo{
				Releases: []*release.ReleaseInfoV2{
					{
						ReleaseInfo: release.ReleaseInfo{
							ReleaseSpec: release.ReleaseSpec{
								Name:      "A",
								ChartName: "chartA",
							},
						},
					},
					{
						ReleaseInfo: release.ReleaseInfo{
							ReleaseSpec: release.ReleaseSpec{
								Name:      "B",
								ChartName: "chartB",
							},
						},
					},
				},
			},
			releaseParams: &release.ReleaseRequestV2{
				ReleaseRequest: release.ReleaseRequest{
					Name:      "C",
					ChartName: "chartC",
				},
			},
			updatedReleaseParams: &release.ReleaseRequestV2{
				ReleaseRequest: release.ReleaseRequest{
					Name:         "C",
					ChartName:    "chartC",
					Dependencies: map[string]string{"chartB": "B"},
				},
			},
			affectedReleases: []*release.ReleaseRequestV2{
				{
					ReleaseRequest: release.ReleaseRequest{
						Name:         "A",
						ChartName:    "chartA",
						Dependencies: map[string]string{"chartC": "C"},
					},
				},
			},
		},
		{
			initMock: func() {
				refreshMocks()
				mockHelm.On("GetChartAutoDependencies", mock.Anything, mock.Anything, mock.Anything).Return(func(repo, chart, version string) (result []string) {
					if chart == "chartA" {
						result = append(result, "chartC")
					} else if chart == "chartC" {
						result = append(result, "chartB")
					}
					return
				}, nil)
			},
			projectInfo: &project.ProjectInfo{
				Releases: []*release.ReleaseInfoV2{
					{
						ReleaseInfo: release.ReleaseInfo{
							ReleaseSpec: release.ReleaseSpec{
								Name:      "A",
								ChartName: "chartA",
								Dependencies: map[string]string{"chartC": "CC"},
							},
						},
					},
					{
						ReleaseInfo: release.ReleaseInfo{
							ReleaseSpec: release.ReleaseSpec{
								Name:      "B",
								ChartName: "chartB",
							},
						},
					},
				},
			},
			releaseParams: &release.ReleaseRequestV2{
				ReleaseRequest: release.ReleaseRequest{
					Name:      "C",
					ChartName: "chartC",
					Dependencies: map[string]string{"chartB": "BB"},
				},
			},
			updatedReleaseParams: &release.ReleaseRequestV2{
				ReleaseRequest: release.ReleaseRequest{
					Name:         "C",
					ChartName:    "chartC",
					Dependencies: map[string]string{"chartB": "BB"},
				},
			},
			affectedReleases: []*release.ReleaseRequestV2{
			},
		},
		{
			initMock: func() {
				refreshMocks()
			},
			isRemove: true,
			projectInfo: &project.ProjectInfo{
				Releases: []*release.ReleaseInfoV2{
					{
						ReleaseInfo: release.ReleaseInfo{
							ReleaseSpec: release.ReleaseSpec{
								Name:      "A",
								ChartName: "chartA",
								Dependencies: map[string]string{"chartB": "B"},
							},
						},
					},
					{
						ReleaseInfo: release.ReleaseInfo{
							ReleaseSpec: release.ReleaseSpec{
								Name:      "B",
								ChartName: "chartB",
							},
						},
					},
				},
			},
			releaseParams: &release.ReleaseRequestV2{
				ReleaseRequest: release.ReleaseRequest{
					Name:      "B",
					ChartName: "chartB",
				},
			},
			affectedReleases: []*release.ReleaseRequestV2{
				{
					ReleaseRequest: release.ReleaseRequest{
						Name:         "A",
						ChartName:    "chartA",
						Dependencies: map[string]string{"chartB": ""},
					},
				},
			},
		},
		{
			initMock: func() {
				refreshMocks()
			},
			isRemove: true,
			projectInfo: &project.ProjectInfo{
				Releases: []*release.ReleaseInfoV2{
					{
						ReleaseInfo: release.ReleaseInfo{
							ReleaseSpec: release.ReleaseSpec{
								Name:      "A",
								ChartName: "chartA",
								Dependencies: map[string]string{"chartB": "B"},
							},
						},
					},
					{
						ReleaseInfo: release.ReleaseInfo{
							ReleaseSpec: release.ReleaseSpec{
								Name:      "B",
								ChartName: "chartB",
							},
						},
					},
				},
			},
			releaseParams: &release.ReleaseRequestV2{
				ReleaseRequest: release.ReleaseRequest{
					Name:      "A",
					ChartName: "chartA",
				},
			},
			affectedReleases: []*release.ReleaseRequestV2{
			},
		},
	}

	for _, test := range tests {
		test.initMock()
		affectedReleases, err := mockProjectManager.autoUpdateReleaseDependencies(test.projectInfo, test.releaseParams, test.isRemove)
		assert.IsType(t, test.err, err)
		assert.ElementsMatch(t, test.affectedReleases, affectedReleases)
		if !test.isRemove{
			assert.Equal(t, test.updatedReleaseParams, test.releaseParams)
		}


		mockProjectCache.AssertExpectations(t)
		mockHelm.AssertExpectations(t)
		mockTask.AssertExpectations(t)
		mockReleaseUseCase.AssertExpectations(t)

		mockTaskState.AssertExpectations(t)
	}

}
