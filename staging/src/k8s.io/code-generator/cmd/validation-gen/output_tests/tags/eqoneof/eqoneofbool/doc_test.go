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

package eqoneofbool

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"
)

func Test(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	// Test valid values.
	st.Value(&Struct{
		BoolOnlyTrueField:         true,
		BoolPtrOnlyFalseField:     ptr.To(false),
		BoolBothAllowedField:      false,
		ValidatedTypedefBoolField: false,
	}).ExpectValid()

	// Test invalid values.
	invalid := &Struct{
		BoolOnlyTrueField:         false,
		BoolPtrOnlyFalseField:     ptr.To(true),
		BoolBothAllowedField:      true,
		ValidatedTypedefBoolField: true,
	}
	st.Value(invalid).ExpectInvalid(
		field.NotSupported(field.NewPath("boolOnlyTrueField"), "false", []string{"true"}),
		field.NotSupported(field.NewPath("boolPtrOnlyFalseField"), "true", []string{"false"}),
		field.NotSupported(field.NewPath("validatedTypedefBoolField"), "true", []string{"false"}),
	)

	// Test nil pointer.
	st.Value(&Struct{
		BoolOnlyTrueField:         true,
		BoolPtrOnlyFalseField:     nil,
		BoolBothAllowedField:      true,
		ValidatedTypedefBoolField: false,
	}).ExpectValid()

	// Test ratcheting.
	st.Value(invalid).OldValue(invalid).ExpectValid()
}
