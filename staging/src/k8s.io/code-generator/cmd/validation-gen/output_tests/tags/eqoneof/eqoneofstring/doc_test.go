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

package eqoneofstring

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"
)

func Test(t *testing.T) {
	st := localSchemeBuilder.Test(t)
	// Test valid values.
	st.Value(&Struct{
		StringField:           "valid2",
		StringPtrField:        ptr.To("ptr-valid1"),
		TypedefStringField:    "typedef-valid2",
		TypedefStringPtrField: ptr.To(StringType("ptr-typedef-valid1")),
		StringWithSpacesField: "space valid1",
		ValidatedTypedefField: "validated-valid2",
	}).ExpectValid()

	// Test invalid values.
	invalid := &Struct{
		StringField:           "invalid-string",
		StringPtrField:        ptr.To("ptr-invalid"),
		TypedefStringField:    "typedef-invalid",
		TypedefStringPtrField: ptr.To(StringType("ptr-typedef-invalid")),
		StringWithSpacesField: "space invalid",
		ValidatedTypedefField: "validated-invalid",
	}
	st.Value(invalid).ExpectInvalid(
		field.NotSupported(field.NewPath("stringField"), "invalid-string", []string{"valid1", "valid2", "valid3"}),
		field.NotSupported(field.NewPath("stringPtrField"), "ptr-invalid", []string{"ptr-valid1", "ptr-valid2"}),
		field.NotSupported(field.NewPath("typedefStringField"), "typedef-invalid", []string{"typedef-valid1", "typedef-valid2"}),
		field.NotSupported(field.NewPath("typedefStringPtrField"), "ptr-typedef-invalid", []string{"ptr-typedef-valid1", "ptr-typedef-valid2"}),
		field.NotSupported(field.NewPath("stringWithSpacesField"), "space invalid", []string{"space valid1", "space valid2"}),
		field.NotSupported(field.NewPath("validatedTypedefField"), "validated-invalid", []string{"validated-valid1", "validated-valid2"}),
	)

	// Test nil pointers.
	st.Value(&Struct{
		StringField:           "valid1",
		StringPtrField:        nil,
		TypedefStringField:    "typedef-valid1",
		TypedefStringPtrField: nil,
		StringWithSpacesField: "space valid2",
		ValidatedTypedefField: "validated-valid1",
	}).ExpectValid()

	// Test ratcheting.
	st.Value(invalid).OldValue(invalid).ExpectValid()
}
