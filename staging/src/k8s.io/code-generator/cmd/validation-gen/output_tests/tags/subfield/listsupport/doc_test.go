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

package listsupport

import (
	"testing"

	"k8s.io/utils/ptr"
)

func TestSubfieldListAccess(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	// --- Base Cases for MainStruct.Conditions ---
	st.Value(&MainStruct{
		Conditions: []MyCondition{
			{Type: "Ready", Status: "Any"},
			{Type: "Progressing", Status: "Any"},
			{Type: "Degraded", Status: "Any"},
		},
		DirectStruct: MyCondition{DefaultType: "Any"},
	}).ExpectValidateFalseByPath(map[string][]string{
		"conditions[Ready]":        {"Ready.Status is being validated (and will fail)"},
		"conditions[Progressing]":  {"Progressing.Status is being validated (and will fail)"},
		"conditions[Degraded]":     {"Degraded.Status is being validated (and will fail)"},
		"directStruct.defaultType": {"Direct.DefaultType is being validated (and will fail)"},
	})

	// Test a single rule triggering: only "Ready" condition exists
	st.Value(&MainStruct{
		Conditions: []MyCondition{
			{Type: "Ready", Status: "Any"},
		},
		DirectStruct: MyCondition{DefaultType: "Any"},
	}).ExpectValidateFalseByPath(map[string][]string{
		"conditions[Ready]":        {"Ready.Status is being validated (and will fail)"},
		"directStruct.defaultType": {"Direct.DefaultType is being validated (and will fail)"},
	})

	st.Value(&MainStruct{
		Conditions: []MyCondition{
			{Type: "Progressing", Status: "Any"},
		},
		DirectStruct: MyCondition{DefaultType: "Any"},
	}).ExpectValidateFalseByPath(map[string][]string{
		"conditions[Progressing]":  {"Progressing.Status is being validated (and will fail)"},
		"directStruct.defaultType": {"Direct.DefaultType is being validated (and will fail)"},
	})

	// --- Edge Cases for MainStruct.Conditions ---
	// No conditions matching the rules, so no errors from list access.
	st.Value(&MainStruct{
		Conditions:   []MyCondition{},
		DirectStruct: MyCondition{DefaultType: "Any"},
	}).ExpectValidateFalseByPath(map[string][]string{
		"directStruct.defaultType": {"Direct.DefaultType is being validated (and will fail)"},
	})

	// Nil conditions list.
	st.Value(&MainStruct{
		Conditions:   nil,
		DirectStruct: MyCondition{DefaultType: "Any"},
	}).ExpectValidateFalseByPath(map[string][]string{
		"directStruct.defaultType": {"Direct.DefaultType is being validated (and will fail)"},
	})

	// Conditions present, but none match the 'type' or 'reason' rules.
	st.Value(&MainStruct{
		Conditions: []MyCondition{
			{Type: "UnknownType1", Status: "Anything"},
			{Type: "UnknownType2", Status: "GoesHere"},
		},
		DirectStruct: MyCondition{DefaultType: "Any"},
	}).ExpectValidateFalseByPath(map[string][]string{
		"directStruct.defaultType": {"Direct.DefaultType is being validated (and will fail)"},
	})

	// Condition matching by 'reason'.
	st.Value(&MainStruct{
		Conditions: []MyCondition{
			{Type: "SomeType", Reason: "Special", Status: "Any"},
		},
		DirectStruct: MyCondition{DefaultType: "Any"},
	}).ExpectValidateFalseByPath(map[string][]string{
		"conditions[Special]":      {"SpecialReason.Status is being validated (and will fail)"}, // Path is conditions[VALUE_OF_REASON_FIELD]
		"directStruct.defaultType": {"Direct.DefaultType is being validated (and will fail)"},
	})

	// Condition matching rule for "NonExistent" type.
	st.Value(&MainStruct{
		Conditions: []MyCondition{
			{Type: "NonExistent", Status: "AnyValueWillFail"},
		},
		DirectStruct: MyCondition{DefaultType: "Any"},
	}).ExpectValidateFalseByPath(map[string][]string{
		"conditions[NonExistent]":  {"NonExistent.Status is being validated (and will fail if it does exist)"},
		"directStruct.defaultType": {"Direct.DefaultType is being validated (and will fail)"},
	})

	// No conditions matching, only direct struct validation.
	st.Value(&MainStruct{
		Conditions: []MyCondition{
			{Type: "Other", Status: "DoesNotMatter"},
		},
		DirectStruct: MyCondition{DefaultType: "Any"},
	}).ExpectValidateFalseByPath(map[string][]string{
		"directStruct.defaultType": {"Direct.DefaultType is being validated (and will fail)"},
	})

	// --- DirectStruct Test (standalone) ---
	st.Value(&MainStruct{
		Conditions:   nil,
		DirectStruct: MyCondition{DefaultType: "AnyValueWillFailThis"},
	}).ExpectValidateFalseByPath(map[string][]string{
		"directStruct.defaultType": {"Direct.DefaultType is being validated (and will fail)"},
	})

	// --- Tests for StructWithPointerField.Items ---
	st.Value(&StructWithPointerField{
		Items: []ElementWithPointer{
			{Key: "Target", Value: ptr.To("AnyValueWillFail")},
			{Key: "Other", Value: ptr.To("SomethingElse")},
		},
	}).ExpectValidateFalseByPath(map[string][]string{
		"items[Target].value": {"PointerInElement.Value is being validated (and will fail)"}, // Path is items[VALUE_OF_KEY_FIELD].value
	})

	// TargetNil element's 'value' field is nil, so validateNil (applied to the .value field) passes.
	st.Value(&StructWithPointerField{
		Items: []ElementWithPointer{
			{Key: "TargetNil", Value: nil}, // Value field is nil
			{Key: "Other", Value: ptr.To("SomethingElse")},
		},
	}).ExpectValid()

	st.Value(&StructWithPointerField{
		Items: []ElementWithPointer{
			{Key: "TargetNil", Value: nil}, // Value field is nil
		},
	}).ExpectValid()

	// TargetNil element's 'value' field is ptr.To("HasValue") (not nil), so validateNil (applied to .value field) fails.
	st.Value(&StructWithPointerField{
		Items: []ElementWithPointer{
			{Key: "TargetNil", Value: ptr.To("HasValue")},
		},
	}).ExpectValidateFalseByPath(map[string][]string{
		"items[TargetNil].value": {"PointerInElement.Value must be nil for TargetNil"},
	})

	// Both rules trigger on their respective '.value' fields.
	st.Value(&StructWithPointerField{
		Items: []ElementWithPointer{
			{Key: "Target", Value: ptr.To("AnyValue")},    // validateFalse on this .value
			{Key: "TargetNil", Value: ptr.To("HasValue")}, // validateNil on this .value
		},
	}).ExpectValidateFalseByPath(map[string][]string{
		"items[Target].value":    {"PointerInElement.Value is being validated (and will fail)"},
		"items[TargetNil].value": {"PointerInElement.Value must be nil for TargetNil"},
	})
}
