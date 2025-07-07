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
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/code-generator/cmd/validation-gen/util"
	"k8s.io/gengo/v2/codetags"
	"k8s.io/gengo/v2/types"
)

const (
	itemTagName = "k8s:item"
)

type itemValidation struct {
	matcherPairs [][2]string
	valueTag     codetags.Tag
	elemType     *types.Type
}

type itemMetadata struct {
	items []itemValidation
}

type itemTagValidator struct {
	validator   Validator
	byFieldPath map[string]*itemMetadata
}

func (itv *itemTagValidator) Init(cfg Config) {
	itv.validator = cfg.Validator
	if itv.byFieldPath == nil {
		itv.byFieldPath = make(map[string]*itemMetadata)
	}
}

func (itemTagValidator) TagName() string {
	return itemTagName
}

var itemTagValidScopes = sets.New(ScopeField)

func (itemTagValidator) ValidScopes() sets.Set[Scope] {
	return itemTagValidScopes
}

func (itv *itemTagValidator) GetValidations(context Context, tag codetags.Tag) (Validations, error) {
	// Parse key-value pairs from named args.
	matcherPairs := [][2]string{}
	processedKeys := sets.NewString()

	for _, arg := range tag.Args {
		if arg.Name == "" {
			return Validations{}, fmt.Errorf("all arguments must be named (ex: fieldName:value)")
		}
		if processedKeys.Has(arg.Name) {
			return Validations{}, fmt.Errorf("duplicate key %q in item", arg.Name)
		}
		processedKeys.Insert(arg.Name)
		matcherPairs = append(matcherPairs, [2]string{arg.Name, arg.Value})
	}

	if len(matcherPairs) == 0 {
		return Validations{}, fmt.Errorf("item requires at least one key-value pair")
	}

	if tag.ValueType != codetags.ValueTypeTag {
		return Validations{}, fmt.Errorf("item requires a validation tag as its value payload")
	}

	if tag.ValueTag == nil {
		return Validations{}, fmt.Errorf("item requires a non-nil validation tag as its value payload")
	}

	// This tag can apply to value and pointer fields, as well as typedefs
	// (which should never be pointers). We need to check the concrete type.
	t := util.NonPointer(util.NativeType(context.Type))

	if t.Kind != types.Slice {
		return Validations{}, fmt.Errorf("can only be used on list types")
	}

	elemT := util.NonPointer(util.NativeType(t.Elem))
	if elemT.Kind != types.Struct {
		return Validations{}, fmt.Errorf("can only be used on list of structs")
	}

	// Store metadata for the field validator to use.
	fieldPath := context.Path.String()
	if itv.byFieldPath[fieldPath] == nil {
		itv.byFieldPath[fieldPath] = &itemMetadata{}
	}

	itv.byFieldPath[fieldPath].items = append(itv.byFieldPath[fieldPath].items, itemValidation{
		matcherPairs: matcherPairs,
		valueTag:     *tag.ValueTag,
		elemType:     elemT,
	})

	// This tag doesn't generate validations directly, the itemFieldValidator does.
	return Validations{}, nil
}

