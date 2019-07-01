package mocks

import (
	"github.com/stretchr/testify/mock"
	"WarpCloud/walm/pkg/models/release"
)

type Cache struct {
	mock.Mock
}


func (cache *Cache) GetReleaseCache(namespace, name string)(*release.ReleaseCache, error) {
	args := cache.Called(namespace, name)
	return args.Get(0).(*release.ReleaseCache), args.Error(1)
}