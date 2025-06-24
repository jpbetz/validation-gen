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
	"testing"

	"k8s.io/apimachinery/pkg/api/operation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"
)

type StructComparable struct {
	S string
	I int
	B bool
}

func TestFrozenByCompare(t *testing.T) {
	structA := StructComparable{"abc", 123, true}
	structA2 := structA
	structB := StructComparable{"xyz", 456, false}

	for _, tc := range []struct {
		name string
		fn   func(operation.Operation, *field.Path) field.ErrorList
		fail bool
	}{{
		name: "nil both values",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return FrozenByCompare[int](context.Background(), op, fld, nil, nil)
		},
	}, {
		name: "nil value",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return FrozenByCompare(context.Background(), op, fld, nil, ptr.To(123))
		},
		fail: true,
	}, {
		name: "nil oldValue",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return FrozenByCompare(context.Background(), op, fld, ptr.To(123), nil)
		},
		fail: true,
	}, {
		name: "int",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return FrozenByCompare(context.Background(), op, fld, ptr.To(123), ptr.To(123))
		},
	}, {
		name: "int fail",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return FrozenByCompare(context.Background(), op, fld, ptr.To(123), ptr.To(456))
		},
		fail: true,
	}, {
		name: "string",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return FrozenByCompare(context.Background(), op, fld, ptr.To("abc"), ptr.To("abc"))
		},
	}, {
		name: "string fail",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return FrozenByCompare(context.Background(), op, fld, ptr.To("abc"), ptr.To("xyz"))
		},
		fail: true,
	}, {
		name: "bool",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return FrozenByCompare(context.Background(), op, fld, ptr.To(true), ptr.To(true))
		},
	}, {
		name: "bool fail",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return FrozenByCompare(context.Background(), op, fld, ptr.To(true), ptr.To(false))
		},
		fail: true,
	}, {
		name: "same struct",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return FrozenByCompare(context.Background(), op, fld, ptr.To(structA), ptr.To(structA))
		},
	}, {
		name: "equal struct",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return FrozenByCompare(context.Background(), op, fld, ptr.To(structA), ptr.To(structA2))
		},
	}, {
		name: "struct fail",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return FrozenByCompare(context.Background(), op, fld, ptr.To(structA), ptr.To(structB))
		},
		fail: true,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			errs := tc.fn(operation.Operation{Type: operation.Create}, field.NewPath(""))
			if len(errs) != 0 {
				t.Errorf("case %q (create): expected success: %v", tc.name, errs)
			}
			errs = tc.fn(operation.Operation{Type: operation.Update}, field.NewPath(""))
			if tc.fail && len(errs) == 0 {
				t.Errorf("case %q (update): expected failure", tc.name)
			} else if !tc.fail && len(errs) != 0 {
				t.Errorf("case %q (update): expected success: %v", tc.name, errs)
			}
		})
	}
}

type StructNonComparable struct {
	S   string
	SP  *string
	I   int
	IP  *int
	B   bool
	BP  *bool
	SS  []string
	MSS map[string]string
}

