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

package zeroroneof

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestItemZeroOrOneOf(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	st.Value(&Pipeline{
		Tasks: []Task{
			{Name: "other", State: "Other"},
			{Name: "another", State: "Running"},
		},
	}).ExpectValid()

	st.Value(&Pipeline{
		Tasks: []Task{
			{Name: "succeeded", State: "Succeeded"},
			{Name: "other", State: "Other"},
		},
	}).ExpectValid()

	st.Value(&Pipeline{
		Tasks: []Task{},
	}).ExpectValid()

	invalidMultipleMembers := &Pipeline{
		Tasks: []Task{
			{Name: "succeeded", State: "Succeeded"},
			{Name: "failed", State: "Failed"},
			{Name: "other", State: "Other"},
		},
	}
	st.Value(invalidMultipleMembers).ExpectMatches(
		field.ErrorMatcher{},
		field.ErrorList{
			field.Invalid(field.NewPath("tasks"), "{Tasks[{\"name\": \"failed\"}], Tasks[{\"name\": \"succeeded\"}]}",
				"must specify at most one of: `Tasks[{\"name\": \"succeeded\"}]`, `Tasks[{\"name\": \"failed\"}]`"),
		},
	)

	// Test ratcheting.
	st.Value(invalidMultipleMembers).OldValue(invalidMultipleMembers).ExpectValid()
}
