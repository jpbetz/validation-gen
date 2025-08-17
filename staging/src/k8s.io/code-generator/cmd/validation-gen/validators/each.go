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
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/code-generator/cmd/validation-gen/util"
	"k8s.io/gengo/v2/codetags"
	"k8s.io/gengo/v2/types"
)

const (
	listTypeTagName   = "k8s:listType"
	ListMapKeyTagName = "k8s:listMapKey"
	uniqueTagName     = "k8s:unique"
	eachValTagName    = "k8s:eachVal"
	eachKeyTagName    = "k8s:eachKey"
)

// We keep the eachVal and eachKey validators around because the main
// code-generation logic calls them directly.  We could move them into the main
// pkg, but it's easier and cleaner to leave them here.
var globalEachVal *eachValTagValidator
var globalEachKey *eachKeyTagValidator

func init() {
	// Lists with list-map semantics are comprised of multiple tags, which need
	// to share metadata about the list between them.
	listMeta := map[string]*listMetadata{} // keyed by the field or type path

	// Accumulate list metadata via tags.
	RegisterTagValidator(listTypeTagValidator{byPath: listMeta})
	RegisterTagValidator(listMapKeyTagValidator{byPath: listMeta})
	RegisterTagValidator(uniqueTagValidator{byPath: listMeta})

	// Finish work on the accumulated list metadata.
	RegisterFieldValidator(listValidator{byPath: listMeta})
	RegisterTypeValidator(listValidator{byPath: listMeta})

	// Processing item tags requires the list metadata.
	RegisterTagValidator(&itemTagValidator{listByPath: listMeta})

	// Iterating values of lists and maps is a special tag, which can be called
	// directly by the code-generator logic.
	globalEachVal = &eachValTagValidator{byPath: listMeta, validator: nil}
	RegisterTagValidator(globalEachVal)

	// Iterating keys of maps is a special tag, which can be called directly by
	// the code-generator logic.
	globalEachKey = &eachKeyTagValidator{validator: nil}
	RegisterTagValidator(globalEachKey)
}

// This applies to all tags in this file.
var listTagsValidScopes = sets.New(ScopeType, ScopeField, ScopeListVal, ScopeMapKey, ScopeMapVal)

type listSemantic string

// Known list semantics.
const (
	listMap    listSemantic = "map"
	listSet    listSemantic = "set"
	listAtomic listSemantic = "atomic"
)

// listMetadata collects information about a single list with map or set semantics.
type listMetadata struct {
	// These will be checked for correctness elsewhere.
	semantic  listSemantic
	unique    listSemantic // if set, overrides semantic for uniqueness
	keyFields []string     // iff semantic == listMap or unique == listMap
	keyNames  []string     // iff semantic == listMap or unique == listMap
}

// makeListMapMatchFunc generates a function that compares two list-map
// elements by their list-map key fields.
func (lm *listMetadata) makeListMapMatchFunc(t *types.Type) FunctionLiteral {
	if lm.semantic != listMap && lm.unique != listMap {
		panic("makeListMapMatchFunc called on a non-map list")
	}
	// If no keys are defined, we will throw a good error later.

	matchFn := FunctionLiteral{
		Parameters: []ParamResult{{"a", t}, {"b", t}},
		Results:    []ParamResult{{"", types.Bool}},
	}
	buf := strings.Builder{}
	buf.WriteString("return ")
	// Note: this does not handle pointer fields, which are not
	// supposed to be used as listMap keys.
	for i, fld := range lm.keyFields {
		if i > 0 {
			buf.WriteString(" && ")
		}
		buf.WriteString(fmt.Sprintf("a.%s == b.%s", fld, fld))
	}
	matchFn.Body = buf.String()
	return matchFn
}

type listTypeTagValidator struct {
	byPath map[string]*listMetadata
}

func (listTypeTagValidator) Init(Config) {}

func (listTypeTagValidator) TagName() string {
	return listTypeTagName
}

func (listTypeTagValidator) ValidScopes() sets.Set[Scope] {
	return listTagsValidScopes
}

