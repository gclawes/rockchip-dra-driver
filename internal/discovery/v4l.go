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
	"strconv"
	"strings"
)

// v4lNode is one /sys/class/video4linux entry. Index is only a tiebreak.
// videoN is not a stable name across boots.
type v4lNode struct {
	Index      int
	Driver     string
	CardName   string
	Compatible string
	DeviceNode string
}

func discoverV4L(cfg Config) []v4lNode {
	classDir := filepath.Join(cfg.SysfsRoot, "class", "video4linux")
	entries, err := os.ReadDir(classDir)
	if err != nil {
		return nil
	}
	var nodes []v4lNode
	for _, e := range entries {
		name := e.Name()
		index, ok := videoIndex(name)
		if !ok {
			continue
		}
		classPath := filepath.Join(classDir, name)
		devicePath := filepath.Join(classPath, "device")
		driver := ueventDriver(classPath)
		if driver == "" {
			driver = ueventDriver(devicePath)
		}
		if driver == "" {
			driver = driverSymlink(devicePath)
		}
		devNode := filepath.Join(cfg.DevRoot, name)
		if !exists(devNode) {
			devNode = filepath.Join("/dev", name)
		}
		nodes = append(nodes, v4lNode{
			Index:      index,
			Driver:     driver,
			CardName:   readTrimmed(filepath.Join(classPath, "name")),
			Compatible: readCompatible(devicePath),
			DeviceNode: devNode,
		})
	}
	return nodes
}

func videoIndex(name string) (int, bool) {
	rest, ok := strings.CutPrefix(name, "video")
	if !ok || rest == "" {
		return 0, false
	}
	n, err := strconv.Atoi(rest)
	if err != nil {
		return 0, false
	}
	return n, true
}

func driverSymlink(devicePath string) string {
	target, err := os.Readlink(filepath.Join(devicePath, "driver"))
	if err != nil {
		return ""
	}
	return filepath.Base(target)
}

func readCompatible(devicePath string) string {
	data, err := os.ReadFile(filepath.Join(devicePath, "of_node", "compatible"))
	if err != nil {
		return ""
	}
	parts := strings.Split(string(data), "\x00")
	var kept []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			kept = append(kept, p)
		}
	}
	return strings.Join(kept, ",")
}

func readTrimmed(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
