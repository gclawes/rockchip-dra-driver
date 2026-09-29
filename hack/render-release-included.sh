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

# Rewrite the Included section of a release-train pull request body.
# The body is read from stdin. Included lines are read from the first
# argument, one Markdown bullet per line. The new body is written to stdout.
#
# Markers <!-- included:start --> and <!-- included:end --> bound the
# section when present. Otherwise the ## Included heading through the next
# heading is replaced, so an older body still updates.

set -euo pipefail

items_file="${1:?included items file is required}"
start='<!-- included:start -->'
end='<!-- included:end -->'

section=$(
	printf '%s\n## Included\n\n' "$start"
	if [ -s "$items_file" ]; then
		cat "$items_file"
		printf '\n'
	else
		printf -- '-\n\n'
	fi
	printf '%s\n' "$end"
)

body=$(cat)
if printf '%s\n' "$body" | grep -qF "$start" && printf '%s\n' "$body" | grep -qF "$end"; then
	awk -v section="$section" -v start="$start" -v end="$end" '
		$0 == start { skip = 1; printf "%s\n\n", section; next }
		skip && $0 == end { skip = 0; next }
		!skip { print }
	' <<<"$body"
	exit 0
fi

if printf '%s\n' "$body" | grep -qx '## Included'; then
	awk -v section="$section" '
		$0 == "## Included" && !done { skip = 1; printf "%s\n\n", section; next }
		skip && /^## / { skip = 0; done = 1 }
		!skip { print }
	' <<<"$body"
	exit 0
fi

# No section yet. Insert it before the next heading after the intro, or append.
awk -v section="$section" '
	!inserted && /^## / && seen { printf "%s\n\n", section; inserted = 1 }
	/^## / { seen = 1 }
	{ print }
	END { if (!inserted) printf "\n%s\n\n", section }
' <<<"$body"
