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
package simple

import "k8s.io/code-generator/cmd/validation-gen/testscheme"

var localSchemeBuilder = testscheme.New()

type Struct struct {
	TypeMeta int `json:"typeMeta"`

	// +k8s:listType=map
	// +k8s:listMapKey=key
	// +k8s:listMapItem(`[["key","target"]]`)=+k8s:validateFalse="listMapItem SingleKey[key=target]"
	// +k8s:listMapItem(`[["key","fixed"]]`)=+k8s:immutable
	SingleKey []Item `json:"singleKey"`

	// +k8s:listType=map
	// +k8s:listMapKey=key1
	// +k8s:listMapKey=key2
	// +k8s:listMapItem(`[["key1","a"],["key2","b"]]`)=+k8s:validateFalse="listMapItem MultiKey[key1=a,key2=b]"
	MultiKey []MultiItem `json:"multiKey"`

	// +k8s:listType=map
	// +k8s:listMapKey=key
	// +k8s:listMapItem(`[["key","target"]]`)=+k8s:subfield(stringField)=+k8s:validateFalse="listMapItem WithSubfield[key=target].stringField"
	WithSubfield []SubfieldItem `json:"withSubfield"`

	// +k8s:listType=map
	// +k8s:listMapKey=key
	// +k8s:listMapItem(`[["key",""]]`)=+k8s:validateFalse="listMapItem EmptyKey[key=]"
	EmptyKey []Item `json:"emptyKey"`

	// Special characters in key values
	// +k8s:listType=map
	// +k8s:listMapKey=key
	// +k8s:listMapItem(`[["key","with\"quotes"]]`)=+k8s:validateFalse="listMapItem Special[key=with\"quotes]"
	// +k8s:listMapItem(`[["key","multi\nline"]]`)=+k8s:validateFalse="listMapItem Special[key=multi\nline]"
	// +k8s:listMapItem(`[["key","unicode-🚀"]]`)=+k8s:validateFalse="listMapItem Special[key=unicode-🚀]"
	Special []Item `json:"special"`

	// Opaque type handling
	// +k8s:listType=map
	// +k8s:listMapKey=key
	// +k8s:listMapItem(`[["key","opaque"]]`)=+k8s:opaqueType
	// +k8s:listMapItem(`[["key","validated"]]`)=+k8s:validateFalse="listMapItem OpaqueList[key=validated]"
	OpaqueList []OpaqueItem `json:"opaqueList"`
}

type Item struct {
	Key  string `json:"key"`
	Data string `json:"data"`
}

type MultiItem struct {
	Key1 string `json:"key1"`
	Key2 string `json:"key2"`
	Data string `json:"data"`
}

type SubfieldItem struct {
	Key         string `json:"key"`
	StringField string `json:"stringField"`
}

type OpaqueItem struct {
	Key        string `json:"key"`
	OpaqueData string `json:"opaqueData"`
}
