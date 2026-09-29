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

# Print the version semantic-release would publish if head were merged into
# base, or "none" if those commits would not publish.
#
# The range is every commit on base since its latest tag, plus every commit
# on head that is not on base. That is the set master (or a maintenance
# branch) will contain after the merge.
#
# Exit 2: a breaking commit would publish 1.0.0 while the latest tag is 0.x.
# Exit 3: a maintenance base contains a feat or breaking commit. Those
# branches only publish patches.

set -euo pipefail

usage() {
	cat <<'EOF'
Usage: next-release-version.sh --base-ref REF --base-name NAME --head-ref REF

Prints the next version, or "none".
EOF
}

base_ref=""
base_name=""
head_ref=""

while [ $# -gt 0 ]; do
	case "$1" in
	--base-ref)
		base_ref="${2:-}"
		shift 2
		;;
	--base-name)
		base_name="${2:-}"
		shift 2
		;;
	--head-ref)
		head_ref="${2:-}"
		shift 2
		;;
	-h | --help)
		usage
		exit 0
		;;
	*)
		echo "unknown argument: $1" >&2
		usage >&2
		exit 1
		;;
	esac
done

if [ -z "$base_ref" ] || [ -z "$base_name" ] || [ -z "$head_ref" ]; then
	usage >&2
	exit 1
fi

rank() {
	case "$1" in
	none) echo 0 ;;
	patch) echo 1 ;;
	minor) echo 2 ;;
	major) echo 3 ;;
	*)
		echo "unknown release level: $1" >&2
		exit 1
		;;
	esac
}

higher() {
	if [ "$(rank "$1")" -ge "$(rank "$2")" ]; then
		echo "$1"
	else
		echo "$2"
	fi
}

# Angular preset: feat is minor, fix/perf/revert are patch, a breaking
# marker is major. Anything else does not publish.
classify_message() {
	local msg="$1"
	local subject type breaking=0
	subject=$(printf '%s\n' "$msg" | head -n 1)
	if printf '%s\n' "$msg" | grep -Eq '^BREAKING[- ]CHANGE:'; then
		breaking=1
	fi
	# Stored in a variable so "!" is not parsed as a [[ negation.
	local conventional_re='^(feat|fix|perf|revert|docs|chore|ci|test|refactor|style)(\([^)]+\))?(!)?:[[:space:]]+'
	if [[ "$subject" =~ $conventional_re ]]; then
		type="${BASH_REMATCH[1]}"
		if [ -n "${BASH_REMATCH[3]:-}" ]; then
			breaking=1
		fi
	else
		echo none
		return
	fi
	if [ "$breaking" -eq 1 ]; then
		echo major
		return
	fi
	case "$type" in
	feat) echo minor ;;
	fix | perf | revert) echo patch ;;
	*) echo none ;;
	esac
}

is_maintenance() {
	case "$1" in
	[0-9]*.x | [0-9]*.[0-9]*.x | [0-9]*.[0-9]*.[0-9]*.x) return 0 ;;
	*) return 1 ;;
	esac
}

from_version=$(git describe --tags --abbrev=0 "$base_ref")
from_version=${from_version#v}

declare -A seen=()
highest=none

consider_range() {
	local range="$1"
	local hash body level
	while read -r hash; do
		[ -z "$hash" ] && continue
		if [ -n "${seen[$hash]:-}" ]; then
			continue
		fi
		seen[$hash]=1
		body=$(git show -s --format='%B' "$hash")
		level=$(classify_message "$body")
		highest=$(higher "$highest" "$level")
	done < <(git log --format='%H' "$range")
}

consider_range "${from_version}..${base_ref}"
consider_range "${base_ref}..${head_ref}"

if is_maintenance "$base_name"; then
	case "$highest" in
	minor | major)
		echo "maintenance branch ${base_name} only publishes patches; found ${highest}" >&2
		exit 3
		;;
	esac
fi

if [[ "$from_version" == 0.* && "$highest" == major ]]; then
	echo "breaking change would publish 1.0.0 from ${from_version}; that tag is cut by a human" >&2
	exit 2
fi

if [ "$highest" = none ]; then
	echo none
	exit 0
fi

IFS=. read -r major minor patch <<<"$from_version"
case "$highest" in
major) echo "$((major + 1)).0.0" ;;
minor) echo "${major}.$((minor + 1)).0" ;;
patch) echo "${major}.${minor}.$((patch + 1))" ;;
esac
