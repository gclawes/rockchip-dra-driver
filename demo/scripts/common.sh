#!/usr/bin/env bash

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

SCRIPTS_DIR="$(cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd)"

: ${DRIVER_NAME:=rockchip-dra-driver}
: ${DRIVER_IMAGE_REGISTRY:="ghcr.io/gclawes"}
: ${DRIVER_IMAGE_NAME:="${DRIVER_NAME}"}
: ${DRIVER_IMAGE_TAG:="$(git describe --tags --always --dirty)"}
: ${DRIVER_IMAGE:="${DRIVER_IMAGE_REGISTRY}/${DRIVER_IMAGE_NAME}:${DRIVER_IMAGE_TAG}"}
: ${KIND_CLUSTER_NAME:="${DRIVER_NAME}-cluster"}
: ${KIND_K8S_TAG:="v1.37.0"}
: ${KIND_IMAGE:="kindest/node:${KIND_K8S_TAG}"}

if [[ -z "${CONTAINER_TOOL:-}" ]]; then
    if command -v docker >/dev/null 2>&1; then
        CONTAINER_TOOL=docker
    elif command -v podman >/dev/null 2>&1; then
        CONTAINER_TOOL=podman
    else
        echo "No container tool detected. Please install Docker or Podman." >&2
        return 1
    fi
fi

: ${KIND:="env KIND_EXPERIMENTAL_PROVIDER=${CONTAINER_TOOL} kind"}
