package redis

import (
	"testing"
	"WarpCloud/walm/pkg/redis/mocks"
	"github.com/stretchr/testify/assert"
	"WarpCloud/walm/pkg/redis"
	"errors"
	"encoding/json"
	"WarpCloud/walm/pkg/models/k8s"
	"WarpCloud/walm/pkg/models/release"
	"github.com/stretchr/testify/mock"
)

func TestCache_GetReleaseCache(t *testing.T) {
	var mockRedis *mocks.Redis
	var mockCache *Cache

	refreshMocks := func() {
		mockRedis = &mocks.Redis{}
		mockCache = NewCache(mockRedis)
	}

	tests := []struct {
		initMock func()
		err      error
	}{
		{
			initMock: func() {
				refreshMocks()
				mockRedis.On("GetFieldValue", redis.WalmReleasesKey, "testns", "testnm").Return("", errors.New(""))
			},
			err: errors.New(""),
		},
		{
			initMock: func() {
				refreshMocks()
				mockRedis.On("GetFieldValue", redis.WalmReleasesKey, "testns", "testnm").Return("notvalid", nil)
			},
			err: &json.SyntaxError{},
		},
		{
			initMock: func() {
				refreshMocks()
				mockRedis.On("GetFieldValue", redis.WalmReleasesKey, "testns", "testnm").Return("{}", nil)
			},
			err: nil,
		},
	}

	for _, test := range tests {
		test.initMock()
		_, err := mockCache.GetReleaseCache("testns", "testnm")
		assert.IsType(t, test.err, err)

		mockRedis.AssertExpectations(t)
	}
}

func TestCache_GetReleaseCaches(t *testing.T) {
	var mockRedis *mocks.Redis
	var mockCache *Cache

	refreshMocks := func() {
		mockRedis = &mocks.Redis{}
		mockCache = NewCache(mockRedis)
	}

	tests := []struct {
		initMock func()
		err      error
	}{
		{
			initMock: func() {
				refreshMocks()
				mockRedis.On("GetFieldValues", redis.WalmReleasesKey, "testns").Return(nil, errors.New(""))
			},
			err: errors.New(""),
		},
		{
			initMock: func() {
				refreshMocks()
				mockRedis.On("GetFieldValues", redis.WalmReleasesKey, "testns").Return([]string{"notvalid"}, nil)
			},
			err: &json.SyntaxError{},
		},
		{
			initMock: func() {
				refreshMocks()
				mockRedis.On("GetFieldValues", redis.WalmReleasesKey, "testns").Return([]string{"{}"}, nil)
			},
			err: nil,
		},
	}

	for _, test := range tests {
		test.initMock()
		_, err := mockCache.GetReleaseCaches("testns")
		assert.IsType(t, test.err, err)

		mockRedis.AssertExpectations(t)
	}
}

func TestCache_GetReleaseCachesByReleaseConfigs(t *testing.T) {
	var mockRedis *mocks.Redis
	var mockCache *Cache

	refreshMocks := func() {
		mockRedis = &mocks.Redis{}
		mockCache = NewCache(mockRedis)
	}

	tests := []struct {
		initMock       func()
		releaseConfigs []*k8s.ReleaseConfig
		err            error
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
				mockRedis.On("GetFieldValuesByNames", redis.WalmReleasesKey, "testns/testnm1", "testns/testnm2").Return(nil, errors.New(""))
			},
			releaseConfigs: []*k8s.ReleaseConfig{
				{
					Meta: k8s.Meta{Namespace: "testns", Name: "testnm1"},
				},
				{
					Meta: k8s.Meta{Namespace: "testns", Name: "testnm2"},
				},
			},
			err: errors.New(""),
		},
		{
			initMock: func() {
				refreshMocks()
				mockRedis.On("GetFieldValuesByNames", redis.WalmReleasesKey, "testns/testnm1", "testns/testnm2").Return([]string{"notvalid", "notvalid"}, nil)
			},
			releaseConfigs: []*k8s.ReleaseConfig{
				{
					Meta: k8s.Meta{Namespace: "testns", Name: "testnm1"},
				},
				{
					Meta: k8s.Meta{Namespace: "testns", Name: "testnm2"},
				},
			},
			err: &json.SyntaxError{},
		},
		{
			initMock: func() {
				refreshMocks()
				mockRedis.On("GetFieldValuesByNames", redis.WalmReleasesKey, "testns/testnm1", "testns/testnm2").Return([]string{"", "{}"}, nil)
			},
			releaseConfigs: []*k8s.ReleaseConfig{
				{
					Meta: k8s.Meta{Namespace: "testns", Name: "testnm1"},
				},
				{
					Meta: k8s.Meta{Namespace: "testns", Name: "testnm2"},
				},
			},
			err: nil,
		},
	}

	for _, test := range tests {
		test.initMock()
		_, err := mockCache.GetReleaseCachesByReleaseConfigs(test.releaseConfigs)
		assert.IsType(t, test.err, err)

		mockRedis.AssertExpectations(t)
	}
}

