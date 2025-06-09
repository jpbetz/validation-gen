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

package validate

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"testing"

	"k8s.io/apimachinery/pkg/api/operation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type TestStruct struct {
	I int
	D string
}

type TestStructWithKey struct {
	Key string
	I   int
	D   string
}

type NonComparableStruct struct {
	I int
	S []string
}

type NonComparableStructWithKey struct {
	Key string
	I   int
	S   []string
}

func TestEachSliceVal(t *testing.T) {
	testEachSliceVal(t, "valid", []int{11, 12, 13})
	testEachSliceVal(t, "valid", []string{"a", "b", "c"})
	testEachSliceVal(t, "valid", []TestStruct{{11, "a"}, {12, "b"}, {13, "c"}})

	testEachSliceVal(t, "empty", []int{})
	testEachSliceVal(t, "empty", []string{})
	testEachSliceVal(t, "empty", []TestStruct{})

	testEachSliceVal[int](t, "nil", nil)
	testEachSliceVal[string](t, "nil", nil)
	testEachSliceVal[TestStruct](t, "nil", nil)

	testEachSliceValUpdate(t, "valid", []int{11, 12, 13})
	testEachSliceValUpdate(t, "valid", []string{"a", "b", "c"})
	testEachSliceValUpdate(t, "valid", []TestStruct{{11, "a"}, {12, "b"}, {13, "c"}})

	testEachSliceValUpdate(t, "empty", []int{})
	testEachSliceValUpdate(t, "empty", []string{})
	testEachSliceValUpdate(t, "empty", []TestStruct{})

	testEachSliceValUpdate[int](t, "nil", nil)
	testEachSliceValUpdate[string](t, "nil", nil)
	testEachSliceValUpdate[TestStruct](t, "nil", nil)
}

func testEachSliceVal[T any](t *testing.T, name string, input []T) {
	t.Helper()
	var zero T
	t.Run(fmt.Sprintf("%s(%T)", name, zero), func(t *testing.T) {
		calls := 0
		vfn := func(ctx context.Context, op operation.Operation, fldPath *field.Path, newVal, oldVal *T) field.ErrorList {
			if oldVal != nil {
				t.Errorf("expected nil oldVal, got %v", *oldVal)
			}
			calls++
			return nil
		}
		_ = EachSliceVal(context.Background(), operation.Operation{}, field.NewPath("test"), input, nil, nil, nil, vfn)
		if calls != len(input) {
			t.Errorf("expected %d calls, got %d", len(input), calls)
		}
	})
}

func testEachSliceValUpdate[T any](t *testing.T, name string, input []T) {
	t.Helper()
	var zero T
	t.Run(fmt.Sprintf("%s(%T)", name, zero), func(t *testing.T) {
		calls := 0
		vfn := func(ctx context.Context, op operation.Operation, fldPath *field.Path, newVal, oldVal *T) field.ErrorList {
			if oldVal == nil {
				t.Fatalf("expected non-nil oldVal")
			}
			if !reflect.DeepEqual(*newVal, *oldVal) {
				t.Errorf("expected oldVal == newVal, got %v, %v", *oldVal, *newVal)
			}
			calls++
			return nil
		}
		old := make([]T, len(input))
		copy(old, input)
		slices.Reverse(old)
		cmp := func(a, b T) bool { return reflect.DeepEqual(a, b) }
		_ = EachSliceVal(context.Background(), operation.Operation{}, field.NewPath("test"), input, old, cmp, cmp, vfn)
		if calls != len(input) {
			t.Errorf("expected %d calls, got %d", len(input), calls)
		}
	})
}

