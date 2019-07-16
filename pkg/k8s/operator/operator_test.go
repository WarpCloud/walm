package operator

import (
	"testing"
	"WarpCloud/walm/pkg/models/release"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"github.com/ghodss/yaml"
)

func Test_BuildPvcStorage(t *testing.T) {
	testStorageClass := "test-storage-class"
	tests := []struct {
		pvc     v1.PersistentVolumeClaim
		storage *release.ReleaseResourceStorage
	}{
		{
			pvc: v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pvc",
				},
				Spec: v1.PersistentVolumeClaimSpec{
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceStorage: resource.MustParse("100Gi"),
						},
					},
					StorageClassName: &testStorageClass,
				},
			},
			storage: &release.ReleaseResourceStorage{
				Name:         "test-pvc",
				StorageClass: "test-storage-class",
				Type:         release.PvcPodStorageType,
				Size:         100,
			},
		},
		{
			pvc: v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pvc",
					Annotations: map[string]string{
						storageClassAnnotationKey: "test-storage-class1",
					},
				},
				Spec: v1.PersistentVolumeClaimSpec{
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceStorage: resource.MustParse("100Gi"),
						},
					},
				},
			},
			storage: &release.ReleaseResourceStorage{
				Name:         "test-pvc",
				StorageClass: "test-storage-class1",
				Type:         release.PvcPodStorageType,
				Size:         100,
			},
		},
	}

	for _, test := range tests {
		storage := buildPvcStorage(test.pvc)
		assert.Equal(t, test.storage, storage)
	}
}

func Test_BuildReleaseResourceBase(t *testing.T) {
	unstructuredString := `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: test-statefulset
spec:
  template:
    spec:
      volumes:
      - name: apacheds-log
        tosDisk:
          accessMode: ReadWriteOnce
          capability: 20Gi
          name: apacheds-log
          storageType: silver
`
	tests := []struct {
		r               *unstructured.Unstructured
		podTemplateSpec v1.PodTemplateSpec
		pvcs            []v1.PersistentVolumeClaim
		releaseResource release.ReleaseResourceBase
		err             error
	}{
		{
			r: convertYamlStringToUnstructured(unstructuredString, t),
			podTemplateSpec: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("0.2"),
									v1.ResourceMemory: resource.MustParse("200Mi"),
								},
								Limits: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("2"),
									v1.ResourceMemory: resource.MustParse("2Gi"),
								},
							},
						},
					},
				},
			},
			pvcs: []v1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-pvc",
						Annotations: map[string]string{
							storageClassAnnotationKey: "test-storage-class1",
						},
					},
					Spec: v1.PersistentVolumeClaimSpec{
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: resource.MustParse("100Gi"),
							},
						},
					},
				},
			},
			releaseResource: release.ReleaseResourceBase{
				Name: "test-statefulset",
				PodRequests: &release.ReleaseResourcePod{
					Memory: 200,
					Cpu:    0.2,
					Storage: []*release.ReleaseResourceStorage{
						{
							Name:         "apacheds-log",
							StorageClass: "silver",
							Type:         release.TosDiskPodStorageType,
							Size:         20,
						},
						{
							Name:         "test-pvc",
							StorageClass: "test-storage-class1",
							Type:         release.PvcPodStorageType,
							Size:         100,
						},
					},
				},
				PodLimits: &release.ReleaseResourcePod{
					Memory: 2048,
					Cpu:    2,
				},
			},
			err: nil,
		},
	}

	for _, test := range tests {
		releaseResource, err := buildReleaseResourceBase(test.r, test.podTemplateSpec, test.pvcs)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.releaseResource, releaseResource)
	}
}

func Test_BuildReleaseResourceDeployment(t *testing.T) {
	unstructuredString := `
kind: Deployment
metadata:
  name: test-deployment
spec:
  replicas: 3
  template:
    spec:
      containers:
      - resources:
          limits:
            cpu: "2"
            memory: 4Gi
          requests:
            cpu: "1"
            memory: 4Gi
      volumes:
      - name: apacheds-log
        tosDisk:
          accessMode: ReadWriteOnce
          capability: 20Gi
          name: apacheds-log
          storageType: silver
`
	tests := []struct {
		r               *unstructured.Unstructured
		releaseResource *release.ReleaseResourceDeployment
		err             error
	}{
		{
			r: convertYamlStringToUnstructured(unstructuredString, t),
			releaseResource: &release.ReleaseResourceDeployment{
				Replicas: 3,
				ReleaseResourceBase: release.ReleaseResourceBase{
					Name: "test-deployment",
					PodRequests: &release.ReleaseResourcePod{
						Memory: 4096,
						Cpu:    1,
						Storage: []*release.ReleaseResourceStorage{
							{
								Name:         "apacheds-log",
								StorageClass: "silver",
								Type:         release.TosDiskPodStorageType,
								Size:         20,
							},
						},
					},
					PodLimits: &release.ReleaseResourcePod{
						Memory: 4096,
						Cpu:    2,
					},
				},
			},
			err: nil,
		},
	}

	for _, test := range tests {
		releaseResource, err := buildReleaseResourceDeployment(test.r)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.releaseResource, releaseResource)
	}
}

