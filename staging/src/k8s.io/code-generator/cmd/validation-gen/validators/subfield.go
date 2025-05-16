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
	"k8s.io/gengo/v2/parser/tags"
	"k8s.io/gengo/v2/types"
)

const (
	subfieldTagName = "k8s:subfield"
)

func init() {
	RegisterTagValidator(&subfieldTagValidator{})
}

type subfieldTagValidator struct {
	validator Validator
}

func (stv *subfieldTagValidator) Init(cfg Config) {
	stv.validator = cfg.Validator
}

func (subfieldTagValidator) TagName() string {
	return subfieldTagName
}

var subfieldTagValidScopes = sets.New(ScopeAny)

func (subfieldTagValidator) ValidScopes() sets.Set[Scope] {
	return subfieldTagValidScopes
}

type parsedSubfieldArg struct {
	MatcherMap map[string]string
	SubName    string
}

var (
	validateSubfield            = types.Name{Package: libValidationPkg, Name: "Subfield"}
	validateListMapElementByKey = types.Name{Package: libValidationPkg, Name: "ListMapElementByKey"}
)

func (stv *subfieldTagValidator) GetValidations(context Context, args []string, payload string) (Validations, error) {
	if len(args) != 1 {
		return Validations{}, fmt.Errorf("requires exactly one arg")
	}
	configStr := args[0]

	parsedArg, err := parseSubfieldArg(configStr)
	if err != nil {
		return Validations{}, err
	}

	// This tag can apply to value and pointer fields, as well as typedefs
	// (which should never be pointers). We need to check the concrete type.
	t := nonPointer(nativeType(context.Type))
	fakeComments := []string{payload}

	if parsedArg.SubName != "" { // +k8s:subfield(subname) usage
		if t.Kind != types.Struct {
			return Validations{}, fmt.Errorf("can only be used on struct types")
		}

		subname := parsedArg.SubName
		submemb := getMemberByJSON(t, subname)
		if submemb == nil {
			return Validations{}, fmt.Errorf("no field for json name %q", subname)
		}

		result := Validations{}

		subContext := Context{
			Scope:  ScopeField,
			Type:   submemb.Type,
			Parent: t,
			Path:   context.Path.Child(subname),
			Member: submemb,
		}
		if validations, err := stv.validator.ExtractValidations(subContext, fakeComments); err != nil {
			return Validations{}, err
		} else {
			if len(validations.Variables) > 0 {
				return Validations{}, fmt.Errorf("variable generation is not supported")
			}

			for _, vfn := range validations.Functions {
				nilableStructType := context.Type
				if !isNilableType(nilableStructType) {
					nilableStructType = types.PointerTo(nilableStructType)
				}
				nilableFieldType := submemb.Type
				fieldExprPrefix := ""
				if !isNilableType(nilableFieldType) {
					nilableFieldType = types.PointerTo(nilableFieldType)
					fieldExprPrefix = "&"
				}

				getFn := FunctionLiteral{
					Parameters: []ParamResult{{"o", nilableStructType}},
					Results:    []ParamResult{{"", nilableFieldType}},
				}
				getFn.Body = fmt.Sprintf("return %so.%s", fieldExprPrefix, submemb.Name)
				f := Function(subfieldTagName, vfn.Flags, validateSubfield, subname, getFn, WrapperFunction{vfn, submemb.Type})
				result.Functions = append(result.Functions, f)
				result.Variables = append(result.Variables, validations.Variables...)
			}
		}
		return result, nil

	} else { // +k8s:subfield({"listMapElems":{"..."}}) usage
		if t.Kind != types.Slice && t.Kind != types.Array {
			return Validations{}, fmt.Errorf("can only be used on list types")
		}

		elemT := nonPointer(nativeType(t.Elem))
		if elemT.Kind != types.Struct {
			return Validations{}, fmt.Errorf("can only be used on list of structs")
		}

		if context.Member == nil {
			// TODO(aaron-prindle) more descriptive error would be helpful here
			// currently not sure what cases this occurs for
			return Validations{}, fmt.Errorf("unexpected nil context member")
		}
		listMapKey := parseListMapKey(context.Member.CommentLines)
		if len(listMapKey) == 0 {
			return Validations{}, fmt.Errorf("must have '+listType=map' and '+listMapKey=...' annotations to use subfield with matcher map")
		}
		if _, ok := parsedArg.MatcherMap[listMapKey]; !ok {
			return Validations{}, fmt.Errorf("subfield matcher map must contain listMap key")
		}

		for k := range parsedArg.MatcherMap {
			if getMemberByJSON(elemT, k) == nil {
				return Validations{}, fmt.Errorf("element type %s has no field with JSON name %q", elemT.Name.String(), k)
			}
		}

		// generates context path like Struct.Conditions[status="true",type="Approved"]
		subContextPath := generatePathForMap(parsedArg.MatcherMap)
		subContext := Context{
			Scope: ScopeField,
			Type:  elemT,
			// TODO(aaron-prindle) for +k8s:unionMember support need to plumb this.
			Parent: nil,
			Path:   context.Path.Key(subContextPath),
			// TODO(aaron-prindle) for +k8s:unionMember support need to plumb this.
			Member: nil,
		}

		if validations, err := stv.validator.ExtractValidations(subContext, fakeComments); err != nil {
			return Validations{}, err
		} else {

			result := Validations{}
			result.Variables = append(result.Variables, validations.Variables...)

			matchFn, err := createMatchFn(elemT, parsedArg.MatcherMap)
			if err != nil {
				return Validations{}, err
			}

			for _, vfn := range validations.Functions {
				f := Function(
					subfieldTagName,
					vfn.Flags,
					validateListMapElementByKey,
					matchFn,
					WrapperFunction{vfn, elemT},
				)
				result.Functions = append(result.Functions, f)
			}
			return result, nil
		}

	}
}

