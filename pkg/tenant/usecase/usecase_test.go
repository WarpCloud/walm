package usecase

import (
	k8sMocks "WarpCloud/walm/pkg/k8s/mocks"
	errorModel "WarpCloud/walm/pkg/models/error"
	"WarpCloud/walm/pkg/models/release"
	"WarpCloud/walm/pkg/models/tenant"

	"WarpCloud/walm/pkg/release/mocks"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)


func TestTenant_CreateTenant(t *testing.T) {
	var mockK8sCache *k8sMocks.Cache
	var mockK8sOperator *k8sMocks.Operator
	var mockReleaseUseCase *mocks.UseCase
	var mockTenantManager *Tenant

	refreshMocks := func() {
		mockK8sCache = &k8sMocks.Cache{}
		mockK8sOperator = &k8sMocks.Operator{}
		mockReleaseUseCase = &mocks.UseCase{}
		mockTenantManager = NewTenant(mockK8sCache, mockK8sOperator, mockReleaseUseCase)
	}

	tests := []struct {
		initMock func()
		err      error
	}{
		{
			initMock: func() {
				refreshMocks()
				mockK8sCache.On("GetTenant", mock.Anything).Return(nil, nil)
			},
			err: nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockK8sCache.On("GetTenant", mock.Anything).Return(nil, errors.New("failed"))
			},
			err: errors.New("failed"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockK8sCache.On("GetTenant", mock.Anything).Return(nil, errorModel.NotFoundError{})
				mockK8sOperator.On("CreateNamespace", mock.Anything).Return(errors.New("failed"))
			},
			err: errors.New("failed"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockK8sCache.On("GetTenant", mock.Anything).Return(nil, errorModel.NotFoundError{})
				mockK8sOperator.On("CreateNamespace", mock.Anything).Return(nil)
				mockK8sOperator.On("CreateResourceQuota", mock.Anything).Return(nil)
				mockK8sOperator.On("CreateLimitRange", mock.Anything).Return(nil)
			},
			err: nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockK8sCache.On("GetTenant", mock.Anything).Return(nil, errorModel.NotFoundError{})
				mockK8sOperator.On("CreateNamespace", mock.Anything).Return(nil)
				mockK8sOperator.On("CreateResourceQuota", mock.Anything).Return(nil)
				mockK8sOperator.On("CreateLimitRange", mock.Anything).Return(errors.New("failed"))
				mockK8sOperator.On("DeleteNamespace", mock.Anything).Return(nil)
			},
			err: errors.New("failed"),
		},
	}
	testTenantParams := &tenant.TenantParams{
		TenantLabels:      map[string]string{"labels1": "labels1"},
		TenantQuotas: []*tenant.TenantQuotaParams{
			{
				QuotaName: "test-quota",
				Hard: &tenant.TenantQuotaInfo{
					LimitCpu:        "1",
					LimitMemory:     "500m",
				},
			},
		},
	}
	for _, test := range tests {
		test.initMock()
		err := mockTenantManager.CreateTenant("test-tenant", testTenantParams)
		assert.IsType(t, test.err, err)
		mockK8sCache.AssertExpectations(t)
		mockK8sOperator.AssertExpectations(t)
		mockReleaseUseCase.AssertExpectations(t)
	}
}

func TestTenant_GetTenant(t *testing.T) {
	var mockK8sCache *k8sMocks.Cache
	var mockK8sOperator *k8sMocks.Operator
	var mockReleaseUseCase *mocks.UseCase
	var mockTenantManager *Tenant

	refreshMocks := func() {
		mockK8sCache = &k8sMocks.Cache{}
		mockK8sOperator = &k8sMocks.Operator{}
		mockReleaseUseCase = &mocks.UseCase{}
		mockTenantManager = NewTenant(mockK8sCache, mockK8sOperator, mockReleaseUseCase)
	}

	tests := []struct {
		initMock   func()
		tenantInfo *tenant.TenantInfo
		err        error
	}{
		{
			initMock: func() {
				refreshMocks()
				mockK8sCache.On("GetTenant", mock.Anything).Return(nil, errors.New(""))
			},
			tenantInfo: nil,
			err:        errors.New(""),
		},
		{
			initMock: func() {
				refreshMocks()
				mockK8sCache.On("GetTenant", mock.Anything).Return(
					&tenant.TenantInfo{
						TenantName:   "test-tenant",
						TenantLabels: map[string]string{"test1": "test1"},
						Ready:        true,
					}, nil)
			},
			tenantInfo: &tenant.TenantInfo{
				TenantName:   "test-tenant",
				TenantLabels: map[string]string{"test1": "test1"},
				Ready:        true,
			},
			err: nil,
		},
	}
	for _, test := range tests {
		test.initMock()
		tenantInfo, err := mockTenantManager.GetTenant("test-tenant")
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.tenantInfo, tenantInfo)
		mockK8sCache.AssertExpectations(t)
		mockK8sOperator.AssertExpectations(t)
		mockReleaseUseCase.AssertExpectations(t)
	}
}