func TestFrozenByReflect(t *testing.T) {
	structA := StructNonComparable{
		S:   "abc",
		SP:  ptr.To("abc"),
		I:   123,
		IP:  ptr.To(123),
		B:   true,
		BP:  ptr.To(true),
		SS:  []string{"a", "b", "c"},
		MSS: map[string]string{"a": "b", "c": "d"},
	}

	structA2 := structA
	structA2.SP = ptr.To("abc")
	structA2.IP = ptr.To(123)
	structA2.BP = ptr.To(true)
	structA2.SS = []string{"a", "b", "c"}
	structA2.MSS = map[string]string{"a": "b", "c": "d"}

	structB := StructNonComparable{
		S:   "xyz",
		SP:  ptr.To("xyz"),
		I:   456,
		IP:  ptr.To(456),
		B:   false,
		BP:  ptr.To(false),
		SS:  []string{"x", "y", "z"},
		MSS: map[string]string{"x": "X", "y": "Y"},
	}

	for _, tc := range []struct {
		name string
		fn   func(operation.Operation, *field.Path) field.ErrorList
		fail bool
	}{{
		name: "nil both values",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return FrozenByReflect[*int](context.Background(), op, fld, nil, nil)
		},
	}, {
		name: "nil value",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return FrozenByReflect(context.Background(), op, fld, nil, ptr.To(123))
		},
		fail: true,
	}, {
		name: "nil oldValue",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return FrozenByReflect(context.Background(), op, fld, ptr.To(123), nil)
		},
		fail: true,
	}, {
		name: "int",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return FrozenByReflect(context.Background(), op, fld, ptr.To(123), ptr.To(123))
		},
	}, {
		name: "int fail",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return FrozenByReflect(context.Background(), op, fld, ptr.To(123), ptr.To(456))
		},
		fail: true,
	}, {
		name: "string",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return FrozenByReflect(context.Background(), op, fld, ptr.To("abc"), ptr.To("abc"))
		},
	}, {
		name: "string fail",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return FrozenByReflect(context.Background(), op, fld, ptr.To("abc"), ptr.To("xyz"))
		},
		fail: true,
	}, {
		name: "bool",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return FrozenByReflect(context.Background(), op, fld, ptr.To(true), ptr.To(true))
		},
	}, {
		name: "bool fail",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return FrozenByReflect(context.Background(), op, fld, ptr.To(true), ptr.To(false))
		},
		fail: true,
	}, {
		name: "same struct",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return FrozenByReflect(context.Background(), op, fld, ptr.To(structA), ptr.To(structA))
		},
	}, {
		name: "equal struct",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return FrozenByReflect(context.Background(), op, fld, ptr.To(structA), ptr.To(structA2))
		},
	}, {
		name: "struct fail",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return FrozenByReflect(context.Background(), op, fld, ptr.To(structA), ptr.To(structB))
		},
		fail: true,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			errs := tc.fn(operation.Operation{Type: operation.Create}, field.NewPath(""))
			if len(errs) != 0 {
				t.Errorf("case %q (create): expected success: %v", tc.name, errs)
			}
			errs = tc.fn(operation.Operation{Type: operation.Update}, field.NewPath(""))
			if tc.fail && len(errs) == 0 {
				t.Errorf("case %q (update): expected failure", tc.name)
			} else if !tc.fail && len(errs) != 0 {
				t.Errorf("case %q (update): expected success: %v", tc.name, errs)
			}
		})
	}
}

func TestFrozenVariantsConsistency(t *testing.T) {
	for _, tc := range []struct {
		name     string
		oldValue *string
		newValue *string
	}{
		{"string both nil", nil, nil},
		{"string nil to empty", nil, ptr.To("")},
		{"string nil to value", nil, ptr.To("hello")},
		{"string empty to value", ptr.To(""), ptr.To("hello")},
		{"string value to empty", ptr.To("hello"), ptr.To("")},
		{"string same value", ptr.To("hello"), ptr.To("hello")},
		{"string different values", ptr.To("hello"), ptr.To("world")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			op := operation.Operation{Type: operation.Update}
			path := field.NewPath("test")

			errs1 := FrozenByCompare(ctx, op, path, tc.newValue, tc.oldValue)
			errs2 := FrozenByReflect(ctx, op, path, tc.newValue, tc.oldValue)

			if len(errs1) != len(errs2) {
				t.Errorf("FrozenByCompare and FrozenByReflect differ: %v, %v",
					errs1, errs2)
			}
		})
	}
}

func TestImmutableValueByCompare(t *testing.T) {
	for _, tc := range []struct {
		name string
		fn   func(operation.Operation, *field.Path) field.ErrorList
		fail bool
	}{{
		name: "nil both values",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return ImmutableValueByCompare[int](context.Background(), op, fld, nil, nil)
		},
	}, {
		name: "nil value pointer",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return ImmutableValueByCompare(context.Background(), op, fld, nil, ptr.To(123))
		},
		fail: true,
	}, {
		name: "nil oldValue pointer",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return ImmutableValueByCompare(context.Background(), op, fld, ptr.To(123), nil)
		},
	}, {
		name: "int zero to non-zero (unset to set)",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return ImmutableValueByCompare(context.Background(), op, fld, ptr.To(123), ptr.To(0))
		},
	}, {
		name: "int non-zero to zero (clear)",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return ImmutableValueByCompare(context.Background(), op, fld, ptr.To(0), ptr.To(123))
		},
		fail: true,
	}, {
		name: "int modify",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return ImmutableValueByCompare(context.Background(), op, fld, ptr.To(456), ptr.To(123))
		},
		fail: true,
	}, {
		name: "int same value",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return ImmutableValueByCompare(context.Background(), op, fld, ptr.To(123), ptr.To(123))
		},
	}, {
		name: "string empty to non-empty (unset to set)",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return ImmutableValueByCompare(context.Background(), op, fld, ptr.To("abc"), ptr.To(""))
		},
	}, {
		name: "string non-empty to empty (clear)",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return ImmutableValueByCompare(context.Background(), op, fld, ptr.To(""), ptr.To("abc"))
		},
		fail: true,
	}, {
		name: "string modify",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return ImmutableValueByCompare(context.Background(), op, fld, ptr.To("xyz"), ptr.To("abc"))
		},
		fail: true,
	}, {
		name: "bool false to true (unset to set)",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return ImmutableValueByCompare(context.Background(), op, fld, ptr.To(true), ptr.To(false))
		},
	}, {
		name: "bool true to false (clear)",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return ImmutableValueByCompare(context.Background(), op, fld, ptr.To(false), ptr.To(true))
		},
		fail: true,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			errs := tc.fn(operation.Operation{Type: operation.Create}, field.NewPath(""))
			if len(errs) != 0 {
				t.Errorf("case %q (create): expected success: %v", tc.name, errs)
			}
			errs = tc.fn(operation.Operation{Type: operation.Update}, field.NewPath(""))
			if tc.fail && len(errs) == 0 {
				t.Errorf("case %q (update): expected failure", tc.name)
			} else if !tc.fail && len(errs) != 0 {
				t.Errorf("case %q (update): expected success: %v", tc.name, errs)
			}
		})
	}
}

