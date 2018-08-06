package informer

import (
	"testing"

	"walm/pkg/k8s/client"
	"fmt"
	"k8s.io/apimachinery/pkg/util/wait"
	"encoding/json"
)

func Test(t *testing.T) {

	client1, err := client.CreateFakeApiserverClient("", "C:/kubernetes/0.5/kubeconfig")
	if err != nil {
		fmt.Println(err.Error())
	}

	clientEx, err := client.CreateFakeApiserverClientEx("", "C:/kubernetes/0.5/kubeconfig")
	if err != nil {
		fmt.Println(err.Error())
	}

	factory := newInformerFactory(client1, clientEx, 0)
	factory.Start(wait.NeverStop)
	factory.WaitForCacheSync(wait.NeverStop)

	factory.factory.WaitForCacheSync(wait.NeverStop)
	factory.factoryEx.WaitForCacheSync(wait.NeverStop)

	inst, err := factory.InstanceLister.ApplicationInstances("txsql3").Get("txsql-txsql3")
	e, err := json.Marshal(inst)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(e))

	deployment, err := factory.DeploymentLister.Deployments("walm").Get("walm-server")
	e, err = json.Marshal(deployment)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(e))
}