func (lttv listTypeTagValidator) GetValidations(context Context, tag codetags.Tag) (Validations, error) {
	// NOTE: pointers to lists are not supported, so we should never see a pointer here.
	t := util.NativeType(context.Type)
	if t.Kind != types.Slice && t.Kind != types.Array {
		return Validations{}, fmt.Errorf("can only be used on list types (%s)", t.Kind)
	}

	lm := lttv.byPath[context.Path.String()]
	if lm == nil {
		lm = &listMetadata{}
		lttv.byPath[context.Path.String()] = lm
	}
	if lm.semantic != "" {
		return Validations{}, fmt.Errorf("list was already declared as %q", lm.semantic)
	}

	switch tag.Value {
	case "atomic":
		// We don't do much with atomic, but this ensures no conflicts between
		// tags on typedefs and tags on fields which use those typedefs.
		lm.semantic = listAtomic
	case "set":
		lm.semantic = listSet
		// NOTE: we validate uniqueness in the listValidator.
	case "map":
		// NOTE: maps of pointers are not supported, so we should never see a pointer here.
		if util.NativeType(t.Elem).Kind != types.Struct {
			return Validations{}, fmt.Errorf("only lists of structs can be list-maps")
		}

		// Save the fact that this list is a map.
		lm.semantic = listMap
		// NOTE: we validate uniqueness of the keys in the listValidator.
	default:
		return Validations{}, fmt.Errorf("unknown list type %q", tag.Value)
	}

	// This tag doesn't generate any validations.  It just accumulates
	// information for other tags to use.
	return Validations{}, nil
}

func (lttv listTypeTagValidator) Docs() TagDoc {
	doc := TagDoc{
		Tag:            lttv.TagName(),
		StabilityLevel: Stable,
		Scopes:         lttv.ValidScopes().UnsortedList(),
		Description:    "Declares a list field's semantic type and ownership behavior. atomic: single ownership, set: shared ownership with uniqueness, map: shared ownership with key-based uniqueness.",
		Payloads: []TagPayloadDoc{{
			Description: "<type>",
			Docs:        "atomic | map | set",
		}},
		PayloadsType:     codetags.ValueTypeString,
		PayloadsRequired: true,
	}
	return doc
}

type listMapKeyTagValidator struct {
	byPath map[string]*listMetadata
}

func (listMapKeyTagValidator) Init(Config) {}

func (listMapKeyTagValidator) TagName() string {
	return ListMapKeyTagName
}

func (listMapKeyTagValidator) ValidScopes() sets.Set[Scope] {
	return listTagsValidScopes
}

func (lmktv listMapKeyTagValidator) GetValidations(context Context, tag codetags.Tag) (Validations, error) {
	// NOTE: pointers to lists are not supported, so we should never see a pointer here.
	t := util.NativeType(context.Type)
	if t.Kind != types.Slice && t.Kind != types.Array {
		return Validations{}, fmt.Errorf("can only be used on list types (%s)", t.Kind)
	}
	// NOTE: lists of pointers are not supported, so we should never see a pointer here.
	if util.NativeType(t.Elem).Kind != types.Struct {
		return Validations{}, fmt.Errorf("only lists of structs can be list-maps")
	}

	var fieldName string
	if memb := util.GetMemberByJSON(util.NativeType(t.Elem), tag.Value); memb == nil {
		return Validations{}, fmt.Errorf("no field for JSON name %q", tag.Value)
	} else if k := util.NativeType(memb.Type).Kind; k != types.Builtin {
		return Validations{}, fmt.Errorf("only primitive types can be list-map keys (%s)", k)
	} else {
		fieldName = memb.Name
	}

	lm := lmktv.byPath[context.Path.String()]
	if lm == nil {
		lm = &listMetadata{}
		lmktv.byPath[context.Path.String()] = lm
	}
	lm.keyFields = append(lm.keyFields, fieldName)
	lm.keyNames = append(lm.keyNames, tag.Value)

	// This tag doesn't generate any validations.  It just accumulates
	// information for other tags to use.
	return Validations{}, nil
}

func (lmktv listMapKeyTagValidator) Docs() TagDoc {
	doc := TagDoc{
		Tag:            lmktv.TagName(),
		StabilityLevel: Stable,
		Scopes:         lmktv.ValidScopes().UnsortedList(),
		Description:    "Declares a named sub-field of a list's value-type to be part of the list-map key.",
		Payloads: []TagPayloadDoc{{
			Description: "<field-json-name>",
			Docs:        "The name of the field.",
		}},
		PayloadsType:     codetags.ValueTypeString,
		PayloadsRequired: true,
	}
	return doc
}

