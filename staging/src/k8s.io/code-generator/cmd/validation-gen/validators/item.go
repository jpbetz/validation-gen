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
	"fmt"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/code-generator/cmd/validation-gen/util"
	"k8s.io/gengo/v2/codetags"
	"k8s.io/gengo/v2/types"
)

const (
	itemTagName = "k8s:item"
)

type keyValuePair struct {
	key       string
	value     any
	valueType codetags.ValueType
}

type itemValidation struct {
	// matcherPairs contains the key value pairs to match list items
	matcherPairs []keyValuePair
	// valueTag is the validation tag to apply to matched items
	valueTag codetags.Tag
}

type itemMetadata struct {
	items []itemValidation
}

type itemTagValidator struct {
	byPath map[string]*itemMetadata
}

func (itv *itemTagValidator) Init(cfg Config) {}

func (itemTagValidator) TagName() string {
	return itemTagName
}

var itemTagValidScopes = sets.New(ScopeAny)

func (itemTagValidator) ValidScopes() sets.Set[Scope] {
	return itemTagValidScopes
}

func (itv *itemTagValidator) GetValidations(context Context, tag codetags.Tag) (Validations, error) {
	// TODO: Support regular maps with syntax like:
	// +k8s:item("map-key")=+k8s:immutable

	// Parse key-value pairs from named args.
	matcherPairs := []keyValuePair{}
	processedKeys := sets.NewString()

	for _, arg := range tag.Args {
		if arg.Name == "" {
			return Validations{}, fmt.Errorf("all arguments must be named (ex: fieldName:value)")
		}
		if processedKeys.Has(arg.Name) {
			return Validations{}, fmt.Errorf("duplicate key %q", arg.Name)
		}
		processedKeys.Insert(arg.Name)

		parsedValue, valueType, err := parseTypedValue(arg.Value, arg.Type)
		if err != nil {
			return Validations{}, fmt.Errorf("invalid value for key %q: %w", arg.Name, err)
		}

		matcherPairs = append(matcherPairs, keyValuePair{
			key:       arg.Name,
			value:     parsedValue,
			valueType: valueType,
		})
	}

	if len(matcherPairs) == 0 {
		return Validations{}, fmt.Errorf("no selection criteria was specified")
	}

	if tag.ValueType != codetags.ValueTypeTag || tag.ValueTag == nil {
		return Validations{}, fmt.Errorf("requires a validation tag as its value payload")
	}

	// This tag can apply to value and pointer fields, as well as typedefs
	// (which should never be pointers). We need to check the concrete type.
	t := util.NonPointer(util.NativeType(context.Type))

	if t.Kind != types.Slice {
		return Validations{}, fmt.Errorf("can only be used on list types")
	}

	elemT := util.NonPointer(util.NativeType(t.Elem))
	if elemT.Kind != types.Struct {
		return Validations{}, fmt.Errorf("can only be used on lists of structs")
	}

	// Store metadata for the field validator to use.
	if itv.byPath[context.Path.String()] == nil {
		itv.byPath[context.Path.String()] = &itemMetadata{}
	}

	itv.byPath[context.Path.String()].items = append(itv.byPath[context.Path.String()].items, itemValidation{
		matcherPairs: matcherPairs,
		valueTag:     *tag.ValueTag,
	})

	// This tag doesn't generate validations directly, the itemFieldValidator does.
	return Validations{}, nil
}

// parseTypedValue parses a value based on its detected type
func parseTypedValue(value string, argType codetags.ArgType) (any, codetags.ValueType, error) {
	switch argType {
	case codetags.ArgTypeString:
		return value, codetags.ValueTypeString, nil
	case codetags.ArgTypeInt:
		intVal, err := strconv.Atoi(value)
		if err != nil {
			return nil, "", fmt.Errorf("invalid integer: %w", err)
		}
		return intVal, codetags.ValueTypeInt, nil
	case codetags.ArgTypeBool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return nil, "", fmt.Errorf("invalid boolean: %w", err)
		}
		return boolVal, codetags.ValueTypeBool, nil
	default:
		// Default to string
		return value, codetags.ValueTypeString, nil
	}
}

