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
)

// knownSoCs is ordered most-specific first so rk3588s wins over a generic
// rockchip compatible string.
var knownSoCs = []string{
	"rk3588s",
	"rk3588",
	"rk3576",
	"rk3568",
	"rk3566",
	"rk3562",
}

func detectSoC(sysfsRoot string) string {
	data, err := os.ReadFile(filepath.Join(sysfsRoot, "firmware", "devicetree", "base", "compatible"))
	if err != nil {
		return ""
	}
	// Device-tree compatible is NUL-separated.
	compat := strings.ReplaceAll(string(data), "\x00", ",")
	compat = strings.ToLower(compat)
	for _, soc := range knownSoCs {
		if strings.Contains(compat, soc) {
			return soc
		}
	}
	return ""
}

func ueventDriver(path string) string {
	data, err := os.ReadFile(filepath.Join(path, "uevent"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		key, val, ok := strings.Cut(line, "=")
		if ok && key == "DRIVER" {
			return strings.TrimSpace(val)
		}
	}
	return ""
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
