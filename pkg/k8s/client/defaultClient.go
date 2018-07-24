package client

import (
	"walm/pkg/setting"

	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	. "walm/pkg/util/log"
)

var DefaultApiserverClient *kubernetes.Clientset
var DefaultRestConfig *restclient.Config

func GetDefaultClient() *kubernetes.Clientset {
	var err error
	if DefaultApiserverClient == nil {
		DefaultApiserverClient, err = CreateApiserverClient(setting.Config.Kube.MasterHost, setting.Config.Kube.Config)
	}
	if err != nil {
		Log.Fatalf("create apiserver client failed:%v", err)
	}
	return DefaultApiserverClient
}

func GetDefaultRestConfig() *restclient.Config {
	var err error
	if DefaultRestConfig == nil {
		DefaultRestConfig, err = clientcmd.BuildConfigFromFlags(setting.Config.Kube.MasterHost, setting.Config.Kube.Config)
	}
	if err != nil {
		Log.Fatalf("get default rest config= failed:%v", err)
	}
	return DefaultRestConfig
}
