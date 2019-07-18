package cache

import (
	"testing"
	"github.com/stretchr/testify/assert"
	hapirelease "k8s.io/helm/pkg/hapi/release"
	"WarpCloud/walm/pkg/release"
	"k8s.io/helm/pkg/chart"
	"encoding/json"
	"WarpCloud/walm/pkg/release/manager/metainfo"
	//"gopkg.in/yaml.v2"
	"WarpCloud/walm/pkg/util/transwarpjsonnet"
	"github.com/ghodss/yaml"
	"transwarp/release-config/pkg/apis/transwarp/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBuildHScanFilter(t *testing.T) {
	tests := []struct {
		namespace string
		filter    string
		result    string
	}{
		{
			namespace: "",
			filter:    "",
			result:    "*/*",
		},
		{
			namespace: "test_ns",
			filter:    "",
			result:    "test_ns/*",
		},
		{
			namespace: "",
			filter:    "test_name",
			result:    "*/test_name",
		},
		{
			namespace: "test_ns",
			filter:    "test_name",
			result:    "test_ns/test_name",
		},
	}

	for _, test := range tests {
		result := buildHScanFilter(test.namespace, test.filter)
		assert.Equal(t, test.result, result)
	}
}

func TestBuildHelmReleasesMap(t *testing.T) {
	tests := []struct {
		helmReleases []*hapirelease.Release
		result       map[string]*hapirelease.Release
	}{
		{
			helmReleases: []*hapirelease.Release{
				{
					Name:      "rel1",
					Namespace: "default",
					Version:   2,
				},
				{
					Name:      "rel1",
					Namespace: "default",
					Version:   1,
				},
				{
					Name:      "rel2",
					Namespace: "default",
					Version:   1,
				},
			},
			result: map[string]*hapirelease.Release{
				"default/rel1": {
					Name:      "rel1",
					Namespace: "default",
					Version:   2,
				},
				"default/rel2": {
					Name:      "rel2",
					Namespace: "default",
					Version:   1,
				},
			},
		},
		{
			helmReleases: []*hapirelease.Release{
				{
					Name:      "rel1",
					Namespace: "default",
					Version:   1,
				},
				{
					Name:      "rel1",
					Namespace: "default",
					Version:   2,
				},
				{
					Name:      "rel2",
					Namespace: "default",
					Version:   1,
				},
			},
			result: map[string]*hapirelease.Release{
				"default/rel1": {
					Name:      "rel1",
					Namespace: "default",
					Version:   2,
				},
				"default/rel2": {
					Name:      "rel2",
					Namespace: "default",
					Version:   1,
				},
			},
		},
	}

	for _, test := range tests {
		result := buildHelmReleasesMap(test.helmReleases)
		assert.Equal(t, test.result, result)
	}
}

