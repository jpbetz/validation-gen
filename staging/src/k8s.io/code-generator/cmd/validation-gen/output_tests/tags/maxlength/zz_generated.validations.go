//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright The Kubernetes Authors.

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

// Code generated by validation-gen. DO NOT EDIT.

package maxlength

import (
	context "context"
	fmt "fmt"

	operation "k8s.io/apimachinery/pkg/api/operation"
	safe "k8s.io/apimachinery/pkg/api/safe"
	validate "k8s.io/apimachinery/pkg/api/validate"
	field "k8s.io/apimachinery/pkg/util/validation/field"
	testscheme "k8s.io/code-generator/cmd/validation-gen/testscheme"
)

func init() { localSchemeBuilder.Register(RegisterValidations) }

// RegisterValidations adds validation functions to the given scheme.
// Public to allow building arbitrary schemes.
func RegisterValidations(scheme *testscheme.Scheme) error {
	scheme.AddValidationFunc((*Struct)(nil), func(ctx context.Context, op operation.Operation, obj, oldObj interface{}) field.ErrorList {
		switch op.Request.SubresourcePath() {
		case "/":
			return Validate_Struct(ctx, op, nil /* fldPath */, obj.(*Struct), safe.Cast[*Struct](oldObj))
		}
		return field.ErrorList{field.InternalError(nil, fmt.Errorf("no validation found for %T, subresource: %v", obj, op.Request.SubresourcePath()))}
	})
	return nil
}

func Validate_Max0Type(ctx context.Context, op operation.Operation, fldPath *field.Path, obj, oldObj *Max0Type) (errs field.ErrorList) {
	// type Max0Type
	if op.Type == operation.Update && (obj == oldObj || (obj != nil && oldObj != nil && *obj == *oldObj)) {
		return nil // no changes
	}
	errs = append(errs, validate.MaxLength(ctx, op, fldPath, obj, oldObj, 0)...)

	return errs
}

func Validate_Max10Type(ctx context.Context, op operation.Operation, fldPath *field.Path, obj, oldObj *Max10Type) (errs field.ErrorList) {
	// type Max10Type
	if op.Type == operation.Update && (obj == oldObj || (obj != nil && oldObj != nil && *obj == *oldObj)) {
		return nil // no changes
	}
	errs = append(errs, validate.MaxLength(ctx, op, fldPath, obj, oldObj, 10)...)

	return errs
}

