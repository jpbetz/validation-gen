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

// Package listsupport contains test types for testing subfield list element access.
package listsupport

import "k8s.io/code-generator/cmd/validation-gen/testscheme"

var localSchemeBuilder = testscheme.New()

// MainStruct demonstrates subfield validation.
type MainStruct struct {
	TypeMeta int `json:"typeMeta"`

	// List access using "config-in-payload-prefix" style with explicit sub-targeting:
	// These apply to the *entire MyCondition element* selected.
	// +k8s:subfield={"type":"Ready"}=+k8s:validateFalse="Ready.Status is being validated (and will fail)"
	// +k8s:subfield={"type":"Progressing"}=+k8s:validateFalse="Progressing.Status is being validated (and will fail)"
	// +k8s:subfield={"type":"Degraded"}=+k8s:validateFalse="Degraded.Status is being validated (and will fail)"
	// +k8s:subfield={"type":"NonExistent"}=+k8s:validateFalse="NonExistent.Status is being validated (and will fail if it does exist)"
	// +k8s:subfield={"reason":"Special"}=+k8s:validateFalse="SpecialReason.Status is being validated (and will fail)"
	Conditions []MyCondition `json:"conditions,omitempty"`

	// +k8s:subfield(defaultType)=+k8s:validateFalse="Direct.DefaultType is being validated (and will fail)"
	DirectStruct MyCondition `json:"directStruct"`
}

// MyCondition is an element type for the list.
type MyCondition struct {
	Type        string `json:"type"`
	Status      string `json:"status"`
	Reason      string `json:"reason,omitempty"`
	DefaultType string `json:"defaultType,omitempty"`
}

// StructWithPointerField tests subfield access to pointer fields within a list element.
type StructWithPointerField struct {
	TypeMeta int `json:"typeMeta"`

	// +k8s:subfield={"key":"Target"}=+k8s:subfield(value)=+k8s:validateFalse="PointerInElement.Value is being validated (and will fail)"
	// +k8s:subfield:{"key":"TargetNil"}=+k8s:subfield(value)=+k8s:validateNil="PointerInElement.Value must be nil for TargetNil"
	Items []ElementWithPointer `json:"items,omitempty"`
}

type ElementWithPointer struct {
	Key   string  `json:"key"`
	Value *string `json:"value,omitempty"` // This is the field targeted by the second subfield tag
}
