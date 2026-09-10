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
	"slices"
	"testing"

	"github.com/gclawes/rockchip-dra-driver/internal/discovery"
	"github.com/gclawes/rockchip-dra-driver/pkg/consts"
)

func TestDeviceEditsEnvScopedByType(t *testing.T) {
	npu := deviceEdits(discovery.Device{
		Type:       consts.TypeNPU,
		SoC:        "rk3588",
		KMD:        consts.KMDRocket,
		DeviceNode: "/dev/accel/accel0",
	}, true)
	wantNPU := []string{
		"DRA_ROCKCHIP_SOC=rk3588",
		"DRA_ROCKCHIP_NPU_KMD=rocket",
		"DRA_ROCKCHIP_NPU_DEVICE=/dev/accel/accel0",
	}
	if !slices.Equal(npu.Env, wantNPU) {
		t.Fatalf("npu env: got %v want %v", npu.Env, wantNPU)
	}

	gpu := deviceEdits(discovery.Device{
		Type:       consts.TypeGPU,
		SoC:        "rk3588",
		KMD:        consts.KMDPanthor,
		DeviceNode: "/dev/dri/renderD128",
	}, true)
	wantGPU := []string{
		"DRA_ROCKCHIP_SOC=rk3588",
		"DRA_ROCKCHIP_GPU_KMD=panthor",
		"DRA_ROCKCHIP_GPU_DEVICE=/dev/dri/renderD128",
	}
	if !slices.Equal(gpu.Env, wantGPU) {
		t.Fatalf("gpu env: got %v want %v", gpu.Env, wantGPU)
	}
}
