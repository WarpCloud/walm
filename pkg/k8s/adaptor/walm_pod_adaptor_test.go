package adaptor

import (
	"testing"
	"walm/pkg/k8s/client"
	"fmt"
	corev1 "k8s.io/api/core/v1"
)

func TestPodAdaptor(t *testing.T) {
	client, err := client.CreateFakeApiserverClient("", "C:/kubernetes/0.5/kubeconfig")
	if err != nil {
		println(err.Error())
		return
	}
	tail := int64(5)
	logs, err := client.CoreV1().Pods("tosshengfen").GetLogs("txsql-48d9q-2", &corev1.PodLogOptions{TailLines: &tail}).Do().Raw()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(logs))
}