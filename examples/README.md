# Example workloads

These manifests show how to claim the NPU and GPU this driver advertises.
They are **wiring examples**, not inference demos. The containers only print
`DRA_ROCKCHIP_*` environment variables and the injected device nodes, then
sleep. Swap the image when a rocket-capable workload image exists.

They use `debian:bookworm-slim`, which publishes `linux/arm64` and runs on
RK3588. Apply them against a cluster that already has this driver installed
and `DRAConsumableCapacity` enabled.

```bash
kubectl apply -f examples/npu-deployment.yaml
kubectl apply -f examples/gpu-deployment.yaml
kubectl apply -f examples/npu-and-gpu-deployment.yaml
kubectl apply -f examples/npu-shared-replicas.yaml
```

Each example is a `Deployment` plus a `ResourceClaimTemplate`. Kubernetes
creates one `ResourceClaim` per replica. The default request consumes one
`shares` unit (Helm defaults: 3 concurrent NPU claims, 8 GPU).

Check a running pod:

```bash
kubectl exec deploy/npu-example -- env | grep DRA_ROCKCHIP
kubectl exec deploy/npu-example -- ls -l /dev/accel
```
