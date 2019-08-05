package sync

import (
	"testing"
	"github.com/stretchr/testify/assert"
	releaseModel "WarpCloud/walm/pkg/models/release"
	helmmocks "WarpCloud/walm/pkg/helm/mocks"
	k8smocks "WarpCloud/walm/pkg/k8s/mocks"
	taskmocks "WarpCloud/walm/pkg/task/mocks"
	"github.com/stretchr/testify/mock"
	"errors"
	errorModel "WarpCloud/walm/pkg/models/error"
)

func Test_buildReleaseCachesFromHelmMap(t *testing.T) {
	tests := []struct {
		releaseCaches []*releaseModel.ReleaseCache
		result        map[string]*releaseModel.ReleaseCache
		err           error
	}{
		{
			releaseCaches: []*releaseModel.ReleaseCache{
				{
					ReleaseSpec: releaseModel.ReleaseSpec{
						Name:      "rel1",
						Namespace: "default",
						Version:   2,
					},
				},
				{
					ReleaseSpec: releaseModel.ReleaseSpec{
						Name:      "rel1",
						Namespace: "default",
						Version:   1,
					},
				},
				{
					ReleaseSpec: releaseModel.ReleaseSpec{
						Name:      "rel2",
						Namespace: "default",
						Version:   1,
					},
				},
			},
			result: map[string]*releaseModel.ReleaseCache{
				"default/rel1": {
					ReleaseSpec: releaseModel.ReleaseSpec{
						Name:      "rel1",
						Namespace: "default",
						Version:   2,
					},
				},
				"default/rel2": {
					ReleaseSpec: releaseModel.ReleaseSpec{
						Name:      "rel2",
						Namespace: "default",
						Version:   1,
					},
				},
			},
		},
		{
			releaseCaches: []*releaseModel.ReleaseCache{
				{
					ReleaseSpec: releaseModel.ReleaseSpec{
						Name:      "rel1",
						Namespace: "default",
						Version:   1,
					},
				},
				{
					ReleaseSpec: releaseModel.ReleaseSpec{
						Name:      "rel1",
						Namespace: "default",
						Version:   2,
					},
				},
				{
					ReleaseSpec: releaseModel.ReleaseSpec{
						Name:      "rel2",
						Namespace: "default",
						Version:   1,
					},
				},
			},
			result: map[string]*releaseModel.ReleaseCache{
				"default/rel1": {
					ReleaseSpec: releaseModel.ReleaseSpec{
						Name:      "rel1",
						Namespace: "default",
						Version:   2,
					},
				},
				"default/rel2": {
					ReleaseSpec: releaseModel.ReleaseSpec{
						Name:      "rel2",
						Namespace: "default",
						Version:   1,
					},
				},
			},
		},
	}

	for _, test := range tests {
		result, err := buildReleaseCachesFromHelmMap(test.releaseCaches)
		assert.IsType(t, test.err, err)

		expectedResult, err := convertReleaseCachesMapToStrMap(test.result)
		assert.IsType(t, test.err, err)

		assert.Equal(t, expectedResult, result)
	}
}

func TestSync_buildReleaseTasksToDel(t *testing.T) {
	var mockHelm *helmmocks.Helm
	var mockK8sCache *k8smocks.Cache
	var mockTask *taskmocks.Task
	var mockTaskState *taskmocks.TaskState
	var mockSync *Sync

	refreshMocks := func() {
		mockHelm = &helmmocks.Helm{}
		mockK8sCache = &k8smocks.Cache{}
		mockTask = &taskmocks.Task{}
		mockTaskState = &taskmocks.TaskState{}

		mockSync = NewSync(nil, mockHelm, mockK8sCache, mockTask, "", "", "")
	}

	tests := []struct {
		initMock             func()
		releaseTasksFromHelm map[string]string
		releaseTaskInRedis   map[string]string
		results              []string
		err                  error
	}{
		{
			initMock: func() {
				refreshMocks()
			},
			releaseTasksFromHelm: map[string]string{"test": "{}"},
			releaseTaskInRedis: map[string]string{"test": "{}"},
			results: []string{},
			err: nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockTask.On("GetTaskState", mock.Anything).Return(nil, errors.New(""))
			},
			releaseTasksFromHelm: map[string]string{},
			releaseTaskInRedis: map[string]string{"test": "{}"},
			results: nil,
			err: errors.New(""),
		},
		{
			initMock: func() {
				refreshMocks()
				mockTask.On("GetTaskState", mock.Anything).Return(nil, errorModel.NotFoundError{})
			},
			releaseTasksFromHelm: map[string]string{},
			releaseTaskInRedis: map[string]string{"test": "{}"},
			results: []string{"test"},
			err: nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockTask.On("GetTaskState", mock.Anything).Return(mockTaskState, nil)
				mockTaskState.On("IsFinished").Return(true)
			},
			releaseTasksFromHelm: map[string]string{},
			releaseTaskInRedis: map[string]string{"test": "{}"},
			results: []string{"test"},
			err: nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockTask.On("GetTaskState", mock.Anything).Return(mockTaskState, nil)
				mockTaskState.On("IsFinished").Return(false)
				mockTaskState.On("IsTimeout").Return(true)
			},
			releaseTasksFromHelm: map[string]string{},
			releaseTaskInRedis: map[string]string{"test": "{}"},
			results: []string{"test"},
			err: nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockTask.On("GetTaskState", mock.Anything).Return(mockTaskState, nil)
				mockTaskState.On("IsFinished").Return(false)
				mockTaskState.On("IsTimeout").Return(false)
			},
			releaseTasksFromHelm: map[string]string{},
			releaseTaskInRedis: map[string]string{"test": "{}"},
			results: []string{},
			err: nil,
		},
	}

	for _, test := range tests {
		test.initMock()
		result, err := mockSync.buildReleaseTasksToDel(test.releaseTasksFromHelm, test.releaseTaskInRedis)
		assert.IsType(t, test.err, err)
		assert.ElementsMatch(t, test.results, result)

		mockHelm.AssertExpectations(t)
		mockK8sCache.AssertExpectations(t)
		mockTask.AssertExpectations(t)
	}
}

