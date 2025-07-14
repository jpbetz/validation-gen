/*
Copyright 2021 The Kubernetes Authors.

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
	"regexp"
	"slices"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/code-generator/cmd/validation-gen/util"
	"k8s.io/gengo/v2/codetags"
	"k8s.io/gengo/v2/parser/tags"
	"k8s.io/gengo/v2/types"
)

var discriminatedUnionValidator = types.Name{Package: libValidationPkg, Name: "DiscriminatedUnion"}
var unionValidator = types.Name{Package: libValidationPkg, Name: "Union"}

var newDiscriminatedUnionMembership = types.Name{Package: libValidationPkg, Name: "NewDiscriminatedUnionMembership"}
var newUnionMembership = types.Name{Package: libValidationPkg, Name: "NewUnionMembership"}
var unionVariablePrefix = "unionMembershipFor"

func init() {
	// Unions are comprised of multiple tags, which need to share information
	// between them.  The tags are on struct fields, but the validation
	// actually pertains to the struct itself.
	shared := map[string]unions{}
	RegisterTypeValidator(unionTypeOrFieldValidator{shared})
	RegisterFieldValidator(unionTypeOrFieldValidator{shared})
	RegisterTagValidator(unionDiscriminatorTagValidator{shared})
	RegisterTagValidator(unionMemberTagValidator{shared})
}

type unionTypeOrFieldValidator struct {
	shared map[string]unions
}

func (unionTypeOrFieldValidator) Init(_ Config) {}

func (unionTypeOrFieldValidator) Name() string {
	return "unionTypeOrFieldValidator"
}

func (utfv unionTypeOrFieldValidator) GetValidations(context Context) (Validations, error) {
	// Gengo does not treat struct definitions as aliases, which is
	// inconsistent but unlikely to change. That means we don't REALLY need to
	// handle it here, but let's be extra careful and extract the most concrete
	// type possible.
	if k := util.NonPointer(util.NativeType(context.Type)).Kind; k != types.Struct && k != types.Slice {
		return Validations{}, nil
	}

	unions := utfv.shared[context.Path.String()]
	if len(unions) == 0 {
		return Validations{}, nil
	}

	return processUnionValidations(context, unions, unionVariablePrefix,
		unionMemberTagName, unionValidator, discriminatedUnionValidator)
}

func toSliceAny[T any](t []T) []any {
	result := make([]any, len(t))
	for i, v := range t {
		result[i] = v
	}
	return result
}

const (
	unionDiscriminatorTagName = "k8s:unionDiscriminator"
	unionMemberTagName        = "k8s:unionMember"
)

type unionDiscriminatorTagValidator struct {
	shared map[string]unions
}

func (unionDiscriminatorTagValidator) Init(_ Config) {}

func (unionDiscriminatorTagValidator) TagName() string {
	return unionDiscriminatorTagName
}

// Shared between unionDiscriminatorTagValidator and unionMemberTagValidator.
var unionTagValidScopes = sets.New(ScopeField, ScopeListVal)

func (unionDiscriminatorTagValidator) ValidScopes() sets.Set[Scope] {
	return unionTagValidScopes
}

func (udtv unionDiscriminatorTagValidator) GetValidations(context Context, tag codetags.Tag) (Validations, error) {
	err := processDiscriminatorValidations(udtv.shared, context, tag)
	if err != nil {
		return Validations{}, err
	}
	// This tag does not actually emit any validations, it just accumulates
	// information. The validation is done by the unionTypeOrFieldValidator.
	return Validations{}, nil
}

func (udtv unionDiscriminatorTagValidator) Docs() TagDoc {
	return TagDoc{
		Tag:         udtv.TagName(),
		Scopes:      udtv.ValidScopes().UnsortedList(),
		Description: "Indicates that this field is the discriminator for a union.",
		Args: []TagArgDoc{{
			Name:        "union",
			Description: "<string>",
			Docs:        "the name of the union, if more than one exists",
			Type:        codetags.ArgTypeString,
		}},
	}
}

type unionMemberTagValidator struct {
	shared map[string]unions
}

func (unionMemberTagValidator) Init(_ Config) {}

func (unionMemberTagValidator) TagName() string {
	return unionMemberTagName
}

func (unionMemberTagValidator) ValidScopes() sets.Set[Scope] {
	return unionTagValidScopes
}

func (umtv unionMemberTagValidator) GetValidations(context Context, tag codetags.Tag) (Validations, error) {
	err := processMemberValidations(umtv.shared, context, tag)
	if err != nil {
		return Validations{}, err
	}
	// This tag does not actually emit any validations, it just accumulates
	// information. The validation is done by the unionTypeOrFieldValidator.
	return Validations{}, nil
}

func (umtv unionMemberTagValidator) Docs() TagDoc {
	return TagDoc{
		Tag:         umtv.TagName(),
		Scopes:      umtv.ValidScopes().UnsortedList(),
		Description: "Indicates that this field is a member of a union.",
		Args: []TagArgDoc{{
			Name:        "union",
			Description: "<string>",
			Docs:        "the name of the union, if more than one exists",
			Type:        codetags.ArgTypeString,
		}, {
			Name:        "memberName",
			Description: "<string>",
			Docs:        "the discriminator value for this member",
			Default:     "the field's name",
			Type:        codetags.ArgTypeString,
		}},
	}
}

// union defines how a union validation will be generated, based
// on +k8s:unionMember and +k8s:unionDiscriminator tags found in a go struct.
type union struct {
	// fields provides field information about all the members of the union.
	// Each item provides a fieldName and memberName pair, where [0] identifies
	// the field name and [1] identifies the union member Name. fields is index
	// aligned with fieldMembers.
	// If member name is not set, it defaults to the go struct field name.
	fields [][2]string
	// fieldMembers describes all the members of the union.
	fieldMembers []*types.Member

	// discriminator is the name of the discriminator field
	discriminator *string
	// discriminatorMember describes the discriminator field.
	discriminatorMember *types.Member

	// itemMatchers stores matcher criteria for list item unions.
	// key is the path, value is the matcher map.
	itemMatchers map[string]map[string]any
}

// unions represents all the unions for a go struct.
type unions map[string]*union

// getOrCreate gets a union by name, or initializes a new union by the given name.
func (us unions) getOrCreate(name string) *union {
	var u *union
	var ok bool
	if u, ok = us[name]; !ok {
		u = &union{
			itemMatchers: make(map[string]map[string]any),
		}
		us[name] = u
	}
	return u
}

func processUnionValidations(context Context, unions unions, varPrefix string,
	tagName string, undiscriminatedValidator types.Name, discriminatedValidator types.Name,
) (Validations, error) {
	result := Validations{}

	// Sort the keys for stable output.
	keys := make([]string, 0, len(unions))
	for k := range unions {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, unionName := range keys {
		u := unions[unionName]
		if len(u.fieldMembers) > 0 || u.discriminator != nil || len(u.itemMatchers) > 0 {
			// TODO: Avoid the "local" here. This was added to to avoid errors caused when the package is an empty string.
			//       The correct package would be the output package but is not known here. This does not show up in generated code.
			// TODO: Append a consistent hash suffix to avoid generated name conflicts?
			varBaseName := sanitizeName(context.Path.String() + "_" + unionName)
			supportVarName := PrivateVar{Name: varPrefix + "_" + varBaseName, Package: "local"}

			var extractorArgs []any
			ptrType := types.PointerTo(context.Type)

			// Handle field unions
			for _, member := range u.fieldMembers {
				extractor := createMemberExtractor(ptrType, member)
				extractorArgs = append(extractorArgs, extractor)
			}

			// Handle list item unions for lists
			if context.Type.Kind == types.Slice && len(u.itemMatchers) > 0 {
				elemType := util.NonPointer(context.Type.Elem)

				// Sort matcher paths for stable output
				matcherPaths := make([]string, 0, len(u.itemMatchers))
				for path := range u.itemMatchers {
					matcherPaths = append(matcherPaths, path)
				}
				slices.Sort(matcherPaths)

				for _, fullPath := range matcherPaths {
					matcher := u.itemMatchers[fullPath]
					extractor, err := generateItemExtractor(context.Type, elemType, matcher)
					if err != nil {
						return Validations{}, err
					}
					extractorArgs = append(extractorArgs, extractor)
				}
			}

			if u.discriminator != nil {
				supportVar := Variable(supportVarName,
					Function(tagName, DefaultFlags, newDiscriminatedUnionMembership,
						append([]any{*u.discriminator}, toSliceAny(getDisplayFields(u, context))...)...))
				result.Variables = append(result.Variables, supportVar)

				discriminatorExtractor := FunctionLiteral{
					Parameters: []ParamResult{{Name: "obj", Type: ptrType}},
					Results:    []ParamResult{{Type: types.String}},
					Body:       fmt.Sprintf("if obj == nil {return \"\"}; return string(obj.%s)", u.discriminatorMember.Name), // Cast to string
				}

				extraArgs := append([]any{supportVarName, discriminatorExtractor}, extractorArgs...)
				fn := Function(tagName, DefaultFlags, discriminatedValidator, extraArgs...)
				result.Functions = append(result.Functions, fn)
			} else {
				supportVar := Variable(supportVarName, Function(tagName, DefaultFlags, newUnionMembership, toSliceAny(getDisplayFields(u, context))...))
				result.Variables = append(result.Variables, supportVar)

				extraArgs := append([]any{supportVarName}, extractorArgs...)
				fn := Function(tagName, DefaultFlags, undiscriminatedValidator, extraArgs...)
				result.Functions = append(result.Functions, fn)
			}
		}
	}

	return result, nil
}

func createMemberExtractor(ptrType *types.Type, member *types.Member) FunctionLiteral {
	extractor := FunctionLiteral{
		Parameters: []ParamResult{{Name: "obj", Type: ptrType}},
		Results:    []ParamResult{{Type: types.Bool}},
	}
	nt := util.NativeType(member.Type)
	switch nt.Kind {
	case types.Pointer, types.Map, types.Slice:
		extractor.Body = fmt.Sprintf("if obj == nil {return false}; return obj.%s != nil", member.Name)
	case types.Builtin:
		extractor.Body = fmt.Sprintf("if obj == nil {return false}; var z %s; return obj.%s != z", member.Type, member.Name)
	default:
		// This should be caught before we get here, but JIC.
		extractor.Body = fmt.Sprintf("if obj == nil {return false}; return false /* unsupported union member kind: %s */", nt.Kind)
	}
	return extractor
}