func TestCache_CreateOrUpdateReleaseCache(t *testing.T) {
	var mockRedis *mocks.Redis
	var mockCache *Cache

	refreshMocks := func() {
		mockRedis = &mocks.Redis{}
		mockCache = NewCache(mockRedis)
	}

	tests := []struct {
		initMock     func()
		releaseCache *release.ReleaseCache
		err          error
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
				mockRedis.On("SetFieldValues", redis.WalmReleasesKey, mock.Anything).Return(errors.New(""))
			},
			releaseCache: &release.ReleaseCache{},
			err:          errors.New(""),
		},
		{
			initMock: func() {
				refreshMocks()
				mockRedis.On("SetFieldValues", redis.WalmReleasesKey, mock.Anything).Return(nil)
			},
			releaseCache: &release.ReleaseCache{},
			err:          nil,
		},
	}

	for _, test := range tests {
		test.initMock()
		err := mockCache.CreateOrUpdateReleaseCache(test.releaseCache)
		assert.IsType(t, test.err, err)

		mockRedis.AssertExpectations(t)
	}
}

func TestCache_DeleteReleaseCache(t *testing.T) {
	var mockRedis *mocks.Redis
	var mockCache *Cache

	refreshMocks := func() {
		mockRedis = &mocks.Redis{}
		mockCache = NewCache(mockRedis)
	}

	tests := []struct {
		initMock func()
		err      error
	}{
		{
			initMock: func() {
				refreshMocks()
				mockRedis.On("DeleteField", redis.WalmReleasesKey, "testns", "testnm").Return(errors.New(""))
			},
			err: errors.New(""),
		},
		{
			initMock: func() {
				refreshMocks()
				mockRedis.On("DeleteField", redis.WalmReleasesKey, "testns", "testnm").Return(nil)
			},
			err: nil,
		},
	}

	for _, test := range tests {
		test.initMock()
		err := mockCache.DeleteReleaseCache("testns", "testnm")
		assert.IsType(t, test.err, err)

		mockRedis.AssertExpectations(t)
	}
}

func TestCache_GetReleaseTask(t *testing.T) {
	var mockRedis *mocks.Redis
	var mockCache *Cache

	refreshMocks := func() {
		mockRedis = &mocks.Redis{}
		mockCache = NewCache(mockRedis)
	}

	tests := []struct {
		initMock func()
		err      error
	}{
		{
			initMock: func() {
				refreshMocks()
				mockRedis.On("GetFieldValue", redis.WalmReleaseTasksKey, "testns", "testnm").Return("", errors.New(""))
			},
			err: errors.New(""),
		},
		{
			initMock: func() {
				refreshMocks()
				mockRedis.On("GetFieldValue", redis.WalmReleaseTasksKey, "testns", "testnm").Return("notvalid", nil)
			},
			err: &json.SyntaxError{},
		},
		{
			initMock: func() {
				refreshMocks()
				mockRedis.On("GetFieldValue", redis.WalmReleaseTasksKey, "testns", "testnm").Return("{}", nil)
			},
			err: nil,
		},
	}

	for _, test := range tests {
		test.initMock()
		_, err := mockCache.GetReleaseTask("testns", "testnm")
		assert.IsType(t, test.err, err)

		mockRedis.AssertExpectations(t)
	}
}

func TestCache_GetReleaseTasks(t *testing.T) {
	var mockRedis *mocks.Redis
	var mockCache *Cache

	refreshMocks := func() {
		mockRedis = &mocks.Redis{}
		mockCache = NewCache(mockRedis)
	}

	tests := []struct {
		initMock func()
		err      error
	}{
		{
			initMock: func() {
				refreshMocks()
				mockRedis.On("GetFieldValues", redis.WalmReleaseTasksKey, "testns").Return(nil, errors.New(""))
			},
			err: errors.New(""),
		},
		{
			initMock: func() {
				refreshMocks()
				mockRedis.On("GetFieldValues", redis.WalmReleaseTasksKey, "testns").Return([]string{"notvalid"}, nil)
			},
			err: &json.SyntaxError{},
		},
		{
			initMock: func() {
				refreshMocks()
				mockRedis.On("GetFieldValues", redis.WalmReleaseTasksKey, "testns").Return([]string{"{}"}, nil)
			},
			err: nil,
		},
	}

	for _, test := range tests {
		test.initMock()
		_, err := mockCache.GetReleaseTasks("testns")
		assert.IsType(t, test.err, err)

		mockRedis.AssertExpectations(t)
	}
}

