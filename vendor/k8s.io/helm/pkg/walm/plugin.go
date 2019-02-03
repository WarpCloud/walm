package walm

import (
	"k8s.io/helm/pkg/hapi/release"
	"k8s.io/helm/pkg/tiller/environment"
)

var pluginRunners map[string]*WalmPluginRunner

func Register(name string, runner *WalmPluginRunner) {
	if pluginRunners == nil {
		pluginRunners = map[string]*WalmPluginRunner{}
	}
	pluginRunners[name] = runner
}

type WalmPluginRunner struct {
	Run func(context *WalmPluginManagerContext, args string) error
	Type RunnerType
}

type RunnerType string

const (
	Pre_Install  RunnerType = "pre_install"
	Post_Install RunnerType = "post_install"
	Unknown      RunnerType = "unknown"

	WalmPluginConfigKey string = "Walm-Plugin-Key"
)

type WalmPlugin struct {
	Name string
	Args string
}

func (walmPlugin *WalmPlugin) getRunner() *WalmPluginRunner {
	if pluginRunners == nil {
		return nil
	}
	return pluginRunners[walmPlugin.Name]
}

type WalmPluginManager struct {
	context *WalmPluginManagerContext
	plugins map[RunnerType][]*WalmPlugin
}

type WalmPluginManagerContext struct {
	KubeClient environment.KubeClient
	R          *release.Release
}

func NewWalmPluginManager(kubeClient environment.KubeClient, r *release.Release) (manager *WalmPluginManager) {
	manager = &WalmPluginManager{
		plugins: map[RunnerType][]*WalmPlugin{},
		context: &WalmPluginManagerContext{
			KubeClient: kubeClient,
			R:          r,
		},
	}
	if len(r.Config) > 0 {
		walmPlugins, ok := r.Config[WalmPluginConfigKey]
		if ok {
			for _, plugin := range walmPlugins.([]*WalmPlugin){
				runner := plugin.getRunner()
				if runner != nil {
					manager.plugins[runner.Type] = append(manager.plugins[runner.Type], plugin)
				}
			}
		}
	}
	return
}

func (manager *WalmPluginManager) ExecPlugins(runnerType RunnerType) error{
	for _, plugin := range manager.plugins[runnerType] {
		runner := plugin.getRunner()
		if runner != nil {
			err := runner.Run(manager.context, plugin.Args)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