type uniqueTagValidator struct {
	byPath map[string]*listMetadata
}

func (uniqueTagValidator) Init(Config) {}

func (uniqueTagValidator) TagName() string {
	return uniqueTagName
}

func (uniqueTagValidator) ValidScopes() sets.Set[Scope] {
	return listTagsValidScopes
}

func (utv uniqueTagValidator) GetValidations(context Context, tag codetags.Tag) (Validations, error) {
	// NOTE: pointers to lists are not supported, so we should never see a pointer here.
	t := util.NativeType(context.Type)
	if t.Kind != types.Slice && t.Kind != types.Array {
		return Validations{}, fmt.Errorf("can only be used on list types (%s)", t.Kind)
	}

	lm := utv.byPath[context.Path.String()]
	if lm == nil {
		lm = &listMetadata{}
		utv.byPath[context.Path.String()] = lm
	}

	switch tag.Value {
	case "set":
		lm.unique = listSet
		// NOTE: we validate uniqueness in the listValidator.
	case "map":
		// NOTE: maps of pointers are not supported, so we should never see a pointer here.
		if util.NativeType(t.Elem).Kind != types.Struct {
			return Validations{}, fmt.Errorf("only lists of structs can be list-maps")
		}

		// Save the fact that this list is a map.
		lm.unique = listMap
		// NOTE: we validate uniqueness of the keys in the listValidator.
	default:
		return Validations{}, fmt.Errorf("unknown unique type %q", tag.Value)
	}

	// This tag doesn't generate any validations.  It just accumulates
	// information for other tags to use.
	return Validations{}, nil
}

func (utv uniqueTagValidator) Docs() TagDoc {
	doc := TagDoc{
		Tag:         utv.TagName(),
		Scopes:      utv.ValidScopes().UnsortedList(),
		Description: "Declares that a list field's elements are unique. This tag can be used with listType=atomic to add uniqueness constraints, or independently to specify uniqueness semantics.",
		Payloads: []TagPayloadDoc{{
			Description: "<type>",
			Docs:        "map | set",
		}},
		PayloadsType:     codetags.ValueTypeString,
		PayloadsRequired: true,
	}
	return doc
}

type listValidator struct {
	byPath map[string]*listMetadata
}

func (listValidator) Init(_ Config) {}

func (listValidator) Name() string {
	return "listValidator"
}

var (
	validateUnique = types.Name{Package: libValidationPkg, Name: "Unique"}
)

func (lv listValidator) GetValidations(context Context) (Validations, error) {
	nt := util.NativeType(context.Type)
	if nt.Kind != types.Slice && nt.Kind != types.Array {
		return Validations{}, nil
	}

	// Look up the list metadata which is defined on this field or type.
	lm := lv.byPath[context.Path.String()]

	// NOTE: We don't really support list-of-list or map-of-list, so this does
	// not consider the case of ScopeListVal or ScopeMapVal. If we want to
	// support those, we need to look at this and make sure the paths work the
	// way we need.
	if context.Scope == ScopeField {
		// If this is a field, look up the list metadata for the type.
		// TypeValidators happen before FieldValidators, so this is safe.
		tm := lv.byPath[context.Type.String()]
		if lm != nil && tm != nil {
			return Validations{}, fmt.Errorf("found list metadata for both a field and its type: %s", context.Path)
		}
		// TODO(thockin): enable this once the whole codebase is converted or
		// if we only run against fields which are opted-in.
		// if lm == nil && tm == nil {
		// 	 return Validations{}, fmt.Errorf("found a list field without list metadata")
		// }
	}

	if lm == nil {
		// If we don't have metadata for this field, we might have it for the
		// field's type.
		return Validations{}, nil
	}

	// Do this after the above - if we only get one error, the one(s) above
	// this are more important.
	if err := lv.check(lm); err != nil {
		return Validations{}, err
	}

	result := Validations{}

	// Generate uniqueness checks for lists with higher-order semantics.
	if lm.semantic == listSet || lm.unique == listSet {
		// Only compare primitive values when possible. Slices and maps are not
		// comparable, and structs might hold pointer fields, which are directly
		// comparable but not what we need.
		//
		// NOTE: lists of pointers are not supported, so we should never see a pointer here.
		matchArg := validateSemanticDeepEqual
		if util.IsDirectComparable(util.NonPointer(util.NativeType(nt.Elem))) {
			matchArg = validateDirectEqual
		}
		comment := ""
		if lm.semantic == listSet {
			comment = "lists with set semantics require unique values"
		} else { // lm.unique == listSet
			comment = "unique=set requires unique values"
		}
		f := Function("listValidator", DefaultFlags, validateUnique, Identifier(matchArg)).
			WithComment(comment)
		result.AddFunction(f)
	}
	if lm.semantic == listMap || lm.unique == listMap {
		// TODO: There are some fields which are declared as maps which do not
		// enforce uniqueness in manual validation. Those either need to not be
		// maps or we need to allow types to opt-out from this validation.  SSA
		// is also not able to handle these well.
		matchArg := lm.makeListMapMatchFunc(nt.Elem)
		comment := ""
		if lm.semantic == listMap {
			// comment = "listType=map requires unique keys"
		} else { // lm.unique == listMap
			comment = "unique=map requires unique keys"
		}
		// TODO: Remove this check once we can support listType=map uniqueness.
		if lm.unique == listMap {
			f := Function("listValidator", DefaultFlags, validateUnique, matchArg).
				WithComment(comment)
			result.AddFunction(f)
		}
	}

	return result, nil
}

