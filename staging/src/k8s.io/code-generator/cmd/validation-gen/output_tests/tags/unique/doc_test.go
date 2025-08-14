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

package unique

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestUnique(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	// Test empty struct (should be valid)
	st.Value(&Struct{}).ExpectValid()

	// Test valid cases with no duplicates
	st.Value(&Struct{
		SliceSetField: []string{"aaa", "bbb"},
		SliceMapField: []Item{
			{Key: "key1", Data: "one"},
			{Key: "key2", Data: "two"},
		},
		SliceSetFieldWithStruct: []Item{
			{Key: "key1", Data: "one"},
			{Key: "key2", Data: "two"},
		},
		SliceMapFieldWithMultipleKeys: []ItemWithMultipleKeys{
			{Key1: "a", Key2: "x", Data: "first"},
			{Key1: "a", Key2: "y", Data: "second"},
		},
		AtomicListUniqueSet: []Item{
			{Key: "key1", Data: "one"},
			{Key: "key2", Data: "two"},
		},
		AtomicListUniqueMap: []Item{
			{Key: "key1", Data: "one"},
			{Key: "key2", Data: "two"},
		},
		IntSlice:            []int{1, 2, 3},
		BoolSlice:           []bool{true, false},
		SliceWithZeroValues: []string{"a", "b", "c"},
		SliceWithEmptyKeys: []Item{
			{Key: "key1", Data: "one"},
			{Key: "key2", Data: "two"},
		},
	}).ExpectValid()

	// Test empty lists
	st.Value(&Struct{
		SliceSetField:                 []string{},
		SliceMapField:                 []Item{},
		SliceSetFieldWithStruct:       []Item{},
		SliceMapFieldWithMultipleKeys: []ItemWithMultipleKeys{},
		AtomicListUniqueSet:           []Item{},
		AtomicListUniqueMap:           []Item{},
		IntSlice:                      []int{},
		BoolSlice:                     []bool{},
		SliceWithZeroValues:           []string{},
		SliceWithEmptyKeys:            []Item{},
	}).ExpectValid()

	// Test single element lists
	st.Value(&Struct{
		SliceSetField:                 []string{"single"},
		SliceMapField:                 []Item{{Key: "single", Data: "one"}},
		SliceSetFieldWithStruct:       []Item{{Key: "single", Data: "one"}},
		SliceMapFieldWithMultipleKeys: []ItemWithMultipleKeys{{Key1: "a", Key2: "b", Data: "one"}},
		AtomicListUniqueSet:           []Item{{Key: "single", Data: "one"}},
		AtomicListUniqueMap:           []Item{{Key: "single", Data: "one"}},
		IntSlice:                      []int{42},
		BoolSlice:                     []bool{true},
		SliceWithZeroValues:           []string{"single"},
		SliceWithEmptyKeys:            []Item{{Key: "single", Data: "one"}},
	}).ExpectValid()

	// Test duplicate values (should fail validation)
	st.Value(&Struct{
		SliceSetField: []string{"aaa", "bbb", "ccc", "ccc", "bbb", "aaa"},
		SliceMapField: []Item{
			{Key: "key1", Data: "one"},
			{Key: "key2", Data: "two"},
			{Key: "key1", Data: "three"},
		},
		SliceSetFieldWithStruct: []Item{
			{Key: "key1", Data: "one"},
			{Key: "key2", Data: "two"},
			{Key: "key1", Data: "one"},
		},
		SliceMapFieldWithMultipleKeys: []ItemWithMultipleKeys{
			{Key1: "a", Key2: "x", Data: "first"},
			{Key1: "a", Key2: "y", Data: "second"},
			{Key1: "a", Key2: "x", Data: "third"},
		},
		AtomicListUniqueSet: []Item{
			{Key: "key1", Data: "one"},
			{Key: "key2", Data: "two"},
			{Key: "key1", Data: "one"},
		},
		AtomicListUniqueMap: []Item{
			{Key: "key1", Data: "one"},
			{Key: "key2", Data: "two"},
			{Key: "key1", Data: "three"},
		},
		IntSlice:            []int{1, 2, 3, 3, 2, 1},
		BoolSlice:           []bool{true, false, true},
		SliceWithZeroValues: []string{"a", "b", "a"},
		SliceWithEmptyKeys: []Item{
			{Key: "key1", Data: "one"},
			{Key: "key2", Data: "two"},
			{Key: "key1", Data: "three"},
		},
	}).ExpectMatches(field.ErrorMatcher{}.ByType().ByField(), field.ErrorList{
		field.Duplicate(field.NewPath("sliceSetField").Index(3), "ccc"),
		field.Duplicate(field.NewPath("sliceSetField").Index(4), "bbb"),
		field.Duplicate(field.NewPath("sliceSetField").Index(5), "aaa"),
		field.Duplicate(field.NewPath("sliceMapField").Index(2), nil),
		field.Duplicate(field.NewPath("sliceSetFieldWithStruct").Index(2), Item{Key: "key1", Data: "one"}),
		field.Duplicate(field.NewPath("sliceMapFieldWithMultipleKeys").Index(2), nil),
		field.Duplicate(field.NewPath("atomicListUniqueSet").Index(2), Item{Key: "key1", Data: "one"}),
		field.Duplicate(field.NewPath("atomicListUniqueMap").Index(2), nil),
		field.Duplicate(field.NewPath("intSlice").Index(3), 3),
		field.Duplicate(field.NewPath("intSlice").Index(4), 2),
		field.Duplicate(field.NewPath("intSlice").Index(5), 1),
		field.Duplicate(field.NewPath("boolSlice").Index(2), true),
		field.Duplicate(field.NewPath("sliceWithZeroValues").Index(2), "a"),
		field.Duplicate(field.NewPath("sliceWithEmptyKeys").Index(2), nil),
	})

	// Test with zero values and empty strings
	st.Value(&Struct{
		SliceWithZeroValues: []string{"", "a", ""},
		SliceWithEmptyKeys: []Item{
			{Key: "", Data: "one"},
			{Key: "a", Data: "two"},
			{Key: "", Data: "three"},
		},
	}).ExpectMatches(field.ErrorMatcher{}.ByType().ByField(), field.ErrorList{
		field.Duplicate(field.NewPath("sliceWithZeroValues").Index(2), ""),
		field.Duplicate(field.NewPath("sliceWithEmptyKeys").Index(2), nil),
	})

	// Test with zero values in primitive types
	st.Value(&Struct{
		IntSlice:  []int{0, 1, 0},
		BoolSlice: []bool{false, true, false},
	}).ExpectMatches(field.ErrorMatcher{}.ByType().ByField(), field.ErrorList{
		field.Duplicate(field.NewPath("intSlice").Index(2), 0),
		field.Duplicate(field.NewPath("boolSlice").Index(2), false),
	})
}

