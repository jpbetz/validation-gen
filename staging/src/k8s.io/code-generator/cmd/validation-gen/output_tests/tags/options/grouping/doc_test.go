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

package grouping

import (
	"testing"
)

func Test(t *testing.T) {

	st := localSchemeBuilder.Test(t)

	st.Value(&Struct{
		// All zero values
	}).ExpectValidateFalseByPath(map[string][]string{
		// All ifDisabled validations should trigger
		"primitiveField": {
			"field Struct.PrimitiveField disabled validation 1",
			"field Struct.PrimitiveField disabled validation 2",
		},
	})

	st.Value(&Struct{
		XEnabledListUniqueField: []Item{
			{Key: "a"},
			{Key: "a"},
		},
	}).Opts([]string{"FeatureX", "FeatureA", "FeatureB"}).ExpectValidateFalseByPath(map[string][]string{
		// All ifEnabled validations should trigger
		"primitiveField": {
			"field Struct.PrimitiveField enabled validation 1",
			"field Struct.PrimitiveField enabled validation 2",
			"field Struct.PrimitiveField feature B",
		},
	})

	st.Value(&Struct{
		// All zero values
	}).Opts([]string{"FeatureA"}).ExpectValidateFalseByPath(map[string][]string{
		// All ifEnabled(FeatureA) validations should trigger
		// All ifDisabled(<non-FeatureA>) validations should trigger
		"primitiveField": {
			"field Struct.PrimitiveField enabled validation 1",
			"field Struct.PrimitiveField enabled validation 2",
		},
	})

	st.Value(&Struct{
		// All zero values
	}).Opts([]string{"FeatureB"}).ExpectValidateFalseByPath(map[string][]string{
		// All ifEnabled(FeatureB) validations should trigger
		// All ifDisabled(<non-FeatureB>) validations should trigger
		"primitiveField": {
			"field Struct.PrimitiveField disabled validation 1",
			"field Struct.PrimitiveField disabled validation 2",
			"field Struct.PrimitiveField feature B",
		},
	})
}
