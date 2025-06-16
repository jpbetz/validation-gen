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

package union

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestItemUnion(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	validSingleState := &Pipeline{
		Tasks: []Task{
			{Name: "succeeded", State: "Succeeded"},
			{Name: "other", State: "Other"},
		},
	}
	st.Value(validSingleState).ExpectValid()

	invalidMultipleStates := &Pipeline{
		Tasks: []Task{
			{Name: "succeeded", State: "Succeeded"},
			{Name: "failed", State: "Failed"},
			{Name: "other", State: "Other"},
		},
	}
	st.Value(invalidMultipleStates).ExpectInvalid(
		field.Invalid(nil, "{Tasks[name=\"succeeded\"], Tasks[name=\"failed\"]}",
			"must specify exactly one of: `Tasks[name=\"succeeded\"]`, `Tasks[name=\"failed\"]`, `Tasks[name=\"running\"]`, `Tasks[name=\"pending\"]`"),
	)

	invalidNoUnionMembers := &Pipeline{
		Tasks: []Task{
			{Name: "other", State: "Other"},
		},
	}
	st.Value(invalidNoUnionMembers).ExpectInvalid(
		field.Invalid(nil, "",
			"must specify one of: `Tasks[name=\"succeeded\"]`, `Tasks[name=\"failed\"]`, `Tasks[name=\"running\"]`, `Tasks[name=\"pending\"]`"),
	)
}
