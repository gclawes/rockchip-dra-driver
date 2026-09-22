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

# Soft checks / host integration each container start (do not fail the session).

if [[ -n "${SSH_AUTH_SOCK:-}" ]]; then
  if [[ -S "${SSH_AUTH_SOCK}" ]]; then
    if [[ ! -r "${SSH_AUTH_SOCK}" || ! -w "${SSH_AUTH_SOCK}" ]]; then
      echo "warning: SSH_AUTH_SOCK=${SSH_AUTH_SOCK} is not read/write for $(id -un)" >&2
    fi
  else
    echo "warning: SSH_AUTH_SOCK=${SSH_AUTH_SOCK} is not a socket" >&2
  fi
else
  echo "warning: SSH_AUTH_SOCK unset; ssh to nodes will not use the host agent" >&2
fi

if [[ ! -r "${HOME}/.ssh/config" && ! -d "${HOME}/.ssh" ]]; then
  echo "warning: ~/.ssh is not mounted; node debugging over ssh will fail" >&2
fi

if [[ ! -f "${HOME}/.kube/config" && -z "${KUBECONFIG:-}" ]]; then
  echo "warning: no kubeconfig at ~/.kube/config; mount host ~/.kube or set KUBECONFIG" >&2
elif [[ -f "${HOME}/.kube/config" && ! -r "${HOME}/.kube/config" ]]; then
  echo "warning: ~/.kube/config is not readable by $(id -un)" >&2
fi

# Optional host custom CA (for example a homelab *.home.arpa root).
# Missing CA is not an error; public cluster APIs do not need it.
HOST_CA="/usr/local/share/rockchip-dra-devcontainer/host/ca.home.arpa-root_ca.pem"
if [[ -s "${HOST_CA}" ]]; then
  if sudo cp "${HOST_CA}" /etc/pki/ca-trust/source/anchors/ca.home.arpa-root_ca.pem \
    && sudo update-ca-trust; then
    echo "installed host CA: ca.home.arpa-root_ca.pem"
  else
    echo "warning: failed to install host CA into trust store" >&2
  fi
fi

exit 0
