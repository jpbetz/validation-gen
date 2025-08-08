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

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/gengo/v2/codetags"
	"k8s.io/gengo/v2/types"
)

const (
	ifEnabledTag  = "k8s:ifEnabled"
	ifDisabledTag = "k8s:ifDisabled"
)

func init() {
	RegisterTagValidator(&ifTagValidator{true, nil})
	RegisterTagValidator(&ifTagValidator{false, nil})
}

type ifTagValidator struct {
	enabled   bool
	validator Validator
}

func (iotv *ifTagValidator) Init(cfg Config) {
	iotv.validator = cfg.Validator
}

func (iotv ifTagValidator) TagName() string {
	if iotv.enabled {
		return ifEnabledTag
	}
	return ifDisabledTag
}

var ifEnabledDisabledTagValidScopes = sets.New(ScopeAny)

func (ifTagValidator) ValidScopes() sets.Set[Scope] {
	return ifEnabledDisabledTagValidScopes
}

var (
	ifOption = types.Name{Package: libValidationPkg, Name: "IfOption"}
)

func (iotv ifTagValidator) GetValidations(context Context, tag codetags.Tag) (Validations, error) {
	optionArg, ok := tag.NamedArg("option")
	if !ok {
		return Validations{}, fmt.Errorf("missing required argument: option")
	}
	result := Validations{}
	if validations, err := iotv.validator.ExtractValidations(context, *tag.ValueTag); err != nil {
		return Validations{}, err
	} else {
		for _, fn := range validations.Functions {
			f := Function(iotv.TagName(), fn.Flags, ifOption, optionArg.Value, iotv.enabled, WrapperFunction{Function: fn, ObjType: context.Type})
			result.Variables = append(result.Variables, validations.Variables...)
			result.AddFunction(f)
		}
		return result, nil
	}
}

func (iotv ifTagValidator) Docs() TagDoc {
	doc := TagDoc{
		Tag: iotv.TagName(),
		Args: []TagArgDoc{{
			Name:     "option",
			Type:     codetags.ArgTypeString,
			Required: true,
		}},
		Scopes: iotv.ValidScopes().UnsortedList(),
	}

	doc.PayloadsType = codetags.ValueTypeTag
	doc.PayloadsRequired = true
	if iotv.enabled {
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