func TestHelmCache_BuildHelmCaches(t *testing.T) {
	nameNodeRoleImage := "test_value"
	tests := []struct {
		releases map[string]*hapirelease.Release
		result   map[string]*release.ReleaseCache
		err      error
	}{
		{
			releases: map[string]*hapirelease.Release{
				"default/test_name": {
					Namespace: "default",
					Name:      "test_name",
					Version:   1,
					Config: map[string]interface{}{
						"test_key": "test_value",
					},
					Chart: &chart.Chart{
						Metadata: &chart.Metadata{
							Name:       "test_chart",
							Version:    "1.0",
							AppVersion: "2.0",
						},
						Values: map[string]interface{}{
							"test_key":  "test_value1",
							"test_key2": "test_value2",
						},
						Files: []*chart.File{
							{
								Name: transwarpjsonnet.TranswarpMetadataDir + transwarpjsonnet.TranswarpMetaInfoFileName,
								Data: convertMetaInfoToBytes(&metainfo.ChartMetaInfo{
									FriendlyName: "hdfs",
									ChartRoles: []*metainfo.MetaRoleConfig{
										{
											Name: "namenode",
											RoleBaseConfig: &metainfo.MetaRoleBaseConfig{
												Image: &metainfo.MetaStringConfig{
													MapKey: "test_key",
												},
											},
										},
									},
								}),
							},
						},
					},
				},
			},
			result: map[string]*release.ReleaseCache{
				"default/test_name": {
					ReleaseSpec: release.ReleaseSpec{
						Namespace:       "default",
						Name:            "test_name",
						Version:         int32(1),
						ChartName:       "test_chart",
						ChartVersion:    "1.0",
						ChartAppVersion: "2.0",
						Dependencies:    map[string]string{},
						ConfigValues: map[string]interface{}{
							"test_key": "test_value",
						},
					},
					ComputedValues: map[string]interface{}{
						"test_key":  "test_value",
						"test_key2": "test_value2",
					},
					MetaInfoValues: &metainfo.MetaInfoParams{
						Roles: []*metainfo.MetaRoleConfigValue{
							{
								Name: "namenode",
								RoleBaseConfigValue: &metainfo.MetaRoleBaseConfigValue{
									Image: &nameNodeRoleImage,
								},
							},
						},
					},
				},
			},
			err: nil,
		},
	}

	helmCache := &HelmCache{
		getReleaseResourceMetas: func(helmRelease *hapirelease.Release) (resources []release.ReleaseResourceMeta, err error) {
			return
		},
	}

	for _, test := range tests {
		rcs, err := helmCache.buildReleaseCaches(test.releases)
		if assert.IsType(t, test.err, err) {
			result := convertBytesToReleaseCacheMap(rcs)
			assert.Equal(t, test.result, result)
		}
	}
}

func convertBytesToReleaseCacheMap(rcs map[string]interface{}) map[string]*release.ReleaseCache {
	result := map[string]*release.ReleaseCache{}
	for key, value := range rcs {
		rc := &release.ReleaseCache{}
		_ = json.Unmarshal(value.([]byte), rc)
		result[key] = rc
	}
	return result
}

func convertMetaInfoToBytes(metaInfo *metainfo.ChartMetaInfo) []byte {
	bytes, _ := yaml.Marshal(metaInfo)
	return bytes
}

func Test_BuildReleaseCacheKeysToDel(t *testing.T) {
	tests := []struct {
		releaseCacheKeysFromRedis []string
		releaseCachesFromHelm     map[string]interface{}
		result                    []string
	}{
		{
			releaseCacheKeysFromRedis: []string{
				"test_name1",
				"test_name2",
			},
			releaseCachesFromHelm: map[string]interface{}{
				"test_name2": nil,
			},
			result: []string{
				"test_name1",
			},
		},
	}

	for _, test := range tests {
		result := buildReleaseCacheKeysToDel(test.releaseCacheKeysFromRedis, test.releaseCachesFromHelm)
		assert.ElementsMatch(t, test.result, result)
	}
}

func Test_BuildProjectCachesFromReleaseConfigs(t *testing.T) {
	tests := []struct {
		releaseConfigs []*v1beta1.ReleaseConfig
		result         map[string]string
		err            error
	}{
		{
			releaseConfigs: []*v1beta1.ReleaseConfig{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Labels: map[string]string{
							ProjectNameLabelKey: "project1",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Labels: map[string]string{
							ProjectNameLabelKey: "project2",
						},
					},
				},
			},
			result: map[string]string{
				"default/project1": "{\"name\":\"project1\",\"namespace\":\"default\",\"latestTaskSignature\":null,\"latestTaskTimeoutSec\":0}",
				"default/project2": "{\"name\":\"project2\",\"namespace\":\"default\",\"latestTaskSignature\":null,\"latestTaskTimeoutSec\":0}",
			},
			err: nil,
		},
	}

	for _, test := range tests {
		result, err := buildProjectCachesFromReleaseConfigs(test.releaseConfigs)
		if assert.IsType(t, test.err, err) {
			assert.Equal(t, test.result, result)
		}
	}
}

