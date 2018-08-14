package helm

import (
	"io/ioutil"
	"fmt"
	"os"
	"net/url"
	"strings"

	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/repo"
	"github.com/ghodss/yaml"
	"path/filepath"
)

func FindChartInChartMuseumRepoURL(repoURL, username, password, chartName, chartVersion string) (string, getter.Getter, error) {
	tempIndexFile, err := ioutil.TempFile("", "tmp-repo-file")
	if err != nil {
		return "", nil, fmt.Errorf("cannot write index file for repository requested")
	}
	defer os.Remove(tempIndexFile.Name())

	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return "", nil, err
	}
	parsedURL.Path = strings.TrimSuffix(parsedURL.Path, "/") + "/index.yaml"

	indexURL := parsedURL.String()
	httpGetter, err := getter.NewHTTPGetter(repoURL, "", "", "")
	httpGetter.SetCredentials(username, password)
	resp, err := httpGetter.Get(indexURL)
	if err != nil {
		return "", nil, err
	}
	index, err := ioutil.ReadAll(resp)
	if err != nil {
		return "", nil, err
	}
	repoIndex := &repo.IndexFile{}
	if err := yaml.Unmarshal(index, repoIndex); err != nil {
		return "", nil, err
	}
	errMsg := fmt.Sprintf("chart %q", chartName)
	if chartVersion != "" {
		errMsg = fmt.Sprintf("%s version %q", errMsg, chartVersion)
	}
	cv, err := repoIndex.Get(chartName, chartVersion)
	if err != nil {
		return "", nil, fmt.Errorf("%s not found in %s repository", errMsg, repoURL)
	}
	if len(cv.URLs) == 0 {
		return "", nil, fmt.Errorf("%s has no downloadable URLs", errMsg)
	}
	chartURL := cv.URLs[0]

	absoluteChartURL, err := repo.ResolveReferenceURL(repoURL, chartURL)
	if err != nil {
		return "", nil, fmt.Errorf("failed to make chart URL absolute: %v", err)
	}

	return absoluteChartURL, httpGetter, nil
}

func ChartMuseumDownloadTo(ref, dest string, getter getter.Getter) (string, error) {
	data, err := getter.Get(ref)
	if err != nil {
		return "", err
	}

	name := filepath.Base(ref)
	destfile := filepath.Join(dest, name)
	if err := ioutil.WriteFile(destfile, data.Bytes(), 0644); err != nil {
		return destfile, err
	}

	return destfile, nil
}