func Test_BuildReleaseResourceStatefulSet(t *testing.T) {
	unstructuredString := `
kind: StatefulSet
metadata:
  name: test-statefulset
spec:
  replicas: 3
  template:
    spec:
      containers:
      - resources:
          limits:
            cpu: "2"
            memory: 4Gi
          requests:
            cpu: "1"
            memory: 4Gi
      volumes:
      - name: apacheds-log
        tosDisk:
          accessMode: ReadWriteOnce
          capability: 20Gi
          name: apacheds-log
          storageType: silver
  volumeClaimTemplates:
  - metadata:
      annotations:
        volume.beta.kubernetes.io/storage-class: silver
      name: apacheds-data
    spec:
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 100Gi
      volumeMode: Filesystem
`
	tests := []struct {
		r               *unstructured.Unstructured
		releaseResource *release.ReleaseResourceStatefulSet
		err             error
	}{
		{
			r: convertYamlStringToUnstructured(unstructuredString, t),
			releaseResource: &release.ReleaseResourceStatefulSet{
				Replicas: 3,
				ReleaseResourceBase: release.ReleaseResourceBase{
					Name: "test-statefulset",
					PodRequests: &release.ReleaseResourcePod{
						Memory: 4096,
						Cpu:    1,
						Storage: []*release.ReleaseResourceStorage{
							{
								Name:         "apacheds-log",
								StorageClass: "silver",
								Type:         release.TosDiskPodStorageType,
								Size:         20,
							},
							{
								Name:         "apacheds-data",
								StorageClass: "silver",
								Type:         release.PvcPodStorageType,
								Size:         100,
							},
						},
					},
					PodLimits: &release.ReleaseResourcePod{
						Memory: 4096,
						Cpu:    2,
					},
				},
			},
			err: nil,
		},
	}

	for _, test := range tests {
		releaseResource, err := buildReleaseResourceStatefulSet(test.r)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.releaseResource, releaseResource)
	}
}

func Test_BuildReleaseResourceDaemonSet(t *testing.T) {
	unstructuredString := `
kind: DaemonSet
metadata:
  name: test-daemonset
spec:
  template:
    spec:
      nodeSelector:
        master: "true"
      containers:
      - resources:
          limits:
            cpu: "2"
            memory: 4Gi
          requests:
            cpu: "1"
            memory: 4Gi
`
	tests := []struct {
		r               *unstructured.Unstructured
		releaseResource *release.ReleaseResourceDaemonSet
		err             error
	}{
		{
			r: convertYamlStringToUnstructured(unstructuredString, t),
			releaseResource: &release.ReleaseResourceDaemonSet{
				NodeSelector: map[string]string{
					"master": "true",
				},
				ReleaseResourceBase: release.ReleaseResourceBase{
					Name: "test-daemonset",
					PodRequests: &release.ReleaseResourcePod{
						Memory: 4096,
						Cpu:    1,
						Storage: []*release.ReleaseResourceStorage{
						},
					},
					PodLimits: &release.ReleaseResourcePod{
						Memory: 4096,
						Cpu:    2,
					},
				},
			},
			err: nil,
		},
	}

	for _, test := range tests {
		releaseResource, err := buildReleaseResourceDaemonSet(test.r)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.releaseResource, releaseResource)
	}
}

func Test_BuildReleaseResourceJob(t *testing.T) {
	unstructuredString := `
kind: Job
metadata:
  name: test-job
spec:
  template:
    spec:
      containers:
      - resources:
          limits:
            cpu: "2"
            memory: 4Gi
          requests:
            cpu: "1"
            memory: 4Gi
`

	resource := convertYamlStringToUnstructured(unstructuredString, t)
	resource.Object["spec"].(map[string]interface{})["parallelism"] = int32(2)
	resource.Object["spec"].(map[string]interface{})["completions"] = int32(2)

	tests := []struct {
		r               *unstructured.Unstructured
		releaseResource *release.ReleaseResourceJob
		err             error
	}{
		{
			r: resource,
			releaseResource: &release.ReleaseResourceJob{
				Completions: 2,
				Parallelism: 2,
				ReleaseResourceBase: release.ReleaseResourceBase{
					Name: "test-job",
					PodRequests: &release.ReleaseResourcePod{
						Memory: 4096,
						Cpu:    1,
						Storage: []*release.ReleaseResourceStorage{
						},
					},
					PodLimits: &release.ReleaseResourcePod{
						Memory: 4096,
						Cpu:    2,
					},
				},
			},
			err: nil,
		},
	}

	for _, test := range tests {
		releaseResource, err := buildReleaseResourceJob(test.r)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.releaseResource, releaseResource)
	}
}

func Test_BuildReleaseResourcePvc(t *testing.T) {
	unstructuredString := `
kind: PersistentVolumeClaim
metadata:
  name: test-pvc
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 8Gi
  storageClassName: silver
  volumeMode: Filesystem
`
	tests := []struct {
		r               *unstructured.Unstructured
		releaseResource *release.ReleaseResourceStorage
		err             error
	}{
		{
			r: convertYamlStringToUnstructured(unstructuredString, t),
			releaseResource: &release.ReleaseResourceStorage{
				Name:         "test-pvc",
				StorageClass: "silver",
				Type:         release.PvcPodStorageType,
				Size:         8,
			},
			err: nil,
		},
	}

	for _, test := range tests {
		releaseResource, err := buildReleaseResourcePvc(test.r)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.releaseResource, releaseResource)
	}
}

func convertYamlStringToUnstructured(objStr string, t *testing.T) *unstructured.Unstructured {
	obj := map[string]interface{}{}
	err := yaml.Unmarshal([]byte(objStr), &obj)
	if err != nil {
		assert.Fail(t, err.Error())
		return nil
	}
	return &unstructured.Unstructured{Object: obj}
}
