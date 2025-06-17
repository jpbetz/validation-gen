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

// Package eqoneofstring tests string validation for eqOneOf tag.
package eqoneofstring

import "k8s.io/code-generator/cmd/validation-gen/testscheme"

var localSchemeBuilder = testscheme.New()

type Struct struct {
	TypeMeta int `json:"typeMeta"`

	// +k8s:eqOneOf=`["valid1","valid2","valid3"]`
	StringField string `json:"stringField"`

	// +k8s:eqOneOf=`["ptr-valid1","ptr-valid2"]`
	StringPtrField *string `json:"stringPtrField"`

	// +k8s:eqOneOf=`["typedef-valid1","typedef-valid2"]`
	TypedefStringField StringType `json:"typedefStringField"`

	// +k8s:eqOneOf=`["ptr-typedef-valid1","ptr-typedef-valid2"]`
	TypedefStringPtrField *StringType `json:"typedefStringPtrField"`

	// +k8s:eqOneOf=`["space valid1","space valid2"]`
	StringWithSpacesField string `json:"stringWithSpacesField"`

	ValidatedTypedefField ValidatedStringType `json:"validatedTypedefField"`
}

type StringType string

// +k8s:eqOneOf=`["validated-valid1","validated-valid2"]`
type ValidatedStringType string
