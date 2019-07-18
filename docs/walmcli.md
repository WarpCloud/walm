# :book: Walmctl 使用指南

### :ship: Introduction

`walmctl cli` 是用户与 `walm` 进行交互的命令行工具， 通过调用`apiserver` 的接口来与 `walm`服务进行交互, 旨在替代 `helm` 命令行， 方便用户的使用。

***walmctl 执行每条命令都需要指定 `walm host(server)`和操作的 `namespace`，为了方便，可以设定env变量，eg:***

```shell
walmctl -s/--server xxx.xxx.xxx.xxx:xxx -n/--namespace p1138 [commands] ...
export WALM_HOST=0.0.0.0:9001
```



### :rocket: Commands

#### :one: 创建资源： `create`

- 根据 json/yaml 文件创建一个名称为releaseName的release

```shell
walmctl -n/--namespace xxx create release releaseName -f xxx.json/xxx.yaml
```

- 根据 json/yaml 文件创建一个名称为projectName的project

```
walmctl -n/--namespace xxx create project projectName -f xxx.json/xxx.yaml
```

**可选参数**

--async 异步与否（true/false, 默认true）

--timeoutSec 设置超时（默认0）

```shell
walmctl -n xxx create project projectName -f xxx.json --async false --timeoutSec 600
```

#### :two: 删除资源:   `delete`

- 删除一个 project

```shell
walmctl -n xxx delete project projectName
```

- 删除一个 release

```shell
walmctl -n xxx delete release releaseName
```

- 删除某个project的组件（release）

```shell
walmctl -n xxx delete release releaseName -p/--project projectName
```

**可选参数**

--async 异步与否（true/false, 默认true）

--timeoutSec 设置超时（默认0）

--deletePvcs 是否删除release管理的statefulSet关联的所有pvc (默认 true)

```shell
walmctl -n xxx delete release releaseName --async true --timeoutSec 0 --deletePvcs false
```

#### :three: 列举资源:   `list`

- 列举 namespace 下的所有release (:warning: **labelSelector will be support in the future**) 

```shell
walmctl -n xxx list release
```

- 列举 namespace下的 所有 project

```shell
walmctl -n xxx list project
```

- 列举 namespace 下的 某个 project的所有release

```shell
walmctl -n xxx list release -p/--project projectName
```

#### :four: 获取资源详细信息:   `get`

- 获取某个release 信息

```shell
walmctl -n xxx get release releaseName -o json
```

- 获取某个project 信息

```shell
walmctl -n xxx get project projectName -o yaml
```

**必选参数**

-o/--output  json/yaml :warning: 以json或者yaml的形式输出资源的完整信息， 目前建议用 json

``` 
walmctl -n xxx get release releaseName -o/--output json
```

#### :five: 更新已有的资源：  `update`

:warning: 更新资源是在已有资源的基础上根据configPath 和 本地 chart 对 资源进行更新

Required Flags:

--set-string:  set values on the command line (can specify multiple or separate values with commas: key1=val1,dependencies.guardian=…  

**eg:  如果一个release的详细信息是这样的：**

```json
{
    "name": "txsql-dzy",
    "repo_name": "",
    "config_values": {
        "App": {
            "txsql": {
                "image": "txsql:transwarp-5.2.2-final",
                "replicas": 1,
                "resources": {
                    "cpu_limit": 2,
                    "cpu_request": 1,
                    "memory_limit": 3,
                    "memory_request": 1
                }
            }
        }
    },
    "release_status": {
        "services": [
            {
                ......
                "ports": [
                    {
                        "name": "mysql-port",
                        "protocol": "TCP",
                        "port": 3306,
                        "target_port": "3306",
                        "node_port": 0,
                        "endpoints": []
                    }
                ],
                "cluster_ip": "None",
                "service_type": "ClusterIP"
            }
        ],
    }
    ......
}
    
```

:warning: 修改 name 的话 不是升级 release， 而是会创建一个 新的release, 非法的或者不存在的 configPath (属性路径) 会有报错提示

1. 修改 replicas为2  *（single）*

```shell
--set-string config_values.App.txsql.replicas=1
```

2. 修改 cpu_limit为3, memory_limit 为 4  *（multiply）*

```shell
--set-string config_values.App.txsql.cpu_limit=3,config_values.App.txsql.memory_limit=4
```

3. 修改 ports的下标为0的元素的 protocol 属性

```shell
--set-string release_status.services.ports.0.protocol=IP
```

4. 支持同时根据`--set-string`和 `-f/--file`对资源进行更新， 其中 `--set-string`的优先级高于`-f/--file`。

```shell
--set-string config_values.App.txsql.replicas=1 -f txsql.yaml 
```

- 通过`flag: --set-string`编辑 需要更改的release 属性进行 release 升级 **（不需要更改的属性不必添加在configPath中）**

```shell
walmctl -n xxx update release releaseName --set-string a=xxx,b=yyy,c=zzz
```

- 根据本地 chart  进行 release 的 升级

```shell
walmctl -n xxx update release releaseName --withchart /Users/corndai/Desktop/txsql-6.0.0.tgz
```

Advanced:

```shell
walmctl -n xxx update release releaseName --withchart /Users/corndai/Desktop/txsql-6.0.0.tgz --set-string config_values.App.txsql.resources.cpu_limit=4
```

#### :six: 格式检查 && dryrun： `lint`

**check:**
- metainfo是否符合yaml格式
  - 文件后缀名是否为yaml
  - 相同层级的元素左侧是否对齐
  - 对象后的 value 与冒号之间必须要有一到多个空格
- metainfo中 baseconfig中 每个元素都要有 mapKey, type对象, 且 不能是 mapkey, TYPE, MAPKEY 这样的字段类型。
而 resources type 可以不存在
- mapKey 不能为空字符串
- mapKey 在 values是否存在
- metainfo mapKey 类型与 values 中的 键值 类型是否对应
- 添加对开源chart的metainfo的check support

**dryrun:**
模拟charts 是否能够正常安装
```
walmctl lint --chartPath transwarp-native-charts/$chart
```
#### :seven: charts打包： `package`

提供charts的打包功能， 对于开源的charts，可以通过 `helm package`完成，非开源charts 可通过 `walmctl package`完成。
以[application-helmcharts](https://github.com/WarpCloud/walm-charts)中的charts为例。
对于transwarp-jsonnetcharts， 通过以下命令进行打包， 其中 `destination`指输出路径， 若不填代表charts打包输出在当前目录下。
```
walmctl package --chartPath transwarp-jsonnetcharts/${transwarpchart}/${appVersion} --destination ${OUTPUTDIR}
```

#### :eight: 编辑服务端资源： `edit`
目前仅限编辑release资源
```
walmctl -s xxx -n edit release releaseName -o yaml/json
```


### :apple: Todo

- [ ] [get] 不加flag —output/-o 后只返回主要信息而不是全部信息
- [ ]   安全验证， 权限管理 view， admin， namespace  ！！！
- [ ]   检索namespace下所有release添加 labelSelector支持

