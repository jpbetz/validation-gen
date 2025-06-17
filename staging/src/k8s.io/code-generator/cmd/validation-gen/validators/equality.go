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

package validators

import (
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/code-generator/cmd/validation-gen/util"
	"k8s.io/gengo/v2/codetags"
	"k8s.io/gengo/v2/types"
)

const (
	eqOneOfTagName = "k8s:eqOneOf"
)

func init() {
	RegisterTagValidator(eqOneOfTagValidator{})
}

type eqOneOfTagValidator struct{}

func (eqOneOfTagValidator) Init(_ Config) {}

func (eqOneOfTagValidator) TagName() string {
	return eqOneOfTagName
}

var eqOneOfTagValidScopes = sets.New(ScopeAny)

func (eqOneOfTagValidator) ValidScopes() sets.Set[Scope] {
	return eqOneOfTagValidScopes
}

var (
	eqOneOfValidator = types.Name{Package: libValidationPkg, Name: "EqOneOf"}
)

func buildSliceLiteral[T any](fieldType *types.Type, nativeType *types.Type, values []T, format func(T) string) Literal {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("[]%s{", fieldType.Name.Name))

	for i, v := range values {
		if i > 0 {
			b.WriteString(", ")
		}

		// If it's a typedef, we need to cast.
		if fieldType != nativeType {
			b.WriteString(fmt.Sprintf("%s(%s)", fieldType.Name.Name, format(v)))
		} else {
			b.WriteString(format(v))
		}
	}

	b.WriteString("}")
	return Literal(b.String())
}

func (v eqOneOfTagValidator) GetValidations(context Context, tag codetags.Tag) (Validations, error) {
	t := util.NonPointer(util.NativeType(context.Type))
	fieldType := util.NonPointer(context.Type)

	if !util.IsDirectComparable(t) {
		return Validations{}, fmt.Errorf("can only be used on comparable types (e.g. string, int, bool), but got %s", rootTypeString(context.Type, t))
	}

	if tag.ValueType != codetags.ValueTypeString {
		return Validations{}, fmt.Errorf("missing required payload in backticks")
	}

	var rawValues []interface{}
	if err := json.Unmarshal([]byte(tag.Value), &rawValues); err != nil {
		return Validations{}, fmt.Errorf("payload must be a valid JSON array, got: %s (error: %w)", tag.Value, err)
	}

	if len(rawValues) == 0 {
		return Validations{}, fmt.Errorf("array cannot be empty")
	}

	var literal Literal

	switch t {
	case types.String:
		values := make([]string, 0, len(rawValues))
		for i, raw := range rawValues {
			str, ok := raw.(string)
			if !ok {
				return Validations{}, fmt.Errorf("array element at index %d must be a string, got %T", i, raw)
			}
			values = append(values, str)
		}
		literal = buildSliceLiteral(fieldType, types.String, values, func(s string) string {
			return fmt.Sprintf("%q", s)
		})

	case types.Bool:
		values := make([]bool, 0, len(rawValues))
		for i, raw := range rawValues {
			b, ok := raw.(bool)
			if !ok {
				return Validations{}, fmt.Errorf("array element at index %d must be a bool, got %T", i, raw)
			}
			values = append(values, b)
		}
		literal = buildSliceLiteral(fieldType, types.Bool, values, func(b bool) string {
			return fmt.Sprintf("%t", b)
		})

	default:
		if types.IsInteger(t) {
			values := make([]int, 0, len(rawValues))
			for i, raw := range rawValues {
				// JSON unmarshals numbers as float64.
				f, ok := raw.(float64)
				if !ok {
					return Validations{}, fmt.Errorf("array element at index %d must be a number, got %T", i, raw)
				}
					// Check if whole number by checking float == (float -> int -> float).
					intVal := int(f)
					if float64(intVal) != f {
					return Validations{}, fmt.Errorf("array element at index %d must be an integer, got %v", i, f)
				}
				values = append(values, intVal)
			}
			literal = buildSliceLiteral(fieldType, t, values, func(i int) string {
				return fmt.Sprintf("%d", i)
			})
		} else {
			return Validations{}, fmt.Errorf("unsupported type for 'eqOneOf' tag: %s", t.Name)
		}
	}

	fn := Function(v.TagName(), DefaultFlags, eqOneOfValidator, literal)
	return Validations{Functions: []FunctionGen{fn}}, nil
}

func (v eqOneOfTagValidator) Docs() TagDoc {
	return TagDoc{
		Tag:              v.TagName(),
		Scopes:           v.ValidScopes().UnsortedList(),
		Description:      "Verifies the field's value is equal to one of the allowed values. Supports string, integer, and bool types.",
		PayloadsRequired: true,
		PayloadsType:     codetags.ValueTypeString,
		Payloads: []TagPayloadDoc{{
			Description: `JSON array`,
			Docs:        `A JSON array of allowed values. Examples: ["a","b","c"] for strings, [1,2,3] for integers, [true,false] for bools.`,
		}},
		Usage: `+k8s:eqOneOf=["a","b","c"] or +k8s:eqOneOf=[1,2,3] or +k8s:eqOneOf=[true,false]`,
	}
}
