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

package validators

import (
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/gengo/v2/codetags"
	"k8s.io/gengo/v2/types"
)

type validatorErrorTestCase struct {
	name       string
	tagName    string
	tagValue   string
	targetType *types.Type
	scope      Scope // defaults to ScopeField if empty
	expectErr  string
}

var (
	// Helper types for testing
	strType        = types.String
	intType        = &types.Type{Kind: types.Builtin, Name: types.Name{Name: "int"}}
	structType     = &types.Type{Kind: types.Struct, Name: types.Name{Name: "MyStruct"}}
	sliceOfStrings = &types.Type{Kind: types.Slice, Elem: strType}
	mapStringInt   = &types.Type{Kind: types.Map, Key: strType, Elem: intType}
)

func TestValidatorErrors(t *testing.T) {

	tests := []validatorErrorTestCase{
		// Required/Optional/Forbidden
		{
			name:       "required on plain struct",
			tagName:    "k8s:required",
			targetType: structType,
			expectErr:  `non-pointer structs cannot use the "k8s:required" tag`,
		},
		{
			name:       "optional on plain struct",
			tagName:    "k8s:optional",
			targetType: structType,
			expectErr:  `non-pointer structs cannot use the "k8s:optional" tag`,
		},
		{
			name:       "forbidden on plain struct",
			tagName:    "k8s:forbidden",
			targetType: structType,
			expectErr:  `non-pointer structs cannot use the "k8s:forbidden" tag`,
		},

		// Enum
		{
			name:       "enum on non-string",
			tagName:    "k8s:enum",
			targetType: intType,
			scope:      ScopeType,
			expectErr:  `can only be used on string types`,
		},

		// List Types (listType, listMapKey, unique, customUnique)
		{
			name:       "listType on non-list",
			tagName:    "k8s:listType",
			tagValue:   "atomic",
			targetType: structType,
			expectErr:  `can only be used on list types`,
		},
		{
			name:       "listType unknown value",
			tagName:    "k8s:listType",
			tagValue:   "unknown",
			targetType: sliceOfStrings,
			expectErr:  `unknown list type "unknown"`,
		},
		{
			name:       "listType=map on non-struct slice",
			tagName:    "k8s:listType",
			tagValue:   "map",
			targetType: sliceOfStrings,
			expectErr:  `only lists of structs can be list-maps`,
		},
		{
			name:       "listMapKey on non-list",
			tagName:    "k8s:listMapKey",
			tagValue:   "someField",
			targetType: structType,
			expectErr:  `can only be used on list types`,
		},
		{
			name:       "listMapKey on non-struct slice",
			tagName:    "k8s:listMapKey",
			tagValue:   "someField",
			targetType: sliceOfStrings,
			expectErr:  `only lists of structs can be list-maps`,
		},
		{
			name:       "unique on non-list",
			tagName:    "k8s:unique",
			tagValue:   "set",
			targetType: structType,
			expectErr:  `can only be used on list types`,
		},
		{
			name:       "unique=map on non-struct slice",
			tagName:    "k8s:unique",
			tagValue:   "map",
			targetType: sliceOfStrings,
			expectErr:  `only lists of structs can be list-maps`,
		},
		{
			name:       "unique unknown value",
			tagName:    "k8s:unique",
			tagValue:   "unknown",
			targetType: sliceOfStrings,
			expectErr:  `unknown unique type "unknown"`,
		},
		{
			name:       "customUnique on non-list",
			tagName:    "k8s:customUnique",
			targetType: structType,
			expectErr:  `can only be used on list types`,
		},

		// Limits (min/max/length)
		{
			name:       "maxLength on non-string",
			tagName:    "k8s:maxLength",
			tagValue:   "10",
			targetType: intType,
			expectErr:  `can only be used on string types`,
		},
		{
			name:       "maxLength negative",
			tagName:    "k8s:maxLength",
			tagValue:   "-1",
			targetType: strType,
			expectErr:  `must be greater than or equal to zero`,
		},
		{
			name:       "maxItems on non-list",
			tagName:    "k8s:maxItems",
			tagValue:   "10",
			targetType: structType,
			expectErr:  `can only be used on list types`,
		},
		{
			name:       "maxItems negative",
			tagName:    "k8s:maxItems",
			tagValue:   "-1",
			targetType: sliceOfStrings,
			expectErr:  `must be greater than or equal to zero`,
		},
		{
			name:       "maxProperties negative",
			tagName:    "k8s:maxProperties",
			tagValue:   "-1",
			targetType: mapStringInt,
			expectErr:  `must be greater than or equal to zero`,
		},
		{
			name:       "minimum on non-integer",
			tagName:    "k8s:minimum",
			tagValue:   "10",
			targetType: strType,
			expectErr:  `can only be used on integer types`,
		},
		{
			name:       "minLength on non-string",
			tagName:    "k8s:minLength",
			tagValue:   "1",
			targetType: intType,
			expectErr:  `can only be used on string types`,
		},
		{
			name:       "minLength zero",
			tagName:    "k8s:minLength",
			tagValue:   "0",
			targetType: strType,
			expectErr:  `must be greater than or equal to one`,
		},

		// Format
		{
			name:       "format on non-string",
			tagName:    "k8s:format",
			tagValue:   "k8s-uuid",
			targetType: intType,
			expectErr:  `can only be used on string types`,
		},
		{
			name:       "format unknown",
			tagName:    "k8s:format",
			tagValue:   "unknown",
			targetType: strType,
			expectErr:  `unsupported validation format "unknown"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runValidatorErrorTest(t, tt)
		})
	}
}

func runValidatorErrorTest(t *testing.T, tt validatorErrorTestCase) {
	for k := range globalListMeta {
		delete(globalListMeta, k)
	}

	val := globalRegistry.tagValidators[tt.tagName]
	if val == nil {
		t.Fatalf("Common validator %s not found. Registered validators: %v", tt.tagName, globalRegistry.tagIndex)
	}

	scope := tt.scope
	if scope == "" {
		scope = ScopeField
	}

	ctx := Context{
		Type:  tt.targetType,
		Scope: scope,
		Member: &types.Member{
			Name: "DummyMember",
			Type: tt.targetType,
		},
		Path: field.NewPath("testPath"),
	}
	tag := codetags.Tag{
		Name:  tt.tagName,
		Value: tt.tagValue,
	}

	_, err := val.GetValidations(ctx, tag)
	if len(tt.expectErr) > 0 {
		if err == nil {
			t.Errorf("Expected error containing %q, got nil", tt.expectErr)
		} else if !strings.Contains(err.Error(), tt.expectErr) {
			t.Errorf("Expected error containing %q, got %q", tt.expectErr, err.Error())
		}
	} else {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}
}
