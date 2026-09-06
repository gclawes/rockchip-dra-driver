#!/usr/bin/env bash

# Copyright 2024 The Kubernetes Authors.
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

set -euo pipefail

ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)"
# shellcheck source=../../demo/scripts/common.sh
source "${ROOT}/demo/scripts/common.sh"

HELM_CHART_PATH="${HELM_CHART_PATH:-${ROOT}/deployments/helm/rockchip-dra-driver}"

if [[ "${HELM_CHART_PATH}" != oci://* ]]; then
	"${ROOT}/demo/build-driver.sh"
fi

if ! ${KIND} get clusters | grep -qw "${KIND_CLUSTER_NAME}"; then
	${KIND} create cluster --name "${KIND_CLUSTER_NAME}" --image "${KIND_IMAGE}"
fi

if [[ "${HELM_CHART_PATH}" != oci://* ]]; then
	${KIND} load docker-image "${DRIVER_IMAGE}" --name "${KIND_CLUSTER_NAME}"
fi

helm upgrade -i \
	--create-namespace \
	--namespace rockchip-dra-driver \
	--wait \
	--timeout 5m \
	--set mockDevices=true \
	--set image.repository="${DRIVER_IMAGE_REGISTRY}/${DRIVER_IMAGE_NAME}" \
	--set image.tag="${DRIVER_IMAGE_TAG}" \
	--set image.pullPolicy=IfNotPresent \
	rockchip-dra-driver \
	"${HELM_CHART_PATH}"
