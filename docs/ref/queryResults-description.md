## 查询结果分类和解析

以下以kafka组件信息查询结果为例， 包含丰富的信息，如statefulset的状态, deployments的信息， pod的状态

```json
    {
      "name": "ka2",                               // 应用名
      "repoName": "qa",                            // 仓库名
      "configValues": {},                          // 映射到chart的配置属性
      "version": 4,                                // 应用版本
      "namespace": "helmv3",											 // 命名空间
      "dependencies": {                            // 依赖了一个 zookeeper 组件
        "zookeeper": "zk2"
      },
      "chartName": "kafka",                        // chart 名
      "chartVersion": "6.1.0",                     // chart 版本
      "chartAppVersion": "6.1",                    // chart 应用版本
      "helmExtraLabels": null,
      "ready": true,
      "message": "",
      "releaseStatus": {
        "services": [
          {
            "name": "ka2-kafka-hl",
            "namespace": "helmv3",
            "kind": "Service",
            "state": {
              "status": "Ready",
              "reason": "",
              "message": ""
            },
            "ports": [
              {
                "name": "web",
                "protocol": "TCP",
                "port": 9092,
                "targetPort": "9092",
                "nodePort": 0,
                "endpoints": [
                  "10.16.11.179:9092",
                  "10.16.80.61:9092",
                  "10.16.94.251:9092"
                ]
              }
            ],
            "clusterIp": "None",
            "serviceType": "ClusterIP"
          },
          {
            "name": "ka2-kafka",
            "namespace": "helmv3",
            "kind": "Service",
            "state": {
              "status": "Ready",
              "reason": "",
              "message": ""
            },
            "ports": [
              {
                "name": "web",
                "protocol": "TCP",
                "port": 9092,
                "targetPort": "9092",
                "nodePort": 30704,
                "endpoints": [
                  "10.16.11.179:9092",
                  "10.16.80.61:9092",
                  "10.16.94.251:9092"
                ]
              }
            ],
            "clusterIp": "10.10.168.9",
            "serviceType": "NodePort"
          }
        ],
        "configmaps": [
          {
            "name": "ka2-kafka-confd-conf",
            "namespace": "helmv3",
            "kind": "ConfigMap",
            "state": {
              "status": "Ready",
              "reason": "",
              "message": ""
            },
            "data": {
              "consumer.properties.tmpl": "",
              "jaas.conf.tmpl": "",
              "kafka-confd.conf": "",
              "kafka-env.sh.tmpl": "",
              "kafka.toml": "",
              "producer.properties.tmpl": "",
              "server.properties.tmpl": "",
              "tdh-env.sh.tmpl": ""
            }
          },
          {
            "name": "ka2-kafka-entrypoint",
            "namespace": "helmv3",
            "kind": "ConfigMap",
            "state": {
              "status": "Ready",
              "reason": "",
              "message": ""
            },
            "data": {
              "entrypoint.sh": ""
            }
          }
        ],
        "daemonsets": [],
        "deployments": [],
        "ingresses": [],
        "jobs": [],
        "secrets": [],
        "statefulsets": [
          {
            "name": "ka2-kafka",
            "namespace": "helmv3",
            "kind": "StatefulSet",
            "state": {
              "status": "Ready",
              "reason": "",
              "message": ""
            },
            "labels": {
              "app.kubernetes.io/component": "kafka",
              "app.kubernetes.io/instance": "ka2",
              "app.kubernetes.io/managed-by": "walm",
              "app.kubernetes.io/name": "kafka",
              "app.kubernetes.io/part-of": "kafka",
              "app.kubernetes.io/version": "6.1"
            },
            "annotations": null,
            "expectedReplicas": 3,
            "readyReplicas": 3,
            "currentVersion": "ka2-kafka-6ffd77fdb6",
            "updateVersion": "ka2-kafka-6ffd77fdb6",
            "pods": [
              {
                "name": "ka2-kafka-0",
                "namespace": "helmv3",
                "kind": "Pod",
                "state": {
                  "status": "Ready",
                  "reason": "",
                  "message": ""
                },
                "labels": {
                  "app.kubernetes.io/instance": "ka2",
                  "app.kubernetes.io/name": "kafka",
                  "app.kubernetes.io/version": "6.1",
                  "controller-revision-hash": "ka2-kafka-6ffd77fdb6",
                  "statefulset.kubernetes.io/pod-name": "ka2-kafka-0"
                },
                "annotations": {
                  "k8s.v1.cni.cncf.io/networks-status": "[{\n    \"name\": \"\",\n    \"ips\": [\n        \"10.16.80.61\"\n    ],\n    \"default\": true,\n    \"dns\": {}\n}]",
                  "tos.network.staticIP": "true",
                  "tosdisksubpool": "{\"log\":\"silver1\"}",
                  "transwarp.replicaid": "0",
                  "transwarp/configmap.md5": "540790c5030838d4d4c6dc3679a52932"
                },
                "hostIp": "172.26.0.5",
                "podIp": "10.16.80.61",
                "containers": [
                  {
                    "name": "kafka",
                    "image": "172.16.1.99/gold/kafka:transwarp-6.0",
                    "ready": true,
                    "restartCount": 0,
                    "state": {
                      "status": "Running",
                      "reason": "",
                      "message": ""
                    }
                  }
                ],
                "age": "22h"
              },
              {
                "name": "ka2-kafka-2",
                "namespace": "helmv3",
                "kind": "Pod",
                "state": {
                  "status": "Ready",
                  "reason": "",
                  "message": ""
                },
                "labels": {
                  "app.kubernetes.io/instance": "ka2",
                  "app.kubernetes.io/name": "kafka",
                  "app.kubernetes.io/version": "6.1",
                  "controller-revision-hash": "ka2-kafka-6ffd77fdb6",
                  "statefulset.kubernetes.io/pod-name": "ka2-kafka-2"
                },
                "annotations": {
                  "k8s.v1.cni.cncf.io/networks-status": "[{\n    \"name\": \"\",\n    \"ips\": [\n        \"10.16.11.179\"\n    ],\n    \"default\": true,\n    \"dns\": {}\n}]",
                  "tos.network.staticIP": "true",
                  "tosdisksubpool": "{\"log\":\"silver1\"}",
                  "transwarp.replicaid": "2",
                  "transwarp/configmap.md5": "540790c5030838d4d4c6dc3679a52932"
                },
                "hostIp": "172.26.0.6",
                "podIp": "10.16.11.179",
                "containers": [
                  {
                    "name": "kafka",
                    "image": "172.16.1.99/gold/kafka:transwarp-6.0",
                    "ready": true,
                    "restartCount": 1,
                    "state": {
                      "status": "Running",
                      "reason": "",
                      "message": ""
                    }
                  }
                ],
                "age": "26d"
              },
              {
                "name": "ka2-kafka-1",
                "namespace": "helmv3",
                "kind": "Pod",
                "state": {
                  "status": "Ready",
                  "reason": "",
                  "message": ""
                },
                "labels": {
                  "app.kubernetes.io/instance": "ka2",
                  "app.kubernetes.io/name": "kafka",
                  "app.kubernetes.io/version": "6.1",
                  "controller-revision-hash": "ka2-kafka-6ffd77fdb6",
                  "statefulset.kubernetes.io/pod-name": "ka2-kafka-1"
                },
                "annotations": {
                  "k8s.v1.cni.cncf.io/networks-status": "[{\n    \"name\": \"\",\n    \"ips\": [\n        \"10.16.94.251\"\n    ],\n    \"default\": true,\n    \"dns\": {}\n}]",
                  "tos.network.staticIP": "true",
                  "tosdisksubpool": "{\"log\":\"silver1\"}",
                  "transwarp.replicaid": "1",
                  "transwarp/configmap.md5": "540790c5030838d4d4c6dc3679a52932"
                },
                "hostIp": "172.26.0.8",
                "podIp": "10.16.94.251",
                "containers": [
                  {
                    "name": "kafka",
                    "image": "172.16.1.99/gold/kafka:transwarp-6.0",
                    "ready": true,
                    "restartCount": 0,
                    "state": {
                      "status": "Running",
                      "reason": "",
                      "message": ""
                    }
                  }
                ],
                "age": "15d"
              }
            ]
          }
        ]
      },
      "dependenciesConfigValues": {
        "ZOOKEEPER_CLIENT_CONFIG": {
          "zookeeper_addresses": "zk2-zookeeper-0.zk2-zookeeper-hl.helmv3.svc,zk2-zookeeper-1.zk2-zookeeper-hl.helmv3.svc,zk2-zookeeper-2.zk2-zookeeper-hl.helmv3.svc",
          "zookeeper_auth_type": "none",
          "zookeeper_port": "2181"
        }
      },
      "computedValues": {
        "ZOOKEEPER_CLIENT_CONFIG": {
          "zookeeper_addresses": "zk2-zookeeper-0.zk2-zookeeper-hl.helmv3.svc,zk2-zookeeper-1.zk2-zookeeper-hl.helmv3.svc,zk2-zookeeper-2.zk2-zookeeper-hl.helmv3.svc",
          "zookeeper_auth_type": "none",
          "zookeeper_port": "2181"
        },
        "advanceConfig": {
          "kafka": {},
          "server_properties": {
            "default.replication.factor": 2,
            "log.dirs": "/data",
            "log.flush.interval.messages": 10000,
            "log.flush.interval.ms": 1000,
            "log.retention.bytes": 1073741824,
            "log.retention.check.interval.ms": 300000,
            "log.retention.hours": 6,
            "message.max.bytes": 100000000,
            "num.io.threads": 8,
            "num.network.threads": 3,
            "num.partitions": 3,
            "num.recovery.threads.per.data.dir": 1,
            "replica.fetch.max.bytes": 100000000,
            "socket.receive.buffer.bytes": 102400,
            "socket.request.max.bytes": 104857600,
            "socket.send.buffer.bytes": 102400,
            "zookeeper.connection.timeout.ms": 6000
          }
        },
        "appConfig": {
          "kafka": {
            "env_list": [],
            "image": "172.16.1.99/gold/kafka:transwarp-6.0",
            "priority": 0,
            "replicas": 3,
            "resources": {
              "cpu_limit": 2,
              "cpu_request": 0.5,
              "memory_limit": "4Gi",
              "memory_request": "1Gi",
              "storage": {
                "data": {
                  "accessMode": "ReadWriteOnce",
                  "limit": {},
                  "size": "100Gi",
                  "storageClass": "silver"
                },
                "log": {
                  "accessMode": "ReadWriteOnce",
                  "limit": {},
                  "size": "100Gi",
                  "storageClass": "silver"
                }
              }
            },
            "use_host_network": false
          }
        },
        "transwarpConfig": {
          "transwarpApplicationPause": false,
          "transwarpCniNetwork": "overlay",
          "transwarpGlobalIngress": {
            "httpPort": 80,
            "httpsPort": 443
          },
          "transwarpLicenseAddress": "",
          "transwarpMetrics": {
            "enable": true
          }
        }
      },
      "outputConfigValues": {},
      "releaseLabels": {
        "auto-gen": "true"
      },
      "plugins": [],
      "metaInfoValues": {
        "params": null,
        "roles": [
          {
            "name": "kafka",
            "baseConfig": {
              "image": "172.16.1.99/gold/kafka:transwarp-6.0",
              "priority": null,
              "replicas": 3,
              "env": null,
              "useHostNetwork": null,
              "others": null
            },
            "resources": {
              "limitsMemory": 4096,
              "limitsCpu": 2,
              "limitsGpu": 0,
              "requestsMemory": 1024,
              "requestsCpu": 0,
              "requestsGpu": 0,
              "storageResources": [
                {
                  "name": "data",
                  "value": {
                    "accessModes": null,
                    "storageClass": "silver",
                    "size": 100
                  }
                },
                {
                  "name": "log",
                  "value": {
                    "accessModes": null,
                    "storageClass": "silver",
                    "size": 100
                  }
                }
              ]
            }
          }
        ]
      },
      "paused": false,
      "chartImage": ""
    }
```

