package metainfo

import (
	"github.com/tidwall/gjson"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/resource"
	"strconv"
	"fmt"
	"encoding/json"
	"walm/pkg/util"
)

type MetaResourceIntConfig struct {
	IntConfig
}

func (config *MetaResourceIntConfig) BuildDefaultValue(jsonStr string) {
	config.DefaultValue = config.BuildIntConfigValue(jsonStr)
}

func (config *MetaResourceIntConfig) BuildIntConfigValue(jsonStr string) int64 {
	quantity := parseK8sResourceQuantity(jsonStr, config.MapKey)
	if quantity != nil {
		return quantity.Value() / util.K8sResourceMemoryScale
	}
	return 0
}

func parseK8sResourceQuantity(jsonStr, mapKey string) *resource.Quantity {
	if jsonStr == "" || mapKey == "" {
		return nil
	}
	strValue := gjson.Get(jsonStr, mapKey).String()
	if strValue == "" {
		return nil
	}
	quantity, err := resource.ParseQuantity(strValue)
	if err != nil {
		logrus.Warnf("failed to parse quantity %s : %s", strValue, err.Error())
		return nil
	}
	return &quantity
}

type MetaResourceFloatConfig struct {
	FloatConfig
}

func (config *MetaResourceFloatConfig) BuildDefaultValue(jsonStr string) {
	config.DefaultValue = config.BuildFloatConfigValue(jsonStr)
}

func (config *MetaResourceFloatConfig) BuildFloatConfigValue(jsonStr string) float64 {
	quantity := parseK8sResourceQuantity(jsonStr, config.MapKey)
	if quantity != nil {
		return float64(quantity.MilliValue()) / util.K8sResourceCpuScale
	}
	return 0
}

type ResourceStorage struct {
	AccessModes  []string `json:"accessModes, omitempty" description:"storage access modes"`
	StorageClass string   `json:"storageClass" description:"storage class"`
}

type MetaResourceStorage struct {
	ResourceStorage
	Size int64 `json:"size" description:"storage size"`
}

type MetaResourceStorageWithStringSize struct {
	ResourceStorage
	Size string `json:"size" description:"storage size"`
}

type MetaResourceStorageConfig struct {
	Name         string               `json:"name" description:"config name"`
	MapKey       string               `json:"mapKey" description:"config map values.yaml key"`
	DefaultValue *MetaResourceStorage `json:"defaultValue" description:"default value of mapKey"`
	Description  string               `json:"description" description:"config description"`
	Type         string               `json:"type" description:"config type"`
	Required     bool                 `json:"required" description:"required"`
}

func (config *MetaResourceStorageConfig) BuildDefaultValue(jsonStr string) {
	config.DefaultValue = config.BuildStorageConfigValue(jsonStr).Value
}

func (config *MetaResourceStorageConfig) BuildStorageConfigValue(jsonStr string) *MetaResourceStorageConfigValue {
	resourceStorageConfigValue := &MetaResourceStorageConfigValue{
		Name: config.Name,
	}
	resourceStorageWithStringSize := parseResourceStorageWithStringSize(jsonStr, config.MapKey)
	if resourceStorageWithStringSize != nil {
		resourceStorageConfigValue.Value = &MetaResourceStorage{
			ResourceStorage: resourceStorageWithStringSize.ResourceStorage,
		}

		if resourceStorageWithStringSize.Size != "" {
			quantity, err := resource.ParseQuantity(resourceStorageWithStringSize.Size)
			if err != nil {
				logrus.Warnf("failed to parse quantity %s : %s", resourceStorageWithStringSize.Size, err.Error())
				return nil
			}
			resourceStorageConfigValue.Value.Size = quantity.Value() / util.K8sResourceStorageScale
		}
	}
	return resourceStorageConfigValue
}

func parseResourceStorageWithStringSize(jsonStr, mapKey string) *MetaResourceStorageWithStringSize {
	rawMsg := gjson.Get(jsonStr, mapKey).Raw
	if rawMsg == "" {
		return nil
	}
	resourceStorage := &MetaResourceStorageWithStringSize{}
	err := json.Unmarshal([]byte(rawMsg), resourceStorage)
	if err != nil {
		logrus.Warnf("failed to unmarshal %s : %s", rawMsg, err.Error())
		return nil
	}
	return resourceStorage
}

func convertResourceBinaryIntByUnit(i *int64, unit string) string {
	return strconv.FormatInt(*i, 10) + unit
}

func convertResourceDecimalFloat(f *float64) string {
	return fmt.Sprintf("%g", *f)
}