func Test_BuildReleaseTasksFromHelm(t *testing.T) {
	tests := []struct {
		releaseCachesFromHelm map[string]interface{}
		result                map[string]string
		err                   error
	}{
		{
			releaseCachesFromHelm: map[string]interface{}{
				"default/rel1": convertReleaseCacheToBytes(release.ReleaseCache{
					ReleaseSpec: release.ReleaseSpec{
						Namespace: "default",
						Name:      "rel1",
					},
				}),
				"default/rel2": convertReleaseCacheToBytes(release.ReleaseCache{
					ReleaseSpec: release.ReleaseSpec{
						Namespace: "default",
						Name:      "rel2",
					},
				}),
			},
			result: map[string]string{
				"default/rel1": "{\"name\":\"rel1\",\"namespace\":\"default\",\"latestReleaseTaskSignature\":null}",
				"default/rel2": "{\"name\":\"rel2\",\"namespace\":\"default\",\"latestReleaseTaskSignature\":null}",
			},
			err: nil,
		},
	}

	for _, test := range tests {
		result, err := buildReleaseTasksFromHelm(test.releaseCachesFromHelm)
		if assert.IsType(t, test.err, err) {
			assert.Equal(t, test.result, result)
		}
	}
}

func convertReleaseCacheToBytes(releaseCache release.ReleaseCache) []byte {
	bytes, _ := json.Marshal(releaseCache)
	return bytes
}

func TestHelmCache_BuildProjectCachesToDel(t *testing.T) {
	tests := []struct {
		helmCache                       *HelmCache
		projectCachesFromReleaseConfigs map[string]string
		projectCacheInRedis             map[string]string
		result                          []string
		err                             error
	}{
		{
			helmCache: &HelmCache{},
			projectCachesFromReleaseConfigs: map[string]string{
				"default/project1": "{\"name\":\"project1\",\"namespace\":\"default\",\"latestTaskSignature\":null,\"latestTaskTimeoutSec\":0}",
			},
			projectCacheInRedis: map[string]string{
				"default/project1": "{\"name\":\"project1\",\"namespace\":\"default\",\"latestTaskSignature\":null,\"latestTaskTimeoutSec\":0}",
			},
			result: []string{},
			err:    nil,
		},
		{
			helmCache: &HelmCache{
				isProjectTaskFinishedOrTimeout: func(projectCache *ProjectCache) bool {
					return true
				},
			},
			projectCachesFromReleaseConfigs: map[string]string{
			},
			projectCacheInRedis: map[string]string{
				"default/project1": "{\"name\":\"project1\",\"namespace\":\"default\",\"latestTaskSignature\":null,\"latestTaskTimeoutSec\":0}",
			},
			result: []string{"default/project1"},
			err:    nil,
		},
		{
			helmCache: &HelmCache{
				isProjectTaskFinishedOrTimeout: func(projectCache *ProjectCache) bool {
					return false
				},
			},
			projectCachesFromReleaseConfigs: map[string]string{
			},
			projectCacheInRedis: map[string]string{
				"default/project1": "{\"name\":\"project1\",\"namespace\":\"default\",\"latestTaskSignature\":null,\"latestTaskTimeoutSec\":0}",
			},
			result: []string{},
			err:    nil,
		},
	}

	for _, test := range tests {
		result, err := test.helmCache.buildProjectCachesToDel(test.projectCachesFromReleaseConfigs, test.projectCacheInRedis)
		if assert.IsType(t, test.err, err) {
			assert.Equal(t, test.result, result)
		}
	}
}

func Test_buildProjectCachesToSet(t *testing.T) {
	tests := []struct {
		projectCachesFromReleaseConfigs map[string]string
		projectCacheInRedis             map[string]string
		result                          map[string]interface{}
	}{
		{
			projectCachesFromReleaseConfigs: map[string]string{
				"default/project1": "{\"name\":\"project1\",\"namespace\":\"default\",\"latestTaskSignature\":null,\"latestTaskTimeoutSec\":0}",
			},
			projectCacheInRedis: map[string]string{
				"default/project1": "{\"name\":\"project1\",\"namespace\":\"default\",\"latestTaskSignature\":null,\"latestTaskTimeoutSec\":0}",
			},
			result: map[string]interface{}{
			},
		},
		{
			projectCachesFromReleaseConfigs: map[string]string{
				"default/project1": "{\"name\":\"project1\",\"namespace\":\"default\",\"latestTaskSignature\":null,\"latestTaskTimeoutSec\":0}",
			},
			projectCacheInRedis: map[string]string{
			},
			result: map[string]interface{}{
				"default/project1": "{\"name\":\"project1\",\"namespace\":\"default\",\"latestTaskSignature\":null,\"latestTaskTimeoutSec\":0}",
			},
		},
	}

	for _, test := range tests {
		result := buildProjectCachesToSet(test.projectCachesFromReleaseConfigs, test.projectCacheInRedis)
		assert.Equal(t, test.result, result)
	}
}

