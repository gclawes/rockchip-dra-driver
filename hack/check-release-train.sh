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

# Required CI check for every pull request. Exits 0 unless the head branch
# is prep/X.Y.Z. On a train PR it fails until the branch name matches the
# version those commits would publish and docs/release-train.md is gone.
#
# Environment (set by CI):
#   RELEASE_TRAIN_EVENT  pull_request or push
#   RELEASE_TRAIN_HEAD   github.head_ref
#   RELEASE_TRAIN_BASE   github.base_ref

set -euo pipefail

event="${RELEASE_TRAIN_EVENT:-}"
head="${RELEASE_TRAIN_HEAD:-}"
base="${RELEASE_TRAIN_BASE:-}"

if [ -z "$event" ]; then
	echo "RELEASE_TRAIN_EVENT is required" >&2
	exit 1
fi

if [ "$event" != "pull_request" ]; then
	echo "not a pull request"
	exit 0
fi

case "$head" in
prep/*) ;;
*)
	echo "not a release train"
	exit 0
	;;
esac

if [ -z "$base" ]; then
	echo "RELEASE_TRAIN_BASE is required for a prep pull request" >&2
	exit 1
fi

root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
if [ -f "${root}/docs/release-train.md" ]; then
	echo "delete docs/release-train.md before merging this train" >&2
	exit 1
fi

expected=${head#prep/}
base_ref="origin/${base}"
if ! git rev-parse --verify --quiet "${base_ref}^{commit}" >/dev/null; then
	base_ref="$base"
fi

got=$(bash "${root}/hack/next-release-version.sh" \
	--base-ref "$base_ref" \
	--base-name "$base" \
	--head-ref HEAD)

if [ "$got" = none ]; then
	echo "prep/${expected} would not publish; add a feat, fix, perf, or revert, or close the train" >&2
	exit 1
fi

if [ "$got" != "$expected" ]; then
	echo "branch prep/${expected} would publish ${got}; rename the branch and retitle the pull request" >&2
	exit 1
fi

echo "prep/${expected} matches the commits"
