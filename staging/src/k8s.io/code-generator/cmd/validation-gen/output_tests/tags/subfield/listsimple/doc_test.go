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

package listsimple

import (
	"testing"

	"k8s.io/utils/ptr"
)

func TestSubfieldListSimple(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	st.Value(&Struct{
		Conditions: []MyCondition{
			{Type: "Approved", Status: "False"},
			{Type: "OtherType", Status: "False"},
		},
	}).ExpectValidateFalseByPath(map[string][]string{
		`conditions[0]`: {"subfield Conditions[type=Approved]"},
	})

	st.Value(&Struct{
		Conditions: []MyCondition{
			{Type: "Approved", Status: "False"},
		},
	}).ExpectValidateFalseByPath(map[string][]string{
		`conditions[0]`: {"subfield Conditions[type=Approved]"},
	})

	st.Value(&Struct{
		Conditions: []MyCondition{
			{Type: "UnknownType1", Status: "Anything"},
		},
	}).ExpectValid()

	st.Value(&Struct{
		Conditions: []MyCondition{
			{Type: "Approved", Status: "True"},
		},
	}).ExpectValidateFalseByPath(map[string][]string{
		`conditions[0]`: {
			"subfield Conditions[type=Approved]",
			"subfield Conditions[status=True,type=Approved]",
		},
	})

	st.Value(&Struct{
		Conditions: []MyCondition{
			{Type: "Approved", Status: "OnlyOneEntryNotTwo"},
		},
	}).ExpectValidateFalseByPath(map[string][]string{
		`conditions[0]`: {"subfield Conditions[type=Approved]"},
	})

	st.Value(&Struct{
		Conditions: []MyCondition{
			{Type: "Approved", StringPtr: ptr.To("Target")},
		},
	}).ExpectValidateFalseByPath(map[string][]string{
		`conditions[0]`: {
			"subfield Conditions[stringPtr=Target,type=Approved]",
			"subfield Conditions[type=Approved]"},
	})

	st.Value(&Struct{
		Conditions: []MyCondition{
			{Type: "UnknownType", StringPtr: ptr.To("NotTarget")},
		},
	}).ExpectValid()

	st.Value(&Struct{
		Conditions: []MyCondition{
			{Type: "UnknownType", StringPtr: nil},
		},
	}).ExpectValid()
}
