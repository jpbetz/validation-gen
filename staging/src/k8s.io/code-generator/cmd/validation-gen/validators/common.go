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
	"encoding/json"
	"fmt"
	"reflect"

	"k8s.io/code-generator/cmd/validation-gen/util"
	"k8s.io/gengo/v2"
	"k8s.io/gengo/v2/types"
)

const (
	// libValidationPkg is the pkgpath to our "standard library" of validation
	// functions.
	libValidationPkg = "k8s.io/apimachinery/pkg/api/validate"
)

// rootTypeString returns a string representation of the relationship between
// src and dst types, for use in error messages.
func rootTypeString(src, dst *types.Type) string {
	if src == dst {
		return src.String()
	}
	return src.String() + " -> " + dst.String()
}

// hasZeroDefault returns whether the field has a default value and whether
// that default value is the zero value for the field's type.
func hasZeroDefault(context Context) (bool, bool, error) {
	t := util.NonPointer(util.NativeType(context.Type))
	// This validator only applies to fields, so Member must be valid.
	tagsByName, err := gengo.ExtractFunctionStyleCommentTags("+", []string{defaultTagName}, context.Member.CommentLines)
	if err != nil {
		return false, false, fmt.Errorf("failed to read tags: %w", err)
	}

	tags, hasDefault := tagsByName[defaultTagName]
	if !hasDefault {
		return false, false, nil
	}
	if len(tags) == 0 {
		return false, false, fmt.Errorf("+default tag with no value")
	}
	if len(tags) > 1 {
		return false, false, fmt.Errorf("+default tag with multiple values: %q", tags)
	}

	payload := tags[0].Value
	var defaultValue any
	if err := json.Unmarshal([]byte(payload), &defaultValue); err != nil {
		return false, false, fmt.Errorf("failed to parse default value %q: %w", payload, err)
	}
	if defaultValue == nil {
		return false, false, fmt.Errorf("failed to parse default value %q: unmarshalled to nil", payload)
	}

	zero, found := typeZeroValue[t.String()]
	if !found {
		return false, false, fmt.Errorf("unknown zero-value for type %s", t.String())
	}

	return true, reflect.DeepEqual(defaultValue, zero), nil
}

// This is copied from defaulter-gen.
// TODO: move this back to gengo as Type.ZeroValue()?
var typeZeroValue = map[string]any{
	"uint":        0.,
	"uint8":       0.,
	"uint16":      0.,
	"uint32":      0.,
	"uint64":      0.,
	"int":         0.,
	"int8":        0.,
	"int16":       0.,
	"int32":       0.,
	"int64":       0.,
	"byte":        0.,
	"float64":     0.,
	"float32":     0.,
	"bool":        false,
	"time.Time":   "",
	"string":      "",
	"integer":     0.,
	"number":      0.,
	"boolean":     false,
	"[]byte":      "", // base64 encoded characters
	"interface{}": interface{}(nil),
	"any":         interface{}(nil),
}
