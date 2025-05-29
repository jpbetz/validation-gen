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

package transitions

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

func Test(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	oldStruct := &Struct{
		Items: []Item{
			{ID: "existing", Data: "data"},
		},
	}
	newStruct := &Struct{
		Items: []Item{
			{ID: "existing", Data: "data"},
			{ID: "temp", Data: "temporary"},
			{ID: "new", Data: "added"},
		},
	}
	st.Value(newStruct).OldValue(oldStruct).ExpectValidateFalseByPath(map[string][]string{
		`items[1]`: {"listMapItem Items[id=temp]"},
	})

	deletedStruct := &Struct{
		Items: []Item{
			{ID: "existing", Data: "data"},
		},
	}
	st.Value(deletedStruct).OldValue(newStruct).ExpectValidateFalseByPath(map[string][]string{
		`items`: {"listMapItem Items[id=temp]"},
	})

	beforeReorder := &Struct{
		Items: []Item{
			{ID: "low", Data: "d1"},
			{ID: "medium", Data: "d2"},
			{ID: "high", Data: "d3"},
		},
	}
	afterReorder := &Struct{
		Items: []Item{
			{ID: "high", Data: "d3"},
			{ID: "medium", Data: "d2"},
			{ID: "low", Data: "d1"},
		},
	}
	st.Value(afterReorder).OldValue(beforeReorder).ExpectValidateFalseByPath(map[string][]string{
		`items[0]`: {"listMapItem Items[id=high]"},
	})

	oldMulti := &Struct{
		MultiKey: []MultiKeyItem{
			{Key1: "a", Key2: "1", StringField: "s1"},
			{Key1: "b", Key2: "1", StringField: "s2"},
		},
	}
	newMulti := &Struct{
		MultiKey: []MultiKeyItem{
			{Key1: "a", Key2: "1", StringField: "changed"},
			{Key1: "b", Key2: "1", StringField: "changed"},
		},
	}
	st.Value(newMulti).OldValue(oldMulti).ExpectInvalid(
		field.Forbidden(field.NewPath("multiKey").Index(0), "field is immutable"),
		field.Forbidden(field.NewPath("multiKey").Index(1).Child("stringField"), "field is immutable"),
	)
}
