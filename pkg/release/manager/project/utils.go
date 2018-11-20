package project

import (
	"walm/pkg/release"
	"time"
	"fmt"
)

func buildProjectReleaseName(projectName, releaseName string) string {
	return fmt.Sprintf("%s--%s", projectName, releaseName)
}

func mergeValues(dest map[string]interface{}, src map[string]interface{}) map[string]interface{} {
	for k, v := range src {
		// If the key doesn't exist already, then just set the key to that value
		if _, exists := dest[k]; !exists {
			dest[k] = v
			continue
		}
		nextMap, ok := v.(map[string]interface{})
		// If it isn't another map, overwrite the value
		if !ok {
			dest[k] = v
			continue
		}
		// Edge case: If the key exists in the destination, but isn't a map
		destMap, isMap := dest[k].(map[string]interface{})
		// If the source map has a map for this key, prefer it
		if !isMap {
			dest[k] = v
			continue
		}
		// If we got to this point, it is a map in both, so merge them
		dest[k] = mergeValues(destMap, nextMap)
	}

	return dest
}

func buildReleaseRequest(projectInfo *release.ProjectInfo, releaseName string) *release.ReleaseRequest {
	var releaseRequest release.ReleaseRequest
	found := false
	for _, releaseInfo := range projectInfo.Releases {
		if releaseInfo.Name != releaseName {
			continue
		}
		releaseRequest.ConfigValues = make(map[string]interface{})
		releaseRequest.ConfigValues["UPDATE"] = time.Now().String()
		releaseRequest.Dependencies = make(map[string]string)
		for k, v := range releaseInfo.Dependencies {
			releaseRequest.Dependencies[k] = v
		}
		releaseRequest.Name = buildProjectReleaseName(projectInfo.Name, releaseInfo.Name)
		releaseRequest.ChartName = releaseInfo.ChartName
		releaseRequest.ChartVersion = releaseInfo.ChartVersion
		found = true
		break
	}

	if !found {
		return nil
	}
	return &releaseRequest
}
