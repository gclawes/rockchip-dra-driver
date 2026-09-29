# rockchip-dra-driver

Kubernetes [Dynamic Resource Allocation (DRA)](https://kubernetes.io/docs/concepts/scheduling-eviction/dynamic-resource-allocation/)
driver for Rockchip accelerators.

The first supported SoC is the **RK3588**, advertising:

- **NPU** via the mainline [`accel/rocket`](https://docs.kernel.org/accel/rocket/index.html) driver (`/dev/accel/accel*`)
- **GPU** via the mainline `panthor` DRM driver (Mali-G610 render node)
- **VPU** via mainline `rkvdec` and `hantro-vpu` (`/dev/video*`), one device per block rather than per video index

The driver is **mainline-only**. It does not support the proprietary Rockchip
`rknpu` kernel module or RKNN-Toolkit2. CPU cores are out of scope; use
[`kubernetes-sigs/dra-driver-cpu`](https://github.com/kubernetes-sigs/dra-driver-cpu)
alongside this driver if you need exclusive CPU pinning.

## Status

Pre-1.0. The public API is not stable. See [CONTRIBUTING.md](CONTRIBUTING.md)
for how to send changes and [AGENTS.md](AGENTS.md) for project conventions.

## Requirements

- Kubernetes **1.35+**. Consumable capacity (`DRAConsumableCapacity`) is
  required; it is alpha and off by default in 1.35, and beta default-on in 1.36+.
- Linux **6.18+** for rocket (NPU); **6.10+** for panthor (GPU)
- Container runtime with [CDI](https://github.com/cncf-tags/container-device-interface) support

## Install

```bash
helm upgrade -i rockchip-dra-driver \
  oci://ghcr.io/gclawes/rockchip-dra-driver/charts/rockchip-dra-driver \
  --namespace rockchip-dra-driver \
  --create-namespace
```

Helm knobs:

| Value | Default | Meaning |
|---|---|---|
| `npu.enabled` | `true` | Publish NPU devices |
| `npu.maxAllocations` | `0` (SoC default; RK3588 → 3) | Concurrent NPU claim cap |
| `gpu.enabled` | `true` | Publish GPU devices |
| `gpu.maxAllocations` | `0` (default 8) | Concurrent GPU claim cap |
| `vpu.enabled` | `true` | Publish VPU devices |
| `vpu.maxAllocations` | `0` (default 1 per block) | Concurrent claims of each VPU |
| `mockDevices` | `false` | Fake RK3588 devices for kind |
| `discoveryInterval` | `10s` | How often to re-read sysfs. Keep this under 30s |

DeviceClasses: `npu.rockchip.com`, `gpu.rockchip.com`, and, when VPU
discovery is enabled, `vpu-rkvdec.rockchip.com`,
`vpu-hantro-dec.rockchip.com`, `vpu-hantro-enc.rockchip.com`,
`vpu-hantro-av1.rockchip.com`. There is no generic `vpu.rockchip.com`.
Hantro decode and AV1 decode share a driver, so classes select
`type == vpu && block == …`. Driver name: `dra.rockchip.com`.

The plugin reports device health to the kubelet and republishes the
ResourceSlice when sysfs changes. A char device that has not appeared yet
(udev) is `Unknown` and is not tainted. Once a device has been seen, losing
its kernel driver or char device marks it `Unhealthy` and taints it
`dra.rockchip.com/unhealthy` with effect `NoExecute` (this also blocks new
scheduling). The taint value is `node-missing` or `not-discovered`. New
prepares fail until the device is usable again. The device stays published
while a claim is still prepared, and is removed once that claim is
unprepared. Device taints and health status need Kubernetes 1.36+
(`DRADeviceTaints` and the DRA resource-health service); rediscovery and
failed prepares still apply on 1.35.

Char devices are injected with the host mode and owner. On a typical board
the NPU and GPU nodes are mode `0660` and group `render`. VPU nodes are
typically group `video`. The gid is not
stable across distros, so this driver does not chmod the node and does not
guess a supplemental group. When the node exists, the ResourceSlice attribute
`deviceGid` and a CDI variable carry the host gid. NPU and GPU use
`DRA_ROCKCHIP_<TYPE>_GID`. VPU uses `DRA_ROCKCHIP_VPU_<BLOCK>_GID`
(`DRA_ROCKCHIP_VPU_RKVDEC_GID`, `DRA_ROCKCHIP_VPU_HANTRO_DEC_GID`, …) so two
blocks in one pod do not overwrite each other.
Set that value on the pod:

```yaml
securityContext:
  supplementalGroups: [<deviceGid>]
```

A pod that claims both an NPU and a VPU needs every distinct gid. Mock
discovery omits `deviceGid` because there is no host node.

A claim that omits `shares` gets 1. Requests must fall in `1..capacity`
(step 1 when capacity is at least 2). Asking for the whole published count
is how to take the device exclusively. There is no core-mask UAPI.

Deployable `Deployment` examples (NPU, GPU, both, shared NPU replicas, an
exclusive NPU claim, and an rkvdec claim) are in [`examples/`](examples/).

## Development

Go, lint, and cluster tools run in the [devcontainer](.devcontainer/README.md),
not on the host. From the repo root:

```bash
devcontainer up --workspace-folder .
devcontainer exec --workspace-folder . make test
```

Add `--docker-path podman` when the engine is Podman, not Docker. Do not
hard-code that flag for Docker. See [`.devcontainer/README.md`](.devcontainer/README.md).

Once inside the container: `make cmds`, `make test`, `make lint`.

Kind e2e (`make setup-e2e test-e2e teardown-e2e`) needs the host container
engine. CI runs it with mock devices.

Releases are automated with [semantic-release](https://semantic-release.org/)
from Angular conventional commits (`feat:`, `fix:`, …). A single releasing
pull request merged to `master` publishes immediately. Several that should
share a version land on a `prep/X.Y.Z` branch first. See
[CONTRIBUTING.md](CONTRIBUTING.md) for the commit format and the train.

## License

Apache License 2.0. See [LICENSE](LICENSE) and [NOTICE](NOTICE).
