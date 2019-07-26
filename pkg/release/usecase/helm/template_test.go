package helm

import (
	"testing"
	"WarpCloud/walm/pkg/models/release"
	"github.com/stretchr/testify/assert"
	"WarpCloud/walm/pkg/release/mocks"
	helmMocks "WarpCloud/walm/pkg/helm/mocks"
	k8sMocks "WarpCloud/walm/pkg/k8s/mocks"
	"github.com/stretchr/testify/mock"
	"errors"
	errorModel "WarpCloud/walm/pkg/models/error"
)

func TestHelm_ComputeResourcesByDryRunRelease(t *testing.T) {

	var mockReleaseCache  *mocks.Cache
	var mockHelm *helmMocks.Helm
	var mockK8sOperator *k8sMocks.Operator
	var mockReleaseManager *Helm

	refreshMocks := func() {
		mockReleaseCache = &mocks.Cache{}
		mockHelm = &helmMocks.Helm{}
		mockK8sOperator = &k8sMocks.Operator{}
		mockReleaseManager = &Helm{}
		mockReleaseManager.releaseCache = mockReleaseCache
		mockReleaseManager.helm = mockHelm
		mockReleaseManager.k8sOperator = mockK8sOperator
	}

	tests := []struct{
		initMock func()
		resources *release.ReleaseResources
		err error
	} {
		{
			initMock: func() {
				refreshMocks()
				mockReleaseCache.On("GetReleaseCache", mock.Anything, mock.Anything).Return(nil, errors.New("failed to get release cache"))
			},
			resources: nil,
			err: errors.New("failed to get release cache"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockReleaseCache.On("GetReleaseCache", mock.Anything, mock.Anything).Return(nil, errorModel.NotFoundError{})
				mockHelm.On("InstallOrCreateRelease", mock.Anything, mock.Anything, mock.Anything, true, false, (*release.ReleaseInfoV2)(nil), (*bool)(nil)).Return(&release.ReleaseCache{Manifest: "test-manifest"}, nil)
				mockK8sOperator.On("ComputeReleaseResourcesByManifest", mock.Anything, "test-manifest").Return(&release.ReleaseResources{}, nil)
			},
			resources: &release.ReleaseResources{},
			err: nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockReleaseCache.On("GetReleaseCache", mock.Anything, mock.Anything).Return(nil, errorModel.NotFoundError{})
				mockHelm.On("InstallOrCreateRelease", mock.Anything, mock.Anything, mock.Anything, true, false, (*release.ReleaseInfoV2)(nil), (*bool)(nil)).Return(&release.ReleaseCache{Manifest: "test-manifest"}, nil)
				mockK8sOperator.On("ComputeReleaseResourcesByManifest", mock.Anything, "test-manifest").Return(nil, errors.New("failed to compute"))
			},
			resources: nil,
			err: errors.New("failed to compute"),
		},
	}

	for _, test := range tests {
		test.initMock()
		releaseRequest := &release.ReleaseRequestV2{}
		releaseRequest.Name = "test-release"
		resources, err := mockReleaseManager.ComputeResourcesByDryRunRelease("test-ns", releaseRequest, nil)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.resources, resources)

		mockReleaseCache.AssertExpectations(t)
		mockHelm.AssertExpectations(t)
		mockK8sOperator.AssertExpectations(t)
	}
}

func TestHelm_DryRunRelease(t *testing.T) {

	var mockReleaseCache  *mocks.Cache
	var mockHelm *helmMocks.Helm
	var mockK8sOperator *k8sMocks.Operator
	var mockReleaseManager *Helm

	refreshMocks := func() {
		mockReleaseCache = &mocks.Cache{}
		mockHelm = &helmMocks.Helm{}
		mockK8sOperator = &k8sMocks.Operator{}
		mockReleaseManager = &Helm{}
		mockReleaseManager.releaseCache = mockReleaseCache
		mockReleaseManager.helm = mockHelm
		mockReleaseManager.k8sOperator = mockK8sOperator
	}

	tests := []struct{
		initMock func()
		resources []map[string]interface{}
		err error
	} {
		{
			initMock: func() {
				refreshMocks()
				mockReleaseCache.On("GetReleaseCache", mock.Anything, mock.Anything).Return(nil, errors.New("failed to get release cache"))
			},
			resources: nil,
			err: errors.New("failed to get release cache"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockReleaseCache.On("GetReleaseCache", mock.Anything, mock.Anything).Return(nil, errorModel.NotFoundError{})
				mockHelm.On("InstallOrCreateRelease", mock.Anything, mock.Anything, mock.Anything, true, false, (*release.ReleaseInfoV2)(nil), (*bool)(nil)).Return(&release.ReleaseCache{Manifest: "test-manifest"}, nil)
				mockK8sOperator.On("BuildManifestObjects", mock.Anything, "test-manifest").Return([]map[string]interface{}{}, nil)
			},
			resources: []map[string]interface{}{},
			err: nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockReleaseCache.On("GetReleaseCache", mock.Anything, mock.Anything).Return(nil, errorModel.NotFoundError{})
				mockHelm.On("InstallOrCreateRelease", mock.Anything, mock.Anything, mock.Anything, true, false, (*release.ReleaseInfoV2)(nil), (*bool)(nil)).Return(&release.ReleaseCache{Manifest: "test-manifest"}, nil)
				mockK8sOperator.On("BuildManifestObjects", mock.Anything, "test-manifest").Return(nil, errors.New("failed to build"))
			},
			resources: nil,
			err: errors.New("failed to build"),
		},
	}

	for _, test := range tests {
		test.initMock()
		releaseRequest := &release.ReleaseRequestV2{}
		releaseRequest.Name = "test-release"
		resources, err := mockReleaseManager.DryRunRelease("test-ns", releaseRequest, nil)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.resources, resources)

		mockReleaseCache.AssertExpectations(t)
		mockHelm.AssertExpectations(t)
		mockK8sOperator.AssertExpectations(t)
	}
}