func TestImmutablePointerByCompare(t *testing.T) {
	for _, tc := range []struct {
		name string
		fn   func(operation.Operation, *field.Path) field.ErrorList
		fail bool
	}{{
		name: "nil both values",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return ImmutablePointerByCompare[int](context.Background(), op, fld, nil, nil)
		},
	}, {
		name: "nil to non-nil (unset to set)",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return ImmutablePointerByCompare(context.Background(), op, fld, ptr.To(123), nil)
		},
	}, {
		name: "non-nil to nil (clear)",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return ImmutablePointerByCompare(context.Background(), op, fld, nil, ptr.To(123))
		},
		fail: true,
	}, {
		name: "int pointer same value",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return ImmutablePointerByCompare(context.Background(), op, fld, ptr.To(123), ptr.To(123))
		},
	}, {
		name: "int pointer modify",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return ImmutablePointerByCompare(context.Background(), op, fld, ptr.To(456), ptr.To(123))
		},
		fail: true,
	}, {
		name: "string pointer nil to empty string (unset to set)",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			var oldVal *string = nil
			return ImmutablePointerByCompare(context.Background(), op, fld, ptr.To(""), oldVal)
		},
	}, {
		name: "string pointer empty to non-empty (modify)",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return ImmutablePointerByCompare(context.Background(), op, fld, ptr.To("abc"), ptr.To(""))
		},
		fail: true,
	}, {
		name: "bool pointer nil to false (unset to set)",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			var oldVal *bool = nil
			return ImmutablePointerByCompare(context.Background(), op, fld, ptr.To(false), oldVal)
		},
	}, {
		name: "bool pointer false to true (modify)",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return ImmutablePointerByCompare(context.Background(), op, fld, ptr.To(true), ptr.To(false))
		},
		fail: true,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			errs := tc.fn(operation.Operation{Type: operation.Create}, field.NewPath(""))
			if len(errs) != 0 {
				t.Errorf("case %q (create): expected success: %v", tc.name, errs)
			}
			errs = tc.fn(operation.Operation{Type: operation.Update}, field.NewPath(""))
			if tc.fail && len(errs) == 0 {
				t.Errorf("case %q (update): expected failure", tc.name)
			} else if !tc.fail && len(errs) != 0 {
				t.Errorf("case %q (update): expected success: %v", tc.name, errs)
			}
		})
	}
}

