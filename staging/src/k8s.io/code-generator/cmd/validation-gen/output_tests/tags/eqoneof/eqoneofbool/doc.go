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

// +k8s:validation-gen=TypeMeta
// +k8s:validation-gen-scheme-registry=k8s.io/code-generator/cmd/validation-gen/testscheme.Scheme

// Package eqoneofbool tests boolean validation for eqOneOf tag.
package eqoneofbool

import "k8s.io/code-generator/cmd/validation-gen/testscheme"

var localSchemeBuilder = testscheme.New()

type Struct struct {
	TypeMeta int `json:"typeMeta"`

	// +k8s:eqOneOf=`[true]`
	BoolOnlyTrueField bool `json:"boolOnlyTrueField"`

	// +k8s:eqOneOf=`[false]`
	BoolPtrOnlyFalseField *bool `json:"boolPtrOnlyFalseField"`

	// +k8s:eqOneOf=`[true,false]`
	BoolBothAllowedField bool `json:"boolBothAllowedField"`

	ValidatedTypedefBoolField ValidatedBoolType `json:"validatedTypedefBoolField"`
}

// +k8s:eqOneOf=`[false]`
type ValidatedBoolType bool