func TestTenant_ListTenants(t *testing.T) {
	var mockK8sCache *k8sMocks.Cache
	var mockK8sOperator *k8sMocks.Operator
	var mockReleaseUseCase *mocks.UseCase
	var mockTenantManager *Tenant

	refreshMocks := func() {
		mockK8sCache = &k8sMocks.Cache{}
		mockK8sOperator = &k8sMocks.Operator{}
		mockReleaseUseCase = &mocks.UseCase{}
		mockTenantManager = NewTenant(mockK8sCache, mockK8sOperator, mockReleaseUseCase)
	}

	tests := []struct {
		initMock       func()
		tenantInfoList *tenant.TenantInfoList
		err            error
	}{
		{
			initMock: func() {
				refreshMocks()
				mockK8sCache.On("ListTenants", mock.Anything).Return(
					&tenant.TenantInfoList{
						Items: []*tenant.TenantInfo{
							{
								TenantName:   "test-tenant",
								TenantLabels: map[string]string{"test1": "test1"},
								Ready:        true,
							},
						}}, nil)
			},
			tenantInfoList: &tenant.TenantInfoList{
				Items: []*tenant.TenantInfo{
					{
						TenantName:   "test-tenant",
						TenantLabels: map[string]string{"test1": "test1"},
						Ready:        true,
					},
				}},
			err: nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockK8sCache.On("ListTenants", mock.Anything).Return( nil, errors.New("failed"))
			},
			err: errors.New("failed"),
		},
	}
	for _, test := range tests {
		test.initMock()
		tenantInfoList, err := mockTenantManager.ListTenants()
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.tenantInfoList, tenantInfoList)
		mockK8sCache.AssertExpectations(t)
		mockK8sOperator.AssertExpectations(t)
		mockReleaseUseCase.AssertExpectations(t)
	}
}

