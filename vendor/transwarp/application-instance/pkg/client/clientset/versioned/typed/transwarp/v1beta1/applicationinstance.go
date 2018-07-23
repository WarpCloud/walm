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
	v1beta1 "transwarp/application-instance/pkg/apis/transwarp/v1beta1"
	scheme "transwarp/application-instance/pkg/client/clientset/versioned/scheme"
)

// ApplicationInstancesGetter has a method to return a ApplicationInstanceInterface.
// A group's client should implement this interface.
type ApplicationInstancesGetter interface {
	ApplicationInstances(namespace string) ApplicationInstanceInterface
}

// ApplicationInstanceInterface has methods to work with ApplicationInstance resources.
type ApplicationInstanceInterface interface {
	Create(*v1beta1.ApplicationInstance) (*v1beta1.ApplicationInstance, error)
	Update(*v1beta1.ApplicationInstance) (*v1beta1.ApplicationInstance, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1beta1.ApplicationInstance, error)
	List(opts v1.ListOptions) (*v1beta1.ApplicationInstanceList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.ApplicationInstance, err error)
	ApplicationInstanceExpansion
}

// applicationInstances implements ApplicationInstanceInterface
type applicationInstances struct {
	client rest.Interface
	ns     string
}

// newApplicationInstances returns a ApplicationInstances
func newApplicationInstances(c *TranswarpV1beta1Client, namespace string) *applicationInstances {
	return &applicationInstances{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the applicationInstance, and returns the corresponding applicationInstance object, and an error if there is any.
func (c *applicationInstances) Get(name string, options v1.GetOptions) (result *v1beta1.ApplicationInstance, err error) {
	result = &v1beta1.ApplicationInstance{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("applicationinstances").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ApplicationInstances that match those selectors.
func (c *applicationInstances) List(opts v1.ListOptions) (result *v1beta1.ApplicationInstanceList, err error) {
	result = &v1beta1.ApplicationInstanceList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("applicationinstances").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested applicationInstances.
func (c *applicationInstances) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("applicationinstances").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a applicationInstance and creates it.  Returns the server's representation of the applicationInstance, and an error, if there is any.
func (c *applicationInstances) Create(applicationInstance *v1beta1.ApplicationInstance) (result *v1beta1.ApplicationInstance, err error) {
	result = &v1beta1.ApplicationInstance{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("applicationinstances").
		Body(applicationInstance).
		Do().
		Into(result)
	return
}

// Update takes the representation of a applicationInstance and updates it. Returns the server's representation of the applicationInstance, and an error, if there is any.
func (c *applicationInstances) Update(applicationInstance *v1beta1.ApplicationInstance) (result *v1beta1.ApplicationInstance, err error) {
	result = &v1beta1.ApplicationInstance{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("applicationinstances").
		Name(applicationInstance.Name).
		Body(applicationInstance).
		Do().
		Into(result)
	return
}

// Delete takes name of the applicationInstance and deletes it. Returns an error if one occurs.
func (c *applicationInstances) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("applicationinstances").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *applicationInstances) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("applicationinstances").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched applicationInstance.
func (c *applicationInstances) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.ApplicationInstance, err error) {
	result = &v1beta1.ApplicationInstance{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("applicationinstances").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
