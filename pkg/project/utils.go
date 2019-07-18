package project

import (
	"WarpCloud/walm/pkg/release"
)

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

func buildReleaseRequest(projectInfo *ProjectInfo, releaseName string) *release.ReleaseRequestV2 {
	var releaseRequest *release.ReleaseRequestV2
	for _, releaseInfo := range projectInfo.Releases {
		if releaseInfo.Name == releaseName {
			releaseRequest = releaseInfo.BuildReleaseRequestV2()
			break
		}
	}

	return releaseRequest
}
