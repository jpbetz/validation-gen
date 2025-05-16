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

package listcomplex

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"
)

func TestListComplexSubfield(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	st.Value(&Struct{
		Items: []Item{
			{Type: "foo", Status: "True", NestedStruct: nil},
		},
	}).ExpectInvalid(
		field.Required(field.NewPath("items").Index(0).Child("nestedStruct"), ""),
	)

	st.Value(&Struct{
		Items: []Item{
			{Type: "foo", Status: "True", NestedStruct: &NestedStruct{Value: ptr.To("present")}},
		},
	}).ExpectValid()

	st.Value(&Struct{
		Items: []Item{
			{Type: "foo", Status: "False", NestedStruct: nil},
		},
	}).ExpectValid()

}
