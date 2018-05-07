// Copyright 2016 The Kubernetes Authors All rights reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build windows

package hompath

import (
	"testing"
)

func TestHelmHome(t *testing.T) {
	hh := Home("r:\\")
	isEq := func(t *testing.T, a, b string) {
		if a != b {
			t.Errorf("Expected %q, got %q", b, a)
		}
	}

	isEq(t, hh.String(), "r:\\")
	isEq(t, hh.TLSCaCert(), "r:\\ca.pem")
	isEq(t, hh.TLSCert(), "r:\\cert.pem")
	isEq(t, hh.TLSKey(), "r:\\key.pem")
}