// generateItemExtractor creates an extractor function for list item union members.
// It generates code that uses validate.SliceItem to check if an item matching
// the criteria exists in the list.
func generateItemExtractor(listType *types.Type, elemType *types.Type, matcher map[string]any) (FunctionLiteral, error) {
	// Build matcher conditions
	var conditions []string
	for key, value := range matcher {
		member := util.GetMemberByJSON(elemType, key)
		if member == nil {
			return FunctionLiteral{}, fmt.Errorf("struct %s has no field with JSON name %q", elemType, key)
		}
		var condition string
		switch v := value.(type) {
		case string:
			condition = fmt.Sprintf("item.%s == %q", member.Name, v)
		case int:
			condition = fmt.Sprintf("item.%s == %d", member.Name, v)
		case bool:
			condition = fmt.Sprintf("item.%s == %t", member.Name, v)
		default:
			condition = fmt.Sprintf("item.%s == %v", member.Name, v)
		}
		conditions = append(conditions, condition)
	}

	extractor := FunctionLiteral{
		Parameters: []ParamResult{{Name: "list", Type: listType}},
		Results:    []ParamResult{{Type: types.Bool}},
	}

	// Build the function body
	body := fmt.Sprintf(`var matched *%s`, elemType.Name.Name) + "\n"
	body += fmt.Sprintf(`validate.SliceItem(ctx, op, fldPath, list, nil, func(item *%s) bool { return %s }, `,
		elemType.Name.Name, strings.Join(conditions, " && "))

	if util.IsDirectComparable(elemType) {
		body += "validate.DirectEqual"
	} else {
		body += "validate.SemanticDeepEqual"
	}

	body += fmt.Sprintf(`, func(ctx context.Context, op operation.Operation, itemPath *field.Path, newItem, oldItem *%s) field.ErrorList { matched = newItem; return nil })`, elemType.Name.Name) + "\n"
	body += "return matched != nil"

	extractor.Body = body
	return extractor, nil
}