为了更加直观且层次性地了解返回的应用相关信息， 我们对其进行归类介绍

#### 配置相关

- dependenciesConfigValues（依赖的配置）

```
"dependenciesConfigValues": {
     "ZOOKEEPER_CLIENT_CONFIG": {         // zookeeper应用配置                
        "zookeeper_addresses": "zk2-zookeeper-0.zk2-zookeeper-hl.helmv3.svc,
        zk2-zookeeper-1.zk2-zookeeper-hl.helmv3.svc,zk2-zookeeper-2.zk2-zookeeper-           hl.helmv3.svc",                           // zookeeper服务地址 
        "zookeeper_auth_type": "none",   // zookeeper验证方式
        "zookeeper_port": "2181"         // 暴露端口
      }
},
```

- ComputedValues（用来渲染chart模版的配置）

```
"computedValues": {
        "ZOOKEEPER_CLIENT_CONFIG": {
          "zookeeper_addresses": "zk2-zookeeper-0.zk2-zookeeper-hl.helmv3.svc,zk2-zookeeper-1.zk2-zookeeper-hl.helmv3.svc,zk2-zookeeper-2.zk2-zookeeper-hl.helmv3.svc",
          "zookeeper_auth_type": "none",
          "zookeeper_port": "2181"
        },
        "advanceConfig": {                                     // 高级配置
          "kafka": {},                                        
          "server_properties": {                              
            "default.replication.factor": 2,
            "log.dirs": "/data",
            "log.flush.interval.messages": 10000,
            "log.flush.interval.ms": 1000,
            "log.retention.bytes": 1073741824,
            "log.retention.check.interval.ms": 300000,
            "log.retention.hours": 6,
            "message.max.bytes": 100000000,
            "num.io.threads": 8,
            "num.network.threads": 3,
            "num.partitions": 3,
            "num.recovery.threads.per.data.dir": 1,
            "replica.fetch.max.bytes": 100000000,
            "socket.receive.buffer.bytes": 102400,
            "socket.request.max.bytes": 104857600,
            "socket.send.buffer.bytes": 102400,
            "zookeeper.connection.timeout.ms": 6000
          }
        },
        "appConfig": {                                          // 应用配置
          "kafka": {
            "env_list": [],
            "image": "172.16.1.99/gold/kafka:transwarp-6.0",
            "priority": 0,
            "replicas": 3,
            "resources": {
              "cpu_limit": 2,
              "cpu_request": 0.5,
              "memory_limit": "4Gi",
              "memory_request": "1Gi",
              "storage": {
                "data": {
                  "accessMode": "ReadWriteOnce",
                  "limit": {},
                  "size": "100Gi",
                  "storageClass": "silver"
                },
                "log": {
                  "accessMode": "ReadWriteOnce",
                  "limit": {},
                  "size": "100Gi",
                  "storageClass": "silver"
                }
              }
            },
            "use_host_network": false
          }
        },
        "transwarpConfig": {                                   // 星环内部产品配置
          "transwarpApplicationPause": false,
          "transwarpCniNetwork": "overlay",
          "transwarpGlobalIngress": {
            "httpPort": 80,
            "httpsPort": 443
          },
          "transwarpLicenseAddress": "",
          "transwarpMetrics": {
            "enable": true
          }
        }
      }
```

