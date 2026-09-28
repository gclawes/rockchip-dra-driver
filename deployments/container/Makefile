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
DOCKERFILE := $(CURDIR)/deployments/container/Dockerfile
SELF_MAKEFILE := $(lastword $(MAKEFILE_LIST))
COMMA := ,

include $(CURDIR)/common.mk

ifeq ($(strip $(VERSION)),)
VERSION := latest
endif

IMAGE_TAG := $(VERSION)
IMAGE = $(IMAGE_NAME):$(IMAGE_TAG)
BASE_IMAGE ?= gcr.io/distroless/static-debian13

# Rockchip nodes are arm64. A comma-separated list publishes an index of ARM
# platforms (for example linux/arm64,linux/arm/v7). amd64 is not a target.
PLATFORMS ?= linux/arm64

BUILDX_BUILDER ?= rockchip-dra-driver-builder
# mode=min is the useful attestation without embedding full build args.
# Attestations are attached on push only. buildx uploads those blobs;
# a later "docker push" does not, and GHCR rejects the manifest with
# "unknown blob".
BUILDX_PROVENANCE ?= mode=min
BUILDX_SBOM ?= true

IMAGE_BUILD_ARGS = \
	--build-arg GO_VERSION="$(GO_VERSION)" \
	--build-arg BASE_IMAGE="$(BASE_IMAGE)" \
	--build-arg VERSION="$(VERSION)"

IMAGE_BUILD_FLAGS = \
	--pull \
	--tag $(IMAGE) \
	$(IMAGE_BUILD_ARGS) \
	-f $(DOCKERFILE) \
	$(CURDIR)

BUILDX_BUILD_FLAGS = \
	$(IMAGE_BUILD_FLAGS) \
	--builder $(BUILDX_BUILDER) \
	--platform $(PLATFORMS)

# Passed only to registry pushes. --load cannot store attestations.
BUILDX_ATTEST_FLAGS = \
	--provenance=$(BUILDX_PROVENANCE) \
	--sbom=$(BUILDX_SBOM)

.PHONY: build push check-platforms ensure-buildx-builder

check-platforms:
	@if [ -z "$(PLATFORMS)" ]; then echo "PLATFORMS is empty"; exit 1; fi
	@for p in $(subst $(COMMA), ,$(PLATFORMS)); do \
		case "$$p" in \
			linux/arm*) ;; \
			*) echo "unsupported platform $$p: only linux/arm* is published"; exit 1 ;; \
		esac; \
	done

# Load a single platform into the local engine. Provenance is omitted because
# the Docker image store cannot keep attestations.
build: check-platforms
	@if echo "$(PLATFORMS)" | grep -q ','; then \
		echo "local build supports one platform; push to publish $(PLATFORMS)"; \
		exit 1; \
	fi
ifeq ($(CONTAINER_TOOL),docker)
	$(MAKE) -f $(SELF_MAKEFILE) ensure-buildx-builder
	$(CONTAINER_TOOL) buildx build $(BUILDX_BUILD_FLAGS) --load --provenance=false
else
	$(CONTAINER_TOOL) build --platform $(PLATFORMS) $(IMAGE_BUILD_FLAGS)
endif

# Always push with buildx so provenance and SBOM blobs are uploaded with the
# manifest. A comma-separated PLATFORMS value publishes an ARM image index.
push: check-platforms
	@if [ "$(CONTAINER_TOOL)" != "docker" ]; then \
		echo "image push requires docker buildx, got $(CONTAINER_TOOL)"; \
		exit 1; \
	fi
	@echo "Pushing $(PLATFORMS) with provenance=$(BUILDX_PROVENANCE) sbom=$(BUILDX_SBOM)"
	$(MAKE) -f $(SELF_MAKEFILE) ensure-buildx-builder
	$(CONTAINER_TOOL) buildx build $(BUILDX_BUILD_FLAGS) $(BUILDX_ATTEST_FLAGS) --push

ensure-buildx-builder:
ifeq ($(CONTAINER_TOOL),docker)
	@docker buildx inspect $(BUILDX_BUILDER) >/dev/null 2>&1 || \
		docker buildx create --name $(BUILDX_BUILDER) --driver docker-container
	@docker buildx inspect --bootstrap $(BUILDX_BUILDER) >/dev/null
else
	@echo "ensure-buildx-builder requires CONTAINER_TOOL=docker"
	@exit 1
endif
