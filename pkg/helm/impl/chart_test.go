package impl

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"WarpCloud/walm/pkg/models/release"
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
