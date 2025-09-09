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
package grouping

import "k8s.io/code-generator/cmd/validation-gen/testscheme"

var localSchemeBuilder = testscheme.New()

type Struct struct {
	TypeMeta int

	// +k8s:ifEnabled(FeatureX)=+k8s:listType=atomic
	// +k8s:ifEnabled(FeatureX)=+k8s:listMapKey=key
	// +k8s:ifEnabled(FeatureX)=+k8s:unique=map
	XEnabledListUniqueField []Item `json:"listMapField"`

	// +k8s:ifEnabled(FeatureA)=+k8s:validateFalse="field Struct.PrimitiveField enabled validation 1"
	// +k8s:ifEnabled(FeatureA)=+k8s:validateFalse="field Struct.PrimitiveField enabled validation 2"
	// +k8s:ifDisabled(FeatureA)=+k8s:validateFalse="field Struct.PrimitiveField disabled validation 1"
	// +k8s:ifDisabled(FeatureA)=+k8s:validateFalse="field Struct.PrimitiveField disabled validation 2"
	// +k8s:ifEnabled(FeatureB)=+k8s:validateFalse="field Struct.PrimitiveField feature B"
	PrimitiveField string `json:"primitiveField"`
}

type Item struct {
	Key   string `json:"key"`
	Value int    `json:"value"`
}