func TestEachSliceValRatcheting(t *testing.T) {
	testEachSliceValRatcheting(t, "ComparableStruct same data different order",
		[]TestStruct{
			{11, "a"}, {12, "b"}, {13, "c"},
		},
		[]TestStruct{
			{11, "a"}, {13, "c"}, {12, "b"},
		},
		SemanticDeepEqual,
		nil,
	)
	testEachSliceValRatcheting(t, "ComparableStruct less data in new, exist in old",
		[]TestStruct{
			{11, "a"}, {12, "b"}, {13, "c"},
		},
		[]TestStruct{
			{11, "a"}, {13, "c"},
		},
		DirectEqual,
		nil,
	)
	testEachSliceValRatcheting(t, "Comparable struct with key same data different order",
		[]TestStructWithKey{
			{Key: "a", I: 11, D: "a"}, {Key: "b", I: 12, D: "b"}, {Key: "c", I: 13, D: "c"},
		},
		[]TestStructWithKey{
			{Key: "a", I: 11, D: "a"}, {Key: "c", I: 13, D: "c"}, {Key: "b", I: 12, D: "b"},
		},
		CompareFunc[TestStructWithKey](func(a, b TestStructWithKey) bool {
			return a.Key == b.Key
		}),
		DirectEqual,
	)
	testEachSliceValRatcheting(t, "Comparable struct with key less data in new, exist in old",
		[]TestStructWithKey{
			{Key: "a", I: 11, D: "a"}, {Key: "b", I: 12, D: "b"}, {Key: "c", I: 13, D: "c"},
		},
		[]TestStructWithKey{
			{Key: "a", I: 11, D: "a"}, {Key: "c", I: 13, D: "c"},
		},
		CompareFunc[TestStructWithKey](func(a, b TestStructWithKey) bool {
			return a.Key == b.Key
		}),
		DirectEqual,
	)
	testEachSliceValRatcheting(t, "NonComparableStruct same data different order",
		[]NonComparableStruct{
			{I: 11, S: []string{"a"}}, {I: 12, S: []string{"b"}}, {I: 13, S: []string{"c"}},
		},
		[]NonComparableStruct{
			{I: 11, S: []string{"a"}}, {I: 13, S: []string{"c"}}, {I: 12, S: []string{"b"}},
		},
		SemanticDeepEqual,
		nil,
	)
	testEachSliceValRatcheting(t, "NonComparableStructWithKey same data different order",
		[]NonComparableStructWithKey{
			{Key: "a", I: 11, S: []string{"a"}}, {Key: "b", I: 12, S: []string{"b"}}, {Key: "c", I: 13, S: []string{"c"}},
		},
		[]NonComparableStructWithKey{
			{Key: "a", I: 11, S: []string{"a"}}, {Key: "b", I: 12, S: []string{"b"}}, {Key: "c", I: 13, S: []string{"c"}},
		},
		CompareFunc[NonComparableStructWithKey](func(a, b NonComparableStructWithKey) bool {
			return a.Key == b.Key
		}),
		SemanticDeepEqual,
	)

}

func testEachSliceValRatcheting[T any](t *testing.T, name string, old, new []T, cmp, equiv CompareFunc[T]) {
	t.Helper()
	var zero T
	t.Run(fmt.Sprintf("%s(%T)", name, zero), func(t *testing.T) {
		vfn := func(ctx context.Context, op operation.Operation, fldPath *field.Path, newVal, oldVal *T) field.ErrorList {
			return field.ErrorList{field.Invalid(fldPath, *newVal, "expected no calls")}
		}
		errs := EachSliceVal(context.Background(), operation.Operation{Type: operation.Update}, field.NewPath("test"), new, old, cmp, equiv, vfn)
		if len(errs) > 0 {
			t.Errorf("expected no errors, got %d: %s", len(errs), fmtErrs(errs))
		}
	})
}

func TestEachMapVal(t *testing.T) {
	testEachMapVal(t, "valid", map[string]int{"one": 11, "two": 12, "three": 13})
	testEachMapVal(t, "valid", map[string]string{"A": "a", "B": "b", "C": "c"})
	testEachMapVal(t, "valid", map[string]TestStruct{"one": {11, "a"}, "two": {12, "b"}, "three": {13, "c"}})

	testEachMapVal(t, "empty", map[string]int{})
	testEachMapVal(t, "empty", map[string]string{})
	testEachMapVal(t, "empty", map[string]TestStruct{})

	testEachMapVal[int](t, "nil", nil)
	testEachMapVal[string](t, "nil", nil)
	testEachMapVal[TestStruct](t, "nil", nil)
}

func testEachMapVal[T any](t *testing.T, name string, input map[string]T) {
	t.Helper()
	var zero T
	t.Run(fmt.Sprintf("%s(%T)", name, zero), func(t *testing.T) {
		calls := 0
		vfn := func(ctx context.Context, op operation.Operation, fldPath *field.Path, newVal, oldVal *T) field.ErrorList {
			if oldVal != nil {
				t.Errorf("expected nil oldVal, got %v", *oldVal)
			}
			calls++
			return nil
		}
		_ = EachMapVal(context.Background(), operation.Operation{}, field.NewPath("test"), input, nil, vfn)
		if calls != len(input) {
			t.Errorf("expected %d calls, got %d", len(input), calls)
		}
	})
}

type StringType string

func TestEachMapKey(t *testing.T) {
	testEachMapKey(t, "valid", map[string]int{"one": 11, "two": 12, "three": 13})
	testEachMapKey(t, "valid", map[StringType]string{"A": "a", "B": "b", "C": "c"})
}

