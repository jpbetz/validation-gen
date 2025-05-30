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
package special_keys

import "k8s.io/code-generator/cmd/validation-gen/testscheme"

var localSchemeBuilder = testscheme.New()

type Struct struct {
	TypeMeta int `json:"typeMeta"`

	// Empty string key
	// +k8s:listType=map
	// +k8s:listMapKey=key
	// +k8s:listMapItem(`[["key",""]]`)=+k8s:validateFalse="listMapItem EmptyKey[key=]"
	EmptyKey []Item `json:"emptyKey"`

	// Special characters
	// +k8s:listType=map
	// +k8s:listMapKey=key
	// +k8s:listMapItem(`[["key","with\"quotes"]]`)=+k8s:validateFalse="listMapItem Special[key=with\"quotes]"
	// +k8s:listMapItem(`[["key","multi\nline"]]`)=+k8s:validateFalse="listMapItem Special[key=multi\nline]"
	// +k8s:listMapItem(`[["key","unicode-🚀"]]`)=+k8s:validateFalse="listMapItem Special[key=unicode-🚀]"
	Special []Item `json:"special"`
}

type Item struct {
	Key  string `json:"key"`
	Data string `json:"data"`
}
