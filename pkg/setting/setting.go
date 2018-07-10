package setting

import (
	"fmt"
	"time"
	"walm/pkg/setting/homepath"

	"os"
	"path/filepath"

	"github.com/spf13/pflag"
	"k8s.io/client-go/util/homedir"
)

var DefaultWalmHome = filepath.Join(homedir.HomeDir(), ".walm")

type Config struct {
	HTTPPort     int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	PageSize  int
	JwtSecret string

	TillerConnectionTimeout int64
	// Home is the local path to the Helm home directory.
	Home homepath.Home
	// Debug indicates whether or not Helm is running in Debug mode.
	Debug bool
	// KubeContext is the name of the kubeconfig context.
	KubeContext string

	ZipkinUrl string
}

var envMap = map[string]string{
	"debug": DebugEnvVar,
	"home":  HomeEnvVar,
	"port":  PortEnvVar,

	"httpreadtimeout":  HTTPRTimeOutEnvVar,
	"httpwritetimeout": HTTPWTimeOutEnvVar,

	"zipkinurl":                 ZipkinUrl,
	"tiller-connection-timeout": TillerConnTimeOut,
	"kube-context":              KubeContext,
	"JwtSecret":                 JwtSecret,
}

// AddFlags binds flags to the given flagset.
func (conf *Config) AddFlags(fs *pflag.FlagSet) {

	fs.BoolVar(&conf.Debug, "debug", false, "enable verbose output")
	fs.StringVar((*string)(&conf.Home), "home", DefaultWalmHome, "location of your Walm config. Overrides $WALM_HOME")
	fs.IntVar(&conf.HTTPPort, "port", 8000, "api server port")

	fs.StringVar(&conf.JwtSecret, "jwtsecret", "", "value of jwtsecrect")

	fs.DurationVar(&conf.ReadTimeout, "httpreadtimeout", time.Duration(0), "httpreadtimeout")
	fs.DurationVar(&conf.WriteTimeout, "httpwritetimeout", time.Duration(0), "httpwritetimeout")

	fs.StringVar(&conf.ZipkinUrl, "zipkinurl", "", "zipkin url")
	fs.Int64Var(&conf.TillerConnectionTimeout, "tiller-connection-timeout", int64(300), "the duration (in seconds) Helm will wait to establish a connection to tiller")
	fs.StringVar(&conf.KubeContext, "kube-context", "", "name of the kubeconfig context to use")
}

// Init sets values from the environment.
func (conf *Config) Init(fs *pflag.FlagSet) {
	for name, envar := range envMap {
		conf.setFlagFromEnv(name, envar, fs)
	}

	{
		ensureDirectories(conf.Home)
	}
}

func (conf *Config) setFlagFromEnv(name, envar string, fs *pflag.FlagSet) {
	if fs.Changed(name) {
		return
	}
	if v, ok := os.LookupEnv(envar); ok {
		fs.Set(name, v)
	}
}

// Deprecated
const (
	HomeEnvVar  = "WALM_HOME"
	PortEnvVar  = "WALM_HTTP_PORT"
	DebugEnvVar = "WALM_DEBUG"

	HTTPRTimeOutEnvVar = "HTTP_READ_TIMEOUT"
	HTTPWTimeOutEnvVar = "HTTP_WRITE_TIMEOUT"

	ZipkinUrl         = "ZIPKIN_URL"
	TillerConnTimeOut = "TILLER_CONN_TIMEOUT"
	KubeContext       = "KUBE_CONTEXT"
	JwtSecret         = "WALM_JWTSECRET"
)

// envMap maps flag names to envvars

func ensureDirectories(home homepath.Home) error {
	configDirectories := []string{
		home.String(),
	}

	for _, p := range configDirectories {
		if fi, err := os.Stat(p); err != nil {
			if err := os.MkdirAll(p, 0755); err != nil {
				return fmt.Errorf("Could not create %s: %s", p, err)
			}
		} else if !fi.IsDir() {
			return fmt.Errorf("%s must be a directory", p)
		}
	}

	return nil
}
