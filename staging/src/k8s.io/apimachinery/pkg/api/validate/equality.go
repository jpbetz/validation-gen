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

	"k8s.io/apimachinery/pkg/api/operation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// EqOneOf validates that the specified comparable value is equal to one of the allowed values.
func EqOneOf[T comparable](_ context.Context, _ operation.Operation, fldPath *field.Path, value, _ *T, allowed []T) field.ErrorList {
	if value == nil {
		return nil
	}
	for _, a := range allowed {
		if *value == a {
			return nil
		}
	}
	// Convert allowed values to strings for NotSupported fn.
	allowedStrs := make([]string, len(allowed))
	for i, v := range allowed {
		allowedStrs[i] = fmt.Sprintf("%v", v)
	}
	return field.ErrorList{
		field.NotSupported(fldPath, fmt.Sprintf("%v", *value), allowedStrs).WithOrigin("k8s:eqOneOf"),
	}
}
