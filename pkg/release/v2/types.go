package v2

import "walm/pkg/release"

type ReleaseInfoV2 struct {
	release.ReleaseInfo
	DependenciesConfigValues map[string]interface{} `json:"dependencies_config_values" description:"release's dependencies' config values"`
	ComputedValues           map[string]interface{} `json:"computed_values" description:"config values to render chart templates"`
	OutputConfigValues       map[string]interface{} `json:"output_config_values" description:"release's output config values'"`
}

type ReleaseRequestV2 struct {
	release.ReleaseRequest
}

type ReleaseInfoV2List struct {
	Num   int              `json:"num" description:"release num"`
	Items []*ReleaseInfoV2 `json:"items" description:"release infos"`
}
