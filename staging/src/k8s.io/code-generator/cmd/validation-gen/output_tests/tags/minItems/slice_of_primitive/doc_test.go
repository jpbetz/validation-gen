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

package sliceofprimitive

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

func Test(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	st.Value(&Struct{
		Min1Field:        make([]int, 1),
		Min5Field:        make([]int, 5),
		Min1TypedefField: make([]IntType, 1),
		Min5TypedefField: make([]IntType, 5),
	}).ExpectValid()

	st.Value(&Struct{
		Min1Field:        make([]int, 2),
		Min5Field:        make([]int, 10),
		Min1TypedefField: make([]IntType, 3),
		Min5TypedefField: make([]IntType, 6),
	}).ExpectValid()

	testVal := &Struct{
		Min1Field:        make([]int, 0),
		Min5Field:        make([]int, 3),
		Min1TypedefField: make([]IntType, 0),
		Min5TypedefField: make([]IntType, 4),
	}
	st.Value(testVal).ExpectMatches(field.ErrorMatcher{}.ByType().ByField(), field.ErrorList{
		field.TooFew(field.NewPath("min1Field"), 0, 1),
		field.TooFew(field.NewPath("min5Field"), 3, 5),
		field.TooFew(field.NewPath("min1TypedefField"), 0, 1),
		field.TooFew(field.NewPath("min5TypedefField"), 4, 5),
	})

	// Test validation ratcheting
	st.Value(&Struct{
		Min1Field:        make([]int, 0),
		Min5Field:        make([]int, 3),
		Min1TypedefField: make([]IntType, 0),
		Min5TypedefField: make([]IntType, 4),
	}).OldValue(&Struct{
		Min1Field:        make([]int, 0),
		Min5Field:        make([]int, 3),
		Min1TypedefField: make([]IntType, 0),
		Min5TypedefField: make([]IntType, 4),
	}).ExpectValid()
}