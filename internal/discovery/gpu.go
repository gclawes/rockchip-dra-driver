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

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gclawes/rockchip-dra-driver/pkg/consts"
)

func discoverGPU(cfg Config, soc string) []Device {
	drmDir := filepath.Join(cfg.SysfsRoot, "class", "drm")
	entries, err := os.ReadDir(drmDir)
	if err != nil {
		return nil
	}

	var renderNode string
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, "renderD") {
			continue
		}
		classPath := filepath.Join(drmDir, name)
		driver := ueventDriver(classPath)
		if driver == "" {
			driver = ueventDriver(filepath.Join(classPath, "device"))
		}
		if driver != consts.KMDPanthor {
			continue
		}
		devNode := filepath.Join(cfg.DevRoot, "dri", name)
		if !exists(devNode) {
			devNode = filepath.Join("/dev/dri", name)
		}
		renderNode = devNode
		break
	}
	if renderNode == "" {
		return nil
	}

	maxAlloc := int64(cfg.GPUMaxAllocations)
	if maxAlloc <= 0 {
		maxAlloc = defaultGPUMaxAllocations()
	}

	model, shaderCores := gpuModelForSoC(soc)

	return []Device{{
		Name:           "gpu-0",
		Type:           consts.TypeGPU,
		SoC:            soc,
		Vendor:         consts.VendorARM,
		Model:          model,
		KMD:            consts.KMDPanthor,
		ShaderCores:    shaderCores,
		DeviceNode:     renderNode,
		MaxAllocations: maxAlloc,
	}}
}

func gpuModelForSoC(soc string) (string, int64) {
	switch soc {
	case "rk3588", "rk3588s":
		return "Mali-G610", 4
	default:
		return "Mali", 0
	}
}
