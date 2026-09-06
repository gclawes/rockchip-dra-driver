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
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
)

func TestSchemeRoundTrip(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme: %v", err)
	}

	npu := DefaultNpuConfig()
	if gvk := npu.GroupVersionKind(); gvk.Group != GroupName || gvk.Kind != NpuConfigKind {
		t.Fatalf("unexpected NpuConfig GVK: %v", gvk)
	}
	if npu.DeepCopyObject() == nil {
		t.Fatal("NpuConfig DeepCopyObject returned nil")
	}

	gpu := DefaultGpuConfig()
	if gvk := gpu.GroupVersionKind(); gvk.Group != GroupName || gvk.Kind != GpuConfigKind {
		t.Fatalf("unexpected GpuConfig GVK: %v", gvk)
	}
	if gpu.DeepCopyObject() == nil {
		t.Fatal("GpuConfig DeepCopyObject returned nil")
	}
}
