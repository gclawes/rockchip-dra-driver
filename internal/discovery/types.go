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

// Config controls how devices are enumerated.
type Config struct {
	// SysfsRoot is the mount of sysfs (default /sys). Tests pass a fixture.
	SysfsRoot string
	// DevRoot is the mount of /dev (default /dev).
	DevRoot string
	// Mock advertises a synthetic RK3588 NPU+GPU instead of reading sysfs.
	Mock bool
	// NPUEnabled / GPUEnabled skip a device class when false.
	NPUEnabled bool
	GPUEnabled bool
	// NPUMaxAllocations / GPUMaxAllocations override SoC-aware defaults when
	// greater than zero.
	NPUMaxAllocations int
	GPUMaxAllocations int
}

// Device is one accelerator discovered on the node.
type Device struct {
	Name           string
	Type           string
	SoC            string
	Vendor         string
	Model          string
	KMD            string
	CoreCount      int64
	ShaderCores    int64
	DeviceNode     string
	MaxAllocations int64
}

func defaultConfig(c Config) Config {
	if c.SysfsRoot == "" {
		c.SysfsRoot = "/sys"
	}
	if c.DevRoot == "" {
		c.DevRoot = "/dev"
	}
	if !c.NPUEnabled && !c.GPUEnabled {
		// Zero-value Config means "discover everything".
		c.NPUEnabled = true
		c.GPUEnabled = true
	}
	return c
}

func defaultNPUMaxAllocations(soc string, coreCount int64) int64 {
	if coreCount > 0 {
		return coreCount
	}
	switch soc {
	case "rk3588", "rk3588s":
		return 3
	default:
		return 1
	}
}

func defaultGPUMaxAllocations() int64 {
	return consts.DefaultGPUMaxAllocations
}