func TestTenant_DeleteTenant(t *testing.T) {
	var mockK8sCache *k8sMocks.Cache
	var mockK8sOperator *k8sMocks.Operator
	var mockReleaseUseCase *mocks.UseCase
	var mockTenantManager *Tenant

	refreshMocks := func() {
		mockK8sCache = &k8sMocks.Cache{}
		mockK8sOperator = &k8sMocks.Operator{}
		mockReleaseUseCase = &mocks.UseCase{}
		mockTenantManager = NewTenant(mockK8sCache, mockK8sOperator, mockReleaseUseCase)
	}
	tests := []struct {
		initMock func()
		err      error
	}{
		{
			initMock: func() {
				refreshMocks()
				mockK8sCache.On("GetTenant", mock.Anything).Return(nil, errors.New("failed"))
			},
			err: errors.New("failed"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockK8sCache.On("GetTenant", mock.Anything).Return(nil, errorModel.NotFoundError{})
			},
			err: nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockK8sCache.On("GetTenant", mock.Anything).Return(nil, nil)
				mockReleaseUseCase.On("ListReleases", mock.Anything).Return(nil, errors.New("failed"))
			},
			err: errors.New("failed"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockK8sCache.On("GetTenant", mock.Anything).Return(nil, nil)
				mockReleaseUseCase.On("ListReleases", mock.Anything).Return([]*release.ReleaseInfoV2{
					{
						ReleaseInfo: release.ReleaseInfo{
							ReleaseSpec: release.ReleaseSpec{
								Name:      "test-release",
								Namespace: "test-tenant",
							},
						},
					},
				}, nil)
				mockReleaseUseCase.On("DeleteReleaseWithRetry", mock.Anything, mock.Anything,false,false, int64(0)).Return(
					errors.New("failed"),
				)
			},
			err: errors.New("failed"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockK8sCache.On("GetTenant", mock.Anything).Return(nil, nil)
				mockReleaseUseCase.On("ListReleases", mock.Anything).Return([]*release.ReleaseInfoV2{
					{
						ReleaseInfo: release.ReleaseInfo{
							ReleaseSpec: release.ReleaseSpec{
								Name:      "test-release",
								Namespace: "test-tenant",
							},
						},
					},
				}, nil)
				mockReleaseUseCase.On("DeleteReleaseWithRetry", mock.Anything, mock.Anything,false,false,int64(0)).Return(
					nil,
				)
				mockK8sOperator.On("DeleteNamespace", mock.Anything).Return(errors.New("failed"))
			},
			err: errors.New("failed"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockK8sCache.On("GetTenant", mock.Anything).Return(nil, nil)
				mockReleaseUseCase.On("ListReleases", mock.Anything).Return([]*release.ReleaseInfoV2{
					{
						ReleaseInfo: release.ReleaseInfo{
							ReleaseSpec: release.ReleaseSpec{
								Name:      "test-release",
								Namespace: "test-tenant",
							},
						},
					},
				}, nil)
				mockReleaseUseCase.On("DeleteReleaseWithRetry", mock.Anything, mock.Anything,false,false, int64(0)).Return(
					nil,
				)
				mockK8sOperator.On("DeleteNamespace", mock.Anything).Return(nil)
			},
			err: nil,
		},
	}
	for _, test := range tests {
		test.initMock()
		err := mockTenantManager.DeleteTenant("test-tenant")
		assert.IsType(t, test.err, err)
		mockK8sCache.AssertExpectations(t)
		mockK8sOperator.AssertExpectations(t)
		mockReleaseUseCase.AssertExpectations(t)
	}
}

func TestTenant_UpdateTenant(t *testing.T) {
	var mockK8sCache *k8sMocks.Cache
	var mockK8sOperator *k8sMocks.Operator
	var mockReleaseUseCase *mocks.UseCase
	var mockTenantManager *Tenant

	refreshMocks := func() {
		mockK8sCache = &k8sMocks.Cache{}
		mockK8sOperator = &k8sMocks.Operator{}
		mockReleaseUseCase = &mocks.UseCase{}

		mockTenantManager = NewTenant(mockK8sCache, mockK8sOperator, mockReleaseUseCase)
	}
	tests := []struct {
		initMock func()
		err      error
	}{
		{
			initMock: func() {
				refreshMocks()
				mockK8sCache.On("GetTenant", mock.Anything).Return(nil, errors.New("failed"))
			},
			err: errors.New("failed"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockK8sCache.On("GetTenant", mock.Anything).Return(nil, nil)
				mockK8sOperator.On("UpdateNamespace", mock.Anything).Return(errors.New("failed"))
			},
			err: errors.New("failed"),
		},
		{
			initMock: func() {
				refreshMocks()
				mockK8sCache.On("GetTenant", mock.Anything).Return(nil, nil)
				mockK8sOperator.On("UpdateNamespace", mock.Anything).Return(nil)
				mockK8sOperator.On("CreateOrUpdateResourceQuota", mock.Anything).Return(nil)
			},
			err: nil,
		},
		{
			initMock: func() {
				refreshMocks()
				mockK8sCache.On("GetTenant", mock.Anything).Return(nil, nil)
				mockK8sOperator.On("UpdateNamespace", mock.Anything).Return(nil)
				mockK8sOperator.On("CreateOrUpdateResourceQuota", mock.Anything).Return(errors.New("failed"))
			},
			err: errors.New("failed"),
		},
	}
	for _, test := range tests {
		test.initMock()
		err := mockTenantManager.UpdateTenant("test-tenant", &tenant.TenantParams{
			TenantLabels: map[string]string{"test1": "test1"},
			TenantQuotas:      []*tenant.TenantQuotaParams{
				{
					QuotaName: "test-quota",
					Hard: &tenant.TenantQuotaInfo{
						LimitCpu: limitRangeDefaultCpu,
					},
				},
			},
		})
		assert.IsType(t, test.err, err)
		mockK8sCache.AssertExpectations(t)
		mockK8sOperator.AssertExpectations(t)
		mockReleaseUseCase.AssertExpectations(t)
	}
}
