package setting

import (
	"io/ioutil"
	"github.com/sirupsen/logrus"
	"github.com/ghodss/yaml"
)

var Config WalmConfig

type HttpConfig struct {
	HTTPPort     int           `json:"port,default=9999"`
	TLS          bool          `json:"tls"`
	TlsKey       string        `json:"tlsKey"`
	TlsCert      string        `json:"tlsCert"`
}

type ChartRepo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type KubeConfig struct {
	Config string `json:"config"`
}

type RedisConfig struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

type KafkaConfig struct {
	Enable    bool     `json:"enable"`
	Brokers   []string `json:"brokers"`
	CertFile  string   `json:"certFile"`
	KeyFile   string   `json:"keyFile"`
	CaFile    string   `json:"caFile"`
	VerifySsl bool     `json:"verifySsl"`
}

type TaskConfig struct {
	Broker          string `json:"broker"`
	DefaultQueue    string `json:"default_queue"`
	ResultBackend   string `json:"result_backend"`
	ResultsExpireIn int    `json:"results_expire_in"`
}

type WalmConfig struct {
	Debug         bool           `json:"debug"`
	LogConfig     *LogConfig     `json:"logConfig"`
	HttpConfig    *HttpConfig    `json:"serverConfig"`
	RepoList      []*ChartRepo   `json:"repoList"`
	KubeConfig    *KubeConfig    `json:"kubeConfig"`
	RedisConfig   *RedisConfig   `json:"redisConfig"`
	KafkaConfig   *KafkaConfig   `json:"kafkaConfig"`
	TaskConfig    *TaskConfig    `json:"taskConfig"`
	JsonnetConfig *JsonnetConfig `json:"jsonnetConfig"`
	ChartImageConfig *ChartImageConfig `json:"chartImageConfig"`

	//only for test
	ChartImageRegistry string `json:"chartImageRegistry"`
}

type ChartImageConfig struct {
	CacheRootDir string  `json:"cacheRootDir"`
}

type LogConfig struct {
	Level string `json:"level"`
}

type JsonnetConfig struct {
	CommonTemplateFilesPath string `json:"commonTemplateFilesPath"`
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
