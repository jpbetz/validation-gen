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

package listsimple

import "k8s.io/code-generator/cmd/validation-gen/testscheme"

var localSchemeBuilder = testscheme.New()

type Struct struct {
	TypeMeta int `json:"typeMeta"`

	// +listType=map
	// +listMapKey=type
	// +k8s:subfield({"type":"Approved"})=+k8s:validateFalse="subfield Conditions[type=Approved]"
	// +k8s:subfield({"status":"True","type":"Approved"})=+k8s:validateFalse="subfield Conditions[status=True,type=Approved]"
	// +k8s:subfield({"stringPtr":"Target", "type":"Approved"})=+k8s:validateFalse="subfield Conditions[stringPtr=Target,type=Approved]"
	Conditions []MyCondition `json:"conditions"`
}

type MyCondition struct {
	Type      string  `json:"type"`
	Status    string  `json:"status"`
	StringPtr *string `json:"stringPtr"`
}
