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

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/code-generator/cmd/validation-gen/util"
	"k8s.io/gengo/v2/codetags"
	"k8s.io/gengo/v2/types"
)

const (
	neqStringTagName = "k8s:NEQString"
)

func init() {
	RegisterTagValidator(neqStringTagValidator{})
}

type neqStringTagValidator struct{}

func (neqStringTagValidator) Init(_ Config) {}

func (neqStringTagValidator) TagName() string {
	return neqStringTagName
}

var neqStringTagValidScopes = sets.New(ScopeAny)

func (neqStringTagValidator) ValidScopes() sets.Set[Scope] {
	return neqStringTagValidScopes
}

var (
	neqStringValidator = types.Name{Package: libValidationPkg, Name: "NEQString"}
)

func (v neqStringTagValidator) GetValidations(context Context, tag codetags.Tag) (Validations, error) {
	var result Validations

	// This tag can apply to value and pointer fields, as well as typedefs
	// (which should never be pointers). We need to check the concrete type.
	t := util.NonPointer(util.NativeType(context.Type))
	if t != types.String {
		return Validations{}, fmt.Errorf("can only be used on string types (got %s)", rootTypeString(context.Type, t))
	}

	if len(tag.Args) != 1 {
		return Validations{}, fmt.Errorf("tag must have exactly one argument")
	}
	arg := tag.Args[0]
	if arg.Name != "" {
		return Validations{}, fmt.Errorf("argument must be positional, not named")
	}

	result.AddFunction(Function(v.TagName(), DefaultFlags, neqStringValidator, arg.Value))
	return result, nil
}

func (v neqStringTagValidator) Docs() TagDoc {
	return TagDoc{
		Tag:         v.TagName(),
		Usage:       "k8s:NEQString(<disallowed-string>)",
		Scopes:      v.ValidScopes().UnsortedList(),
		Description: "Validates the field's value is not equal to a specified string.",
		Args: []TagArgDoc{{
			Description: "<disallowed-string>",
			Docs:        "The string value that is not allowed.",
			Type:        codetags.ArgTypeString,
			Required:    true,
		}},
	}
}
