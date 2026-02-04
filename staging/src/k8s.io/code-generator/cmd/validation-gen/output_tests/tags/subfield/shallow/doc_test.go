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

package shallow

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"
)

func Test(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	st.Value(&Struct{
		StructField: OtherStruct{
			StringField:  "",
			PointerField: ptr.To(""),
			StructField:  SmallStruct{},
			SliceField:   []string{},
			MapField:     map[string]string{},
		},
		StructPtrField: &OtherStruct{
			StringField:  "",
			PointerField: ptr.To(""),
			StructField:  SmallStruct{},
			SliceField:   []string{},
			MapField:     map[string]string{},
		},
	}).ExpectValidateFalseByPath(map[string][]string{
		"structField.stringField":     {"subfield Struct.StructField.StringField 1", "subfield Struct.StructField.StringField 2"},
		"structField.pointerField":    {"subfield Struct.StructField.PointerField"},
		"structField.structField":     {"subfield Struct.StructField.StructField"},
		"structField.sliceField":      {"subfield Struct.StructField.SliceField"},
		"structField.mapField":        {"subfield Struct.StructField.MapField"},
		"structPtrField.stringField":  {"subfield Struct.StructPtrField.StringField 1", "subfield Struct.StructPtrField.StringField 2"},
		"structPtrField.pointerField": {"subfield Struct.StructPtrField.PointerField"},
		"structPtrField.structField":  {"subfield Struct.StructPtrField.StructField"},
		"structPtrField.sliceField":   {"subfield Struct.StructPtrField.SliceField"},
		"structPtrField.mapField":     {"subfield Struct.StructPtrField.MapField"},
	})
}

func TestListInsideSubfield(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	// Valid list elements
	st.Value(&ListInsideSubfield{
		Lists: ListStruct{
			ListTypeMap: []ListItem{{Name: "a", Val: "1"}, {Name: "b", Val: "2"}},
			ListTypeSet: []string{"a", "b"},
		},
	}).ExpectValid()

	// Invalid list elements (duplicates)
	st.Value(&ListInsideSubfield{
		Lists: ListStruct{
			ListTypeMap: []ListItem{{Name: "a", Val: "1"}, {Name: "a", Val: "2"}}, // Dupe by key
			ListTypeSet: []string{"dup", "dup"},                                   // Dupe by set value
		},
	}).ExpectMatches(field.ErrorMatcher{}.ByType().ByField(), field.ErrorList{
		field.Duplicate(field.NewPath("lists").Child("listTypeMap").Index(1), "a"),
		field.Duplicate(field.NewPath("lists").Child("listTypeSet").Index(1), "dup"),
	})
}

func TestUpdateInsideSubfield(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	// Valid update: unchanged
	st.Value(&UpdateInsideSubfield{
		Updatable: UpdateStruct{StringField: "old"},
	}).OldValue(&UpdateInsideSubfield{
		Updatable: UpdateStruct{StringField: "old"},
	}).ExpectValid()

	// Invalid update: modified inside subfield
	st.Value(&UpdateInsideSubfield{
		Updatable: UpdateStruct{StringField: "new"},
	}).OldValue(&UpdateInsideSubfield{
		Updatable: UpdateStruct{StringField: "old"},
	}).ExpectMatches(field.ErrorMatcher{}.ByType().ByField(), field.ErrorList{
		field.Invalid(field.NewPath("updatable").Child("stringField"), "new", "field is immutable"),
	})
}

func TestDuplicateAccumulatorStruct(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	// Valid composite key elements
	st.Value(&DuplicateAccumulatorStruct{
		ListTypeMap: []ListItem{{Name: "a", Val: "1"}, {Name: "a", Val: "2"}}, // Dupe by name, different val
	}).ExpectValid()

	// Invalid composite key elements (duplicates by both keys)
	st.Value(&DuplicateAccumulatorStruct{
		ListTypeMap: []ListItem{{Name: "a", Val: "1"}, {Name: "a", Val: "1"}}, // Dupe by name and val
	}).ExpectMatches(field.ErrorMatcher{}.ByType().ByField(), field.ErrorList{
		field.Duplicate(field.NewPath("listTypeMap").Index(1), "a"),
	})
}

func TestAggregatedUpdateStruct(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	// Valid update: unchanged
	st.Value(&AggregatedUpdateStruct{
		StringField: ptr.To("old"),
	}).OldValue(&AggregatedUpdateStruct{
		StringField: ptr.To("old"),
	}).ExpectValid()

	// Invalid update 1: NoModify triggers
	st.Value(&AggregatedUpdateStruct{
		StringField: ptr.To("new"),
	}).OldValue(&AggregatedUpdateStruct{
		StringField: ptr.To("old"),
	}).ExpectMatches(field.ErrorMatcher{}.ByType().ByField(), field.ErrorList{
		field.Invalid(field.NewPath("stringField"), ptr.To("new"), "field is immutable"),
	})

	// Invalid update 2: NoUnset triggers
	st.Value(&AggregatedUpdateStruct{
		StringField: nil,
	}).OldValue(&AggregatedUpdateStruct{
		StringField: ptr.To("old"),
	}).ExpectMatches(field.ErrorMatcher{}.ByType().ByField(), field.ErrorList{
		field.Invalid(field.NewPath("stringField"), nil, "field cannot be unset"),
	})
}
