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

package int_key

import (
	"testing"
)

func Test(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	st.Value(&Struct{
		ListField: []Item{
			{IntKey: 5, StringField: "s1"},
			{IntKey: 15, StringField: "s2"},
		},
	}).ExpectValid()

	st.Value(&Struct{
		ListField: []Item{
			{IntKey: 5, StringField: "s1"},
			{IntKey: -10, StringField: "s2"},
			{IntKey: 15, StringField: "s3"},
		},
	}).ExpectValidateFalseByPath(map[string][]string{
		`listField[1]`: {"item ListField[intKey=-10]"},
	})

	st.Value(&Struct{
		MultiKey: []MultiKeyItem{
			{Key1: -1, Key2: "a", StringField: "s1"},
			{Key1: 2, Key2: "b", StringField: "s2"},
		},
	}).ExpectValidateFalseByPath(map[string][]string{
		`multiKey[0]`: {"item MultiKey[key1=-1,key2=a]"},
	})

	st.Value(&StructWithTypedef{
		TypedefField: TypedefList{
			{IntField: 50, StringField: "s1"},
			{IntField: -100, StringField: "s2"},
		},
	}).ExpectValidateFalseByPath(map[string][]string{
		`typedefField[1]`: {"item TypedefList[intField=-100]"},
	})
}
