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

func discoverNPU(cfg Config, soc string) []Device {
	accelDir := filepath.Join(cfg.SysfsRoot, "class", "accel")
	entries, err := os.ReadDir(accelDir)
	if err != nil {
		return nil
	}

	var nodes []string
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, "accel") {
			continue
		}
		classPath := filepath.Join(accelDir, name)
		if ueventDriver(classPath) != consts.KMDRocket && ueventDriver(filepath.Join(classPath, "device")) != consts.KMDRocket {
			continue
		}
		devNode := filepath.Join(cfg.DevRoot, "accel", name)
		if !exists(devNode) {
			// Still advertise: the node may appear after udev.
			devNode = filepath.Join("/dev/accel", name)
		}
		nodes = append(nodes, devNode)
	}
	if len(nodes) == 0 {
		return nil
	}

	cores := countRocketCores(cfg.SysfsRoot)
	maxAlloc := int64(cfg.NPUMaxAllocations)
	if maxAlloc <= 0 {
		maxAlloc = defaultNPUMaxAllocations(soc, cores)
	}

	return []Device{{
		Name:           "npu-0",
		Type:           consts.TypeNPU,
		SoC:            soc,
		Vendor:         consts.VendorRockchip,
		Model:          "rknn",
		KMD:            consts.KMDRocket,
		CoreCount:      cores,
		DeviceNode:     nodes[0],
		MaxAllocations: maxAlloc,
	}}
}

func countRocketCores(sysfsRoot string) int64 {
	// Each bound rocket platform device is one NPU core.
	dir := filepath.Join(sysfsRoot, "bus", "platform", "drivers", "rocket")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	var n int64
	for _, e := range entries {
		name := e.Name()
		if name == "bind" || name == "unbind" || name == "uevent" || strings.HasPrefix(name, "module") {
			continue
		}
		n++
	}
	return n
}
