package walm

import (
	"k8s.io/helm/pkg/hapi/release"
	"k8s.io/helm/pkg/tiller/environment"
	"k8s.io/apimachinery/pkg/runtime"
	"bytes"
	"strings"
	"github.com/ghodss/yaml"
)

var pluginRunners map[string]*WalmPluginRunner

func Register(name string, runner *WalmPluginRunner) {
	if pluginRunners == nil {
		pluginRunners = map[string]*WalmPluginRunner{}
	}
	pluginRunners[name] = runner
}

type WalmPluginRunner struct {
	Run  func(context *WalmPluginManagerContext, args string) error
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
	Name    string `json:"name" description:"plugin name"`
	Args    string `json:"args" description:"plugin args"`
	Version string `json:"version" description:"plugin version"`
	Disable bool   `json:"disable" description:"disable plugin"`
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
	Log        func(string, ...interface{})
	Resources  []runtime.Object
}

func NewWalmPluginManager(kubeClient environment.KubeClient, r *release.Release, log func(string, ...interface{})) (manager *WalmPluginManager) {
	manager = &WalmPluginManager{
		plugins: map[RunnerType][]*WalmPlugin{},
		context: &WalmPluginManagerContext{
			KubeClient: kubeClient,
			R:          r,
			Log:        log,
		},
	}
	if len(r.Config) > 0 {
		walmPlugins, ok := r.Config[WalmPluginConfigKey]
		if ok {
			for _, plugin := range walmPlugins.([]*WalmPlugin) {
				if plugin.Disable {
					continue
				}
				runner := plugin.getRunner()
				if runner != nil {
					manager.plugins[runner.Type] = append(manager.plugins[runner.Type], plugin)
				} else {
					log("failed to get runner of plugin %s", plugin.Name)
				}
			}
		}
	}
	return
}

func (manager *WalmPluginManager) ExecPlugins(runnerType RunnerType) error {
	manager.context.Log("start to exec %s plugins of release %s/%s", runnerType, manager.context.R.Namespace, manager.context.R.Name)
	if runnerType == Pre_Install {
		resources, err := manager.context.KubeClient.BuildUnstructured(manager.context.R.Namespace, bytes.NewBufferString(manager.context.R.Manifest))
		if err != nil {
			manager.context.Log("failed to build k8s resources : %s", err.Error())
			return err
		}
		manager.context.Resources = []runtime.Object{}
		for _, resource := range resources {
			manager.context.Resources = append(manager.context.Resources, resource.Object)
		}
	}

	for _, plugin := range manager.plugins[runnerType] {
		runner := plugin.getRunner()
		if runner != nil {
			manager.context.Log("start to exec %s plugin %s", runnerType, plugin.Name)
			err := runner.Run(manager.context, plugin.Args)
			if err != nil {
				manager.context.Log("failed to exec %s plugin %s : %s", runnerType, plugin.Name, err.Error())
				return err
			}
			manager.context.Log("succeed to exec %s plugin %s", runnerType, plugin.Name)
		}
	}

	if runnerType == Pre_Install {
		manifest, err := buildManifest(manager.context.Resources)
		if err != nil {
			manager.context.Log("failed to build manifest : %s", err.Error())
			return err
		}
		manager.context.R.Manifest = manifest
	}
	manager.context.Log("succeed to exec %s plugins of release %s/%s", runnerType, manager.context.R.Namespace, manager.context.R.Name)
	return nil
}

func buildManifest(resources []runtime.Object) (string, error) {
	var sb strings.Builder
	for _, resource := range resources {
		resourceBytes, err := yaml.Marshal(resource)
		if err != nil {
			return "", err
		}
		sb.WriteString("\n---\n")
		sb.Write(resourceBytes)
	}
	return sb.String(), nil
}
