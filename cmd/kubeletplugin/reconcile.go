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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"k8s.io/dynamic-resource-allocation/kubeletplugin"
	"k8s.io/dynamic-resource-allocation/resourceslice"
	"k8s.io/klog/v2"

	"github.com/gclawes/rockchip-dra-driver/internal/discovery"
	"github.com/gclawes/rockchip-dra-driver/pkg/consts"
)

const (
	// defaultDiscoveryInterval is how often the plugin re-reads sysfs.
	// Polling rather than inotify: sysfs bind/unbind notifications are
	// unreliable, and this stays under the kubelet health lease.
	defaultDiscoveryInterval = 10 * time.Second

	// healthLease is the kubelet's default device-health timeout. Reports
	// that are not refreshed within this window are treated as unknown.
	healthLease = 30 * time.Second
)

// healthResendInterval is how often WatchHealthStatus repeats the last
// report so the kubelet does not expire it. A variable so tests can
// shorten it. Must stay under healthLease.
var healthResendInterval = 15 * time.Second

// statDeviceNode is how the plugin checks that an advertised char device
// exists. Tests replace it so a missing node does not depend on the host /dev.
var statDeviceNode = os.Stat

// trackedDevice is a discovered accelerator plus the health the plugin last
// decided for it. Retained devices are no longer enumerated but stay published
// because a prepared claim still names them.
type trackedDevice struct {
	discovery.Device
	status     kubeletplugin.HealthStatus
	message    string
	nodeSeen   bool
	taintedAt  time.Time
	taintValue string
	retained   bool
}

type reconcileResult struct {
	ResourcesChanged bool
	HealthChanged    bool
	Resources        resourceslice.DriverResources
}

// Reconcile re-reads sysfs and updates the published device set.
//
// A device that disappears is kept, tainted, and reported unhealthy while any
// prepared claim still references it. It is dropped only after the last such
// claim is unprepared. A char device that has never been seen is reported
// unknown and is not tainted: udev may still be creating it.
func (s *DeviceState) Reconcile(ctx context.Context) (reconcileResult, error) {
	devices, err := discovery.Enumerate(s.discCfg)
	if err != nil {
		return reconcileResult{}, fmt.Errorf("enumerate devices: %w", err)
	}
	s.Lock()
	defer s.Unlock()
	return s.applyDiscoveryLocked(ctx, devices)
}

func (s *DeviceState) applyDiscoveryLocked(ctx context.Context, found []discovery.Device) (reconcileResult, error) {
	logger := klog.FromContext(ctx)

	cp, err := readCheckpoint(s.checkpointPath)
	if err != nil {
		return reconcileResult{}, fmt.Errorf("read checkpoint: %w", err)
	}
	claims, err := readClaimDevices(s.claimDevicesPath)
	if err != nil {
		return reconcileResult{}, fmt.Errorf("read claim devices: %w", err)
	}
	if pruneClaimDevices(cp, &claims) {
		if err := writeClaimDevices(s.claimDevicesPath, claims); err != nil {
			return reconcileResult{}, fmt.Errorf("prune claim devices: %w", err)
		}
	}
	snapshots := claimedSnapshots(cp, claims)

	foundByName := make(map[string]discovery.Device, len(found))
	for _, d := range found {
		foundByName[d.Name] = d
	}

	next := make(map[string]trackedDevice, len(found)+len(snapshots))
	for _, d := range found {
		prev, had := s.tracked[d.Name]
		next[d.Name] = assessEnumerated(d, prev, had, s.deviceNodeExists(d.DeviceNode))
	}
	for name, snap := range snapshots {
		if _, ok := foundByName[name]; ok {
			continue
		}
		prev, had := s.tracked[name]
		if !had {
			// Restart, or the in-memory set was empty: republish the snapshot
			// taken when the claim was prepared.
			prev = trackedDevice{Device: snap}
		}
		td := markUnavailable(prev, consts.TaintValueNotDiscovered, "device no longer discovered")
		td.retained = true
		next[name] = td
	}

	for name, td := range next {
		prev, had := s.tracked[name]
		switch {
		case !had && td.status == kubeletplugin.HealthStatusUnhealthy:
			logger.Info("Retaining device for prepared claim", "device", name, "message", td.message)
		case !had:
			logger.Info("Discovered device", "device", name, "type", td.Type, "health", td.status, "node", td.DeviceNode)
		case prev.status != td.status || prev.message != td.message:
			logger.Info("Device health changed", "device", name, "from", prev.status, "to", td.status, "message", td.message)
		}
	}
	for name := range s.tracked {
		if _, ok := next[name]; !ok {
			logger.Info("Removed device", "device", name)
		}
	}

	resources := devicesToResources(s.nodeName, trackedList(next))
	result := reconcileResult{
		ResourcesChanged: !resourcesEqual(s.driverResources, resources),
		HealthChanged:    !healthEqual(s.tracked, next),
		Resources:        resources,
	}
	s.tracked = next
	s.driverResources = resources
	return result, nil
}

