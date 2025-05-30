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
package transitions

import "k8s.io/code-generator/cmd/validation-gen/testscheme"

var localSchemeBuilder = testscheme.New()

type Struct struct {
	TypeMeta int `json:"typeMeta"`

	// +k8s:listType=map
	// +k8s:listMapKey=id
	// +k8s:listMapItem(`[["id","temp"]]`)=+k8s:validateFalse="listMapItem Items[id=temp]"
	// +k8s:listMapItem(`[["id","high"]]`)=+k8s:validateFalse="listMapItem Items[id=high]"
	Items []Item `json:"items"`

	// +k8s:listType=map
	// +k8s:listMapKey=key1
	// +k8s:listMapKey=key2
	// +k8s:listMapItem(`[["key1","a"],["key2","1"]]`)=+k8s:immutable
	// +k8s:listMapItem(`[["key1","b"],["key2","1"]]`)=+k8s:subfield(stringField)=+k8s:immutable
	MultiKey []MultiKeyItem `json:"multiKey"`
}

type Item struct {
	ID   string `json:"id"`
	Data string `json:"data"`
}

type MultiKeyItem struct {
	Key1        string `json:"key1"`
	Key2        string `json:"key2"`
	StringField string `json:"stringField"`
}
