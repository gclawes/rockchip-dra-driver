# Agent instructions for rockchip-dra-driver

This file is the working contract for humans and coding agents contributing to
this repository.

## Project

`rockchip-dra-driver` is a Kubernetes Dynamic Resource Allocation (DRA) driver
for Rockchip accelerators.

- First target: RK3588 NPU + GPU.
- Kernel stack is **mainline only**: NPU is `accel/rocket` (`/dev/accel/accel*`);
  GPU is `panthor` (Mali render node). Do not add proprietary `rknpu`,
  RKNN-Toolkit2, or the Mali blob.
- CPU cores are **out of scope**. Point users at
  `kubernetes-sigs/dra-driver-cpu`. Do not advertise CPUs from this driver.
- The API and discovery path must stay SoC-generic (`soc` attribute, discovery
  from sysfs/DT). RK3588 is the first implementation, not the only shape.
- Future device types (RGA, VPU, …) should be additional `type` values under
  the same DRA driver (`dra.rockchip.com`), not a new API group.

Minimum cluster version is Kubernetes **1.35** with `DRAConsumableCapacity`
enabled (alpha, off by default). 1.36+ has that gate on by default. Develop
against the latest Kubernetes release's client libraries.

## Layout

```
.devcontainer/                        dev image (go, kubectl, helm, ssh)
api/resource.rockchip.com/v1alpha1/   opaque NpuConfig / GpuConfig
cmd/kubeletplugin/                    DRA kubelet plugin
internal/discovery/                   sysfs + mock device enumeration
pkg/flags/                            kube client and logging flags
pkg/metrics/                          Prometheus metrics
deployments/helm/rockchip-dra-driver/ Helm chart (DeviceClasses, DaemonSet)
deployments/container/                image build
demo/scripts/                         image/chart push helpers
examples/                             deployable NPU/GPU workload manifests
test/e2e/                             kind e2e (mock discovery)
docs/upstream.md                      notes on upstream tracking
```

## Development tools

`go`, `gofmt`, `golangci-lint`, `kubectl`, and `helm` come from the
devcontainer. Do not install them on the host to complete a task. `kind` is
not in the image; e2e stays on CI or a host engine.

### Devcontainer

- Preferred portable workspace: [`.devcontainer/`](.devcontainer/) (Fedora 44 image).
- Operator guide: [`.devcontainer/README.md`](.devcontainer/README.md).
- Host-agnostic on macOS and Aurora/Bluefin DX (Docker or Podman).
- Host bind mounts cover kubeconfig (`~/.kube`), SSH agent + `~/.ssh`, and
  git/`gh` config so a running cluster or node can be inspected from the
  container. Do not copy kubeconfigs or private keys into the image or the repo.

#### Running tools from an agent

When a required tool is missing on the host, run it **inside the
devcontainer**, not by installing it on the host. Prefer the container even
when a host binary exists if the task needs the image's Go toolchain
(`GOTOOLCHAIN=local`, version taken from `go.mod` at image build) or the
mounted cluster credentials.

Check the image's tool list in [`.devcontainer/README.md`](.devcontainer/README.md)
(table under **What you get**). Do not assume a host install, a pinned CLI
version, or invent extra packages. Rebuild the image after the `go` line in
`go.mod` changes.

From the repo root:

```sh
devcontainer exec --workspace-folder . <command>
```

If the workspace uses **Podman** (not Docker), add `--docker-path podman` to
`devcontainer` invocations (`up`, `exec`, and similar). Do not hard-code that
flag when Docker is the engine. Detect from the environment (for example
`docker info` vs `podman info`, or `DOCKER_HOST` pointing at a podman sock)
or from the operator's existing `devcontainer` alias.

Bring the container up first if `exec` fails because no workspace is running
(`devcontainer up --workspace-folder .`, with `--docker-path podman` when
applicable). Operator details stay in [`.devcontainer/README.md`](.devcontainer/README.md).

Kind clusters and driver image builds need the host container engine. That
socket is not mounted. Run those on the host engine or in CI; do not install
Go on the host as a workaround.

## Commits (Angular / conventional commits)

