package setting

import (
	"path/filepath"
	"time"

	"github.com/spf13/viper"

	"k8s.io/client-go/util/homedir"
)

var ConfigPath = "/home/hanbing/myworkspace/go/src/viper_demo"

var DefaultWalmHome = filepath.Join(homedir.HomeDir(), ".walm")

type Config struct {
	Home  string `yaml:"home"`
	Debug bool   `yaml:"debug"`

	Http struct {
		HTTPPort     int           `yaml:"port"`
		ReadTimeout  time.Duration `yaml:"read_timeout"`
		WriteTimeout time.Duration `yaml:"write_timeout"`
	} `yaml:"http"`

	Secret struct {
		Tls       bool   `yaml:"tls"`
		TlsVerify bool   `yaml:"tls-verify"`
		TlsKey    string `yaml:"tls-key"`
		TlsCert   string `yaml:"tls-cert"`
		TlsCaCert string `yaml:"tls-ca-cert"`
	} `yaml:"secret"`

	Helm struct {
		TillerConnectionTimeout time.Duration `yaml:"tiller_time_out"`
		TillerHost              string        `yaml:"tillerHost"`
	} `yaml:"helm"`

	Kube struct {
		KubeContext string `yaml:"config"`
		KubeConfig  string `yaml:"context"`
	} `yaml:"kube"`

	Trace struct {
		ZipkinUrl string `yaml:"zipkin_url"`
	} `yaml:"trace"`

	Auth struct {
		Enable    bool   `yaml:"enalbe"`
		JwtSecret string `yaml:"jwtsecret"`
	} `yaml:"auth"`
}

// Init sets values from the environment.
func (conf *Config) Init() {
	viper.SetConfigType("yaml")
	viper.SetConfigName("conf")
	viper.SetDefault("home", DefaultWalmHome)
	viper.AddConfigPath(ConfigPath)
	viper.ReadInConfig()
	viper.Unmarshal(conf)
}
