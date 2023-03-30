package plugins

import (
	"testing"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"encoding/json"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func convertObjToUnstructured(obj interface{}) *unstructured.Unstructured {
	objBytes, _ := json.Marshal(obj)
	objMap := map[string]interface{}{}
	json.Unmarshal(objBytes, &objMap)
	return &unstructured.Unstructured{
		Object: objMap,
	}
}

func Test_setNestedStringMap(t *testing.T) {
	tests := []struct {
		obj            map[string]interface{}
		stringMapToAdd map[string]string
		fields         []string
		err            error
		result         map[string]interface{}
	}{
		{
			obj: convertObjToUnstructured(&appsv1.StatefulSet{}).Object,
			stringMapToAdd: map[string]string{"test": "test"},
			fields: []string{"spec", "template", "metadata", "labels"},
			result: convertObjToUnstructured(&appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"test": "test"},
						},
					},
				},
			}).Object,
		},
	}

	for _, test := range tests {
		err := setNestedStringMap(test.obj, test.stringMapToAdd, test.fields...)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.result, test.obj)
	}
}

