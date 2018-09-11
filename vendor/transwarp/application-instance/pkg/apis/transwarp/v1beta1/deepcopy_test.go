package v1beta1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	labelSelector = map[string]string{"transwarp.install": "1", "transwarp.name": "kube"}
)

func newApplicationInstance(namespace, instanceName string, labelSelector map[string]string, status ApplicationInstanceStatus) *ApplicationInstance {
	return &ApplicationInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instanceName,
			Namespace: namespace,
			Labels:    labelSelector,
		},
		Spec: ApplicationInstanceSpec{
			ApplicationRef: ApplicationReference{
				Name:    "foo-app",
				Version: "1.0",
			},
			InstanceId: "123",
			Configs: map[string]interface{}{
				"foo-attr": "foo-value",
				"App":      map[string]interface{}{},
			},
		},
		Status: status,
	}
}
func TestDeepCopyInstanceSpec(t *testing.T) {
	ins := newApplicationInstance("default", "fake-ins", labelSelector, ApplicationInstanceStatus{})
	newIns := ins.DeepCopy()
	t.Logf("%v", newIns.Spec.Configs)
	if !reflect.DeepEqual(newIns.Spec.Configs, ins.Spec.Configs) || !reflect.DeepEqual(newIns.Spec, ins.Spec) || !reflect.DeepEqual(newIns, ins) {
		t.Errorf("\n[Failed to copy instance %v ]\nexpected:\n\t%+v unchanged! \nbut  got:\n\t%+v", ins.GetName(), ins.Spec.Configs, newIns.Spec.Configs)
	}
	newIns.Spec.Configs["test"] = nil
	if reflect.DeepEqual(newIns.Spec.Configs, ins.Spec.Configs) || reflect.DeepEqual(newIns.Spec, ins.Spec) || reflect.DeepEqual(newIns, ins) {
		t.Errorf("\n[Failed to copy instance %v ]\nexpected:\n\t%+v unchanged! \nbut  got:\n\t%+v", ins.GetName(), ins.Spec.Configs, newIns.Spec.Configs)
	}

}
