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
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	resourceapi "k8s.io/api/resource/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/dynamic-resource-allocation/kubeletplugin"

	"github.com/gclawes/rockchip-dra-driver/pkg/consts"
)

func TestReconcileNodeMissingTaintsAndFailsPrepare(t *testing.T) {
	sysfs, dev := rocketFixture(t, true)
	st := newTestState(t, sysfs, dev, false)
	if got := mustTracked(t, st, "npu-0"); got.status != kubeletplugin.HealthStatusHealthy || len(publishedTaints(t, st, "npu-0")) != 0 {
		t.Fatalf("expected healthy untainted npu, got %+v", got)
	}

	nodePresent := false
	st.Lock()
	orig := statDeviceNode
	statDeviceNode = func(string) (os.FileInfo, error) {
		if nodePresent {
			return orig(filepath.Join(dev, "accel/accel0"))
		}
		return nil, os.ErrNotExist
	}
	st.Unlock()
	t.Cleanup(func() { statDeviceNode = orig })

	result, err := st.Reconcile(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !result.ResourcesChanged || !result.HealthChanged {
		t.Fatalf("expected publish and health change, got %+v", result)
	}
	got := mustTracked(t, st, "npu-0")
	if got.status != kubeletplugin.HealthStatusUnhealthy || got.taintValue != consts.TaintValueNodeMissing || got.retained {
		t.Fatalf("expected node-missing taint, got %+v", got)
	}
	taints := publishedTaints(t, st, "npu-0")
	if len(taints) != 1 || taints[0].Effect != resourceapi.DeviceTaintEffectNoExecute || taints[0].TimeAdded == nil {
		t.Fatalf("unexpected published taint: %+v", taints)
	}
	added := taints[0].TimeAdded.Time

	again, err := st.Reconcile(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if again.ResourcesChanged || again.HealthChanged {
		t.Fatalf("unchanged rediscovery should not republish: %+v", again)
	}
	taints = publishedTaints(t, st, "npu-0")
	if taints[0].TimeAdded == nil || !taints[0].TimeAdded.Time.Equal(added) {
		t.Fatalf("taint time changed: %v -> %v", added, taints[0].TimeAdded)
	}

	_, err = st.Prepare(context.Background(), npuClaim("claim-1"))
	if err == nil || !strings.Contains(err.Error(), "unhealthy") {
		t.Fatalf("expected prepare to fail unhealthy, got %v", err)
	}
}

func TestReconcileUnknownUntilNodeAppears(t *testing.T) {
	sysfs, dev := rocketFixture(t, false)
	orig := statDeviceNode
	statDeviceNode = func(string) (os.FileInfo, error) { return nil, os.ErrNotExist }
	t.Cleanup(func() { statDeviceNode = orig })

	st := newTestState(t, sysfs, dev, false)
	got := mustTracked(t, st, "npu-0")
	if got.status != kubeletplugin.HealthStatusUnknown || len(publishedTaints(t, st, "npu-0")) != 0 {
		t.Fatalf("expected unknown and untainted, got %+v taints %+v", got, publishedTaints(t, st, "npu-0"))
	}
	if _, err := st.Prepare(context.Background(), npuClaim("claim-1")); err == nil {
		t.Fatal("expected prepare to fail while the node is missing")
	}

	statDeviceNode = func(string) (os.FileInfo, error) { return orig(filepath.Join(dev, "accel/accel0")) }
	// The fixture has no char device; pretend the stat succeeded via a stand-in file.
	standIn := filepath.Join(t.TempDir(), "node")
	if err := os.WriteFile(standIn, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	statDeviceNode = func(string) (os.FileInfo, error) { return os.Stat(standIn) }

	if _, err := st.Reconcile(context.Background()); err != nil {
		t.Fatal(err)
	}
	got = mustTracked(t, st, "npu-0")
	if got.status != kubeletplugin.HealthStatusHealthy || !got.nodeSeen {
		t.Fatalf("expected healthy after node appeared, got %+v", got)
	}
}

func TestReconcileRetainsDisappearedDeviceUntilUnprepare(t *testing.T) {
	sysfs, dev := rocketFixture(t, true)
	st := newTestState(t, sysfs, dev, false)
	if _, err := st.Prepare(context.Background(), npuClaim("claim-1")); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll(filepath.Join(sysfs, "bus/platform/drivers/rocket")); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(sysfs, "class/accel")); err != nil {
		t.Fatal(err)
	}
	result, err := st.Reconcile(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !result.ResourcesChanged {
		t.Fatal("expected retained device to stay published with a taint")
	}
	got := mustTracked(t, st, "npu-0")
	if !got.retained || got.status != kubeletplugin.HealthStatusUnhealthy || got.taintValue != consts.TaintValueNotDiscovered {
		t.Fatalf("expected retained not-discovered device, got %+v", got)
	}
	if _, err := st.Prepare(context.Background(), npuClaim("claim-2")); err == nil {
		t.Fatal("expected new prepare to fail")
	}

	if err := st.Unprepare("claim-1"); err != nil {
		t.Fatal(err)
	}
	result, err = st.Reconcile(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !result.ResourcesChanged {
		t.Fatal("expected device to be dropped after the last claim")
	}
	st.Lock()
	_, ok := st.tracked["npu-0"]
	st.Unlock()
	if ok {
		t.Fatal("device still published after unprepare")
	}
}

func TestReconcileWithoutSnapshotDoesNotInventDevice(t *testing.T) {
	sysfs, dev := rocketFixture(t, true)
	st := newTestState(t, sysfs, dev, false)
	if err := writeCheckpoint(st.checkpointPath, checkpoint{PreparedClaims: []types.UID{"old-claim"}}); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(sysfs, "bus/platform/drivers/rocket")); err != nil {
		t.Fatal(err)
	}
	if _, err := st.Reconcile(context.Background()); err != nil {
		t.Fatal(err)
	}
	st.Lock()
	defer st.Unlock()
	if _, ok := st.tracked["npu-0"]; ok {
		t.Fatal("checkpoint without a device snapshot must not keep a disappeared device")
	}
}

func TestReconcileDropsUnclaimedDevice(t *testing.T) {
	sysfs, dev := rocketFixture(t, true)
	st := newTestState(t, sysfs, dev, false)
	if err := os.RemoveAll(filepath.Join(sysfs, "bus/platform/drivers/rocket")); err != nil {
		t.Fatal(err)
	}
	if _, err := st.Reconcile(context.Background()); err != nil {
		t.Fatal(err)
	}
	st.Lock()
	defer st.Unlock()
	if len(st.tracked) != 0 {
		t.Fatalf("expected no devices, got %+v", st.tracked)
	}
}

func TestReconcileDeviceAppears(t *testing.T) {
	sysfs := t.TempDir()
	dev := t.TempDir()
	writeFile(t, filepath.Join(sysfs, "firmware/devicetree/base/compatible"), "rockchip,rk3588\x00")
	st := newTestState(t, sysfs, dev, false)
	if n := len(st.Resources().Pools["node-a"].Slices[0].Devices); n != 0 {
		t.Fatalf("expected no devices, got %d", n)
	}

	writeRocket(t, sysfs, dev, true)
	result, err := st.Reconcile(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !result.ResourcesChanged || !result.HealthChanged {
		t.Fatalf("expected new device to publish, got %+v", result)
	}
	if mustTracked(t, st, "npu-0").status != kubeletplugin.HealthStatusHealthy {
		t.Fatal("appeared device should be healthy")
	}
}

func TestMockDevicesStayHealthyWithoutNodes(t *testing.T) {
	st := newTestState(t, t.TempDir(), t.TempDir(), true)
	got := mustTracked(t, st, "npu-0")
	if got.status != kubeletplugin.HealthStatusHealthy {
		t.Fatalf("mock device unhealthy: %+v", got)
	}
	report := st.healthReport()
	if len(report.Devices) != 2 {
		t.Fatalf("expected npu and gpu health, got %+v", report.Devices)
	}
	for _, d := range report.Devices {
		if d.PoolName != "node-a" || d.Health != kubeletplugin.HealthStatusHealthy || d.HealthCheckTimeout != healthLease {
			t.Fatalf("unexpected health entry: %+v", d)
		}
	}
	if _, err := st.Prepare(context.Background(), npuClaim("claim-1")); err != nil {
		t.Fatal(err)
	}
}

func TestWatchHealthStatusResendsAndUnblocks(t *testing.T) {
	old := healthResendInterval
	healthResendInterval = 20 * time.Millisecond
	t.Cleanup(func() { healthResendInterval = old })

	st := newTestState(t, t.TempDir(), t.TempDir(), true)
	d := &driver{state: st, healthChanged: make(chan struct{}, 1)}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	reports := make(chan kubeletplugin.DeviceHealthReport, 4)
	errCh := make(chan error, 1)
	go func() { errCh <- d.WatchHealthStatus(ctx, reports) }()

	waitReport := func() kubeletplugin.DeviceHealthReport {
		t.Helper()
		select {
		case report := <-reports:
			return report
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for health report")
			return kubeletplugin.DeviceHealthReport{}
		}
	}
	if len(waitReport().Devices) != 2 {
		t.Fatal("expected an initial report covering both devices")
	}
	if len(waitReport().Devices) != 2 {
		t.Fatal("expected a resent report")
	}
	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("WatchHealthStatus did not return on cancel")
	}

	blocked := make(chan kubeletplugin.DeviceHealthReport)
	ctx2, cancel2 := context.WithCancel(context.Background())
	errCh = make(chan error, 1)
	go func() { errCh <- d.WatchHealthStatus(ctx2, blocked) }()
	time.Sleep(20 * time.Millisecond)
	cancel2()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("WatchHealthStatus blocked after cancel")
	}
}

func newTestState(t *testing.T, sysfs, dev string, mock bool) *DeviceState {
	t.Helper()
	pluginRoot := t.TempDir()
	cfg := &Config{flags: &Flags{
		nodeName:                    "node-a",
		driverName:                  consts.DriverName,
		cdiRoot:                     t.TempDir(),
		kubeletPluginsDirectoryPath: pluginRoot,
		sysfsRoot:                   sysfs,
		devRoot:                     dev,
		mockDevices:                 mock,
		npuEnabled:                  true,
		gpuEnabled:                  mock,
	}}
	if err := os.MkdirAll(cfg.DriverPluginPath(), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cfg.flags.cdiRoot, 0o750); err != nil {
		t.Fatal(err)
	}
	st, err := NewDeviceState(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	return st
}

func rocketFixture(t *testing.T, withNode bool) (string, string) {
	t.Helper()
	sysfs := t.TempDir()
	dev := t.TempDir()
	writeFile(t, filepath.Join(sysfs, "firmware/devicetree/base/compatible"), "rockchip,rk3588\x00")
	writeRocket(t, sysfs, dev, withNode)
	return sysfs, dev
}

func writeRocket(t *testing.T, sysfs, dev string, withNode bool) {
	t.Helper()
	writeFile(t, filepath.Join(sysfs, "bus/platform/drivers/rocket/fdab0000.npu"), "")
	writeFile(t, filepath.Join(sysfs, "class/accel/accel0/uevent"), "MAJOR=261\nMINOR=0\nDEVNAME=accel/accel0\n")
	writeFile(t, filepath.Join(sysfs, "class/accel/accel0/device/uevent"), "MODALIAS=platform:rknn\n")
	if withNode {
		writeFile(t, filepath.Join(dev, "accel/accel0"), "")
	}
}

func npuClaim(uid string) *resourceapi.ResourceClaim {
	return &resourceapi.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{UID: types.UID(uid), Name: "npu", Namespace: "default"},
		Status: resourceapi.ResourceClaimStatus{
			Allocation: &resourceapi.AllocationResult{
				Devices: resourceapi.DeviceAllocationResult{
					Results: []resourceapi.DeviceRequestAllocationResult{{
						Driver:  consts.DriverName,
						Pool:    "node-a",
						Device:  "npu-0",
						Request: "npu",
					}},
				},
			},
		},
	}
}

func mustTracked(t *testing.T, st *DeviceState, name string) trackedDevice {
	t.Helper()
	st.Lock()
	defer st.Unlock()
	td, ok := st.tracked[name]
	if !ok {
		t.Fatalf("device %s not tracked", name)
	}
	return td
}

func publishedTaints(t *testing.T, st *DeviceState, name string) []resourceapi.DeviceTaint {
	t.Helper()
	st.Lock()
	defer st.Unlock()
	for _, dev := range st.driverResources.Pools[st.nodeName].Slices[0].Devices {
		if dev.Name == name {
			return dev.Taints
		}
	}
	t.Fatalf("device %s not published", name)
	return nil
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
