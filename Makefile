# Copyright © 2022 The OpenEBS Authors.  All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# list only populator source code directories
PACKAGES = $(shell go list ./... | grep -v 'vendor')

UNIT_TEST_PACKAGES = $(shell go list ./... | grep -v 'vendor')

# Specify the name for the rsync-populator binary
RSYNC_POPULATOR=rsync-populator
# Specify the name for the data-populator binary
DATA_POPULATOR=data-populator

RSYNC_DAEMON=rsync-daemon
RSYNC_CLIENT=rsync-client

# The images can be pushed to any docker/image registeries
# like docker hub, quay. The registries are specified in
# the `build/push` script.
#
# The images of a project or company can then be grouped
# or hosted under a unique organization key like `openebs`
#
# Each component (container) will be pushed to a unique
# repository under an organization.
# Putting all this together, an unique uri for a given
# image comprises of:
#   <registry url>/<image org>/<image repo>:<image-tag>
#
# IMAGE_ORG can be used to customize the organization
# under which images should be pushed.
# By default the organization name is `openebs`.

ifeq (${IMAGE_ORG}, )
  IMAGE_ORG="openebs"
  export IMAGE_ORG
endif

# Specify the date of build
DBUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

# Specify the docker arg for repository url
ifeq (${DBUILD_REPO_URL}, )
  DBUILD_REPO_URL="https://github.com/openebs/data-populator"
  export DBUILD_REPO_URL
endif

# Specify the docker arg for website url
ifeq (${DBUILD_SITE_URL}, )
  DBUILD_SITE_URL="https://openebs.io"
  export DBUILD_SITE_URL
endif

ifeq (${IMAGE_TAG}, )
  IMAGE_TAG = ci
  export IMAGE_TAG
endif

# Determine the arch/os
ifeq (${XC_OS}, )
  XC_OS:=$(shell go env GOOS)
endif
export XC_OS
ifeq (${XC_ARCH}, )
  XC_ARCH:=$(shell go env GOARCH)
endif
export XC_ARCH
ARCH:=${XC_OS}_${XC_ARCH}
export ARCH

export DBUILD_ARGS=--build-arg DBUILD_DATE=${DBUILD_DATE} --build-arg DBUILD_REPO_URL=${DBUILD_REPO_URL} --build-arg DBUILD_SITE_URL=${DBUILD_SITE_URL} --build-arg BRANCH=${BRANCH} --build-arg RELEASE_TAG=${RELEASE_TAG}

.PHONY: all
all: license-check golint test manifests populator-images

## golangci-lint tool used to check linting tools in codebase
## Example: golangci-lint document is not recommending
##			to use `go get <path>`. For more info:
##          https://golangci-lint.run/usage/install/#install-from-source
##
## Install golangci-lint only if tool doesn't exist in system
.PHONY: install-golangci-lint
install-golangci-lint:
	$(if $(shell which golangci-lint), echo "golangci-lint already exist in system", (curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sudo sh -s -- -b "${GOPATH}/bin" v1.40.1))

.PHONY: controller-gen
controller-gen:
	TMP_DIR=$(shell mktemp -d) && cd $$TMP_DIR && go mod init tmp && go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.5.0 && rm -rf $$TMP_DIR;

## Currently we are running with Default options + other options
## Explanation for explicitly mentioned linters:
## exportloopref: checks for pointers to enclosing loop variables
## revive: Drop-in replacement of golint. It allows to enable or disable
##         rules using configuration file.
## bodyclose: checks whether HTTP response body is closed successfully
## goconst: Find repeated strings that could be replaced by a constant
## misspell: Finds commonly misspelled English words in comments
##
## NOTE: Disabling structcheck since it is reporting false positive cases
##       for more information look at https://github.com/golangci/golangci-lint/issues/537
.PHONY: golint
golint:
	@echo "--> Running golint"
	golangci-lint run -E exportloopref,revive,bodyclose,goconst,misspell -D structcheck --timeout 5m0s
	@echo "Completed golangci-lint no recommendations !!"
	@echo "--------------------------------"
	@echo ""

.PHONY: deps
deps:
	@echo "--> Tidying up submodules"
	@go mod tidy
	@echo "--> Verifying submodules"
	@go mod verify

.PHONY: verify-deps
verify-deps: deps
	@if !(git diff --quiet HEAD -- go.sum go.mod); then \
		echo "go module files are out of date, please commit the changes to go.mod and go.sum"; exit 1; \
	fi

