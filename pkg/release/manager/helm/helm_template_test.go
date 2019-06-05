package helm

import (
	"testing"
	"WarpCloud/walm/pkg/release"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var testStatefulSet1 = `
kind: StatefulSet
metadata:
  name: apacheds-master-kmd9d
  namespace: xinlin
spec:
  replicas: 1
  selector:
    matchLabels:
      transwarp.install: kmd9d
      transwarp.name: apacheds-master
  serviceName: apacheds-master-kmd9d
  template:
    metadata:
      labels:
        release: xinlin-sysctx--xinlin-guardian
        transwarp.install: kmd9d
        transwarp.name: apacheds-master
    spec:
      containers:
      - name: apacheds-master
        resources:
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
      - configMap:
          defaultMode: 420
          items:
          - key: entrypoint-apacheds.sh
            mode: 493
            path: entrypoint.sh
          - key: entrypoint-gencerts.sh
            mode: 493
            path: entrypoint-gencerts.sh
          name: guardian-entrypoint-kmd9d
        name: apacheds-entrypoint
      - configMap:
          defaultMode: 420
          items:
          - key: guardian-confd.conf
            path: guardian-confd.conf
          - key: apacheds.toml
            path: conf.d/apacheds.toml
          - key: guardian-ds.properties.tmpl
            path: templates/guardian-ds.properties.tmpl
          - key: log4j.properties.raw
            path: templates/log4j.properties.raw
          name: guardian-confd-conf-kmd9d
        name: apacheds-confd-conf
  volumeClaimTemplates:
  - metadata:
      annotations:
        volume.beta.kubernetes.io/storage-class: silver
      labels:
        release: xinlin-sysctx--xinlin-guardian
        transwarp.install: kmd9d
        transwarp.name: apacheds-data
      name: apacheds-data
    spec:
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 100Gi
      volumeMode: Filesystem
`

var testDeployment1 = `
kind: Deployment
metadata:
  name: apacheds-master-kmd9d
  namespace: xinlin
spec:
  replicas: 3
  selector:
    matchLabels:
      transwarp.install: kmd9d
      transwarp.name: apacheds-master
  template:
    metadata:
      labels:
        release: xinlin-sysctx--xinlin-guardian
        transwarp.install: kmd9d
        transwarp.name: apacheds-master
    spec:
      containers:
      - name: apacheds-master
        resources:
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
      - configMap:
          defaultMode: 420
          items:
          - key: entrypoint-apacheds.sh
            mode: 493
            path: entrypoint.sh
          - key: entrypoint-gencerts.sh
            mode: 493
            path: entrypoint-gencerts.sh
          name: guardian-entrypoint-kmd9d
        name: apacheds-entrypoint
      - configMap:
          defaultMode: 420
          items:
          - key: guardian-confd.conf
            path: guardian-confd.conf
          - key: apacheds.toml
            path: conf.d/apacheds.toml
          - key: guardian-ds.properties.tmpl
            path: templates/guardian-ds.properties.tmpl
          - key: log4j.properties.raw
            path: templates/log4j.properties.raw
          name: guardian-confd-conf-kmd9d
        name: apacheds-confd-conf
`

func Test_BuildReleaseResourceStatefulSet(t *testing.T) {
	tests := []struct {
		objectStr                   string
		releaseResourceStatefulSets *release.ReleaseResourceStatefulSet
		err                         error
	}{
		{
			objectStr: testStatefulSet1,
			releaseResourceStatefulSets: &release.ReleaseResourceStatefulSet{
				Replicas: 1,
				ReleaseResourceBase: release.ReleaseResourceBase{
					Name: "apacheds-master-kmd9d",
					PodRequests: &release.ReleaseResourcePod{
						Memory: 4096,
						Cpu:    1,
						Storage: []*release.ReleaseResourceStorage{
							{
								Name:         "apacheds-log",
								Type:         release.TosDiskPodStorageType,
								Size:         20,
								StorageClass: "silver",
							},
							{
								Name:         "apacheds-data",
								Type:         release.PvcPodStorageType,
								Size:         100,
								StorageClass: "silver",
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
		object := map[string]interface{}{}
		err := yaml.Unmarshal([]byte(test.objectStr), &object)
		assert.Nil(t, err)

		releaseResourceStatefulSet, err := buildReleaseResourceStatefulSet(&unstructured.Unstructured{object})
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.releaseResourceStatefulSets, releaseResourceStatefulSet)
	}
}

func Test_BuildReleaseResourceDeployment(t *testing.T) {
	tests := []struct {
		objectStr                 string
		releaseResourceDeployment *release.ReleaseResourceDeployment
		err                       error
	}{
		{
			objectStr: testDeployment1,
			releaseResourceDeployment: &release.ReleaseResourceDeployment{
				Replicas: 3,
				ReleaseResourceBase: release.ReleaseResourceBase{
					Name: "apacheds-master-kmd9d",
					PodRequests: &release.ReleaseResourcePod{
						Memory: 4096,
						Cpu:    1,
						Storage: []*release.ReleaseResourceStorage{
							{
								Name:         "apacheds-log",
								Type:         release.TosDiskPodStorageType,
								Size:         20,
								StorageClass: "silver",
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
		object := map[string]interface{}{}
		err := yaml.Unmarshal([]byte(test.objectStr), &object)
		assert.Nil(t, err)

		releaseResourceDeployment, err := buildReleaseResourceDeployment(&unstructured.Unstructured{object})
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.releaseResourceDeployment, releaseResourceDeployment)
	}
}
