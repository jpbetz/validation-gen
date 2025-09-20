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
	"sort"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/gengo/v2/codetags"
	"k8s.io/gengo/v2/types"
)

const (
	ifEnabledTag  = "k8s:ifEnabled"
	ifDisabledTag = "k8s:ifDisabled"
)

// globalIfOptions maps field paths to option names to enabled/disabled to the tags to apply.
// It is populated by the ifTagValidator and consumed by the ifValidator.
var globalIfOptions = map[string]map[string]map[bool][]codetags.Tag{}

func init() {
	// Accumulate conditional validations via tags.
	RegisterTagValidator(&ifTagValidator{true, nil, globalIfOptions})
	RegisterTagValidator(&ifTagValidator{false, nil, globalIfOptions})

	// Finish work on the accumulated conditional validations.
	RegisterFieldValidator(&ifValidator{byPath: globalIfOptions})
	RegisterTypeValidator(&ifValidator{byPath: globalIfOptions})
}

// ifTagValidator collects conditional validations from k8s:ifEnabled and
// k8s:ifDisabled tags and stores them in the globalIfOptions map.
type ifTagValidator struct {
	enabled   bool
	validator Validator
	byPath    map[string]map[string]map[bool][]codetags.Tag
}

func (itv *ifTagValidator) Init(cfg Config) {
	itv.validator = cfg.Validator
}

func (itv ifTagValidator) TagName() string {
	if itv.enabled {
		return ifEnabledTag
	}
	return ifDisabledTag
}

var ifEnabledDisabledTagValidScopes = sets.New(ScopeType, ScopeField, ScopeListVal, ScopeMapKey, ScopeMapVal, ScopeConst)

func (ifTagValidator) ValidScopes() sets.Set[Scope] {
	return ifEnabledDisabledTagValidScopes
}

var (
	ifOption = types.Name{Package: libValidationPkg, Name: "IfOption"}
)

// GetValidations for ifTagValidator does not return any validations directly.
// Instead, it populates the globalIfOptions map with the validations that
// should be applied when the specified option is enabled or disabled.
func (itv ifTagValidator) GetValidations(context Context, tag codetags.Tag) (Validations, error) {
	optionArg, ok := tag.PositionalArg()
	if !ok {
		return Validations{}, fmt.Errorf("missing required option name positional argument")
	}
	optionName := optionArg.Value

	if tag.ValueTag == nil {
		return Validations{}, fmt.Errorf("tag %q: missing value tag", itv.TagName())
	}

	fieldPath := context.Path.String()
	if itv.byPath[fieldPath] == nil {
		itv.byPath[fieldPath] = make(map[string]map[bool][]codetags.Tag)
	}
	if itv.byPath[fieldPath][optionName] == nil {
		itv.byPath[fieldPath][optionName] = make(map[bool][]codetags.Tag)
	}

	itv.byPath[fieldPath][optionName][itv.enabled] = append(itv.byPath[fieldPath][optionName][itv.enabled], *tag.ValueTag)

	return Validations{}, nil
}

func (itv ifTagValidator) Docs() TagDoc {
	doc := TagDoc{
		Tag:            itv.TagName(),
		StabilityLevel: Alpha,
		Args: []TagArgDoc{{
			Description: "<option>",
			Type:        codetags.ArgTypeString,
			Required:    true,
		}},
		Scopes: itv.ValidScopes().UnsortedList(),
	}

	doc.PayloadsType = codetags.ValueTypeTag
	doc.PayloadsRequired = true
	if itv.enabled {
		doc.Description = "Declares a validation that only applies when an option is enabled."
		doc.Payloads = []TagPayloadDoc{{
			Description: "<validation-tag>",
			Docs:        "This validation tag will be evaluated only if the validation option is enabled.",
		}}
	} else {
		doc.Description = "Declares a validation that only applies when an option is disabled."
		doc.Payloads = []TagPayloadDoc{{
			Description: "<validation-tag>",
			Docs:        "This validation tag will be evaluated only if the validation option is disabled.",
		}}
	}
	return doc
}

// ifValidator generates the final validation functions for conditional
// validations that were collected by ifTagValidator.
type ifValidator struct {
	byPath    map[string]map[string]map[bool][]codetags.Tag
	validator Validator
}

func (v *ifValidator) Init(cfg Config) {
	v.validator = cfg.Validator
}

func (ifValidator) Name() string {
	return "ifValidator"
}

// GetValidations for ifValidator generates the validation functions for the
// conditional validations that were collected by ifTagValidator.
func (v *ifValidator) GetValidations(context Context) (Validations, error) {
	fieldPath := context.Path.String()
	optionsForField, ok := v.byPath[fieldPath]
	if !ok {
		return Validations{}, nil
	}

	result := Validations{}
	optionNames := make([]string, 0, len(optionsForField))
	for name := range optionsForField {
		optionNames = append(optionNames, name)
	}
	sort.Strings(optionNames)

	for _, optionName := range optionNames {
		enabledMap := optionsForField[optionName]

		enabledKeys := make([]bool, 0, len(enabledMap))
		for enabled := range enabledMap {
			enabledKeys = append(enabledKeys, enabled)
		}
		sort.Slice(enabledKeys, func(i, j int) bool { return !enabledKeys[i] }) // true then false

		for _, enabled := range enabledKeys {
			tags := enabledMap[enabled]
			if len(tags) == 0 {
				continue
			}

			newContext := context
			// The path needs to be unique to distinguish between the `ifEnabled` and `ifDisabled` branches for the same option.
			// This constructs a virtual field path as `<original-path>.(<option>=<enabled>)`.
			newContext.Path = context.Path.Child(fmt.Sprintf("(%s=%v)", optionName, enabled))
			validations, err := v.validator.ExtractValidations(newContext, tags...)
			if err != nil {
				return Validations{}, err
			}

			if validations.Empty() {
				continue
			}

			result.Variables = append(result.Variables, validations.Variables...)
			validations.Variables = nil

			tagName := ifEnabledTag
			if !enabled {
				tagName = ifDisabledTag
			}

			f := Function(tagName, DefaultFlags, ifOption, optionName, enabled, WrapperFunction{validations.Functions, context.Type})
			result.AddFunction(f)
		}
	}

	return result, nil
}
