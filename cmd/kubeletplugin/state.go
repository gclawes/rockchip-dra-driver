/*
 * Copyright The Kubernetes Authors.
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
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"

	resourceapi "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/dynamic-resource-allocation/resourceslice"
	"k8s.io/klog/v2"

	"github.com/gclawes/rockchip-dra-driver/internal/discovery"
)

type preparedDevice struct {
	RequestNames []string
	PoolName     string
	DeviceName   string
	CDIDeviceIDs []string
	ShareID      *types.UID
	discovered   discovery.Device
}

type checkpoint struct {
	PreparedClaims []types.UID `json:"preparedClaims"`
}

type DeviceState struct {
	sync.Mutex
	driverName      string
	cdi             *CDIHandler
	driverResources resourceslice.DriverResources
	discovered      map[string]discovery.Device
	checkpointPath  string
}

func NewDeviceState(config *Config) (*DeviceState, error) {
	devices, err := discovery.Enumerate(config.discoveryConfig())
	if err != nil {
		return nil, fmt.Errorf("enumerate devices: %w", err)
	}

	discovered := make(map[string]discovery.Device, len(devices))
	for _, d := range devices {
		discovered[d.Name] = d
	}

	cdi, err := NewCDIHandler(config.flags.cdiRoot, config.flags.driverName, config.flags.mockDevices)
	if err != nil {
		return nil, fmt.Errorf("unable to create CDI handler: %w", err)
	}

	return &DeviceState{
		driverName:      config.flags.driverName,
		cdi:             cdi,
		driverResources: devicesToResources(config.flags.nodeName, devices),
		discovered:      discovered,
		checkpointPath:  filepath.Join(config.DriverPluginPath(), driverPluginCheckpointFile),
	}, nil
}

func (s *DeviceState) Prepare(ctx context.Context, claim *resourceapi.ResourceClaim) ([]preparedDevice, error) {
	s.Lock()
	defer s.Unlock()

	cp, err := readCheckpoint(s.checkpointPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read checkpoint: %w", err)
	}

	prepared, err := s.computeDevices(claim)
	if err != nil {
		return nil, err
	}

	if err := s.cdi.CreateClaimSpecFile(string(claim.UID), prepared); err != nil {
		return nil, fmt.Errorf("unable to create CDI spec file for claim: %w", err)
	}

	if !slices.Contains(cp.PreparedClaims, claim.UID) {
		cp.PreparedClaims = append(cp.PreparedClaims, claim.UID)
		if err := writeCheckpoint(s.checkpointPath, cp); err != nil {
			return nil, fmt.Errorf("unable to write checkpoint: %w", err)
		}
	}

	klog.FromContext(ctx).Info("Prepared claim", "uid", claim.UID, "devices", len(prepared))
	return prepared, nil
}

func (s *DeviceState) Unprepare(claimUID types.UID) error {
	s.Lock()
	defer s.Unlock()

	cp, err := readCheckpoint(s.checkpointPath)
	if err != nil {
		return fmt.Errorf("unable to read checkpoint: %w", err)
	}

	if err := s.cdi.DeleteClaimSpecFile(string(claimUID)); err != nil {
		return fmt.Errorf("unable to delete CDI spec file for claim: %w", err)
	}

	cp.PreparedClaims = slices.DeleteFunc(cp.PreparedClaims, func(u types.UID) bool { return u == claimUID })
	if err := writeCheckpoint(s.checkpointPath, cp); err != nil {
		return fmt.Errorf("unable to write checkpoint: %w", err)
	}
	return nil
}

func (s *DeviceState) computeDevices(claim *resourceapi.ResourceClaim) ([]preparedDevice, error) {
	if claim.Status.Allocation == nil {
		return nil, fmt.Errorf("claim not yet allocated")
	}

	var prepared []preparedDevice
	for _, result := range claim.Status.Allocation.Devices.Results {
		if result.Driver != s.driverName {
			continue
		}
		d, ok := s.discovered[result.Device]
		if !ok {
			return nil, fmt.Errorf("requested device is not allocatable: %s", result.Device)
		}
		prepared = append(prepared, preparedDevice{
			RequestNames: []string{result.Request},
			PoolName:     result.Pool,
			DeviceName:   result.Device,
			CDIDeviceIDs: s.cdi.GetClaimDevices(string(claim.UID), result.Device, result.ShareID),
			ShareID:      result.ShareID,
			discovered:   d,
		})
	}
	return prepared, nil
}

func readCheckpoint(path string) (checkpoint, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return checkpoint{}, nil
		}
		return checkpoint{}, err
	}
	var cp checkpoint
	if err := json.Unmarshal(data, &cp); err != nil {
		return checkpoint{}, err
	}
	return cp, nil
}

func writeCheckpoint(path string, cp checkpoint) error {
	data, err := json.MarshalIndent(cp, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