func TestHelmCache_BuildReleaseTasksToDel(t *testing.T) {
	tests := []struct {
		helmCache            *HelmCache
		releaseTasksFromHelm map[string]string
		releaseTaskInRedis   map[string]string
		result               []string
		err                  error
	}{
		{
			helmCache: &HelmCache{},
			releaseTasksFromHelm: map[string]string{
				"default/rel1": "{\"name\":\"rel1\",\"namespace\":\"default\",\"latestReleaseTaskSignature\":null}",
			},
			releaseTaskInRedis: map[string]string{
				"default/rel1": "{\"name\":\"rel1\",\"namespace\":\"default\",\"latestReleaseTaskSignature\":null}",
			},
			result: []string{},
			err:    nil,
		},
		{
			helmCache: &HelmCache{},
			releaseTasksFromHelm: map[string]string{
				"default/rel1": "{\"name\":\"rel1\",\"namespace\":\"default\",\"latestReleaseTaskSignature\":null}",
			},
			releaseTaskInRedis: map[string]string{
				"default/rel1": "{\"name\":\"rel1\",\"namespace\":\"default\",\"latestReleaseTaskSignature\":null}",
			},
			result: []string{},
			err:    nil,
		},
		{
			helmCache: &HelmCache{isReleaseTaskFinishedOrTimeout: func(releaseTask *ReleaseTask) bool {
				return true
			}},
			releaseTasksFromHelm: map[string]string{
			},
			releaseTaskInRedis: map[string]string{
				"default/rel1": "{\"name\":\"rel1\",\"namespace\":\"default\",\"latestReleaseTaskSignature\":null}",
			},
			result: []string{"default/rel1"},
			err:    nil,
		},
		{
			helmCache: &HelmCache{isReleaseTaskFinishedOrTimeout: func(releaseTask *ReleaseTask) bool {
				return false
			}},
			releaseTasksFromHelm: map[string]string{
			},
			releaseTaskInRedis: map[string]string{
				"default/rel1": "{\"name\":\"rel1\",\"namespace\":\"default\",\"latestReleaseTaskSignature\":null}",
			},
			result: []string{},
			err:    nil,
		},
	}

	for _, test := range tests {
		result, err := test.helmCache.buildReleaseTasksToDel(test.releaseTasksFromHelm, test.releaseTaskInRedis)
		if assert.IsType(t, test.err, err) {
			assert.Equal(t, test.result, result)
		}
	}
}

func Test_BuildReleaseTasksToSet(t *testing.T) {
	tests := []struct {
		releaseTasksFromHelm map[string]string
		releaseTaskInRedis   map[string]string
		result               map[string]interface{}
	}{
		{
			releaseTasksFromHelm: map[string]string{
				"default/rel1": "{\"name\":\"rel1\",\"namespace\":\"default\",\"latestReleaseTaskSignature\":null}",
			},
			releaseTaskInRedis: map[string]string{
				"default/rel1": "{\"name\":\"rel1\",\"namespace\":\"default\",\"latestReleaseTaskSignature\":null}",
			},
			result: map[string]interface{}{
			},
		},
		{
			releaseTasksFromHelm: map[string]string{
				"default/rel1": "{\"name\":\"rel1\",\"namespace\":\"default\",\"latestReleaseTaskSignature\":null}",
			},
			releaseTaskInRedis: map[string]string{
			},
			result: map[string]interface{}{
				"default/rel1": "{\"name\":\"rel1\",\"namespace\":\"default\",\"latestReleaseTaskSignature\":null}",
			},
		},
	}

	for _, test := range tests {
		result := buildReleaseTasksToSet(test.releaseTasksFromHelm, test.releaseTaskInRedis)
		assert.Equal(t, test.result, result)
	}
}

