package adaptor

//import (
//	"testing"
//	"walm/pkg/k8s/client"
//	"fmt"
//	"encoding/json"
//	"k8s.io/apimachinery/pkg/apis/meta/v1"
//)
//
//func TestJobAdaptor(t *testing.T) {
//	client, err := client.CreateApiserverClient("", "C:/kubernetes/0.5/kubeconfig")
//	if err != nil {
//		println(err.Error())
//		return
//	}
//
//	watch, err := client.BatchV1().Jobs("default").Watch(v1.ListOptions{})
//	if err != nil {
//		println(err.Error())
//		return
//	}
//	for {
//		select {
//		case event, ok := <-watch.ResultChan():
//			if !ok {
//				break
//			}
//			e, err := json.Marshal(event.Type)
//			if err != nil {
//				fmt.Println(err)
//				return
//			}
//			fmt.Println(string(e))
//			e, err = json.Marshal(event.Object)
//			if err != nil {
//				fmt.Println(err)
//				return
//			}
//			fmt.Println(string(e))
//		}
//	}
//
//}