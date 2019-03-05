package plugins

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

// this method only can convert struct with legacy scheme
// if CRD is needed to be converted, please register the scheme first
// For example:
// import (
//    clientsetscheme "k8s.io/client-go/kubernetes/scheme"
//    transwarpscheme "transwarp/release-config/pkg/client/clientset/versioned/scheme"
// )
// transwarpscheme.AddToScheme(clientsetscheme.Scheme)
func convertUnstructured(unStruct *unstructured.Unstructured) (runtime.Object, error) {
	unStructBytes, err := unStruct.MarshalJSON()
	if err != nil {
		return nil, err
	}

	defaultGVK := unStruct.GetObjectKind().GroupVersionKind()

	decoder := scheme.Codecs.UniversalDecoder(defaultGVK.GroupVersion())
	obj, _, err := decoder.Decode(unStructBytes, &defaultGVK, nil)
	if err != nil {
		return nil,  err
	}
	return obj, nil
}
