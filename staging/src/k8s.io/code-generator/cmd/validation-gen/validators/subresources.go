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
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/gengo/v2/codetags"
)

const (
	subresourceTag    = "k8s:subresource"
	notSubresourceTag = "k8s:notSubresource"
)

func init() {
	RegisterTagValidator(&isSubresourceValidator{true, nil})
	RegisterTagValidator(&isSubresourceValidator{false, nil})
}

type isSubresourceValidator struct {
	matchSubresource bool
	validator        Validator
}

func (sv *isSubresourceValidator) Init(cfg Config) {
	sv.validator = cfg.Validator
}

func (sv isSubresourceValidator) TagName() string {
	if sv.matchSubresource {
		return subresourceTag
	}
	return notSubresourceTag
}

var isSubresourceTagValidScopes = sets.New(ScopeAny)

func (isSubresourceValidator) ValidScopes() sets.Set[Scope] {
	return isSubresourceTagValidScopes
}

func (sv isSubresourceValidator) GetValidations(context Context, tag codetags.Tag) (Validations, error) {
	subresourceName, _ := tag.PositionalArg()

	result := Validations{}

	if validations, err := sv.validator.ExtractValidations(context, *tag.ValueTag); err != nil {
		return Validations{}, err
	} else {
		for _, fn := range validations.Functions {
			if sv.matchSubresource {
				result.Functions = append(result.Functions, fn.WithConditions(Conditions{IsSubresource: subresourceName.Value}))
			} else {
				result.Functions = append(result.Functions, fn.WithConditions(Conditions{IsNotSubresource: subresourceName.Value}))
			}
			result.Variables = append(result.Variables, validations.Variables...)
		}
		return result, nil
	}

}

func (sv isSubresourceValidator) Docs() TagDoc {
	doc := TagDoc{
		Tag: sv.TagName(),
		Args: []TagArgDoc{{
			Description: "<subresource-name>",
			Required:    true,
			Type:        codetags.ArgTypeString,
		}},
		Scopes: sv.ValidScopes().UnsortedList(),
	}

	if sv.matchSubresource {
		doc.Description = "Declares a validation that only applies when the requested subresource is the specified subresource."
		doc.Payloads = []TagPayloadDoc{{
			Description: "<validation-tag>",
			Docs:        "This validation tag will be evaluated only if the requested subresource is the specified subresource.",
		}}
	} else {
		doc.Description = "Declares a validation that only applies when the requested subresource is not the specified subresource."
		doc.Payloads = []TagPayloadDoc{{
			Description: "<validation-tag>",
			Docs:        "This validation tag will be evaluated only if the requested subresource is not the specified subresource.",
		}}
	}
	return doc
}
