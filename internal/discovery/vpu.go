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
	"strings"

	"github.com/gclawes/rockchip-dra-driver/pkg/consts"
)

type vpuMatch struct {
	node     v4lNode
	block    string
	function string
}

func discoverVPU(cfg Config, soc string) []Device {
	var matches []vpuMatch
	for _, n := range discoverV4L(cfg) {
		block, function, ok := classifyVPU(n)
		if !ok {
			continue
		}
		matches = append(matches, vpuMatch{node: n, block: block, function: function})
	}
	slices.SortFunc(matches, func(a, b vpuMatch) int {
		if a.block != b.block {
			return strings.Compare(a.block, b.block)
		}
		return a.node.Index - b.node.Index
	})

	maxAlloc := int64(cfg.VPUMaxAllocations)
	if maxAlloc <= 0 {
		maxAlloc = consts.DefaultVPUMaxAllocations
	}

	counts := map[string]int{}
	devices := make([]Device, 0, len(matches))
	for _, m := range matches {
		n := counts[m.block]
		counts[m.block] = n + 1
		dev := Device{
			Name:           fmt.Sprintf("vpu-%s-%d", m.block, n),
			Type:           consts.TypeVPU,
			SoC:            soc,
			Vendor:         consts.VendorRockchip,
			Model:          m.node.Compatible,
			KMD:            m.node.Driver,
			Function:       m.function,
			Block:          m.block,
			DeviceNode:     m.node.DeviceNode,
			MaxAllocations: maxAlloc,
		}
		applyNodeGID(&dev)
		devices = append(devices, dev)
	}
	return devices
}

// classifyVPU maps a video4linux node to a block. More specific Hantro
// strings are checked first so av1 and vepu cannot fall through to the
// H.264 decoder. Anything else, including RGA and ISP, is skipped.
func classifyVPU(n v4lNode) (block, function string, ok bool) {
	compat := strings.ToLower(n.Compatible)
	card := strings.ToLower(n.CardName)
	switch n.Driver {
	case consts.KMDRkvdec:
		return consts.BlockRkvdec, consts.FunctionDecode, true
	case consts.KMDHantro:
		switch {
		case strings.Contains(compat, "av1") || strings.Contains(card, "av1"):
			return consts.BlockHantroAV1, consts.FunctionDecode, true
		case strings.Contains(compat, "vepu") || strings.Contains(card, "vepu"):
			return consts.BlockHantroEnc, consts.FunctionEncode, true
		case strings.Contains(compat, "vpu121") || strings.Contains(card, "vpu-dec"):
			return consts.BlockHantroDec, consts.FunctionDecode, true
		default:
			return "", "", false
		}
	default:
		return "", "", false
	}
}
