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

// +k8s:validation-gen=TypeMeta
// +k8s:validation-gen-scheme-registry=k8s.io/code-generator/cmd/validation-gen/testscheme.Scheme

// This is a test package.
package immutable

import "k8s.io/code-generator/cmd/validation-gen/testscheme"

var localSchemeBuilder = testscheme.New()

type Struct struct {
	TypeMeta int

	// +k8s:immutable
	StringField string `json:"stringField"`

	// +k8s:immutable
	StringPtrField *string `json:"stringPtrField"`

	// +k8s:immutable
	StructField ComparableStruct `json:"structField"`

	// +k8s:immutable
	StructPtrField *ComparableStruct `json:"structPtrField"`

	// +k8s:immutable
	NonComparableStructField NonComparableStruct `json:"noncomparableStructField"`

	// +k8s:immutable
	NonComparableStructPtrField *NonComparableStruct `json:"noncomparableStructPtrField"`

	// +k8s:immutable
	SliceField []string `json:"sliceField"`

	// +k8s:immutable
	MapField map[string]string `json:"mapField"`

	ImmutableField ImmutableType `json:"immutableField"`

	ImmutablePtrField *ImmutableType `json:"immutablePtrField"`

	// +k8s:immutable
	IntPtrField *int `json:"intPtrField"`

	// +k8s:immutable
	BoolPtrField *bool `json:"boolPtrField"`

	// +k8s:optional
	// +default="defaultValue"
	// +k8s:immutable
	StringWithDefault string `json:"stringWithDefault"`

	// +k8s:optional
	// +default=42
	// +k8s:immutable
	IntPtrWithDefault *int32 `json:"intPtrWithDefault"`

	// +k8s:optional
	// +default=0
	// +k8s:immutable
	IntWithZeroDefault int32 `json:"intWithZeroDefault"`

	// +k8s:optional
	// +default=""
	// +k8s:immutable
	StringWithZeroDefault string `json:"stringWithZeroDefault"`

	// +k8s:required
	// +k8s:immutable
	RequiredImmutableField string `json:"requiredImmutableField"`

	// +k8s:optional
	// +k8s:immutable
	OptionalImmutableField *string `json:"optionalImmutableField"`
}

type ComparableStruct struct {
	StringField    string  `json:"stringField"`
	StringPtrField *string `json:"stringPtrField"`
}

type NonComparableStruct struct {
	SliceField []string `json:"sliceField"`
}

// +k8s:immutable
type ImmutableType string