// make sure a given listMetadata makes sense.
func (lv listValidator) check(lm *listMetadata) error {
	// Check some fundamental constraints on list tags.

	// If we have listMapKey but no map semantics, that's an error
	if len(lm.keyFields) > 0 && lm.semantic != listMap && lm.unique != listMap {
		return fmt.Errorf("found listMapKey without listType=map or unique=map")
	}

	// If we have map semantics but no keys, that's an error
	if (lm.semantic == listMap || lm.unique == listMap) && len(lm.keyFields) == 0 {
		return fmt.Errorf("found listType=map or unique=map without listMapKey")
	}

	// If we have both listType and unique with the same semantics, that's redundant
	if lm.semantic != "" && lm.unique != "" && lm.semantic == lm.unique {
		return fmt.Errorf("redundant declaration: listType=%s and unique=%s", lm.semantic, lm.unique)
	}

	return nil
}

type eachValTagValidator struct {
	byPath    map[string]*listMetadata
	validator Validator
}

func (evtv *eachValTagValidator) Init(cfg Config) {
	evtv.validator = cfg.Validator
}

func (eachValTagValidator) TagName() string {
	return eachValTagName
}

func (eachValTagValidator) ValidScopes() sets.Set[Scope] {
	return listTagsValidScopes
}

// LateTagValidator indicates that this validator has to run AFTER the listType
// and listMapKey tags.
func (eachValTagValidator) LateTagValidator() {}

var (
	validateEachSliceVal      = types.Name{Package: libValidationPkg, Name: "EachSliceVal"}
	validateEachMapVal        = types.Name{Package: libValidationPkg, Name: "EachMapVal"}
	validateSemanticDeepEqual = types.Name{Package: libValidationPkg, Name: "SemanticDeepEqual"}
	validateDirectEqual       = types.Name{Package: libValidationPkg, Name: "DirectEqual"}
	validateDirectEqualPtr    = types.Name{Package: libValidationPkg, Name: "DirectEqualPtr"}
)

func (evtv eachValTagValidator) GetValidations(context Context, tag codetags.Tag) (Validations, error) {
	// NOTE: pointers to lists and maps are not supported, so we should never see a pointer here.
	t := context.Type
	nt := util.NativeType(t)
	switch nt.Kind {
	case types.Slice, types.Array, types.Map:
	default:
		return Validations{}, fmt.Errorf("can only be used on list or map types (%s)", nt.Kind)
	}

	elemContext := Context{
		// Scope is initialized below.
		Type:       nt.Elem,
		Path:       context.Path.Key("(vals)"),
		Member:     nil, // NA for list/map values
		ParentPath: context.Path,
	}
	switch nt.Kind {
	case types.Slice, types.Array:
		elemContext.Scope = ScopeListVal
		elemContext.ListSelector = []ListSelectorTerm{} // empty == "all"
	case types.Map:
		elemContext.Scope = ScopeMapVal
		// TODO: We may need map selectors at some point.
	}
	if tag.ValueTag == nil {
		return Validations{}, fmt.Errorf("missing validation tag")
	}
	if validations, err := evtv.validator.ExtractValidations(elemContext, *tag.ValueTag); err != nil {
		return Validations{}, err
	} else {
		if validations.Empty() && !validations.OpaqueKeyType && !validations.OpaqueValType && !validations.OpaqueType {
			return Validations{}, fmt.Errorf("no validation functions found")
		}
		if len(validations.Variables) > 0 {
			return Validations{}, fmt.Errorf("variable generation is not supported")
		}
		// Pass the real (possibly alias) type.
		return evtv.getValidations(context.Path, t, validations)
	}
}