This repo uses [semantic-release](https://semantic-release.gitbook.io/) with the
Angular convention. Every commit subject must be:

```
type(scope): subject
```

Common types: `feat`, `fix`, `perf`, `revert`, `docs`, `chore`, `ci`, `test`,
`refactor`, `style`.

Common scopes: `npu`, `gpu`, `api`, `helm`, `ci`, `docs`, `discovery`.

Rules:

- **One changelog-worthy change per commit.** Do not mix unrelated features,
  Helm, CI, and API edits in a single commit. `CHANGELOG.md` is generated from
  these subjects; a blob commit produces a useless changelog entry.
- Subject is imperative, lowercase after the type, and ≤72 characters.
- The body explains *why*, not a dump of the diff.
- `feat` and `fix` (and `perf` / `revert`) create releases. `chore`, `docs`,
  `ci`, `test`, `style`, and `refactor` do not, unless they include a breaking
  footer (see 0.x policy below).

### 0.x until a human cuts 1.0.0

The repo is seeded with annotated tag `0.0.0`. Until **1.0.0**:

| Commit | Result |
|---|---|
| `feat:` | `0.x+1.0` |
| `fix:` | `0.x.y+1` |
| `feat!` or `BREAKING CHANGE:` footer | jumps to **1.0.0** — **forbidden** |
| `chore:` / `docs:` / `ci:` / `test:` | no release |

During 0.x, breaking API changes are still `feat(api): …` with a note in the
body (not a `BREAKING CHANGE:` footer). SemVer 0.y.z is not a stable public
API.

**Do not declare 1.0.0 from a coding agent.** The maintainer will cut 1.0.0
manually when the API is stable (typically with an intentional `feat!` /
`BREAKING CHANGE:` commit).

## Watching dra-example-driver

This project started from the structure of
[`kubernetes-sigs/dra-example-driver`](https://github.com/kubernetes-sigs/dra-example-driver),
not as a git fork. Upstream keeps picking up new DRA framework features.

**When to check:** each Kubernetes minor release, and whenever example-driver
tags a release.

**What to look at:**

1. `go.mod` versions of `k8s.io/dynamic-resource-allocation`, `k8s.io/api`,
   `k8s.io/client-go`, `k8s.io/kubelet`.
2. kubeletplugin helper API (`kubeletplugin.Start` options, Prepare/Unprepare,
   health, rolling update, device metadata, ShareID).
3. Helm values, CDI handling, e2e harness, metrics.
4. The Kubernetes `v1.xx DRA updates` blog for that cycle.

**Port:** library upgrades, helper API changes, CDI/Helm/e2e/metrics patterns,
and new *stable or beta default-on* DRA features that apply to **node-local
shared accelerators** (consumable capacity, device taints, extended resource
names, device metadata).

**Do not port:** dummy GPU/net device profiles, TimeSlicing /
SpacePartitioning, `example.com` API types, `cloudbuild.yaml` / k8s-infra
release flow, BindingConditions (those are for fabric-attached / FPGA-style
devices; rocket and panthor are node-local).

Prefer bumping `k8s.io/dynamic-resource-allocation` over copy-pasting example
driver internals. Record skipped upstream changes in
[`docs/upstream.md`](docs/upstream.md).

## DRA features we use

Use the latest Kubernetes DRA feature set **as appropriate** for these devices:

- `resource.k8s.io/v1` ResourceSlice / ResourceClaim / DeviceClass
- Consumable capacity: `allowMultipleAllocations` + `shares` + `requestPolicy`
- Extended resource names on DeviceClasses (optional Helm values)
- Device taints when a discovered device is unhealthy
- Device metadata via the kubeletplugin helper
- ShareID on shared allocations

Do not enable BindingConditions, partitionable-device sharedCounters, or node
allocatable resource mapping unless a later SoC actually needs them.

Opaque `NpuConfig` / `GpuConfig` stay empty TypeMeta stubs until there is a
real configuration knob. Do not invent NVIDIA-style sharing strategies.

## Discovery and sharing model

- One NPU device and one GPU device per node (not per core).
- `allowMultipleAllocations: true` with Helm-tunable `maxAllocations`.
- If Helm leaves the cap unset (`0`), use SoC-aware defaults (RK3588 NPU → 3
  cores as a congestion heuristic; GPU → 8).
- Core count / shader core count are **attributes**, not allocatable capacity.
  Mainline rocket has no core-mask UAPI; panthor has no shader-core isolation.
- Missing KMD: omit that device type, do not crash the plugin.
- Kind CI uses `--mock-devices`. Hardware tests need self-hosted RK3588 runners.

## License

Apache-2.0. Keep existing `Copyright … The Kubernetes Authors` headers on
derived files; add `Copyright 2026 Graeme Lawes` for modifications. New files
use the Graeme Lawes header from `hack/boilerplate.go.txt`.

## Tests

- Unit tests and `go test ./...` on every PR. Run them inside the devcontainer
  (`devcontainer exec --workspace-folder . go test ./...`, with
  `--docker-path podman` only when Podman is the engine).
- Kind e2e with mock discovery on GitHub-hosted runners. Do not expect
  `make setup-e2e` to work inside the devcontainer; the host engine socket
  is not mounted.
- Do not assume `/dev/accel` or panthor exist in CI.
