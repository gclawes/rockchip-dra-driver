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
	"os"
	"path/filepath"
	"testing"

	"github.com/gclawes/rockchip-dra-driver/pkg/consts"
)

func TestMockEnumerateRK3588(t *testing.T) {
	devs, err := Enumerate(Config{Mock: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(devs) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(devs))
	}
	if devs[0].Type != consts.TypeNPU || devs[0].CoreCount != 3 || devs[0].MaxAllocations != 3 {
		t.Fatalf("unexpected NPU: %+v", devs[0])
	}
	if devs[1].Type != consts.TypeGPU || devs[1].ShaderCores != 4 || devs[1].MaxAllocations != 8 {
		t.Fatalf("unexpected GPU: %+v", devs[1])
	}
}

func TestMockEnumerateCapsOverride(t *testing.T) {
	devs, err := Enumerate(Config{Mock: true, NPUMaxAllocations: 1, GPUMaxAllocations: 2, NPUEnabled: true, GPUEnabled: true})
	if err != nil {
		t.Fatal(err)
	}
	if devs[0].MaxAllocations != 1 || devs[1].MaxAllocations != 2 {
		t.Fatalf("overrides not applied: %+v", devs)
	}
}

func TestMockEnumerateDisableGPU(t *testing.T) {
	devs, err := Enumerate(Config{Mock: true, NPUEnabled: true, GPUEnabled: false})
	if err != nil {
		t.Fatal(err)
	}
	if len(devs) != 1 || devs[0].Type != consts.TypeNPU {
		t.Fatalf("expected only NPU, got %+v", devs)
	}
}

func TestSysfsEnumerate(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "firmware/devicetree/base/compatible"), "rockchip,rk3588\x00rockchip,rk3588-rock-5b\x00")
	writeFile(t, filepath.Join(root, "class/accel/accel0/uevent"), "DRIVER=rocket\n")
	writeFile(t, filepath.Join(root, "bus/platform/drivers/rocket/fdab0000.npu"), "")
	writeFile(t, filepath.Join(root, "bus/platform/drivers/rocket/fdac0000.npu"), "")
	writeFile(t, filepath.Join(root, "bus/platform/drivers/rocket/fdad0000.npu"), "")
	writeFile(t, filepath.Join(root, "class/drm/renderD128/device/uevent"), "DRIVER=panthor\n")

	devRoot := t.TempDir()
	mustMkdir(t, filepath.Join(devRoot, "accel"))
	mustMkdir(t, filepath.Join(devRoot, "dri"))
	writeFile(t, filepath.Join(devRoot, "accel/accel0"), "")
	writeFile(t, filepath.Join(devRoot, "dri/renderD128"), "")

	devs, err := Enumerate(Config{SysfsRoot: root, DevRoot: devRoot, NPUEnabled: true, GPUEnabled: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(devs) != 2 {
		t.Fatalf("expected 2 devices, got %+v", devs)
	}
	if devs[0].SoC != "rk3588" || devs[0].KMD != consts.KMDRocket || devs[0].CoreCount != 3 {
		t.Fatalf("unexpected NPU: %+v", devs[0])
	}
	if devs[1].KMD != consts.KMDPanthor || devs[1].Model != "Mali-G610" {
		t.Fatalf("unexpected GPU: %+v", devs[1])
	}
}

func TestMissingKMDOmitsDevice(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "firmware/devicetree/base/compatible"), "rockchip,rk3588\x00")
	devs, err := Enumerate(Config{SysfsRoot: root, DevRoot: t.TempDir(), NPUEnabled: true, GPUEnabled: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(devs) != 0 {
		t.Fatalf("expected no devices, got %+v", devs)
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	mustMkdir(t, filepath.Dir(path))
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mustMkdir(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
}