.PHONY: crd-gen
crd-gen:
	rm -rf deploy/crds
	controller-gen crd:trivialVersions=true,preserveUnknownFields=false paths="./apis/openebs.io/..." output:crd:artifacts:config=deploy/crds/

.PHONY: format
format:
	@echo "--> Running go fmt"
	@go fmt $(PACKAGES)

.PHONY: test
test: format
	@echo "--> Running go test" ;
	@go test $(UNIT_TEST_PACKAGES)

.PHONY: vendor
vendor: go.mod go.sum deps
	@go mod vendor

manifests:
	@echo "--------------------------------"
	@echo "+ Generating data populator crds"
	@echo "--------------------------------"
	$(PWD)/buildscripts/generate-manifests.sh

.PHONY: populator-images
populator-images: rsync-daemon-image rsync-client-image rsync-populator-image data-populator-image

.PHONY: rsync-populator
rsync-populator: format
	@echo "--------------------------------"
	@echo "--> Building ${RSYNC_POPULATOR}        "
	@echo "--------------------------------"
	mkdir -p bin
	rm -rf bin/rsync-populator
	CGO_ENABLED=0 go build -o bin/rsync-populator ./app/populator/rsync/*

.PHONY: data-populator
data-populator: format
	@echo "--------------------------------"
	@echo "--> Building ${DATA_POPULATOR}        "
	@echo "--------------------------------"
	mkdir -p bin
	rm -rf bin/data-populator
	CGO_ENABLED=0 go build -o bin/data-populator ./app/populator/data/

.PHONY: rsync-populator-image
rsync-populator-image: rsync-populator
	@echo "--------------------------------"
	@echo "+ Generating ${RSYNC_POPULATOR} image"
	@echo "--------------------------------"
	sudo docker build -t ${IMAGE_ORG}/${RSYNC_POPULATOR}:${IMAGE_TAG} ${DBUILD_ARGS} -f buildscripts/populator/rsync/Dockerfile . && sudo docker tag ${IMAGE_ORG}/${RSYNC_POPULATOR}:${IMAGE_TAG} quay.io/${IMAGE_ORG}/${RSYNC_POPULATOR}:${IMAGE_TAG}

.PHONY: data-populator-image
data-populator-image: data-populator
	@echo "--------------------------------"
	@echo "+ Generating ${DATA_POPULATOR} image"
	@echo "--------------------------------"
	sudo docker build -t ${IMAGE_ORG}/${DATA_POPULATOR}:${IMAGE_TAG} ${DBUILD_ARGS} -f buildscripts/populator/data/Dockerfile . && sudo docker tag ${IMAGE_ORG}/${DATA_POPULATOR}:${IMAGE_TAG} quay.io/${IMAGE_ORG}/${DATA_POPULATOR}:${IMAGE_TAG}

.PHONY: rsync-daemon-image
rsync-daemon-image:
	@echo "--------------------------------"
	@echo "+ Generating ${RSYNC_DAEMON} image"
	@echo "--------------------------------"
	sudo docker build -t ${IMAGE_ORG}/${RSYNC_DAEMON}:${IMAGE_TAG} ${DBUILD_ARGS} -f buildscripts/rsync/daemon/Dockerfile . && sudo docker tag ${IMAGE_ORG}/${RSYNC_DAEMON}:${IMAGE_TAG} quay.io/${IMAGE_ORG}/${RSYNC_DAEMON}:${IMAGE_TAG}

.PHONY: rsync-client-image
rsync-client-image:
	@echo "--------------------------------"
	@echo "+ Generating ${RSYNC_CLIENT} image"
	@echo "--------------------------------"
	sudo docker build -t ${IMAGE_ORG}/${RSYNC_CLIENT}:${IMAGE_TAG} ${DBUILD_ARGS} -f buildscripts/rsync/client/Dockerfile . && sudo docker tag ${IMAGE_ORG}/${RSYNC_CLIENT}:${IMAGE_TAG} quay.io/${IMAGE_ORG}/${RSYNC_CLIENT}:${IMAGE_TAG}

.PHONY: license-check
license-check:
	@echo "--> Checking license header..."
	@licRes=$$(for file in $$(find . -type f -regex '.*\.sh\|.*\.go\|.*Docker.*\|.*\Makefile*' ! -path './vendor/*' ) ; do \
               awk 'NR<=5' $$file | grep -Eq "(Copyright|generated|GENERATED)" || echo $$file; \
       done); \
       if [ -n "$${licRes}" ]; then \
               echo "license header checking failed:"; echo "$${licRes}"; \
               exit 1; \
       fi
	@echo "--> Done checking license."
	@echo