func (itv itemTagValidator) Docs() TagDoc {
	doc := TagDoc{
		Tag:    itv.TagName(),
		Scopes: itv.ValidScopes().UnsortedList(),
		Description: "Declares a validation for an item of a slice declared as a +k8s:listType=map. " +
			"The item to match is declared by providing field-value pair arguments. All +k8s:listMapKey fields must be included in the field-value pair arguments.",
		Usage: "+k8s:item(stringKey: \"value\", intKey: 42, boolKey: true)=<validation-tag>",
		Docs: "Arguments must be named with the JSON names of the list map key fields. " +
			"Values can be strings (quoted or unquoted), integers, or booleans. " +
			"For example: +k8s:item(name: \"myname\", priority: 10, enabled: true)=<chained-validation-tag>",
		AcceptsUnknownArgs: true,
		Payloads: []TagPayloadDoc{{
			Description: "<validation-tag>",
			Docs:        "The tag to evaluate for the matching list item.",
		}},
		PayloadsType:     codetags.ValueTypeTag,
		PayloadsRequired: true,
	}
	return doc
}

type itemValidator struct {
	validator  Validator
	listByPath map[string]*listMetadata
	itemByPath map[string]*itemMetadata
}

func (iv *itemValidator) Init(cfg Config) {
	iv.validator = cfg.Validator
}

func (itemValidator) Name() string {
	return "itemFieldValidator"
}

var (
	validateSliceItem = types.Name{Package: libValidationPkg, Name: "SliceItem"}
)

func (iv itemValidator) GetValidations(context Context) (Validations, error) {
	itemMeta, ok := iv.itemByPath[context.Path.String()]
	if !ok || len(itemMeta.items) == 0 {
		return Validations{}, nil
	}

	// For fields, list metadata can fall back to the type.
	// For types, list metadata must be defined on the type itself.
	listMeta, ok := iv.listByPath[context.Path.String()]
	if !ok {
		if context.Scope == ScopeField {
			typePath := context.Type.String()
			listMeta, ok = iv.listByPath[typePath]
		}
	}

	// Fields inherit list metadata from typedefs, but not vice-versa.
	// If we find no listMetadata then something is wrong.
	if !ok || !listMeta.declaredAsMap || len(listMeta.keyFields) == 0 {
		return Validations{}, fmt.Errorf("must have +k8s:listType=map and at least one '+k8s:listMapKey=...' annotation to use +k8s:item")
	}

	t := util.NonPointer(util.NativeType(context.Type))
	elemT := util.NonPointer(util.NativeType(t.Elem))

	result := Validations{}

	for _, item := range itemMeta.items {
		if len(item.matcherPairs) != len(listMeta.keyNames) {
			return Validations{}, fmt.Errorf("number of arguments does not match number of listMapKey fields")
		}

		// Validate that all listMapKeys are provided and types match
		foundKeys := make(map[string]bool)
		for _, keyName := range listMeta.keyNames {
			for _, pair := range item.matcherPairs {
				if pair.key == keyName {
					member := util.GetMemberByJSON(elemT, pair.key)
					if member != nil {
						if err := validateTypeMatch(member.Type, pair.value); err != nil {
							return Validations{}, fmt.Errorf("key %q: %w", pair.key, err)
						}
					}
					foundKeys[keyName] = true
					break
				}
			}
		}
		if len(foundKeys) != len(listMeta.keyNames) {
			missing := []string{}
			for _, k := range listMeta.keyNames {
				if !foundKeys[k] {
					missing = append(missing, k)
				}
			}
			return Validations{}, fmt.Errorf("missing required listMapKey fields: %v", missing)
		}

		// Extract validations from the stored tag
		subContextPath := generateFieldPathForMap(item.matcherPairs)
		subContext := Context{
			Scope:  ScopeListVal,
			Type:   elemT,
			Parent: context.Type,
			Path:   context.Path.Key(subContextPath),
			Member: nil,
		}

		validations, err := iv.validator.ExtractValidations(subContext, item.valueTag)
		if err != nil {
			return Validations{}, err
		}

		result.Variables = append(result.Variables, validations.Variables...)

		matchFn, err := createMatchFn(elemT, item.matcherPairs)
		if err != nil {
			return Validations{}, err
		}

		for _, vfn := range validations.Functions {
			f := Function(
				itemTagName,
				vfn.Flags,
				validateSliceItem,
				matchFn,
				WrapperFunction{vfn, elemT},
			)
			result.Functions = append(result.Functions, f)
		}
	}

	return result, nil
}

