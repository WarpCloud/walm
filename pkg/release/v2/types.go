package v2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ReleaseConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReleaseConfigSpec   `json:"spec"`
	Status ReleaseConfigStatus `json:"status"`
}

// ReleaseConfigSpec is the spec for a ReleaseConfig resource
type ReleaseConfigSpec struct {
	InputConfig  ReleaseInputConfig `json:"input_config"`
	OutputConfig map[string]interface{} `json:"output_config"`
}

type ReleaseInputConfig struct {
	UserConfig  map[string]interface{} `json:"user_config"`
	CharName     string                 `json:"chart_name"`
	ChartVersion string                 `json:"chart_version"`
	Dependencies map[string]string      `json:"dependencies"`
}

// ReleaseConfigStatus is the status for a ReleaseConfig resource
type ReleaseConfigStatus struct {
	// 判断release config有没有被operator处理
	ObservedGeneration int64 `json:"observed_generation"`
	// inputConfig 有没有生效
	InputConfigState ReleaseConfigState
	// 上游release有没有更新依赖配置
	OutputConfigState ReleaseConfigState
}

type ReleaseConfigState struct {
	Status string
	Message string
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ReleaseConfigList is a list of ReleaseConfig resources
type ReleaseConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ReleaseConfig `json:"items"`
}
