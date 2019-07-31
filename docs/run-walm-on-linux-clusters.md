# 目录

* [walm部署](#walm部署)
    * [1. 使用kubeadm安装kubernetes 1.14](#1-使用kubeadm安装kubernetes 1.14)
    * [2. 安装helm](#2-安装helm)
    * [3. 安装chartmuseum](#3-安装chartmuseum)
    * [4. hostPath实现本地存储](#4-hostPath实现本地存储)
    * [5. redis部署](#5-redis部署)
    * [6. walm部署](#6-walm部署)

# walm部署

:warning: 常见问题汇总
- Docker拉取镜像失败
  启动Docker后, 对Docker进行如下配置
  ```shell script
  vi /usr/lib/systemd/system/docker.service
  # config registry-mirror && insecure-registry
  ExecStart=/usr/bin/dockerd -H fd:// --containerd=/run/containerd/containerd.sock --registry-mirror xxx1 --registry-mirror xxx2 --insecure-registry xxx1 --insecure-registry xxx2
  # restart
  systemctl daemon-reload
  systemctl restart docker
  ```
  
- 集群无法连接外网
  自定义DNS配置, 选择其他如Google, Tencent的 DNS服务
  ```shell script
  vi /etc/sysconfig/network-scripts/ifcfg-eth0
  # 最后两行添加
  DNS1=223.5.5.5
  DNS2=114.114.114.114
  # 重启
  systemctl restart network
  ```
  
- walm, redis chart 来源

​       chart的来源可以参考提供的[walm-chart](https://github.com/WarpCloud/walm-charts)。

## 1. 使用kubeadm安装kubernetes 1.14

**版本建议**

docker 18.06及以上

kubeadm 1.14.0

kubernates 1.14.0

**开源 registry 使用**

```
docker pull registry:2
docker run -d -p 5000:5000 --restart=always --name registry registry:2
```

[kubernetes安装参考](https://blog.frognew.com/2019/04/kubeadm-install-kubernetes-1.14.html)

## 2. 安装helm

由于开源的helm v3版本尚在开发，存在一些bug，目前可以使用我们自己定制的helm(可以跳过下面的 helm push 插件安装步骤)。[下载地址](https://github.com/WarpCloud/helm/releases/tag/v3.0%2Bunreleased)

## 3. 安装chartmuseum

安装`helm push`插件和`chartmuseum`

```shell
# install helm push
# on Linux
helm plugin install https://github.com/chartmuseum/helm-push

# install chartmuseum
# on Linux
curl -LO https://s3.amazonaws.com/chartmuseum/release/latest/bin/linux/amd64/chartmuseum
# enable cli
chmod a+x ./chartmuseum
mv ./chartmuseum /usr/local/bin

# start chartmuseum
chartmuseum --port=8089 \
  --storage="local" \
  --storage-local-rootdir="./chartstorage" >/tmp/chartmuseum.log 2>&1 &

# add chartRepo to helm
helm repo add chartmuseum https://my.chart.repo.com(eg: 127.0.0.1:8189)

# upload charts to repo chartmuseum， 可以上传 redis， walm的chart到chartmuseum上
helm push mychart-0.1.0.tgz chartmuseum       # push .tgz from "helm package"
helm push . chartmuseum                       # package and push chart directory
helm push . --version="7c4d121" chartmuseum   # override version in Chart.yaml
helm push . https://my.chart.repo.com         # push directly to chart repo URL
```

## 4. hostPath实现本地存储(也可采用其他方式如local-volume-pre)

创建 PersistentVolume和PersistentVolumeClaim

```yaml
kind: PersistentVolume
apiVersion: v1
metadata:
  name: redis-data-walm-redis-master-0
  labels:
    type: local
spec:
  storageClassName: hostpath
  capacity:
    storage: 8Gi
  accessModes:
    - ReadWriteOnce
  hostPath:
    path: "/tmp/redis"
---
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: redis-data-walm-redis-master-0
spec:
  storageClassName: stan
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 8Gi
```

## 5. redis部署

redis的chart可以从[walm-charts](https://github.com/WarpCloud/walm-charts)获取到， 修改 `redis/values.yaml`中 `persistence`的`storageClass`为`hostpath`, 创建 redis。

```shell
$ helm install -n namespace --name walm-redis ./redis
```

## 6. walm部署

- 从[walm-charts](https://github.com/WarpCloud/walm-charts)获取到walm的chart， 修改 walm/values.yaml ， 指定安装的`redis svc`。 

```yaml
# 1. 查询 svc name
kubectl -n kube-system get svc

# 2. 在 values.yaml 中对svc进行配置 host: svc-name.namespace.svc
redisConfig:
  ##
  ## Use the redis chart dependency.
  ## Set to false if bringing your own redis.
  enabled: true
  ##
  host: walm-redis-master.kube-system.svc      
  password: "123456"
  port: 6379
  db: 0
  default_queue: machinery_tasks
  results_expire_in: 360000
```

- 若需要自定义walm中chart的repo源， 可在 values.yaml中对 configmap中的repoList 进行配置。

```yaml
configmap:
  conf.yaml: |-
    .......
    - name: chartmuseum
      url: 之前配置的chartmuseum的地址
    // 如果需要使用外部的chartRepo源， 按如下方法进行配置
    - name: apphub
      url: https://apphub.aliyuncs.com
    ......
  
```

- 部署walm

```shell
$ helm install --name walm ./walm
```

- 环境变量设置

```shell
export HELM_DRIVER=configmap && helm ls
```

