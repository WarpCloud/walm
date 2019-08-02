package release

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"encoding/json"
)

func TestMetaInfoParams_BuildConfigValues(t *testing.T) {
	image := "test-image"
	replicas := int64(2)
	useHostNetwork := true
	priority := int64(1)

	requestCpu := float64(0.1)
	limitsCpu := float64(0.2)
	requestsMem := int64(200)
	limitsMem := int64(400)
	requestsGpu := float64(1)
	limitsGpu := float64(2)

	tests := []struct {
		params       *MetaInfoParams
		metaInfo     *ChartMetaInfo
		configValues string
		err          error
	}{
		{
			params: &MetaInfoParams{
				Roles: []*MetaRoleConfigValue{
					{
						Name: "test-role",
						RoleBaseConfigValue: &MetaRoleBaseConfigValue{
							Image:          &image,
							Replicas:       &replicas,
							UseHostNetwork: &useHostNetwork,
							Priority:       &priority,
							Env: []MetaEnv{
								{
									Name:  "test-key",
									Value: "test-value",
								},
							},
							Others: []*MetaCommonConfigValue{
								{
									Name:  "test-other",
									Type:  "string",
									Value: "\"test-other-value\"",
								},
							},
						},
						RoleResourceConfigValue: &MetaResourceConfigValue{
							RequestsCpu:    &requestCpu,
							LimitsCpu:      &limitsCpu,
							RequestsMemory: &requestsMem,
							LimitsMemory:   &limitsMem,
							RequestsGpu:    &requestsGpu,
							LimitsGpu:      &limitsGpu,
							StorageResources: []*MetaResourceStorageConfigValue{
								{
									Name: "test-storage",
									Value: &MetaResourceStorage{
										ResourceStorage: ResourceStorage{
											AccessModes:  []string{"ReadOnly"},
											StorageClass: "silver",
										},
										Size: 100,
									},
								},
							},
						},
					},
				},
				Params: []*MetaCommonConfigValue{
					{
						Name:  "test-params",
						Type:  "string",
						Value: "\"test-params-value\"",
					},
				},
			},
			metaInfo: &ChartMetaInfo{
				ChartRoles: []*MetaRoleConfig{
					{
						Name: "test-role",
						RoleBaseConfig: &MetaRoleBaseConfig{
							Image: &MetaStringConfig{
								MapKey: "image.application.image",
							},
							Env: &MetaEnvConfig{
								MapKey: "envs",
							},
							Priority: &MetaIntConfig{
								IntConfig: IntConfig{
									MapKey: "priority",
								},
							},
							UseHostNetwork: &MetaBoolConfig{
								MapKey: "useHostNetwork",
							},
							Replicas: &MetaIntConfig{
								IntConfig: IntConfig{
									MapKey: "replicas",
								},
							},
							Others: []*MetaCommonConfig{
								{
									Name:   "test-other",
									MapKey: "test-other",
								},
							},
						},
						RoleResourceConfig: &MetaResourceConfig{
							RequestsCpu: &MetaResourceCpuConfig{
								FloatConfig: FloatConfig{
									MapKey: "resources.requestsCpu",
								},
							},
							LimitsCpu: &MetaResourceCpuConfig{
								FloatConfig: FloatConfig{
									MapKey: "resources.limitsCpu",
								},
							},
							RequestsMemory: &MetaResourceMemoryConfig{
								IntConfig: IntConfig{
									MapKey: "resources.requestsMem",
								},
							},
							LimitsMemory: &MetaResourceMemoryConfig{
								IntConfig: IntConfig{
									MapKey: "resources.LimitsMem",
								},
							},
							RequestsGpu: &MetaResourceCpuConfig{
								FloatConfig: FloatConfig{
									MapKey: "resources.requestsGpu",
								},
							},
							LimitsGpu: &MetaResourceCpuConfig{
								FloatConfig: FloatConfig{
									MapKey: "resources.limitsGpu",
								},
							},
							StorageResources: []*MetaResourceStorageConfig{
								{
									Name:   "test-storage",
									MapKey: "storage",
								},
							},
						},
					},
				},
				ChartParams: []*MetaCommonConfig{
					{
						Name:   "test-params",
						MapKey: "image.java.command",
					},
				},
			},
			configValues: "{\"envs\":[{\"name\":\"test-key\",\"value\":\"test-value\"}],\"image\":{\"application\":{\"image\":\"test-image\"},\"java\":{\"command\":\"test-params-value\"}},\"priority\":1,\"replicas\":2,\"resources\":{\"LimitsMem\":\"400Mi\",\"limitsCpu\":\"0.2\",\"limitsGpu\":\"2\",\"requestsCpu\":\"0.1\",\"requestsGpu\":\"1\",\"requestsMem\":\"200Mi\"},\"storage\":{\"accessModes\":[\"ReadOnly\"],\"size\":\"100Gi\",\"storageClass\":\"silver\"},\"test-other\":\"test-other-value\",\"useHostNetwork\":true}",
			err:          nil,
		},
	}

	for _, test := range tests {
		configValues, err := test.params.BuildConfigValues(test.metaInfo)
		assert.IsType(t, test.err, err)

		configValuesStr, err := json.Marshal(configValues)
		assert.IsType(t, nil, err)
		assert.Equal(t, test.configValues, string(configValuesStr))
	}
}
