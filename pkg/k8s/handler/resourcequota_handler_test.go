package handler

import (
	"testing"
	k8sclient "walm/pkg/k8s/client"
	"fmt"
	"encoding/json"
	"walm/pkg/k8s/informer"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/api/core/v1"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"
)

func TestResourceQuotaHandler(t *testing.T) {
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

	quotaHandler := ResourceQuotaHandler{client, factory.ResourceQuotaLister}

	quota := ResourceQuotaBuilder{}.Namespace("default").Name("test").AddLabel("label1","true").AddAnnotations("ann1", "true").
		AddHardResourceLimit(v1.ResourceCPU, *k8sresource.NewQuantity(10, k8sresource.DecimalExponent)).AddResourceQuotaScope(v1.ResourceQuotaScopeNotTerminating).Build()

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