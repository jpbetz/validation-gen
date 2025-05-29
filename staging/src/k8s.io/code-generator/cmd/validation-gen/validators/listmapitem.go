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
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/gengo/v2/types"
)

const (
	listMapItemTagName = "k8s:listMapItem"
)

// listMapItemTagValidator init() part of each.go for
// shared usage of byFieldPath

type listMapItemTagValidator struct {
	validator   Validator
	byFieldPath map[string]*listMetadata
}

func (stv *listMapItemTagValidator) Init(cfg Config) {
	stv.validator = cfg.Validator
}

func (listMapItemTagValidator) TagName() string {
	return listMapItemTagName
}

var listMapItemTagValidScopes = sets.New(ScopeField)

func (listMapItemTagValidator) ValidScopes() sets.Set[Scope] {
	return listMapItemTagValidScopes
}

// LateTagValidator ensures this runs after listMapKey tags are processed
func (listMapItemTagValidator) LateTagValidator() {}

type parsedListMapItemKVs struct {
	MatcherPairs [][2]string
}

var (
	validateListMapItemByKeyValues = types.Name{Package: libValidationPkg, Name: "ListMapItemByKeyValues"}
)

func (stv *listMapItemTagValidator) GetValidations(context Context, args []string, payload string) (Validations, error) {
	if len(args) != 1 {
		return Validations{}, fmt.Errorf("requires exactly one arg")
	}
	parsedArg, err := parseListMapItemArg(args[0])
	if err != nil {
		return Validations{}, err
	}

	// This tag can apply to value and pointer fields, as well as typedefs
	// (which should never be pointers). We need to check the concrete type.
	t := NonPointer(NativeType(context.Type))
	fakeComments := []string{payload}

	if !(t.Kind == types.Slice) {
		return Validations{}, fmt.Errorf("can only be used on list types")
	}

	elemT := NonPointer(NativeType(t.Elem))
	if elemT.Kind != types.Struct {
		return Validations{}, fmt.Errorf("can only be used on list of structs")
	}

	if context.Member == nil {
		return Validations{}, fmt.Errorf("unexpected nil context member")
	}

	listMap, found := stv.byFieldPath[context.Path.String()]
	if !found || !listMap.declaredAsMap || len(listMap.keyFields) == 0 {
		return Validations{}, fmt.Errorf("must have +k8s:listType=map and '+k8s:listMapKey=...' annotations")
	}

	foundRequiredKeys := 0
	for _, fieldName := range listMap.keyFields {
		for _, pair := range parsedArg.MatcherPairs {
			if member := getMemberByJSON(elemT, pair[0]); member != nil && member.Name == fieldName {
				foundRequiredKeys++
				break
			}
		}
	}

	if foundRequiredKeys != len(listMap.keyFields) {
		return Validations{}, fmt.Errorf("listMapItem field-value pairs must include all +k8s:listMapKey fields. ")
	}

	for _, pair := range parsedArg.MatcherPairs {
		if getMemberByJSON(elemT, pair[0]) == nil {
			return Validations{}, fmt.Errorf("list item has has no field with JSON name %q", pair[0])
		}
	}

	// Generates context path like Struct.Conditions[status="true",type="Approved"].
	subContextPath := generatePathForMap(parsedArg.MatcherPairs)
	fakeMember := createFakeMember(elemT, parsedArg.MatcherPairs)

	subContext := Context{
		Member: fakeMember,
		Scope:  ScopeField,
		Type:   elemT,
		// TODO(aaron-prindle) for +k8s:unionMember support need to plumb this.
		Parent: nil,
		Path:   context.Path.Key(subContextPath),
	}

	if validations, err := stv.validator.ExtractValidations(subContext, fakeComments); err != nil {
		return Validations{}, err
	} else {

		result := Validations{}
		result.Variables = append(result.Variables, validations.Variables...)

		matchFn, err := createMatchFn(elemT, parsedArg.MatcherPairs)
		if err != nil {
			return Validations{}, err
		}

		for _, vfn := range validations.Functions {
			f := Function(
				listMapItemTagName,
				vfn.Flags,
				validateListMapItemByKeyValues,
				matchFn,
				WrapperFunction{vfn, elemT},
			)
			result.Functions = append(result.Functions, f)
		}
		return result, nil

	}
}

func parseListMapItemArg(argStr string) (*parsedListMapItemKVs, error) {
	var matcherPairs [][2]string
	// Remove backticks from raw string arg.
	argStr = strings.Trim(argStr, "`")
	if err := json.Unmarshal([]byte(argStr), &matcherPairs); err == nil {
		if len(matcherPairs) == 0 {
			return nil, fmt.Errorf("listMapItem matcher pairs cannot be empty")
		}

		for i, pair := range matcherPairs {
			if len(pair) != 2 {
				return nil, fmt.Errorf("listMapItem pair at index %d must have exactly 2 elements", i)
			}
		}
		// Sort by key for consistent output
		sort.Slice(matcherPairs, func(i, j int) bool {
			return matcherPairs[i][0] < matcherPairs[j][0]
		})
		return &parsedListMapItemKVs{
			MatcherPairs: matcherPairs,
		}, nil
	}
	return nil, fmt.Errorf("listMapItem arguments incorrect, JSON parsing failed")
}

func createMatchFn(elemT *types.Type, matcherPairs [][2]string) (FunctionLiteral, error) {
	var matchFuncBody strings.Builder
	matchFuncBody.WriteString("if item == nil { return false }\n")

	var conditions []string

	for _, pair := range matcherPairs {
		jsonKey := pair[0]
		value := pair[1]
		member := getMemberByJSON(elemT, jsonKey)

		var condition string
		if NativeType(member.Type).Kind != types.Builtin {
			return FunctionLiteral{}, fmt.Errorf("key field %q for listMapItem must be of type string or an alias to string, got %s", member.Name, member.Type.String())
		}
		condition = fmt.Sprintf("item.%s == %q", member.Name, value)
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

func createFakeMember(itemType *types.Type, matcherPairs [][2]string) *types.Member {
	var keyParts []string
	for _, pair := range matcherPairs {
		keyParts = append(keyParts, fmt.Sprintf("%s=%s", pair[0], pair[1]))
	}
	memberName := fmt.Sprintf("_listItem[%s]", strings.Join(keyParts, ","))

	fakeMember := &types.Member{
		Name:         memberName,
		Type:         itemType,
		Embedded:     false,
		CommentLines: []string{},
		Tags:         "",
	}

	return fakeMember
}

func (stv listMapItemTagValidator) Docs() TagDoc {
	doc := TagDoc{
		Tag:    stv.TagName(),
		Scopes: stv.ValidScopes().UnsortedList(),
		Description: "Declares a validation for an item of a slice declared as a +k8s:listType=map." +
			"The item to match is declared by providing field-value pair arguments. All +k8s:listMapKey fields must be included in the field-value pair arguments.",
		Args: []TagArgDoc{
			{
				Description: `[["<list-map-key-field-json-name>","<value>"], ["<list-map-key-field-json-name>", "<value>"], ...]`,
			},
		},
		Payloads: []TagPayloadDoc{{
			Description: "<validation-tag>",
			Docs:        "The tag to evaluate for the matching list item.",
		}},
	}
	return doc
}
