# Copyright 2020-2024 Open Networking Foundation (ONF) and the ONF Contributors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

$(if $(DEBUG),$(warning ENTER))

.PHONY: help clean help test
.DEFAULT_GOAL := help

##-------------------##
##---]  GLOBALS  [---##
##-------------------##
TOP ?=$(strip \
  $(dir \
    $(abspath $(lastword $(MAKEFILE_LIST)))\
   )\
)

##--------------------##
##---]  INCLUDES  [---##
##--------------------##
include $(TOP)/config.mk#                # configure

# -----------------------------------------------------------------------
# https://jira.opencord.org/browse/VOL-5163
#   - only a tiny subset of logic is enabled right now.
#   - several targets can also be refactored across several repos
#     into additional library targets.
# -----------------------------------------------------------------------
include $(TOP)/makefiles/include.mk      # top level include

# Variables
VERSION                    ?= $(shell cat ./VERSION)

DOCKER_LABEL_VCS_DIRTY     = false
ifneq ($(shell git ls-files --others --modified --exclude-standard 2>/dev/null | wc -l | sed -e 's/ //g'),0)
    DOCKER_LABEL_VCS_DIRTY = true
endif
## Docker related
DOCKER_EXTRA_ARGS          ?=
DOCKER_REGISTRY            ?=
DOCKER_REPOSITORY          ?=
DOCKER_TAG                 ?= ${VERSION}$(shell [[ ${DOCKER_LABEL_VCS_DIRTY} == "true" ]] && echo "-dirty" || true)
IMAGENAME           := ${DOCKER_REGISTRY}${DOCKER_REPOSITORY}bbsim-sadis-server
TYPE                       ?= minimal

## Docker labels. Only set ref and commit date if committed
DOCKER_LABEL_VCS_URL       ?= $(shell git remote get-url $(shell git remote))
DOCKER_LABEL_VCS_REF       = $(shell git rev-parse HEAD)
DOCKER_LABEL_BUILD_DATE    ?= $(shell date -u "+%Y-%m-%dT%H:%M:%SZ")
DOCKER_LABEL_COMMIT_DATE   = $(shell git show -s --format=%cd --date=iso-strict HEAD)

DOCKER_BUILD_ARGS ?= \
	${DOCKER_EXTRA_ARGS} \
	--build-arg org_label_schema_version="${VERSION}" \
	--build-arg org_label_schema_vcs_url="${DOCKER_LABEL_VCS_URL}" \
	--build-arg org_label_schema_vcs_ref="${DOCKER_LABEL_VCS_REF}" \
	--build-arg org_label_schema_build_date="${DOCKER_LABEL_BUILD_DATE}" \
	--build-arg org_opencord_vcs_commit_date="${DOCKER_LABEL_COMMIT_DATE}" \
	--build-arg org_opencord_vcs_dirty="${DOCKER_LABEL_VCS_DIRTY}"

DOCKER_BUILD_ARGS_LOCAL ?= ${DOCKER_BUILD_ARGS} \
	--build-arg LOCAL_PROTOS=${LOCAL_PROTOS}

# tool containers
VOLTHA_TOOLS_VERSION ?= 2.4.0

GO                = docker run --rm --user $$(id -u):$$(id -g) -v ${CURDIR}:/app $(shell test -t 0 && echo "-it") -v gocache:/.cache -v gocache-${VOLTHA_TOOLS_VERSION}:/go/pkg voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-golang go
GO_JUNIT_REPORT   = docker run --rm --user $$(id -u):$$(id -g) -v ${CURDIR}:/app -i voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-go-junit-report go-junit-report
GOCOVER_COBERTURA = docker run --rm --user $$(id -u):$$(id -g) -v ${CURDIR}:/app/src/github.com/opencord/bbsim-sadis-server -i voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-gocover-cobertura gocover-cobertura
GOLANGCI_LINT     = docker run --rm --user $$(id -u):$$(id -g) -v ${CURDIR}:/app $(shell test -t 0 && echo "-it") -v gocache:/.cache -v gocache-${VOLTHA_TOOLS_VERSION}:/go/pkg voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-golangci-lint golangci-lint
HADOLINT          = docker run --rm --user $$(id -u):$$(id -g) -v ${CURDIR}:/app $(shell test -t 0 && echo "-it") voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-hadolint hadolint

help ::
	@echo "TODO write the help"

.PHONY: build
build:
	$(call banner-enter,Target $@)
	@${GO} build -mod=vendor ./cmd/bbsim-sadis-server.go
	$(call banner-leave,Target $@)

