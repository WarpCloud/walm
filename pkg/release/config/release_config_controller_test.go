package config

import (
	"WarpCloud/walm/pkg/k8s/mocks"
	releasemocks "WarpCloud/walm/pkg/release/mocks"
	kafkamocks "WarpCloud/walm/pkg/kafka/mocks"
	"testing"
	"time"
	"github.com/stretchr/testify/mock"
	"transwarp/release-config/pkg/apis/transwarp/v1beta1"
	"WarpCloud/walm/pkg/models/k8s"
	"WarpCloud/walm/pkg/release"
	"errors"
	"WarpCloud/walm/pkg/kafka"
	errorModel "WarpCloud/walm/pkg/models/error"
	"github.com/stretchr/testify/assert"
)

func TestReleaseConfigController_enqueueReleaseConfig(t *testing.T) {
	var mockK8sCache *mocks.Cache
	var mockReleaseUseCase *releasemocks.UseCase
	var mockKafka *kafkamocks.Kafka
	var mockReleaseConfigController *ReleaseConfigController

	refreshMocks := func() {
		mockK8sCache = &mocks.Cache{}
		mockReleaseUseCase = &releasemocks.UseCase{}
		mockKafka = &kafkamocks.Kafka{}
		mockReleaseConfigController.k8sCache = mockK8sCache
		mockReleaseConfigController.releaseUseCase = mockReleaseUseCase
		mockReleaseConfigController.kafka = mockKafka
	}

	mockReleaseConfigController = NewReleaseConfigController(mockK8sCache, mockReleaseUseCase, mockKafka, 1)
	refreshMocks()
	mockK8sCache.On("AddReleaseConfigHandler", mock.Anything, mock.Anything, mock.Anything).Return()

	stopChan := make(chan struct{})
	defer close(stopChan)
	go mockReleaseConfigController.Start(stopChan)

	// wait release config controller start
	time.Sleep(time.Millisecond * 200)
	mockK8sCache.AssertExpectations(t)

	tests := []struct {
		initMock func()
		waitTime time.Duration
	}{
		{
			initMock: func() {
				refreshMocks()
				mockK8sCache.On("ListReleaseConfigs", "", "").Return(nil, nil)
			},
			waitTime: time.Millisecond * 200,
		},
		{
			initMock: func() {
				refreshMocks()
				releaseConfig := &k8s.ReleaseConfig{}
				releaseConfig.Namespace = "testns"
				releaseConfig.Name = "testnm1"
				releaseConfig.Dependencies = map[string]string{"test": "testnm"}
				mockK8sCache.On("ListReleaseConfigs", "", "").Return([]*k8s.ReleaseConfig{releaseConfig}, nil)

				mockReleaseUseCase.On("ReloadRelease", "testns", "testnm1").Return(nil)
			},
			waitTime: time.Millisecond * 200,
		},
		{
			initMock: func() {
				refreshMocks()
				releaseConfig := &k8s.ReleaseConfig{}
				releaseConfig.Namespace = "testns"
				releaseConfig.Name = "testnm1"
				releaseConfig.Dependencies = map[string]string{"test": "testnm"}
				mockK8sCache.On("ListReleaseConfigs", "", "").Return([]*k8s.ReleaseConfig{releaseConfig}, nil)

				retryTimes := 0
				mockReleaseUseCase.On("ReloadRelease", "testns", "testnm1").Return(func(string, string) error {
					if retryTimes == 0 {
						retryTimes ++
						return errors.New(release.WaitReleaseTaskMsgPrefix)
					} else {
						return nil
					}
				}).Twice()
			},
			waitTime: time.Millisecond * 1200,
		},
		{
			initMock: func() {
				refreshMocks()
				releaseConfig := &k8s.ReleaseConfig{}
				releaseConfig.Namespace = "testns"
				releaseConfig.Name = "testnm1"
				releaseConfig.Dependencies = map[string]string{"test": "testnm"}
				mockK8sCache.On("ListReleaseConfigs", "", "").Return([]*k8s.ReleaseConfig{releaseConfig}, nil)

				mockReleaseUseCase.On("ReloadRelease", "testns", "testnm1").Return(errors.New("other")).Once()
			},
			waitTime: time.Millisecond * 1200,
		},
	}

	for _, test := range tests {
		test.initMock()
		rc := &v1beta1.ReleaseConfig{}
		rc.Namespace = "testns"
		rc.Name = "testnm"

		mockReleaseConfigController.enqueueReleaseConfig(rc)
		time.Sleep(test.waitTime)

		mockReleaseUseCase.AssertExpectations(t)
		mockK8sCache.AssertExpectations(t)
		mockKafka.AssertExpectations(t)
	}
}

