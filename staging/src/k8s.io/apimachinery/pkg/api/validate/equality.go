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

	"k8s.io/apimachinery/pkg/api/operation"
	"k8s.io/apimachinery/pkg/api/validate/content"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// NEQString validates that the specified value is not equal to the disallowed string.
func NEQString[T ~string](_ context.Context, _ operation.Operation, fldPath *field.Path, value, _ *T, disallowed string) field.ErrorList {
	if value == nil {
		return nil
	}
	if string(*value) == disallowed {
		return field.ErrorList{
			field.Invalid(fldPath, *value, content.NEQStringError(disallowed)).WithOrigin("NEQString"),
		}
	}
	return nil
}
