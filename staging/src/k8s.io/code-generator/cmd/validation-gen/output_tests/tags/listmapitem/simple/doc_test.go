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

package simple

import (
	"testing"
)

func Test(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	singleKeyItemsMsg := "listMapItem SingleKeyItems[key=Target]"
	multipleKeyItemsMsg := "listMapItem MultipleKeyItems[key=Target,key2=Target2]"

	st.Value(&Struct{SingleKeyItems: []Item{}}).ExpectValid()
	st.Value(&Struct{SingleKeyItems: nil}).ExpectValid()

	st.Value(&Struct{
		SingleKeyItems: []Item{{Key: "Target"}},
	}).ExpectValidateFalseByPath(map[string][]string{
		`singleKeyItems[0]`: {singleKeyItemsMsg},
	})

	st.Value(&Struct{
		SingleKeyItems: []Item{{Key: "NotTarget"}},
	}).ExpectValid()

	st.Value(&Struct{
		MultipleKeyItems: []Item{{Key: "Target", Key2: "Target2"}},
	}).ExpectValidateFalseByPath(map[string][]string{
		`multipleKeyItems[0]`: {multipleKeyItemsMsg},
	})

	st.Value(&Struct{
		MultipleKeyItems: []Item{{Key: "Target", Key2: "NotTarget"}},
	}).ExpectValid()

	st.Value(&Struct{
		MultipleKeyItems: []Item{{Key: "NotTarget", Key2: "Target2"}},
	}).ExpectValid()

	st.Value(&Struct{MultipleKeyItems: []Item{}}).ExpectValid()
	st.Value(&Struct{MultipleKeyItems: nil}).ExpectValid()

	st.Value(&Struct{
		SingleKeyItems:   []Item{{Key: "Target"}},
		MultipleKeyItems: []Item{{Key: "Target", Key2: "Target2"}},
	}).ExpectValidateFalseByPath(map[string][]string{
		`singleKeyItems[0]`:   {singleKeyItemsMsg},
		`multipleKeyItems[0]`: {multipleKeyItemsMsg},
	})

	st.Value(&Struct{
		SingleKeyItems:   []Item{{Key: "NotTarget"}},
		MultipleKeyItems: []Item{{Key: "NotTarget", Key2: "NotTarget"}},
	}).ExpectValid()
}