func TestReleaseConfigController_enqueueKafka(t *testing.T) {
	var mockK8sCache *mocks.Cache
	var mockReleaseUseCase *releasemocks.UseCase
	var mockKafka *kafkamocks.Kafka
	var mockReleaseConfigController *ReleaseConfigController

	refreshMocks := func() {
		mockK8sCache = &mocks.Cache{}
		mockReleaseUseCase = &releasemocks.UseCase{}
		mockKafka = &kafkamocks.Kafka{}
		mockReleaseConfigController.k8sCache = mockK8sCache
		mockReleaseConfigController.releaseUseCase = mockReleaseUseCase
		mockReleaseConfigController.kafka = mockKafka
	}

	mockReleaseConfigController = NewReleaseConfigController(mockK8sCache, mockReleaseUseCase, mockKafka, 1)
	refreshMocks()
	mockK8sCache.On("AddReleaseConfigHandler", mock.Anything, mock.Anything, mock.Anything).Return()

	stopChan := make(chan struct{})
	defer close(stopChan)
	go mockReleaseConfigController.Start(stopChan)

	// wait release config controller start
	time.Sleep(time.Millisecond * 200)
	mockK8sCache.AssertExpectations(t)

	tests := []struct {
		initMock func()
	}{
		{
			initMock: func() {
				refreshMocks()
				mockK8sCache.On("GetResource", k8s.ReleaseConfigKind, "testns", "testnm").Return(nil, errors.New(""))
			},
		},
		{
			initMock: func() {
				refreshMocks()
				mockK8sCache.On("GetResource", k8s.ReleaseConfigKind, "testns", "testnm").Return(nil, errorModel.NotFoundError{})
				mockKafka.On("SyncSendMessage", kafka.ReleaseConfigTopic, "{\"type\":\"Delete\",\"data\":{\"name\":\"testnm\",\"namespace\":\"testns\",\"kind\":\"\",\"state\":{\"status\":\"\",\"reason\":\"\",\"message\":\"\"},\"labels\":null,\"configValues\":null,\"dependenciesConfigValues\":null,\"dependencies\":null,\"chartName\":\"\",\"chartVersion\":\"\",\"chartAppVersion\":\"\",\"outputConfig\":null,\"repo\":\"\",\"chartImage\":\"\"}}").Return(nil)
			},
		},
		{
			initMock: func() {
				refreshMocks()
				mockK8sCache.On("GetResource", k8s.ReleaseConfigKind, "testns", "testnm").Return(&k8s.ReleaseConfig{
					Meta: k8s.Meta{
						Name:      "testnm",
						Namespace: "testns",
					},
				}, nil)
				mockReleaseUseCase.On("GetRelease", "testns", "testnm").Return(nil, errorModel.NotFoundError{})
			},
		},
		{
			initMock: func() {
				refreshMocks()
				mockK8sCache.On("GetResource", k8s.ReleaseConfigKind, "testns", "testnm").Return(&k8s.ReleaseConfig{
					Meta: k8s.Meta{
						Name:      "testnm",
						Namespace: "testns",
					},
				}, nil)
				mockReleaseUseCase.On("GetRelease", "testns", "testnm").Return(nil, nil)
				mockKafka.On("SyncSendMessage", kafka.ReleaseConfigTopic, "{\"type\":\"CreateOrUpdate\",\"data\":{\"name\":\"testnm\",\"namespace\":\"testns\",\"kind\":\"\",\"state\":{\"status\":\"\",\"reason\":\"\",\"message\":\"\"},\"labels\":null,\"configValues\":null,\"dependenciesConfigValues\":null,\"dependencies\":null,\"chartName\":\"\",\"chartVersion\":\"\",\"chartAppVersion\":\"\",\"outputConfig\":null,\"repo\":\"\",\"chartImage\":\"\"}}").Return(nil)
			},
		},
	}

	for _, test := range tests {
		test.initMock()
		rc := &v1beta1.ReleaseConfig{}
		rc.Namespace = "testns"
		rc.Name = "testnm"

		mockReleaseConfigController.enqueueKafka(rc)
		time.Sleep(time.Millisecond * 200)

		mockReleaseUseCase.AssertExpectations(t)
		mockK8sCache.AssertExpectations(t)
		mockKafka.AssertExpectations(t)
	}
}

func Test_needsEnqueueUpdatedReleaseConfig(t *testing.T) {
	tests := []struct {
		old          *v1beta1.ReleaseConfig
		cur          *v1beta1.ReleaseConfig
		needsEnqueue bool
	}{
		{
			old: &v1beta1.ReleaseConfig{
				Spec: v1beta1.ReleaseConfigSpec{
					OutputConfig: map[string]interface{}{
						"test": "true",
					},
				},
			},
			cur: &v1beta1.ReleaseConfig{
				Spec: v1beta1.ReleaseConfigSpec{
					OutputConfig: map[string]interface{}{
						"test": "false",
					},
				},
			},
			needsEnqueue: true,
		},
		{
			old: &v1beta1.ReleaseConfig{
				Spec: v1beta1.ReleaseConfigSpec{
					OutputConfig: nil,
				},
			},
			cur: &v1beta1.ReleaseConfig{
				Spec: v1beta1.ReleaseConfigSpec{
					OutputConfig: map[string]interface{}{
					},
				},
			},
			needsEnqueue: false,
		},
	}

	for _, test := range tests {
		needsEnqueue := needsEnqueueUpdatedReleaseConfig(test.old, test.cur)
		assert.Equal(t, test.needsEnqueue, needsEnqueue)
	}
}