func TestCache_GetReleaseTasksByReleaseConfigs(t *testing.T) {
	var mockRedis *mocks.Redis
	var mockCache *Cache

	refreshMocks := func() {
		mockRedis = &mocks.Redis{}
		mockCache = NewCache(mockRedis)
	}

	tests := []struct {
		initMock       func()
		releaseConfigs []*k8s.ReleaseConfig
		err            error
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
				mockRedis.On("GetFieldValuesByNames", redis.WalmReleaseTasksKey, "testns/testnm1", "testns/testnm2").Return(nil, errors.New(""))
			},
			releaseConfigs: []*k8s.ReleaseConfig{
				{
					Meta: k8s.Meta{Namespace: "testns", Name: "testnm1"},
				},
				{
					Meta: k8s.Meta{Namespace: "testns", Name: "testnm2"},
				},
			},
			err: errors.New(""),
		},
		{
			initMock: func() {
				refreshMocks()
				mockRedis.On("GetFieldValuesByNames", redis.WalmReleaseTasksKey, "testns/testnm1", "testns/testnm2").Return([]string{"notvalid", "notvalid"}, nil)
			},
			releaseConfigs: []*k8s.ReleaseConfig{
				{
					Meta: k8s.Meta{Namespace: "testns", Name: "testnm1"},
				},
				{
					Meta: k8s.Meta{Namespace: "testns", Name: "testnm2"},
				},
			},
			err: &json.SyntaxError{},
		},
		{
			initMock: func() {
				refreshMocks()
				mockRedis.On("GetFieldValuesByNames", redis.WalmReleaseTasksKey, "testns/testnm1", "testns/testnm2").Return([]string{"", "{}"}, nil)
			},
			releaseConfigs: []*k8s.ReleaseConfig{
				{
					Meta: k8s.Meta{Namespace: "testns", Name: "testnm1"},
				},
				{
					Meta: k8s.Meta{Namespace: "testns", Name: "testnm2"},
				},
			},
			err: nil,
		},
	}

	for _, test := range tests {
		test.initMock()
		_, err := mockCache.GetReleaseTasksByReleaseConfigs(test.releaseConfigs)
		assert.IsType(t, test.err, err)

		mockRedis.AssertExpectations(t)
	}
}

func TestCache_CreateOrUpdateReleaseTask(t *testing.T) {
	var mockRedis *mocks.Redis
	var mockCache *Cache

	refreshMocks := func() {
		mockRedis = &mocks.Redis{}
		mockCache = NewCache(mockRedis)
	}

	tests := []struct {
		initMock    func()
		releaseTask *release.ReleaseTask
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
				mockRedis.On("SetFieldValues", redis.WalmReleaseTasksKey, mock.Anything).Return(errors.New(""))
			},
			releaseTask: &release.ReleaseTask{},
			err:         errors.New(""),
		},
		{
			initMock: func() {
				refreshMocks()
				mockRedis.On("SetFieldValues", redis.WalmReleaseTasksKey, mock.Anything).Return(nil)
			},
			releaseTask: &release.ReleaseTask{},
			err:         nil,
		},
	}

	for _, test := range tests {
		test.initMock()
		err := mockCache.CreateOrUpdateReleaseTask(test.releaseTask)
		assert.IsType(t, test.err, err)

		mockRedis.AssertExpectations(t)
	}
}

func TestCache_DeleteReleaseTask(t *testing.T) {
	var mockRedis *mocks.Redis
	var mockCache *Cache

	refreshMocks := func() {
		mockRedis = &mocks.Redis{}
		mockCache = NewCache(mockRedis)
	}

	tests := []struct {
		initMock func()
		err      error
	}{
		{
			initMock: func() {
				refreshMocks()
				mockRedis.On("DeleteField", redis.WalmReleaseTasksKey, "testns", "testnm").Return(errors.New(""))
			},
			err: errors.New(""),
		},
		{
			initMock: func() {
				refreshMocks()
				mockRedis.On("DeleteField", redis.WalmReleaseTasksKey, "testns", "testnm").Return(nil)
			},
			err: nil,
		},
	}

	for _, test := range tests {
		test.initMock()
		err := mockCache.DeleteReleaseTask("testns", "testnm")
		assert.IsType(t, test.err, err)

		mockRedis.AssertExpectations(t)
	}
}