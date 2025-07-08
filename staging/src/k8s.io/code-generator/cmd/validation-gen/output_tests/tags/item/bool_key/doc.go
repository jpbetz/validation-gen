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
package bool_key

import "k8s.io/code-generator/cmd/validation-gen/testscheme"

var localSchemeBuilder = testscheme.New()

type Struct struct {
	TypeMeta int `json:"typeMeta"`

	// +k8s:listType=map
	// +k8s:listMapKey=boolKey
	// +k8s:item(boolKey: true)=+k8s:validateFalse="item ListField[boolKey=true]"
	ListField []Item `json:"listField"`

	// +k8s:listType=map
	// +k8s:listMapKey=key1
	// +k8s:listMapKey=key2
	// +k8s:listMapKey=key3
	// +k8s:item(key1: true, key2: 5, key3: "x")=+k8s:validateFalse="item MultiKey[key1=true,key2=5,key3=x]"
	MultiKey []MultiKeyItem `json:"multiKey"`
}

type Item struct {
	BoolKey     bool   `json:"boolKey"`
	StringField string `json:"stringField"`
}

type MultiKeyItem struct {
	Key1        bool   `json:"key1"`
	Key2        int    `json:"key2"`
	Key3        string `json:"key3"`
	StringField string `json:"stringField"`
}

// +k8s:listType=map
// +k8s:listMapKey=boolField
// +k8s:item(boolField: true)=+k8s:validateFalse="item TypedefList[boolField=true]"
type TypedefList []TypedefItem

type TypedefItem struct {
	BoolField   bool   `json:"boolField"`
	StringField string `json:"stringField"`
}

type StructWithTypedef struct {
	TypeMeta int `json:"typeMeta"`

	TypedefField TypedefList `json:"typedefField"`
}
