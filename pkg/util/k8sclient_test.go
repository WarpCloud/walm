package util


import (
	"testing"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test(t *testing.T) {
	client, err := createApiserverClient("http://172.16.1.70:8080", "")
	//client, err := createApiserverClient("", "C:/kubernetes/kubeconfig")
	if err == nil {
		nodelist, err := client.CoreV1().Nodes().List(v1.ListOptions{})
		if err == nil {
			for _, node := range nodelist.Items {
				println(node.Name)
			}
		}
	}

}