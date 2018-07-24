package instance


import (
	"testing"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"walm/pkg/k8s/client"
	"walm/pkg/instance/lister"
	"encoding/json"
	"fmt"
)

func Test(t *testing.T) {
	clientEx, err := client.CreateApiserverClientEx("", "C:/kubernetes/0.5/kubeconfig")
	if err != nil {
		println(err.Error())
		return
	}

	inst, err := clientEx.TranswarpV1beta1().ApplicationInstances("guardian").Get("guardian-guardian", v1.GetOptions{})
	if err != nil {
		println(err.Error())
		return
	}

	client, err := client.CreateApiserverClient("", "C:/kubernetes/0.5/kubeconfig")
	if err != nil {
		println(err.Error())
		return
	}

	lister := lister.K8sResourceLister{client}
	instManager := InstanceManager{lister}
	walmInst, err := instManager.BuildWalmApplicationInstance(*inst)
	if err != nil {
		println(err.Error())
		return
	}

	e, err := json.Marshal(walmInst.Status.WalmModules)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(e))

}


