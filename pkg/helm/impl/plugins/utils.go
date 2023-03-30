package plugins

import (
	"encoding/json"
	"github.com/tidwall/sjson"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
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
		return nil, err
	}
	return obj, nil
}

func buildStatefulSet(obj runtime.Object) (*appsv1.StatefulSet, error) {
	if statefulSet, ok := obj.(*appsv1.StatefulSet); ok {
		return statefulSet, nil
	} else {
		objBytes, err := json.Marshal(obj)
		if err != nil {
			return nil, err
		}
		objStr := string(objBytes)
		objStr, err = sjson.Set(objStr, "apiVersion", "apps/v1")
		if err != nil {
			return nil, err
		}
		statefulSet = &appsv1.StatefulSet{}
		err = json.Unmarshal([]byte(objStr), statefulSet)
		if err != nil {
			return nil, err
		}
		return statefulSet, nil
	}
}

func buildJob(obj runtime.Object) (*batchv1.Job, error) {
	if job, ok := obj.(*batchv1.Job); ok {
		return job, nil
	} else {
		objBytes, err := json.Marshal(obj)
		if err != nil {
			return nil, err
		}
		objStr := string(objBytes)
		objStr, err = sjson.Set(objStr, "apiVersion", "batch/v1")
		if err != nil {
			return nil, err
		}
		job = &batchv1.Job{}
		err = json.Unmarshal([]byte(objStr), job)
		if err != nil {
			return nil, err
		}
		return job, nil
	}
}

func buildDaemonSet(obj runtime.Object) (*appsv1.DaemonSet, error) {
	if daemonSet, ok := obj.(*appsv1.DaemonSet); ok {
		return daemonSet, nil
	} else {
		objBytes, err := json.Marshal(obj)
		if err != nil {
			return nil, err
		}
		objStr := string(objBytes)
		objStr, err = sjson.Set(objStr, "apiVersion", "apps/v1")
		if err != nil {
			return nil, err
		}
		daemonSet = &appsv1.DaemonSet{}
		err = json.Unmarshal([]byte(objStr), daemonSet)
		if err != nil {
			return nil, err
		}
		return daemonSet, nil
	}
}

func buildDeployment(obj runtime.Object) (*appsv1.Deployment, error) {
	if deployment, ok := obj.(*appsv1.Deployment); ok {
		return deployment, nil
	} else {
		objBytes, err := json.Marshal(obj)
		if err != nil {
			return nil, err
		}
		objStr := string(objBytes)
		objStr, err = sjson.Set(objStr, "apiVersion", "apps/v1")
		if err != nil {
			return nil, err
		}
		deployment = &appsv1.Deployment{}
		err = json.Unmarshal([]byte(objStr), deployment)
		if err != nil {
			return nil, err
		}
		return deployment, nil
	}
}

func buildConfigmap(obj runtime.Object) (*v1.ConfigMap, error) {
	if configmap, ok := obj.(*v1.ConfigMap); ok {
		return configmap, nil
	} else {
		objBytes, err := json.Marshal(obj)
		if err != nil {
			return nil, err
		}
		objStr := string(objBytes)
		objStr, err = sjson.Set(objStr, "apiVersion", "v1")
		if err != nil {
			return nil, err
		}
		configmap = &v1.ConfigMap{}
		err = json.Unmarshal([]byte(objStr), configmap)
		if err != nil {
			return nil, err
		}
		return configmap, nil
	}
}

func buildService(obj runtime.Object) (*v1.Service, error) {
	if service, ok := obj.(*v1.Service); ok {
		return service, nil
	} else {
		objBytes, err := json.Marshal(obj)
		if err != nil {
			return nil, err
		}
		objStr := string(objBytes)
		objStr, err = sjson.Set(objStr, "apiVersion", "v1")
		if err != nil {
			return nil, err
		}
		service = &v1.Service{}
		err = json.Unmarshal([]byte(objStr), service)
		if err != nil {
			return nil, err
		}
		return service, nil
	}
}

func buildIngress(obj runtime.Object) (*v1beta1.Ingress, error) {
	if ingress, ok := obj.(*v1beta1.Ingress); ok {
		return ingress, nil
	} else {
		objBytes, err := json.Marshal(obj)
		if err != nil {
			return nil, err
		}
		objStr := string(objBytes)
		objStr, err = sjson.Set(objStr, "apiVersion", "extensions/v1beta1")
		if err != nil {
			return nil, err
		}
		ingress = &v1beta1.Ingress{}
		err = json.Unmarshal([]byte(objStr), ingress)
		if err != nil {
			return nil, err
		}
		return ingress, nil
	}
}

func convertToUnstructured(obj runtime.Object) (runtime.Object, error) {
	var unstructuredObj unstructured.Unstructured
	objBytes, err := json.Marshal(obj)
	if err != nil {
		return unstructuredObj.DeepCopyObject(), err
	}
	objMap := make(map[string]interface{}, 0)
	err = json.Unmarshal(objBytes, &objMap)
	if err != nil {
		return unstructuredObj.DeepCopyObject(), err
	}

	unstructuredObj.SetUnstructuredContent(objMap)
	return unstructuredObj.DeepCopyObject(), nil
}
