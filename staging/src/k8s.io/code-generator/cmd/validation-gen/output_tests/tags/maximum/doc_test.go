/*
Copyright 2024 The Kubernetes Authors.

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

package maximum

import (
	"testing"

	"k8s.io/apimachinery/pkg/api/validate/content"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"
)

func Test(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	st.Value(&Struct{
		IntField:        2,
		IntPtrField:     ptr.To(2),
		Int16Field:      2,
		Int32Field:      2,
		Int64Field:      2,
		UintField:       2,
		Uint16Field:     2,
		Uint32Field:     2,
		Uint64Field:     2,
		UintPtrField:    ptr.To(uint(2)),
		TypedefField:    IntType(2),
		TypedefPtrField: ptr.To(IntType(2)),
		IntLimit:        1,
	}).ExpectInvalid(
		field.Invalid(field.NewPath("intField"), 2, content.MaxFieldError("intLimit")),
		field.Invalid(field.NewPath("intPtrField"), 2, content.MaxError(1)),
		field.Invalid(field.NewPath("int16Field"), 2, content.MaxError(1)),
		field.Invalid(field.NewPath("int32Field"), 2, content.MaxError(1)),
		field.Invalid(field.NewPath("int64Field"), 2, content.MaxError(1)),
		field.Invalid(field.NewPath("uintField"), uint(2), content.MaxError(1)),
		field.Invalid(field.NewPath("uintPtrField"), uint(2), content.MaxError(1)),
		field.Invalid(field.NewPath("uint16Field"), uint(2), content.MaxError(1)),
		field.Invalid(field.NewPath("uint32Field"), uint(2), content.MaxError(1)),
		field.Invalid(field.NewPath("uint64Field"), uint(2), content.MaxError(1)),
		field.Invalid(field.NewPath("typedefField"), 2, content.MaxError(1)),
		field.Invalid(field.NewPath("typedefPtrField"), 2, content.MaxError(1)),
	)

	st.Value(&Struct{
		IntField:        1,
		IntPtrField:     ptr.To(1),
		Int16Field:      1,
		Int32Field:      1,
		Int64Field:      1,
		UintField:       1,
		Uint16Field:     1,
		Uint32Field:     1,
		Uint64Field:     1,
		UintPtrField:    ptr.To(uint(1)),
		TypedefField:    IntType(1),
		TypedefPtrField: ptr.To(IntType(1)),
		IntLimit:        2,
	}).ExpectValid()
}
