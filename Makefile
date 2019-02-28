# Copyright 2017 The OpenPitrix Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.

GIT_SHA    = $(shell git rev-parse --short HEAD)
GIT_TAG    = $(shell git describe --tags --abbrev=0 2>/dev/null)
BUILD_DATE = $(shell date -u +%Y%m%d-%H:%M:%S)

LDFLAGS += -X walm/pkg/version.Version=${GIT_TAG}
LDFLAGS += -X walm/pkg/version.GitSha1Version=${GIT_SHA}
LDFLAGS += -X walm/pkg/version.BuildDate=${BUILD_DATE}

.PHONY: build
build:
	GOOS=linux GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -o _output/walm
	GOOS=linux GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -o _output/walmcli walm/walmcli

build_darwin:
	GOOS=darwin GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -o _output/walm-darwin-amd64
	GOOS=linux GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -o _output/walmcli-darwin-amd64 walm/walmcli

build_windows:
	GOOS=windows GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -o _output/walm-windows-amd64.exe
	GOOS=linux GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -o _output/walmcli-windows-amd64.exe walm/walmcli

all: build build_darwin build_windows
