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

package multiple_keys

import (
	"testing"
)

func Test(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	st.Value(&Struct{
		Items: []Item{
			{Key1: "a", Key2: "b", Data: "match"},   
			{Key1: "a", Key2: "c", Data: "no match"},  
			{Key1: "b", Key2: "b", Data: "no match"},  
			{Key1: "c", Key2: "d", Data: "different"}, 
		},
	}).ExpectValidateFalseByPath(map[string][]string{
		`items[0]`: {"listMapItem Items[key1=a,key2=b]"},
	})

	st.Value(&Struct{
		Items: []Item{
			{Key1: "x", Key2: "y", Data: "d1"},
			{Key1: "a", Key2: "y", Data: "d2"}, 
			{Key1: "x", Key2: "b", Data: "d3"}, 
		},
	}).ExpectValid()
}
