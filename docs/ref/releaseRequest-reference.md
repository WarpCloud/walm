# 1.结构示意图

[结构图](http://naotu.baidu.com/file/f109591df91807624546d23de4f7c91e?token=f4ecee6d93f50070)

# 2.json

`releasePrettyParams`字段已弃用。

```
{
  "chartImage": "string",             // chart image url
  "chartName": "string",              // chart 名
  "chartVersion": "string",           // chart 版本
  "configValues": {},                 // 应用配置
  "dependencies": {},                 // 应用依赖
  "metaInfoParams": {                 // metaInfoParams 应用配置，优先级高于configValues，最后与
                                      // configValues merge
    "params": [                       // 参数配置
      {
        "name": "string",             // 参数名
        "type": "string",             // 参数类型
        "value": ``                   // 参数值， json raw string
      }
    ],
    "roles": [                        // 应用组件部分，一个或多个
      {
        "baseConfig": {               // 基本配置
          "env": [                    // 环境列表
            {
              "name": "string",
              "value": "string"
            }
          ],
          "image": "string",          // 镜像地址
          "others": [                 // 其他配置
            {
              "name": "string",
              "type": "string",
              "value": {}
            }
          ],
          "priority": 0,              // 优先级
          "replicas": 0,              // 副本数
          "useHostNetwork": true      // 使用节点网络
        },
        "name": "string",             // 应用组件名称
        "resources": {                // 资源配置
          "limitsCpu": 0,             // cpu 上限
          "limitsGpu": 0,             // gpu 上限
          "limitsMemory": 0,          // 内存上限
          "requestsCpu": 0,           // cpu 请求数
          "requestsGpu": 0,           
          "requestsMemory": 0,
          "storageResources": [       //存储资源配置
            {
              "name": "string",       // 资源名
              "value": {              
                "accessModes": [      // 访问模式
                                      // ReadWriteOnce——该卷可以被单个节点以读/写模式挂载
                                      // ReadOnlyMany——该卷可以被多个节点以只读模式挂载
                                      // ReadWriteMany——该卷可以被多个节点以读/写模式挂载        
                  "string"
                ],
                "size": 0,            // 分配资源大小
                "storageClass": "string".   // 为管理员提供了描述存储"class（类)"的方法
              }
            }
          ]
        }
      }
    ]
  },
  "name": "string",                   // 应用名
  "plugins": [                        // 要安装的插件列表(不能为已安装插件)
    {
      "args": "string",               // 参数名
      "disable": true,                
      "name": "string",           
      "version": "string" 
    }
  ],
  "releaseLabels": {},                // 应用标签
  "repoName": "string"                // 存放chart的仓库名
}
```