func assessEnumerated(d discovery.Device, prev trackedDevice, had, nodeOK bool) trackedDevice {
	td := trackedDevice{Device: d}
	if had {
		td.nodeSeen = prev.nodeSeen
		td.taintedAt = prev.taintedAt
	}
	if nodeOK {
		td.nodeSeen = true
		td.status = kubeletplugin.HealthStatusHealthy
		td.taintedAt = time.Time{}
		td.taintValue = ""
		return td
	}
	if td.nodeSeen {
		return markUnavailable(td, consts.TaintValueNodeMissing, "device node missing")
	}
	td.status = kubeletplugin.HealthStatusUnknown
	td.message = "device node not yet present"
	td.taintedAt = time.Time{}
	td.taintValue = ""
	return td
}

func markUnavailable(td trackedDevice, value, message string) trackedDevice {
	td.status = kubeletplugin.HealthStatusUnhealthy
	td.message = message
	td.taintValue = value
	if td.taintedAt.IsZero() {
		td.taintedAt = time.Now()
	}
	return td
}

func (s *DeviceState) deviceNodeExists(path string) bool {
	if s.mock || path == "" {
		return s.mock
	}
	info, err := statDeviceNode(path)
	return err == nil && !info.IsDir()
}

func (s *DeviceState) healthReport() kubeletplugin.DeviceHealthReport {
	s.Lock()
	defer s.Unlock()
	devices := trackedList(s.tracked)
	report := kubeletplugin.DeviceHealthReport{
		Devices: make([]kubeletplugin.DeviceHealth, 0, len(devices)),
	}
	now := time.Now()
	for _, td := range devices {
		status := td.status
		if status == "" {
			status = kubeletplugin.HealthStatusUnknown
		}
		report.Devices = append(report.Devices, kubeletplugin.DeviceHealth{
			PoolName:           s.nodeName,
			DeviceName:         td.Name,
			Health:             status,
			LastUpdated:        now,
			HealthCheckTimeout: healthLease,
			Message:            td.message,
		})
	}
	return report
}

func (td trackedDevice) prepareError(mock bool, nodeExists bool) error {
	if td.status == kubeletplugin.HealthStatusUnhealthy {
		return fmt.Errorf("device %s is unhealthy (%s=%s): %s", td.Name, consts.TaintUnhealthy, td.taintValue, td.message)
	}
	if !mock && td.DeviceNode != "" && !nodeExists {
		return fmt.Errorf("device %s node %s is not available", td.Name, td.DeviceNode)
	}
	return nil
}

func trackedList(m map[string]trackedDevice) []trackedDevice {
	out := make([]trackedDevice, 0, len(m))
	for _, td := range m {
		out = append(out, td)
	}
	return out
}

func healthEqual(a, b map[string]trackedDevice) bool {
	if len(a) != len(b) {
		return false
	}
	for name, da := range a {
		db, ok := b[name]
		if !ok || da.status != db.status || da.message != db.message {
			return false
		}
	}
	return true
}

func resourcesEqual(a, b resourceslice.DriverResources) bool {
	aj, errA := json.Marshal(a)
	bj, errB := json.Marshal(b)
	if errA != nil || errB != nil {
		return false
	}
	return bytes.Equal(aj, bj)
}

func claimedSnapshots(cp checkpoint, file claimDevicesFile) map[string]discovery.Device {
	prepared := make(map[string]struct{}, len(cp.PreparedClaims))
	for _, uid := range cp.PreparedClaims {
		prepared[string(uid)] = struct{}{}
	}
	out := map[string]discovery.Device{}
	for uid, devs := range file.Claims {
		if _, ok := prepared[uid]; !ok {
			continue
		}
		for _, d := range devs {
			if d.Name == "" {
				continue
			}
			out[d.Name] = d
		}
	}
	return out
}
