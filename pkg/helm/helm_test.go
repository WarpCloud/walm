package helm

import (
	"testing"
	"WarpCloud/walm/pkg/models/release"
	"github.com/stretchr/testify/assert"
	"k8s.io/helm/pkg/walm"
	"k8s.io/helm/pkg/walm/plugins"
)

func Test_BuildReleasePluginsByConfigValues(t *testing.T) {
	tests := []struct {
		configValues          map[string]interface{}
		releasePlugins        []*release.ReleasePlugin
		hasPauseReleasePlugin bool
		err                   error
	}{
		{
			configValues: map[string]interface{}{
				walm.WalmPluginConfigKey: []interface{}{
					map[string]interface{}{
						"name": "test-plugin",
						"args": "test-args",
						"version": "test-version",
						"disable": false,
					},
				},
			},
			releasePlugins: []*release.ReleasePlugin{
				{
					Name: "test-plugin",
					Args: "test-args",
					Version: "test-version",
					Disable: false,
				},
			},
			hasPauseReleasePlugin: false,
			err: nil,
		},
		{
			configValues: map[string]interface{}{
				walm.WalmPluginConfigKey: []interface{}{
					map[string]interface{}{
						"name": "test-plugin",
						"args": "test-args",
						"version": "test-version",
						"disable": true,
					},
				},
			},
			releasePlugins: []*release.ReleasePlugin{
			},
			hasPauseReleasePlugin: false,
			err: nil,
		},
		{
			configValues: map[string]interface{}{
				walm.WalmPluginConfigKey: []interface{}{
					map[string]interface{}{
						"name": plugins.ValidateReleaseConfigPluginName,
						"args": "test-args",
						"version": "test-version",
						"disable": false,
					},
				},
			},
			releasePlugins: []*release.ReleasePlugin{
			},
			hasPauseReleasePlugin: false,
			err: nil,
		},
		{
			configValues: map[string]interface{}{
				walm.WalmPluginConfigKey: []interface{}{
					map[string]interface{}{
						"name": "test-plugin",
						"args": "test-args",
						"version": "test-version",
						"disable": false,
					},
					map[string]interface{}{
						"name": plugins.PauseReleasePluginName,
						"args": "",
						"version": "",
						"disable": false,
					},
				},
			},
			releasePlugins: []*release.ReleasePlugin{
				{
					Name: "test-plugin",
					Args: "test-args",
					Version: "test-version",
					Disable: false,
				},
				{
					Name: plugins.PauseReleasePluginName,
					Disable: false,
				},
			},
			hasPauseReleasePlugin: true,
			err: nil,
		},
	}

	for _, test := range tests {
		releasePlugins, hasPauseReleasePlugin, err := BuildReleasePluginsByConfigValues(test.configValues)
		assert.IsType(t, test.err, err)
		assert.ElementsMatch(t, test.releasePlugins, releasePlugins)
		assert.Equal(t, test.hasPauseReleasePlugin, hasPauseReleasePlugin)
	}
}
