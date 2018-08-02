package instance

//import (
//	"testing"
//	"walm/pkg/k8s/client"
//	"encoding/json"
//	"fmt"
//	"walm/pkg/instance/lister"
//	"walm/pkg/k8s/handler"
//	"walm/pkg/k8s/informer"
//	"k8s.io/apimachinery/pkg/util/wait"
//	"time"
//)
//
//func Test(t *testing.T) {
//	clientEx, err := client.CreateApiserverClientEx("", "C:/kubernetes/0.5/kubeconfig")
//	if err != nil {
//		println(err.Error())
//		return
//	}
//
//	client, err := client.CreateApiserverClient("", "C:/kubernetes/0.5/kubeconfig")
//	if err != nil {
//		println(err.Error())
//		return
//	}
//
//	factory := informer.NewInformerFactory(client, clientEx, 0)
//	factory.Start(wait.NeverStop)
//	factory.WaitForCacheSync(wait.NeverStop)
//
//	for {
//		inst, err := handler.NewInstanceHandler(clientEx, factory.InstanceLister).GetInstance("default","cy-test2")
//		if err != nil {
//			println(err.Error())
//			time.Sleep(3 * time.Second)
//			continue
//		}
//
//		lister := lister.K8sResourceLister{factory, client}
//		instManager := InstanceManager{lister}
//		walmInst, err := instManager.BuildWalmApplicationInstance(*inst)
//		if err != nil {
//			println(err.Error())
//			return
//		}
//
//		e, err := json.Marshal(walmInst.Status.WalmModules)
//		if err != nil {
//			fmt.Println(err)
//			return
//		}
//		fmt.Println(string(e))
//
//		e, err = json.Marshal(walmInst.Status.InstanceState)
//		if err != nil {
//			fmt.Println(err)
//			return
//		}
//		fmt.Println(string(e))
//		time.Sleep(3 * time.Second)
//	}
//
//
//}


