# 目录

* [Deploying walm on google kubernetes engine](#Deploying walm on google kubernetes engine)
    * [1. Prerequisites](#1-Prerequisites)
    * [2. Understanding Yaml](#2-Understanding Yaml)
    * [3. Deploying Redis](#3-Deploying Redis)
    * [4. Deploying walm](#4-Deploying walm)

# Deploying walm on google kubernetes engine

## 1. Prerequisites

配置对多集群的访问， 获取到gke集群上下文， 设置本地 kubeconfig 环境变量， 通过本地的 kubectl 对gke集群进行访问。[官方参考](https://kubernetes.io/zh/docs/tasks/access-application-cluster/configure-access-multiple-clusters/)

## 2. Understanding Yaml

示例YAML文件中的某些行有破折号， 有些行没有， 由于 `apiVersion`的声明有时不止一个，  所以一个yaml文件内部会有多个对象， 它们之间用 `---`分开。如果想要深入了解 yaml结构， 可以参考 [blog from Mirantis](https://www.mirantis.com/blog/introduction-to-yaml-creating-a-kubernetes-deployment/) 。

## 3. Deploying Redis

```yaml
apiVersion: v1
kind: Service
metadata:
  name: walm-redis-master
spec:
  ports:
    - name: redis
      port: 6379
      protocol: TCP
      targetPort: redis
  clusterIP: None
  selector:
    app: redis
---
apiVersion: apps/v1beta2
kind: StatefulSet
metadata:
  name: walm-redis-master
spec:
  selector:
    matchLabels:
      app: redis
  serviceName: redis
  replicas: 1
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
        - name: redis
          image: redis:4.0.12
          imagePullPolicy: Always
          args: ["--requirepass", "$(REDIS_PASS)"]
          ports:
            - containerPort: 6379
              name: redis
          env:
          - name: REDIS_PASS
            value: "123456" # redis的连接密码
```

拷贝上面的 `YAML` 并命名为 `redis.yaml`， 然后运行 `kubectl -n kube-system create -f redis.yaml`在gke集群上创建 redis的 `Service`和`StatefulSet`。

运行 `kubectl -n kube-system get statefulsets`, `kubectl -n kube-system get service`,`kubectl -n kube-system get pods`来检查 redis服务的状态。

当redis 正常运行时， 

```shell
$ kubectl -n kube-system get statefulsets
NAME                READY     AGE
walm-redis-master   1/1       40h
$ kubectl -n kube-system get svc
NAME                   TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)          AGE
walm-redis-master      ClusterIP   10.7.246.210   <none>        6379/TCP         40h
$ kubectl -n kube-system get pods
NAME                    READY     STATUS    RESTARTS   AGE
walm-redis-master-0     1/1       Running   0          20h
```

对 redis 密码进行验证

```shell
$ kubectl -n kube-system exec -it walm-redis-master-0 /bin/bash
root@walm-redis-master-0:/data# redis-cli
127.0.0.1:6379> Auth 123456
OK
```



## 4. Deploying walm

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: walm
  name: walm
  namespace: kube-system
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  annotations:
    helm.sh/hook: crd-install
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"apiextensions.k8s.io/v1beta1","kind":"CustomResourceDefinition","metadata":{"annotations":{"helm.sh/hook":"crd-install"},"name":"releaseconfigs.apiextensions.transwarp.io"},"spec":{"group":"apiextensions.transwarp.io","names":{"kind":"ReleaseConfig","plural":"releaseconfigs","shortNames":["relconf"],"singular":"releaseconfig"},"scope":"Namespaced","version":"v1beta1"}}
  name: releaseconfigs.apiextensions.transwarp.io
spec:
  conversion:
    strategy: None
  group: apiextensions.transwarp.io
  names:
    kind: ReleaseConfig
    listKind: ReleaseConfigList
    plural: releaseconfigs
    shortNames:
    - relconf
    singular: releaseconfig
  scope: Namespaced
  version: v1beta1
  versions:
  - name: v1beta1
    served: true
    storage: true
status:
  acceptedNames:
    kind: ReleaseConfig
    listKind: ReleaseConfigList
    plural: releaseconfigs
    shortNames:
    - relconf
    singular: releaseconfig
  conditions:
  - lastTransitionTime: null
    message: no conflicts found
    reason: NoConflicts
    status: "True"
    type: NamesAccepted
  - lastTransitionTime: null
    message: the initial names have been accepted
    reason: InitialNamesAccepted
    status: "True"
    type: Established
  storedVersions:
  - v1beta1
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    app: walm
  name: walm
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: walm
  namespace: kube-system
---
apiVersion: v1
data:
  conf.yaml: |-
    debug: true
    jsonnetConfig:
      commonTemplateFilesPath: "/opt/ksonnet-lib"
    kubeConfig: {}
    repoList:
    - name: stable
      url: http://172.16.1.41:8882/stable/
    - name: transwarp
      url: https://harbor.kube-system.ingress.lan/api/chartrepo/transwarp/charts/
    serverConfig:
      port: 9001
      readTimeout: 0
      tls: false
      tlsCACert: ''
      tlsCert: ''
      tlsKey: ''
      tlsVerify: false
      writeTimeout: 0
    kafkaConfig:
      brokers: []
      enable: false
    redisConfig:
      addr: walm-redis-master.kube-system.svc:6379
      db: 0
      password: 123456
    taskConfig:
      broker: redis://123456@walm-redis-master.kube-system.svc:6379
      default_queue: machinery_tasks
      result_backend: redis://123456@walm-redis-master.kube-system.svc:6379
      results_expire_in:  360000
kind: ConfigMap
metadata:
  labels:
    app: walm
  name: walm-conf
  namespace: kube-system
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: walm
  name: walm
  namespace: kube-system
spec:
  ports:
  - name: port
    nodePort: 31607
    port: 9001
    protocol: TCP
    targetPort: 9001
  selector:
    app: walm
  sessionAffinity: None
  type: NodePort
status:
  loadBalancer: {}
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: walm
  name: walm-hl
  namespace: kube-system
spec:
  clusterIP: None
  ports:
  - name: port
    port: 9001
    protocol: TCP
    targetPort: 9001
  selector:
    app: walm
  sessionAffinity: None
  type: ClusterIP
status:
  loadBalancer: {}
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  labels:
    app: walm
  name: walm
  namespace: kube-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: walm
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
    type: RollingUpdate
  template:
    metadata:
      annotations:
        cni.networks: overlay
      creationTimestamp: null
      labels:
        app: walm
    spec:
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchLabels:
                app: walm
            namespaces:
            - kube-system
            topologyKey: kubernetes.io/hostname
      containers:
      - args:
        - walm
        - serv
        - --config
        - /etc/walm/conf.yaml
        env:
        - name: Pod_Name
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.name
        - name: Pod_Namespace
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
        image: docker.io/corndai1997/walm:dev
        imagePullPolicy: Always
        name: walm
        resources:
          limits:
            cpu: "1"
            memory: 1Gi
          requests:
            cpu: "0.5"
            memory: 1Gi
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /etc/walm
          name: walm-conf
      dnsPolicy: ClusterFirst
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      serviceAccount: walm
      serviceAccountName: walm
      terminationGracePeriodSeconds: 30
      tolerations:
      - key: master
        operator: Exists
      volumes:
      - configMap:
          defaultMode: 420
          items:
          - key: conf.yaml
            path: conf.yaml
          name: walm-conf
        name: walm-conf
```

拷贝上面的 `YAML` 并命名为 `walm.yaml`， 然后运行 `kubectl -n kube-system create -f walm.yaml`, 会在 gke集群上创建 walm的 `Service`,` ServiceAccount`,`Deployment`,` ConfigMap`,` ClusterRoleBinding`, `CustomResourceDefinition`。

:warning: 注意

- ConfigMap的redisConfig中， 对应的 addr填写格式为releaseName.namespace.svc.port (如walm-redis-master.kube-system.svc:6379)，password 同之前设置的redis密码。
- Deployment中的 requestCpu， requestMemory 根据实际集群资源进行配置

- 如果需要对walm自带的chart repo进行配置， 可以在 `ConfigMap`中对 `repoList`进行自定义配置。



walm 部署成功后，虽然pod状态属于running状态， 但目前还是无法访问的， 需要创建防火墙规则以允许TCP流量进入节点端口。[参考](https://cloud.google.com/kubernetes-engine/docs/how-to/exposing-apps?hl=zh-cn)

首先我们查看集群中节点的外部ip地址：

```shell
kubectl get nodes --output wide
NAME     STATUS    ROLES     AGE       VERSION          INTERNAL-IP   EXTERNAL-IP       OS-IMAGE    KERNEL-VERSION   CONTAINER-RUNTIME
gke-standard-cluster-1-default-pool-22417cff-08q3   Ready     <none>    2d3h      v1.13.6-gke.13   10.140.0.2    35.201.223.108    Container-Optimized OS from Google   4.14.127+        docker://18.9.3
gke-standard-cluster-1-default-pool-22417cff-482x   Ready     <none>    2d3h      v1.13.6-gke.13   10.140.0.3    104.199.225.254   Container-Optimized OS from Google   4.14.127+        docker://18.9.3
gke-standard-cluster-1-default-pool-22417cff-604h   Ready     <none>    2d3h      v1.13.6-gke.13   10.140.0.4    34.80.131.56      Container-Optimized OS from Google   4.14.127+        docker://18.9.3
```

通过`kubectl describe node `我们发现 `walm-56c4b49bb6-8994h`部署在 外部ip地址为 `35.201.223.108`的节点上。我们记下这个ip地址。

通过`kubectl -n kube-system get svc walm -o yaml` 我们可以查看到walm服务的 nodePort为 31607，该端口为我们访问walm服务的端口。

创建一条防火墙规则以允许 TCP 流量进入节点端口：

方法一：

确保你已安装[Cloud SDK](https://cloud.google.com/sdk/downloads?hl=zh-cn)。

打开[控制台](https://console.cloud.google.com)， 点击右上角的 cloud shell 按钮，连接到google cloud shell 机器。

```shell
gcloud compute firewall-rules create walm-node-port --allow tcp:31607
```

方法二：

打开 [防火墙设置页面](https://console.cloud.google.com/networking/firewalls/),点击**创建防火墙规则**

<div>
    <img src="assets/gke-firewall-settings.png">
</div>

浏览器打开 http://35.201.223.108:31607/swagger-ui/ 访问walm server。