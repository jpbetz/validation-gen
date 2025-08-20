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

package validate

import (
	"context"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/api/operation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestShortName(t *testing.T) {
	ctx := context.Background()
	fldPath := field.NewPath("test")

	testCases := []struct {
		name     string
		input    string
		wantErrs field.ErrorList
	}{{
		name:     "valid name",
		input:    "valid-name",
		wantErrs: nil,
	}, {
		name:     "valid single character name",
		input:    "a",
		wantErrs: nil,
	}, {
		name:     "valid name with numbers",
		input:    "123-abc",
		wantErrs: nil,
	}, {
		name:  "invalid: uppercase characters",
		input: "Invalid-Name",
		wantErrs: field.ErrorList{
			field.Invalid(fldPath, nil, "").WithOrigin("format=k8s-short-name"),
		},
	}, {
		name:  "invalid: starts with dash",
		input: "-invalid-name",
		wantErrs: field.ErrorList{
			field.Invalid(fldPath, nil, "").WithOrigin("format=k8s-short-name"),
		},
	}, {
		name:  "invalid: ends with dash",
		input: "invalid-name-",
		wantErrs: field.ErrorList{
			field.Invalid(fldPath, nil, "").WithOrigin("format=k8s-short-name"),
		},
	}, {
		name:  "invalid: contains dots",
		input: "invalid.name",
		wantErrs: field.ErrorList{
			field.Invalid(fldPath, nil, "").WithOrigin("format=k8s-short-name"),
		},
	}, {
		name:  "invalid: contains special characters",
		input: "invalid@name",
		wantErrs: field.ErrorList{
			field.Invalid(fldPath, nil, "").WithOrigin("format=k8s-short-name"),
		},
	}, {
		name:  "invalid: too long",
		input: "a" + strings.Repeat("b", 62) + "c", // 64 characters
		wantErrs: field.ErrorList{
			field.Invalid(fldPath, nil, "").WithOrigin("format=k8s-short-name"),
		},
	}, {
		name:     "valid: max length",
		input:    "a" + strings.Repeat("b", 61) + "c", // 63 characters
		wantErrs: nil,
	}, {
		name:  "invalid: empty string",
		input: "",
		wantErrs: field.ErrorList{
			field.Invalid(fldPath, nil, "").WithOrigin("format=k8s-short-name"),
		},
	}}

	matcher := field.ErrorMatcher{}.ByType().ByField().ByOrigin()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value := tc.input
			gotErrs := ShortName(ctx, operation.Operation{}, fldPath, &value, nil)

			matcher.Test(t, tc.wantErrs, gotErrs)
		})
	}
}

func TestLongName(t *testing.T) {
	ctx := context.Background()
	fldPath := field.NewPath("test")

	testCases := []struct {
		name     string
		input    string
		wantErrs field.ErrorList
	}{{
		name:     "valid single label",
		input:    "valid-label",
		wantErrs: nil,
	}, {
		name:     "valid subdomain",
		input:    "this-is.a-valid.subdomain",
		wantErrs: nil,
	}, {
		name:     "valid single character elements",
		input:    "a.b.c",
		wantErrs: nil,
	}, {
		name:     "valid elements with numbers",
		input:    "123.abc-123.456-def",
		wantErrs: nil,
	}, {
		name:     "all number elements",
		input:    "1.2.3.4",
		wantErrs: nil,
	}, {
		name:  "invalid: uppercase characters",
		input: "Invalid.Subdomain",
		wantErrs: field.ErrorList{
			field.Invalid(fldPath, nil, "").WithOrigin("format=k8s-long-name"),
		},
	}, {
		name:  "invalid: starts with dash",
		input: "this-is.-an-invalid.subdomain",
		wantErrs: field.ErrorList{
			field.Invalid(fldPath, nil, "").WithOrigin("format=k8s-long-name"),
		},
	}, {
		name:  "invalid: ends with dash",
		input: "this-is.an-invalid-.subdomain",
		wantErrs: field.ErrorList{
			field.Invalid(fldPath, nil, "").WithOrigin("format=k8s-long-name"),
		},
	}, {
		name:  "invalid: contains double dots",
		input: "invalid..subdomain",
		wantErrs: field.ErrorList{
			field.Invalid(fldPath, nil, "").WithOrigin("format=k8s-long-name"),
		},
	}, {
		name:  "invalid: contains special characters",
		input: "inv@lid.subdoma!n",
		wantErrs: field.ErrorList{
			field.Invalid(fldPath, nil, "").WithOrigin("format=k8s-long-name"),
		},
	}, {
		name:  "invalid: too long single label",
		input: "a" + strings.Repeat("b", 252) + "c", // 254 characters
		wantErrs: field.ErrorList{
			field.Invalid(fldPath, nil, "").WithOrigin("format=k8s-long-name"),
		},
	}, {
		name: "invalid: too long multiple labels",
		input: strings.Join([]string{
			strings.Repeat("a", 60), // 61 with the "."
			strings.Repeat("b", 60), // 122 with the "."
			strings.Repeat("c", 60), // 183 with the "."
			strings.Repeat("d", 60), // 244 with the "."
			strings.Repeat("e", 10), // 254 characters
		}, "."),
		wantErrs: field.ErrorList{
			field.Invalid(fldPath, nil, "").WithOrigin("format=k8s-long-name"),
		},
	}, {
		name:     "valid: max length single label",     // supported for compat
		input:    "a" + strings.Repeat("b", 251) + "c", // 253 characters
		wantErrs: nil,
	}, {
		name:  "invalid: empty string",
		input: "",
		wantErrs: field.ErrorList{
			field.Invalid(fldPath, nil, "").WithOrigin("format=k8s-long-name"),
		},
	}}

	matcher := field.ErrorMatcher{}.ByType().ByField().ByOrigin()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value := tc.input
			gotErrs := LongName(ctx, operation.Operation{}, fldPath, &value, nil)

			matcher.Test(t, tc.wantErrs, gotErrs)
		})
	}
}

