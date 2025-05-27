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

package empty

import (
	"testing"
)

func Test(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	st.Value(&Struct{
		ListField: []Item{},
	}).ExpectValid()

	st.Value(&Struct{
		ListField: nil,
	}).ExpectValid()

	oldStruct := &Struct{ListField: nil}
	newStruct := &Struct{ListField: []Item{}}
	st.Value(newStruct).OldValue(oldStruct).ExpectValid()
	st.Value(oldStruct).OldValue(newStruct).ExpectValid()

	emptyList := &Struct{ListField: []Item{}}
	populatedList := &Struct{
		ListField: []Item{
			{Key: "fail", Data: "data"},
			{Key: "fixed", Data: "immutable"},
			{Key: "normal", Data: "ok"},
		},
	}
	st.Value(populatedList).OldValue(emptyList).ExpectValidateFalseByPath(map[string][]string{
		`listField[0]`: {"listMapItem ListField[key=fail]"},
	})
}
