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

	vpu := deviceEdits(discovery.Device{
		Type:       consts.TypeVPU,
		SoC:        "rk3588",
		KMD:        consts.KMDRkvdec,
		Block:      consts.BlockRkvdec,
		DeviceNode: "/dev/video2",
	}, true)
	wantVPU := []string{
		"DRA_ROCKCHIP_SOC=rk3588",
		"DRA_ROCKCHIP_VPU_RKVDEC_KMD=rkvdec",
		"DRA_ROCKCHIP_VPU_RKVDEC_DEVICE=/dev/video2",
	}
	if !slices.Equal(vpu.Env, wantVPU) {
		t.Fatalf("vpu env: got %v want %v", vpu.Env, wantVPU)
	}
	if slices.Contains(vpu.Env, "DRA_ROCKCHIP_VPU_DEVICE=/dev/video2") {
		t.Fatalf("bare VPU env would collide across blocks: %v", vpu.Env)
	}

	rga := deviceEdits(discovery.Device{
		Type:       consts.TypeRGA,
		SoC:        "rk3588",
		KMD:        consts.KMDRGA,
		DeviceNode: "/dev/video0",
	}, true)
	wantRGA := []string{
		"DRA_ROCKCHIP_SOC=rk3588",
		"DRA_ROCKCHIP_RGA_KMD=rockchip-rga",
		"DRA_ROCKCHIP_RGA_DEVICE=/dev/video0",
	}
	if !slices.Equal(rga.Env, wantRGA) {
		t.Fatalf("rga env: got %v want %v", rga.Env, wantRGA)
	}
}

func TestDeviceEditsIncludesGID(t *testing.T) {
	npu := deviceEdits(discovery.Device{
		Type:           consts.TypeNPU,
		SoC:            "rk3588",
		KMD:            consts.KMDRocket,
		DeviceNode:     "/dev/accel/accel0",
		DeviceGID:      44,
		DeviceGIDKnown: true,
	}, true)
	want := "DRA_ROCKCHIP_NPU_GID=44"
	if !slices.Contains(npu.Env, want) {
		t.Fatalf("missing %s in %v", want, npu.Env)
	}

	rootOwned := deviceEdits(discovery.Device{
		Type:           consts.TypeNPU,
		KMD:            consts.KMDRocket,
		DeviceNode:     "/dev/accel/accel0",
		DeviceGID:      0,
		DeviceGIDKnown: true,
	}, true)
	if !slices.Contains(rootOwned.Env, "DRA_ROCKCHIP_NPU_GID=0") {
		t.Fatalf("gid 0 should still be published: %v", rootOwned.Env)
	}
}