// t is expected to be the top-most type of the list or map. For example, if
// this is a typedef to a list, this is the alias type, not the underlying
// type.
func (evtv eachValTagValidator) getValidations(fldPath *field.Path, t *types.Type, validations Validations) (Validations, error) {
	switch util.NativeType(t).Kind {
	case types.Slice, types.Array:
		return evtv.getListValidations(fldPath, t, validations)
	case types.Map:
		return evtv.getMapValidations(t, validations)
	}
	return Validations{}, fmt.Errorf("non-iterable type: %v", t)
}

// ForEachVal returns a validation that applies a function to each element of
// a list or map. The type argument is expected to be the top-most type of the
// list or map. For example, if this is a typedef to a list, this is the alias
// type, not the underlying type.
func ForEachVal(fldPath *field.Path, t *types.Type, fn FunctionGen) (Validations, error) {
	return globalEachVal.getValidations(fldPath, t, Validations{Functions: []FunctionGen{fn}})
}

// t is expected to be the top-most type of the list. For example, if this is a
// typedef to a list, this is the alias type, not the underlying type.
func (evtv eachValTagValidator) getListValidations(fldPath *field.Path, t *types.Type, validations Validations) (Validations, error) {
	result := Validations{}
	result.OpaqueValType = validations.OpaqueType

	// This type is a "late" validator, so it runs after all the keys are
	// registered.  See LateTagValidator() above.
	listMetadata := evtv.byPath[fldPath.String()]
	if listMetadata == nil {
		// If we don't have metadata for this field, we might have it for the
		// field's type.
		listMetadata = evtv.byPath[t.String()]
	}

	nt := util.NativeType(t)

	// matchArg is the function that is used to lookup the correlated element in the old list.
	var matchArg any = Literal("nil")

	// equivArg is the function that is used to compare the correlated elements in the old and new lists.
	// It would be "nil" if the matchArg is a full comparison function.
	var equivArg any = Literal("nil")

	// directComparable is used to determine whether we can use the direct
	// comparison operator "==" or need to use the semantic DeepEqual when
	// looking up and comparing correlated list elements for validation ratcheting.
	directComparable := util.IsDirectComparable(util.NonPointer(util.NativeType(nt.Elem)))

	switch {
	case listMetadata != nil && (listMetadata.semantic == listMap || listMetadata.unique == listMap):
		// For listType=map, we use key to lookup the correlated element in the old list.
		// And use equivFunc to compare the correlated elements in the old and new lists.
		matchArg = listMetadata.makeListMapMatchFunc(nt.Elem)
		if directComparable {
			equivArg = Identifier(validateDirectEqual)
		} else {
			equivArg = Identifier(validateSemanticDeepEqual)
		}
	case listMetadata != nil && (listMetadata.semantic == listSet || listMetadata.unique == listSet):
		// For listType=set, matchArg is the equivalence check, so equivArg is nil.
		if directComparable {
			matchArg = Identifier(validateDirectEqual)
		} else {
			matchArg = Identifier(validateSemanticDeepEqual)
		}
	default:
		// For non-map and non-set list, we don't lookup the correlated element in the old list.
		// The matchArg and equivArg are both nil.
	}

	for _, vfn := range validations.Functions {
		comm := vfn.Comments
		vfn.Comments = nil
		f := Function(eachValTagName, vfn.Flags, validateEachSliceVal, matchArg, equivArg, WrapperFunction{vfn, nt.Elem}).WithComments(comm...)
		result.AddFunction(f)
	}

	return result, nil
}

