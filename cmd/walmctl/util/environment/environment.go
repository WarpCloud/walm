package environment

import (
	"os"

	"github.com/spf13/pflag"
)

// EnvSettings describes all of the environment settings.
type EnvSettings struct {
	// Debug indicates whether or not Helm is running in Debug mode.
	Debug bool
}

// AddFlags binds flags to the given flagset.
func (s *EnvSettings) AddFlags(fs *pflag.FlagSet) {
}

// Init sets values from the environment.
func (s *EnvSettings) Init(fs *pflag.FlagSet) {
	for name, envar := range envMap {
		setFlagFromEnv(name, envar, fs)
	}
}

// envMap maps flag names to envvars
var envMap = map[string]string{
}

func setFlagFromEnv(name, envar string, fs *pflag.FlagSet) {
	if fs.Changed(name) {
		return
	}
	if v, ok := os.LookupEnv(envar); ok {
		fs.Set(name, v)
	}
}

// Deprecated
const (
)
