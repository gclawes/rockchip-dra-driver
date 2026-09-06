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
	resourceapi "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/dynamic-resource-allocation/resourceslice"
	"k8s.io/utils/ptr"

	"github.com/gclawes/rockchip-dra-driver/internal/discovery"
	"github.com/gclawes/rockchip-dra-driver/pkg/consts"
)

func devicesToResources(nodeName string, devices []discovery.Device) resourceslice.DriverResources {
	var sliceDevices []resourceapi.Device
	for _, d := range devices {
		sliceDevices = append(sliceDevices, toResourceDevice(d))
	}
	return resourceslice.DriverResources{
		Pools: map[string]resourceslice.Pool{
			nodeName: {
				Slices: []resourceslice.Slice{{Devices: sliceDevices}},
			},
		},
	}
}

func toResourceDevice(d discovery.Device) resourceapi.Device {
	attrs := map[resourceapi.QualifiedName]resourceapi.DeviceAttribute{
		consts.AttrType:       {StringValue: ptr.To(d.Type)},
		consts.AttrVendor:     {StringValue: ptr.To(d.Vendor)},
		consts.AttrModel:      {StringValue: ptr.To(d.Model)},
		consts.AttrKMD:        {StringValue: ptr.To(d.KMD)},
		consts.AttrDeviceNode: {StringValue: ptr.To(d.DeviceNode)},
	}
	if d.SoC != "" {
		attrs[consts.AttrSoC] = resourceapi.DeviceAttribute{StringValue: ptr.To(d.SoC)}
	}
	if d.CoreCount > 0 {
		attrs[consts.AttrCoreCount] = resourceapi.DeviceAttribute{IntValue: ptr.To(d.CoreCount)}
	}
	if d.ShaderCores > 0 {
		attrs[consts.AttrShaderCores] = resourceapi.DeviceAttribute{IntValue: ptr.To(d.ShaderCores)}
	}

	one := resource.MustParse("1")
	shares := resource.NewQuantity(d.MaxAllocations, resource.DecimalSI)
	return resourceapi.Device{
		Name:                     d.Name,
		Attributes:               attrs,
		AllowMultipleAllocations: ptr.To(true),
		Capacity: map[resourceapi.QualifiedName]resourceapi.DeviceCapacity{
			consts.CapacityShares: {
				Value: *shares,
				RequestPolicy: &resourceapi.CapacityRequestPolicy{
					Default: &one,
				},
			},
		},
	}
}
