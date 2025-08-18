
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

package enum

type ConditionalStruct struct {
	TypeMeta int

	ConditionalEnumField    ConditionalEnum  `json:"conditionalEnumField"`
	ConditionalEnumPtrField *ConditionalEnum `json:"conditionalEnumPtrField"`
}

// +k8s:enum
type ConditionalEnum string

const (
	// +k8s:ifEnabled(FeatureA)=+k8s:enumExclude
	ConditionalA ConditionalEnum = "A"

	// +k8s:ifDisabled(FeatureB)=+k8s:enumExclude
	ConditionalB ConditionalEnum = "B"

	// This value is always included.
	ConditionalC ConditionalEnum = "C"

	// +k8s:ifEnabled(FeatureA)=+k8s:enumExclude
	// +k8s:ifEnabled(FeatureB)=+k8s:enumExclude
	ConditionalD ConditionalEnum = "D"

	// +k8s:ifDisabled(FeatureC)=+k8s:enumExclude
	// +k8s:ifDisabled(FeatureD)=+k8s:enumExclude
	ConditionalE ConditionalEnum = "E"

	// +k8s:ifDisabled(FeatureC)=+k8s:enumExclude
	// +k8s:ifEnabled(FeatureD)=+k8s:enumExclude
	ConditionalF ConditionalEnum = "F"
)