func (itv itemTagValidator) Docs() TagDoc {
	doc := TagDoc{
		Tag:    itv.TagName(),
		Scopes: itv.ValidScopes().UnsortedList(),
		Description: "Declares a validation for an item of a slice declared as a +k8s:listType=map. " +
			"The item to match is declared by providing field-value pair arguments. All +k8s:listMapKey fields must be included in the field-value pair arguments.",
		Usage: "+k8s:item(key: value)=<validation-tag>",
		Docs: "Arguments must be named with the JSON names of the list map key fields. " +
			"For example, if the list has +k8s:listMapKey=name, use: +k8s:item(name: myname)=<chained-validation-tag>",
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

type itemFieldValidator struct {
	validator       Validator
	listByFieldPath map[string]*listMetadata
	itemByFieldPath map[string]*itemMetadata
}

func (ifv *itemFieldValidator) Init(cfg Config) {
	ifv.validator = cfg.Validator
}

func (itemFieldValidator) Name() string {
	return "itemFieldValidator"
}

var (
	validateItemByKeyValues = types.Name{Package: libValidationPkg, Name: "ItemByKeyValues"}
)

func (ifv itemFieldValidator) GetValidations(context Context) (Validations, error) {
	itemMeta, ok := ifv.itemByFieldPath[context.Path.String()]
	if !ok || itemMeta == nil || len(itemMeta.items) == 0 {
		return Validations{}, nil
	}

	listMeta, ok := ifv.listByFieldPath[context.Path.String()]
	if !ok || !listMeta.declaredAsMap || len(listMeta.keyFields) == 0 {
		return Validations{}, fmt.Errorf("must have +k8s:listType=map and at least one '+k8s:listMapKey=...' annotation to use +k8s:item")
	}

	result := Validations{}

	for _, item := range itemMeta.items {
		// Validate that all listMapKeys are provided
		foundRequiredKeys := 0
		for _, fieldName := range listMeta.keyFields {
			for _, pair := range item.matcherPairs {
				if member := util.GetMemberByJSON(item.elemType, pair[0]); member != nil && member.Name == fieldName {
					foundRequiredKeys++
					break
				}
			}
		}
		if foundRequiredKeys != len(listMeta.keyFields) {
			return Validations{}, fmt.Errorf("item field-value pairs must include all +k8s:listMapKey fields (expected: %v)", listMeta.keyFields)
		}

		// Validate that the keys in the tag correspond to actual fields
		for _, pair := range item.matcherPairs {
			if util.GetMemberByJSON(item.elemType, pair[0]) == nil {
				return Validations{}, fmt.Errorf("list item has no field with JSON name %q", pair[0])
			}
		}

		// Extract validations from the stored tag
		subContextPath := generatePathForMap(item.matcherPairs)
		subContext := Context{
			Scope:  ScopeField,
			Type:   item.elemType,
			Parent: nil,
			Path:   context.Path.Key(subContextPath),
			Member: nil,
		}

		validations, err := ifv.validator.ExtractValidations(subContext, item.valueTag)
		if err != nil {
			return Validations{}, err
		}

		result.Variables = append(result.Variables, validations.Variables...)

		matchFn, err := createMatchFn(item.elemType, item.matcherPairs)
		if err != nil {
			return Validations{}, err
		}

		for _, vfn := range validations.Functions {
			f := Function(
				itemTagName,
				vfn.Flags,
				validateItemByKeyValues,
				matchFn,
				WrapperFunction{vfn, item.elemType},
			)
			result.Functions = append(result.Functions, f)
		}
	}

	return result, nil
}

func createMatchFn(elemT *types.Type, matcherPairs [][2]string) (FunctionLiteral, error) {
	var matchFuncBody strings.Builder
	matchFuncBody.WriteString("if item == nil { return false }\n")

	var conditions []string

	for _, pair := range matcherPairs {
		jsonKey := pair[0]
		value := pair[1]
		member := util.GetMemberByJSON(elemT, jsonKey)

		// TODO: Support all comparable primitive types (int, bool, etc.)
		// Currently only string types are supported.
		if util.NativeType(member.Type).Kind != types.Builtin || util.NativeType(member.Type) != types.String {
			return FunctionLiteral{}, fmt.Errorf("key field %q for item must be of type string or an alias to string", member.Name)
		}
		condition := fmt.Sprintf("item.%s == %q", member.Name, value)
		conditions = append(conditions, condition)
	}

	matchFuncBody.WriteString(fmt.Sprintf("return %s", strings.Join(conditions, " && ")))
	return FunctionLiteral{
		Parameters: []ParamResult{{"item", types.PointerTo(elemT)}},
		Results:    []ParamResult{{"", types.Bool}},
		Body:       matchFuncBody.String(),
	}, nil
}

func generatePathForMap(matcherPairs [][2]string) string {
	var sb strings.Builder
	for i, pair := range matcherPairs {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf("%s=%q", pair[0], pair[1]))
	}
	return sb.String()
}
