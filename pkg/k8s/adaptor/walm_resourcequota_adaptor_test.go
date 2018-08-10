package adaptor

import (
	"testing"
	"fmt"
	"encoding/json"
	"k8s.io/api/core/v1"
	"walm/pkg/k8s/informer"
	k8sclient "walm/pkg/k8s/client"
	"k8s.io/apimachinery/pkg/util/wait"
	"walm/pkg/k8s/handler"
)

func TestResourceQuotaAdaptor(t *testing.T) {
	walmResourceQuota := &WalmResourceQuota{
		WalmMeta: buildWalmMetaWithoutState("ResourceQuota", "default", "test"),
		ResourceLimits: map[v1.ResourceName]string{v1.ResourceCPU: "10", v1.ResourceMemory: "5Gi"},
	}

	quota, err := BuildResourceQuota(walmResourceQuota)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	e, err := json.Marshal(quota)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	fmt.Println(string(e))

	walmQuota := BuildWalmResourceQuota(quota)
	e, err = json.Marshal(walmQuota)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	fmt.Println(string(e))
}

func TestCreateResourceQuota(t *testing.T) {
	client, err := k8sclient.CreateFakeApiserverClient("", "C:/kubernetes/0.5/kubeconfig")
	if err != nil {
		println(err.Error())
		return
	}

	clientEx, err := k8sclient.CreateFakeApiserverClientEx("", "C:/kubernetes/0.5/kubeconfig")
	if err != nil {
		println(err.Error())
		return
	}

	factory := informer.NewFakeInformerFactory(client, clientEx, 0)
	factory.Start(wait.NeverStop)
	factory.WaitForCacheSync(wait.NeverStop)

	quotaHandler := handler.NewFakeResourceQuotaHandler(client, factory.ResourceQuotaLister)
	walmResourceQuota := &WalmResourceQuota{
		WalmMeta: buildWalmMetaWithoutState("ResourceQuota", "default", "test"),
		ResourceLimits: map[v1.ResourceName]string{v1.ResourceCPU: "10", v1.ResourceMemory: "5Gi"},
	}

	quota, err := BuildResourceQuota(walmResourceQuota)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	newQuota, err := quotaHandler.CreateResourceQuota("default", quota)
	if err != nil {
		fmt.Println(err)
	}

	e, err := json.Marshal(newQuota)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(e))

	newQuota, err = quotaHandler.GetResourceQuota("default", "test")
	if err != nil {
		fmt.Println(err)
	}

	e, err = json.Marshal(newQuota)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(e))

	err = quotaHandler.DeleteResourceQuota("default", "test")
	if err != nil {
		fmt.Println(err)
	}
}