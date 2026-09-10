# rockchip-dra-driver

Kubernetes [Dynamic Resource Allocation (DRA)](https://kubernetes.io/docs/concepts/scheduling-eviction/dynamic-resource-allocation/)
driver for Rockchip accelerators.

The first supported SoC is the **RK3588**, advertising:

- **NPU** via the mainline [`accel/rocket`](https://docs.kernel.org/accel/rocket/index.html) driver (`/dev/accel/accel*`)
- **GPU** via the mainline `panthor` DRM driver (Mali-G610 render node)

The driver is **mainline-only**. It does not support the proprietary Rockchip
`rknpu` kernel module or RKNN-Toolkit2. CPU cores are out of scope; use
[`kubernetes-sigs/dra-driver-cpu`](https://github.com/kubernetes-sigs/dra-driver-cpu)
alongside this driver if you need exclusive CPU pinning.

## Status

Pre-1.0. The public API is not stable. See [AGENTS.md](AGENTS.md) for project
conventions.

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
| `mockDevices` | `false` | Fake RK3588 devices for kind |

DeviceClasses: `npu.rockchip.com`, `gpu.rockchip.com`. Driver name:
`dra.rockchip.com`.

Deployable `Deployment` examples (NPU, GPU, both, and shared NPU replicas)
are in [`examples/`](examples/).

## Development

```bash
make cmds
make test
make setup-e2e test-e2e teardown-e2e   # kind + mock devices
```

## License

Apache License 2.0. See [LICENSE](LICENSE) and [NOTICE](NOTICE).
