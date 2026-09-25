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

	resourceapi "k8s.io/api/resource/v1"

	"github.com/gclawes/rockchip-dra-driver/internal/discovery"
	"github.com/gclawes/rockchip-dra-driver/pkg/consts"
)

func TestDevicesToResources(t *testing.T) {
	devs, err := discovery.Enumerate(discovery.Config{Mock: true})
	if err != nil {
		t.Fatal(err)
	}
	res := devicesToResources("node-a", devs)
	pool, ok := res.Pools["node-a"]
	if !ok || len(pool.Slices) != 1 || len(pool.Slices[0].Devices) != 2 {
		t.Fatalf("unexpected pool: %+v", res.Pools)
	}
	npu := pool.Slices[0].Devices[0]
	if npu.Name != "npu-0" || npu.AllowMultipleAllocations == nil || !*npu.AllowMultipleAllocations {
		t.Fatalf("unexpected npu device: %+v", npu)
	}
	cap, ok := npu.Capacity[consts.CapacityShares]
	if !ok || cap.Value.CmpInt64(3) != 0 || cap.RequestPolicy == nil || cap.RequestPolicy.Default == nil {
		t.Fatalf("unexpected npu capacity: %+v", cap)
	}
	assertShareRange(t, cap.RequestPolicy, 3)
	gpu := pool.Slices[0].Devices[1]
	gpuCap := gpu.Capacity[consts.CapacityShares]
	assertShareRange(t, gpuCap.RequestPolicy, 8)
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
