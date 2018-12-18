package hook

import (
	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
	"os"
	"reflect"
	"testing"
	"walm/pkg/release"
)

func Test_Merge(t *testing.T) {
	prettyChartParams := release.PrettyChartParams{}
	err := yaml.Unmarshal([]byte(hdfsPrettyParams), &prettyChartParams)
	if err != nil {
		logrus.Printf("hdfsPrettyParams Unmarshal %v\n", err)
	}
	logrus.Printf("prettyParams %+v", prettyChartParams)

	request := release.ReleaseRequest{}
	request.Name = "test"
	request.ChartName = "test"
	request.ConfigValues = make(map[string]interface{}, 0)
	request.ReleasePrettyParams = prettyChartParams

	ProcessPrettyParams(&request)
	logrus.Printf("ConfigValues %+v\n", request.ConfigValues)

	var a interface{}
	logrus.Printf("%v %v\n", reflect.TypeOf(a), reflect.ValueOf(a).Kind().String())
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

var hdfsPrettyParams = `
commonConfig:
  roles:
  - name: hdfsnamenode
    description: "hdfsnamenode服务"
    baseConfig:
    - variable: image
      default: 172.16.1.99/gold/hdfs:transwarp-5.2
      description: "镜像"
      type: string
    - variable: priority
      default: 10
      description: "优先级"
      type: number
    - variable: replicas
      default: 2
      description: "副本个数"
      type: number
    - variable: env_list
      default: []
      description: "额外环境变量"
      type: list
    - variable: use_host_network
      default: false
      description: "是否使用主机网络"
      type: bool
    resouceConfig:
      cpu_limit: 2
      cpu_request: 1
      memory_limit: 8
      memory_request: 4
      gpu_limit: 0
      gpu_request: 0
      extra_resources: []
      storage:
      - name: data
        type: pvc
        storageClass: "silver"
        size: "100Gi"
        accessModes: ["ReadWriteOnce"]
        limit: {}
      - name: log
        type: tosDisk
        storageClass: "silver"
        size: "20Gi"
        accessMode: "ReadWriteOnce"
        limit: {}
  - name: hdfszkfc
    description: "hdfszkfc服务"
    baseConfig:
    - variable: image
      default: 172.16.1.99/gold/hdfs:transwarp-5.2
      description: "镜像"
      type: string
    - variable: env_list
      default: []
      description: "额外环境变量"
      type: list
    resouceConfig:
      cpu_limit: 0.5
      cpu_request: 0.1
      memory_limit: 1
      memory_request: 0.5
      gpu_limit: 0
      gpu_request: 0
      extra_resources: []
  - name: hdfsdatanode
    description: "hdfsdatanode服务"
    baseConfig:
    - variable: image
      default: 172.16.1.99/gold/hdfs:transwarp-5.2
      description: "镜像"
      type: string
    - variable: priority
      default: 10
      description: "优先级"
      type: number
    - variable: replicas
      default: 3
      description: "副本个数"
      type: number
    - variable: env_list
      default: []
      description: "额外环境变量"
      type: list
    - variable: use_host_network
      default: false
      description: "是否使用主机网络"
      type: bool
    resouceConfig:
      cpu_limit: 2
      cpu_request: 0.5
      memory_limit: 4
      memory_request: 1
      gpu_limit: 0
      gpu_request: 0
      extra_resources: []
      storage:
      - name: data
        type: pvc
        storageClass: "silver"
        size: "500Gi"
        accessModes: ["ReadWriteOnce"]
        limit: {}
      - name: log
        type: tosDisk
        storageClass: "silver"
        size: "20Gi"
        accessMode: "ReadWriteOnce"
        limit: {}
  - name: hdfsjournalnode
    description: "hdfsjournalnode服务"
    baseConfig:
    - variable: image
      default: 172.16.1.99/gold/hdfs:transwarp-5.2
      description: "镜像"
      type: string
    - variable: priority
      default: 10
      description: "优先级"
      type: number
    - variable: replicas
      default: 3
      description: "副本个数"
      type: number
    - variable: env_list
      default: []
      description: "额外环境变量"
      type: list
    - variable: use_host_network
      default: false
      description: "是否使用主机网络"
      type: bool
    resouceConfig:
      cpu_limit: 2
      cpu_request: 0.5
      memory_limit: 4
      memory_request: 1
      gpu_limit: 0
      gpu_request: 0
      extra_resources: []
      storage:
      - name: data
        type: pvc
        storageClass: "silver"
        size: "500Gi"
        accessModes: ["ReadWriteOnce"]
        limit: {}
      - name: log
        type: tosDisk
        storageClass: "silver"
        size: "20Gi"
        accessMode: "ReadWriteOnce"
        limit: {}
  - name: httpfs
    description: "httpfs服务"
    baseConfig:
    - variable: image
      default: 172.16.1.99/gold/httpfs:transwarp-5.2
      description: "镜像"
      type: string
    - variable: priority
      default: 10
      description: "优先级"
      type: number
    - variable: replicas
      default: 2
      description: "副本个数"
      type: number
    - variable: env_list
      default: []
      description: "额外环境变量"
      type: list
    - variable: use_host_network
      default: false
      description: "是否使用主机网络"
      type: bool
    resouceConfig:
      cpu_limit: 2
      cpu_request: 0.5
      memory_limit: 4
      memory_request: 1
      gpu_limit: 0
      gpu_request: 0
      extra_resources: []
      storage:
      - name: log
        type: tosDisk
        storageClass: "silver"
        size: "20Gi"
        accessMode: "ReadWriteOnce"
        limit: {}
transwarpBundleConfig:
- variable: Transwarp_Config.Transwarp_Metric_Enable
  default: true
  description: "是否开启组件metrics服务"
  type: bool
- variable: Transwarp_Config.Transwarp_Auto_Injected_Volumes
  default: []
  description: "自动挂载keytab目录"
  type: list
- variable: Transwarp_Config.security.auth_type
  default: "none"
  description: "开启安全类型"
  type: string
- variable: Transwarp_Config.security.guardian_principal_host
  default: "tos"
  description: "开启安全服务Principal主机名"
  type: string
- variable: Transwarp_Config.security.guardian_principal_user
  default: "hdfs"
  description: "开启安全服务Principal用户名"
  type: string
- variable: Transwarp_Config.security.guardian_spnego_principal_host
  default: "tos"
  description: "Httpfs开启安全服务Principal主机名"
  type: string
- variable: Transwarp_Config.security.guardian_spnego_principal_user
  default: "HTTP"
  description: "Httpfs开启安全服务Principal用户名"
  type: string
- variable: Transwarp_Config.Ingress
  default: {}
  description: "HDFS Ingress配置参数"
  type: yaml
advanceConfig:
- variable: Advance_Config.hdfs
  default: {}
  description: "hdfs guardian配置"
  type: yaml
- variable: Advance_Config.core_site
  default: {}
  description: "hdfs core-site配置"
  type: yaml
- variable: Advance_Config.hdfs_site
  default: {}
  description: "hdfs hdfs-site配置"
  type: yaml
- variable: Advance_Config.httpfs_site
  default: {}
  description: "hdfs httpfs-site配置"
  type: yaml
`
