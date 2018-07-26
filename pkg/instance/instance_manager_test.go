package instance

import (
	"testing"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"walm/pkg/k8s/client"
	"encoding/json"
	"fmt"
	"walm/pkg/instance/lister"
)

func Test(t *testing.T) {
	clientEx, err := client.CreateApiserverClientEx("", "C:/kubernetes/0.5/kubeconfig")
	if err != nil {
		println(err.Error())
		return
	}

	inst, err := clientEx.TranswarpV1beta1().ApplicationInstances("txsql3").Get("txsql-txsql3", v1.GetOptions{})
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

	e, err = json.Marshal(walmInst.Status.Events)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(e))

}


