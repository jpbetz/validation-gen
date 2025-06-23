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

type itemTagValidator struct {
	validator   Validator
	byFieldPath map[string]*listMetadata
}

func (stv *itemTagValidator) Init(cfg Config) {
	stv.validator = cfg.Validator
	// Initialize byFieldPath if it's nil
	if stv.byFieldPath == nil {
		stv.byFieldPath = make(map[string]*listMetadata)
	}
}

func (itemTagValidator) TagName() string {
	return itemTagName
}

var itemTagValidScopes = sets.New(ScopeField)

func (itemTagValidator) ValidScopes() sets.Set[Scope] {
	return itemTagValidScopes
}

// LateTagValidator ensures this runs after listMapKey tags are processed
func (itemTagValidator) LateTagValidator() {}

var (
	validateItemByKeyValues = types.Name{Package: libValidationPkg, Name: "ItemByKeyValues"}
)

func (stv *itemTagValidator) GetValidations(context Context, tag codetags.Tag) (Validations, error) {
	matcherPairs, elemT, err := stv.validateAndParseTag(context, tag)
	if err != nil {
		return Validations{}, err
	}

	// Generates context path like Struct.Conditions[status="true",type="Approved"].
	subContextPath := generatePathForMap(matcherPairs)

	result := Validations{}
	subContext := Context{
		Scope:  ScopeField,
		Type:   elemT,
		Parent: context.Parent,
		Path:   context.Path.Key(subContextPath),
		Member: context.Member,
		VirtualField: &ItemExtractorField{
			listFieldName: context.Member.Name,
			elemType:      elemT,
			matcherPairs:  matcherPairs,
		},
	}

	validations, err := stv.validator.ExtractValidations(subContext, *tag.ValueTag)
	if err != nil {
		return Validations{}, err
	}

	// If chained validator uses extractor pattern, don't wrap.
	if tag.ValueTag != nil && tag.ValueTag.Name != "" {
		if ValidatorUsesExtractorPattern(tag.ValueTag.Name) {
			return validations, nil
		}
	}

	matchFn, err := createMatchFn(elemT, matcherPairs)
	if err != nil {
		return Validations{}, err
	}

	for _, vfn := range validations.Functions {
		f := Function(
			itemTagName,
			vfn.Flags,
			validateItemByKeyValues,
			matchFn,
			WrapperFunction{vfn, elemT},
		)
		result.Functions = append(result.Functions, f)
	}
	result.Variables = append(result.Variables, validations.Variables...)

	return result, nil
}

// validateAndParseTag validates the tag arguments and context, returning the matcher pairs and element type
func (stv *itemTagValidator) validateAndParseTag(context Context, tag codetags.Tag) ([][2]string, *types.Type, error) {
	matcherPairs := [][2]string{}
	processedKeys := sets.NewString()

	for _, arg := range tag.Args {
		if arg.Name == "" {
			return nil, nil, fmt.Errorf("all arguments must be named (e.g., fieldName:\"value\")")
		}
		if processedKeys.Has(arg.Name) {
			return nil, nil, fmt.Errorf("duplicate key %q in item", arg.Name)
		}
		processedKeys.Insert(arg.Name)
		matcherPairs = append(matcherPairs, [2]string{arg.Name, arg.Value})
	}

	if len(matcherPairs) == 0 {
		return nil, nil, fmt.Errorf("item requires at least one key-value pair")
	}

	// This tag can apply to value and pointer fields, as well as typedefs
	// (which should never be pointers). We need to check the concrete type.
	t := util.NonPointer(util.NativeType(context.Type))

	if t.Kind != types.Slice {
		return nil, nil, fmt.Errorf("can only be used on list types")
	}

	elemT := util.NonPointer(util.NativeType(t.Elem))
	if elemT.Kind != types.Struct {
		return nil, nil, fmt.Errorf("can only be used on list of structs")
	}

	if context.Member == nil {
		return nil, nil, fmt.Errorf("unexpected nil context member")
	}

	// Ensure the field is a list-map.
	listMap, found := stv.byFieldPath[context.Path.String()]
	if !found || !listMap.declaredAsMap || len(listMap.keyFields) == 0 {
		return nil, nil, fmt.Errorf("must have +k8s:listType=map and at least one '+k8s:listMapKey=...' annotation to use +k8s:item")
	}

	// Ensure all defined listMapKeys are provided in the tag.
	foundRequiredKeys := 0
	for _, fieldName := range listMap.keyFields {
		for _, pair := range matcherPairs {
			if member := util.GetMemberByJSON(elemT, pair[0]); member != nil && member.Name == fieldName {
				foundRequiredKeys++
				break
			}
		}
	}

	if foundRequiredKeys != len(listMap.keyFields) {
		return nil, nil, fmt.Errorf("item field-value pairs must include all +k8s:listMapKey fields (expected: %v)", listMap.keyFields)
	}

	for _, pair := range matcherPairs {
		if util.GetMemberByJSON(elemT, pair[0]) == nil {
			return nil, nil, fmt.Errorf("list item has no field with JSON name %q", pair[0])
		}
	}

	if tag.ValueType != codetags.ValueTypeTag {
		return nil, nil, fmt.Errorf("item requires a validation tag as its value payload")
	}

	if tag.ValueTag == nil {
		return nil, nil, fmt.Errorf("item requires a non-nil validation tag as its value payload")
	}

	return matcherPairs, elemT, nil
}

