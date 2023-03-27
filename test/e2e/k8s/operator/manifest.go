package operator

import (
	. "github.com/onsi/gomega"
	. "github.com/onsi/ginkgo"
	"WarpCloud/walm/test/e2e/framework"
	"WarpCloud/walm/pkg/k8s/operator"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"fmt"
	"WarpCloud/walm/pkg/models/release"
)

var _ = Describe("K8sOperatorManifest", func() {

	var (
		namespace   string
		k8sOperator *operator.Operator
		err         error
	)

	BeforeEach(func() {
		By("create namespace")
		namespace, err = framework.CreateRandomNamespace("k8sOperatorPodTest", nil)
		Expect(err).NotTo(HaveOccurred())
		k8sOperator = operator.NewOperator(nil, nil, framework.GetKubeClient(), nil)
	})

	AfterEach(func() {
		By("delete namespace")
		err = framework.DeleteNamespace(namespace)
		Expect(err).NotTo(HaveOccurred())
	})

	It("test build objects by manifest", func() {
		manifest := `
apiVersion: v1
kind: Service
metadata:
  name: my-nginx-svc
  labels:
    app: nginx
spec:
  type: LoadBalancer
  ports:
  - port: 80
  selector:
    app: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-nginx
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
        ports:
        - containerPort: 80
`
		objects, err := k8sOperator.BuildManifestObjects(namespace, manifest)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(objects)).To(Equal(2))

		objectStrings, err := convertToObjectsStrings(objects)
		Expect(err).NotTo(HaveOccurred())
		Expect(objectStrings).To(Equal([]string{
			fmt.Sprintf("{\"apiVersion\":\"v1\",\"kind\":\"Service\",\"metadata\":{\"labels\":{\"app\":\"nginx\"},\"name\":\"my-nginx-svc\",\"namespace\":\"%s\"},\"spec\":{\"ports\":[{\"port\":80}],\"selector\":{\"app\":\"nginx\"},\"type\":\"LoadBalancer\"}}", namespace),
			fmt.Sprintf("{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"labels\":{\"app\":\"nginx\"},\"name\":\"my-nginx\",\"namespace\":\"%s\"},\"spec\":{\"replicas\":3,\"selector\":{\"matchLabels\":{\"app\":\"nginx\"}},\"template\":{\"metadata\":{\"labels\":{\"app\":\"nginx\"}},\"spec\":{\"containers\":[{\"image\":\"nginx:1.7.9\",\"name\":\"nginx\",\"ports\":[{\"containerPort\":80}]}]}}}}",	namespace)	}))
	})

	It("test Compute Release Resources by manifest", func() {
		manifest := `
apiVersion: v1
kind: Service
metadata:
  name: my-nginx-svc
  labels:
    app: nginx
spec:
  type: LoadBalancer
  ports:
  - port: 80
  selector:
    app: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-nginx
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
        resources:
          requests:
            cpu: 0.1
            memory: 100Mi
          limits:
            cpu: 1
            memory: 1Gi
        ports:
        - containerPort: 80
`
		releaseResources, err := k8sOperator.ComputeReleaseResourcesByManifest(namespace, manifest)
		Expect(err).NotTo(HaveOccurred())
		Expect(releaseResources).To(Equal(&release.ReleaseResources{
			Deployments: []*release.ReleaseResourceDeployment{
				{
					Replicas: 3,
					ReleaseResourceBase: release.ReleaseResourceBase{
						Name: "my-nginx",
						PodLimits: &release.ReleaseResourcePod{
							Cpu: 1,
							Memory: 1024,
						},
						PodRequests: &release.ReleaseResourcePod{
							Cpu: 0.1,
							Memory: 100,
							Storage: []*release.ReleaseResourceStorage{},
						},
					},
				},
			},
		}))
	})
})

func convertToObjectsStrings(objects []map[string]interface{}) ([]string, error) {
	objectStrings := []string{}
	for _, value := range objects {
		valueString, err := json.Marshal(value)
		if err != nil {
			logrus.Errorf("failed to marshal value : %s", err.Error())
			return nil, err
		}
		objectStrings = append(objectStrings, string(valueString))
	}
	return objectStrings, nil
}