func TestK8sUUID(t *testing.T) {
	ctx := context.Background()
	fldPath := field.NewPath("test")

	testCases := []struct {
		name     string
		input    string
		wantErrs field.ErrorList
	}{{
		name:     "valid uuid with hyphens",
		input:    "123e4567-e89b-12d3-a456-426614174000",
		wantErrs: nil,
	}, {
		name:  "invalid uuid with hyphens uppercase",
		input: "123E4567-E89B-12D3-A456-426614174000",
		wantErrs: field.ErrorList{
			field.Invalid(fldPath, "123E4567-E89B-12D3-A456-426614174000", "must be a lowercase UUID in 8-4-4-4-12 format").WithOrigin("format=k8s-uuid"),
		},
	}, {
		name:  "invalid uuid with urn prefix",
		input: "urn:uuid:123e4567-e89b-12d3-a456-426614174000",
		wantErrs: field.ErrorList{
			field.Invalid(fldPath, "urn:uuid:123e4567-e89b-12d3-a456-426614174000", "must be a lowercase UUID in 8-4-4-4-12 format").WithOrigin("format=k8s-uuid"),
		},
	}, {
		name:  "invalid uuid without hyphens",
		input: "123e4567e89b12d3a456426614174000",
		wantErrs: field.ErrorList{
			field.Invalid(fldPath, "123e4567e89b12d3a456426614174000", "must be a lowercase UUID in 8-4-4-4-12 format").WithOrigin("format=k8s-uuid"),
		},
	}, {
		name:  "invalid: wrong length",
		input: "123e4567-e89b-12d3-a456-42661417400",
		wantErrs: field.ErrorList{
			field.Invalid(fldPath, "123e4567-e89b-12d3-a456-42661417400", "must be a lowercase UUID in 8-4-4-4-12 format").WithOrigin("format=k8s-uuid"),
		},
	}, {
		name:  "invalid: wrong characters",
		input: "123e4567-e89b-12d3-a456-42661417400g",
		wantErrs: field.ErrorList{
			field.Invalid(fldPath, "123e4567-e89b-12d3-a456-42661417400g", "must be a lowercase UUID in 8-4-4-4-12 format").WithOrigin("format=k8s-uuid"),
		},
	}, {
		name:  "invalid: misplaced hyphens",
		input: "123e4567-e89b-12d3-a4564-26614174000",
		wantErrs: field.ErrorList{
			field.Invalid(fldPath, "123e4567-e89b-12d3-a4564-26614174000", "must be a lowercase UUID in 8-4-4-4-12 format").WithOrigin("format=k8s-uuid"),
		},
	}, {
		name:  "empty string",
		input: "",
		wantErrs: field.ErrorList{
			field.Invalid(fldPath, "", "must be a lowercase UUID in 8-4-4-4-12 format").WithOrigin("format=k8s-uuid"),
		},
	}, {
		name:  "not a uuid",
		input: "not-a-uuid",
		wantErrs: field.ErrorList{
			field.Invalid(fldPath, "not-a-uuid", "must be a lowercase UUID in 8-4-4-4-12 format").WithOrigin("format=k8s-uuid"),
		},
	}}

	matcher := field.ErrorMatcher{}.ByType().ByField().ByOrigin().ByDetailExact()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value := tc.input
			gotErrs := UUID(ctx, operation.Operation{}, fldPath, &value, nil)
			matcher.Test(t, tc.wantErrs, gotErrs)
		})
	}
}