func testEachMapKey[K ~string, V any](t *testing.T, name string, input map[K]V) {
	t.Helper()
	var zero K
	t.Run(fmt.Sprintf("%s(%T)", name, zero), func(t *testing.T) {
		calls := 0
		vfn := func(ctx context.Context, op operation.Operation, fldPath *field.Path, newVal, oldVal *K) field.ErrorList {
			if oldVal != nil {
				t.Errorf("expected nil oldVal, got %v", *oldVal)
			}
			calls++
			return nil
		}
		_ = EachMapKey(context.Background(), operation.Operation{}, field.NewPath("test"), input, nil, vfn)
		if calls != len(input) {
			t.Errorf("expected %d calls, got %d", len(input), calls)
		}
	})
}

func TestUniqueByCompare(t *testing.T) {
	testUniqueByCompare(t, "int_nil", []int(nil), 0)
	testUniqueByCompare(t, "int_empty", []int{}, 0)
	testUniqueByCompare(t, "int_uniq", []int{1, 2, 3}, 0)
	testUniqueByCompare(t, "int_dup", []int{1, 2, 3, 2, 1}, 2)

	testUniqueByCompare(t, "string_nil", []string(nil), 0)
	testUniqueByCompare(t, "string_empty", []string{}, 0)
	testUniqueByCompare(t, "string_uniq", []string{"a", "b", "c"}, 0)
	testUniqueByCompare(t, "string_dup", []string{"a", "a", "c", "b", "a"}, 2)

	type isComparable struct {
		I int
		S string
	}

	testUniqueByCompare(t, "struct_nil", []isComparable(nil), 0)
	testUniqueByCompare(t, "struct_empty", []isComparable{}, 0)
	testUniqueByCompare(t, "struct_uniq", []isComparable{{1, "a"}, {2, "b"}, {3, "c"}}, 0)
	testUniqueByCompare(t, "struct_dup", []isComparable{{1, "a"}, {2, "b"}, {3, "c"}, {2, "b"}, {1, "a"}}, 2)
}

func testUniqueByCompare[T comparable](t *testing.T, name string, input []T, wantErrs int) {
	t.Helper()
	var zero T
	t.Run(fmt.Sprintf("%s(%T)", name, zero), func(t *testing.T) {
		errs := UniqueByCompare(context.Background(), operation.Operation{}, field.NewPath("test"), input, nil)
		if len(errs) != wantErrs {
			t.Errorf("expected %d errors, got %d: %s", wantErrs, len(errs), fmtErrs(errs))
		}
	})
}

func TestUniqueByReflect(t *testing.T) {
	type nonComparable struct {
		I int
		S []string
	}

	testUniqueByReflect(t, "noncomp_nil", []nonComparable(nil), 0)
	testUniqueByReflect(t, "noncomp_empty", []nonComparable{}, 0)
	testUniqueByReflect(t, "noncomp_uniq", []nonComparable{{1, []string{"a"}}, {2, []string{"b"}}, {3, []string{"c"}}}, 0)
	testUniqueByReflect(t, "noncomp_dup", []nonComparable{
		{1, []string{"a"}},
		{2, []string{"b"}},
		{3, []string{"c"}},
		{2, []string{"b"}},
		{1, []string{"a"}}}, 2)
}

func testUniqueByReflect[T any](t *testing.T, name string, input []T, wantErrs int) {
	t.Helper()
	var zero T
	t.Run(fmt.Sprintf("%s(%T)", name, zero), func(t *testing.T) {
		errs := UniqueByReflect(context.Background(), operation.Operation{}, field.NewPath("test"), input, nil)
		if len(errs) != wantErrs {
			t.Errorf("expected %d errors, got %d: %s", wantErrs, len(errs), fmtErrs(errs))
		}
	})
}

type listItem struct {
	key1 string
	key2 string
	val  string
}

func TestUniqueByFunc(t *testing.T) {
	testUniqueByFunc(t, "valid", []listItem{{"key1", "key2", "val1"}, {"key1", "key2", "val2"}, {"key1", "key3", "val3"}}, 1)
	testUniqueByFunc(t, "valid", []listItem{{"key1", "key2", "val1"}, {"key1", "key2", "val2"}, {"key1", "key2", "val3"}}, 2)
	testUniqueByFunc(t, "valid", []listItem{{"key1", "key2", "val1"}, {"key3", "key4", "val2"}, {"key5", "key6", "val3"}}, 0)
}

func testUniqueByFunc(t *testing.T, name string, input []listItem, wantErrs int) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		errs := UniqueByFunc(context.Background(), operation.Operation{}, field.NewPath("test"), input, nil, func(a listItem, b listItem) bool {
			return a.key1 == b.key1 && a.key2 == b.key2
		})
		if len(errs) != wantErrs {
			t.Errorf("expected %d errors, got %d: %s", wantErrs, len(errs), fmtErrs(errs))
		}
	})
}
