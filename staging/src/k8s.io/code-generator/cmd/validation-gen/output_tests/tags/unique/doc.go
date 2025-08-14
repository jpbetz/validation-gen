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

// This is a test package.
package unique

import "k8s.io/code-generator/cmd/validation-gen/testscheme"

var localSchemeBuilder = testscheme.New()

type Struct struct {
	TypeMeta int

	// Basic unique=set on primitive slice
	// +k8s:unique=set
	SliceSetField []string `json:"sliceSetField"`

	// Basic unique=map with single key
	// +k8s:unique=map
	// +k8s:listMapKey=key
	SliceMapField []Item `json:"sliceMapField"`

	// Basic unique=set on struct slice
	// +k8s:unique=set
	SliceSetFieldWithStruct []Item `json:"sliceSetFieldWithStruct"`

	// unique=map with multiple keys
	// +k8s:unique=map
	// +k8s:listMapKey=key1
	// +k8s:listMapKey=key2
	SliceMapFieldWithMultipleKeys []ItemWithMultipleKeys `json:"sliceMapFieldWithMultipleKeys"`

	// atomic + unique=set combination
	// +k8s:listType=atomic
	// +k8s:unique=set
	AtomicListUniqueSet []Item `json:"atomicListUniqueSet"`

	// atomic + unique=map combination
	// +k8s:listType=atomic
	// +k8s:unique=map
	// +k8s:listMapKey=key
	AtomicListUniqueMap []Item `json:"atomicListUniqueMap"`

	// Test with primitive types that support direct comparison
	// +k8s:unique=set
	IntSlice []int `json:"intSlice"`

	// Test with primitive types that support direct comparison
	// +k8s:unique=set
	BoolSlice []bool `json:"boolSlice"`

	// Test with zero values
	// +k8s:unique=set
	SliceWithZeroValues []string `json:"sliceWithZeroValues"`

	// Test with empty string keys
	// +k8s:unique=map
	// +k8s:listMapKey=key
	SliceWithEmptyKeys []Item `json:"sliceWithEmptyKeys"`
}

type Item struct {
	Key  string `json:"key"`
	Data string `json:"data"`
}

type ItemWithMultipleKeys struct {
	Key1 string `json:"key1"`
	Key2 string `json:"key2"`
	Data string `json:"data"`
}
