/*
Copyright 2019 The Kubernetes Authors.

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

package fake

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
	v1beta1 "transwarp/release-config/pkg/apis/transwarp/v1beta1"
)

// FakeReleaseConfigs implements ReleaseConfigInterface
type FakeReleaseConfigs struct {
	Fake *FakeTranswarpV1beta1
	ns   string
}

var releaseconfigsResource = schema.GroupVersionResource{Group: "transwarp.k8s.io", Version: "v1beta1", Resource: "releaseconfigs"}

var releaseconfigsKind = schema.GroupVersionKind{Group: "transwarp.k8s.io", Version: "v1beta1", Kind: "ReleaseConfig"}

// Get takes name of the releaseConfig, and returns the corresponding releaseConfig object, and an error if there is any.
func (c *FakeReleaseConfigs) Get(name string, options v1.GetOptions) (result *v1beta1.ReleaseConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(releaseconfigsResource, c.ns, name), &v1beta1.ReleaseConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.ReleaseConfig), err
}

// List takes label and field selectors, and returns the list of ReleaseConfigs that match those selectors.
func (c *FakeReleaseConfigs) List(opts v1.ListOptions) (result *v1beta1.ReleaseConfigList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(releaseconfigsResource, releaseconfigsKind, c.ns, opts), &v1beta1.ReleaseConfigList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.ReleaseConfigList{}
	for _, item := range obj.(*v1beta1.ReleaseConfigList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested releaseConfigs.
func (c *FakeReleaseConfigs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(releaseconfigsResource, c.ns, opts))

}

// Create takes the representation of a releaseConfig and creates it.  Returns the server's representation of the releaseConfig, and an error, if there is any.
func (c *FakeReleaseConfigs) Create(releaseConfig *v1beta1.ReleaseConfig) (result *v1beta1.ReleaseConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(releaseconfigsResource, c.ns, releaseConfig), &v1beta1.ReleaseConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.ReleaseConfig), err
}

// Update takes the representation of a releaseConfig and updates it. Returns the server's representation of the releaseConfig, and an error, if there is any.
func (c *FakeReleaseConfigs) Update(releaseConfig *v1beta1.ReleaseConfig) (result *v1beta1.ReleaseConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(releaseconfigsResource, c.ns, releaseConfig), &v1beta1.ReleaseConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.ReleaseConfig), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeReleaseConfigs) UpdateStatus(releaseConfig *v1beta1.ReleaseConfig) (*v1beta1.ReleaseConfig, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(releaseconfigsResource, "status", c.ns, releaseConfig), &v1beta1.ReleaseConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.ReleaseConfig), err
}

// Delete takes name of the releaseConfig and deletes it. Returns an error if one occurs.
func (c *FakeReleaseConfigs) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(releaseconfigsResource, c.ns, name), &v1beta1.ReleaseConfig{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeReleaseConfigs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(releaseconfigsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1beta1.ReleaseConfigList{})
	return err
}

// Patch applies the patch and returns the patched releaseConfig.
func (c *FakeReleaseConfigs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.ReleaseConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(releaseconfigsResource, c.ns, name, data, subresources...), &v1beta1.ReleaseConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.ReleaseConfig), err
}
