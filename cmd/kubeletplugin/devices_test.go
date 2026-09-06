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
}
