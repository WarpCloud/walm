# Copyright 2017 The OpenPitrix Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.

GIT_SHA    = $(shell git rev-parse --short HEAD)
GIT_TAG    = $(shell git describe --tags --abbrev=0 2>/dev/null)
BUILD_DATE = $(shell date -u +%Y%m%d-%H:%M:%S)

LDFLAGS += -X WarpCloud/walm/pkg/version.Version=${GIT_TAG}
LDFLAGS += -X WarpCloud/walm/pkg/version.GitSha1Version=${GIT_SHA}
LDFLAGS += -X WarpCloud/walm/pkg/version.BuildDate=${BUILD_DATE}

PKG = WarpCloud/walm

.PHONY: build
build:
	GOOS=linux GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -o _output/walm
	GOOS=linux GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -o _output/walmctl WarpCloud/walm/cmd/walmctl

build_arm64:
	GOOS=linux GOARCH=arm64 go build -ldflags '$(LDFLAGS)' -o _output/arm64/walm
	GOOS=linux GOARCH=arm64 go build -ldflags '$(LDFLAGS)' -o _output/arm64/walmctl WarpCloud/walm/cmd/walmctl

build_darwin:
	GOOS=darwin GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -o _output/walm-darwin-amd64
	GOOS=darwin GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -o _output/walmctl-darwin-amd64 WarpCloud/walm/cmd/walmctl

build_windows:
	GOOS=windows GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -o _output/walm-windows-amd64.exe
	GOOS=windows GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -o _output/walmctl-windows-amd64.exe WarpCloud/walm/cmd/walmctl

all: build build_darwin build_windows

.PHONY: test
test:
	@go test -race $(shell go list ${PKG}/... | grep -v vendor | grep -v '/test') -coverprofile ut-coverage-report.out

.PHONY: e2e-test
e2e-test:
	@ginkgo version || go get -u github.com/onsi/ginkgo/ginkgo
	@rm -f ./e2e-test-coverage-report.out
	@ginkgo -randomizeAllSpecs -flakeAttempts=2 -trace -outputdir=./ -coverprofile=e2e-test-coverage-report.out -coverpkg=WarpCloud/walm/pkg/... ./test/