- MetainfoValues（metainfo值）

```
"metaInfoValues": {                  
        "params": null,
        "roles": [
          {
            "name": "kafka",
            "baseConfig": {
              "image": "172.16.1.99/gold/kafka:transwarp-6.0",
              "priority": null,
              "replicas": 3,
              "env": null,
              "useHostNetwork": null,
              "others": null
            },
            "resources": {
              "limitsMemory": 4096,
              "limitsCpu": 2,
              "limitsGpu": 0,
              "requestsMemory": 1024,
              "requestsCpu": 0,
              "requestsGpu": 0,
              "storageResources": [
                {
                  "name": "data",
                  "value": {
                    "accessModes": null,
                    "storageClass": "silver",
                    "size": 100
                  }
                },
                {
                  "name": "log",
                  "value": {
                    "accessModes": null,
                    "storageClass": "silver",
                    "size": 100
                  }
                }
              ]
            }
          }
        ]
      }
```

- outputConfigValues（此处以zookeeper应用的输出配置为例）

`outputConfigValues`结构与`dependenciesConfigValues`结构类似，区别在于， `OutputConfigValues`属于`zookeeper`应用当作依赖暴露给其他应用（`kafka`）的配置, 因而 `outputConfigValues`里不需要`ZOOKEEPER_CLIENT_CONFIG`,属性直接暴露给其他应用。 而在`dependenciesConfigValues`中， 需要指定某一个依赖的具体配置，所以需要`ZOOKEEPER_CLIENT_CONFIG`。

