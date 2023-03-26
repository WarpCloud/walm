package helm

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"WarpCloud/walm/pkg/helm/impl/plugins"
	"WarpCloud/walm/pkg/models/k8s"
)

func Test_BuildReleasePluginsByConfigValues(t *testing.T) {
	tests := []struct {
		configValues          map[string]interface{}
		releasePlugins        []*k8s.ReleasePlugin
		hasPauseReleasePlugin bool
		err                   error
	}{
		{
			configValues: map[string]interface{}{
				plugins.WalmPluginConfigKey: []interface{}{
					map[string]interface{}{
						"name": "test-plugin",
						"args": "test-args",
						"version": "test-version",
						"disable": false,
					},
				},
			},
			releasePlugins: []*k8s.ReleasePlugin{
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
				plugins.WalmPluginConfigKey: []interface{}{
					map[string]interface{}{
						"name": "test-plugin",
						"args": "test-args",
						"version": "test-version",
						"disable": true,
					},
				},
			},
			releasePlugins: []*k8s.ReleasePlugin{
			},
			hasPauseReleasePlugin: false,
			err: nil,
		},
		{
			configValues: map[string]interface{}{
				plugins.WalmPluginConfigKey: []interface{}{
					map[string]interface{}{
						"name": plugins.ValidateReleaseConfigPluginName,
						"args": "test-args",
						"version": "test-version",
						"disable": false,
					},
				},
			},
			releasePlugins: []*k8s.ReleasePlugin{
			},
			hasPauseReleasePlugin: false,
			err: nil,
		},
		{
			configValues: map[string]interface{}{
				plugins.WalmPluginConfigKey: []interface{}{
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
			releasePlugins: []*k8s.ReleasePlugin{
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
