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

# Pin tool versions (bump intentionally). GO_VERSION must match common.mk.
# Keep kubectl aligned with the k8s.io/* modules in go.mod and KIND_K8S_TAG
# in demo/scripts/common.sh.
ARG GO_VERSION=1.26.2
ARG GOLANGCI_LINT_VERSION=2.13.2
ARG KUBECTL_VERSION=1.37.0
ARG HELM_VERSION=4.3.0
ARG KIND_VERSION=0.33.0
ARG YQ_VERSION=4.53.6
ARG GH_VERSION=2.101.0

ENV LANG=C.UTF-8 \
    LC_ALL=C.UTF-8 \
    HOME=/home/${USERNAME} \
    GOPATH=/home/${USERNAME}/go \
    GOCACHE=/home/${USERNAME}/.cache/go-build \
    GOTOOLCHAIN=local \
    CGO_ENABLED=1 \
    PATH=/usr/local/go/bin:/home/${USERNAME}/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

USER root

RUN dnf -y upgrade --refresh \
    && dnf -y install \
        bash-completion \
        ca-certificates \
        curl \
        diffutils \
        findutils \
        gcc \
        git \
        glibc-devel \
        jq \
        less \
        make \
        openssh-clients \
        procps-ng \
        sudo \
        tar \
        unzip \
        vim-enhanced \
        which \
        wget \
        xz \
    && dnf clean all \
    && rm -rf /var/cache/dnf

RUN groupadd --gid "${USER_GID}" "${USERNAME}" \
    && useradd --uid "${USER_UID}" --gid "${USER_GID}" -m -s /bin/bash "${USERNAME}" \
    && echo "${USERNAME} ALL=(root) NOPASSWD:ALL" > "/etc/sudoers.d/${USERNAME}" \
    && chmod 0440 "/etc/sudoers.d/${USERNAME}"

# Resolve arch once for release assets (Docker TARGETARCH is amd64/arm64).
RUN set -euo pipefail; \
    arch="${TARGETARCH:-}"; \
    if [ -z "${arch}" ]; then \
      case "$(uname -m)" in \
        x86_64) arch=amd64 ;; \
        aarch64|arm64) arch=arm64 ;; \
        *) echo "unsupported arch: $(uname -m)" >&2; exit 1 ;; \
      esac; \
    fi; \
    case "${arch}" in \
      amd64|arm64) echo "${arch}" > /tmp/arch ;; \
      *) echo "unsupported TARGETARCH: ${arch}" >&2; exit 1 ;; \
    esac

# Go. Official tarball, not the Fedora package, so the pin matches common.mk.
# gcc/glibc-devel above are for `go test -race` (cgo).
RUN set -euo pipefail; \
    arch="$(cat /tmp/arch)"; \
    curl -fsSL -o /tmp/go.tgz \
      "https://go.dev/dl/go${GO_VERSION}.linux-${arch}.tar.gz"; \
    rm -rf /usr/local/go; \
    tar -C /usr/local -xzf /tmp/go.tgz; \
    rm -f /tmp/go.tgz

# golangci-lint
RUN set -euo pipefail; \
    arch="$(cat /tmp/arch)"; \
    curl -fsSL -o /tmp/golangci-lint.tar.gz \
      "https://github.com/golangci/golangci-lint/releases/download/v${GOLANGCI_LINT_VERSION}/golangci-lint-${GOLANGCI_LINT_VERSION}-linux-${arch}.tar.gz"; \
    tar -xzf /tmp/golangci-lint.tar.gz -C /tmp; \
    mv "/tmp/golangci-lint-${GOLANGCI_LINT_VERSION}-linux-${arch}/golangci-lint" /usr/local/bin/golangci-lint; \
    chmod +x /usr/local/bin/golangci-lint; \
    rm -rf /tmp/golangci-lint.tar.gz "/tmp/golangci-lint-${GOLANGCI_LINT_VERSION}-linux-${arch}"

# kubectl
RUN set -euo pipefail; \
    arch="$(cat /tmp/arch)"; \
    curl -fsSL -o /usr/local/bin/kubectl \
      "https://dl.k8s.io/release/v${KUBECTL_VERSION}/bin/linux/${arch}/kubectl"; \
    chmod +x /usr/local/bin/kubectl

# Helm
RUN set -euo pipefail; \
    arch="$(cat /tmp/arch)"; \
    curl -fsSL -o /tmp/helm.tar.gz \
      "https://get.helm.sh/helm-v${HELM_VERSION}-linux-${arch}.tar.gz"; \
    tar -xzf /tmp/helm.tar.gz -C /tmp; \
    mv "/tmp/linux-${arch}/helm" /usr/local/bin/helm; \
    chmod +x /usr/local/bin/helm; \
    rm -rf /tmp/helm.tar.gz "/tmp/linux-${arch}"

# kind (binary only; creating a cluster still needs a host container engine)
RUN set -euo pipefail; \
    arch="$(cat /tmp/arch)"; \
    curl -fsSL -o /usr/local/bin/kind \
      "https://github.com/kubernetes-sigs/kind/releases/download/v${KIND_VERSION}/kind-linux-${arch}"; \
    chmod +x /usr/local/bin/kind

# GitHub CLI
RUN set -euo pipefail; \
    arch="$(cat /tmp/arch)"; \
    curl -fsSL -o /tmp/gh.tar.gz \
      "https://github.com/cli/cli/releases/download/v${GH_VERSION}/gh_${GH_VERSION}_linux_${arch}.tar.gz"; \
    tar -xzf /tmp/gh.tar.gz -C /tmp; \
    mv "/tmp/gh_${GH_VERSION}_linux_${arch}/bin/gh" /usr/local/bin/gh; \
    chmod +x /usr/local/bin/gh; \
    rm -rf /tmp/gh.tar.gz "/tmp/gh_${GH_VERSION}_linux_${arch}"

# yq
RUN set -euo pipefail; \
    arch="$(cat /tmp/arch)"; \
    curl -fsSL -o /usr/local/bin/yq \
      "https://github.com/mikefarah/yq/releases/download/v${YQ_VERSION}/yq_linux_${arch}"; \
    chmod +x /usr/local/bin/yq

# Pre-create ~/.config as the container user. Bind-mounting ~/.config/gh
# otherwise makes the engine create ~/.config as root.
RUN mkdir -p \
        /home/${USERNAME}/.local/bin \
        /home/${USERNAME}/go/bin \
        /home/${USERNAME}/.cache/go-build \
        /home/${USERNAME}/.config \
    && printf '%s\n' \
        'export PATH="/usr/local/go/bin:$HOME/go/bin:$PATH"' \
        >> /home/${USERNAME}/.bashrc \
    && chown -R ${USERNAME}:${USERNAME} /home/${USERNAME}

COPY scripts/post-create.sh /usr/local/share/rockchip-dra-devcontainer/post-create.sh
COPY scripts/post-start.sh /usr/local/share/rockchip-dra-devcontainer/post-start.sh
RUN chmod +x /usr/local/share/rockchip-dra-devcontainer/*.sh \
    && chown root:root /usr/local/bin/* \
    && chmod 755 /usr/local/bin/* \
    && rm -f /tmp/arch

USER ${USERNAME}
WORKDIR /workspaces/rockchip-dra-driver

RUN go version \
    && golangci-lint version \
    && kubectl version --client \
    && helm version \
    && kind version \
    && gh --version \
    && yq --version \
    && ssh -V \
    && gcc --version