```
"outputConfigValues": {
     "zookeeper_addresses": "zookeeper-zhiyang-zookeeper-0.zookeeper-zhiyang-zookeeper-hl.zhiyang.svc,zookeeper-zhiyang-zookeeper-1.zookeeper-zhiyang-zookeeper-hl.zhiyang.svc,zookeeper-zhiyang-zookeeper-2.zookeeper-zhiyang-zookeeper-hl.zhiyang.svc",
     "zookeeper_auth_type": "none",
     "zookeeper_port": "2181"
 }
```

#### chart信息相关

```
"repoName": "qa",											 // 应用仓库名
"version": 4,                          //应用版本
"chartVersion": "6.1.0",               // chart 版本
"chartName": "kafka",                  // chart 名称
"chartAppVersion": "6.1"               // jsonnet 应用版本
```

#### 应用组件相关

```
"name": "ka2",                         // 应用名
"configValues": {},                    //额外加入到chart的配置值     
"namespace": "zhiyang",                //应用所属命名空间      
"dependencies": {"zookeeper": "zk2"},  //应用依赖
"releaseLabels": {                     //应用标签 Map <String,String>
   "auto-gen": "true"
},
"plugins": [                           //插件列表
      {
        "name": "guardian",            //插件名
        "args": "",                    //插件参数
        "version": "1",                //插件版本
        "disable": true                //disable=true 删除插件
      },
      {
        "name": "ingress",
        "args": "",
        "version": "1",
        "disable": false               //disable=false，启用插件
      }
    ],
    
"paused": false,                     // 是否处于暂停状态

```


