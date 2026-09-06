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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const GpuConfigKind = "GpuConfig"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GpuConfig holds opaque parameters for a Rockchip GPU allocation.
//
// Mainline panthor multiplexes clients through drm_sched. There is no
// time-slicing or shader-core isolation UAPI, so this type is an empty stub
// reserved for future knobs.
type GpuConfig struct {
	metav1.TypeMeta `json:",inline"`
}

// DefaultGpuConfig returns an empty GPU config with TypeMeta filled in.
func DefaultGpuConfig() *GpuConfig {
	return &GpuConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupName + "/" + Version,
			Kind:       GpuConfigKind,
		},
	}
}
