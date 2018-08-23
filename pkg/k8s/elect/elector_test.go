package elect

import (
	"testing"
	"walm/pkg/k8s/client"
	"fmt"
	"time"
)

func Test(t *testing.T) {
	client, err := client.CreateFakeApiserverClient("", "C:/kubernetes/0.5/kubeconfig")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	config1 := &ElectorConfig{
		Client: client,
		ElectionId: "walm-test",
		LockIdentity: "client1",
		LockNamespace: "walmtest",
		OnStartedLeadingFunc: func(stop <-chan struct{}) {
			fmt.Println("client1 is leader")
		},
		OnStoppedLeadingFunc: func() {
			fmt.Println("client1 is not leader any more")
		},
		OnNewLeaderFunc: func(identity string) {
			fmt.Println(fmt.Sprintf("%s is leader now", identity))
		},
	}

	config2 := &ElectorConfig{
		Client: client,
		ElectionId: "walm-test",
		LockIdentity: "client2",
		LockNamespace: "walmtest",
		OnStartedLeadingFunc: func(stop <-chan struct{}) {
			fmt.Println("client2 is leader")
		},
		OnStoppedLeadingFunc: func() {
			fmt.Println("client2 is not leader any more")
		},
		OnNewLeaderFunc: func(identity string) {
			fmt.Println(fmt.Sprintf("%s is leader now", identity))
		},
	}

	elector1, err := NewElector(config1)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	elector2, err := NewElector(config2)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	go elector1.Run()
	go elector2.Run()

	time.Sleep(5 * time.Second)
}
