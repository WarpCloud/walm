package walm

import (
	"k8s.io/helm/pkg/hapi/release"
	"k8s.io/helm/pkg/tiller/environment"
	"k8s.io/apimachinery/pkg/runtime"
	"bytes"
	"strings"
	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes/scheme"
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
	Resources  []runtime.Object
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
			for _, plugin := range walmPlugins.([]*WalmPlugin) {
				runner := plugin.getRunner()
				if runner != nil {
					manager.plugins[runner.Type] = append(manager.plugins[runner.Type], plugin)
				}
			}
		}
	}
	return
}

func convertUnstructured(unStruct *unstructured.Unstructured) (runtime.Object, error) {
	unStructBytes, err := unStruct.MarshalJSON()
	if err != nil {
		return nil, err
	}

	defaultGVK := unStruct.GetObjectKind().GroupVersionKind()

	decoder := scheme.Codecs.UniversalDecoder(defaultGVK.GroupVersion())
	obj, _, err := decoder.Decode(unStructBytes, &defaultGVK, nil)
	if err != nil {
		return nil,  err
	}
	return obj, nil
}

func (manager *WalmPluginManager) ExecPlugins(runnerType RunnerType) error {
	if runnerType == Pre_Install {
		resources, err := manager.context.KubeClient.BuildUnstructured(manager.context.R.Namespace, bytes.NewBufferString(manager.context.R.Manifest))
		if err != nil {
			return err
		}
		manager.context.Resources = []runtime.Object{}
		for _, resource := range resources {
			unStruct := resource.Object.(*unstructured.Unstructured)
			obj, err := convertUnstructured(unStruct)
			if err != nil {
				return err
			}
			manager.context.Resources = append(manager.context.Resources, obj)
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
