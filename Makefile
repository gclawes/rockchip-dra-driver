# Copyright 2023 The Kubernetes Authors.
# Copyright 2026 Graeme Lawes.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

CONTAINER_TOOL ?= docker
MKDIR    ?= mkdir
DIST_DIR ?= $(CURDIR)/dist
HELM     ?= "go run helm.sh/helm/v3/cmd/helm@latest"

export GIT_TAG ?= $(shell git describe --tags --always --dirty)

include $(CURDIR)/common.mk

CMDS := $(patsubst ./cmd/%/,%,$(sort $(dir $(wildcard ./cmd/*/))))
CMD_TARGETS := $(patsubst %,cmd-%, $(CMDS))

.PHONY: binaries cmds build test fmt assert-fmt vet lint setup-e2e test-e2e teardown-e2e push-release-artifacts

GOOS ?= linux

binaries: cmds
ifneq ($(PREFIX),)
cmd-%: COMMAND_BUILD_OPTIONS = -o $(PREFIX)/$(*)
endif
cmds: $(CMD_TARGETS)
$(CMD_TARGETS): cmd-%:
	CGO_LDFLAGS_ALLOW='-Wl,--unresolved-symbols=ignore-in-object-files' GOOS=$(GOOS) \
		go build -ldflags "-s -w -X main.version=$(VERSION)" $(COMMAND_BUILD_OPTIONS) $(MODULE)/cmd/$(*)

build:
	GOOS=$(GOOS) go build ./...

fmt:
	go list -f '{{.Dir}}' $(MODULE)/... \
		| xargs gofmt -s -l -w

assert-fmt:
	go list -f '{{.Dir}}' $(MODULE)/... \
		| xargs gofmt -s -l > fmt.out
	@if [ -s fmt.out ]; then \
		echo "\nERROR: The following files are not formatted:\n"; \
		cat fmt.out; \
		rm fmt.out; \
		exit 1; \
	else \
		rm fmt.out; \
	fi

lint:
	golangci-lint run --build-tags=e2e ./...

vet:
	go vet $(MODULE)/...

test:
	go test -v -race ./...

setup-e2e:
	test/e2e/setup-e2e.sh

test-e2e:
	test/e2e/test-e2e.sh

teardown-e2e:
	test/e2e/teardown-e2e.sh

.PHONY: push-release-artifacts
push-release-artifacts:
	CHART_VERSION="${GIT_TAG}" \
		HELM=$(HELM) \
		demo/scripts/push-driver-chart.sh
	export DRIVER_IMAGE_TAG="${GIT_TAG}"; \
	demo/scripts/push-driver-image.sh