func processDiscriminatorValidations(shared map[string]unions, context Context, tag codetags.Tag) error {
	// This tag can apply to value and pointer fields, as well as typedefs
	// (which should never be pointers). We need to check the concrete type.
	if t := util.NonPointer(util.NativeType(context.Type)); t != types.String {
		return fmt.Errorf("can only be used on string types (%s)", rootTypeString(context.Type, t))
	}
	if shared[context.ParentPath.String()] == nil {
		shared[context.ParentPath.String()] = unions{}
	}
	unionArg, _ := tag.NamedArg("union") // optional
	u := shared[context.ParentPath.String()].getOrCreate(unionArg.Value)

	var discriminatorFieldName string
	if jsonAnnotation, ok := tags.LookupJSON(*context.Member); ok {
		discriminatorFieldName = jsonAnnotation.Name
		u.discriminator = &discriminatorFieldName
		u.discriminatorMember = context.Member
	}

	return nil
}

func processMemberValidations(shared map[string]unions, context Context, tag codetags.Tag) error {
	var fieldName string
	var unionArg codetags.Arg

	unionArg, _ = tag.NamedArg("union") // optional

	if context.Scope == ScopeListVal {
		if context.Parent != nil && context.Parent.Kind == types.Alias {
			return fmt.Errorf("list item union members are not supported on typedef types")
		}

		if context.Path == nil {
			return fmt.Errorf("no path for list val union member")
		}
		fieldName = context.Path.String() // eg: "<path>/Pipeline.Tasks[{"name": "succeeded"}]"
	} else {
		nt := util.NativeType(context.Member.Type)
		switch nt.Kind {
		case types.Pointer, types.Map, types.Slice, types.Builtin:
			// OK
		default:
			// In particular non-pointer structs are not supported.
			return fmt.Errorf("can only be used on nilable and primitive types (%s)", nt.Kind)
		}

		jsonTag, ok := tags.LookupJSON(*context.Member)
		if !ok {
			return fmt.Errorf("field %q is a union member but has no JSON struct field tag", context.Member)
		}
		fieldName = jsonTag.Name
		if len(fieldName) == 0 {
			return fmt.Errorf("field %q is a union member but has no JSON name", context.Member)
		}
	}

	if shared[context.ParentPath.String()] == nil {
		shared[context.ParentPath.String()] = unions{}
	}

	var memberName string
	if memberNameArg, ok := tag.NamedArg("memberName"); ok { // optional
		memberName = memberNameArg.Value
	} else if context.Scope != ScopeListVal {
		memberName = context.Member.Name // default
	}

	u := shared[context.ParentPath.String()].getOrCreate(unionArg.Value)
	u.fields = append(u.fields, [2]string{fieldName, memberName})

	if context.Scope == ScopeListVal {
		matcher, err := extractMatcherFromPath(fieldName)
		if err != nil {
			return fmt.Errorf("failed to extract matcher from path %s: %w", fieldName, err)
		}
		u.itemMatchers[fieldName] = matcher
	} else {
		u.fieldMembers = append(u.fieldMembers, context.Member)
	}

	return nil
}

