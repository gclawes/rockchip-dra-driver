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
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto copies all properties of this object into another object of the
// same type.
func (in *NpuConfig) DeepCopyInto(out *NpuConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
}

// DeepCopy creates a new instance of the type and copies all properties from this object.
func (in *NpuConfig) DeepCopy() *NpuConfig {
	if in == nil {
		return nil
	}
	out := new(NpuConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject returns a generically typed copy of an object.
func (in *NpuConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto copies all properties of this object into another object of the
// same type.
func (in *GpuConfig) DeepCopyInto(out *GpuConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
}

// DeepCopy creates a new instance of the type and copies all properties from this object.
func (in *GpuConfig) DeepCopy() *GpuConfig {
	if in == nil {
		return nil
	}
	out := new(GpuConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject returns a generically typed copy of an object.
func (in *GpuConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