func TestImmutableByReflect(t *testing.T) {
	emptySlice := []string{}
	nonEmptySlice := []string{"a", "b", "c"}
	emptyMap := map[string]string{}
	nonEmptyMap := map[string]string{"key": "value"}

	structA := StructNonComparable{
		S:   "abc",
		SP:  ptr.To("abc"),
		I:   123,
		IP:  ptr.To(123),
		B:   true,
		BP:  ptr.To(true),
		SS:  []string{"a", "b", "c"},
		MSS: map[string]string{"a": "b", "c": "d"},
	}

	structB := StructNonComparable{
		S:   "xyz",
		SP:  ptr.To("xyz"),
		I:   456,
		IP:  ptr.To(456),
		B:   false,
		BP:  ptr.To(false),
		SS:  []string{"x", "y", "z"},
		MSS: map[string]string{"x": "X", "y": "Y"},
	}

	structZero := StructNonComparable{}

	for _, tc := range []struct {
		name string
		fn   func(operation.Operation, *field.Path) field.ErrorList
		fail bool
	}{{
		name: "slice nil to empty (both unset)",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			var nilSlice []string
			return ImmutableByReflect(context.Background(), op, fld, emptySlice, nilSlice)
		},
	}, {
		name: "slice empty to nil (both unset)",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			var nilSlice []string
			return ImmutableByReflect(context.Background(), op, fld, nilSlice, emptySlice)
		},
	}, {
		name: "slice nil to non-empty (unset to set)",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			var nilSlice []string
			return ImmutableByReflect(context.Background(), op, fld, nonEmptySlice, nilSlice)
		},
	}, {
		name: "slice non-empty to nil (clear)",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			var nilSlice []string
			return ImmutableByReflect(context.Background(), op, fld, nilSlice, nonEmptySlice)
		},
		fail: true,
	}, {
		name: "slice modify",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			slice1 := []string{"a", "b"}
			slice2 := []string{"x", "y"}
			return ImmutableByReflect(context.Background(), op, fld, slice2, slice1)
		},
		fail: true,
	}, {
		name: "map nil to empty (both unset)",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			var nilMap map[string]string
			return ImmutableByReflect(context.Background(), op, fld, emptyMap, nilMap)
		},
	}, {
		name: "map nil to non-empty (unset to set)",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			var nilMap map[string]string
			return ImmutableByReflect(context.Background(), op, fld, nonEmptyMap, nilMap)
		},
	}, {
		name: "map non-empty to nil (clear)",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			var nilMap map[string]string
			return ImmutableByReflect(context.Background(), op, fld, nilMap, nonEmptyMap)
		},
		fail: true,
	}, {
		name: "struct zero to non-zero (unset to set)",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return ImmutableByReflect(context.Background(), op, fld, structA, structZero)
		},
	}, {
		name: "struct non-zero to zero (clear)",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return ImmutableByReflect(context.Background(), op, fld, structZero, structA)
		},
		fail: true,
	}, {
		name: "struct modify",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return ImmutableByReflect(context.Background(), op, fld, structB, structA)
		},
		fail: true,
	}, {
		name: "struct same value",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			structA2 := structA
			return ImmutableByReflect(context.Background(), op, fld, structA2, structA)
		},
	}, {
		name: "pointer to struct - nil to non-nil (unset to set)",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return ImmutableByReflect(context.Background(), op, fld, &structA, (*StructNonComparable)(nil))
		},
	}, {
		name: "pointer to struct - non-nil to nil (clear)",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return ImmutableByReflect(context.Background(), op, fld, (*StructNonComparable)(nil), &structA)
		},
		fail: true,
	}, {
		name: "int value zero to non-zero (unset to set)",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return ImmutableByReflect(context.Background(), op, fld, 123, 0)
		},
	}, {
		name: "string empty to non-empty (unset to set)",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return ImmutableByReflect(context.Background(), op, fld, "hello", "")
		},
	}, {
		name: "bool false to true (unset to set)",
		fn: func(op operation.Operation, fld *field.Path) field.ErrorList {
			return ImmutableByReflect(context.Background(), op, fld, true, false)
		},
	}} {
		t.Run(tc.name, func(t *testing.T) {
			errs := tc.fn(operation.Operation{Type: operation.Create}, field.NewPath(""))
			if len(errs) != 0 {
				t.Errorf("case %q (create): expected success: %v", tc.name, errs)
			}
			errs = tc.fn(operation.Operation{Type: operation.Update}, field.NewPath(""))
			if tc.fail && len(errs) == 0 {
				t.Errorf("case %q (update): expected failure", tc.name)
			} else if !tc.fail && len(errs) != 0 {
				t.Errorf("case %q (update): expected success: %v", tc.name, errs)
			}
		})
	}
}

func TestImmutableVariantsConsistency(t *testing.T) {
	for _, tc := range []struct {
		name     string
		oldValue *string
		newValue *string
	}{
		{"string both nil", nil, nil},
		{"string nil to empty", nil, ptr.To("")},
		{"string nil to value", nil, ptr.To("hello")},
		{"string empty to value", ptr.To(""), ptr.To("hello")},
		{"string value to empty", ptr.To("hello"), ptr.To("")},
		{"string same value", ptr.To("hello"), ptr.To("hello")},
		{"string different values", ptr.To("hello"), ptr.To("world")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			op := operation.Operation{Type: operation.Update}
			path := field.NewPath("test")

			errs1 := ImmutablePointerByCompare(ctx, op, path, tc.newValue, tc.oldValue)
			errs2 := ImmutableByReflect(ctx, op, path, tc.newValue, tc.oldValue)

			if len(errs1) != len(errs2) {
				t.Errorf("ImmutablePointerByCompare and ImmutableByReflect differ: %v, %v",
					errs1, errs2)
			}

			if tc.oldValue != nil && tc.newValue != nil {
				errs3 := ImmutableValueByCompare(ctx, op, path, tc.newValue, tc.oldValue)
				errs4 := ImmutableByReflect(ctx, op, path, *tc.newValue, *tc.oldValue)

				if len(errs3) != len(errs4) {
					t.Errorf("ImmutableValueByCompare and ImmutableByReflect differ: %v, %v",
						errs3, errs4)
				}
			}
		})
	}
}
