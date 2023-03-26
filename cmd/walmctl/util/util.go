package util

import (
	"WarpCloud/walm/pkg/models/k8s"
	"WarpCloud/walm/pkg/models/release"
	"WarpCloud/walm/pkg/util"
	"encoding/json"
	"github.com/pkg/errors"
	"k8s.io/klog"
	"strconv"
)

var PLUGINS_KEY = "plugins"
var RELEASES_KEY = "releases"
var METAINFO_KEY = "metaInfoParams"
var METAINFOPARAMS_KEY = "params"

const (
	STRING = "string"
	INT    = "int"
	FLOAT  = "float"
	YAML   = "yaml"
	JSON   = "json"
	BOOL   = "boolean"
)

type CtlMetaInfoParam struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

type CtlPluginParam struct {
	Name    string      `json:"name"`
	Version string      `json:"version"`
	Args    interface{} `json:"args"`
	Disable bool        `json:"disable"`
}

func convertReleasePlugins(pluginValuesBytes []byte) ([]*k8s.ReleasePlugin, error) {
	releasePlugins := make([]*k8s.ReleasePlugin, 0)
	ctlPluginParams := make([]CtlPluginParam, 0)

	err := json.Unmarshal(pluginValuesBytes, &ctlPluginParams)
	if err != nil {
		klog.Errorf("json Unmarshal error %v", err)
		return releasePlugins, err
	}
	for _, ctlPluginParam := range ctlPluginParams {
		releasePlugin := k8s.ReleasePlugin{
			Name:    ctlPluginParam.Name,
			Args:    "{}",
			Version: ctlPluginParam.Version,
			Disable: ctlPluginParam.Disable,
		}
		ctlArgsBytes, err := json.Marshal(ctlPluginParam.Args)
		if err != nil {
			klog.Errorf("json Marshal error %v", err)
			return releasePlugins, err
		}
		releasePlugin.Args = string(ctlArgsBytes[:])
		releasePlugins = append(releasePlugins, &releasePlugin)
	}

	return releasePlugins, nil
}

func convertMetaInfoParams(metaInfoParamsBytes []byte) ([]*release.MetaCommonConfigValue, error) {
	metaInfoParams := make([]*release.MetaCommonConfigValue, 0)
	ctlMetaInfoParams := make([]CtlMetaInfoParam, 0)

	err := json.Unmarshal(metaInfoParamsBytes, &ctlMetaInfoParams)
	if err != nil {
		klog.Errorf("json Unmarshal error %v", err)
		return metaInfoParams, err
	}
	for _, ctlParam := range ctlMetaInfoParams {
		var metaInfoParamValue string
		metaInfoParam := release.MetaCommonConfigValue{
			Name:  ctlParam.Name,
			Value: ctlParam.Value,
		}
		switch ctlParam.Type {
		case STRING:
			marshalBytes, err := json.Marshal(ctlParam.Value)
			if err != nil {
				return metaInfoParams, err
			}
			metaInfoParamValue = string(marshalBytes[:])
		case INT:
			metaInfoParamInt, err := strconv.Atoi(ctlParam.Value)
			if err != nil {
				return metaInfoParams, err
			}
			marshalBytes, err := json.Marshal(metaInfoParamInt)
			if err != nil {
				return metaInfoParams, err
			}
			metaInfoParamValue = string(marshalBytes[:])
		case FLOAT:
			metaInfoParamFloat, err := strconv.ParseFloat(ctlParam.Value, 64)
			if err != nil {
				return metaInfoParams, err
			}
			marshalBytes, err := json.Marshal(metaInfoParamFloat)
			if err != nil {
				return metaInfoParams, err
			}
			metaInfoParamValue = string(marshalBytes[:])
		case BOOL:
			metaInfoParamBool, err := strconv.ParseBool(ctlParam.Value)
			if err != nil {
				return metaInfoParams, err
			}
			marshalBytes, err := json.Marshal(metaInfoParamBool)
			if err != nil {
				return metaInfoParams, err
			}
			metaInfoParamValue = string(marshalBytes[:])
		case JSON:
			//TODO
		case YAML:
			//TODO
		default:
			// default as string
			marshalBytes, err := json.Marshal(ctlParam.Value)
			if err != nil {
				return metaInfoParams, err
			}
			metaInfoParamValue = string(marshalBytes[:])
		}
		metaInfoParam.Value = metaInfoParamValue
		metaInfoParams = append(metaInfoParams, &metaInfoParam)
	}
	return metaInfoParams, nil
}

func SmartProjectConfigValues(projectConfigValues map[string]interface{}) (map[string]interface{}, error) {
	destConfigValues := make(map[string]interface{}, 0)
	util.MergeValues(destConfigValues, projectConfigValues, false)

	if releasesConfigValues, ok := destConfigValues[RELEASES_KEY].([]interface{}); ok {
		destReleasesValues := make([]interface{}, 0)
		for _, releaseValues := range releasesConfigValues {
			if releaseMap, ok := releaseValues.(map[string]interface{}); ok {
				if releaseMap[PLUGINS_KEY] != nil {
					pluginValuesBytes, err := json.Marshal(releaseMap[PLUGINS_KEY])
					if err != nil {
						klog.Errorf("json Marshal error %v", err)
						return nil, err
					}
					releasePlugins, err := convertReleasePlugins(pluginValuesBytes)
					if err != nil {
						klog.Errorf("json convert release plugins error %v", err)
						return nil, err
					}
					releaseMap[PLUGINS_KEY] = releasePlugins
				}
			}
			destReleasesValues = append(destReleasesValues, releaseValues)
		}
		destConfigValues[RELEASES_KEY] = destReleasesValues
	}

	return destConfigValues, nil
}

func SmartConfigValues(configValues map[string]interface{}) (destConfigValues map[string]interface{}, metaInfoParamParams []*release.MetaCommonConfigValue, releasePlugins []*k8s.ReleasePlugin, retErr error) {
	destConfigValues = make(map[string]interface{}, 0)
	util.MergeValues(destConfigValues, configValues, false)
	pluginValues, ok := configValues[PLUGINS_KEY]
	if ok {
		pluginValuesBytes, err := json.Marshal(pluginValues)
		if err != nil {
			klog.Errorf("json Marshal error %v", err)
			retErr = err
			return
		}
		releasePlugins, err = convertReleasePlugins(pluginValuesBytes)
		if err != nil {
			klog.Errorf("json convert release plugins error %v", err)
			return
		}

		destConfigValues[PLUGINS_KEY] = releasePlugins
	}
	metaInfoValuesInterface, ok := configValues[METAINFO_KEY]
	if ok {
		metaInfoValuesMap, ok2 := metaInfoValuesInterface.(map[string]interface{})
		if !ok2 {
			klog.Errorf("json convert release plugins error %v", metaInfoValuesInterface)
			retErr = errors.New("invaild metaInfo values")
			return
		}
		metaInfoParamsValues, ok3 := metaInfoValuesMap[METAINFOPARAMS_KEY]
		if ok3 {
			metaInfoParamsBytes, err := json.Marshal(metaInfoParamsValues)
			if err != nil {
				klog.Errorf("json marshal metainfo params error %v", metaInfoParamsValues)
				retErr = err
				return
			}
			metaInfoParamParams, err = convertMetaInfoParams(metaInfoParamsBytes)
			if err != nil {
				klog.Errorf("json convert metainfo params error %v", err)
				retErr = err
				return
			}

			destConfigValuesMap := destConfigValues[METAINFO_KEY].(map[string]interface{})
			destConfigValuesMap[METAINFOPARAMS_KEY] = metaInfoParamParams
		}
	}

	return
}
