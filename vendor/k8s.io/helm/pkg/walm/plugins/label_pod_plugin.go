package plugins

import (
	"k8s.io/helm/pkg/walm"
	"encoding/json"
	"k8s.io/api/extensions/v1beta1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
)

const (
	LabelPodPluginName = "LabelPod"
)

// ValidateReleaseConfig plugin is used to make sure:
// 1. release have and only have one ReleaseConfig
// 2. ReleaseConfig has the same namespace and name with the release

func init() {
	walm.Register(LabelPodPluginName, &walm.WalmPluginRunner{
		Run:  LabelPod,
		Type: walm.Pre_Install,
	})
}

type LabelPodArgs struct {
	LabelsToAdd      map[string]string   `json:"labelsToAdd" description:"labels to add"`
	AnnotationsToAdd map[string]string   `json:"annotationsToAdd" description:"annotations to add"`
}

func LabelPod(context *walm.WalmPluginManagerContext, args string) (err error) {
	if args == "" {
		return nil
	}
	labelPodArgs := &LabelPodArgs{}
	err = json.Unmarshal([]byte(args), labelPodArgs)
	if err != nil {
		return err
	}

	for _, resource := range context.Resources {
		switch resource.GetObjectKind().GroupVersionKind().Kind {
		case "Deployment":
			deployment := resource.(*v1beta1.Deployment)
			labelDeploymentPod(deployment, labelPodArgs)
		case "DaemonSet":
			daemonSet := resource.(*v1beta1.DaemonSet)
			labelDaemonSetPod(daemonSet, labelPodArgs)
		case "StatefulSet":
			statefulSet := resource.(*appsv1beta1.StatefulSet)
			labelStatefulSetPod(statefulSet, labelPodArgs)
		default:
		}
	}
	return
}

func labelStatefulSetPod(statefulSet *appsv1beta1.StatefulSet, labelPodArgs *LabelPodArgs) {
	if statefulSet.Spec.Template.Labels == nil {
		statefulSet.Spec.Template.Labels = labelPodArgs.LabelsToAdd
	} else {
		for k, v := range labelPodArgs.LabelsToAdd {
			statefulSet.Spec.Template.Labels[k] = v
		}
	}
	if statefulSet.Spec.Template.Annotations == nil {
		statefulSet.Spec.Template.Annotations = labelPodArgs.AnnotationsToAdd
	} else {
		for k, v := range labelPodArgs.AnnotationsToAdd {
			statefulSet.Spec.Template.Annotations[k] = v
		}
	}
}

func labelDaemonSetPod(daemonSet *v1beta1.DaemonSet, labelPodArgs *LabelPodArgs) {
	if daemonSet.Spec.Template.Labels == nil {
		daemonSet.Spec.Template.Labels = labelPodArgs.LabelsToAdd
	} else {
		for k, v := range labelPodArgs.LabelsToAdd {
			daemonSet.Spec.Template.Labels[k] = v
		}
	}
	if daemonSet.Spec.Template.Annotations == nil {
		daemonSet.Spec.Template.Annotations = labelPodArgs.AnnotationsToAdd
	} else {
		for k, v := range labelPodArgs.AnnotationsToAdd {
			daemonSet.Spec.Template.Annotations[k] = v
		}
	}
}

func labelDeploymentPod(deployment *v1beta1.Deployment, labelPodArgs *LabelPodArgs) {
	if deployment.Spec.Template.Labels == nil {
		deployment.Spec.Template.Labels = labelPodArgs.LabelsToAdd
	} else {
		for k, v := range labelPodArgs.LabelsToAdd {
			deployment.Spec.Template.Labels[k] = v
		}
	}
	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = labelPodArgs.AnnotationsToAdd
	} else {
		for k, v := range labelPodArgs.AnnotationsToAdd {
			deployment.Spec.Template.Annotations[k] = v
		}
	}
}

