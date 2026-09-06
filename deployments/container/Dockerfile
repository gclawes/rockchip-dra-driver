# Copyright 2023 The Kubernetes Authors.
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

ARG GO_VERSION=undefined
ARG BASE_IMAGE=undefined
FROM golang:${GO_VERSION} AS build

WORKDIR /build
COPY . .

RUN mkdir /artifacts
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    make PREFIX=/artifacts cmds

FROM ${BASE_IMAGE}

LABEL io.k8s.display-name="Rockchip DRA Driver"
LABEL name="Rockchip DRA Driver"
LABEL vendor="gclawes"
ARG VERSION="N/A"
LABEL version=${VERSION}
LABEL summary="Kubernetes DRA driver for Rockchip NPU and GPU"
LABEL description="Mainline rocket NPU and panthor GPU DRA driver"

COPY --from=build /artifacts/kubeletplugin /usr/bin/kubeletplugin
