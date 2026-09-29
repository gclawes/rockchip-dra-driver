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
	"strings"
	"syscall"
	"testing"

	"github.com/gclawes/rockchip-dra-driver/pkg/consts"
)

func TestMockEnumerateOmitsGID(t *testing.T) {
	devs, err := Enumerate(Config{Mock: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range devs {
		if d.DeviceGIDKnown {
			t.Fatalf("mock device should not invent a gid: %+v", d)
		}
	}
}

func TestNodeGID(t *testing.T) {
	if _, ok := nodeGID(filepath.Join(t.TempDir(), "missing")); ok {
		t.Fatal("missing path should not have a gid")
	}
	path := filepath.Join(t.TempDir(), "node")
	writeFile(t, path, "")
	gid, ok := nodeGID(path)
	if !ok {
		t.Fatal("expected gid")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || gid != int64(stat.Gid) {
		t.Fatalf("gid %d != stat", gid)
	}
}

func TestMockEnumerateRK3588(t *testing.T) {
	devs, err := Enumerate(Config{Mock: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(devs) != 3 {
		t.Fatalf("expected 3 devices, got %d: %+v", len(devs), devs)
	}
	if devs[0].Type != consts.TypeNPU || devs[0].CoreCount != 3 || devs[0].MaxAllocations != 3 {
		t.Fatalf("unexpected NPU: %+v", devs[0])
	}
	if devs[1].Type != consts.TypeGPU || devs[1].ShaderCores != 4 || devs[1].MaxAllocations != 8 {
		t.Fatalf("unexpected GPU: %+v", devs[1])
	}
	vpu := devs[2]
	if vpu.Name != "vpu-rkvdec-0" || vpu.Block != consts.BlockRkvdec || vpu.MaxAllocations != 1 || vpu.DeviceGIDKnown {
		t.Fatalf("unexpected VPU: %+v", vpu)
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
	// Real rocket (RK3588 / Turing RK1): accel class uevent has no DRIVER=,
	// the parent is unbound platform:rknn, and DRIVER=rocket is on the cores.
	writeFile(t, filepath.Join(root, "class/accel/accel0/uevent"), "MAJOR=261\nMINOR=0\nDEVNAME=accel/accel0\nDEVTYPE=accel_minor\n")
	writeFile(t, filepath.Join(root, "class/accel/accel0/device/uevent"), "MODALIAS=platform:rknn\n")
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
	if devs[0].DeviceNode != filepath.Join(devRoot, "accel/accel0") {
		t.Fatalf("unexpected NPU node: %+v", devs[0])
	}
	if !devs[0].DeviceGIDKnown || !devs[1].DeviceGIDKnown {
		t.Fatalf("expected gids from fixture nodes, got %+v", devs)
	}
	if devs[0].DeviceGID != fileGID(t, devs[0].DeviceNode) || devs[1].DeviceGID != fileGID(t, devs[1].DeviceNode) {
		t.Fatalf("gid mismatch: %+v", devs)
	}
	if devs[1].KMD != consts.KMDPanthor || devs[1].Model != "Mali-G610" {
		t.Fatalf("unexpected GPU: %+v", devs[1])
	}
}

func TestAccelWithoutRocketOmitsNPU(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "firmware/devicetree/base/compatible"), "rockchip,rk3588\x00")
	writeFile(t, filepath.Join(root, "class/accel/accel0/uevent"), "DEVNAME=accel/accel0\n")
	devRoot := t.TempDir()
	writeFile(t, filepath.Join(devRoot, "accel/accel0"), "")

	devs, err := Enumerate(Config{SysfsRoot: root, DevRoot: devRoot, NPUEnabled: true, GPUEnabled: false})
	if err != nil {
		t.Fatal(err)
	}
	if len(devs) != 0 {
		t.Fatalf("expected no NPU without bound rocket cores, got %+v", devs)
	}
}

func TestRocketCoresWithoutAccelClass(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "firmware/devicetree/base/compatible"), "rockchip,rk3588\x00")
	writeFile(t, filepath.Join(root, "bus/platform/drivers/rocket/fdab0000.npu"), "")

	devs, err := Enumerate(Config{SysfsRoot: root, DevRoot: t.TempDir(), NPUEnabled: true, GPUEnabled: false})
	if err != nil {
		t.Fatal(err)
	}
	if len(devs) != 1 || devs[0].DeviceNode != "/dev/accel/accel0" || devs[0].CoreCount != 1 {
		t.Fatalf("expected fallback accel node, got %+v", devs)
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

func TestVPUSysfsClassify(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "firmware/devicetree/base/compatible"), "rockchip,rk3588\x00")
	// Indexes are deliberately not in probe order. Names come from the block.
	writeV4L(t, root, "video9", "rkvdec", "rkvdec", "rockchip,rk3588-vdec\x00")
	writeV4L(t, root, "video1", "hantro-vpu", "hantro-vpu", "rockchip,rk3588-vpu121\x00")
	writeV4L(t, root, "video7", "hantro-vpu", "hantro-vpu", "rockchip,rk3588-vpu121\x00")
	writeV4L(t, root, "video4", "hantro-vpu", "hantro-vpu", "rockchip,rk3588-av1-vpu\x00")
	writeV4L(t, root, "video3", "hantro-vpu", "hantro-vpu", "rockchip,rk3588-vepu121\x00")
	writeV4L(t, root, "video0", "rockchip-rga", "rockchip-rga", "rockchip,rk3588-rga\x00")
	writeV4L(t, root, "video5", "rkisp1", "rkisp1", "rockchip,rk3588-rkisp\x00")

	devRoot := t.TempDir()
	for _, name := range []string{"video0", "video1", "video3", "video4", "video5", "video7", "video9"} {
		writeFile(t, filepath.Join(devRoot, name), "")
	}

	devs, err := Enumerate(Config{SysfsRoot: root, DevRoot: devRoot, VPUEnabled: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(devs) != 5 {
		t.Fatalf("expected 5 vpu devices, got %+v", devs)
	}
	byName := map[string]Device{}
	for _, d := range devs {
		byName[d.Name] = d
	}
	dec := byName["vpu-hantro-dec-0"]
	if dec.Function != consts.FunctionDecode || dec.KMD != consts.KMDHantro || !strings.Contains(dec.Model, "vpu121") {
		t.Fatalf("unexpected decoder: %+v", dec)
	}
	if dec.DeviceNode != filepath.Join(devRoot, "video1") || !dec.DeviceGIDKnown {
		t.Fatalf("unexpected decoder node: %+v", dec)
	}
	dec1 := byName["vpu-hantro-dec-1"]
	if dec1.DeviceNode != filepath.Join(devRoot, "video7") {
		t.Fatalf("second decoder should follow video index, got %+v", dec1)
	}
	av1 := byName["vpu-hantro-av1-0"]
	if av1.Function != consts.FunctionDecode || av1.Block != consts.BlockHantroAV1 {
		t.Fatalf("unexpected av1: %+v", av1)
	}
	enc := byName["vpu-hantro-enc-0"]
	if enc.Function != consts.FunctionEncode || enc.Block != consts.BlockHantroEnc {
		t.Fatalf("unexpected encoder: %+v", enc)
	}
	rkv := byName["vpu-rkvdec-0"]
	if rkv.Block != consts.BlockRkvdec || rkv.DeviceNode != filepath.Join(devRoot, "video9") || rkv.MaxAllocations != 1 {
		t.Fatalf("unexpected rkvdec: %+v", rkv)
	}
}

func TestVPUNameFallbackAndCap(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "firmware/devicetree/base/compatible"), "rockchip,rk3588\x00")
	writeV4L(t, root, "video2", "hantro-vpu", "rockchip-vpu-dec", "")
	writeV4L(t, root, "video8", "hantro-vpu", "unmatched-hantro", "")

	devs, err := Enumerate(Config{SysfsRoot: root, DevRoot: t.TempDir(), VPUEnabled: true, VPUMaxAllocations: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(devs) != 1 || devs[0].Block != consts.BlockHantroDec || devs[0].MaxAllocations != 2 {
		t.Fatalf("unexpected devices: %+v", devs)
	}
	if devs[0].DeviceNode != "/dev/video2" || devs[0].DeviceGIDKnown {
		t.Fatalf("missing node should fall back without a gid: %+v", devs[0])
	}
}

func writeV4L(t *testing.T, root, video, driver, card, compatible string) {
	t.Helper()
	class := filepath.Join(root, "class", "video4linux", video)
	writeFile(t, filepath.Join(class, "uevent"), "DEVNAME="+video+"\n")
	writeFile(t, filepath.Join(class, "name"), card+"\n")
	writeFile(t, filepath.Join(class, "device", "uevent"), "DRIVER="+driver+"\n")
	if compatible != "" {
		writeFile(t, filepath.Join(class, "device", "of_node", "compatible"), compatible)
	}
}

func fileGID(t *testing.T, path string) int64 {
	t.Helper()
	gid, ok := nodeGID(path)
	if !ok {
		t.Fatalf("stat %s", path)
	}
	return gid
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
