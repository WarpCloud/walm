package helm

import (
	"github.com/hashicorp/golang-lru"
	"k8s.io/helm/pkg/kube"
)

type Client struct {
	kubeClients *lru.Cache
	kubeConfig  string
}

func (c *Client) GetKubeClient(namespace string) *kube.Client {
	if c.kubeClients == nil {
		c.kubeClients, _ = lru.New(100)
	}

	if kubeClient, ok := c.kubeClients.Get(namespace); ok {
		return kubeClient.(*kube.Client)
	} else {
		kubeClient = createKubeClient(c.kubeConfig, namespace)
		c.kubeClients.Add(namespace, kubeClient)
		return kubeClient.(*kube.Client)
	}
}

func createKubeClient(kubeConfig string, namespace string) (*kube.Client) {
	cfg := kube.GetConfig(kubeConfig, "", namespace)
	client := kube.New(cfg)

	return client
}

func NewHelmKubeClient(kubeConfig string) *Client {
	kubeClients, _ := lru.New(100)
	return &Client{
		kubeClients: kubeClients,
		kubeConfig:  kubeConfig,
	}
}
