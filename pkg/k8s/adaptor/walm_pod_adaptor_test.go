package adaptor

import (
	"testing"
)

func TestPodLogStream(t *testing.T) {
	//client, err := client.CreateFakeApiserverClient("", "C:/kubernetes/0.5/kubeconfig")
	//if err != nil {
	//	println(err.Error())
	//	return
	//}
	//tail := int64(5)
	//readCloser, err := client.CoreV1().Pods("kube-system").GetLogs("walm-bsddx-5548d8fcfc-9wvgs", &corev1.PodLogOptions{
	//	TailLines: &tail,
	//	Follow:true,
	//}).Stream()
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//
	//defer readCloser.Close()
	//io.Copy(os.Stdout, readCloser)
	//fmt.Println("finished")
}