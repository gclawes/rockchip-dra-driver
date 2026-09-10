/*
 * Copyright 2023 The Kubernetes Authors.
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
	"fmt"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/types"
	cdiapi "tags.cncf.io/container-device-interface/pkg/cdi"
	cdiparser "tags.cncf.io/container-device-interface/pkg/parser"
	cdispec "tags.cncf.io/container-device-interface/specs-go"

	"github.com/gclawes/rockchip-dra-driver/internal/discovery"
)

const cdiClass = "dra"

type CDIHandler struct {
	cache      *cdiapi.Cache
	driverName string
	mock       bool
}

func NewCDIHandler(root, driverName string, mock bool) (*CDIHandler, error) {
	cache, err := cdiapi.NewCache(cdiapi.WithSpecDirs(root))
	if err != nil {
		return nil, fmt.Errorf("unable to create a new CDI cache: %w", err)
	}
	return &CDIHandler{cache: cache, driverName: driverName, mock: mock}, nil
}

func (cdi *CDIHandler) CreateClaimSpecFile(claimUID string, devices []preparedDevice) error {
	specName := cdiapi.GenerateTransientSpecName(cdi.vendor(), cdiClass, claimUID)
	spec := &cdispec.Spec{
		Kind:    cdi.kind(),
		Devices: []cdispec.Device{},
	}

	for _, device := range devices {
		edits := deviceEdits(device.discovered, cdi.mock)
		spec.Devices = append(spec.Devices, cdispec.Device{
			Name:           cdiDeviceName(claimUID, device.DeviceName, uidString(device.ShareID)),
			ContainerEdits: edits,
		})
	}

	var err error
	spec.Version, err = cdispec.MinimumRequiredVersion(spec)
	if err != nil {
		return fmt.Errorf("failed to get minimum required CDI spec version: %w", err)
	}
	return cdi.cache.WriteSpec(spec, specName)
}

func (cdi *CDIHandler) DeleteClaimSpecFile(claimUID string) error {
	specName := cdiapi.GenerateTransientSpecName(cdi.vendor(), cdiClass, claimUID)
	return cdi.cache.RemoveSpec(specName)
}

func (cdi *CDIHandler) GetClaimDevices(claimUID, deviceName string, shareID *types.UID) []string {
	return []string{
		cdiparser.QualifiedName(cdi.vendor(), cdiClass, cdiDeviceName(claimUID, deviceName, uidString(shareID))),
	}
}

func (cdi *CDIHandler) kind() string {
	return cdi.vendor() + "/" + cdiClass
}

func (cdi *CDIHandler) vendor() string {
	return "k8s." + cdi.driverName
}

func cdiDeviceName(claimUID, deviceName, shareID string) string {
	name := claimUID + "-" + deviceName
	if shareID != "" {
		name += "-" + shareID
	}
	return name
}

func uidString(id *types.UID) string {
	if id == nil {
		return ""
	}
	return string(*id)
}

func deviceEdits(d discovery.Device, mock bool) cdispec.ContainerEdits {
	prefix := "DRA_ROCKCHIP_" + strings.ToUpper(d.Type)
	edits := cdispec.ContainerEdits{
		Env: []string{
			fmt.Sprintf("DRA_ROCKCHIP_SOC=%s", d.SoC),
			fmt.Sprintf("%s_KMD=%s", prefix, d.KMD),
			fmt.Sprintf("%s_DEVICE=%s", prefix, d.DeviceNode),
		},
	}
	if !mock && d.DeviceNode != "" {
		if _, err := os.Stat(d.DeviceNode); err == nil {
			edits.DeviceNodes = []*cdispec.DeviceNode{{
				Path:     d.DeviceNode,
				HostPath: d.DeviceNode,
			}}
		}
	}
	return edits
}
