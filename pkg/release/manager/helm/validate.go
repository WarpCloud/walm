package helm
//
//import (
//	"fmt"
//	"strings"
//	"bytes"
//
//	"WarpCloud/walm/pkg/k8s/client"
//	"WarpCloud/walm/pkg/release"
//
//	"github.com/sirupsen/logrus"
//	"github.com/ghodss/yaml"
//	"k8s.io/helm/pkg/chartutil"
//	"k8s.io/helm/pkg/engine"
//)
//
//func (client *HelmClient)ValidateChart(namespace string, releaseRequest release.ReleaseRequest) (release.ChartValicationInfo, error) {
//
//	logrus.Debugf("Begin ValidateChart %v\n", releaseRequest)
//
//	var chartValicationInfo release.ChartValicationInfo
//	chartValicationInfo.ChartName = releaseRequest.ChartName
//	chartValicationInfo.Name = releaseRequest.Name
//	chartValicationInfo.ConfigValues = releaseRequest.ConfigValues
//	chartValicationInfo.ChartVersion = releaseRequest.ChartVersion
//	chartValicationInfo.Dependencies = releaseRequest.Dependencies
//	chartValicationInfo.Namespace = namespace
//
//
//	chartPath, err := client.downloadChart(releaseRequest.RepoName, releaseRequest.ChartName, releaseRequest.ChartVersion)
//	if err != nil {
//		return chartValicationInfo, err
//	}
//	chartRequested, err := chartutil.Load(chartPath)
//	if err != nil {
//		return chartValicationInfo, err
//	}
//
//	if namespace == "" {
//		namespace = "default"
//	}
//
//	rawVals, err := yaml.Marshal(releaseRequest.ConfigValues)
//	if err != nil {
//		return chartValicationInfo, err
//	}
//	config := &chart.Config{Raw: string(rawVals), Values: map[string]*chart.Value{}}
//
//	var links []string
//	for k, v := range releaseRequest.Dependencies {
//		tmpStr := k + "=" + v
//		links = append(links, tmpStr)
//	}
//
//	out := make(map[string]string)
//	if chartRequested.Metadata.Engine == "jsonnet" {
//
//		if len(links) > 0 {
//
//			out, err = renderWithDependencies(chartRequested, namespace, rawVals, "1.9", "", links)
//
//		} else {
//
//			out, err = render(chartRequested, namespace, rawVals, "1.9")
//		}
//
//	} else {
//
//		if req, err := chartutil.LoadRequirements(chartRequested); err == nil {
//
//			if err := checkDependencies(chartRequested, req); err != nil {
//				return chartValicationInfo, fmt.Errorf("checkDependencies: %v", err)
//			}
//
//		} else if err != chartutil.ErrRequirementsNotFound {
//			return  chartValicationInfo, fmt.Errorf("checkDependencies: %v", err)
//		}
//
//		options := chartutil.ReleaseOptions{
//			Name:      releaseRequest.Name,
//			IsInstall: false,
//			IsUpgrade: false,
//			Time:      timeconv.Now(),
//			Namespace: namespace,
//		}
//
//		err = chartutil.ProcessRequirementsEnabled(chartRequested, config)
//		if err != nil {
//			return chartValicationInfo, err
//		}
//		err = chartutil.ProcessRequirementsImportValues(chartRequested)
//		if err != nil {
//			return chartValicationInfo, err
//		}
//
//		// Set up engine.
//		renderer := engine.New()
//
//		caps := &chartutil.Capabilities{
//			APIVersions:   chartutil.DefaultVersionSet,
//			KubeVersion:   chartutil.DefaultKubeVersion,
//		}
//
//		vals, err := chartutil.ToRenderValuesCaps(chartRequested, config, options, caps)
//		if err != nil {
//			return chartValicationInfo, err
//		}
//
//		out, err = renderer.Render(chartRequested, vals)
//		if err != nil {
//			return chartValicationInfo, err
//		}
//
//	}
//
//	if err != nil {
//		return chartValicationInfo, err
//	}
//
//	chartValicationInfo.RenderStatus = "ok"
//	chartValicationInfo.RenderResult = out
//
//	resultMap, errFlag := dryRunK8sResource(out, namespace)
//	if errFlag {
//		chartValicationInfo.DryRunStatus = "failed"
//		chartValicationInfo.ErrorMessage = "dry run check fail"
//	}else {
//		chartValicationInfo.DryRunStatus = "ok"
//		chartValicationInfo.ErrorMessage = " test pass "
//	}
//
//	chartValicationInfo.DryRunResult = resultMap
//	return chartValicationInfo, nil
//
//}
//
//func dryRunK8sResource(out map[string]string, namespace string) (map[string]string, bool) {
//	resultMap := make(map[string]string)
//	errFlag := false
//	kubeClient := client.GetKubeClient()
//	for name, content := range out {
//
//		if strings.HasSuffix(name, "NOTES.txt") {
//			continue
//		}
//
//		r := bytes.NewReader([]byte(content))
//		_, err := kubeClient.BuildUnstructured(namespace, r)
//		if err != nil {
//			resultMap[name] = err.Error()
//			errFlag = true
//		}else {
//			resultMap[name] = " dry run suc"
//		}
//
//	}
//
//	return  resultMap, errFlag
//}
//
//func renderWithDependencies(chartRequested *chart.Chart, namespace string, userVals []byte, kubeVersion string, kubeContext string, links []string) (map[string]string, error) {
//	//
//	//depLinks := map[string]interface{}{}
//	//for _, value := range links {
//	//	if err := strvals.ParseInto(value, depLinks); err != nil {
//	//		return nil, fmt.Errorf("failed parsing --set data: %s", err)
//	//	}
//	//}
//	//
//	//err := transwarp.CheckDepencies(chartRequested, depLinks)
//	//if err != nil {
//	//	return nil, err
//	//}
//	//
//	//// init k8s transwarp client
//	//k8sTranswarpClient := client.GetDefaultClientEx()
//	//
//	//// init k8s client
//	//k8sClient := client.GetDefaultClient()
//	//
//	//depVals, err := transwarp.GetDepenciesConfig(k8sTranswarpClient, k8sClient, namespace, depLinks)
//	//if err != nil {
//	//	return nil, err
//	//}
//	//
//	//newVals, err := transwarp.MergeDepenciesValue(depVals, userVals)
//	//if err != nil {
//	//	return nil, err
//	//}
//	//
//	//return transwarp.Render(chartRequested, namespace, newVals, kubeVersion)
//	return nil, nil
//}
//
//
//func render(chartRequested *chart.Chart, namespace string, userVals []byte, kubeVersion string) (map[string]string, error) {
//	return nil, nil
//}
//
//
//func checkDependencies(ch *chart.Chart, reqs *chartutil.Requirements) error {
//	missing := []string{}
//
//	deps := ch.GetAutoDependencies()
//	for _, r := range reqs.Dependencies {
//		found := false
//		for _, d := range deps {
//			if d.Metadata.Name == r.Name {
//				found = true
//				break
//			}
//		}
//		if !found {
//			missing = append(missing, r.Name)
//		}
//	}
//
//	if len(missing) > 0 {
//		return fmt.Errorf("found in requirements.yaml, but missing in charts/ directory: %s", strings.Join(missing, ", "))
//	}
//	return nil
//}
