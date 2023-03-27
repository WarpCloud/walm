package cache

import (
	"WarpCloud/walm/pkg/k8s/cache/informer"
	errorModel "WarpCloud/walm/pkg/models/error"
	"WarpCloud/walm/test/e2e/framework"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
	"transwarp/release-config/pkg/apis/transwarp/v1beta1"

	k8sModel "WarpCloud/walm/pkg/models/k8s"
	"WarpCloud/walm/pkg/models/release"
	"WarpCloud/walm/pkg/models/tenant"
	storagev1 "k8s.io/api/storage/v1"
)

var _ = Describe("K8sCache", func() {

	var (
		namespace string
		k8sCache  *informer.Informer
		err       error
		stopChan  chan struct{}
	)

	BeforeEach(func() {
		By("create namespace")
		namespace, err = framework.CreateRandomNamespace("k8sCacheTest", nil)
		Expect(err).NotTo(HaveOccurred())
		stopChan = make(chan struct{})
		k8sCache = informer.NewInformer(framework.GetK8sClient(), framework.GetK8sReleaseConfigClient(), framework.GetK8sInstanceClient(), framework.GetK8sMigrationClient(), nil, nil, 0, stopChan)
	})

	AfterEach(func() {
		By("delete namespace")
		close(stopChan)
		err = framework.DeleteNamespace(namespace)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("test resource", func() {
		var (
			storageClass  *storagev1.StorageClass
			storageClass2 *storagev1.StorageClass
		)

		BeforeEach(func() {
			storageClass, err = framework.CreateStorageClass(namespace, "test-storageclass1", map[string]string{"test": "true"})
			Expect(err).NotTo(HaveOccurred())

			storageClass2, err = framework.CreateStorageClass(namespace, "test-storageclass2", map[string]string{"test": "false"})
			Expect(err).NotTo(HaveOccurred())

		})

		AfterEach(func() {
			err = framework.DeleteStorageClass(namespace, storageClass.Name)
			Expect(err).NotTo(HaveOccurred())

			err = framework.DeleteStorageClass(namespace, storageClass2.Name)
			Expect(err).NotTo(HaveOccurred())
		})

		It("test get resource & get resource set", func() {

			By("create resources")
			releaseConfig, err := framework.CreateReleaseConfig(namespace, "test-relconf", nil)
			Expect(err).NotTo(HaveOccurred())

			configMap, err := framework.CreateConfigMap(namespace, "test-cm")
			Expect(err).NotTo(HaveOccurred())

			pvc, err := framework.CreatePvc(namespace, "test-pvc", nil)
			Expect(err).NotTo(HaveOccurred())

			daemonSet, err := framework.CreateDaemonSet(namespace, "test-daemonset")
			Expect(err).NotTo(HaveOccurred())

			deployment, err := framework.CreateDeployment(namespace, "test-deploy")
			Expect(err).NotTo(HaveOccurred())

			service, err := framework.CreateService(namespace, "test-service", deployment.Spec.Selector.MatchLabels)
			Expect(err).NotTo(HaveOccurred())
			serviceWithoutEndpoints, err := framework.CreateService(namespace, "test-svc-noep", map[string]string{"notExist": "true"})
			Expect(err).NotTo(HaveOccurred())

			statefulSet, err := framework.CreateStatefulSet(namespace, "test-sts", nil)
			Expect(err).NotTo(HaveOccurred())

			job, err := framework.CreateJob(namespace, "test-job")
			Expect(err).NotTo(HaveOccurred())

			ingress, err := framework.CreateIngress(namespace, "test-ing")
			Expect(err).NotTo(HaveOccurred())

			secret, err := framework.CreateSecret(namespace, "test-secret", nil)
			Expect(err).NotTo(HaveOccurred())

			replicaSet, err := framework.CreateReplicaSet(namespace, "test-replicaset", nil)
			Expect(err).NotTo(HaveOccurred())

			instance, err := framework.CreateInstance(namespace, "test-instance", nil)
			Expect(err).NotTo(HaveOccurred())

			migration, err := framework.CreateMigration(namespace, "test-migration", nil)
			node, err := framework.GetTestNode()
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(time.Second)

			By("get resources")
			resource, err := k8sCache.GetResource(k8sModel.ReleaseConfigKind, namespace, releaseConfig.Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.GetKind()).To(Equal(k8sModel.ReleaseConfigKind))
			_, err = k8sCache.GetResource(k8sModel.ReleaseConfigKind, namespace, "notExisted")
			Expect(err).To(Equal(errorModel.NotFoundError{}))

			resource, err = k8sCache.GetResource(k8sModel.ConfigMapKind, namespace, configMap.Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.GetKind()).To(Equal(k8sModel.ConfigMapKind))
			_, err = k8sCache.GetResource(k8sModel.ConfigMapKind, namespace, "notExisted")
			Expect(err).To(Equal(errorModel.NotFoundError{}))

			resource, err = k8sCache.GetResource(k8sModel.PersistentVolumeClaimKind, namespace, pvc.Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.GetKind()).To(Equal(k8sModel.PersistentVolumeClaimKind))
			_, err = k8sCache.GetResource(k8sModel.PersistentVolumeClaimKind, namespace, "notExisted")
			Expect(err).To(Equal(errorModel.NotFoundError{}))

			resource, err = k8sCache.GetResource(k8sModel.DaemonSetKind, namespace, daemonSet.Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.GetKind()).To(Equal(k8sModel.DaemonSetKind))
			_, err = k8sCache.GetResource(k8sModel.DaemonSetKind, namespace, "notExisted")
			Expect(err).To(Equal(errorModel.NotFoundError{}))

			resource, err = k8sCache.GetResource(k8sModel.DeploymentKind, namespace, deployment.Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.GetKind()).To(Equal(k8sModel.DeploymentKind))
			_, err = k8sCache.GetResource(k8sModel.DeploymentKind, namespace, "notExisted")
			Expect(err).To(Equal(errorModel.NotFoundError{}))

			resource, err = k8sCache.GetResource(k8sModel.ServiceKind, namespace, service.Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.GetKind()).To(Equal(k8sModel.ServiceKind))
			resource, err = k8sCache.GetResource(k8sModel.ServiceKind, namespace, serviceWithoutEndpoints.Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.GetKind()).To(Equal(k8sModel.ServiceKind))
			_, err = k8sCache.GetResource(k8sModel.ServiceKind, namespace, "notExisted")
			Expect(err).To(Equal(errorModel.NotFoundError{}))

			resource, err = k8sCache.GetResource(k8sModel.StatefulSetKind, namespace, statefulSet.Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.GetKind()).To(Equal(k8sModel.StatefulSetKind))
			_, err = k8sCache.GetResource(k8sModel.StatefulSetKind, namespace, "notExisted")
			Expect(err).To(Equal(errorModel.NotFoundError{}))

			resource, err = k8sCache.GetResource(k8sModel.JobKind, namespace, job.Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.GetKind()).To(Equal(k8sModel.JobKind))
			_, err = k8sCache.GetResource(k8sModel.JobKind, namespace, "notExisted")
			Expect(err).To(Equal(errorModel.NotFoundError{}))

			resource, err = k8sCache.GetResource(k8sModel.IngressKind, namespace, ingress.Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.GetKind()).To(Equal(k8sModel.IngressKind))
			_, err = k8sCache.GetResource(k8sModel.IngressKind, namespace, "notExisted")
			Expect(err).To(Equal(errorModel.NotFoundError{}))

			resource, err = k8sCache.GetResource(k8sModel.SecretKind, namespace, secret.Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.GetKind()).To(Equal(k8sModel.SecretKind))
			_, err = k8sCache.GetResource(k8sModel.SecretKind, namespace, "notExisted")
			Expect(err).To(Equal(errorModel.NotFoundError{}))

			resource, err = k8sCache.GetResource(k8sModel.ReplicaSetKind, namespace, replicaSet.Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.GetKind()).To(Equal(k8sModel.ReplicaSetKind))
			_, err = k8sCache.GetResource(k8sModel.ReplicaSetKind, namespace, "notExisted")
			Expect(err).To(Equal(errorModel.NotFoundError{}))

			resource, err = k8sCache.GetResource(k8sModel.InstanceKind, namespace, instance.Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.GetKind()).To(Equal(k8sModel.InstanceKind))
			_, err = k8sCache.GetResource(k8sModel.InstanceKind, namespace, "notExisted")
			Expect(err).To(Equal(errorModel.NotFoundError{}))

			resource, err = k8sCache.GetResource(k8sModel.MigKind, namespace, migration.Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.GetKind()).To(Equal(k8sModel.MigKind))
			_, err = k8sCache.GetResource(k8sModel.MigKind, namespace, "notExisted")
			Expect(err).To(Equal(errorModel.NotFoundError{}))

			resource, err = k8sCache.GetResource(k8sModel.NodeKind, namespace, node.Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.GetKind()).To(Equal(k8sModel.NodeKind))
			_, err = k8sCache.GetResource(k8sModel.NodeKind, namespace, "notExisted")
			Expect(err).To(Equal(errorModel.NotFoundError{}))

			resource, err = k8sCache.GetResource(k8sModel.StorageClassKind, namespace, storageClass.Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.GetKind()).To(Equal(k8sModel.StorageClassKind))
			_, err = k8sCache.GetResource(k8sModel.StorageClassKind, namespace, "notExisted")
			Expect(err).To(Equal(errorModel.NotFoundError{}))

			resource, err = k8sCache.GetResource("notSupportedKind", "anything", "anything")
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.GetKind()).To(Equal(k8sModel.ResourceKind("notSupportedKind")))

			By("get resourceSet")
			// now only 8 kinds are supported, if resource meta contains not supported kind, it would be ignored.
			// if resource is not found, it would be added to resourceSet
			releaseResourceMetas := []release.ReleaseResourceMeta{
				{
					Namespace: namespace,
					Name:      releaseConfig.Name,
					Kind:      k8sModel.ReleaseConfigKind,
				},
				{
					Namespace: namespace,
					Name:      configMap.Name,
					Kind:      k8sModel.ConfigMapKind,
				},
				{
					Namespace: namespace,
					Name:      daemonSet.Name,
					Kind:      k8sModel.DaemonSetKind,
				},
				{
					Namespace: namespace,
					Name:      deployment.Name,
					Kind:      k8sModel.DeploymentKind,
				},
				{
					Namespace: namespace,
					Name:      service.Name,
					Kind:      k8sModel.ServiceKind,
				},
				{
					Namespace: namespace,
					Name:      statefulSet.Name,
					Kind:      k8sModel.StatefulSetKind,
				},
				{
					Namespace: namespace,
					Name:      job.Name,
					Kind:      k8sModel.JobKind,
				},
				{
					Namespace: namespace,
					Name:      ingress.Name,
					Kind:      k8sModel.IngressKind,
				},
				{
					Namespace: namespace,
					Name:      secret.Name,
					Kind:      k8sModel.SecretKind,
				},
				{
					Namespace: namespace,
					Name:      "not-existed",
					Kind:      k8sModel.SecretKind,
				},
			}
			resourceSet, err := k8sCache.GetResourceSet(releaseResourceMetas)
			Expect(err).NotTo(HaveOccurred())
			Expect(resourceSet.Ingresses).To(HaveLen(1))
			Expect(resourceSet.Jobs).To(HaveLen(1))
			Expect(resourceSet.Services).To(HaveLen(1))
			Expect(resourceSet.Deployments).To(HaveLen(1))
			Expect(resourceSet.DaemonSets).To(HaveLen(1))
			Expect(resourceSet.ConfigMaps).To(HaveLen(1))
			Expect(resourceSet.StatefulSets).To(HaveLen(1))
			Expect(resourceSet.Secrets).To(HaveLen(2))

			By("list storage classes")
			scs, err := k8sCache.ListStorageClasses(namespace, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(len(scs) >= 2).To(BeTrue())

			scs, err = k8sCache.ListStorageClasses(namespace, "test=true")
			Expect(err).NotTo(HaveOccurred())
			Expect(scs).To(HaveLen(1))
		})

	})

	It("test release config", func() {

		By("add release config handler")
		releaseConfigs := map[string]*v1beta1.ReleaseConfig{}
		onAdd := func(obj interface{}) {
			releaseConfig, ok := obj.(*v1beta1.ReleaseConfig)
			Expect(ok).To(BeTrue())
			releaseConfigs[releaseConfig.Name] = releaseConfig
		}
		onUpdate := func(old, cur interface{}) {
			releaseConfig, ok := cur.(*v1beta1.ReleaseConfig)
			Expect(ok).To(BeTrue())
			releaseConfigs[releaseConfig.Name] = releaseConfig
		}
		onDelete := func(obj interface{}) {
			releaseConfig, ok := obj.(*v1beta1.ReleaseConfig)
			Expect(ok).To(BeTrue())
			delete(releaseConfigs, releaseConfig.Name)
		}

		k8sCache.AddReleaseConfigHandler(onAdd, onUpdate, onDelete)

		releaseConfig, err := framework.CreateReleaseConfig(namespace, "test-relconf", nil)
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(time.Millisecond * 500)
		Expect(releaseConfigs).To(HaveKey(releaseConfig.Name))
		Expect(releaseConfigs[releaseConfig.Name].Spec.ChartName).To(Equal(""))

		releaseConfig.Spec.ChartName = "test-chart"
		_, err = framework.UpdateReleaseConfig(releaseConfig)
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(time.Millisecond * 500)
		Expect(releaseConfigs).To(HaveKey(releaseConfig.Name))
		Expect(releaseConfigs[releaseConfig.Name].Spec.ChartName).To(Equal("test-chart"))

		err = framework.DeleteReleaseConfig(namespace, releaseConfig.Name)
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(time.Millisecond * 500)
		Expect(releaseConfigs).NotTo(HaveKey(releaseConfig.Name))

		By("list release configs")
		releaseConfig1, err := framework.CreateReleaseConfig(namespace, "test-relconf1", map[string]string{"test": "true"})
		Expect(err).NotTo(HaveOccurred())
		_, err = framework.CreateReleaseConfig(namespace, "test-relconf2", map[string]string{"test": "false"})
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(time.Millisecond * 500)

		releaseConfigLists, err := k8sCache.ListReleaseConfigs(namespace, "")
		Expect(err).NotTo(HaveOccurred())
		Expect(releaseConfigLists).To(HaveLen(2))

		releaseConfigLists, err = k8sCache.ListReleaseConfigs(namespace, "test=true")
		Expect(err).NotTo(HaveOccurred())
		Expect(releaseConfigLists).To(HaveLen(1))
		Expect(releaseConfigLists[0].Name).To(Equal(releaseConfig1.Name))
	})

	It("test pvc", func() {

		By("list pvcs")
		pvc1, err := framework.CreatePvc(namespace, "test-pvc1", map[string]string{"test": "true"})
		Expect(err).NotTo(HaveOccurred())
		_, err = framework.CreatePvc(namespace, "test-pvc2", map[string]string{"test": "false"})
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(time.Millisecond * 500)

		pvcs, err := k8sCache.ListPersistentVolumeClaims(namespace, "")
		Expect(err).NotTo(HaveOccurred())
		Expect(pvcs).To(HaveLen(2))

		pvcs, err = k8sCache.ListPersistentVolumeClaims(namespace, "test=true")
		Expect(err).NotTo(HaveOccurred())
		Expect(pvcs).To(HaveLen(1))
		Expect(pvcs[0].Name).To(Equal(pvc1.Name))
	})

	Describe("test tenant", func() {

		var (
			namespace2 string
		)

		BeforeEach(func() {
			By("create namespace2")
			namespace2, err = framework.CreateRandomNamespace("k8sCacheTenantTest", map[string]string{"test": "true"})
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			By("delete namespace2")
			err = framework.DeleteNamespace(namespace2)
			Expect(err).NotTo(HaveOccurred())
		})

		It("test get tenant & list tenants", func() {

			By("get tenant")
			_, err := framework.CreateResourceQuota(namespace, "test-rq")
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(time.Millisecond * 500)

			tenant, err := k8sCache.GetTenant(namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(tenant.TenantQuotas).To(HaveLen(1))
			Expect(tenant.UnifyUnitTenantQuotas).To(HaveLen(1))

			_, err = k8sCache.GetTenant("not-existed")
			Expect(err).To(Equal(errorModel.NotFoundError{}))

			By("list tenants")
			tenants, err := k8sCache.ListTenants("")
			Expect(err).NotTo(HaveOccurred())
			Expect(tenants).NotTo(BeNil())
			Expect(len(tenants.Items) >= 2).To(BeTrue())

			tenant1 := getTenantInfo(tenants.Items, namespace)
			Expect(tenant1).NotTo(BeNil())
			Expect(tenant1.TenantQuotas).To(HaveLen(1))
			Expect(tenant1.UnifyUnitTenantQuotas).To(HaveLen(1))

			tenant2 := getTenantInfo(tenants.Items, namespace2)
			Expect(tenant2).NotTo(BeNil())

			tenants, err = k8sCache.ListTenants("test=true")
			Expect(err).NotTo(HaveOccurred())
			Expect(tenants).NotTo(BeNil())
			Expect(len(tenants.Items) >= 1).To(BeTrue())

			tenant1 = getTenantInfo(tenants.Items, namespace)
			Expect(tenant1).To(BeNil())

			tenant2 = getTenantInfo(tenants.Items, namespace2)
			Expect(tenant2).NotTo(BeNil())
		})

	})

	It("test node", func() {

		By("list nodes")
		nodeList, err := framework.ListNodes()
		Expect(err).NotTo(HaveOccurred())
		nodes, err := k8sCache.GetNodes("")
		Expect(err).NotTo(HaveOccurred())
		Expect(nodes).To(HaveLen(len(nodeList.Items)))

		node, err := framework.GetTestNode()
		Expect(err).NotTo(HaveOccurred())

		err = framework.LabelNode(node.Name, map[string]string{"walm-list-nodes-test": "true"}, nil)
		Expect(err).NotTo(HaveOccurred())
		nodes, err = k8sCache.GetNodes("walm-list-nodes-test=true")
		Expect(err).NotTo(HaveOccurred())
		Expect(nodes).To(HaveLen(1))
	})

	It("test stateful sets", func() {

		By("list stateful sets")
		_, err := framework.CreateStatefulSet(namespace, "test-sts1", map[string]string{"test": "true"})
		Expect(err).NotTo(HaveOccurred())
		_, err = framework.CreateStatefulSet(namespace, "test-sts2", map[string]string{"test": "false"})
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(time.Millisecond * 500)

		stss, err := k8sCache.ListStatefulSets(namespace, "")
		Expect(err).NotTo(HaveOccurred())
		Expect(stss).To(HaveLen(2))

		stss, err = k8sCache.ListStatefulSets(namespace, "test=true")
		Expect(err).NotTo(HaveOccurred())
		Expect(stss).To(HaveLen(1))
	})

	It("test pod", func() {

		By("get pod events")
		_, err := framework.CreatePod(namespace, "test-pod")
		Expect(err).NotTo(HaveOccurred())

		err = framework.WaitPodRunning(namespace, "test-pod")
		Expect(err).NotTo(HaveOccurred())

		events, err := k8sCache.GetPodEventList(namespace, "test-pod")
		Expect(err).NotTo(HaveOccurred())
		Expect(len(events.Events) > 0).To(BeTrue())

		_, err = k8sCache.GetPodLogs(namespace, "test-pod", "", 0)
		Expect(err).NotTo(HaveOccurred())

	})

	It("test secrets", func() {

		By("list secrets")
		_, err := framework.CreateSecret(namespace, "test-sct1", map[string]string{"test": "true"})
		Expect(err).NotTo(HaveOccurred())
		_, err = framework.CreateSecret(namespace, "test-sct2", map[string]string{"test": "false"})
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(time.Millisecond * 500)

		scts, err := k8sCache.ListSecrets(namespace, "")
		Expect(err).NotTo(HaveOccurred())
		Expect(len(scts.Items) >= 2).To(BeTrue())

		scts, err = k8sCache.ListSecrets(namespace, "test=true")
		Expect(err).NotTo(HaveOccurred())
		Expect(scts.Items).To(HaveLen(1))
	})

	It("test eventLists", func() {
		By("list deployment eventList")
		_, err := framework.CreateDeployment(namespace, "test-deploy2")
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(time.Millisecond * 500)
		deploymentEventList, err := k8sCache.GetDeploymentEventList(namespace, "test-deploy2")
		Expect(err).NotTo(HaveOccurred())
		Expect(len(deploymentEventList.Events) != 0).To(BeTrue())

		By("list statefulset eventList")
		_, err = framework.CreateStatefulSet(namespace, "test-sts2", nil)
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(time.Millisecond * 500)
		statefulSetEventList, err := k8sCache.GetStatefulSetEventList(namespace, "test-sts2")
		Expect(err).NotTo(HaveOccurred())
		Expect(len(statefulSetEventList.Events) != 0).To(BeTrue())
	})

	It("test migs", func() {
		By("get node migration")
		_, err := framework.CreateMigration(namespace, "test-migration2", map[string]string{"migType": "node", "srcNode": "test-node2"})
		Expect(err).NotTo(HaveOccurred())
		_, err = framework.CreateMigration(namespace, "test-migration3", map[string]string{"migType": "node", "srcNode": "test-node2"})
		Expect(err).NotTo(HaveOccurred())
		_, err = framework.CreateMigration(namespace, "test-migration4", map[string]string{"migType": "node", "srcNode": "test-node3"})
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(time.Millisecond * 500)

		By("list migrations")
		migList, err := k8sCache.GetNodeMigration(namespace, "test-node2")
		Expect(err).NotTo(HaveOccurred())
		Expect(len(migList.Items) == 2).To(BeTrue())
		mig2List, err := k8sCache.ListMigrations(namespace, "migType=node")
		Expect(err).NotTo(HaveOccurred())
		Expect(len(mig2List) == 3).To(BeTrue())
	})
})

func getTenantInfo(items []*tenant.TenantInfo, name string) *tenant.TenantInfo {
	for _, item := range items {
		if item.TenantName == name {
			return item
		}
	}
	return nil
}
