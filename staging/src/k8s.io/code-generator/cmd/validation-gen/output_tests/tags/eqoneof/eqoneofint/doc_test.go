/*
Copyright 2025 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package eqoneofint

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"
)

func Test(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	// Test valid values.
	st.Value(&Struct{
		IntField:                 2,
		IntPtrField:              ptr.To(2),
		Int16Field:               2,
		Int32Field:               2,
		Int64Field:               2,
		UintField:                2,
		UintPtrField:             ptr.To(uint(2)),
		Uint16Field:              2,
		Uint32Field:              2,
		Uint64Field:              2,
		IntWithNegativeField:     -2,
		TypedefField:             2,
		TypedefPtrField:          ptr.To(IntType(2)),
		ValidatedTypedefIntField: 2,
	}).ExpectValid()

	// Test invalid values.
	invalid := &Struct{
		IntField:                 4,
		IntPtrField:              ptr.To(4),
		Int16Field:               4,
		Int32Field:               4,
		Int64Field:               4,
		UintField:                4,
		UintPtrField:             ptr.To(uint(4)),
		Uint16Field:              4,
		Uint32Field:              4,
		Uint64Field:              4,
		IntWithNegativeField:     -4,
		TypedefField:             4,
		TypedefPtrField:          ptr.To(IntType(4)),
		ValidatedTypedefIntField: 4,
	}
	st.Value(invalid).ExpectInvalid(
		field.NotSupported(field.NewPath("intField"), "4", []string{"1", "2", "3"}),
		field.NotSupported(field.NewPath("intPtrField"), "4", []string{"1", "2", "3"}),
		field.NotSupported(field.NewPath("int16Field"), "4", []string{"1", "2", "3"}),
		field.NotSupported(field.NewPath("int32Field"), "4", []string{"1", "2", "3"}),
		field.NotSupported(field.NewPath("int64Field"), "4", []string{"1", "2", "3"}),
		field.NotSupported(field.NewPath("uintField"), "4", []string{"1", "2", "3"}),
		field.NotSupported(field.NewPath("uintPtrField"), "4", []string{"1", "2", "3"}),
		field.NotSupported(field.NewPath("uint16Field"), "4", []string{"1", "2", "3"}),
		field.NotSupported(field.NewPath("uint32Field"), "4", []string{"1", "2", "3"}),
		field.NotSupported(field.NewPath("uint64Field"), "4", []string{"1", "2", "3"}),
		field.NotSupported(field.NewPath("intWithNegativeField"), "-4", []string{"-1", "-2", "-3"}),
		field.NotSupported(field.NewPath("typedefField"), "4", []string{"1", "2", "3"}),
		field.NotSupported(field.NewPath("typedefPtrField"), "4", []string{"1", "2", "3"}),
		field.NotSupported(field.NewPath("validatedTypedefIntField"), "4", []string{"1", "2", "3"}),
	)

	// Test nil pointers.
	st.Value(&Struct{
		IntField:                 3,
		IntPtrField:              nil,
		Int16Field:               3,
		Int32Field:               3,
		Int64Field:               3,
		UintField:                3,
		UintPtrField:             nil,
		Uint16Field:              3,
		Uint32Field:              3,
		Uint64Field:              3,
		IntWithNegativeField:     -3,
		TypedefField:             3,
		TypedefPtrField:          nil,
		ValidatedTypedefIntField: 3,
	}).ExpectValid()

	// Test ratcheting
	st.Value(invalid).OldValue(invalid).ExpectValid()
}