func TestSync_buildProjectTasksToDel(t *testing.T) {
	var mockHelm *helmmocks.Helm
	var mockK8sCache *k8smocks.Cache
	var mockTask *taskmocks.Task
	var mockTaskState *taskmocks.TaskState
	var mockSync *Sync

	refreshMocks := func() {
		mockHelm = &helmmocks.Helm{}
		mockK8sCache = &k8smocks.Cache{}
		mockTask = &taskmocks.Task{}
		mockTaskState = &taskmocks.TaskState{}

		mockSync = NewSync(nil, mockHelm, mockK8sCache, mockTask, "", "", "")
	}

	tests := []struct {
		initMock                      func()
		projectTasksFromReleaseConfig map[string]string
		projectTaskInRedis            map[string]string
		results                       []string
		err                           error
	}{
		{
			initMock: func() {
				refreshMocks()
			},
			projectTasksFromReleaseConfig: map[string]string{"test": "{}"},
			projectTaskInRedis:            map[string]string{"test": "{}"},
			results:                       []string{},
			err:                           nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockTask.On("GetTaskState", mock.Anything).Return(nil, errors.New(""))
			},
			projectTasksFromReleaseConfig: map[string]string{},
			projectTaskInRedis:            map[string]string{"test": "{}"},
			results:                       nil,
			err:                           errors.New(""),
		},
		{
			initMock: func() {
				refreshMocks()
				mockTask.On("GetTaskState", mock.Anything).Return(nil, errorModel.NotFoundError{})
			},
			projectTasksFromReleaseConfig: map[string]string{},
			projectTaskInRedis:            map[string]string{"test": "{}"},
			results:                       []string{"test"},
			err:                           nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockTask.On("GetTaskState", mock.Anything).Return(mockTaskState, nil)
				mockTaskState.On("IsFinished").Return(true)
			},
			projectTasksFromReleaseConfig: map[string]string{},
			projectTaskInRedis:            map[string]string{"test": "{}"},
			results:                       []string{"test"},
			err:                           nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockTask.On("GetTaskState", mock.Anything).Return(mockTaskState, nil)
				mockTaskState.On("IsFinished").Return(false)
				mockTaskState.On("IsTimeout").Return(true)
			},
			projectTasksFromReleaseConfig: map[string]string{},
			projectTaskInRedis:            map[string]string{"test": "{}"},
			results:                       []string{"test"},
			err:                           nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockTask.On("GetTaskState", mock.Anything).Return(mockTaskState, nil)
				mockTaskState.On("IsFinished").Return(false)
				mockTaskState.On("IsTimeout").Return(false)
			},
			projectTasksFromReleaseConfig: map[string]string{},
			projectTaskInRedis:            map[string]string{"test": "{}"},
			results:                       []string{},
			err:                           nil,
		},
	}

	for _, test := range tests {
		test.initMock()
		result, err := mockSync.buildProjectTasksToDel(test.projectTasksFromReleaseConfig, test.projectTaskInRedis)
		assert.IsType(t, test.err, err)
		assert.ElementsMatch(t, test.results, result)

		mockHelm.AssertExpectations(t)
		mockK8sCache.AssertExpectations(t)
		mockTask.AssertExpectations(t)
	}
}