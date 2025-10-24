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

package typedeftoslice

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
		UnvalidatedField: make(UnvalidatedType, 0),
		Min1Field:        make(Min1Type, 1),
		Min5Field:        make(Min5Type, 5),
		Min1TypedefField: make(Min1TypedefType, 1),
		Min5TypedefField: make(Min5TypedefType, 5),
	}).ExpectValid()

	// Struct with more than minItems is also valid
	st.Value(&Struct{
		UnvalidatedField: make(UnvalidatedType, 2),
		Min1Field:        make(Min1Type, 2),
		Min5Field:        make(Min5Type, 10),
		Min1TypedefField: make(Min1TypedefType, 3),
		Min5TypedefField: make(Min5TypedefType, 6),
	}).ExpectValid()

	// Struct with too few items triggers errors
	testVal := &Struct{
		UnvalidatedField: make(UnvalidatedType, 0),
		Min1Field:        make(Min1Type, 0),
		Min5Field:        make(Min5Type, 3),
		Min1TypedefField: make(Min1TypedefType, 0),
		Min5TypedefField: make(Min5TypedefType, 4),
	}
	st.Value(testVal).ExpectMatches(field.ErrorMatcher{}.ByType().ByField(), field.ErrorList{
		field.TooFew(field.NewPath("min1Field"), 0, 1),
		field.TooFew(field.NewPath("min5Field"), 3, 5),
		field.TooFew(field.NewPath("min1TypedefField"), 0, 1),
		field.TooFew(field.NewPath("min5TypedefField"), 4, 5),
	})

	// Test validation ratcheting
    st.Value(&Struct{
        UnvalidatedField:  make(UnvalidatedType, 0),
        Min1Field:         make(Min1Type, 0),
        Min5Field:         make(Min5Type, 3),
        Min1TypedefField:  make(Min1TypedefType, 0),
        Min5TypedefField:  make(Min5TypedefType, 4),
    }).OldValue(&Struct{
        UnvalidatedField:  make(UnvalidatedType, 0),
        Min1Field:         make(Min1Type, 0),
        Min5Field:         make(Min5Type, 3),
        Min1TypedefField:  make(Min1TypedefType, 0),
        Min5TypedefField:  make(Min5TypedefType, 4),
    }).ExpectValid() 
}