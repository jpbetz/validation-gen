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

package immutable

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"
)

func Test(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	structA := Struct{
		StringField:                 "aaa",
		StringPtrField:              ptr.To("aaa"),
		StructField:                 ComparableStruct{"bbb", ptr.To("BBB")},
		StructPtrField:              ptr.To(ComparableStruct{"bbb", ptr.To("BBB")}),
		NonComparableStructField:    NonComparableStruct{[]string{"ccc"}},
		NonComparableStructPtrField: ptr.To(NonComparableStruct{[]string{"ccc"}}),
		SliceField:                  []string{"ddd"},
		MapField:                    map[string]string{"eee": "eee"},
		ImmutableField:              "fff",
		ImmutablePtrField:           ptr.To(ImmutableType("fff")),
		IntPtrField:                 ptr.To(123),
		BoolPtrField:                ptr.To(true),
		RequiredImmutableField:      "required",
		OptionalImmutableField:      ptr.To("optional"),
	}

	st.Value(&structA).OldValue(&structA).ExpectValid()

	structA2 := structA // dup of A but with different pointer values
	structA2.StringPtrField = ptr.To(*structA2.StringPtrField)
	structA2.StructField.StringPtrField = ptr.To("BBB")
	structA2.StructPtrField = ptr.To(*structA2.StructPtrField)
	structA2.StructPtrField.StringPtrField = ptr.To("BBB")
	structA2.NonComparableStructPtrField = ptr.To(*structA2.NonComparableStructPtrField)
	structA2.ImmutablePtrField = ptr.To(*structA2.ImmutablePtrField)
	structA2.IntPtrField = ptr.To(123)
	structA2.BoolPtrField = ptr.To(true)
	structA2.OptionalImmutableField = ptr.To("optional")

	st.Value(&structA).OldValue(&structA2).ExpectValid()
	st.Value(&structA2).OldValue(&structA).ExpectValid()

	structUnset := Struct{
		RequiredImmutableField: "required",
		// All other fields unset or zero values.
	}

	st.Value(&structA).OldValue(&structUnset).ExpectValid()

	structWithNilPointers := Struct{
		RequiredImmutableField: "required",
		// Pointers are nil (unset).
	}

	structPtrToZeroValues := Struct{
		StringPtrField:         ptr.To(""),
		IntPtrField:            ptr.To(0),
		BoolPtrField:           ptr.To(false),
		RequiredImmutableField: "required",
	}

	st.Value(&structPtrToZeroValues).OldValue(&structWithNilPointers).ExpectValid()

	structAModified := Struct{
		StringField:                 "uuu",
		StringPtrField:              ptr.To("uuu"),
		StructField:                 ComparableStruct{"vvv", ptr.To("VVV")},
		StructPtrField:              ptr.To(ComparableStruct{"vvv", ptr.To("VVV")}),
		NonComparableStructField:    NonComparableStruct{[]string{"www"}},
		NonComparableStructPtrField: ptr.To(NonComparableStruct{[]string{"www"}}),
		SliceField:                  []string{"xxx"},
		MapField:                    map[string]string{"yyy": "yyy"},
		ImmutableField:              "zzz",
		ImmutablePtrField:           ptr.To(ImmutableType("zzz")),
		IntPtrField:                 ptr.To(999),
		BoolPtrField:                ptr.To(false),
		RequiredImmutableField:      "different",
		OptionalImmutableField:      ptr.To("changed"),
	}

	st.Value(&structAModified).OldValue(&structA).ExpectInvalid(
		field.Forbidden(field.NewPath("stringField"), "field is immutable"),
		field.Forbidden(field.NewPath("stringPtrField"), "field is immutable"),
		field.Forbidden(field.NewPath("structField"), "field is immutable"),
		field.Forbidden(field.NewPath("structPtrField"), "field is immutable"),
		field.Forbidden(field.NewPath("noncomparableStructField"), "field is immutable"),
		field.Forbidden(field.NewPath("noncomparableStructPtrField"), "field is immutable"),
		field.Forbidden(field.NewPath("sliceField"), "field is immutable"),
		field.Forbidden(field.NewPath("mapField"), "field is immutable"),
		field.Forbidden(field.NewPath("immutableField"), "field is immutable"),
		field.Forbidden(field.NewPath("immutablePtrField"), "field is immutable"),
		field.Forbidden(field.NewPath("intPtrField"), "field is immutable"),
		field.Forbidden(field.NewPath("boolPtrField"), "field is immutable"),
		field.Forbidden(field.NewPath("requiredImmutableField"), "field is immutable"),
		field.Forbidden(field.NewPath("optionalImmutableField"), "field is immutable"),
	)

	st.Value(&structUnset).OldValue(&structA).ExpectInvalid(
		field.Forbidden(field.NewPath("stringField"), "field is immutable"),
		field.Forbidden(field.NewPath("stringPtrField"), "field is immutable"),
		field.Forbidden(field.NewPath("structField"), "field is immutable"),
		field.Forbidden(field.NewPath("structPtrField"), "field is immutable"),
		field.Forbidden(field.NewPath("noncomparableStructField"), "field is immutable"),
		field.Forbidden(field.NewPath("noncomparableStructPtrField"), "field is immutable"),
		field.Forbidden(field.NewPath("sliceField"), "field is immutable"),
		field.Forbidden(field.NewPath("mapField"), "field is immutable"),
		field.Forbidden(field.NewPath("immutableField"), "field is immutable"),
		field.Forbidden(field.NewPath("immutablePtrField"), "field is immutable"),
		field.Forbidden(field.NewPath("intPtrField"), "field is immutable"),
		field.Forbidden(field.NewPath("boolPtrField"), "field is immutable"),
		// optionalImmutableField is not here because OptionalPointer short-circuits when nil.
	)

	structEmptyCollections := Struct{
		SliceField:             []string{},
		MapField:               map[string]string{},
		RequiredImmutableField: "required",
	}

	structNilCollections := Struct{
		// SliceField and MapField are nil.
		RequiredImmutableField: "required",
	}

	st.Value(&structEmptyCollections).OldValue(&structNilCollections).ExpectValid()
	st.Value(&structNilCollections).OldValue(&structEmptyCollections).ExpectValid()

	structWithDefaults := Struct{
		StringWithDefault:      "defaultValue",
		IntPtrWithDefault:      ptr.To(int32(42)),
		IntWithZeroDefault:     0,
		StringWithZeroDefault:  "",
		RequiredImmutableField: "required",
	}

	structModifiedDefaults := Struct{
		StringWithDefault:      "userValue",
		IntPtrWithDefault:      ptr.To(int32(100)),
		IntWithZeroDefault:     5,
		StringWithZeroDefault:  "notEmpty",
		RequiredImmutableField: "required",
	}

	st.Value(&structModifiedDefaults).OldValue(&structWithDefaults).ExpectInvalid(
		field.Forbidden(field.NewPath("stringWithDefault"), "field is immutable"),
		field.Forbidden(field.NewPath("intPtrWithDefault"), "field is immutable"),
		// intWithZeroDefault and stringWithZeroDefault are not here because
		// they have zero defaults which are considered unset (allowing for unset -> set transition)
	)

	structPtrToNonZeroValues := Struct{
		StringPtrField:         ptr.To("value"),
		IntPtrField:            ptr.To(123),
		BoolPtrField:           ptr.To(true),
		RequiredImmutableField: "required",
	}

	st.Value(&structPtrToNonZeroValues).OldValue(&structPtrToZeroValues).ExpectInvalid(
		field.Forbidden(field.NewPath("stringPtrField"), "field is immutable"),
		field.Forbidden(field.NewPath("intPtrField"), "field is immutable"),
		field.Forbidden(field.NewPath("boolPtrField"), "field is immutable"),
	)

	structZeroValues := Struct{
		StringField:            "",
		IntWithZeroDefault:     0,
		StringWithZeroDefault:  "",
		RequiredImmutableField: "required",
	}

	structNonZeroValues := Struct{
		StringField:            "value",
		IntWithZeroDefault:     5,
		StringWithZeroDefault:  "text",
		RequiredImmutableField: "required",
	}

	// Fields with zero value -> non-zero value is unset -> set
	st.Value(&structNonZeroValues).OldValue(&structZeroValues).ExpectValid()
}