func TestRatcheting(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	struct1 := Struct{
		SliceSetField: []string{"aaa", "bbb"},
		SliceMapField: []Item{
			{Key: "key1", Data: "one"},
			{Key: "key2", Data: "two"},
		},
		SliceSetFieldWithStruct: []Item{
			{Key: "key1", Data: "one"},
			{Key: "key2", Data: "two"},
		},
		SliceMapFieldWithMultipleKeys: []ItemWithMultipleKeys{
			{Key1: "a", Key2: "x", Data: "first"},
			{Key1: "a", Key2: "y", Data: "second"},
		},
		AtomicListUniqueSet: []Item{
			{Key: "key1", Data: "one"},
			{Key: "key2", Data: "two"},
		},
		AtomicListUniqueMap: []Item{
			{Key: "key1", Data: "one"},
			{Key: "key2", Data: "two"},
		},
		IntSlice:            []int{1, 2},
		BoolSlice:           []bool{true, false},
		SliceWithZeroValues: []string{"a", "b"},
		SliceWithEmptyKeys: []Item{
			{Key: "key1", Data: "one"},
			{Key: "key2", Data: "two"},
		},
	}

	// Same data, different order.
	struct2 := Struct{
		SliceSetField: []string{"bbb", "aaa"},
		SliceMapField: []Item{
			{Key: "key2", Data: "two"},
			{Key: "key1", Data: "one"},
		},
		SliceSetFieldWithStruct: []Item{
			{Key: "key2", Data: "two"},
			{Key: "key1", Data: "one"},
		},
		SliceMapFieldWithMultipleKeys: []ItemWithMultipleKeys{
			{Key1: "a", Key2: "y", Data: "second"},
			{Key1: "a", Key2: "x", Data: "first"},
		},
		AtomicListUniqueSet: []Item{
			{Key: "key2", Data: "two"},
			{Key: "key1", Data: "one"},
		},
		AtomicListUniqueMap: []Item{
			{Key: "key2", Data: "two"},
			{Key: "key1", Data: "one"},
		},
		IntSlice:            []int{2, 1},
		BoolSlice:           []bool{false, true},
		SliceWithZeroValues: []string{"b", "a"},
		SliceWithEmptyKeys: []Item{
			{Key: "key2", Data: "two"},
			{Key: "key1", Data: "one"},
		},
	}

	// Test that reordering doesn't trigger validation errors
	st.Value(&struct1).OldValue(&struct2).ExpectValid()
	st.Value(&struct2).OldValue(&struct1).ExpectValid()

	// Test that the same data is considered valid regardless of order
	st.Value(&struct1).ExpectValid()
	st.Value(&struct2).ExpectValid()
}
