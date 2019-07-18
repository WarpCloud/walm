# 目录

* [k8s资源管理](#k8s资源管理)
  * [1. pod管理](#1-pod管理)
  * [2. pvc管理](#2-pvc管理)
  * [3. secret管理](#3-secret管理)
  * [4. storageClass](#4-storageClass)
  * [5. 服务状态检查](#5-服务状态检查)
  * [6. 租户管理](#6-租户管理)

# k8s资源管理

## 1. pod管理

用户可以通过接口获取pod对应的事件， 日志， 以及对pod进行重启等操作。

- 通过 `/api/v1/pod/{namespace}/name/{pod}/events`接口获取pod的相关事件

事件信息可用于排错。 例如当前Pod状态为Pending，可通过查看Event事件确认原因 ，一般原因是由于***开启了资源配额管理并且当前Pod的目标节点上恰好没有可用的资源*** 或者 ***正在下载镜像（镜像拉取耗时太久）或镜像下载失败***。时间信息包含：

`type`: 事件类型

`reason`: 事件原因

`message`: 事件信息

`from`: 使用的调度器

`firstTimestamp`: 事件第一次发生时间

`lastTimestamp`： 事件最近一次发生时间

```json
{
  "events": [
    {
      "type": "Warning",
      "reason": "FailedScheduling",
      "message": "pod has unbound immediate PersistentVolumeClaims (repeated 3 times)",
      "from": "default-scheduler",
      "count": 2548,
      "firstTimestamp": "2019-07-05T04:17:59Z",
      "lastTimestamp": "2019-07-07T19:58:04Z"
    }
  ]
}

```

- 通过` /api/v1/pod/{namespace}/name/{pod}/logs`接口获取pod的详细日志

  对pod的日志进行查询时， 除了必须指定的`namespace`和`pod` 名称外， 还可以选择性的设置`container`及`tail`两个字段

  `container`： 容器名， 指定查看特定container的日志

  `tail`: 指定查看日志的最后几行

- 通过`/api/v1/pod/{namespace}/name/{pod}/restart`接口对pod进行重启

## 2. pvc管理

- 获取Namespace下的pvc列表， 支持 `labelselector`节点标签过滤

​      `Get /api/v1/pvc/{namespace}`

- 删除Namespace下满足`labelselector`的Pvc列表

​      `Delete /api/v1/pvc/{namespace}`

- 根据Namespace和pvcname删除一个pvc

​      `Delete /api/v1/pvc/{namespace}/name/{pvcname}`

- 获取pvc的详细信息(包含pvc的名称，租户，种类，状态，标签， 存储类型， 卷名，容量，访问方式，卷模式)

​      `Get /api/v1/pvc/{namespace}/name/{pvcname}`

```json
{
  "name": "redis-data-walm-redis-master-0",
  "namespace": "kube-system",
  "kind": "PersistentVolumeClaim",
  "state": {
    "status": "Bound",
    "reason": "",
    "message": ""
  },
  "labels": {
    "app": "redis",
    "release": "walm-redis",
    "role": "master"
  },
  "storageClass": "hostpath",
  "volumeName": "redis-data-walm-redis-master-0",
  "capacity": "8Gi",
  "accessModes": [
    "ReadWriteOnce"
  ],
  "volumeMode": "Filesystem"
}

```
## 3. secret管理

- 获取Namepace下的所有Secret列表

​      `Get /api/v1/secret/{namespace}`

- 创建一个Secret

​      `Post /api/v1/secret/{namespace}`

- 更新一个Secret

​      `Put /api/v1/secret/{namespace}`

- 删除一个Secret

​      `Delete /api/v1/secret/{namespace}`

- 获取对应Secret的详细信息

​      `Get /api/v1/secret/{namespace}/name/{secretname}`

## 4. storageClass

- 获取StorageClass列表

​       `Get /api/v1/storageclass`

- 获取StorageClass详细信息

​       `Get /api/v1/storageclass/{name}`

## 5. 服务状态检查

- 服务Live状态检查

​       `Get /liveniess`

- 服务Ready状态检查

​      `Get /readiness`

- 获取服务Stats

​       `Get /stats`

## 6. 租户管理

- 获取租户列表

​      `Get /api/v1/tenant`

- 获取租户状态

​      `Get /api/v1/tenant/{tenantName}`

- 创建租户

​      `Post /api/v1/tenant/{tenantName}`

- 更新租户信息

​      `Put /api/v1/tenant/{tenantName}`

- 删除租户

​     `Delete/api/v1/tenant/{tenantName}`


