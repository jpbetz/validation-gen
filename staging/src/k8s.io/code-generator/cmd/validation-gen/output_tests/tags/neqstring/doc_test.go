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

package neqstring

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"
)

func TestNEQString(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	validStruct := &Struct{
		StringField:              "allowed-string",
		StringPtrField:           ptr.To("allowed-pointer"),
		StringTypedefField:       "allowed-typedef",
		StringTypedefPtrField:    ptr.To(StringType("allowed-typedef-pointer")),
		ValidatedTypedefField:    "allowed-on-type",
		ValidatedTypedefPtrField: ptr.To(ValidatedStringType("allowed-on-type-ptr")),
	}
	invalidStruct := &Struct{
		StringField:              "disallowed-string",
		StringPtrField:           ptr.To("disallowed-pointer"),
		StringTypedefField:       "disallowed-typedef",
		StringTypedefPtrField:    ptr.To(StringType("disallowed-typedef-pointer")),
		ValidatedTypedefField:    "disallowed-on-type",
		ValidatedTypedefPtrField: ptr.To(ValidatedStringType("disallowed-on-type")),
	}
	invalidErrs := field.ErrorList{
		field.Invalid(field.NewPath("stringField"), "disallowed-string", `must not be equal to "disallowed-string"`),
		field.Invalid(field.NewPath("stringPtrField"), "disallowed-pointer", `must not be equal to "disallowed-pointer"`),
		field.Invalid(field.NewPath("stringTypedefField"), StringType("disallowed-typedef"), `must not be equal to "disallowed-typedef"`),
		field.Invalid(field.NewPath("stringTypedefPtrField"), StringType("disallowed-typedef-pointer"), `must not be equal to "disallowed-typedef-pointer"`),
		field.Invalid(field.NewPath("validatedTypedefField"), ValidatedStringType("disallowed-on-type"), `must not be equal to "disallowed-on-type"`),
		field.Invalid(field.NewPath("validatedTypedefPtrField"), ValidatedStringType("disallowed-on-type"), `must not be equal to "disallowed-on-type"`),
	}

	st.Value(validStruct).ExpectValid()

	// Test nil string ptr values
	st.Value(&Struct{
		StringPtrField:           nil,
		StringTypedefPtrField:    nil,
		ValidatedTypedefPtrField: nil,
	}).ExpectValid()

	st.Value(invalidStruct).ExpectInvalid(invalidErrs...)

	// Test validation ratcheting
	st.Value(invalidStruct).OldValue(invalidStruct).ExpectValid()
	st.Value(validStruct).OldValue(invalidStruct).ExpectValid()
	st.Value(invalidStruct).OldValue(validStruct).ExpectInvalid(invalidErrs...)
}