.PHONY: build-local
build-local:
	@go build -mod=vendor ./cmd/bbsim-sadis-server.go

## Docker targets
.PHONY: docker-build
docker-build:
	docker build $(DOCKER_BUILD_ARGS) -t ${IMAGENAME}:${DOCKER_TAG} -f build/package/Dockerfile .

.PHONY: docker-push
docker-push:
	docker push ${IMAGENAME}:${DOCKER_TAG}

.PHONY: docker-kind-load
docker-kind-load:
	@if [ "`kind get clusters | grep voltha-$(TYPE)`" = '' ]; then echo "no voltha-$(TYPE) cluster found" && exit 1; fi
	kind load docker-image ${IMAGENAME}:${DOCKER_TAG} --name=voltha-$(TYPE) --nodes $(shell kubectl get nodes --template='{{range .items}}{{.metadata.name}},{{end}}' | rev | cut -c 2- | rev)

## lint and unit tests

.PHONY: lint-dockerfile
lint-dockerfile:
	@echo "Running Dockerfile lint check..."
	@${HADOLINT} $$(find ./build -name "Dockerfile*")
	@echo "Dockerfile lint check OK"

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
lint-mod:
	$(call banner-entry,Target $@)
	@echo "Running dependency check..."
	@${GO} mod verify
	@echo "Dependency check OK. Running vendor check..."
	@git status > /dev/null
	@git diff-index --quiet HEAD -- go.mod go.sum vendor || (echo "ERROR: Staged or modified files must be committed before running this test" && git status -- go.mod go.sum vendor && exit 1)
	@[[ `git ls-files --exclude-standard --others go.mod go.sum vendor` == "" ]] || (echo "ERROR: Untracked files must be cleaned up before running this test" && git status -- go.mod go.sum vendor && exit 1)

	$(HIDE)$(MAKE) --no-print-directory mod-update

	@git status > /dev/null
	@git diff-index --quiet HEAD -- go.mod go.sum vendor || (echo "ERROR: Modified files detected after running go mod tidy / go mod vendor" && git status -- go.mod go.sum vendor && git checkout -- go.mod go.sum vendor && exit 1)
	@[[ `git ls-files --exclude-standard --others go.mod go.sum vendor` == "" ]] || (echo "ERROR: Untracked files detected after running go mod tidy / go mod vendor" && git status -- go.mod go.sum vendor && git checkout -- go.mod go.sum vendor && exit 1)
	@echo "Vendor check OK."
	$(call banner-leave,Target $@)


## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
.PHONY: mod-update
mod-update: go-version mod-tidy mod-vendor

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
.PHONY: go-version
go-version :
	$(call banner-enter,Target $@)
	${GO} version
	$(call banner-leave,Target $@)

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
.PHONY: mod-tidy
mod-tidy :
	$(call banner-enter,Target $@)
	${GO} mod tidy
	$(call banner-leave,Target $@)

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
.PHONY: mod-vendor
mod-vendor : mod-tidy
mod-vendor :
	$(call banner-enter,Target $@)
	$(if $(LOCAL_FIX_PERMS),chmod o+w $(CURDIR))
	${GO} mod vendor
	$(if $(LOCAL_FIX_PERMS),chmod o-w $(CURDIR))
	$(call banner-leave,Target $@)

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
lint: lint-mod lint-dockerfile

## -----------------------------------------------------------------------
## Coverage report: Static code analysis
## -----------------------------------------------------------------------
sca:
	@rm -rf ./sca-report
	@mkdir -p ./sca-report
	@echo "Running static code analysis..."
	@${GOLANGCI_LINT} run --deadline=6m --out-format junit-xml ./... | tee ./sca-report/sca-report.xml
	@echo ""
	@echo "Static code analysis OK"

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
test:
	@mkdir -p ./tests/results
	@${GO} test -mod=vendor -v -coverprofile ./tests/results/go-test-coverage.out -covermode count ./... 2>&1 | tee ./tests/results/go-test-results.out ;\
	RETURN=$$? ;\
	${GO_JUNIT_REPORT} < ./tests/results/go-test-results.out > ./tests/results/go-test-results.xml ;\
	${GOCOVER_COBERTURA} < ./tests/results/go-test-coverage.out > ./tests/results/go-test-coverage.xml ;\
	exit $$RETURN

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
distclean:
	$(RM) -r ./sca-report

# [EOF]
