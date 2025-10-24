/*
Copyright 2025 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*/

// +k8s:validation-gen=TypeMeta
// +k8s:validation-gen-scheme-registry=k8s.io/code-generator/cmd/validation-gen/testscheme.Scheme

// This is a test package.
package typedeftoslice

import "k8s.io/code-generator/cmd/validation-gen/testscheme"

var localSchemeBuilder = testscheme.New()

// Note: no validation here
type UnvalidatedType []int

// +k8s:minItems=1
type Min1Type []int

// +k8s:minItems=5
type Min5Type []int

// Note: no validation here
type UnvalidatedPtrType []*int

type SliceType []int

// +k8s:minItems=1
type Min1TypedefType SliceType

// +k8s:minItems=5
type Min5TypedefType SliceType

type Struct struct {
	TypeMeta int

	UnvalidatedField UnvalidatedType `json:"unvalidatedField"`

	Min1Field Min1Type `json:"min1Field"`

	Min5Field Min5Type `json:"min5Field"`

	Min1TypedefField Min1TypedefType `json:"min1TypedefField"`

	Min5TypedefField Min5TypedefType `json:"min5TypedefField"`
}
