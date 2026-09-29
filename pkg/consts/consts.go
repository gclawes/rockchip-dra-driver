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

	// DefaultVPUMaxAllocations is used when Helm does not override the VPU
	// share cap. One claim owns the node until that cap is raised.
	DefaultVPUMaxAllocations = 1

	// TaintUnhealthy is applied with effect NoExecute when an advertised
	// device loses its kernel driver or char device. NoExecute also blocks
	// new scheduling. The value is TaintValueNodeMissing or
	// TaintValueNotDiscovered.
	TaintUnhealthy = "dra.rockchip.com/unhealthy"

	// TaintValueNodeMissing means the char device disappeared after it had
	// been present.
	TaintValueNodeMissing = "node-missing"

	// TaintValueNotDiscovered means enumeration no longer finds the device
	// (driver unbound or render node gone) while a claim is still prepared.
	TaintValueNotDiscovered = "not-discovered"
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
	AttrDeviceGID   = "deviceGid"
	AttrFunction    = "function"
	AttrBlock       = "block"
)

const (
	TypeNPU = "npu"
	TypeGPU = "gpu"
	TypeVPU = "vpu"

	KMDRocket  = "rocket"
	KMDPanthor = "panthor"
	KMDRkvdec  = "rkvdec"
	KMDHantro  = "hantro-vpu"

	FunctionDecode = "decode"
	FunctionEncode = "encode"

	BlockRkvdec    = "rkvdec"
	BlockHantroDec = "hantro-dec"
	BlockHantroEnc = "hantro-enc"
	BlockHantroAV1 = "hantro-av1"

	VendorRockchip = "rockchip"
	VendorARM      = "arm"
)
