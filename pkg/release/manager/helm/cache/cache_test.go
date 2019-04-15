package cache

import (
	"testing"
	"github.com/stretchr/testify/assert"
	hapirelease "k8s.io/helm/pkg/hapi/release"
	"walm/pkg/release"
	"k8s.io/helm/pkg/chart"
	"encoding/json"
	"walm/pkg/release/manager/metainfo"
	//"gopkg.in/yaml.v2"
	"walm/pkg/util/transwarpjsonnet"
	"github.com/ghodss/yaml"
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

//TODO move to e2e test
//func TestHelmCache_Resync(t *testing.T) {
//	helmClient := helm.NewClient(helm.Host("172.26.0.5:31225"))
//	redisClient := redis.CreateFakeRedisClient()
//	kubeClient := client.CreateFakeKubeClient("", "C:/kubernetes/0.5/kubeconfig")
//
//	helmCache := &HelmCache{
//		redisClient: redisClient,
//		helmClient: helmClient,
//		kubeClient: kubeClient,
//	}
//	wg := sync.WaitGroup{}
//	wg.Add(1)
//	go func() {
//		defer wg.Done()
//		helmCache.Resync()
//	}()
//
//	time.Sleep(500 * time.Millisecond)
//	_, err := redisClient.GetClient().HSet(redis.WalmReleasesKey, "test1/test1", "test").Result()
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	wg.Wait()
//	test, err := redisClient.GetClient().HGet(redis.WalmReleasesKey, "test1/test1").Result()
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	if test != "" {
//		t.Fail()
//	}
//
//}
//
//func TestHelmCache_GetCache(t *testing.T) {
//	helmClient := helm.NewClient(helm.Host("172.26.0.5:31225"))
//	redisClient := redis.CreateFakeRedisClient()
//	kubeClient := client.CreateFakeKubeClient("", "C:/kubernetes/0.5/kubeconfig")
//
//	helmCache := &HelmCache{
//		redisClient: redisClient,
//		helmClient: helmClient,
//		kubeClient: kubeClient,
//	}
//
//	wg := sync.WaitGroup{}
//	wg.Add(1)
//	go func() {
//		defer wg.Done()
//		helmCache.Resync()
//	}()
//	wg.Wait()
//
//	releaseCache, err := helmCache.GetReleaseCache("kube-system", "app-manager")
//	if err != nil {
//		fmt.Println(err)
//		t.Fail()
//	}
//	fmt.Printf("%v\n", *releaseCache)
//
//	releaseCaches, err := helmCache.GetReleaseCaches("", "", 0)
//	if err != nil {
//		fmt.Println(err)
//		t.Fail()
//	}
//	fmt.Println(len(releaseCaches))
//
//	releaseCaches, err = helmCache.GetReleaseCaches("kube-system", "", 0)
//	if err != nil {
//		fmt.Println(err)
//		t.Fail()
//	}
//	fmt.Println(len(releaseCaches))
//
//	releaseCaches, err = helmCache.GetReleaseCaches("", "*security", 0)
//	if err != nil {
//		fmt.Println(err)
//		t.Fail()
//	}
//	fmt.Println(len(releaseCaches))
//
//	releaseCaches, err = helmCache.GetReleaseCaches("", "", 5)
//	if err != nil {
//		fmt.Println(err)
//		t.Fail()
//	}
//	fmt.Println(len(releaseCaches))
//
//	//TODO only return 1 but has 5
//	releaseCaches, err = helmCache.GetReleaseCaches("security", "", 5)
//	if err != nil {
//		fmt.Println(err)
//		t.Fail()
//	}
//	fmt.Println(len(releaseCaches))
//}
//
//func TestHelmList_Tmp(t *testing.T) {
//	helmClient := helm.NewClient(helm.Host("172.26.0.5:31225"))
//	resp, err := helmClient.ListReleases(helm.ReleaseListStatuses(
//		[]hapiRelease.Status_Code{hapiRelease.Status_SUPERSEDED}))
//	if err != nil {
//		logrus.Errorf("failed to list helm releases: %s\n", err.Error())
//	}
//	fmt.Println(len(resp.Releases))
//}
