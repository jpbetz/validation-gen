/*
Copyright 2025 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*/

package sliceofstruct

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

func Test(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	// Zero-value struct should fail for fields with minItems > 0
	st.Value(&Struct{}).ExpectMatches(field.ErrorMatcher{}.ByType().ByField(), field.ErrorList{
		field.TooFew(field.NewPath("min1Field"), 0, 1),
		field.TooFew(field.NewPath("min5Field"), 0, 5),
		field.TooFew(field.NewPath("min1TypedefField"), 0, 1),
		field.TooFew(field.NewPath("min5TypedefField"), 0, 5),
	})

	// Struct with exactly enough items should be valid
	st.Value(&Struct{
		Min1Field:        make([]OtherStruct, 1),
		Min5Field:        make([]OtherStruct, 5),
		Min1TypedefField: make([]OtherTypedefStruct, 1),
		Min5TypedefField: make([]OtherTypedefStruct, 5),
	}).ExpectValid()

	// Struct with more than minItems is also valid
	st.Value(&Struct{
		Min1Field:        make([]OtherStruct, 2),
		Min5Field:        make([]OtherStruct, 10),
		Min1TypedefField: make([]OtherTypedefStruct, 3),
		Min5TypedefField: make([]OtherTypedefStruct, 6),
	}).ExpectValid()

	// Struct with too few items triggers errors
	testVal := &Struct{
		Min1Field:        make([]OtherStruct, 0),
		Min5Field:        make([]OtherStruct, 3),
		Min1TypedefField: make([]OtherTypedefStruct, 0),
		Min5TypedefField: make([]OtherTypedefStruct, 4),
	}
	st.Value(testVal).ExpectMatches(field.ErrorMatcher{}.ByType().ByField(), field.ErrorList{
		field.TooFew(field.NewPath("min1Field"), 0, 1),
		field.TooFew(field.NewPath("min5Field"), 3, 5),
		field.TooFew(field.NewPath("min1TypedefField"), 0, 1),
		field.TooFew(field.NewPath("min5TypedefField"), 4, 5),
	})

	// Test validation ratcheting (if old value was already invalid, new value is allowed)
	st.Value(&Struct{
		Min1Field:        make([]OtherStruct, 0),
		Min5Field:        make([]OtherStruct, 3),
		Min1TypedefField: make([]OtherTypedefStruct, 0),
		Min5TypedefField: make([]OtherTypedefStruct, 4),
	}).OldValue(&Struct{
		Min1Field:        make([]OtherStruct, 0),
		Min5Field:        make([]OtherStruct, 3),
		Min1TypedefField: make([]OtherTypedefStruct, 0),
		Min5TypedefField: make([]OtherTypedefStruct, 4),
	}).ExpectValid()
}