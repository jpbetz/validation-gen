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

	"k8s.io/apimachinery/pkg/util/validation/field"
)

func Test(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	st.Value(&Struct{
		SingleKey: []Item{
			{Key: "other", Data: "d1"},
			{Key: "target", Data: "d2"},
			{Key: "fixed", Data: "d3"},
		},
	}).ExpectValidateFalseByPath(map[string][]string{
		`singleKey[1]`: {"listMapItem SingleKey[key=target]"},
	})

	oldStruct := &Struct{
		SingleKey: []Item{
			{Key: "fixed", Data: "original"},
		},
	}
	newStruct := &Struct{
		SingleKey: []Item{
			{Key: "fixed", Data: "changed"},
		},
	}
	st.Value(newStruct).OldValue(oldStruct).ExpectInvalid(
		field.Forbidden(field.NewPath("singleKey").Index(0), "field is immutable"),
	)

	st.Value(&Struct{
		MultiKey: []MultiItem{
			{Key1: "a", Key2: "b", Data: "match"},
			{Key1: "a", Key2: "c", Data: "no match"},
			{Key1: "b", Key2: "b", Data: "no match"},
		},
	}).ExpectValidateFalseByPath(map[string][]string{
		`multiKey[0]`: {"listMapItem MultiKey[key1=a,key2=b]"},
	})

	st.Value(&Struct{
		WithSubfield: []SubfieldItem{
			{Key: "other", StringField: "any"},
			{Key: "target", StringField: "fails"},
		},
	}).ExpectValidateFalseByPath(map[string][]string{
		`withSubfield[1].stringField`: {"listMapItem WithSubfield[key=target].stringField"},
	})

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

	st.Value(&Struct{
		OpaqueList: []OpaqueItem{
			{Key: "opaque", OpaqueData: "should not validate internals"},
			{Key: "validated", OpaqueData: "should validate"},
			{Key: "normal", OpaqueData: "no validation"},
		},
	}).ExpectValidateFalseByPath(map[string][]string{
		`opaqueList[1]`: {"listMapItem OpaqueList[key=validated]"},
	})
}
