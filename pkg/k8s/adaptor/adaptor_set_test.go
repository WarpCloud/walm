package adaptor

import (
	"testing"
)

func Test(t *testing.T) {
	//clientEx, err := client.CreateFakeApiserverClientEx("", "C:/kubernetes/0.5/kubeconfig")
	//if err != nil {
	//	println(err.Error())
	//	return
	//}
	//
	//client, err := client.CreateFakeApiserverClient("", "C:/kubernetes/0.5/kubeconfig")
	//if err != nil {
	//	println(err.Error())
	//	return
	//}
	//
	//factory := informer.NewFakeInformerFactory(client, clientEx, 0)
	//factory.Start(wait.NeverStop)
	//factory.WaitForCacheSync(wait.NeverStop)
	//
	//handlerSet := handler.NewFakeHandlerSet(client, clientEx, factory)
	//
	//adaptorSet := AdaptorSet{handlerSet: handlerSet}
	//
	//instanceAdaptor := adaptorSet.GetAdaptor("ApplicationInstance")
	//
	//inst, err := instanceAdaptor.GetResource("nosecurity", "zookeeper-nosecurity")
	//
	//e, err := json.Marshal(inst)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//fmt.Println(string(e))

}
