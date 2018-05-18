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

package homepath

import (
	"os"
	"path/filepath"
)

// Home describes the location of a CLI configuration.
//
// This helper builds paths relative to a Helm Home directory.
type Home string

// String returns Home as a string.
//
// Implements fmt.Stringer.
func (h Home) String() string {
	return os.ExpandEnv(string(h))
}

// Path returns Home with elements appended.
func (h Home) Path(elem ...string) string {
	p := []string{h.String()}
	p = append(p, elem...)
	return filepath.Join(p...)
}

// Plugins returns the path to the plugins directory.
func (h Home) Plugins() string {
	return h.Path("plugins")
}

// TLSCaCert returns the path to fetch the CA certificate.
func (h Home) TLSCaCert() string {
	return h.Path("ca.pem")
}

// TLSCert returns the path to fetch the client certificate.
func (h Home) TLSCert() string {
	return h.Path("cert.pem")
}

// TLSKey returns the path to fetch the client public key.
func (h Home) TLSKey() string {
	return h.Path("key.pem")
}
func (h Home) Cache() string {
	return h.Path("cache")
}

func (h Home) Config() string {
	return h.Path("conf")
}
