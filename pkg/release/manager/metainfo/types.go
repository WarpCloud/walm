package metainfo

import (
	"github.com/tidwall/gjson"
	"github.com/sirupsen/logrus"
	"encoding/json"
)

type MetaStringConfig struct {
	MapKey       string `json:"mapKey" description:"config map values.yaml key"`
	DefaultValue string `json:"defaultValue" description:"default value of mapKey"`
	Description  string `json:"description" description:"config description"`
	Type         string `json:"type" description:"config type"`
	Required     bool   `json:"required" description:"required"`
}

func (config *MetaStringConfig) BuildDefaultValue(jsonStr string) {
	config.DefaultValue = config.BuildStringConfigValue(jsonStr)
}

func (config *MetaStringConfig) BuildStringConfigValue(jsonStr string) string {
	return gjson.Get(jsonStr, config.MapKey).String()
}

type IntConfig struct {
	MapKey       string `json:"mapKey" description:"config map values.yaml key"`
	DefaultValue int64  `json:"defaultValue" description:"default value of mapKey"`
	Description  string `json:"description" description:"config description"`
	Type         string `json:"type" description:"config type"`
	Required     bool   `json:"required" description:"required"`
}

type MetaIntConfig struct {
	IntConfig
}

func (config *MetaIntConfig) BuildDefaultValue(jsonStr string) {
	config.DefaultValue = config.BuildIntConfigValue(jsonStr)
}

func (config *MetaIntConfig) BuildIntConfigValue(jsonStr string) int64 {
	return gjson.Get(jsonStr, config.MapKey).Int()
}

type MetaEnvConfig struct {
	MapKey       string    `json:"mapKey" description:"config map values.yaml key"`
	DefaultValue []MetaEnv `json:"defaultValue" description:"default value of mapKey"`
	Description  string    `json:"description" description:"config description"`
	Type         string    `json:"type" description:"config type"`
	Required     bool      `json:"required" description:"required"`
}

func (config *MetaEnvConfig) BuildDefaultValue(jsonStr string) {
	config.DefaultValue = config.BuildEnvConfigValue(jsonStr)
}

func (config *MetaEnvConfig) BuildEnvConfigValue(jsonStr string) []MetaEnv {
	metaEnv := []MetaEnv{}
	rawMsg := gjson.Get(jsonStr, config.MapKey).Raw
	if rawMsg == "" {
		return metaEnv
	}
	err := json.Unmarshal([]byte(rawMsg), &metaEnv)
	if err != nil {
		logrus.Warnf("failed to unmarshal %s : %s", rawMsg, err.Error())
	}
	return metaEnv
}

type MetaEnv struct {
	Name  string `json:"name" description:"env name"`
	Value string `json:"value" description:"env value"`
}

type MetaBoolConfig struct {
	MapKey       string `json:"mapKey" description:"config map values.yaml key"`
	DefaultValue bool   `json:"defaultValue" description:"default value of mapKey"`
	Description  string `json:"description" description:"config description"`
	Type         string `json:"type" description:"config type"`
	Required     bool   `json:"required" description:"required"`
}

func (config *MetaBoolConfig) BuildDefaultValue(jsonStr string) {
	config.DefaultValue = config.BuildBoolConfigValue(jsonStr)
}

func (config *MetaBoolConfig) BuildBoolConfigValue(jsonStr string) bool {
	return gjson.Get(jsonStr, config.MapKey).Bool()
}

type FloatConfig struct {
	MapKey       string  `json:"mapKey" description:"config map values.yaml key"`
	DefaultValue float64 `json:"defaultValue" description:"default value of mapKey"`
	Description  string  `json:"description" description:"config description"`
	Type         string  `json:"type" description:"config type"`
	Required     bool    `json:"required" description:"required"`
}

type MetaFloatConfig struct {
	FloatConfig
}

func (config *MetaFloatConfig) BuildDefaultValue(jsonStr string) {
	config.DefaultValue = config.BuildFloatConfigValue(jsonStr)
}

func (config *MetaFloatConfig) BuildFloatConfigValue(jsonStr string) float64 {
	return gjson.Get(jsonStr, config.MapKey).Float()
}
