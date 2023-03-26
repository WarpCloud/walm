package main

import (
	"WarpCloud/walm/cmd/walmctl/util/walmctlclient"
	"WarpCloud/walm/pkg/helm/impl"
	"WarpCloud/walm/pkg/models/release"
	"WarpCloud/walm/pkg/setting"
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/go-resty/resty"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"helm.sh/helm/pkg/chartutil"
	"helm.sh/helm/pkg/registry"
	"helm.sh/helm/pkg/repo"
	"io"
	"io/ioutil"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/kubectl/util/i18n"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var longSyncHelp = `
To ensure release sync works well, you need to follow these instructions:
When you run sync command on hostA:

1. save releaseConfigs on local:
-s xxx sync release xxx --save ...
2. run release on hostB
-s xxx sync release xxx --target-server hostB

`

type syncCmd struct {
	name         string
	file         string
	targetServer string
	client       *walmctlclient.WalmctlClient
	out          io.Writer
}

func newSyncCmd(out io.Writer) *cobra.Command {
	sync := &syncCmd{out: out}
	cmd := &cobra.Command{
		Use:                   "sync release",
		DisableFlagsInUseLine: false,
		Short:                 i18n.T("sync instance of your app to other k8s cluster or save into files"),
		Long:                  longSyncHelp,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.Errorf("Arguments invalid, format like `sync release zookeeper-test` instead")
			}
			if args[0] != "release" {
				return errors.Errorf("Unsupported sync type, release only currently")
			}
			if namespace == "" {
				return errNamespaceRequired
			}
			sync.name = args[1]
			return sync.run()
		},
	}
	cmd.PersistentFlags().StringVar(&sync.file, "save", "/tmp/walm-sync", "filepath to save instance of app")
	cmd.PersistentFlags().StringVar(&sync.targetServer, "target-server", "", "walm server address of target cluster")
	return cmd
}

func (sync *syncCmd) run() error {
	client, err := walmctlclient.CreateNewClient(walmserver, enableTLS, rootCA)
	if err != nil {
		klog.Errorf("failed to create walmctl client: %s", err.Error())
		return err
	}
	if err := client.ValidateHostConnect(walmserver); err != nil {
		return err
	}
	sync.client = client

	tmpDir, err := createTempDir()
	if err != nil {
		klog.Errorf("failed to create tmp dir : %s", err.Error())
		return err
	}
	defer os.RemoveAll(tmpDir)

	tmpReleaseRequestPath, tmpChartPath, err := sync.saveRelease(tmpDir)
	if err != nil {
		return err
	}

	pathTokens := strings.SplitAfter(tmpChartPath, "/")
	chartName := pathTokens[len(pathTokens)-1]
	targetDir := filepath.Join(sync.file, namespace+"_"+sync.name)

	if sync.targetServer != "" {
		err = sync.deployRelease(tmpReleaseRequestPath, tmpChartPath)
		if err != nil {
			klog.Errorf("failed to deploy release to target server : %s", err.Error())
			return err
		}
		klog.Infof("Sync release to deploy namespace %s of target server.", namespace)
	} else {
		if _, err = copyFile(tmpReleaseRequestPath, filepath.Join(targetDir, "releaseRequest.yaml")); err != nil {
			return errors.Errorf("failed to copy releaseRequest.yaml to %s: %s", targetDir, err.Error())
		}
		if _, err = copyFile(tmpChartPath, filepath.Join(targetDir, chartName)); err != nil {
			return errors.Errorf("failed to copy %s to %s: %s", chartName, targetDir, err.Error())
		}
		klog.Infof("Sync release to store on %s of local succeed.", targetDir)
	}
	fmt.Println("Succeed to sync release.")
	return nil
}

func (sync *syncCmd) saveRelease(tmpDir string) (string, string, error) {

	var releaseInfo release.ReleaseInfoV2
	resp, err := sync.client.GetRelease(namespace, sync.name)
	if err != nil {
		return "", "", err
	}
	err = json.Unmarshal(resp.Body(), &releaseInfo)
	if err != nil {
		return "", "", err
	}
	releaseRequest := releaseInfo.BuildReleaseRequestV2()
	releaseRequestByte, err := json.Marshal(releaseRequest)
	if err != nil {
		return "", "", err
	}
	tmpReleaseRequestPath := filepath.Join(tmpDir, "releaseRequest.yaml")
	if err := ioutil.WriteFile(tmpReleaseRequestPath, releaseRequestByte, 0644); err != nil {
		return "", "", err
	}
	tmpChartPath, err := saveCharts(sync.client, releaseInfo, tmpDir)
	if err != nil {
		return "", "", err

	}
	return tmpReleaseRequestPath, tmpChartPath, err
}

func (sync *syncCmd) deployRelease(tmpReleaseRequestPath string, tmpChartPath string) error {
	targetClient, err := walmctlclient.CreateNewClient(sync.targetServer, enableTLS, rootCA)
	if err != nil {
		klog.Errorf("failed to create walmctl client: %s", err.Error())
		return err
	}
	if err := targetClient.ValidateHostConnect(walmserver); err != nil {
		return err
	}
	releaseRequestByte, err := ioutil.ReadFile(tmpReleaseRequestPath)
	if err != nil {
		return err
	}
	var configValues map[string]interface{}
	err = json.Unmarshal(releaseRequestByte, &configValues)
	if err != nil {
		return err
	}
	_, err = targetClient.CreateRelease(namespace, tmpChartPath, sync.name, false, 0, configValues)
	if err != nil {
		return err
	}
	return nil
}

func saveCharts(client *walmctlclient.WalmctlClient, releaseInfo release.ReleaseInfoV2, tmpDir string) (string, error) {

	chartImage := releaseInfo.ChartImage
	chartRepo := releaseInfo.RepoName
	chartName := releaseInfo.ChartName
	chartVersion := releaseInfo.ChartVersion
	name := ""
	if chartRepo == "" {
		return "", errors.Errorf("repoName for release is empty, no access to fetch charts")
	}
	registryClient, err := impl.NewRegistryClient(&setting.ChartImageConfig{CacheRootDir: "/chart-cache"})
	if err != nil {
		return "", err
	}
	if chartImage != "" {
		ref, err := registry.ParseReference(chartImage)
		if err != nil {
			klog.Errorf("failed to parse chart image %s : %s", chartImage, err.Error())
			return "", errors.Wrapf(err, "failed to parse chart image %s", chartImage)
		}
		ch, err := registryClient.LoadChart(ref)
		if err != nil {
			klog.Errorf("failed to load chart : %s", err.Error())
			return "", err
		}
		// Save the chart to local destination directory
		dest, err := chartutil.Save(ch, tmpDir)
		if err != nil {
			klog.Errorf("failed to save the chart to local destination directory")
			return "", err
		}
		return dest, nil
	} else {
		resp, err := client.GetRepoList()
		if err != nil {
			return "", err
		}

		var repoUrl string
		repos := gjson.Get(string(resp.Body()), "items").Array()
		for _, repo := range repos {
			if repo.Get("repoName").String() == chartRepo {
				repoUrl = repo.Get("repoUrl").String()
				break
			}
		}
		if repoUrl == "" {
			return "", errors.Errorf("release repo %s not exist in repoList, no access to fetch charts.", chartRepo)
		}
		repoIndex := &repo.IndexFile{}
		chartInfoList := new(release.ChartInfoList)
		chartInfoList.Items = make([]*release.ChartInfo, 0)
		parsedURL, err := url.Parse(repoUrl)
		if err != nil {
			return "", err
		}
		parsedURL.Path = strings.TrimSuffix(parsedURL.Path, "/") + "/index.yaml"

		indexURL := parsedURL.String()

		resp, err = resty.R().Get(indexURL)
		if err != nil {
			klog.Errorf("failed to get index : %s", err.Error())
			return "", err
		}

		if err := yaml.Unmarshal(resp.Body(), repoIndex); err != nil {
			return "", err
		}
		cv, err := repoIndex.Get(chartName, chartVersion)
		if err != nil {
			return "", fmt.Errorf("chart %s-%s is not found: %s", chartName, chartVersion, err.Error())
		}
		if len(cv.URLs) == 0 {
			return "", fmt.Errorf("chart %s has no downloadable URLs", chartName)
		}
		chartUrl := cv.URLs[0]
		absoluteChartURL, err := repo.ResolveReferenceURL(repoUrl, chartUrl)
		if err != nil {
			return "", fmt.Errorf("failed to make absolute chart url: %v", err)
		}
		resp, err = resty.R().Get(absoluteChartURL)
		if err != nil {
			klog.Errorf("failed to get chart : %s", err.Error())
			return "", err
		}
		name = filepath.Base(absoluteChartURL)
		dest := filepath.Join(tmpDir, name)
		if err := ioutil.WriteFile(dest, resp.Body(), 0644); err != nil {
			klog.Errorf("failed to write chart : %s", err.Error())
			return "", err
		}
		return dest, nil
	}
}
