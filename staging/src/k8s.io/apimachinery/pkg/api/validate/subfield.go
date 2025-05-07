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
	"strings"

	"k8s.io/apimachinery/pkg/api/operation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// GetFieldFunc is a function that extracts a field from a type and returns a
// nilable value.
type GetFieldFunc[Tstruct any, Tfield any] func(*Tstruct) Tfield

// Subfield validates a subfield of a struct against a validator function.
func Subfield[Tstruct any, Tfield any](ctx context.Context, op operation.Operation, fldPath *field.Path, newStruct, oldStruct *Tstruct,
	fldName string, getField GetFieldFunc[Tstruct, Tfield], validator ValidateFunc[Tfield]) field.ErrorList {
	var errs field.ErrorList
	var newVal Tfield // zero value if newStruct is nil
	if newStruct != nil {
		newVal = getField(newStruct)
	}
	var oldVal Tfield // zero value if oldStruct is nil
	if oldStruct != nil {
		oldVal = getField(oldStruct)
	}
	errs = append(errs, validator(ctx, op, fldPath.Child(fldName), newVal, oldVal)...)
	return errs
}

// ListMapElementByKey validates a specific element in a list (identified by a key-value pair in the element)
// against a validator function.
// TListParent is the type of the struct containing the list field.
// TList is the type of the list itself, constrained to be a slice of TElements (e.g., []MyElement).
// TElement is the type of the elements in the list (e.g., MyElement).
// staging/src/k8s.io/apimachinery/pkg/api/validate/subfield.go
// TItem is the type of the elements in the slice (e.g., MyCondition).
// The elementValidator will be called with *TItem.
func ListMapElementByKey[TList ~[]TItem, TItem any](
	ctx context.Context, op operation.Operation, fldPath *field.Path,
	newList, oldList TList, // newList is e.g. []MyCondition. TItem is MyCondition.
	keyName string,
	keyValue string,
	elementValidator func(ctx context.Context, op operation.Operation, fldPath *field.Path, newObj, oldObj *TItem) field.ErrorList,
) field.ErrorList {
	var errs field.ErrorList

	var foundNewElementPtr *TItem // Pointer to the element in newList
	var foundOldElementPtr *TItem // Pointer to the element in oldList

	for i := range newList {
		// newList[i] is type TItem. We need to check its fields.
		// For getReflectedJSONFieldValueAsString, we need reflect.Value of the struct.
		// TItem itself is the struct type here (e.g. MyCondition).
		val := reflect.ValueOf(newList[i]) // This is MyCondition (not *MyCondition)

		// No pointer check needed here for `val` because TItem is the struct type itself.
		if val.Kind() != reflect.Struct {
			// Should not happen if TItem is a struct type.
			// This check is more for if TItem could be non-struct.
			continue
		}

		fieldStrValue, ok := getReflectedJSONFieldValueAsString(val, keyName)
		if ok && fieldStrValue == keyValue {
			foundNewElementPtr = &newList[i] // Take address of element in slice
			break
		}
	}

	for i := range oldList {
		val := reflect.ValueOf(oldList[i])
		if val.Kind() != reflect.Struct {
			continue
		}
		fieldStrValue, ok := getReflectedJSONFieldValueAsString(val, keyName)
		if ok && fieldStrValue == keyValue {
			foundOldElementPtr = &oldList[i] // Take address of element in slice
			break
		}
	}

	// Only proceed if at least one of them was found (and thus its pointer is non-nil)
	if foundNewElementPtr != nil || foundOldElementPtr != nil {
		elementPath := fldPath.Key(keyValue)
		// If one is nil (not found), elementValidator gets a nil pointer, which is standard.
		errs = append(errs, elementValidator(ctx, op, elementPath, foundNewElementPtr, foundOldElementPtr)...)
	}
	return errs
}

// getReflectedJSONFieldValueAsString gets the string value of a field `jsonKeyName` from a struct `sVal`.
// sVal must be a reflect.Value of Kind reflect.Struct.
func getReflectedJSONFieldValueAsString(sVal reflect.Value, jsonKeyName string) (string, bool) {
	if sVal.Kind() != reflect.Struct {
		return "", false
	}
	typ := sVal.Type()
	for i := 0; i < typ.NumField(); i++ {
		fieldDesc := typ.Field(i)
		if fieldDesc.PkgPath != "" { // Skip unexported fields
			continue
		}

		tag := fieldDesc.Tag.Get("json")
		jsonTagParts := strings.Split(tag, ",")
		nameInTag := jsonTagParts[0]

		currentJsonKeyName := ""
		if nameInTag != "" && nameInTag != "-" {
			currentJsonKeyName = nameInTag
		} else if nameInTag == "-" {
			continue // Field is ignored by JSON
		} else {
			// If no explicit json name, and not ignored, it defaults to field name (case-sensitive)
			// or for K8s usually camelCase from Go field name.
			// For simplicity here, if the tag is `json:""` or `json:",omitempty"`
			// we assume the field name itself might be used or it's an error in tagging.
			// The original code just `continue`d, implying exact match on `jsonKeyName` is needed.
			// Let's stick to that: if `nameInTag` is empty, it won't match `jsonKeyName` unless `jsonKeyName` is also empty.
			continue
		}

		if currentJsonKeyName == jsonKeyName {
			fieldValue := sVal.Field(i)
			if fieldValue.CanInterface() {
				// Prefer direct string conversion if the field is a string or string alias.
				if fieldValue.Kind() == reflect.String {
					return fieldValue.String(), true
				}
				// Handle types that are convertible to string (e.g. type MyString string)
				if fieldValue.Type().Comparable() && fieldValue.Type().ConvertibleTo(reflect.TypeOf("")) {
					return fieldValue.Convert(reflect.TypeOf("")).String(), true
				}
				// As a fallback, try fmt.Stringer interface, but be cautious as this might not be the "key" value.
				// For the specific case of "type":"Denied", type is usually string or string alias.
				// if stringer, ok := fieldValue.Interface().(fmt.Stringer); ok {
				//  return stringer.String(), true
				// }
			}
			// If not a string or directly convertible, or cannot interface, cannot get string value simply.
			return "", false
		}
	}
	return "", false // Field not found
}
