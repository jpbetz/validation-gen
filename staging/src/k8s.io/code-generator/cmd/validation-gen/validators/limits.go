/*
Copyright 2024 The Kubernetes Authors.

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
	"fmt"
	"strconv"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/gengo/v2/types"
)

const (
	maxLengthTagName = "k8s:maxLength"
	maxItemsTagName  = "k8s:maxItems"
	minimumTagName   = "k8s:minimum"
	maximumTagName   = "k8s:maximum"
)

func init() {
	RegisterTagValidator(maxLengthTagValidator{})
	RegisterTagValidator(maxItemsTagValidator{})

	RegisterTagValidator(minimumTagValidator{})
	RegisterTagValidator(maximumTagValidator{})
}

type refLimit struct {
	min, max *string // name of the field that is referenced and provides a min or max limit
}

type maxLengthTagValidator struct{}

func (maxLengthTagValidator) Init(_ Config) {}

func (maxLengthTagValidator) TagName() string {
	return maxLengthTagName
}

var maxLengthTagValidScopes = sets.New(ScopeAny)

func (maxLengthTagValidator) ValidScopes() sets.Set[Scope] {
	return maxLengthTagValidScopes
}

var (
	maxLengthValidator = types.Name{Package: libValidationPkg, Name: "MaxLength"}
)

func (maxLengthTagValidator) GetValidations(context Context, _ []string, payload string) (Validations, error) {
	var result Validations

	// This tag can apply to value and pointer fields, as well as typedefs
	// (which should never be pointers). We need to check the concrete type.
	if t := nonPointer(nativeType(context.Type)); t != types.String {
		return Validations{}, fmt.Errorf("can only be used on string types (%s)", rootTypeString(context.Type, t))
	}

	intVal, err := strconv.Atoi(payload)
	if err != nil {
		return result, fmt.Errorf("failed to parse tag payload as int: %v", err)
	}
	if intVal < 0 {
		return result, fmt.Errorf("must be greater than or equal to zero")
	}
	result.AddFunction(Function(maxLengthTagName, DefaultFlags, maxLengthValidator, intVal))
	return result, nil
}

func (mltv maxLengthTagValidator) Docs() TagDoc {
	return TagDoc{
		Tag:         mltv.TagName(),
		Scopes:      mltv.ValidScopes().UnsortedList(),
		Description: "Indicates that a string field has a limit on its length.",
		Payloads: []TagPayloadDoc{{
			Description: "<non-negative integer>",
			Docs:        "This field must be no more than X characters long.",
		}},
	}
}

type maxItemsTagValidator struct{}

func (maxItemsTagValidator) Init(_ Config) {}

func (maxItemsTagValidator) TagName() string {
	return maxItemsTagName
}

var maxItemsTagValidScopes = sets.New(
	ScopeType,
	ScopeField,
	ScopeListVal,
	ScopeMapVal,
)

func (maxItemsTagValidator) ValidScopes() sets.Set[Scope] {
	return maxItemsTagValidScopes
}

var (
	maxItemsValidator = types.Name{Package: libValidationPkg, Name: "MaxItems"}
)

func (maxItemsTagValidator) GetValidations(context Context, _ []string, payload string) (Validations, error) {
	var result Validations

	// NOTE: pointers to lists are not supported, so we should never see a pointer here.
	if t := nativeType(context.Type); t.Kind != types.Slice && t.Kind != types.Array {
		return Validations{}, fmt.Errorf("can only be used on list types (%s)", rootTypeString(context.Type, t))
	}

	intVal, err := strconv.Atoi(payload)
	if err != nil {
		return result, fmt.Errorf("failed to parse tag payload as int: %v", err)
	}
	if intVal < 0 {
		return result, fmt.Errorf("must be greater than or equal to zero")
	}
	// Note: maxItems short-circuits other validations.
	result.AddFunction(Function(maxItemsTagName, ShortCircuit, maxItemsValidator, intVal))
	return result, nil
}

func (mitv maxItemsTagValidator) Docs() TagDoc {
	return TagDoc{
		Tag:         mitv.TagName(),
		Scopes:      mitv.ValidScopes().UnsortedList(),
		Description: "Indicates that a list field has a limit on its size.",
		Payloads: []TagPayloadDoc{{
			Description: "<non-negative integer>",
			Docs:        "This field must be no more than X items long.",
		}},
	}
}

type minimumTagValidator struct{}

func (minimumTagValidator) Init(_ Config) {}

func (minimumTagValidator) TagName() string {
	return minimumTagName
}

var minimumTagValidScopes = sets.New(
	ScopeAny,
)

func (minimumTagValidator) ValidScopes() sets.Set[Scope] {
	return minimumTagValidScopes
}

var (
	minimumValidator      = types.Name{Package: libValidationPkg, Name: "Minimum"}
	minimumFieldValidator = types.Name{Package: libValidationPkg, Name: "MinimumField"}
)

func (mtv minimumTagValidator) GetValidations(context Context, args []string, payload string) (Validations, error) {
	var result Validations

	if field, ok, err := FieldReference(context, payload); ok {
		if err != nil {
			return result, err
		}
		vfn := Function(minimumTagName, DefaultFlags, minimumFieldValidator, field, field.FieldName)
		result.Functions = append(result.Functions, vfn)
		return result, nil
	}

	// This tag can apply to value and pointer fields, as well as typedefs
	// (which should never be pointers). We need to check the concrete type.
	if t := nonPointer(nativeType(context.Type)); !types.IsInteger(t) {
		return result, fmt.Errorf("can only be used on integer types (%s)", rootTypeString(context.Type, t))
	}

	if len(args) == 1 {
		// TODO
		panic("not implemented")
	}

	intVal, err := strconv.Atoi(payload)
	if err != nil {
		return result, fmt.Errorf("failed to parse tag payload as int: %w", err)
	}
	result.AddFunction(Function(minimumTagName, DefaultFlags, minimumValidator, intVal))
	return result, nil
}

func (mtv minimumTagValidator) Docs() TagDoc {
	return TagDoc{
		Tag:         mtv.TagName(),
		Scopes:      mtv.ValidScopes().UnsortedList(),
		Description: "Indicates that a numeric field has a minimum value.",
		Payloads: []TagPayloadDoc{{
			Description: "<integer>",
			Docs:        "This field must be greater than or equal to x.",
		}},
	}
}

type maximumTagValidator struct{}

func (maximumTagValidator) Init(_ Config) {}

func (maximumTagValidator) TagName() string {
	return maximumTagName
}

var maximumTagValidScopes = sets.New(
	ScopeAny,
)

func (maximumTagValidator) ValidScopes() sets.Set[Scope] {
	return maximumTagValidScopes
}

var (
	maximumValidator      = types.Name{Package: libValidationPkg, Name: "Maximum"}
	maximumFieldValidator = types.Name{Package: libValidationPkg, Name: "MaximumField"}
)

func (mtv maximumTagValidator) GetValidations(context Context, args []string, payload string) (Validations, error) {
	var result Validations

	if field, ok, err := FieldReference(context, payload); ok {
		if err != nil {
			return result, err
		}
		vfn := Function(maximumTagName, DefaultFlags, maximumFieldValidator, field, field.FieldName)
		result.Functions = append(result.Functions, vfn)
		return result, nil
	}

	if t := nonPointer(nativeType(context.Type)); !types.IsInteger(t) {
		return result, fmt.Errorf("can only be used on integer types (%s)", rootTypeString(context.Type, t))
	}

	intVal, err := strconv.Atoi(payload)
	if err != nil {
		return result, fmt.Errorf("failed to parse tag payload as int: %w", err)
	}
	result.AddFunction(Function(maximumTagName, DefaultFlags, maximumValidator, intVal))
	return result, nil
}

func (mtv maximumTagValidator) Docs() TagDoc {
	return TagDoc{
		Tag:         mtv.TagName(),
		Scopes:      mtv.ValidScopes().UnsortedList(),
		Description: "Indicates that a numeric field has a maximum value.",
		Payloads: []TagPayloadDoc{{
			Description: "<integer>",
			Docs:        "This field must be less than or equal to x.",
		}},
	}
}
