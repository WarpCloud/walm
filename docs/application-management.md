# 目录

* [应用生命周期](#应用生命周期)
  * [1. 安装](#1-安装)
    * [2. 升级](#2-升级)
    * [3. 暂停](#3-暂停)
    * [4. 恢复](#4-恢复)
    * [5. 重启](#5-重启)
    * [6. 删除](#6-删除)
    * [7. 查询](#7-查询)


* [基于Redis的应用缓存](#基于Redis的应用缓存)

* [应用状态实时同步](#应用状态实时同步)
  
    

# 应用生命周期

:warning: 在进行应用的安装和升级操作时，需要提前了解下 `releaseRequest` 的定义。[Reference](ref/releaseRequest-reference.md)

## 1. 安装

walm的安装功能通过helm实现kubernetes应用的定制化安装。

**1.1 Chart形式**

- 根据chart in chart repo安装(同时需要repoName,chartName和chartVersion)

  **支持异步安装，async=true 表示异步, async=false 表示同步，timeoutSec默认为0, 表示60秒,如果安装时，timeoutSec设为负值或者值过小，会出现如下错误：**

  ```json
  {
    "errorCode": -1,
    "errMessage": "failed to install release: Timeout reached"
  }
  ```
  异步消息处理,线程启动后会继续执行下一个线程， 无需等待上一个线程的结果。

  同步消息处理，向helm 发送 处理 release 相关请求时， 首先向redis查询是否有相同的 namespace/releaseName 的task存在，若old task还没完成， 等待其finished或timeout 。

  ```shell script
  helm repo list                  #查看repo列表
  helm search repoName            #通过search进行关键字查询
  helm search repoName/chartName  
  helm search chartName              
  #具体可参照 helm 的使用方法
  ```
  
eg: 安装zookeeper应用（/api/v1/release/{namespace} )
  
```json
  {
    "chartName": "zookeeper",
    "chartVersion": "6.1.0",
    "repoName": "qa",
    "name": "zookeeper-xxx"
  }
  ```
  
- 根据chartImage 安装

  通过3.0及以上版本的helm来帮助通过chartImage来安装应用。

  eg:  

  ```shell script
  # 1. 查看已有的chart Image, 列举了各个chartImage的name,ref,name,size,digest,version,created等信息。
  helm chart list
  
  # 2. 从远程仓库下载chart， 将chart 存储在本地仓库以便后续使用
  helm chart pull ref
  
  # 3. 在本地仓库缓存中存储某个chart的备份
  # 对于某个应用产品的tgz包，可以先进行解压，然后将其拷贝到本地仓库缓存中。
  tar -zxvf zookeeper-6.1.0.tgz
  helm chart save /home/zookeeper registryURL/repoName/zookeeper:6.1.0
  
  # 4. 上传一个chart到远程仓库（该ref必须存在于本地仓库缓存中）
  helm chart push ref(ref即registryURL/repoName/zookeeper:6.1.0)
  ```

  

- 根据本地Chart Archive安装

  可以通过helm fetch url or repo/name of the chart获取到产品chart的压缩包，通过`/api/v1/release/{namespace}/withchart`接口进行安装。

  ```json
  {
    "namespace": "corndai",
    "release": "zookeeper-dzy",
    "chart": "zookeeper-6.1.0.tgz"
    "body": {}
  }
  ```


:warning:**优先级：chartArchive 高于 chartImage 高于 chart in chart repo**

如果用户在安装应用时同时采用了以上方式中的两种及两种以上， 按照具有最高优先级的配置进行安装。

eg：

```
#POST /api/v1/release/{namespace}/withchart 用本地chart安装一个release
//如果用户已经上传一个chart archive（chart的压缩包）， 同时在body中都定义了chartImage， chartName， repoName， chartVersion等字段， 示例如下

chart： zookeeper-6.1.0.tgz
body {
	"chartImage": "registryURL/repoName/zookeeper:6.0.0",
	"repoName": "qa",
	"chartVersion": "5.2.0",
	"chartName": "zookeeper"
}

结果实际安装的应用为6.1.0版本的zookeeper， 如果不上传 chart archive， 那么最后实际安装的是6.0.0版本的。
```

- 支持metainfoParams和configValues重写应用组件配置

​       在安装应用组件的过程中，可以通过设置metainfoParams以及configValues的值对chart元信息中应用组件配置进行重写。其中，configValues与原生chart结构类似，而metainfoParams类似于用户自定义的配置，属于一种扁平化结构。**可以同时设置configValues以及metainfoParams的值，相较来说，对于同一个属性的配置，metainfoParams具有更高的优先级。**

eg: 安装zookeeper并通过configValues重写其应用配置

```json
{
  "chartName": "zookeeper",
  "chartVersion": "6.1.0",
  "configValues": {"appConfig": {
      "zookeeper": {
         "priority": 1,
         "replicas": 3,
         "image": "zookeeper:transwarp-5.2",
         "env_list": [],
         "use_host_network": false,
         "resources": {
            "cpu_limit": 3,
            "cpu_request": 0.5,
            "memory_limit": 4Gi,
            "memory_request": 1Gi,
            "storage": {
              "data": {
                "storageClass": "silver",
                "size": "100Gi",
                "accessMode": "ReadWriteOnce",
                "limit": {}
              }
            }
         }
      }
},
  "dependencies": {},
  "name": "zoookeeper-zhiyang",
  "releaseLabels": {},
  "repoName": "qa"
}
```
eg: 安装zookeeper并通过metainfoParams重写其应用配置

```json
{
  "chartName": "zookeeper",
  "chartVersion": "6.1.0",
  "metaInfoParams": {
    "params": [],
    "roles": [
      {
        "baseConfig": {
          "env": [],
          "image": "registryURL/repoName/zookeeper:transwarp-5.2",
          "others": [
            {
              "name": "string",
              "type": "string",
              "value": {}
            }
          ],
          "priority": 0,
          "replicas": 3,
          "useHostNetwork": false
        },
        "name": "zookeeper",
        "resources": {
          "limitsCpu": 2,
          "limitsMemory": 4,
          "requestsCpu": 0.5,
          "requestsMemory": 1,
          "storageResources": [
            {
              "name": "data",
              "value": {
                "accessModes": [
                  "ReadWriteOnce"
                ],
                "size": 100,
                "storageClass": "silver"
              }
            }
          ]
        }
      }
    ]
  },
  "name": "zookeeper-test",
  "releaseLabels": {},
  "repoName": "stable"
}
```

- 模拟安装

​       模拟安装并不会创建资源， 而是验证是否能够正确安装应用组件，如果成功，会返回对应的配置文件。操作方式有三种，与上述三种通过chart安装的方式类似，除了调用的接口不同，可以通过调用接口`/api/v1/release/{namespace}/dryrun`或`/api/v1/release/{namespace}/dryrun/withchart/dryrun`, 对于一个可以正常安装的应用， 会返回该应用的 manifest 配置信息， 包括 `configMap`, `Service`, `StatefulSet`, `ReleaseConfig` 四部分相关的内容。

模拟计算安装一个应用需要多少资源

通过提前预测利用chart安装应用所需的资源量，便于对集群资源作统筹管理和规划。

```
模拟计算安装一个Release需要多少资源
模拟计算用本地chart安装一个Release需要多少资源
分别通过接口/api/v1/release/{namespace}/dryrun/withchart及/api/v1/release/{namespace}/dryrun/withchart/resources进行操作，返回数据如下：
{
  "deployments": null,
  "statefulSets": [
    {
      "replicas": 3,
      "name": "zookeeper-test-zookeeper",
      "podRequests": {
        "cpu": 0.5,
        "memory": 1024,
        "storage": [
          {
            "name": "zkdir",
            "type": "pvc",
            "storageClass": "silver",
            "size": 100
          }
        ]
      },
      "podLimits": {
        "cpu": 2,
        "memory": 4096,
        "storage": null
      }
    }
  ],
  "daemonSets": null,
  "jobs": null,
  "pvcs": null
}
```



**1.2 依赖应用配置注入**

​		依赖应用配置动态注入可以对应用的依赖组件及其配置进行动态更新，主要是为了解耦动态依赖管理，例如**kafka依赖zookeeper, 如果zookeeper的配置进行了更新， kafka会对其进行感知和识别**。

​		依赖注入通过配置`dependencies`字段来实现，格式为`<string, string>`的Map数组，存放的是`name`和`releaseName`的键值对。该信息原先的依赖可以从该组件chart的metainfo文件中获取到（参见项目`application-helmcharts`）。以下以设置kafka应用组件的依赖为例，参考其对应的[metainfo](https://github.com/WarpCloud/walm-charts/blob/master/jsonnetcharts/kafka/transwarp-meta/metainfo.yaml):

```
// 1. 根据metainfo中信息所示，kafka依赖的是zookeeper，若更新kafka依赖某个namespace下名为zookeeper-dzy的zookeeper组件(如果不加namespace， 默认依赖当前目录的对应名称的zookeeper组件)
dependencies{"zookeeper": "namespace/zookeeper-dzy"}

// 2. 如果某个组件本不存在chart的依赖中，那么添加该组件会进行merge

// 3. 如果dependencies字段中， 某个key的值为""(空字符串)，表示将会从原有依赖中删除与该key相匹配的依赖。
dependencies{"zookeeper": ""}
```

**1.3 标签支持**

​	在kubernates中，标签是作为一对key/value,被关联到对象上，被用来划分特定组的对象(如，所有被walm管理的应用组件)，标签可以在创建一个对象的时候直接给与，也可以在后期随时修改，每一个对象可以拥有多个标签，但是，key值必须是唯一的。我们可以通过配置releaseLabels属性来对标签进行设置。格式为`<string, string>`的Map数组， 当前的`releaseLabels`的值会和之前的值进行`merge`, 如果`releaseLabels`字段中， 某个key的值为""(空字符串)， 表示将会从原有标签集合中删除与该key相匹配的标签。

   **在支持labelSelector（标签选择器）的接口中进行标签过滤时， labelSelector支持符合一个或多个的标签查询，格式为 key1=value1，key2=value2，...**

标签示例：

| Key                          | Description                | Example          |
| ---------------------------- | -------------------------- | ---------------- |
| app.kubernetes.io/name       | 应用名                     | **mysql**        |
| app.kubernetes.io/instance   | 标识应用程序实例的唯一名称 | wordpress-abcxzy |
| app.kubernetes.io/version    | 应用程序的当前版本         | 5.7.21           |
| app.kubernetes.io/component  | 体系结构中的组件           | database         |
| app.kubernetes.io/part-of    | 所属的更高级别应用程序名   | **wordpress**    |
| app.kubernetes.io/managed-by | 用于管理应用程序操作的工具 | **walm**         |

使用：

```
//添加标签
"releaseLabels": {"app.kubernetes.io/name":"zookeeper", "app.kubernetes.io/version":"6.1.0"}
//修改标签“app.kubernetes.io/version
"releaseLabels": {"app.kubernetes.io/version":"6.2.0"}
//删除标签
"releaseLabels": {"app.kubernetes.io/version":""}
```



**1.4 支持插件操作**

插件的配置通过`plugins`数组来完成。当前的`plugins`的值会和之前的值进行`merge`, 如果某个插件的disable字段的值为`false`， 表示将会从原有标签集合中删除与该key相匹配的标签。安装插件时，对于已添加的插件，不允许再添加相同`Name`的插件。

eg:

```
// 添加LabelPod插件支持
"plugins": [
        {
          "name": "LabelPod",
          "args": "{"labelsToAdd":{"test1":"test2"}}",
          "version": "1.0",
          "disable": true
        }
      ],
// 删除LabelPod插件
"plugins": [
        {
          "name": "LabelPod",
          "args": "{"labelsToAdd":{"test1":"test2"}}",
          "version": "1.0",
          "disable": false
        }
      ],
```



## 2. 升级

对于一个已经安装成功的应用组件，我们可以对其进行版本升级和配置更新， 实现产品的更新迭代，提高产品的灵活性。

#### 版本升级

- 通过chart in chart Repo进行升级
- 通过chartImage进行升级
- 通过chart archive进行升级

eg:

```
//1.安装一个6.0.0版本的zookeeper, 应用名为zookeeper-demo
//2.通过chart in chart Repo进行升级成6.1.0版本(如果qa repo中存在一个zookeeper 6.1.0的chart)
PUT /api/v1/release/{namespace}
{
  "chartName": "zookeeper",
  "chartVersion": "6.1.0",
  "configValues": {},
  "dependencies": {},
  "name": "zoookeeper-demo",
  "releaseLabels": {},
  "repoName": "qa"
}
// 通过chartImage进行升级
PUT /api/v1/release/{namespace}
{
	"chartImage": "repoUrl/cy-charts/zookeeper:6.1.0",
	"name": "zookeeper-demo",
	"releaseLabels": {}
}
// 通过chart archive进行升级
PUT /api/v1/release/{namespace}/withchart
通过上传对应的6.1.0版本的zookeepr的压缩包进行升级
```

#### 配置更新

在进行版本升级的时候， 可以对应用的配置进行更新，在对应用进行更新的时候， 包含对应用 releaseLabels， dependencies, plugins的重用， 即对于已添加的标签，依赖， 插件不需要再重复添加， 重复添加时会对旧配置进行重写，添加新配置会对其他旧配置进行重用。具体如下：

```
应用升级前配置
body
{
  "releaseLabels": {
	  "app.kubernetes.io/name":"zookeeper", 
	  "app.kubernetes.io/version":"6.1.0",
	  "app.kubernetes.io/managed-by":"walm", 
  },
  "plugins": [
        {
          "name": "nodeAffinity",
          "args": "",
          "version": "1",
          "disable": false
        },
        {
          "name": "guardian",
          "args": "",
          "version": "1",
          "disable": false
        }
      ],
  dependencies{"search": "namespace1/search", "zookeeper": "namespace/zookeeper"}
}

对应用升级时，如果定义body如下
{
  "releaseLabels": {
	  "app.kubernetes.io/version":"6.2.0",
	  "app.kubernetes.io/managed-by":""
  },
  "plugins": [
        {
          "name": "guardian",
          "args": "",
          "version": "1",
          "disable": true
        },
        {
          "name": "ingress",
          "args": "",
          "version": "1",
          "disable": false
        }
      ],
  dependencies{"search": "namespace2/search"}
}

最后应用升级后， 对应配置如下
body
{
  "releaseLabels": {
	  "app.kubernetes.io/name":"zookeeper", 
	  "app.kubernetes.io/version":"6.2.0"
  },
  "plugins": [
        {
          "name": "nodeAffinity",
          "args": "",
          "version": "1",
          "disable": false
        },
        {
          "name": "ingress",
          "args": "",
          "version": "1",
          "disable": false
        }
      ],
  dependencies{"search": "namespace2/search", "zookeeper": "namespace/zookeeper"}
}

需要注意的是， 如果releaseLabels的某个标签的value为 ""，该标签会被删除；dependencies同理。
如果plugins中某个插件的disable属性为true， 表示该插件也会被删除。
```

除了上述，在对应用进行升级时, 可以通过设置`configValues`或`metainfoParams`的值来对应用进行配置，如果自定义metainfoParams, 自定义metainfoParams中的params和roles中的两部分值会和原应用中对应的配置信息，即chartMetainfo中的params和roles进行合并,得到metaInfoConfigs，最后与自定义configValues进行合并，将新的配置进行应用。具体如下：

```
若原应用配置如下
"configValues": {
 		"appConfig": {
      "zookeeper": {
         "priority": 1,
         "replicas": 3,
         "image": "zookeeper:transwarp-5.2",
         "env_list": [],
         "use_host_network": false,
         "resources": {
            "cpu_limit": 3,
            "cpu_request": 0.5,
            "memory_limit": 4Gi,
            "memory_request": 1Gi,
            "storage": {
              "data": {
                "storageClass": "silver",
                "size": "100Gi",
                "accessMode": "ReadWriteOnce",
                "limit": {}
              }
            }
         }
     }
   }
},
应用升级时自定义metainfoParams和configValues
"configValues": {
 		"appConfig": {
      "zookeeper": {
      		"replicas": 2
      }
    }
}
"metaInfoParams": {
    "roles": [
      {
        "name": "zookeeper",
        "resources": {
          "limitsCpu": 4,
          "limitsMemory": 4,
          "requestsCpu": 1,
          "requestsMemory": 2,
        }
      }
    ]
},

发送请求后configValues最终结果如下， 并对应用配置进行更新
"configValues": {
 		"appConfig": {
      "zookeeper": {
         "priority": 2,
         "replicas": 3,
         "image": "zookeeper:transwarp-5.2
         ",
         "env_list": [],
         "use_host_network": false,
         "resources": {
            "cpu_limit": 4,
            "cpu_request": 1,
            "memory_limit": 4Gi,
            "memory_request": 2Gi,
            "storage": {
              "data": {
                "storageClass": "silver",
                "size": "100Gi",
                "accessMode": "ReadWriteOnce",
                "limit": {}
              }
            }
         }
     }
   }
},
```



## 3. 暂停

应用变更时，提供滚动升级策略，失败自动暂停。 该功能通过更新Deployment来改变 pod 和 Replica Set的状态来实现。暂停会删除pod， 但并不会删除应用的相关资源如pv… 暂停后的服务升级后依然保持暂停的状态

通过接口`/api/v1/release/{namespace}/name/{release}/pause`来暂停某个应用服务。

## 4. 恢复

服务状态恢复。对于之前服务被暂停的应用组件，其相关资源并没有被删除，可以通过对应的服务恢复接口`/api/v1/release/{namespace}/name/{release}/recover`进行恢复, 根据应用的Deployment进行弹缩创建出新的pods

## 5. 重启

如果release状态异常或者需要进行更新等其他操作，可以通过`/api/v1/release/{namespace}/name/{release}/restart`来重启（除了job）所有pod。 damenset， sts, deployment

## 6. 删除

通过`/api/v1/release/{namespace}/name/{release}`接口进行单个release的删除。

删除应用时， 支持同步或异步删除，删除release管理的statefulSet关联的所有pvc。

eg:

```
1. 创建一个zookeeper应用
2. 通过接口删除对应应用， 选择 deletePvcs=false
3. 验证查看， kubectl -n namespace get pvc 发现pvc依然存在
NAME                             STATUS    VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
zkdir-zookeeper-pt-zookeeper-0   Bound     pvc-24764968-8bf8-11e9-837f-0cc47ae29dba   100Gi      RWO            silver         14m
zkdir-zookeeper-pt-zookeeper-1   Bound     pvc-24780c38-8bf8-11e9-837f-0cc47ae29dba   100Gi      RWO            silver         14m
zkdir-zookeeper-pt-zookeeper-2   Bound     pvc-2479e6b0-8bf8-11e9-837f-0cc47ae29dba   100Gi      RWO            silver         14m

4. 重新创建一个应用，删除应用时选择 deletePvcs=true， 资源都已全部删除。
```



## 7. 查询

- 获取所有的release列表

```
/api/v1/release
```

`Tips:` 对于 release列表项太多的情况，可以通过`labelselector`进行筛选，可以通过release对应的标签进行查询，release的标签可以通过查看release信息中的`releaseLabels`属性获取（调用获取对应release的详细信息接口），查询格式configA=valueA,configB=valueB……

eg：查看标签属性app.kubernetes.io/managed-by"值为"walm”的release，则`labelselector`所填字段的值应为

app.kubernates.io/managed-by=walm

- 获取Namespace下的所有release列表

```
/api/v1/release/{namespace}
```

- 获取对应release的详细信息

```
/api/v1/release/{namespace}/name/{release}
```

[查询结果分类和解析](ref/queryResults-description.md)

# 基于Redis的应用缓存

walm中release以及project的相关查询都是基于redis缓存实现，通过walm对应用进行创建和升级时，都会经过向redis写入数据的过程。同时， redis会定时从helm和k8s获取数据，完成数据同步， 通过 helm 创建的资源最多5分钟后，会与redis的数据进行匹配， redis中少的数据会被补上，多的数据会被删除。

# 应用状态实时同步

应用会定时向kafka组件发送自己的实时配置，通过创建消费者去消费kafka来获取应用的实时状态。 kafka 命令订阅 topic 获取信息。


