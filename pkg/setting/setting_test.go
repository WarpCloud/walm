/*
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package setting

import (
	"os"
	"strings"
	"testing"

	"walm/pkg/setting/homepath"

	"github.com/spf13/pflag"
)

func TestEnvSettings(t *testing.T) {
	tests := []struct {
		name string

		// input
		args   []string
		envars map[string]string

		// expected values
		home, host, ns, kcontext, plugins string
		debug                             bool
	}{
		{
			name:    "defaults",
			args:    []string{},
			home:    DefaultWalmHome,
			plugins: homepath.Home(DefaultWalmHome).Plugins(),
		},
		{
			name:    "with flags set",
			args:    []string{"--home", "/foo", "--host=here", "--debug"},
			home:    "/foo",
			plugins: homepath.Home("/foo").Plugins(),
			host:    "here",
			debug:   true,
		},
		{
			name:    "with envvars set",
			args:    []string{},
			envars:  map[string]string{"WALM_HOME": "/bar", "WALM_HOST": "there", "WALM_DEBUG": "1"},
			home:    "/bar",
			plugins: homepath.Home("/bar").Plugins(),
			host:    "there",
			debug:   true,
		},
		{
			name:    "with flags and envvars set",
			args:    []string{"--home", "/foo", "--host=here", "--debug"},
			envars:  map[string]string{"WALM_HOME": "/bar", "WALM_HOST": "there", "WALM_DEBUG": "1", "WALM_PLUGIN": "glade"},
			home:    "/foo",
			plugins: "glade",
			host:    "here",
			debug:   true,
		},
	}

	cleanup := resetEnv()
	defer cleanup()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envars {
				os.Setenv(k, v)
			}

			flags := pflag.NewFlagSet("testing", pflag.ContinueOnError)

			settings := &Config{}
			settings.AddFlags(flags)
			flags.Parse(tt.args)

			settings.Init(flags)

			if settings.Home != homepath.Home(tt.home) {
				t.Errorf("expected home %q, got %q", tt.home, settings.Home)
			}
			if settings.Debug != tt.debug {
				t.Errorf("expected debug %t, got %t", tt.debug, settings.Debug)
			}
			if settings.KubeContext != tt.kcontext {
				t.Errorf("expected kube-context %q, got %q", tt.kcontext, settings.KubeContext)
			}

			cleanup()
		})
	}
}

func resetEnv() func() {
	origEnv := os.Environ()

	// ensure any local envvars do not hose us
	for _, e := range envMap {
		os.Unsetenv(e)
	}

	return func() {
		for _, pair := range origEnv {
			kv := strings.SplitN(pair, "=", 2)
			os.Setenv(kv[0], kv[1])
		}
	}
}
