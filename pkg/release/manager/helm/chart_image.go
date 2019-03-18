package helm

import (
	"k8s.io/helm/pkg/registry"
	"k8s.io/helm/pkg/chart"
	"github.com/sirupsen/logrus"
	"net/http"
	"crypto/tls"
	"github.com/containerd/containerd/remotes/docker"
	"os"
	"walm/pkg/setting"
)

type ChartImageClient struct {
	registryClient *registry.Client
}

var chartImageClient *ChartImageClient

func GetDefaultChartImageClient() *ChartImageClient {
	if chartImageClient == nil {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}

		option := &registry.ClientOptions{
			Out: os.Stdout,
			Resolver: registry.Resolver{
				Resolver: docker.NewResolver(docker.ResolverOptions{
					Client: client,
				}),
			},
			CacheRootDir: "/helm-cache",
		}
		if setting.Config.ChartImageConfig != nil {
			option.CacheRootDir = setting.Config.ChartImageConfig.CacheRootDir
		}
		registryClient := registry.NewClient(option)

		chartImageClient = &ChartImageClient{
			registryClient: registryClient,
		}
	}
	return chartImageClient
}

func (c *ChartImageClient) GetChart(chartImage string) (*chart.Chart, error) {
	ref, err := registry.ParseReference(chartImage)
	if err != nil {
		logrus.Errorf("failed to parse chart image %s : %s",chartImage, err.Error())
		return nil, err
	}

	err = c.registryClient.PullChart(ref)
	if err != nil {
		logrus.Errorf("failed to pull chart %s : %s", chartImage, err.Error())
		return nil, err
	}

	chart, err := c.registryClient.LoadChart(ref)
	if err != nil {
		logrus.Errorf("failed to load chart %s : %s", chartImage, err.Error())
		return nil, err
	}
	return chart, nil
}
