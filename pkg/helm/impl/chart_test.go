package impl

import (
	"WarpCloud/walm/pkg/models/release"
	"github.com/stretchr/testify/assert"
	"path"
	"testing"
)

func TestHelm_GetRepoList(t *testing.T) {
	helm := Helm{}
	helm.chartRepoMap = map[string]*ChartRepository{
		"test": {
			Name: "test",
			URL:  "localhost:8880",
		},
	}

	repoList := helm.GetRepoList()
	expectedRepoList := &release.RepoInfoList{
		Items: []*release.RepoInfo{
			{
				TenantRepoName: "test",
				TenantRepoURL:  "localhost:8880",
			},
		},
	}
	assert.Equal(t, expectedRepoList, repoList)
}

func Test_BuildChartInfo(t *testing.T) {
	extension := path.Ext("metainfo/filename.png")
	if len(extension) > 1 {
		assert.Equal(t, extension[1:], "png")
	}
}
