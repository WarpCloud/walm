package informer

import (
	"testing"

	"walm/pkg/k8s/client"
	"fmt"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"
	"k8s.io/apimachinery/pkg/labels"
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

	for {
		insts, _ := factory.InstanceLister.ApplicationInstances("default").List(labels.NewSelector())
		fmt.Println(len(insts))
		//e, err := json.Marshal(insts)
		//if err != nil {
		//	fmt.Println(err)
		//	return
		//}
		//fmt.Println(string(e))
		time.Sleep(2 * time.Second)
	}

	//deployment, err := Factory.DeploymentLister.Deployments("walm").Get("walm-server")
	//e, err = json.Marshal(deployment)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//fmt.Println(string(e))
}
