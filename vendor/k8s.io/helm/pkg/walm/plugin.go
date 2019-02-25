package walm

import (
	"k8s.io/helm/pkg/hapi/release"
	"k8s.io/helm/pkg/tiller/environment"
	"k8s.io/apimachinery/pkg/runtime"
	"bytes"
	"strings"
	"github.com/ghodss/yaml"
	"fmt"
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
	Resources  map[string]runtime.Object
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

func BuildResourceKey(kind, namespace, name string) string {
	return kind + "/" + namespace + "/" + name
}

func (manager *WalmPluginManager) ExecPlugins(runnerType RunnerType) error{
	if runnerType == Pre_Install {
		resources, err := manager.context.KubeClient.BuildUnstructured(manager.context.R.Namespace, bytes.NewBufferString(manager.context.R.Manifest))
		if err != nil {
			return err
		}
		manager.context.Resources = map[string]runtime.Object{}
		for _, resource := range resources {
			resourceKey := BuildResourceKey(resource.Object.GetObjectKind().GroupVersionKind().Kind,
				resource.Namespace, resource.Name)
			if _, ok := manager.context.Resources[resourceKey]; ok {
				return fmt.Errorf("%s %s in namespace %s has already existed", resource.Object.GetObjectKind().GroupVersionKind().Kind,
					resource.Namespace, resource.Name)
			} else {
				manager.context.Resources[resourceKey] = resource.Object
			}
		}
	}

	for _, plugin := range manager.plugins[runnerType] {
		runner := plugin.getRunner()
		if runner != nil {
			err := runner.Run(manager.context, plugin.Args)
			if err != nil {
				return err
			}
		}
	}

	if runnerType == Pre_Install {
		manifest, err := buildManifest(manager.context.Resources)
		if err != nil {
			return err
		}
		manager.context.R.Manifest = manifest
	}

	return nil
}

func buildManifest(resources map[string]runtime.Object) (string,  error) {
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