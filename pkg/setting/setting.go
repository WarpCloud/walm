package setting

import (
	"fmt"
	"time"
	"walm/pkg/setting/homepath"

	"os"
	"path/filepath"
	"strconv"

	"github.com/go-ini/ini"
	"github.com/spf13/pflag"
	"k8s.io/client-go/util/homedir"

	. "walm/pkg/util/log"
)

var DefaultWalmHome = filepath.Join(homedir.HomeDir(), ".walm")

type Config struct {
	Cfg *ini.File

	RunMode string

	HTTPPort int

	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	PageSize  int
	JwtSecret string

	//db config
	DbType, DbName, DbUser, DbPassword, DbHost, TablePrefix string

	TillerConnectionTimeout int64
	// Home is the local path to the Helm home directory.
	Home           homepath.Home
	ValueCachePath string
	// Debug indicates whether or not Helm is running in Debug mode.
	Debug bool
	// KubeContext is the name of the kubeconfig context.
	KubeContext string

	ZipkinUrl string
}

func (conf Config) InitIni() {
	iniPath := conf.Home.Config() + "/app.ini"
	var err error
	conf.Cfg, err = ini.Load(iniPath)
	if err != nil {
		Log.Fatalf("Fail to parse '%s': %v", iniPath, err)
	}

	conf.LoadBase()
	conf.LoadServer()
	conf.LoadApp()
	conf.LoadDatabase()
}

func (conf Config) LoadBase() {
	conf.RunMode = conf.Cfg.Section("").Key("RUN_MODE").MustString("debug")
}

func (conf Config) LoadServer() {
	sec, err := conf.Cfg.GetSection("server")
	if err != nil {
		Log.Fatalf("Fail to get section 'server': %v", err)
	}

	conf.RunMode = conf.Cfg.Section("").Key("RUN_MODE").MustString("debug")

	conf.HTTPPort = sec.Key("HTTP_PORT").MustInt(8000)

	conf.ReadTimeout = time.Duration(sec.Key("READ_TIMEOUT").MustInt(60)) * time.Second
	conf.WriteTimeout = time.Duration(sec.Key("WRITE_TIMEOUT").MustInt(60)) * time.Second
}

func (conf Config) LoadApp() {
	sec, err := conf.Cfg.GetSection("app")
	if err != nil {
		Log.Fatalf("Fail to get section 'app': %v", err)
	}

	conf.JwtSecret = sec.Key("JWT_SECRET").MustString("!@)*#)!@U#@*!@!)")
	conf.PageSize = sec.Key("PAGE_SIZE").MustInt(10)
}

func (conf Config) LoadDatabase() {
	sec, err := conf.Cfg.GetSection("database")
	if err != nil {
		Log.Fatal(2, "Fail to get section 'database': %v", err)
	}

	conf.DbType = sec.Key("TYPE").String()
	conf.DbName = sec.Key("NAME").String()
	conf.DbUser = sec.Key("USER").String()
	conf.DbPassword = sec.Key("PASSWORD").String()
	conf.DbHost = sec.Key("HOST").String()
	conf.TablePrefix = sec.Key("TABLE_PREFIX").String()
}

// AddFlags binds flags to the given flagset.
func (conf Config) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar((*string)(&conf.Home), "home", DefaultWalmHome, "location of your Walm config. Overrides $WALM_HOME")
	fs.StringVar(&conf.KubeContext, "kube-context", "", "name of the kubeconfig context to use")
	fs.BoolVar(&conf.Debug, "debug", false, "enable verbose output")
	fs.Int64Var(&conf.TillerConnectionTimeout, "tiller-connection-timeout", int64(300), "the duration (in seconds) Helm will wait to establish a connection to tiller")
}

// AddFlags binds flags to the given flagset.
func (conf Config) AddServFlags(fs *pflag.FlagSet) {
	fs.Int64Var(&conf.TillerConnectionTimeout, "tiller-connection-timeout", int64(300), "the duration (in seconds) Helm will wait to establish a connection to tiller")
}

// Init sets values from the environment.
func (conf Config) Init(fs *pflag.FlagSet) {
	for name, envar := range envMap {
		conf.setFlagFromEnv(name, envar, fs)
	}

	{
		ensureDirectories(conf.Home)
		conf.ValueCachePath = conf.Home.Cache()
	}

	conf.InitIni()
	conf.EnableEnvValue()
}

func (conf Config) EnableEnvValue() {
	if v, ok := os.LookupEnv(DbNameEnvVar); ok {
		conf.DbName = v
	}

	if v, ok := os.LookupEnv(DbTypeEnvVar); ok {
		conf.DbType = v
	}

	if v, ok := os.LookupEnv(DbUserEnvVar); ok {
		conf.DbUser = v
	}

	if v, ok := os.LookupEnv(DbPassEnvVar); ok {
		conf.DbPassword = v
	}

	if v, ok := os.LookupEnv(DbHostEnvVar); ok {
		conf.DbHost = v
	}

	if v, ok := os.LookupEnv(DbTabPreEnvVar); ok {
		conf.TablePrefix = v
	}

	if v, ok := os.LookupEnv(HTTPRTimeOutEnvVar); ok {
		if r, err := strconv.Atoi(v); err == nil {
			conf.ReadTimeout = time.Duration(r) * time.Second
		}
	}

	if v, ok := os.LookupEnv(HTTPWTimeOutEnvVar); ok {

		if w, err := strconv.Atoi(v); err == nil {
			conf.WriteTimeout = time.Duration(w) * time.Second
		}
	}

	if v, ok := os.LookupEnv(ZipkinUrl); ok {
		conf.ZipkinUrl = v
	}
}

func (conf Config) setFlagFromEnv(name, envar string, fs *pflag.FlagSet) {
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

	DbNameEnvVar   = "WALM_DB_NAME"
	DbTypeEnvVar   = "WALM_DB_TYPE"
	DbUserEnvVar   = "WALM_DB_USER"
	DbPassEnvVar   = "WALM_DB_PASS"
	DbHostEnvVar   = "WALM_DB_HOST"
	DbTabPreEnvVar = "WALM_TABLE_PREFIX"

	HTTPRTimeOutEnvVar = "HTTP_READ_TIMEOUT"
	HTTPWTimeOutEnvVar = "HTTP_WRITE_TIMEOUT"

	ZipkinUrl = "ZIPKIN_URL"
)

// envMap maps flag names to envvars
var envMap = map[string]string{
	"debug": DebugEnvVar,
	"home":  HomeEnvVar,
	"port":  PortEnvVar,
}

func ensureDirectories(home homepath.Home) error {
	configDirectories := []string{
		home.String(),
		home.Cache(),
		home.Config(),
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
