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

# Refresh the Included list on the open release pull request for a train
# branch. Uses the API only. Do not check out the pull request that just
# merged: this runs from pull_request_target.
#
# Environment:
#   GITHUB_REPOSITORY  owner/name
#   TRAIN_REF          base branch the pull request merged into (prep/0.7.0)
#   GH_TOKEN           token that can edit pull requests

set -euo pipefail

repo="${GITHUB_REPOSITORY:?GITHUB_REPOSITORY is required}"
train="${TRAIN_REF:?TRAIN_REF is required}"
root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)

case "$train" in
prep/*) ;;
*)
	echo "not a release train: $train"
	exit 0
	;;
esac

release=$(gh pr list --repo "$repo" --head "$train" --state open --limit 20 \
	--json number,title,body \
	--jq '.[0] // empty')
if [ -z "$release" ]; then
	echo "no open pull request from $train"
	exit 0
fi
number=$(printf '%s' "$release" | jq -r '.number')
body=$(printf '%s' "$release" | jq -r '.body // ""')

items=$(mktemp)
trap 'rm -f "$items"' EXIT
gh pr list --repo "$repo" --base "$train" --state merged --limit 100 \
	--json number,title,mergedAt \
	--jq 'sort_by(.mergedAt) | .[] | "- #\(.number) \(.title)"' >"$items"

updated=$(printf '%s\n' "$body" | bash "${root}/hack/render-release-included.sh" "$items")
if [ "$updated" = "$body" ]; then
	echo "included list unchanged on #${number}"
	exit 0
fi

body_file=$(mktemp)
trap 'rm -f "$items" "$body_file"' EXIT
printf '%s\n' "$updated" >"$body_file"
gh pr edit "$number" --repo "$repo" --body-file "$body_file"
echo "updated included list on #${number}"
