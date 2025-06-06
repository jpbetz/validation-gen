/*
Copyright 2024 The Kubernetes Authors.

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

package required

import (
	"testing"

	"k8s.io/utils/ptr"
)

func Test(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	st.Value(&Struct{ /* All zero-values */ }).ExpectRegexpsByPath(map[string][]string{
		"stringField":           {"Required value"},
		"stringPtrField":        {"Required value"},
		"stringTypedefField":    {"Required value"},
		"stringTypedefPtrField": {"Required value"},
		"intField":              {"Required value"},
		"intPtrField":           {"Required value"},
		"intTypedefField":       {"Required value"},
		"intTypedefPtrField":    {"Required value"},
		"otherStructPtrField":   {"Required value"},
		"sliceField":            {"Required value"},
		"sliceTypedefField":     {"Required value"},
		"mapField":              {"Required value"},
		"mapTypedefField":       {"Required value"},
	})

	// Test validation ratcheting
	st.Value(&Struct{}).OldValue(&Struct{}).ExpectValid()

	st.Value(&Struct{
		StringPtrField:        ptr.To(""),             // satisfies required
		StringTypedefPtrField: ptr.To(StringType("")), // satisfies required
		IntPtrField:           ptr.To(0),              // satisfies required
		IntTypedefPtrField:    ptr.To(IntType(0)),     // satisfies required
		SliceField:            []string{},             // does not satisfy required
		SliceTypedefField:     []string{},             // does not satisfy required
		MapField:              map[string]string{},    // does not satisfy required
		MapTypedefField:       map[string]string{},    // does not satisfy required
	}).ExpectRegexpsByPath(map[string][]string{
		"stringField":           {"Required value"},
		"stringPtrField":        {"field Struct.StringPtrField"},
		"stringTypedefField":    {"Required value"},
		"stringTypedefPtrField": {"field Struct.StringTypedefPtrField", "type StringType"},
		"intField":              {"Required value"},
		"intPtrField":           {"field Struct.IntPtrField"},
		"intTypedefField":       {"Required value"},
		"intTypedefPtrField":    {"field Struct.IntTypedefPtrField", "type IntType"},
		"otherStructPtrField":   {"Required value"},
		"sliceField":            {"Required value"},
		"sliceTypedefField":     {"Required value"},
		"mapField":              {"Required value"},
		"mapTypedefField":       {"Required value"},
	})
	// Test validation ratcheting
	st.Value(&Struct{
		StringPtrField:        ptr.To(""),             // satisfies required
		StringTypedefPtrField: ptr.To(StringType("")), // satisfies required
		IntPtrField:           ptr.To(0),              // satisfies required
		IntTypedefPtrField:    ptr.To(IntType(0)),     // satisfies required
		SliceField:            []string{},             // does not satisfy required
		SliceTypedefField:     []string{},             // does not satisfy required
		MapField:              map[string]string{},    // does not satisfy required
		MapTypedefField:       map[string]string{},    // does not satisfy required
	}).OldValue(&Struct{
		StringPtrField:        ptr.To(""),             // satisfies required
		StringTypedefPtrField: ptr.To(StringType("")), // satisfies required
		IntPtrField:           ptr.To(0),              // satisfies required
		IntTypedefPtrField:    ptr.To(IntType(0)),     // satisfies required
		// nil and empty slices are considered equivalent.
	}).ExpectValid()

	mkInvalid := func() *Struct {
		return &Struct{
			StringField:           "abc",
			StringPtrField:        ptr.To("xyz"),
			StringTypedefField:    StringType("abc"),
			StringTypedefPtrField: ptr.To(StringType("xyz")),
			IntField:              123,
			IntPtrField:           ptr.To(456),
			IntTypedefField:       IntType(123),
			IntTypedefPtrField:    ptr.To(IntType(456)),
			OtherStructPtrField:   &OtherStruct{},
			SliceField:            []string{"a", "b"},
			SliceTypedefField:     SliceType([]string{"a", "b"}),
			MapField:              map[string]string{"a": "b", "c": "d"},
			MapTypedefField:       MapType(map[string]string{"a": "b", "c": "d"}),
		}
	}

	st.Value(mkInvalid()).ExpectRegexpsByPath(map[string][]string{
		"stringField":           {"field Struct.StringField"},
		"stringPtrField":        {"field Struct.StringPtrField"},
		"stringTypedefField":    {"field Struct.StringTypedefField", "type StringType"},
		"stringTypedefPtrField": {"field Struct.StringTypedefPtrField", "type StringType"},
		"intField":              {"field Struct.IntField"},
		"intPtrField":           {"field Struct.IntPtrField"},
		"intTypedefField":       {"field Struct.IntTypedefField", "type IntType"},
		"intTypedefPtrField":    {"field Struct.IntTypedefPtrField", "type IntType"},
		"otherStructPtrField":   {"field Struct.OtherStructPtrField", "type OtherStruct"},
		"sliceField":            {"field Struct.SliceField"},
		"sliceTypedefField":     {"field Struct.SliceTypedefField", "type SliceType"},
		"mapField":              {"field Struct.MapField"},
		"mapTypedefField":       {"field Struct.MapTypedefField", "type MapType"},
	})
	// Test validation ratcheting
	st.Value(mkInvalid()).OldValue(mkInvalid()).ExpectValid()
}
