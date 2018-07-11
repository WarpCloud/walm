package setting

import (
	"time"

	"github.com/spf13/pflag"
)

type Config struct {
	HTTPPort     int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	PageSize  int
	JwtSecret string

	RepoURL string
	TillerHost string
	TillerConnectionTimeout int64
	// Debug indicates whether or not Helm is running in Debug mode.
	Debug bool
	// KubeContext is the name of the kubeconfig context.
	KubeContext string

	ZipkinUrl string
}

// AddFlags binds flags to the given flagset.
func (conf *Config) AddFlags(fs *pflag.FlagSet) {

	fs.BoolVar(&conf.Debug, "debug", false, "enable verbose output")
	fs.IntVar(&conf.HTTPPort, "port", 8000, "api server port")

	fs.StringVar(&conf.JwtSecret, "jwtsecret", "", "value of jwtsecrect")

	fs.DurationVar(&conf.ReadTimeout, "httpreadtimeout", time.Duration(0), "httpreadtimeout")
	fs.DurationVar(&conf.WriteTimeout, "httpwritetimeout", time.Duration(0), "httpwritetimeout")

	fs.StringVar(&conf.ZipkinUrl, "zipkinurl", "", "zipkin url")
	fs.StringVar(&conf.KubeContext, "kube-context", "", "name of the kubeconfig context to use")

	fs.StringVar(&conf.RepoURL, "reporul", "", "default chart repo address")
	fs.StringVar(&conf.TillerHost, "tiller-host", "", "tiller address")
	fs.Int64Var(&conf.TillerConnectionTimeout, "tiller-connection-timeout", int64(300), "the duration (in seconds) Helm will wait to establish a connection to tiller")
}
