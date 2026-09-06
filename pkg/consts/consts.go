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

// Package consts holds driver identity and ResourceSlice attribute names
// shared by discovery, the kubelet plugin, and tests.
package consts

const (
	// DriverName is the DRA driver name published on ResourceSlices.
	DriverName = "dra.rockchip.com"

	// CapacityShares is the consumable-capacity key used to cap concurrent
	// allocations of a shared accelerator.
	CapacityShares = "shares"

	// DefaultGPUMaxAllocations is used when Helm does not override the GPU
	// share cap and discovery cannot infer a better value.
	DefaultGPUMaxAllocations = 8
)

// ResourceSlice attribute names. In CEL these are addressed as
// device.attributes["dra.rockchip.com"].<name>.
const (
	AttrType        = "type"
	AttrSoC         = "soc"
	AttrVendor      = "vendor"
	AttrModel       = "model"
	AttrKMD         = "kmd"
	AttrCoreCount   = "coreCount"
	AttrShaderCores = "shaderCores"
	AttrDeviceNode  = "deviceNode"
)

const (
	TypeNPU = "npu"
	TypeGPU = "gpu"

	KMDRocket  = "rocket"
	KMDPanthor = "panthor"

	VendorRockchip = "rockchip"
	VendorARM      = "arm"
)
