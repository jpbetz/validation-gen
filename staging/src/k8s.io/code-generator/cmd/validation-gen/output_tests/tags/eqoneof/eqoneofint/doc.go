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

// Package eqoneofint tests integer validation for eqOneOf tag.
package eqoneofint

import "k8s.io/code-generator/cmd/validation-gen/testscheme"

var localSchemeBuilder = testscheme.New()

type Struct struct {
	TypeMeta int `json:"typeMeta"`

	// +k8s:eqOneOf=`[1,2,3]`
	IntField int `json:"intField"`
	// +k8s:eqOneOf=`[1,2,3]`
	IntPtrField *int `json:"intPtrField"`

	// +k8s:eqOneOf=`[1,2,3]`
	Int16Field int16 `json:"int16Field"`
	// +k8s:eqOneOf=`[1,2,3]`
	Int32Field int32 `json:"int32Field"`
	// +k8s:eqOneOf=`[1,2,3]`
	Int64Field int64 `json:"int64Field"`

	// +k8s:eqOneOf=`[1,2,3]`
	UintField uint `json:"uintField"`
	// +k8s:eqOneOf=`[1,2,3]`
	UintPtrField *uint `json:"uintPtrField"`

	// +k8s:eqOneOf=`[1,2,3]`
	Uint16Field uint16 `json:"uint16Field"`
	// +k8s:eqOneOf=`[1,2,3]`
	Uint32Field uint32 `json:"uint32Field"`
	// +k8s:eqOneOf=`[1,2,3]`
	Uint64Field uint64 `json:"uint64Field"`

	// +k8s:eqOneOf=`[-1,-2,-3]`
	IntWithNegativeField int `json:"intWithNegativeField"`

	// +k8s:eqOneOf=`[1,2,3]`
	TypedefField IntType `json:"typedefField"`
	// +k8s:eqOneOf=`[1,2,3]`
	TypedefPtrField *IntType `json:"typedefPtrField"`

	ValidatedTypedefIntField ValidatedIntType `json:"validatedTypedefIntField"`
}

type IntType int

// +k8s:eqOneOf=`[1,2,3]`
type ValidatedIntType int