// t is expected to be the top-most type of the map. For example, if this is a
// typedef to a map, this is the alias type, not the underlying type.
func (evtv eachValTagValidator) getMapValidations(t *types.Type, validations Validations) (Validations, error) {
	result := Validations{}
	result.OpaqueValType = validations.OpaqueType

	nt := util.NativeType(t)
	equivArg := Identifier(validateSemanticDeepEqual)
	if util.IsDirectComparable(util.NonPointer(util.NativeType(nt.Elem))) {
		equivArg = Identifier(validateDirectEqual)
	}
	for _, vfn := range validations.Functions {
		comm := vfn.Comments
		vfn.Comments = nil
		f := Function(eachValTagName, vfn.Flags, validateEachMapVal, equivArg, WrapperFunction{vfn, nt.Elem}).WithComments(comm...)
		result.AddFunction(f)
	}

	return result, nil
}

func (evtv eachValTagValidator) Docs() TagDoc {
	doc := TagDoc{
		Tag:            evtv.TagName(),
		StabilityLevel: Alpha,
		Scopes:         evtv.ValidScopes().UnsortedList(),
		Description:    "Declares a validation for each value in a map or list.",
		Payloads: []TagPayloadDoc{{
			Description: "<validation-tag>",
			Docs:        "The tag to evaluate for each value.",
		}},
		PayloadsType:     codetags.ValueTypeTag,
		PayloadsRequired: true,
	}
	return doc
}

type eachKeyTagValidator struct {
	validator Validator
}

func (ektv *eachKeyTagValidator) Init(cfg Config) {
	ektv.validator = cfg.Validator
}

func (eachKeyTagValidator) TagName() string {
	return eachKeyTagName
}

func (eachKeyTagValidator) ValidScopes() sets.Set[Scope] {
	return listTagsValidScopes
}

var (
	validateEachMapKey = types.Name{Package: libValidationPkg, Name: "EachMapKey"}
)

func (ektv eachKeyTagValidator) GetValidations(context Context, tag codetags.Tag) (Validations, error) {
	// NOTE: pointers to lists are not supported, so we should never see a pointer here.
	t := context.Type
	nt := util.NativeType(t)
	if nt.Kind != types.Map {
		return Validations{}, fmt.Errorf("can only be used on map types (%s)", nt.Kind)
	}

	elemContext := Context{
		Scope:      ScopeMapKey,
		Type:       nt.Elem,
		Path:       context.Path.Key("(keys)"),
		Member:     nil, // NA for map keys
		ParentPath: context.Path,
	}

	if validations, err := ektv.validator.ExtractValidations(elemContext, *tag.ValueTag); err != nil {
		return Validations{}, err
	} else {
		if len(validations.Variables) > 0 {
			return Validations{}, fmt.Errorf("variable generation is not supported")
		}

		return ektv.getValidations(t, validations)
	}
}

func (ektv eachKeyTagValidator) getValidations(t *types.Type, validations Validations) (Validations, error) {
	nt := util.NativeType(t)
	result := Validations{}
	result.OpaqueKeyType = validations.OpaqueType
	for _, vfn := range validations.Functions {
		comm := vfn.Comments
		vfn.Comments = nil
		f := Function(eachKeyTagName, vfn.Flags, validateEachMapKey, WrapperFunction{vfn, nt.Key}).WithComments(comm...)
		result.AddFunction(f)
	}
	return result, nil
}

// ForEachKey returns a validation that applies a function to each key of
// a map.
func ForEachKey(_ *field.Path, t *types.Type, fn FunctionGen) (Validations, error) {
	return globalEachKey.getValidations(t, Validations{Functions: []FunctionGen{fn}})
}

func (ektv eachKeyTagValidator) Docs() TagDoc {
	doc := TagDoc{
		Tag:            ektv.TagName(),
		Scopes:         ektv.ValidScopes().UnsortedList(),
		StabilityLevel: Alpha,
		Description:    "Declares a validation for each value in a map or list.",
		Payloads: []TagPayloadDoc{{
			Description: "<validation-tag>",
			Docs:        "The tag to evaluate for each key.",
		}},
		PayloadsType:     codetags.ValueTypeTag,
		PayloadsRequired: true,
	}
	return doc
}
