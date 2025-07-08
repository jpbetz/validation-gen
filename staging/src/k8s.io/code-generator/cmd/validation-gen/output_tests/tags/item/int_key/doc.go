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
package int_key

import "k8s.io/code-generator/cmd/validation-gen/testscheme"

var localSchemeBuilder = testscheme.New()

type Struct struct {
	TypeMeta int `json:"typeMeta"`

	// +k8s:listType=map
	// +k8s:listMapKey=intKey
	// +k8s:item(intKey: -10)=+k8s:validateFalse="item ListField[intKey=-10]"
	ListField []Item `json:"listField"`

	// +k8s:listType=map
	// +k8s:listMapKey=key1
	// +k8s:listMapKey=key2
	// +k8s:item(key1: -1, key2: "a")=+k8s:validateFalse="item MultiKey[key1=-1,key2=a]"
	MultiKey []MultiKeyItem `json:"multiKey"`
}

type Item struct {
	IntKey      int    `json:"intKey"`
	StringField string `json:"stringField"`
}

type MultiKeyItem struct {
	Key1        int    `json:"key1"`
	Key2        string `json:"key2"`
	StringField string `json:"stringField"`
}

// +k8s:listType=map
// +k8s:listMapKey=intField
// +k8s:item(intField: -100)=+k8s:validateFalse="item TypedefList[intField=-100]"
type TypedefList []TypedefItem

type TypedefItem struct {
	IntField    int    `json:"intField"`
	StringField string `json:"stringField"`
}

type StructWithTypedef struct {
	TypeMeta int `json:"typeMeta"`

	TypedefField TypedefList `json:"typedefField"`
}
