package client

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

func TestEx(t *testing.T) {
	clientEx, err := createApiserverClientEx("", "C:/kubernetes/0.5/kubeconfig")
	//client, err := createApiserverClient("", "C:/kubernetes/kubeconfig")
	if err == nil {
		inst, err := clientEx.TranswarpV1beta1().ApplicationInstances("hnnxst1").Get("guardian-hnnxst1", v1.GetOptions{})
		if err == nil {
			println(inst.Name)
		}
	}

}