func Validate_Struct(ctx context.Context, op operation.Operation, fldPath *field.Path, obj, oldObj *Struct) (errs field.ErrorList) {
	// field Struct.TypeMeta has no validation

	// field Struct.Max0Field
	errs = append(errs,
		func(fldPath *field.Path, obj, oldObj *string) (errs field.ErrorList) {
			if op.Type == operation.Update && (obj == oldObj || (obj != nil && oldObj != nil && *obj == *oldObj)) {
				return nil // no changes
			}
			errs = append(errs, validate.MaxLength(ctx, op, fldPath, obj, oldObj, 0)...)
			return
		}(fldPath.Child("max0Field"), &obj.Max0Field, safe.Field(oldObj, func(oldObj *Struct) *string { return &oldObj.Max0Field }))...)

	// field Struct.Max0PtrField
	errs = append(errs,
		func(fldPath *field.Path, obj, oldObj *string) (errs field.ErrorList) {
			if op.Type == operation.Update && (obj == oldObj || (obj != nil && oldObj != nil && *obj == *oldObj)) {
				return nil // no changes
			}
			errs = append(errs, validate.MaxLength(ctx, op, fldPath, obj, oldObj, 0)...)
			return
		}(fldPath.Child("max0PtrField"), obj.Max0PtrField, safe.Field(oldObj, func(oldObj *Struct) *string { return oldObj.Max0PtrField }))...)

	// field Struct.Max10Field
	errs = append(errs,
		func(fldPath *field.Path, obj, oldObj *string) (errs field.ErrorList) {
			if op.Type == operation.Update && (obj == oldObj || (obj != nil && oldObj != nil && *obj == *oldObj)) {
				return nil // no changes
			}
			errs = append(errs, validate.MaxLength(ctx, op, fldPath, obj, oldObj, 10)...)
			return
		}(fldPath.Child("max10Field"), &obj.Max10Field, safe.Field(oldObj, func(oldObj *Struct) *string { return &oldObj.Max10Field }))...)

	// field Struct.Max10PtrField
	errs = append(errs,
		func(fldPath *field.Path, obj, oldObj *string) (errs field.ErrorList) {
			if op.Type == operation.Update && (obj == oldObj || (obj != nil && oldObj != nil && *obj == *oldObj)) {
				return nil // no changes
			}
			errs = append(errs, validate.MaxLength(ctx, op, fldPath, obj, oldObj, 10)...)
			return
		}(fldPath.Child("max10PtrField"), obj.Max10PtrField, safe.Field(oldObj, func(oldObj *Struct) *string { return oldObj.Max10PtrField }))...)

	// field Struct.Max0UnvalidatedTypedefField
	errs = append(errs,
		func(fldPath *field.Path, obj, oldObj *UnvalidatedStringType) (errs field.ErrorList) {
			if op.Type == operation.Update && (obj == oldObj || (obj != nil && oldObj != nil && *obj == *oldObj)) {
				return nil // no changes
			}
			errs = append(errs, validate.MaxLength(ctx, op, fldPath, obj, oldObj, 0)...)
			return
		}(fldPath.Child("max0UnvalidatedTypedefField"), &obj.Max0UnvalidatedTypedefField, safe.Field(oldObj, func(oldObj *Struct) *UnvalidatedStringType { return &oldObj.Max0UnvalidatedTypedefField }))...)

	// field Struct.Max0UnvalidatedTypedefPtrField
	errs = append(errs,
		func(fldPath *field.Path, obj, oldObj *UnvalidatedStringType) (errs field.ErrorList) {
			if op.Type == operation.Update && (obj == oldObj || (obj != nil && oldObj != nil && *obj == *oldObj)) {
				return nil // no changes
			}
			errs = append(errs, validate.MaxLength(ctx, op, fldPath, obj, oldObj, 0)...)
			return
		}(fldPath.Child("max0UnvalidatedTypedefPtrField"), obj.Max0UnvalidatedTypedefPtrField, safe.Field(oldObj, func(oldObj *Struct) *UnvalidatedStringType { return oldObj.Max0UnvalidatedTypedefPtrField }))...)

	// field Struct.Max10UnvalidatedTypedefField
	errs = append(errs,
		func(fldPath *field.Path, obj, oldObj *UnvalidatedStringType) (errs field.ErrorList) {
			if op.Type == operation.Update && (obj == oldObj || (obj != nil && oldObj != nil && *obj == *oldObj)) {
				return nil // no changes
			}
			errs = append(errs, validate.MaxLength(ctx, op, fldPath, obj, oldObj, 10)...)
			return
		}(fldPath.Child("max10UnvalidatedTypedefField"), &obj.Max10UnvalidatedTypedefField, safe.Field(oldObj, func(oldObj *Struct) *UnvalidatedStringType { return &oldObj.Max10UnvalidatedTypedefField }))...)

	// field Struct.Max10UnvalidatedTypedefPtrField
	errs = append(errs,
		func(fldPath *field.Path, obj, oldObj *UnvalidatedStringType) (errs field.ErrorList) {
			if op.Type == operation.Update && (obj == oldObj || (obj != nil && oldObj != nil && *obj == *oldObj)) {
				return nil // no changes
			}
			errs = append(errs, validate.MaxLength(ctx, op, fldPath, obj, oldObj, 10)...)
			return
		}(fldPath.Child("max10UnvalidatedTypedefPtrField"), obj.Max10UnvalidatedTypedefPtrField, safe.Field(oldObj, func(oldObj *Struct) *UnvalidatedStringType { return oldObj.Max10UnvalidatedTypedefPtrField }))...)

	// field Struct.Max0ValidatedTypedefField
	errs = append(errs,
		func(fldPath *field.Path, obj, oldObj *Max0Type) (errs field.ErrorList) {
			errs = append(errs, Validate_Max0Type(ctx, op, fldPath, obj, oldObj)...)
			return
		}(fldPath.Child("max0ValidatedTypedefField"), &obj.Max0ValidatedTypedefField, safe.Field(oldObj, func(oldObj *Struct) *Max0Type { return &oldObj.Max0ValidatedTypedefField }))...)

	// field Struct.Max0ValidatedTypedefPtrField
	errs = append(errs,
		func(fldPath *field.Path, obj, oldObj *Max0Type) (errs field.ErrorList) {
			errs = append(errs, Validate_Max0Type(ctx, op, fldPath, obj, oldObj)...)
			return
		}(fldPath.Child("max0ValidatedTypedefPtrField"), obj.Max0ValidatedTypedefPtrField, safe.Field(oldObj, func(oldObj *Struct) *Max0Type { return oldObj.Max0ValidatedTypedefPtrField }))...)

	// field Struct.Max10ValidatedTypedefField
	errs = append(errs,
		func(fldPath *field.Path, obj, oldObj *Max10Type) (errs field.ErrorList) {
			errs = append(errs, Validate_Max10Type(ctx, op, fldPath, obj, oldObj)...)
			return
		}(fldPath.Child("max10ValidatedTypedefField"), &obj.Max10ValidatedTypedefField, safe.Field(oldObj, func(oldObj *Struct) *Max10Type { return &oldObj.Max10ValidatedTypedefField }))...)

	// field Struct.Max10ValidatedTypedefPtrField
	errs = append(errs,
		func(fldPath *field.Path, obj, oldObj *Max10Type) (errs field.ErrorList) {
			errs = append(errs, Validate_Max10Type(ctx, op, fldPath, obj, oldObj)...)
			return
		}(fldPath.Child("max10ValidatedTypedefPtrField"), obj.Max10ValidatedTypedefPtrField, safe.Field(oldObj, func(oldObj *Struct) *Max10Type { return oldObj.Max10ValidatedTypedefPtrField }))...)

	return errs
}
