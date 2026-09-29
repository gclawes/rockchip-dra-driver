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

root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
script="${root}/hack/next-release-version.sh"
checker="${root}/hack/check-release-train.sh"

fail() {
	echo "FAIL: $*" >&2
	exit 1
}

assert_eq() {
	local got="$1" want="$2" label="$3"
	if [ "$got" != "$want" ]; then
		fail "${label}: expected '${want}', got '${got}'"
	fi
}

assert_rc() {
	local got="$1" want="$2" label="$3"
	if [ "$got" -ne "$want" ]; then
		fail "${label}: expected exit ${want}, got ${got}"
	fi
}

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

git -C "$tmp" init -b master >/dev/null
git -C "$tmp" config user.email test@example.com
git -C "$tmp" config user.name test
git -C "$tmp" commit --allow-empty -m "chore: init" >/dev/null
git -C "$tmp" tag -a 0.6.0 -m 0.6.0

commit_on() {
	local branch="$1" message="$2"
	git -C "$tmp" checkout -q "$branch"
	git -C "$tmp" commit --allow-empty -m "$message" >/dev/null
}

run_version() {
	local base_name="$1"
	git -C "$tmp" checkout -q prep
	(
		cd "$tmp"
		bash "$script" --base-ref master --base-name "$base_name" --head-ref HEAD
	)
}

# feat on the train, and a fix already on master since the tag, is still a minor.
commit_on master "fix: already on master"
git -C "$tmp" checkout -q -b prep 0.6.0
commit_on prep "feat(vpu): advertise decoder nodes"
got=$(run_version master)
assert_eq "$got" "0.7.0" "feat plus an unreleased fix on master"

# fix only.
git -C "$tmp" checkout -q master
git -C "$tmp" reset --hard 0.6.0 >/dev/null
git -C "$tmp" checkout -q -B prep 0.6.0
commit_on prep "fix(npu): keep the node gid"
got=$(run_version master)
assert_eq "$got" "0.6.1" "fix only"

# feat and fix on the train: highest bump wins.
commit_on prep "feat(rga): advertise rockchip-rga"
got=$(run_version master)
assert_eq "$got" "0.7.0" "feat and fix"

# docs do not publish.
git -C "$tmp" checkout -q -B prep 0.6.0
commit_on prep "docs: open release train 0.7.0"
got=$(run_version master)
assert_eq "$got" "none" "docs only"

# perf and revert are patches.
git -C "$tmp" checkout -q -B prep 0.6.0
commit_on prep "perf(discovery): skip unbound nodes"
got=$(run_version master)
assert_eq "$got" "0.6.1" "perf"
commit_on prep "revert: undo the skip"
got=$(run_version master)
assert_eq "$got" "0.6.1" "revert"

# 0.x breaking changes are refused.
git -C "$tmp" checkout -q -B prep 0.6.0
commit_on prep "$(printf 'feat(api)!: rename the attribute\n\nBREAKING CHANGE: attribute name changed\n')"
set +e
got=$(run_version master 2>"$tmp/err")
rc=$?
set -e
assert_rc "$rc" 2 "breaking on 0.x"
grep -q "1.0.0" "$tmp/err" || fail "breaking error should mention 1.0.0"

# A maintenance line only accepts patches.
git -C "$tmp" checkout -q -B prep 0.6.0
commit_on prep "feat(vpu): advertise decoder nodes"
set +e
got=$(run_version 0.6.x 2>"$tmp/err")
rc=$?
set -e
assert_rc "$rc" 3 "feat on a maintenance base"
grep -q "only publishes patches" "$tmp/err" || fail "maintenance error should mention patches"

git -C "$tmp" checkout -q -B prep 0.6.0
commit_on prep "fix(npu): keep the node gid"
got=$(run_version 1.2.x)
assert_eq "$got" "0.6.1" "fix on a maintenance base"

# The checker ignores anything that is not a train pull request.
RELEASE_TRAIN_EVENT=push RELEASE_TRAIN_HEAD= RELEASE_TRAIN_BASE= \
	bash "$checker" >/dev/null
RELEASE_TRAIN_EVENT=pull_request RELEASE_TRAIN_HEAD=feat/vpu RELEASE_TRAIN_BASE=master \
	bash "$checker" >/dev/null

# A train whose version does not match the branch name fails. The checker
# reads the branch from the environment, and the commits from HEAD, so run
# it inside the fixture repo.
git -C "$tmp" checkout -q -B prep 0.6.0
commit_on prep "fix(npu): keep the node gid"
(
	cd "$tmp"
	set +e
	RELEASE_TRAIN_EVENT=pull_request RELEASE_TRAIN_HEAD=prep/0.7.0 RELEASE_TRAIN_BASE=master \
		bash "$checker" >"$tmp/out" 2>"$tmp/err"
	rc=$?
	set -e
	assert_rc "$rc" 1 "mismatched train version"
	grep -q "would publish 0.6.1" "$tmp/err" || fail "mismatch error should name 0.6.1: $(cat "$tmp/err")"
)

echo "ok"
