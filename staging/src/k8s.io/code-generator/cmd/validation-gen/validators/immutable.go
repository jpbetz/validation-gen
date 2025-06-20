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
	frozenTagName = "k8s:frozen"
)

func init() {
	RegisterTagValidator(frozenTagValidator{})
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
		// This is a minor optimization to just compare primitive values when
		// possible. Slices and maps are not comparable, and structs might hold
		// pointer fields, which are directly comparable but not what we need.
		//
		// Note: This compares the pointee, not the pointer itself.
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
