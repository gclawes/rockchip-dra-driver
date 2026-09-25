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

const claimDevicesFileName = "claim-devices.json"

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

type claimDevicesFile struct {
	Claims map[string][]discovery.Device `json:"claims"`
}

type DeviceState struct {
	sync.Mutex
	driverName       string
	nodeName         string
	mock             bool
	discCfg          discovery.Config
	cdi              *CDIHandler
	driverResources  resourceslice.DriverResources
	tracked          map[string]trackedDevice
	checkpointPath   string
	claimDevicesPath string
}

func NewDeviceState(ctx context.Context, config *Config) (*DeviceState, error) {
	cdi, err := NewCDIHandler(config.flags.cdiRoot, config.flags.driverName, config.flags.mockDevices)
	if err != nil {
		return nil, fmt.Errorf("unable to create CDI handler: %w", err)
	}

	s := &DeviceState{
		driverName:       config.flags.driverName,
		nodeName:         config.flags.nodeName,
		mock:             config.flags.mockDevices,
		discCfg:          config.discoveryConfig(),
		cdi:              cdi,
		tracked:          map[string]trackedDevice{},
		checkpointPath:   filepath.Join(config.DriverPluginPath(), driverPluginCheckpointFile),
		claimDevicesPath: filepath.Join(config.DriverPluginPath(), claimDevicesFileName),
	}
	if _, err := s.Reconcile(ctx); err != nil {
		return nil, err
	}
	return s, nil
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
	if err := s.recordClaimDevicesLocked(claim.UID, deviceSnapshots(prepared)); err != nil {
		return nil, fmt.Errorf("unable to record prepared devices: %w", err)
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

	// Drop the claim from the checkpoint before forgetting its device
	// snapshot. A crash in between still retains the device until the next
	// reconcile prunes the orphan snapshot; the other order would drop a
	// device while the claim was still prepared.
	cp.PreparedClaims = slices.DeleteFunc(cp.PreparedClaims, func(u types.UID) bool { return u == claimUID })
	if err := writeCheckpoint(s.checkpointPath, cp); err != nil {
		return fmt.Errorf("unable to write checkpoint: %w", err)
	}
	if err := s.deleteClaimDevicesLocked(claimUID); err != nil {
		return fmt.Errorf("unable to forget prepared devices: %w", err)
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
		td, ok := s.tracked[result.Device]
		if !ok {
			return nil, fmt.Errorf("requested device is not allocatable: %s", result.Device)
		}
		if err := td.prepareError(s.mock, s.deviceNodeExists(td.DeviceNode)); err != nil {
			return nil, err
		}
		prepared = append(prepared, preparedDevice{
			RequestNames: []string{result.Request},
			PoolName:     result.Pool,
			DeviceName:   result.Device,
			CDIDeviceIDs: s.cdi.GetClaimDevices(string(claim.UID), result.Device, result.ShareID),
			ShareID:      result.ShareID,
			discovered:   td.Device,
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

func (s *DeviceState) Resources() resourceslice.DriverResources {
	s.Lock()
	defer s.Unlock()
	return s.driverResources
}

func (s *DeviceState) recordClaimDevicesLocked(uid types.UID, devs []discovery.Device) error {
	file, err := readClaimDevices(s.claimDevicesPath)
	if err != nil {
		return err
	}
	file.Claims[string(uid)] = devs
	return writeClaimDevices(s.claimDevicesPath, file)
}

func (s *DeviceState) deleteClaimDevicesLocked(uid types.UID) error {
	file, err := readClaimDevices(s.claimDevicesPath)
	if err != nil {
		return err
	}
	if _, ok := file.Claims[string(uid)]; !ok {
		return nil
	}
	delete(file.Claims, string(uid))
	return writeClaimDevices(s.claimDevicesPath, file)
}

func deviceSnapshots(prepared []preparedDevice) []discovery.Device {
	out := make([]discovery.Device, 0, len(prepared))
	for _, p := range prepared {
		out = append(out, p.discovered)
	}
	return out
}

func readClaimDevices(path string) (claimDevicesFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return claimDevicesFile{Claims: map[string][]discovery.Device{}}, nil
		}
		return claimDevicesFile{}, err
	}
	var file claimDevicesFile
	if err := json.Unmarshal(data, &file); err != nil {
		return claimDevicesFile{}, err
	}
	if file.Claims == nil {
		file.Claims = map[string][]discovery.Device{}
	}
	return file, nil
}

func writeClaimDevices(path string, file claimDevicesFile) error {
	if file.Claims == nil {
		file.Claims = map[string][]discovery.Device{}
	}
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+"-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpName)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	cleanup = false
	return nil
}

func pruneClaimDevices(cp checkpoint, file *claimDevicesFile) bool {
	if len(file.Claims) == 0 {
		return false
	}
	prepared := make(map[string]struct{}, len(cp.PreparedClaims))
	for _, uid := range cp.PreparedClaims {
		prepared[string(uid)] = struct{}{}
	}
	changed := false
	for uid := range file.Claims {
		if _, ok := prepared[uid]; !ok {
			delete(file.Claims, uid)
			changed = true
		}
	}
	return changed
}
