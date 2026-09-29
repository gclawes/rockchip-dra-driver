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
script="${root}/hack/render-release-included.sh"
items=$(mktemp)
trap 'rm -f "$items"' EXIT

fail() {
	echo "FAIL: $*" >&2
	exit 1
}

render() {
	bash "$script" "$items"
}

printf '%s\n' '- #19 feat(vpu): advertise mainline decoder and encoder nodes' '- #21 feat(rga): advertise rockchip-rga' >"$items"

body='## Upcoming release

notes

## Included

- #19 old

## Not included

- gauges
'
got=$(printf '%s\n' "$body" | render)
printf '%s\n' "$got" | grep -q '^- #19 feat(vpu): advertise mainline decoder and encoder nodes$' || fail "missing vpu bullet: $got"
printf '%s\n' "$got" | grep -q '^- #21 feat(rga): advertise rockchip-rga$' || fail "missing rga bullet: $got"
printf '%s\n' "$got" | grep -q '^- #19 old$' && fail "old bullet kept: $got"
printf '%s\n' "$got" | grep -q '^- gauges$' || fail "not-included section was rewritten: $got"
printf '%s\n' "$got" | grep -q '<!-- included:start -->' || fail "markers not added: $got"

marked='<!-- included:start -->
## Included

- stale

<!-- included:end -->

## Checklist

- [ ] stay
'
got=$(printf '%s\n' "$marked" | render)
printf '%s\n' "$got" | grep -q '^- #21 feat(rga): advertise rockchip-rga$' || fail "markers were not replaced: $got"
printf '%s\n' "$got" | grep -q 'stale' && fail "stale bullet kept: $got"
printf '%s\n' "$got" | grep -q '^- \[ \] stay$' || fail "checklist was rewritten: $got"
start_count=$(printf '%s\n' "$got" | grep -c 'included:start' || true)
end_count=$(printf '%s\n' "$got" | grep -c 'included:end' || true)
[ "$start_count" = 1 ] && [ "$end_count" = 1 ] || fail "marker count start=$start_count end=$end_count"

: >"$items"
got=$(printf '%s\n' "$marked" | render)
printf '%s\n' "$got" | grep -qx '-' || fail "empty list should keep a placeholder: $got"

echo ok
