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
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/code-generator/cmd/validation-gen/util"
	"k8s.io/gengo/v2/codetags"
	"k8s.io/gengo/v2/types"
)

const (
	frozenTagName    = "k8s:frozen"
	immutableTagName = "k8s:immutable"
)

func init() {
	RegisterTagValidator(frozenTagValidator{})
	RegisterTagValidator(immutableTagValidator{})
}

type frozenTagValidator struct{}

func (frozenTagValidator) Init(_ Config) {}

func (frozenTagValidator) TagName() string {
	return frozenTagName
}

var frozenTagValidScopes = sets.New(ScopeField, ScopeType, ScopeMapVal, ScopeListVal)

func (frozenTagValidator) ValidScopes() sets.Set[Scope] {
	return frozenTagValidScopes
}

var (
	frozenCompareValidator = types.Name{Package: libValidationPkg, Name: "FrozenByCompare"}
	frozenReflectValidator = types.Name{Package: libValidationPkg, Name: "FrozenByReflect"}
)

func (frozenTagValidator) GetValidations(context Context, _ codetags.Tag) (Validations, error) {
	var result Validations

	if util.IsDirectComparable(util.NonPointer(util.NativeType(context.Type))) {
		result.AddFunction(Function(frozenTagName, DefaultFlags, frozenCompareValidator))
	} else {
		result.AddFunction(Function(frozenTagName, DefaultFlags, frozenReflectValidator))
	}

	return result, nil
}

func (ftv frozenTagValidator) Docs() TagDoc {
	return TagDoc{
		Tag:         ftv.TagName(),
		Scopes:      ftv.ValidScopes().UnsortedList(),
		Description: "Indicates that a field may not be updated.",
	}
}

type immutableTagValidator struct{}

func (immutableTagValidator) Init(_ Config) {}

func (immutableTagValidator) TagName() string {
	return immutableTagName
}

var immutableTagValidScopes = sets.New(ScopeField, ScopeType, ScopeMapVal, ScopeListVal)

func (immutableTagValidator) ValidScopes() sets.Set[Scope] {
	return immutableTagValidScopes
}

var (
	immutableValueByCompareValidator   = types.Name{Package: libValidationPkg, Name: "ImmutableValueByCompare"}
	immutablePointerByCompareValidator = types.Name{Package: libValidationPkg, Name: "ImmutablePointerByCompare"}
	immutableReflectValidator          = types.Name{Package: libValidationPkg, Name: "ImmutableByReflect"}
)

func (itv immutableTagValidator) GetValidations(context Context, _ codetags.Tag) (Validations, error) {
	var result Validations

	// If validating a field, check for default value.
	if context.Member != nil {
		if hasDefault, zeroDefault, err := hasZeroDefault(context); err != nil {
			return Validations{}, err
		} else if hasDefault && zeroDefault {
			result.AddComment("Zero-value defaults are treated as 'unset' by immutable validation.")
		} else if hasDefault && !zeroDefault {
			result.AddComment("Non-zero defaults are 'always set' and cannot transition from unset to set.")
		}
	}

	if !util.IsDirectComparable(util.NonPointer(util.NativeType(context.Type))) {
		result.AddFunction(Function(immutableTagName, DefaultFlags, immutableReflectValidator))
		return result, nil
	}

	isPointerField := false
	if context.Member != nil {
		memberType := context.Member.Type
		if memberType != nil && memberType.Kind == types.Pointer {
			isPointerField = true
		}
	} else if util.NativeType(context.Type).Kind == types.Pointer {
		isPointerField = true
	}

	if isPointerField {
		result.AddFunction(Function(immutableTagName, DefaultFlags, immutablePointerByCompareValidator))
	} else {
		result.AddFunction(Function(immutableTagName, DefaultFlags, immutableValueByCompareValidator))
	}

	return result, nil
}

func (itv immutableTagValidator) Docs() TagDoc {
	return TagDoc{
		Tag:         itv.TagName(),
		Scopes:      itv.ValidScopes().UnsortedList(),
		Description: "Indicates that a field can be set once (now or at creation), then becomes immutable. Allows transition from unset to set, but forbids modify or clear operations. Fields with default values are considered already set.",
	}
}
