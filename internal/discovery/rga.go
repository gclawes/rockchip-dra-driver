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
	"fmt"
	"slices"

	"github.com/gclawes/rockchip-dra-driver/pkg/consts"
)

func discoverRGA(cfg Config, soc string) []Device {
	var nodes []v4lNode
	for _, n := range discoverV4L(cfg) {
		if n.Driver != consts.KMDRGA {
			continue
		}
		nodes = append(nodes, n)
	}
	slices.SortFunc(nodes, func(a, b v4lNode) int {
		return a.Index - b.Index
	})

	maxAlloc := int64(cfg.RGAMaxAllocations)
	if maxAlloc <= 0 {
		maxAlloc = consts.DefaultRGAMaxAllocations
	}

	devices := make([]Device, 0, len(nodes))
	for i, n := range nodes {
		dev := Device{
			Name:           fmt.Sprintf("rga-%d", i),
			Type:           consts.TypeRGA,
			SoC:            soc,
			Vendor:         consts.VendorRockchip,
			Model:          n.Compatible,
			KMD:            n.Driver,
			DeviceNode:     n.DeviceNode,
			MaxAllocations: maxAlloc,
		}
		applyNodeGID(&dev)
		devices = append(devices, dev)
	}
	return devices
}
