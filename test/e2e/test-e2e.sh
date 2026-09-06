#!/usr/bin/env bash

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

echo "waiting for ResourceSlices from dra.rockchip.com"
for _ in $(seq 1 60); do
	if kubectl get resourceslice -o json | grep -q '"driver": "dra.rockchip.com"'; then
		break
	fi
	sleep 2
done

kubectl get resourceslice -o yaml

if ! kubectl get resourceslice -o json | grep -q '"type"'; then
	echo "ResourceSlice missing type attribute" >&2
	exit 1
fi

if ! kubectl get deviceclass npu.rockchip.com >/dev/null; then
	echo "missing DeviceClass npu.rockchip.com" >&2
	exit 1
fi
if ! kubectl get deviceclass gpu.rockchip.com >/dev/null; then
	echo "missing DeviceClass gpu.rockchip.com" >&2
	exit 1
fi

echo "e2e checks passed"
