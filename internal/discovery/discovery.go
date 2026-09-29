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

package discovery

import "github.com/gclawes/rockchip-dra-driver/pkg/consts"

// Enumerate returns the accelerators visible on this node.
// A missing kernel module omits that device type; it is not an error.
func Enumerate(cfg Config) ([]Device, error) {
	cfg = defaultConfig(cfg)
	if cfg.Mock {
		return mockDevices(cfg), nil
	}

	soc := detectSoC(cfg.SysfsRoot)
	var devices []Device
	if cfg.NPUEnabled {
		devices = append(devices, discoverNPU(cfg, soc)...)
	}
	if cfg.GPUEnabled {
		devices = append(devices, discoverGPU(cfg, soc)...)
	}
	if cfg.VPUEnabled {
		devices = append(devices, discoverVPU(cfg, soc)...)
	}
	if cfg.RGAEnabled {
		devices = append(devices, discoverRGA(cfg, soc)...)
	}
	return devices, nil
}

func mockDevices(cfg Config) []Device {
	soc := "rk3588"
	var devices []Device
	if cfg.NPUEnabled {
		maxAlloc := int64(cfg.NPUMaxAllocations)
		if maxAlloc <= 0 {
			maxAlloc = defaultNPUMaxAllocations(soc, 3)
		}
		devices = append(devices, Device{
			Name:           "npu-0",
			Type:           consts.TypeNPU,
			SoC:            soc,
			Vendor:         consts.VendorRockchip,
			Model:          "rknn",
			KMD:            consts.KMDRocket,
			CoreCount:      3,
			DeviceNode:     "/dev/accel/accel0",
			MaxAllocations: maxAlloc,
		})
	}
	if cfg.GPUEnabled {
		maxAlloc := int64(cfg.GPUMaxAllocations)
		if maxAlloc <= 0 {
			maxAlloc = defaultGPUMaxAllocations()
		}
		devices = append(devices, Device{
			Name:           "gpu-0",
			Type:           consts.TypeGPU,
			SoC:            soc,
			Vendor:         consts.VendorARM,
			Model:          "Mali-G610",
			KMD:            consts.KMDPanthor,
			ShaderCores:    4,
			DeviceNode:     "/dev/dri/renderD128",
			MaxAllocations: maxAlloc,
		})
	}
	if cfg.VPUEnabled {
		maxAlloc := int64(cfg.VPUMaxAllocations)
		if maxAlloc <= 0 {
			maxAlloc = consts.DefaultVPUMaxAllocations
		}
		// video2 matches the rkvdec node observed on rk04. The index is not
		// stable; mock mode never opens it.
		devices = append(devices, Device{
			Name:           "vpu-rkvdec-0",
			Type:           consts.TypeVPU,
			SoC:            soc,
			Vendor:         consts.VendorRockchip,
			Model:          "rockchip,rk3588-vdec",
			KMD:            consts.KMDRkvdec,
			Function:       consts.FunctionDecode,
			Block:          consts.BlockRkvdec,
			DeviceNode:     "/dev/video2",
			MaxAllocations: maxAlloc,
		})
	}
	if cfg.RGAEnabled {
		maxAlloc := int64(cfg.RGAMaxAllocations)
		if maxAlloc <= 0 {
			maxAlloc = consts.DefaultRGAMaxAllocations
		}
		// video0 matches the rockchip-rga node observed on rk04. The index
		// is not stable; mock mode never opens it.
		devices = append(devices, Device{
			Name:           "rga-0",
			Type:           consts.TypeRGA,
			SoC:            soc,
			Vendor:         consts.VendorRockchip,
			Model:          "rockchip,rk3588-rga",
			KMD:            consts.KMDRGA,
			DeviceNode:     "/dev/video0",
			MaxAllocations: maxAlloc,
		})
	}
	return devices
}
