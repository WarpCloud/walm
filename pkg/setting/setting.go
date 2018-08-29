package setting

import (
	"time"

	"io/ioutil"
	"github.com/sirupsen/logrus"
	"github.com/ghodss/yaml"
)

var Config WalmConfig

type HttpConfig struct {
	HTTPPort int `json:"port,default=9999"`
	TLS bool `json:"tls"`
	TlsKey string `json:"tlsKey"`
	TlsCert string `json:"tlsCert"`
	ReadTimeout  time.Duration `json:"readTimeout"`
	WriteTimeout time.Duration `json:"writeTimeout"`
}

type HelmConfig struct {
	TillerConnectionTimeout time.Duration `json:"tillerTimeout"`
	TillerHost string `json:"tillerHost"`
	TillerHome string `json:"tillerHome"`
	TLS bool `json:"tls"`
	TlsKey string `json:"tlsKey"`
	TlsCert string `json:"tlsCert"`
	TlsCACert string `json:"tlsCert"`
}

type ChartRepo struct {
	Name string `json:"name"`
	URL string `json:"url"`
}

type KubeConfig struct {
	Config string `json:"config"`
}

type RedisConfig struct {
	Addr string `json:"addr"`
	Password string `json:"password"`
	DB int `json:"db"`
}

type WalmConfig struct {
	Debug bool `json:"debug"`

	HttpConfig *HttpConfig `json:"serverConfig"`
	SysHelm *HelmConfig `json:"sysHelmConfig"`
	RepoList []*ChartRepo `json:"repoList"`
	KubeConfig *KubeConfig `json:"kubeConfig"`
	RedisConfig *RedisConfig `json:"redisConfig"`
}

// StartResyncReleaseCaches sets values from the environment.
func InitConfig(configPath string) {
	yamlFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		logrus.Fatalf("Read config file faild! %s\n", err.Error())
	}
	err = yaml.Unmarshal(yamlFile, &Config)
	if err != nil {
		logrus.Fatalf("Unmarshal config file faild! %s\n", err.Error())
	}
}
