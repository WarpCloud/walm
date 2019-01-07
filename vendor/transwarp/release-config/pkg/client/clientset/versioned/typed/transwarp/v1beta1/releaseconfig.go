/*
Copyright 2018 The Kubernetes Authors.

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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
	v1beta1 "transwarp/release-config/pkg/apis/transwarp/v1beta1"
	scheme "transwarp/release-config/pkg/client/clientset/versioned/scheme"
)

// ReleaseConfigsGetter has a method to return a ReleaseConfigInterface.
// A group's client should implement this interface.
type ReleaseConfigsGetter interface {
	ReleaseConfigs(namespace string) ReleaseConfigInterface
}

// ReleaseConfigInterface has methods to work with ReleaseConfig resources.
type ReleaseConfigInterface interface {
	Create(*v1beta1.ReleaseConfig) (*v1beta1.ReleaseConfig, error)
	Update(*v1beta1.ReleaseConfig) (*v1beta1.ReleaseConfig, error)
	UpdateStatus(*v1beta1.ReleaseConfig) (*v1beta1.ReleaseConfig, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1beta1.ReleaseConfig, error)
	List(opts v1.ListOptions) (*v1beta1.ReleaseConfigList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.ReleaseConfig, err error)
	ReleaseConfigExpansion
}

// releaseConfigs implements ReleaseConfigInterface
type releaseConfigs struct {
	client rest.Interface
	ns     string
}

// newReleaseConfigs returns a ReleaseConfigs
func newReleaseConfigs(c *TranswarpV1beta1Client, namespace string) *releaseConfigs {
	return &releaseConfigs{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the releaseConfig, and returns the corresponding releaseConfig object, and an error if there is any.
func (c *releaseConfigs) Get(name string, options v1.GetOptions) (result *v1beta1.ReleaseConfig, err error) {
	result = &v1beta1.ReleaseConfig{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("releaseconfigs").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ReleaseConfigs that match those selectors.
func (c *releaseConfigs) List(opts v1.ListOptions) (result *v1beta1.ReleaseConfigList, err error) {
	result = &v1beta1.ReleaseConfigList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("releaseconfigs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested releaseConfigs.
func (c *releaseConfigs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("releaseconfigs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a releaseConfig and creates it.  Returns the server's representation of the releaseConfig, and an error, if there is any.
func (c *releaseConfigs) Create(releaseConfig *v1beta1.ReleaseConfig) (result *v1beta1.ReleaseConfig, err error) {
	result = &v1beta1.ReleaseConfig{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("releaseconfigs").
		Body(releaseConfig).
		Do().
		Into(result)
	return
}

// Update takes the representation of a releaseConfig and updates it. Returns the server's representation of the releaseConfig, and an error, if there is any.
func (c *releaseConfigs) Update(releaseConfig *v1beta1.ReleaseConfig) (result *v1beta1.ReleaseConfig, err error) {
	result = &v1beta1.ReleaseConfig{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("releaseconfigs").
		Name(releaseConfig.Name).
		Body(releaseConfig).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *releaseConfigs) UpdateStatus(releaseConfig *v1beta1.ReleaseConfig) (result *v1beta1.ReleaseConfig, err error) {
	result = &v1beta1.ReleaseConfig{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("releaseconfigs").
		Name(releaseConfig.Name).
		SubResource("status").
		Body(releaseConfig).
		Do().
		Into(result)
	return
}

// Delete takes name of the releaseConfig and deletes it. Returns an error if one occurs.
func (c *releaseConfigs) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("releaseconfigs").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *releaseConfigs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("releaseconfigs").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched releaseConfig.
func (c *releaseConfigs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.ReleaseConfig, err error) {
	result = &v1beta1.ReleaseConfig{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("releaseconfigs").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
