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

package validators

import (
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/gengo/v2/codetags"
	"k8s.io/gengo/v2/types"
)

func TestModeTagValidator(t *testing.T) {
	stringType := types.String
	intType := types.Int
	uint32Type := types.Uint32
	boolType := types.Bool
	structType := &types.Type{Name: types.Name{Name: "Foo"}, Kind: types.Struct}
	ptrType := &types.Type{Elem: stringType, Kind: types.Pointer}

	mtv := &modeTagValidator{shared: make(map[string]modeGroups)}

	tests := []struct {
		name        string
		fieldType   *types.Type
		tag         codetags.Tag
		wantErr     string
		expectGroup string
		setup       func()
	}{
		{
			name:        "valid string mode",
			fieldType:   stringType,
			tag:         codetags.Tag{},
			expectGroup: "default",
		},
		{
			name:        "valid int mode",
			fieldType:   intType,
			tag:         codetags.Tag{},
			expectGroup: "default",
		},
		{
			name:        "valid uint32 mode",
			fieldType:   uint32Type,
			tag:         codetags.Tag{},
			expectGroup: "default",
		},
		{
			name:        "valid bool mode",
			fieldType:   boolType,
			tag:         codetags.Tag{},
			expectGroup: "default",
		},
		{
			name:        "named string mode",
			fieldType:   stringType,
			tag:         codetags.Tag{Args: []codetags.Arg{{Name: "name", Value: "MyMode"}}},
			expectGroup: "MyMode",
		},
		{
			name:      "invalid type struct",
			fieldType: structType,
			tag:       codetags.Tag{},
			wantErr:   "can only be used on string, bool or integer types",
		},
		{
			name:      "invalid type pointer",
			fieldType: ptrType,
			tag:       codetags.Tag{},
			wantErr:   "can only be used on non-pointer types",
		},
		{
			name:      "duplicate mode discriminator",
			fieldType: stringType,
			tag:       codetags.Tag{},
			wantErr:   "duplicate discriminator: \"\"",
			setup: func() {
				mtv.shared["MyStruct"] = make(modeGroups)
				group := mtv.shared["MyStruct"].getOrCreate("")
				group.discriminatorMember = &types.Member{Name: "OtherField"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mtv.shared = make(map[string]modeGroups) // reset
			if tt.setup != nil {
				tt.setup()
			}
			ctx := Context{
				Type:       tt.fieldType,
				Member:     &types.Member{Name: "ModeField", Type: tt.fieldType},
				ParentPath: field.NewPath("MyStruct"),
			}
			_, err := mtv.GetValidations(ctx, tt.tag)

			if (err != nil) != (tt.wantErr != "") {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
			if err != nil && !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}

			if err == nil {
				group := mtv.shared["MyStruct"][tt.expectGroup]
				if group == nil {
					t.Fatalf("expected mode group %q to be created", tt.expectGroup)
				}
				if group.discriminatorMember != ctx.Member {
					t.Fatalf("expected discriminator member to be set")
				}
			}
		})
	}
}

// A mock registry to prevent panic in ExtractTagValidations for tests
type mockValidator struct {
	Validator
}

func (m mockValidator) ExtractTagValidations(context Context, tags ...codetags.Tag) (Validations, error) {
	return Validations{}, nil
}

func TestMemberTagValidator(t *testing.T) {
	tests := []struct {
		name    string
		tag     codetags.Tag
		wantErr string
	}{
		{
			name:    "missing payload",
			tag:     codetags.Tag{Args: []codetags.Arg{{Value: "Val"}}},
			wantErr: "missing required payload",
		},
		{
			name: "missing value",
			tag: codetags.Tag{
				ValueTag: &codetags.Tag{Name: requiredTagName},
			},
			wantErr: "missing required value",
		},
		{
			name: "disallowed payload",
			tag: codetags.Tag{
				Args:     []codetags.Arg{{Value: "Val"}},
				ValueTag: &codetags.Tag{Name: listTypeTagName},
			},
			wantErr: "unsupported payload tag",
		},

		{
			name: "valid member with positional value",
			tag: codetags.Tag{
				Args:     []codetags.Arg{{Value: "A"}},
				ValueTag: &codetags.Tag{Name: requiredTagName},
			},
		},
		{
			name: "valid member with named value",
			tag: codetags.Tag{
				Args:     []codetags.Arg{{Name: "value", Value: "A"}},
				ValueTag: &codetags.Tag{Name: requiredTagName},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mtv := &memberTagValidator{
				shared:    make(map[string]modeGroups),
				validator: &mockValidator{},
			}
			ctx := Context{
				Member:     &types.Member{Name: "MyField"},
				ParentPath: field.NewPath("MyStruct"),
			}
			_, err := mtv.GetValidations(ctx, tt.tag)

			if (err != nil) != (tt.wantErr != "") {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
			if err != nil && !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestModeTypeOrFieldValidator(t *testing.T) {
	structType := &types.Type{Name: types.Name{Name: "Foo"}, Kind: types.Struct}
	stringType := types.String

	// Create a valid struct setup
	validGroup := &modeGroup{
		name:                "default",
		discriminatorMember: &types.Member{Name: "Mode", Type: stringType},
		members: map[string]*fieldModeRules{
			"FieldA": {
				member: &types.Member{Name: "FieldA", Type: stringType},
				rules: []modeRule{
					{value: "A", validations: Validations{Functions: []FunctionGen{Function("test", DefaultFlags, types.Name{Name: "Fake"})}}},
				},
			},
		},
	}

	// Create a missing discriminator struct setup
	missingDiscrimGroup := &modeGroup{
		name: "default",
		members: map[string]*fieldModeRules{
			"FieldA": {
				member: &types.Member{Name: "FieldA", Type: stringType},
				rules:  []modeRule{{value: "A", validations: Validations{}}},
			},
		},
	}

	tests := []struct {
		name    string
		groups  map[string]modeGroups
		wantErr string
	}{
		{
			name: "valid group emits functions",
			groups: map[string]modeGroups{
				"MyStruct": {"default": validGroup},
			},
		},
		{
			name: "missing discriminator field",
			groups: map[string]modeGroups{
				"MyStruct": {"default": missingDiscrimGroup},
			},
			wantErr: "missing discriminator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mtfv := &modeTypeOrFieldValidator{shared: tt.groups}
			ctx := Context{
				Scope: ScopeType,
				Type:  structType,
				Path:  field.NewPath("MyStruct"),
			}

			validations, err := mtfv.GetValidations(ctx)

			if (err != nil) != (tt.wantErr != "") {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
			if err != nil && !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}

			if err == nil && len(tt.groups) > 0 && len(validations.Functions) == 0 {
				t.Fatalf("expected validations to be emitted")
			}
		})
	}
}
