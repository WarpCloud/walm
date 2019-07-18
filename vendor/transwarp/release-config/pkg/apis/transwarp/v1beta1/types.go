/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ReleaseConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReleaseConfigSpec   `json:"spec"`
	Status ReleaseConfigStatus `json:"status"`
}

// ReleaseConfigSpec is the spec for a ReleaseConfig resource
type ReleaseConfigSpec struct {
	ConfigValues             map[string]interface{} `json:"configValues" description:"user config values added to the chart"`
	DependenciesConfigValues map[string]interface{} `json:"dependenciesConfigValues" description:"dependencies' config values added to the chart"`
	Dependencies             map[string]string      `json:"dependencies" description:"map of dependency chart name and release"`
	ChartName                string                 `json:"chartName" description:"chart name"`
	ChartVersion             string                 `json:"chartVersion" description:"chart version"`
	ChartAppVersion          string                 `json:"chartAppVersion" description:"jsonnet app version"`
	OutputConfig             map[string]interface{} `json:"outputConfig"`
	Repo                     string                 `json:"repo" description:"chart repo"`
	ChartImage               string                 `json:"chartImage" description:"chart image"`
}

// ReleaseConfigStatus is the status for a ReleaseConfig resource
type ReleaseConfigStatus struct{}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ReleaseConfigList is a list of ReleaseConfig resources
type ReleaseConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ReleaseConfig `json:"items"`
}
