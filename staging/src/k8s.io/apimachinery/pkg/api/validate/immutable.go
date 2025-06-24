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
	"reflect"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/operation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// FrozenByCompare verifies that the specified value has not changed in the
// course of an update operation.  It does nothing if the old value is not
// provided. If the caller needs to compare types that are not trivially
// comparable, they should use FrozenByReflect instead.
// Semantics:
// - Forbids ALL transitions after creation
// - This includes: set->unset, unset->set, and modify operations
// Caution: structs with pointer fields satisfy comparable, but this function
// will only compare pointer values.  It does not compare the pointed-to
// values.
func FrozenByCompare[T comparable](_ context.Context, op operation.Operation, fldPath *field.Path, value, oldValue *T) field.ErrorList {
	if op.Type != operation.Update {
		return nil
	}
	if value == nil && oldValue == nil {
		return nil
	}
	if value == nil || oldValue == nil || *value != *oldValue {
		return field.ErrorList{
			field.Forbidden(fldPath, "field is frozen"),
		}
	}
	return nil
}

// FrozenByReflect verifies that the specified value has not changed in
// the course of an update operation.  It does nothing if the old value is not
// provided. Unlike ImmutableByCompare, this function can be used with types that are
// not directly comparable, at the cost of performance.
// Semantics:
// - Forbids ALL transitions after creation
// - This includes: set->unset (set), unset->set (clear), and modify
func FrozenByReflect[T any](_ context.Context, op operation.Operation, fldPath *field.Path, value, oldValue T) field.ErrorList {
	if op.Type != operation.Update {
		return nil
	}
	if !equality.Semantic.DeepEqual(value, oldValue) {
		return field.ErrorList{
			field.Forbidden(fldPath, "field is frozen"),
		}
	}
	return nil
}

// ImmutableValueByCompare allows a field to be set
// once then prevents any further changes.
// Semantics:
// - Zero value is considered "unset"
// - Allows ONE transition: unset->set
// - Forbids: modify and clear
// This function is optimized for comparable types.
// For non-comparable types use ImmutableByReflect instead.
func ImmutableValueByCompare[T comparable](ctx context.Context, op operation.Operation, fldPath *field.Path, value, oldValue *T) field.ErrorList {
	return immutableByCompareCheck(op, fldPath, value, oldValue, isUnsetComparable[T])
}

// ImmutablePointerByCompare allows a field to be set
// once then prevents any further changes.
// Semantics:
// - nil is considered "unset"
// - Any non-nil pointer is considered "set" (incl. ptrs to zero values)
// - Allows ONE transition: unset->set (nil -> non-nil)
// - Forbids: modify and clear (non-nil -> nil)
// This function is optimized for comparable types.
// For non-comparable types, use ImmutableByReflect instead.
func ImmutablePointerByCompare[T comparable](ctx context.Context, op operation.Operation, fldPath *field.Path, value, oldValue *T) field.ErrorList {
	return immutableByCompareCheck(op, fldPath, value, oldValue, func(v *T) bool {
		return v == nil
	})
}

// ImmutableByReflect  allows a field to be set
// once then prevents any further changes.
// Semantics:
// - Can be unset at creation
// - Allows ONE transition: set (unset->set)
// - Forbids: modify and clear (set->unset)
// Unlike ImmutableByCompare, this function can be
// used with types that are not directly comparable
// at the cost of performance.
func ImmutableByReflect[T any](_ context.Context, op operation.Operation, fldPath *field.Path, value, oldValue T) field.ErrorList {
	if op.Type != operation.Update {
		return nil
	}
	if equality.Semantic.DeepEqual(value, oldValue) {
		return nil
	}
	oldValueIsUnset := isUnsetForReflect(oldValue)
	valueIsUnset := isUnsetForReflect(value)
	if oldValueIsUnset && !valueIsUnset {
		return nil
	}
	return field.ErrorList{
		field.Forbidden(fldPath, "field is immutable"),
	}
}

func immutableByCompareCheck[T comparable](op operation.Operation,
	fldPath *field.Path, value, oldValue *T,
	isUnset func(*T) bool) field.ErrorList {
	if op.Type != operation.Update {
		return nil
	}

	if value == nil && oldValue == nil {
		return nil
	}
	if oldValue == nil {
		return nil
	}
	if value == nil {
		return field.ErrorList{
			field.Forbidden(fldPath, "field is immutable"),
		}
	}

	oldIsUnset := isUnset(oldValue)
	newIsUnset := isUnset(value)
	if oldIsUnset == newIsUnset && *value == *oldValue {
		return nil
	}
	if oldIsUnset && !newIsUnset {
		return nil
	}
	return field.ErrorList{
		field.Forbidden(fldPath, "field is immutable"),
	}
}

// isUnsetComparable determines if a comparable value should be considered
// "unset" for immutability validation by comparing to its zero value.
func isUnsetComparable[T comparable](v *T) bool {
	if v == nil {
		return true
	}
	var zero T
	return *v == zero
}

// isUnsetReflect determines if value should be considered "unset"
// for immutability validation by comparing to its zero value.
func isUnsetForReflect(value interface{}) bool {
	if value == nil {
		return true
	}

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return true
		}
		elem := v.Elem()

		// If this is a pointer to a struct, check if the struct is zero
		if elem.Kind() == reflect.Struct {
			zero := reflect.Zero(elem.Type())
			return reflect.DeepEqual(elem.Interface(), zero.Interface())
		}
		// For pointers to other types, being non-nil means it's set.
		// Aligns with +k8s:required behavior for pointer fields.
		return false
	}

	switch v.Kind() {
	case reflect.Slice, reflect.Map:
		return v.IsNil() || v.Len() == 0
	case reflect.String:
		return v.String() == ""
	case reflect.Struct:
		return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
	case reflect.Interface:
		return v.IsNil()
	default:
		// For other types check if it's the zero value.
		return v.Interface() == reflect.Zero(v.Type()).Interface()
	}
}
