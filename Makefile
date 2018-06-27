# Copyright 2017 The OpenPitrix Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.

TARG.Name:=walm
#TRAG.Gopkg:=transwarp.io/walm
TRAG.Gopkg:=walm
TRAG.Version:=$(TRAG.Gopkg)/pkg/version

DOCKER_TAGS=latest
RUN_IN_DOCKER:=docker run -it -v `pwd`:/go/src/$(TRAG.Gopkg)  -w /go/src/$(TRAG.Gopkg) -e CGO_ENABLED=0 -e GOBIN=/go/src/$(TRAG.Gopkg)/bin -e USER_ID=`id -u` -e GROUP_ID=`id -g` 172.16.1.99/transwarp/walm-builder:1.0
RUN_IN_DOCKER_WITH_DB:=docker run -it -v `pwd`:/go/src/$(TRAG.Gopkg)  -w /go/src/$(TRAG.Gopkg) -e CGO_ENABLED=0 -e GOBIN=/go/src/$(TRAG.Gopkg)/bin -e USER_ID=`id -u` -e GROUP_ID=`id -g` -e WITH_DB_UNIT_TEST=1 -e WALM_DB_HOST=127.0.0.1:13306 -e WALM_DB_NAME=walm -e WALM_DB_USER=root -e WALM_DB_PASS=passwd 172.16.1.99/transwarp/walm-builder:1.0
GO_FILES:=./cmd ./test ./pkg ./models ./router
WITH_DB_TEST:=WITH_DB_UNIT_TEST=1 WALM_DB_HOST=127.0.0.1:13306 WALM_DB_NAME=walm WALM_DB_USER=root WALM_DB_PASS=passwd

define get_diff_files
    $(eval DIFF_FILES=$(shell git diff --name-only --diff-filter=ad | grep -E "^(test|cmd|pkg|models|router)/.+\.go"))
endef
define get_build_flags
    $(eval SHORT_VERSION=$(shell git describe --tags --always --dirty="-dev"))
    $(eval SHA1_VERSION=$(shell git show --quiet --pretty=format:%H))
	$(eval DATE=$(shell date +'%Y-%m-%dT%H:%M:%S'))
	$(eval BUILD_FLAG= -X $(TRAG.Version).ShortVersion="$(SHORT_VERSION)" \
		-X $(TRAG.Version).GitSha1Version="$(SHA1_VERSION)" \
		-X $(TRAG.Version).BuildDate="$(DATE)")
endef

COMPOSE_APP_SERVICES=openpitrix-runtime-manager openpitrix-app-manager openpitrix-repo-indexer openpitrix-api-gateway openpitrix-repo-manager openpitrix-job-manager openpitrix-task-manager openpitrix-cluster-manager
COMPOSE_DB_CTRL=openpitrix-app-db-ctrl openpitrix-repo-db-ctrl openpitrix-runtime-db-ctrl openpitrix-job-db-ctrl openpitrix-task-db-ctrl openpitrix-cluster-db-ctrl


.PHONY: init-vendor
init-vendor:
	$(RUN_IN_DOCKER) glide init
	@echo "init-vendor done"

.PHONY: update-vendor
update-vendor:
	$(RUN_IN_DOCKER) glide update
	@echo "update-vendor done"

.PHONY: update-builder
update-builder:
	docker pull 172.16.1.99/transwarp/walm-builder:1.0
	@echo "update-builder done"

#all:init-vendor/update-vendor update-builder generate build
.PHONY: all
all: generate build

.PHONY: generate
generate:
	$(RUN_IN_DOCKER) make gen-version
	@echo "generate done"

.PHONY: generate-in-local
gen-version:
	go generate ./pkg/version/


.PHONY: gen-swagger
gen-swagger:
	$(RUN_IN_DOCKER) make generate-swagger
	@echo "gen-swagger done"

.PHONY: gen-swagger-in-local
generate-swagger:
	@swag init -g router/routers.go

.PHONY: build
build:
	@echo "build" $(TARG.Name):$(DOCKER_TAGS)
	@docker build -t $(TARG.Name):$(DOCKER_TAGS) .
	@docker image prune -f 1>/dev/null 2>&1
	@echo "build done"

.PHONY: install
install:
	time go install -v -ldflags '$(BUILD_FLAG)' $(TRAG.Gopkg)/cmd/walm
	#$(RUN_IN_DOCKER) time go install -v -ldflags '$(BUILD_FLAG)' $(TRAG.Gopkg)/cmd/walm

.PHONY: test
test:
	@make unit-test
	@make e2e-test
	@echo "test done"

.PHONY: unit-test
unit-test:
	$(RUN_IN_DOCKER_WITH_DB)  go test -v -a -tags="unit db" ./...
	@echo "unit-test done"

.PHONY: purge-test
purge-test:
	$(RUN_IN_DOCKER) go test -v -a  -tags="purge" ./...
	@echo "purge-test done"

.PHONY: e2e-test
e2e-test:
	$(RUN_IN_DOCKER) go test -v -a -tags="integration db" ./test/...
	@echo "e2e-test done"

.PHONY: ci-test
ci-test:
	# build with production Dockerfile, not dev version
	@docker build -t $(TARG.Name) -f ./Dockerfile .
	@make compose-up
	sleep 20
	@make unit-test
	@make e2e-test
	@echo "ci-test done"

.PHONY: clean
clean:
	-make -C ./pkg/version clean
	@echo "ok"

