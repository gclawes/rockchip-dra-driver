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

FROM docker.io/library/fedora:44

ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=1000
ARG TARGETARCH

# Go is the only pinned toolchain. The version is read from go.mod at build
# time (toolchain directive if present, otherwise the go line). Everything
# else is the current Fedora package. kubectl is not packaged; install the
# current Kubernetes stable release instead.
ENV LANG=C.UTF-8 \
    LC_ALL=C.UTF-8 \
    HOME=/home/${USERNAME} \
    GOPATH=/home/${USERNAME}/go \
    GOCACHE=/home/${USERNAME}/.cache/go-build \
    GOTOOLCHAIN=local \
    CGO_ENABLED=1 \
    PATH=/usr/local/go/bin:/home/${USERNAME}/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

USER root

RUN dnf -y install \
        ca-certificates \
        curl \
        gcc \
        gh \
        git \
        glibc-devel \
        golangci-lint \
        helm \
        jq \
        make \
        openssh-clients \
        sudo \
        tar \
        yq \
    && dnf clean all \
    && rm -rf /var/cache/dnf

RUN groupadd --gid "${USER_GID}" "${USERNAME}" \
    && useradd --uid "${USER_UID}" --gid "${USER_GID}" -m -s /bin/bash "${USERNAME}" \
    && echo "${USERNAME} ALL=(root) NOPASSWD:ALL" > "/etc/sudoers.d/${USERNAME}" \
    && chmod 0440 "/etc/sudoers.d/${USERNAME}" \
    && mkdir -p \
        /home/${USERNAME}/go/bin \
        /home/${USERNAME}/.cache/go-build \
        /home/${USERNAME}/.config \
    && printf '%s\n' 'export PATH="/usr/local/go/bin:$HOME/go/bin:$PATH"' \
        >> /home/${USERNAME}/.bashrc \
    && chown -R ${USERNAME}:${USERNAME} /home/${USERNAME}

# Build context is the repo root so this can see go.mod.
COPY go.mod /tmp/go.mod
RUN set -euo pipefail; \
    case "${TARGETARCH:-}" in \
      amd64|arm64) arch="${TARGETARCH}" ;; \
      "") case "$(uname -m)" in \
            x86_64) arch=amd64 ;; \
            aarch64|arm64) arch=arm64 ;; \
            *) echo "unsupported arch: $(uname -m)" >&2; exit 1 ;; \
          esac ;; \
      *) echo "unsupported TARGETARCH: ${TARGETARCH}" >&2; exit 1 ;; \
    esac; \
    ver="$(awk '/^toolchain go/ { print substr($2, 3); exit } /^go / { v=$2 } END { if (v != "") print v }' /tmp/go.mod)"; \
    if [ -z "${ver}" ]; then echo "go.mod has no go or toolchain version" >&2; exit 1; fi; \
    echo "installing go ${ver}"; \
    curl -fsSL -o /tmp/go.tgz "https://go.dev/dl/go${ver}.linux-${arch}.tar.gz"; \
    rm -rf /usr/local/go; \
    tar -C /usr/local -xzf /tmp/go.tgz; \
    curl -fsSL -o /usr/local/bin/kubectl \
      "https://dl.k8s.io/release/$(curl -fsSL https://dl.k8s.io/release/stable.txt)/bin/linux/${arch}/kubectl"; \
    chmod +x /usr/local/bin/kubectl; \
    rm -f /tmp/go.tgz /tmp/go.mod

COPY .devcontainer/scripts/ /usr/local/share/rockchip-dra-devcontainer/
RUN chmod +x /usr/local/share/rockchip-dra-devcontainer/*.sh

USER ${USERNAME}
WORKDIR /workspaces/rockchip-dra-driver

RUN go version \
    && command -v golangci-lint helm gh yq ssh gcc kubectl
