package adaptor

import (
	"testing"
	"fmt"
	"encoding/json"
	"k8s.io/api/core/v1"
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

