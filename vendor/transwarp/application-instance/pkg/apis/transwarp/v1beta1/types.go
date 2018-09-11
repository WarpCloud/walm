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
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ApplicationInstance struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired behavior of the ApplicationInstance.
	// +optional
	Spec ApplicationInstanceSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// Most recently observed status of the ApplicationInstance. This data
	// may be out of date by some window of time.
	// +optional
	Status ApplicationInstanceStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// An ApplicationInstanceSpec is the specification of an ApplicationInstance.
type ApplicationInstanceSpec struct {
	// Required: application reference determines the template of an ApplicationInstance.
	ApplicationRef ApplicationReference `json:"applicationRef" protobuf:"bytes,1,opt,name=applicationRef"`

	// Install Id to select the resource of the ApplicationInstance.
	// +optional
	InstanceId string `json:"instanceId,omitempty" protobuf:"bytes,2,opt,name=instanceId"`

	// +optional
	Configs map[string]interface{} `json:"configs,omitempty" protobuf:"bytes,3,opt,name=configs"`

	// +optional
	Dependencies []Dependency `json:"dependencies,omitempty" protobuf:"bytes,4,rep,name=dependencies"`
}

type ApplicationReference struct {
	Name        string `json:"name" protobuf:"bytes,1,opt,name=name"`
	Version     string `json:"version" protobuf:"bytes,2,opt,name=version"`
	Storage     string `json:"storage,omitempty" protobuf:"bytes,3,opt,name=storage"`
	StorageName string `json:"storagename,omitempty" protobuf:"bytes,4,opt,name=storagename"`
}

type Dependency struct {
	// +optional
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	// +optional
	DependencyRef v1.ObjectReference `json:"dependencyRef,omitempty" protobuf:"bytes,2,opt,name=dependencyRef"`
}

// ApplicationInstanceStatus represents the current status of an ApplicationInstance.
type ApplicationInstanceStatus struct {
	// The generation observed by the instance manager.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,1,opt,name=observedGeneration"`

	// The current lifecycle phase of the applicationinstance.
	// +optional
	Phase string `json:"phase,omitempty" protobuf:"bytes,2,opt,name=phase,casttype=string"`

	// Whether all of pods in applicationinstance are ready. ?
	// +optional
	Ready bool `json:"ready" protobuf:"varint,3,opt,name=ready"`

	// List of modules of an ApplicationInstance.
	// +optional
	Modules []ResourceReference `json:"modules,omitempty" protobuf:"bytes,4,rep,name=modules"`

	// List of ApplicationInstances that self depends on.
	// +optional
	DependsOn []v1.ObjectReference `json:"dependsOn,omitempty" protobuf:"bytes,5,rep,name=dependsOn"`

	// List of ApplicationInstances that self depended by.
	// +optional
	DependedBy []v1.ObjectReference `json:"dependedBy,omitempty" protobuf:"bytes,6,rep,name=dependedBy"`

	// List of Hooks of an ApplicationInstance.
	// +optional
	Hooks map[string]HookReference `json:"hooks,omitempty" protobuf:"bytes,7,rep,name=hooks"`
}

type ApplicationInstancePhase string

// These are the valid phases of ApplicationInstance.
const (
	// ApplicationInstanceDeploying means one or more of the modules has not been created.
	ApplicationInstanceDeploying string = "deploying"
	// ApplicationInstancePending means the instance has been accepted by the system,
	// but one or more of the modules has not been started.
	ApplicationInstancePending string = "pending"
	// ApplicationInstanceRunning means all of the modules have been started.
	ApplicationInstanceRunning string = "running"
)

// ResourceReference contains enough information about an resource of an ApplicationInstance.
type ResourceReference struct {
	// +optional
	ResourceRef v1.ObjectReference `json:"resourceRef,omitempty" protobuf:"bytes,1,opt,name=resourceRef"`
	// +optional
	Ready bool `json:"ready" protobuf:"varint,2,opt,name=ready"`
	// +optional
	Ports []ServiceNodePort `json:"ports,omitempty" protobuf:"bytes,3,rep,name=ports"`
	// +optional
	ReleaseInfo string `json:"releaseinfo,omitempty" protobuf:"bytes,4,opt,name=releaseinfo"`
}

// HookReference contains enough information about hooks of an ApplicationInstance.
type HookReference struct {
	// +optional
	Objects []ResourceReference `json:"objects,omitempty" protobuf:"bytes,1,rep,name=objects"`
	// +optional
	Succeed bool `json:"succeed" protobuf:"varint,2,opt,name=succeed"`
}

type ServiceNodePort struct {
	// +optional
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	// +optional
	NodePort int32 `json:"nodePort,omitempty" protobuf:"varint,2,opt,name=nodePort"`
	// +optional
	Ready bool `json:"ready,omitempty" protobuf:"varint,3,opt,name=ready"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// ApplicationInstanceList is a collection of ApplicationInstances.
type ApplicationInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []ApplicationInstance `json:"items" protobuf:"bytes,2,rep,name=items"`
}
