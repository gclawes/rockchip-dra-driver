/*
 * Copyright 2026 Graeme Lawes.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"testing"
	"time"

	resourceapi "k8s.io/api/resource/v1"
	"k8s.io/dynamic-resource-allocation/kubeletplugin"

	"github.com/gclawes/rockchip-dra-driver/internal/discovery"
	"github.com/gclawes/rockchip-dra-driver/pkg/consts"
)

func TestDevicesToResources(t *testing.T) {
	devs, err := discovery.Enumerate(discovery.Config{Mock: true})
	if err != nil {
		t.Fatal(err)
	}
	tracked := make([]trackedDevice, len(devs))
	for i, d := range devs {
		tracked[i] = trackedDevice{Device: d, status: kubeletplugin.HealthStatusHealthy}
	}
	res := devicesToResources("node-a", tracked)
	pool, ok := res.Pools["node-a"]
	if !ok || len(pool.Slices) != 1 || len(pool.Slices[0].Devices) != 4 {
		t.Fatalf("unexpected pool: %+v", res.Pools)
	}
	npu := deviceByName(t, pool.Slices[0].Devices, "npu-0")
	if npu.AllowMultipleAllocations == nil || !*npu.AllowMultipleAllocations {
		t.Fatalf("unexpected npu device: %+v", npu)
	}
	if len(npu.Taints) != 0 {
		t.Fatalf("healthy device tainted: %+v", npu.Taints)
	}
	cap, ok := npu.Capacity[consts.CapacityShares]
	if !ok || cap.Value.CmpInt64(3) != 0 || cap.RequestPolicy == nil || cap.RequestPolicy.Default == nil {
		t.Fatalf("unexpected npu capacity: %+v", cap)
	}
	assertShareRange(t, cap.RequestPolicy, 3)
	gpu := deviceByName(t, pool.Slices[0].Devices, "gpu-0")
	assertShareRange(t, gpu.Capacity[consts.CapacityShares].RequestPolicy, 8)
	if _, ok := npu.Attributes[consts.AttrDeviceGID]; ok {
		t.Fatalf("mock device should not publish a gid: %+v", npu.Attributes)
	}
	vpu := deviceByName(t, pool.Slices[0].Devices, "vpu-rkvdec-0")
	assertShareRange(t, vpu.Capacity[consts.CapacityShares].RequestPolicy, 1)
	if vpu.Capacity[consts.CapacityShares].RequestPolicy.ValidRange.Step != nil {
		t.Fatalf("vpu step must be omitted when capacity is 1: %+v", vpu.Capacity[consts.CapacityShares].RequestPolicy)
	}
	if vpu.Attributes[consts.AttrBlock].StringValue == nil || *vpu.Attributes[consts.AttrBlock].StringValue != consts.BlockRkvdec {
		t.Fatalf("unexpected vpu block: %+v", vpu.Attributes)
	}
	// Name order, not discovery order.
	rga := deviceByName(t, pool.Slices[0].Devices, "rga-0")
	assertShareRange(t, rga.Capacity[consts.CapacityShares].RequestPolicy, 1)
	if pool.Slices[0].Devices[0].Name != "gpu-0" || pool.Slices[0].Devices[1].Name != "npu-0" || pool.Slices[0].Devices[2].Name != "rga-0" || pool.Slices[0].Devices[3].Name != "vpu-rkvdec-0" {
		t.Fatalf("devices not sorted: %+v", pool.Slices[0].Devices)
	}
}

func TestDeviceGIDAttribute(t *testing.T) {
	dev := toResourceDevice(trackedDevice{
		Device: discovery.Device{
			Name:           "npu-0",
			Type:           consts.TypeNPU,
			DeviceNode:     "/dev/accel/accel0",
			MaxAllocations: 3,
			DeviceGID:      992,
			DeviceGIDKnown: true,
		},
		status: kubeletplugin.HealthStatusHealthy,
	})
	got := dev.Attributes[consts.AttrDeviceGID]
	if got.IntValue == nil || *got.IntValue != 992 {
		t.Fatalf("unexpected deviceGid: %+v", got)
	}
}

func TestUnhealthyDeviceIsTainted(t *testing.T) {
	when := time.Date(2026, 9, 23, 12, 0, 0, 0, time.UTC)
	dev := toResourceDevice(trackedDevice{
		Device:     discovery.Device{Name: "npu-0", Type: consts.TypeNPU, MaxAllocations: 3},
		status:     kubeletplugin.HealthStatusUnhealthy,
		taintValue: consts.TaintValueNodeMissing,
		taintedAt:  when,
	})
	if len(dev.Taints) != 1 {
		t.Fatalf("expected one taint, got %+v", dev.Taints)
	}
	taint := dev.Taints[0]
	if taint.Key != consts.TaintUnhealthy || taint.Value != consts.TaintValueNodeMissing || taint.Effect != resourceapi.DeviceTaintEffectNoExecute {
		t.Fatalf("unexpected taint: %+v", taint)
	}
	if taint.TimeAdded == nil || !taint.TimeAdded.Time.Equal(when) {
		t.Fatalf("unexpected taint time: %+v", taint.TimeAdded)
	}
}

func TestShareRequestPolicySingleShareOmitsStep(t *testing.T) {
	policy := shareRequestPolicy(1)
	assertShareRange(t, policy, 1)
	if policy.ValidRange.Step != nil {
		t.Fatalf("step must be omitted when capacity is 1: %+v", policy.ValidRange.Step)
	}
}

func assertShareRange(t *testing.T, policy *resourceapi.CapacityRequestPolicy, max int64) {
	t.Helper()
	if policy == nil || policy.Default == nil || policy.Default.CmpInt64(1) != 0 {
		t.Fatalf("unexpected default: %+v", policy)
	}
	rng := policy.ValidRange
	if rng == nil || rng.Min == nil || rng.Max == nil || rng.Min.CmpInt64(1) != 0 || rng.Max.CmpInt64(max) != 0 {
		t.Fatalf("unexpected range for max %d: %+v", max, rng)
	}
	if max >= 2 {
		if rng.Step == nil || rng.Step.CmpInt64(1) != 0 {
			t.Fatalf("expected step 1: %+v", rng.Step)
		}
	}
}

func deviceByName(t *testing.T, devs []resourceapi.Device, name string) resourceapi.Device {
	t.Helper()
	for _, d := range devs {
		if d.Name == name {
			return d
		}
	}
	t.Fatalf("device %s not found in %+v", name, devs)
	return resourceapi.Device{}
}
