package redis

import (
	"testing"
	"WarpCloud/walm/pkg/redis/mocks"
	"github.com/stretchr/testify/assert"
	"WarpCloud/walm/pkg/redis"
	"errors"
	"encoding/json"
	"github.com/stretchr/testify/mock"
	"WarpCloud/walm/pkg/models/project"
)

func TestCache_GetProjectTask(t *testing.T) {
	var mockRedis *mocks.Redis
	var mockCache *Cache

	refreshMocks := func() {
		mockRedis = &mocks.Redis{}
		mockCache = NewProjectCache(mockRedis)
	}

	tests := []struct {
		initMock func()
		err      error
	}{
		{
			initMock: func() {
				refreshMocks()
				mockRedis.On("GetFieldValue", redis.WalmProjectsKey, "testns", "testnm").Return("", errors.New(""))
			},
			err: errors.New(""),
		},
		{
			initMock: func() {
				refreshMocks()
				mockRedis.On("GetFieldValue", redis.WalmProjectsKey, "testns", "testnm").Return("notvalid", nil)
			},
			err: &json.SyntaxError{},
		},
		{
			initMock: func() {
				refreshMocks()
				mockRedis.On("GetFieldValue", redis.WalmProjectsKey, "testns", "testnm").Return("{}", nil)
			},
			err: nil,
		},
	}

	for _, test := range tests {
		test.initMock()
		_, err := mockCache.GetProjectTask("testns", "testnm")
		assert.IsType(t, test.err, err)

		mockRedis.AssertExpectations(t)
	}
}

func TestCache_GetProjectTasks(t *testing.T) {
	var mockRedis *mocks.Redis
	var mockCache *Cache

	refreshMocks := func() {
		mockRedis = &mocks.Redis{}
		mockCache = NewProjectCache(mockRedis)
	}

	tests := []struct {
		initMock func()
		err      error
	}{
		{
			initMock: func() {
				refreshMocks()
				mockRedis.On("GetFieldValues", redis.WalmProjectsKey, "testns").Return(nil, errors.New(""))
			},
			err: errors.New(""),
		},
		{
			initMock: func() {
				refreshMocks()
				mockRedis.On("GetFieldValues", redis.WalmProjectsKey, "testns").Return([]string{"notvalid"}, nil)
			},
			err: &json.SyntaxError{},
		},
		{
			initMock: func() {
				refreshMocks()
				mockRedis.On("GetFieldValues", redis.WalmProjectsKey, "testns").Return([]string{"{}"}, nil)
			},
			err: nil,
		},
	}

	for _, test := range tests {
		test.initMock()
		_, err := mockCache.GetProjectTasks("testns")
		assert.IsType(t, test.err, err)

		mockRedis.AssertExpectations(t)
	}
}

func TestCache_CreateOrUpdateProjectTask(t *testing.T) {
	var mockRedis *mocks.Redis
	var mockCache *Cache

	refreshMocks := func() {
		mockRedis = &mocks.Redis{}
		mockCache = NewProjectCache(mockRedis)
	}

	tests := []struct {
		initMock    func()
		projectTask *project.ProjectTask
		err         error
	}{
		{
			initMock: func() {
				refreshMocks()
			},
			err: nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockRedis.On("SetFieldValues", redis.WalmProjectsKey, mock.Anything).Return(errors.New(""))
			},
			projectTask: &project.ProjectTask{},
			err:         errors.New(""),
		},
		{
			initMock: func() {
				refreshMocks()
				mockRedis.On("SetFieldValues", redis.WalmProjectsKey, mock.Anything).Return(nil)
			},
			projectTask: &project.ProjectTask{},
			err:         nil,
		},
	}

	for _, test := range tests {
		test.initMock()
		err := mockCache.CreateOrUpdateProjectTask(test.projectTask)
		assert.IsType(t, test.err, err)

		mockRedis.AssertExpectations(t)
	}
}

func TestCache_DeleteProjectTask(t *testing.T) {
	var mockRedis *mocks.Redis
	var mockCache *Cache

	refreshMocks := func() {
		mockRedis = &mocks.Redis{}
		mockCache = NewProjectCache(mockRedis)
	}

	tests := []struct {
		initMock func()
		err      error
	}{
		{
			initMock: func() {
				refreshMocks()
				mockRedis.On("DeleteField", redis.WalmProjectsKey, "testns", "testnm").Return(errors.New(""))
			},
			err: errors.New(""),
		},
		{
			initMock: func() {
				refreshMocks()
				mockRedis.On("DeleteField", redis.WalmProjectsKey, "testns", "testnm").Return(nil)
			},
			err: nil,
		},
	}

	for _, test := range tests {
		test.initMock()
		err := mockCache.DeleteProjectTask("testns", "testnm")
		assert.IsType(t, test.err, err)

		mockRedis.AssertExpectations(t)
	}
}