// getDisplayFields formats union field names for user-friendly error messages.
// For list item unions, it converts paths like "<path>/Pipeline.Tasks[{\"name\": \"succeeded\"}]"
// to readable formats like "Tasks[{\"name\": \"succeeded\"}]".
func getDisplayFields(u *union, context Context) [][2]string {
	displayFields := make([][2]string, len(u.fields))
	listFieldName := context.Path.String()
	pathParts := strings.Split(listFieldName, ".")
	if len(pathParts) > 0 {
		listFieldName = pathParts[len(pathParts)-1]
	}
	for i, f := range u.fields {
		fieldName := f[0]
		memberName := f[1]
		if _, isItem := u.itemMatchers[fieldName]; isItem {
			// Extract the JSON part from the input
			bracketIndex := strings.Index(fieldName, "[")
			if bracketIndex != -1 {
				jsonPart := fieldName[bracketIndex:]
				fieldName = listFieldName + jsonPart
			}
		}
		displayFields[i] = [2]string{fieldName, memberName}
	}
	return displayFields
}

// sanitizeName converts a string into a valid Go identifier
func sanitizeName(name string) string {
	name = strings.ReplaceAll(name, ".", "_")
	re := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	return re.ReplaceAllString(name, "_")
}

// extractMatcherFromPath extracts the matcher criteria from a path like "Pipeline.Tasks[{"name": "succeeded"}]"
func extractMatcherFromPath(path string) (map[string]any, error) {
	re := regexp.MustCompile(`\[({.*?})\]`)
	matches := re.FindStringSubmatch(path)
	if len(matches) < 2 {
		return nil, fmt.Errorf("no matcher criteria found in path")
	}

	var matcher map[string]any
	if err := json.Unmarshal([]byte(matches[1]), &matcher); err != nil {
		return nil, fmt.Errorf("failed to parse matcher JSON: %w", err)
	}
	return matcher, nil
}
