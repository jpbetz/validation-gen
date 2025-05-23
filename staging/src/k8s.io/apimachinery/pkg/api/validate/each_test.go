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
		_ = EachSliceVal(context.Background(), operation.Operation{}, field.NewPath("test"), input, nil, nil, false, vfn)
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
		_ = EachSliceVal(context.Background(), operation.Operation{}, field.NewPath("test"), input, old, cmp, true, vfn)
		if calls != len(input) {
			t.Errorf("expected %d calls, got %d", len(input), calls)
		}
	})
}

func TestEachSliceValComparable(t *testing.T) {
	testEachSliceValComparable(t, "valid", []int{11, 12, 13})
	testEachSliceValComparable(t, "valid", []string{"a", "b", "c"})
	testEachSliceValComparable(t, "valid", []TestStruct{{11, "a"}, {12, "b"}, {13, "c"}})
}
func testEachSliceValComparable[T comparable](t *testing.T, name string, input []T) {
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
		_ = EachSliceValComparable(context.Background(), operation.Operation{}, field.NewPath("test"), input, nil, nil, false, vfn)
		if calls != len(input) {
			t.Errorf("expected %d calls, got %d", len(input), calls)
		}
	})
}

func TestEachSliceValRatcheting(t *testing.T) {
	testEachSliceValRatcheting(t, "same data different order",
		[]NonComparableStruct{
			{11, []string{"a"}}, {12, []string{"b"}}, {13, []string{"c"}},
		},
		[]NonComparableStruct{
			{11, []string{"a"}}, {12, []string{"b"}}, {13, []string{"c"}},
		},
		SemanticDeepEqual,
		true,
		0,
	)
	testEachSliceValRatcheting(t, "less data in new, exist in old",
		[]NonComparableStruct{
			{11, []string{"a"}}, {12, []string{"b"}}, {13, []string{"c"}},
		},
		[]NonComparableStruct{
			{11, []string{"a"}}, {13, []string{"c"}},
		},
		SemanticDeepEqual,
		true,
		0,
	)
	testEachSliceValRatcheting(t, "same data different order with key",
		[]NonComparableStructWithKey{
			{Key: "a", I: 11, S: []string{"a"}}, {Key: "b", I: 12, S: []string{"b"}}, {Key: "c", I: 13, S: []string{"c"}},
		},
		[]NonComparableStructWithKey{
			{Key: "a", I: 11, S: []string{"a"}}, {Key: "b", I: 12, S: []string{"b"}}, {Key: "c", I: 13, S: []string{"c"}},
		},
		CompareFunc[NonComparableStructWithKey](func(a, b NonComparableStructWithKey) bool {
			return a.Key == b.Key
		}),
		false,
		0,
	)
	testEachSliceValRatcheting(t, "changed data with key",
		[]NonComparableStructWithKey{
			{Key: "a", I: 11, S: []string{"a"}}, {Key: "b", I: 12, S: []string{"b"}}, {Key: "c", I: 13, S: []string{"c"}},
		},
		[]NonComparableStructWithKey{
			{Key: "a", I: 11, S: []string{"x"}}, {Key: "b", I: 12, S: []string{"y"}}, {Key: "c", I: 13, S: []string{"z"}},
		},
		CompareFunc[NonComparableStructWithKey](func(a, b NonComparableStructWithKey) bool {
			return a.Key == b.Key
		}),
		false,
		3,
	)

	testEachSliceValComparableRatcheting(t, "same data different order", []int{11, 13, 12}, []int{11, 12, 13}, DirectEqual, true, 0)
	testEachSliceValComparableRatcheting(t, "same data different order", []string{"a", "c", "b"}, []string{"a", "b", "c"}, DirectEqual, true, 0)
	testEachSliceValComparableRatcheting(t, "less data in new, not exist in old", []string{"a", "c", "b"}, []string{"b", "c"}, DirectEqual, true, 0)
	testEachSliceValComparableRatcheting(t, "same data different order", []TestStruct{{11, "a"}, {13, "c"}, {12, "b"}}, []TestStruct{{11, "a"}, {12, "b"}, {13, "c"}}, DirectEqual, true, 0)
	testEachSliceValComparableRatcheting(t, "less data in new, not exist in old", []TestStruct{{11, "a"}, {13, "c"}, {12, "b"}}, []TestStruct{{12, "b"}, {13, "c"}}, DirectEqual, true, 0)
	testEachSliceValComparableRatcheting(t, "same data different order with key", []TestStructWithKey{
		{Key: "a", I: 11, D: "a"},
		{Key: "b", I: 12, D: "b"},
		{Key: "c", I: 13, D: "c"},
	},
		[]TestStructWithKey{
			{Key: "a", I: 11, D: "a"},
			{Key: "c", I: 13, D: "c"},
			{Key: "b", I: 12, D: "b"},
		},
		func(a, b TestStructWithKey) bool {
			return a.Key == b.Key
		},
		false,
		0,
	)
	testEachSliceValComparableRatcheting(t, "changed data with key",
		[]TestStructWithKey{
			{Key: "a", I: 11, D: "a"}, {Key: "b", I: 12, D: "b"}, {Key: "c", I: 13, D: "c"},
		},
		[]TestStructWithKey{
			{Key: "a", I: 11, D: "x"}, {Key: "b", I: 12, D: "y"}, {Key: "c", I: 13, D: "z"},
		},
		CompareFunc[TestStructWithKey](func(a, b TestStructWithKey) bool {
			return a.Key == b.Key
		}),
		false,
		3,
	)
}

func testEachSliceValRatcheting[T any](t *testing.T, name string, old, new []T, cmp CompareFunc[T], isEquivalenceCompare bool, wantCalls int) {
	t.Helper()
	var zero T
	t.Run(fmt.Sprintf("%s(%T)", name, zero), func(t *testing.T) {
		calls := 0
		vfn := func(ctx context.Context, op operation.Operation, fldPath *field.Path, newVal, oldVal *T) field.ErrorList {
			calls++
			return nil
		}
		_ = EachSliceVal(context.Background(), operation.Operation{Type: operation.Update}, field.NewPath("test"), new, old, cmp, isEquivalenceCompare, vfn)
		if calls != wantCalls {
			t.Errorf("expected %d calls, got %d", wantCalls, calls)
		}
	})
}

func testEachSliceValComparableRatcheting[T comparable](t *testing.T, name string, old, new []T, cmp CompareFunc[T], isEquivalenceCompare bool, wantCalls int) {
	t.Helper()
	var zero T
	t.Run(fmt.Sprintf("%s(%T)", name, zero), func(t *testing.T) {
		calls := 0
		vfn := func(ctx context.Context, op operation.Operation, fldPath *field.Path, newVal, oldVal *T) field.ErrorList {
			calls++
			return nil
		}
		_ = EachSliceValComparable(context.Background(), operation.Operation{Type: operation.Update}, field.NewPath("test"), new, old, cmp, isEquivalenceCompare, vfn)
		if calls != wantCalls {
			t.Errorf("expected %d calls, got %d", wantCalls, calls)
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