type ItemExtractorField struct {
	listFieldName string
	elemType      *types.Type
	matcherPairs  [][2]string
}

func (lef *ItemExtractorField) ID() string {
	return fmt.Sprintf("%s[%s]", lef.listFieldName, generatePathForMap(lef.matcherPairs))
}

func (lef *ItemExtractorField) Type() *types.Type {
	return lef.elemType
}

// GenerateExtractor creates an extractor function for the parent type
func (lef *ItemExtractorField) GenerateExtractor(parentType *types.Type) FunctionLiteral {
	var conditions []string
	for _, pair := range lef.matcherPairs {
		member := util.GetMemberByJSON(lef.elemType, pair[0])
		conditions = append(conditions, fmt.Sprintf("item.%s == %q", member.Name, pair[1]))
	}

	extractorCode := fmt.Sprintf(`func() interface{} {
		if obj != nil && obj.%s != nil {
			for _, item := range obj.%s {
				if %s {
					return true
				}
			}
		}
		return false
	}()`, lef.listFieldName, lef.listFieldName, strings.Join(conditions, " && "))

	return FunctionLiteral{
		Parameters: []ParamResult{{Name: "obj", Type: types.PointerTo(parentType)}},
		Results:    []ParamResult{{Type: types.Any}},
		Body:       fmt.Sprintf("return %s", extractorCode),
	}
}

func createMatchFn(elemT *types.Type, matcherPairs [][2]string) (FunctionLiteral, error) {
	var matchFuncBody strings.Builder
	matchFuncBody.WriteString("if item == nil { return false }\n")

	var conditions []string

	for _, pair := range matcherPairs {
		jsonKey := pair[0]
		value := pair[1]
		member := util.GetMemberByJSON(elemT, jsonKey)

		if util.NativeType(member.Type).Kind != types.Builtin || util.NativeType(member.Type) != types.String {
			return FunctionLiteral{}, fmt.Errorf("key field %q for item must be of type string or an alias to string, got %s", member.Name, member.Type.String())
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

func (stv itemTagValidator) Docs() TagDoc {
	// We don't specify Args here because the arguments are dynamic based on the list map keys.
	doc := TagDoc{
		Tag:    stv.TagName(),
		Scopes: stv.ValidScopes().UnsortedList(),
		Description: "Declares a validation for an item of a slice declared as a +k8s:listType=map. " +
			"The item to match is declared by providing field-value pair arguments. All +k8s:listMapKey=... fields must be included in the field-value pair arguments.",
		Usage: "+k8s:item(key: value)=<validation-tag>",
		Docs: "Arguments must be named with the JSON names of the list map key fields. " +
			"For example, if the list has +k8s:listMapKey=name, use: +k8s:item(name: myname)=+k8s:immutable",
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
