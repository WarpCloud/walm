package framework

import (
	"strings"
	"walm/pkg/k8s/handler"
	"fmt"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	"walm/pkg/release/manager/helm"
	"os"
	"walm/pkg/util/transwarpjsonnet"
	"k8s.io/helm/pkg/chart/loader"
	"errors"
	"runtime"
)

const (
	maxNameLength                = 62
	randomLength                 = 5
	maxGeneratedRandomNameLength = maxNameLength - randomLength
)

func GenerateRandomName(base string) string {
	if len(base) > maxGeneratedRandomNameLength {
		base = base[:maxGeneratedRandomNameLength]
	}
	return fmt.Sprintf("%s-%s", strings.ToLower(base), utilrand.String(randomLength))
}

func CreateRandomNamespace(base string) (string, error) {
	namespace := GenerateRandomName(base)
	ns := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := handler.GetDefaultHandlerSet().GetNamespaceHandler().CreateNamespace(&ns)
	return namespace, err
}

func DeleteNamespace(namespace string, deleteReleases bool) (error) {
	if deleteReleases {
		releases, err := helm.GetDefaultHelmClient().ListReleases(namespace, "")
		if err != nil {
			return err
		}
		for _, release := range releases {
			err := helm.GetDefaultHelmClient().DeleteRelease(namespace, release.Name, false, false, false, 0)
			if err !=  nil {
				return err
			}
		}
	}

	return handler.GetDefaultHandlerSet().GetNamespaceHandler().DeleteNamespace(namespace)
}

func LoadChartArchive(name string) ([]*loader.BufferedFile, error) {
	if fi, err := os.Stat(name); err != nil {
		return nil, err
	} else if fi.IsDir() {
		return nil, errors.New("cannot load a directory")
	}

	raw, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer raw.Close()
	return transwarpjsonnet.LoadArchive(raw)
}

func GetCurrentFilePath() (string, error) {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		return "", errors.New("Can not get current file info")
	}
	return file, nil
}