// validateTypeMatch ensures the provided value matches the field's type
func validateTypeMatch(fieldType *types.Type, value any) error {
	nativeType := util.NativeType(fieldType)

	switch {
	case nativeType == types.String:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string value")
		}
	case nativeType == types.Bool:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected bool value")
		}
	case types.IsInteger(nativeType):
		if _, ok := value.(int); !ok {
			return fmt.Errorf("expected integer value")
		}
	default:
		return fmt.Errorf("unsupported field type: %s", nativeType.String())
	}
	return nil
}

func createMatchFn(elemT *types.Type, matcherPairs []keyValuePair) (FunctionLiteral, error) {
	var conditions []string

	for _, pair := range matcherPairs {
		member := util.GetMemberByJSON(elemT, pair.key)
		if member == nil {
			return FunctionLiteral{}, fmt.Errorf("no field with JSON name %q", pair.key)
		}
		// Generate the comparison based on the field's actual type
		rhs, err := generateComparisonRHS(member, pair.value)
		if err != nil {
			return FunctionLiteral{}, err
		}
		conditions = append(conditions, fmt.Sprintf("item.%s == %s", member.Name, rhs))
	}

	return FunctionLiteral{
		Parameters: []ParamResult{{"item", types.PointerTo(elemT)}},
		Results:    []ParamResult{{"", types.Bool}},
		Body:       fmt.Sprintf("return %s", strings.Join(conditions, " && ")),
	}, nil
}

func generateComparisonRHS(member *types.Member, value any) (string, error) {
	memberType := util.NativeType(member.Type)
	switch {
	case memberType == types.String:
		strVal, ok := value.(string)
		if !ok {
			return "", fmt.Errorf("type mismatch, field is string but value is not")
		}
		return fmt.Sprintf("%q", strVal), nil

	case memberType == types.Bool:
		boolVal, ok := value.(bool)
		if !ok {
			return "", fmt.Errorf("type mismatch, field is bool but value is not")
		}
		return fmt.Sprintf("%t", boolVal), nil

	case types.IsInteger(memberType):
		intVal, ok := value.(int)
		if !ok {
			return "", fmt.Errorf("type mismatch, field is int but value is not")
		}
		return fmt.Sprintf("%d", intVal), nil

	default:
		return "", fmt.Errorf("unsupported type %s for field %s", memberType.String(), member.Name)
	}
}

// generateFieldPathForMap creates a field path for list map items using a JSON-like syntax.
// The format is {"key": "value", "key2": 42, "key3": true} with quoted keys and appropriately formatted values.
func generateFieldPathForMap(matcherPairs []keyValuePair) string {
	var pairs []string
	for _, pair := range matcherPairs {
		valueStr := formatValueForPath(pair.value)
		pairs = append(pairs, fmt.Sprintf("%q: %s", pair.key, valueStr))
	}
	return fmt.Sprintf("{%s}", strings.Join(pairs, ", "))
}

func formatValueForPath(value any) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("%q", v)
	case int:
		return fmt.Sprintf("%d", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