func parseSubfieldArg(argStr string) (*parsedSubfieldArg, error) {
	var matcherMap map[string]string
	if err := json.Unmarshal([]byte(argStr), &matcherMap); err == nil {
		if len(matcherMap) == 0 {
			return nil, fmt.Errorf("subfield JSON matcher map cannot be empty")
		}

		return &parsedSubfieldArg{
			MatcherMap: matcherMap,
		}, nil
	}
	if argStr == "" {
		return nil, fmt.Errorf("arg cannot be an empty string")
	}
	return &parsedSubfieldArg{
		SubName: argStr,
	}, nil
}

func createMatchFn(elemT *types.Type, listElems map[string]string) (FunctionLiteral, error) {
	var matchFuncBody strings.Builder
	matchFuncBody.WriteString("if item == nil { return false }\n")

	var conditions []string
	sortedKeys := make([]string, 0, len(listElems))
	for k := range listElems {
		sortedKeys = append(sortedKeys, k)
	}
	// Sort keys so that generated code is consistent.
	sort.Strings(sortedKeys)

	for _, jsonKey := range sortedKeys {
		fieldname, err := getFieldNameFromJSONKey(elemT, jsonKey)
		if err != nil {
			return FunctionLiteral{}, err
		}

		fieldMember, err := getMember(elemT, fieldname)
		if err != nil {
			return FunctionLiteral{}, err
		}

		// TODO(aaron-prindle) support additional builtin types
		var condition string
		if fieldMember.Type.Kind == types.Pointer && fieldMember.Type.Elem != nil &&
			fieldMember.Type.Elem.Name.Name == "string" {
			condition = fmt.Sprintf("(item.%s != nil && *item.%s == %q)", fieldname, fieldname, listElems[jsonKey])
		} else if fieldMember.Type.Name.Name == "string" {
			condition = fmt.Sprintf("item.%s == %q", fieldname, listElems[jsonKey])
		} else {
			return FunctionLiteral{}, fmt.Errorf("must be a string or *string, but got type %s", fieldMember.Type.String())
		}
		conditions = append(conditions, condition)
	}

	matchFuncBody.WriteString(fmt.Sprintf("return %s", strings.Join(conditions, " && ")))
	return FunctionLiteral{
		Parameters: []ParamResult{{"item", types.PointerTo(elemT)}},
		Results:    []ParamResult{{"", types.Bool}},
		Body:       matchFuncBody.String(),
	}, nil
}

func getMember(s *types.Type, fieldname string) (types.Member, error) {
	// Assumes 's' is non-pointer struct.
	for _, m := range s.Members {
		if m.Name == fieldname {
			return m, nil
		}
	}
	return types.Member{}, fmt.Errorf("no member in type %s for fieldname %s", s.Kind, fieldname)
}

func getFieldNameFromJSONKey(s *types.Type, jsonKey string) (string, error) {
	// Assumes 's' is non-pointer struct.
	for _, m := range s.Members {
		// Default JSON name is field name if no 'json' tag.
		JSONName := m.Name
		jsonAnnotation, ok := tags.LookupJSON(m)
		if ok && jsonAnnotation.Name != "" {
			JSONName = jsonAnnotation.Name
		}
		if JSONName == jsonKey {
			return m.Name, nil
		}
	}
	return "", fmt.Errorf("no field with JSON name %q in type %s", jsonKey, s.Name.String())
}

func parseListMapKey(commentLines []string) string {
	var key string
	hasListTypeMap := false

	for _, line := range commentLines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "+listType=map" {
			hasListTypeMap = true
		}
		if strings.HasPrefix(trimmedLine, "+listMapKey=") {
			keyPart := strings.TrimPrefix(trimmedLine, "+listMapKey=")
			trimmedKey := strings.TrimSpace(keyPart)
			if trimmedKey != "" {
				key = trimmedKey
			}
		}
	}

	if hasListTypeMap && key != "" {
		return key
	}
	return ""
}

func generatePathForMap(keyValues map[string]string) string {
	keys := make([]string, 0, len(keyValues))
	for k := range keyValues {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf("%s=%q", k, keyValues[k]))
	}
	return sb.String()
}

func (stv subfieldTagValidator) Docs() TagDoc {
	doc := TagDoc{
		Tag:    stv.TagName(),
		Scopes: stv.ValidScopes().UnsortedList(),
		Description: "Declares a validation for a subfield of a struct or a select struct in a list with listType=map" +
			"matching specified criteria. When used for list-map elements, one of the match criteria must be on value of the listMapKey",
		Args: []TagArgDoc{
			{
				Description: "<field-json-name>",
			},
			{
				Description: `{"<jsonKey>":"<value>", ...}`,
			},
		},
		Docs: "When used with `<field-json-name>`, the named subfield must be a direct field. " +
			"When used with a JSON map, it applies to elements of a list of structs.",
		Payloads: []TagPayloadDoc{{
			Description: "<validation-tag>",
			Docs:        "The tag to evaluate for the subfield or selected list elements.",
		}},
	}
	return doc
}
