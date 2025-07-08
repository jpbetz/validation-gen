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

package bool_key

import (
	"testing"
)

func Test(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	st.Value(&Struct{
		ListField: []Item{
			{BoolKey: false, StringField: "s1"},
		},
	}).ExpectValid()

	st.Value(&Struct{
		ListField: []Item{
			{BoolKey: true, StringField: "s1"},
			{BoolKey: false, StringField: "s2"},
		},
	}).ExpectValidateFalseByPath(map[string][]string{
		`listField[0]`: {"item ListField[boolKey=true]"},
	})

	st.Value(&Struct{
		MultiKey: []MultiKeyItem{
			{Key1: true, Key2: 5, Key3: "x", StringField: "s1"},
			{Key1: false, Key2: 10, Key3: "y", StringField: "s2"},
		},
	}).ExpectValidateFalseByPath(map[string][]string{
		`multiKey[0]`: {"item MultiKey[key1=true,key2=5,key3=x]"},
	})

	st.Value(&StructWithTypedef{
		TypedefField: TypedefList{
			{BoolField: true, StringField: "s1"},
		},
	}).ExpectValidateFalseByPath(map[string][]string{
		`typedefField[0]`: {"item TypedefList[boolField=true]"},
	})
}
