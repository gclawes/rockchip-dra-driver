# Example workloads

These manifests show how to claim the NPU and GPU this driver advertises.
They are **wiring examples**, not inference demos. The containers only print
CDI environment variables (`DRA_ROCKCHIP_SOC`, `DRA_ROCKCHIP_NPU_*`,
`DRA_ROCKCHIP_GPU_*`) and the injected device nodes, then sleep. Swap the
image when a rocket-capable workload image exists.

They use `debian:bookworm-slim`, which publishes `linux/arm64` and runs on
RK3588. Apply them against a cluster that already has this driver installed
and `DRAConsumableCapacity` enabled.

```bash
kubectl apply -f examples/npu-deployment.yaml
kubectl apply -f examples/gpu-deployment.yaml
kubectl apply -f examples/npu-and-gpu-deployment.yaml
kubectl apply -f examples/npu-shared-replicas.yaml
kubectl apply -f examples/npu-exclusive.yaml
```

Each example is a `Deployment` plus a `ResourceClaimTemplate`. Kubernetes
creates one `ResourceClaim` per replica. A request that omits `shares`
consumes one unit. Requests must be in `1..capacity` (Helm defaults: NPU
capacity 3 on RK3588, GPU capacity 8). `npu-exclusive.yaml` requests all 3
RK3588 NPU shares. Change that quantity if the published capacity differs.

Check a running pod:

```bash
kubectl exec deploy/npu-example -- env | grep DRA_ROCKCHIP
kubectl exec deploy/npu-example -- ls -l /dev/accel
```
