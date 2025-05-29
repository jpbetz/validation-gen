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

package special_keys

import (
	"testing"
)

func Test(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	st.Value(&Struct{
		EmptyKey: []Item{
			{Key: "", Data: "empty"},
			{Key: "normal", Data: "normal"},
		},
	}).ExpectValidateFalseByPath(map[string][]string{
		`emptyKey[0]`: {"listMapItem EmptyKey[key=]"},
	})

	st.Value(&Struct{
		Special: []Item{
			{Key: `with"quotes`, Data: "d1"},
			{Key: "multi\nline", Data: "d2"},
			{Key: "unicode-🚀", Data: "d3"},
			{Key: "normal", Data: "d4"},
		},
	}).ExpectValidateFalseByPath(map[string][]string{
		`special[0]`: {`listMapItem Special[key=with"quotes]`},
		`special[1]`: {"listMapItem Special[key=multi\nline]"},
		`special[2]`: {"listMapItem Special[key=unicode-🚀]"},
	})

	// Test escaping doesn't affect non-special keys
	st.Value(&Struct{
		Special: []Item{
			{Key: "simple", Data: "d1"},
			{Key: "regular-dash", Data: "d2"},
		},
	}).ExpectValid()
}
