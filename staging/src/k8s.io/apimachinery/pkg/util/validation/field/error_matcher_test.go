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

package field

import (
	"fmt"
	"strings"
	"testing"
)

func TestErrorMatcher_Matches(t *testing.T) {
	baseErr := func() *Error {
		return &Error{
			Type:     ErrorTypeInvalid,
			Field:    "field",
			BadValue: "value",
			Detail:   "detail",
			Origin:   "origin",
		}
	}

	testCases := []struct {
		name      string
		matcher   ErrorMatcher
		wantedErr func() *Error
		actualErr *Error
		matches   bool
	}{{
		name:      "ByType: match",
		matcher:   ErrorMatcher{}.ByType(),
		wantedErr: baseErr,
		actualErr: &Error{Type: ErrorTypeInvalid},
		matches:   true,
	}, {
		name:      "ByType: no match",
		matcher:   ErrorMatcher{}.ByType(),
		wantedErr: baseErr,
		actualErr: &Error{Type: ErrorTypeRequired},
		matches:   false,
	}, {
		name:      "ByField: match",
		matcher:   ErrorMatcher{}.ByField(),
		wantedErr: baseErr,
		actualErr: &Error{Field: "field"},
		matches:   true,
	}, {
		name:      "ByField: no match",
		matcher:   ErrorMatcher{}.ByField(),
		wantedErr: baseErr,
		actualErr: &Error{Field: "other"},
		matches:   false,
	}, {
		name: "ByField with translations: match same path",
		matcher: ErrorMatcher{}.ByField(map[string]string{
			`spec\.devices\.requests\[(\d+)\]\.allocationMode`: "spec.devices.requests[$1].exactly.allocationMode",
		}),
		wantedErr: func() *Error {
			e := baseErr()
			e.Field = "spec.devices.requests[0].allocationMode"
			return e
		},
		actualErr: &Error{Field: "spec.devices.requests[0].exactly.allocationMode"},
		matches:   true,
	}, {
		name: "ByField with translations: match translated path",
		matcher: ErrorMatcher{}.ByField(map[string]string{
			`spec\.devices\.requests\[(\d+)\]\.allocationMode`: "spec.devices.requests[$1].exactly.allocationMode",
		}),
		wantedErr: func() *Error {
			e := baseErr()
			e.Field = "spec.devices.requests[0].exactly.allocationMode"
			return e
		},
		actualErr: &Error{Field: "spec.devices.requests[0].allocationMode"},
		matches:   true,
	}, {
		name: "ByField with translations: no match different index",
		matcher: ErrorMatcher{}.ByField(map[string]string{
			`spec\.devices\.requests\[(\d+)\]\.allocationMode`: "spec.devices.requests[$1].exactly.allocationMode",
		}),
		wantedErr: func() *Error {
			e := baseErr()
			e.Field = "spec.devices.requests[0].allocationMode"
			return e
		},
		actualErr: &Error{Field: "spec.devices.requests[1].exactly.allocationMode"},
		matches:   false,
	}, {
		name: "ByField with translations: multiple translations",
		matcher: ErrorMatcher{}.ByField(map[string]string{
			`spec\.devices\.requests\[(\d+)\]\.deviceClassName`: "spec.devices.requests[$1].exactly.deviceClassName",
			`spec\.devices\.requests\[(\d+)\]\.count`:           "spec.devices.requests[$1].exactly.count",
		}),
		wantedErr: func() *Error {
			e := baseErr()
			e.Field = "spec.devices.requests[2].count"
			return e
		},
		actualErr: &Error{Field: "spec.devices.requests[2].exactly.count"},
		matches:   true,
	}, {
		name: "ByField with translations: no translation needed",
		matcher: ErrorMatcher{}.ByField(map[string]string{
			`spec\.devices\.requests\[(\d+)\]\.allocationMode`: "spec.devices.requests[$1].exactly.allocationMode",
		}),
		wantedErr: func() *Error {
			e := baseErr()
			e.Field = "spec.other.field"
			return e
		},
		actualErr: &Error{Field: "spec.other.field"},
		matches:   true,
	}, {
		name:      "ByValue: match",
		matcher:   ErrorMatcher{}.ByValue(),
		wantedErr: baseErr,
		actualErr: &Error{BadValue: "value"},
		matches:   true,
	}, {
		name:      "ByValue: no match",
		matcher:   ErrorMatcher{}.ByValue(),
		wantedErr: baseErr,
		actualErr: &Error{BadValue: "other"},
		matches:   false,
	}, {
		name:      "ByOrigin: match",
		matcher:   ErrorMatcher{}.ByOrigin(),
		wantedErr: baseErr,
		actualErr: &Error{Origin: "origin"},
		matches:   true,
	}, {
		name:      "ByOrigin: no match",
		matcher:   ErrorMatcher{}.ByOrigin(),
		wantedErr: baseErr,
		actualErr: &Error{Origin: "other"},
		matches:   false,
	}, {
		name:      "ByDetailExact: match",
		matcher:   ErrorMatcher{}.ByDetailExact(),
		wantedErr: baseErr,
		actualErr: &Error{Detail: "detail"},
		matches:   true,
	}, {
		name:      "ByDetailExact: no match",
		matcher:   ErrorMatcher{}.ByDetailExact(),
		wantedErr: baseErr,
		actualErr: &Error{Detail: "other"},
		matches:   false,
	}, {
		name:    "ByDetailSubstring: match empty",
		matcher: ErrorMatcher{}.ByDetailSubstring(),
		wantedErr: func() *Error {
			e := baseErr()
			e.Detail = ""
			return e
		},
		actualErr: &Error{Detail: "this is the detail"},
		matches:   true,
	}, {
		name:    "ByDetailSubstring: match full",
		matcher: ErrorMatcher{}.ByDetailSubstring(),
		wantedErr: func() *Error {
			e := baseErr()
			e.Detail = "is the"
			return e
		},
		actualErr: &Error{Detail: "this is the detail"},
		matches:   true,
	}, {
		name:    "ByDetailSubstring: match start",
		matcher: ErrorMatcher{}.ByDetailSubstring(),
		wantedErr: func() *Error {
			e := baseErr()
			e.Detail = "this is"
			return e
		},
		actualErr: &Error{Detail: "this is the detail"},
		matches:   true,
	}, {
		name:    "ByDetailSubstring: match middle",
		matcher: ErrorMatcher{}.ByDetailSubstring(),
		wantedErr: func() *Error {
			e := baseErr()
			e.Detail = "is the"
			return e
		},
		actualErr: &Error{Detail: "this is the detail"},
		matches:   true,
	}, {
		name:    "ByDetailSubstring: match end",
		matcher: ErrorMatcher{}.ByDetailSubstring(),
		wantedErr: func() *Error {
			e := baseErr()
			e.Detail = "the detail"
			return e
		},
		actualErr: &Error{Detail: "this is the detail"},
		matches:   true,
	}, {
		name:    "ByDetailSubstring: no match",
		matcher: ErrorMatcher{}.ByDetailSubstring(),
		wantedErr: func() *Error {
			e := baseErr()
			e.Detail = "is not the"
			return e
		},
		actualErr: &Error{Detail: "this is the detail"},
		matches:   false,
	}, {
		name:    "ByDetailRegexp: match empty",
		matcher: ErrorMatcher{}.ByDetailRegexp(),
		wantedErr: func() *Error {
			e := baseErr()
			e.Detail = ".*"
			return e
		},
		actualErr: &Error{Detail: "this is the detail"},
		matches:   true,
	}, {
		name:    "ByDetailRegexp: match full",
		matcher: ErrorMatcher{}.ByDetailRegexp(),
		wantedErr: func() *Error {
			e := baseErr()
			e.Detail = "^this is the detail$"
			return e
		},
		actualErr: &Error{Detail: "this is the detail"},
		matches:   true,
	}, {
		name:    "ByDetailRegexp: match start",
		matcher: ErrorMatcher{}.ByDetailRegexp(),
		wantedErr: func() *Error {
			e := baseErr()
			e.Detail = "^this is"
			return e
		},
		actualErr: &Error{Detail: "this is the detail"},
		matches:   true,
	}, {
		name:    "ByDetailRegexp: match middle",
		matcher: ErrorMatcher{}.ByDetailRegexp(),
		wantedErr: func() *Error {
			e := baseErr()
			e.Detail = "is the"
			return e
		},
		actualErr: &Error{Detail: "this is the detail"},
		matches:   true,
	}, {
		name:    "ByDetailRegexp: match end",
		matcher: ErrorMatcher{}.ByDetailRegexp(),
		wantedErr: func() *Error {
			e := baseErr()
			e.Detail = "the detail$"
			return e
		},
		actualErr: &Error{Detail: "this is the detail"},
		matches:   true,
	}, {
		name:    "ByDetailRegexp: match parts",
		matcher: ErrorMatcher{}.ByDetailRegexp(),
		wantedErr: func() *Error {
			e := baseErr()
			e.Detail = "^this .* .* detail$"
			return e
		},
		actualErr: &Error{Detail: "this is the detail"},
		matches:   true,
	}, {
		name:    "ByDetailRegexp: no match",
		matcher: ErrorMatcher{}.ByDetailRegexp(),
		wantedErr: func() *Error {
			e := baseErr()
			e.Detail = "is not the"
			return e
		},
		actualErr: &Error{Detail: "this is the detail"},
		matches:   false,
	}, {
		name:      "Exactly: match",
		matcher:   ErrorMatcher{}.Exactly(),
		wantedErr: baseErr,
		actualErr: baseErr(),
		matches:   true,
	}, {
		name:      "Exactly: no match (type)",
		matcher:   ErrorMatcher{}.Exactly(),
		wantedErr: baseErr,
		actualErr: &Error{Type: ErrorTypeRequired, Field: "field", BadValue: "value", Detail: "detail", Origin: "origin"},
		matches:   false,
	}, {
		name:    "ByDeclarativeOnly: match",
		matcher: ErrorMatcher{}.ByDeclarativeOnly(),
		wantedErr: func() *Error {
			e := baseErr()
			e.DeclarativeOnly = true
			return e
		},
		actualErr: &Error{DeclarativeOnly: true},
		matches:   true,
	}, {
		name:    "ByDeclarativeOnly: no match",
		matcher: ErrorMatcher{}.ByDeclarativeOnly(),
		wantedErr: func() *Error {
			e := baseErr()
			e.DeclarativeOnly = true
			return e
		},
		actualErr: &Error{DeclarativeOnly: false},
		matches:   false,
	}, {
		name:      "RequireOriginWhenInvalid: match",
		matcher:   ErrorMatcher{}.ByOrigin().RequireOriginWhenInvalid(),
		wantedErr: baseErr,
		actualErr: &Error{Type: ErrorTypeInvalid, Origin: "origin"},
		matches:   true,
	}, {
		name:      "RequireOriginWhenInvalid: no match (missing origin)",
		matcher:   ErrorMatcher{}.ByOrigin().RequireOriginWhenInvalid(),
		wantedErr: baseErr,
		actualErr: &Error{Type: ErrorTypeInvalid},
		matches:   false,
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.matcher.Matches(tc.wantedErr(), tc.actualErr) != tc.matches {
				t.Errorf("Matches() = %v, want %v", !tc.matches, tc.matches)
			}
		})
	}
}

// fakeTestIntf is used to test the testing support.
type fakeTestIntf struct {
	errs []string
	logs []string
}

var _ TestIntf = &fakeTestIntf{}

func (*fakeTestIntf) Helper() {}

func (ft *fakeTestIntf) Errorf(format string, args ...any) {
	ft.errs = append(ft.errs, fmt.Sprintf(format, args...))
}

func (ft *fakeTestIntf) Logf(format string, args ...any) {
	ft.logs = append(ft.logs, fmt.Sprintf(format, args...))
}

func TestErrorMatcher_Test(t *testing.T) {
	testCases := []struct {
		name           string
		matcher        ErrorMatcher
		want           ErrorList
		got            ErrorList
		expectedErrors []string
		expectedLogs   []string
	}{{
		name:    "no origin: perfect match",
		matcher: ErrorMatcher{}.ByField(),
		want:    ErrorList{Invalid(NewPath("f"), nil, "")},
		got:     ErrorList{Invalid(NewPath("f"), "v", "d")},
	}, {
		name:           "no origin: got too few errors",
		matcher:        ErrorMatcher{}.ByField(),
		want:           ErrorList{Invalid(NewPath("f"), nil, "")},
		got:            ErrorList{},
		expectedErrors: []string{"expected an error matching:"},
	}, {
		name:           "no origin: got too many errors",
		matcher:        ErrorMatcher{}.ByField(),
		want:           ErrorList{},
		got:            ErrorList{Invalid(NewPath("f"), "v", "d")},
		expectedErrors: []string{"unmatched error:"},
	}, {
		name:           "no origin: got wrong errors",
		matcher:        ErrorMatcher{}.ByField(),
		want:           ErrorList{Invalid(NewPath("f1"), nil, "")},
		got:            ErrorList{Invalid(NewPath("f2"), "v", "d")},
		expectedErrors: []string{"expected an error matching:", "unmatched error:"},
	}, {
		name: "path translations: match v1beta1 to v1",
		matcher: ErrorMatcher{}.ByField(map[string]string{
			`spec\.devices\.requests\[(\d+)\]\.allocationMode`: "spec.devices.requests[$1].exactly.allocationMode",
		}),
		want: ErrorList{Invalid(NewPath("spec", "devices", "requests").Index(0).Child("allocationMode"), nil, "")},
		got:  ErrorList{Invalid(NewPath("spec", "devices", "requests").Index(0).Child("exactly", "allocationMode"), "InvalidMode", "unsupported value")},
	}, {
		name: "path translations: match v1 to v1beta1",
		matcher: ErrorMatcher{}.ByField(map[string]string{
			`spec\.devices\.requests\[(\d+)\]\.allocationMode`: "spec.devices.requests[$1].exactly.allocationMode",
		}),
		want: ErrorList{Invalid(NewPath("spec", "devices", "requests").Index(0).Child("exactly", "allocationMode"), nil, "")},
		got:  ErrorList{Invalid(NewPath("spec", "devices", "requests").Index(0).Child("allocationMode"), "InvalidMode", "unsupported value")},
	}, {
		name: "path translations: multiple fields",
		matcher: ErrorMatcher{}.ByField(map[string]string{
			`spec\.devices\.requests\[(\d+)\]\.allocationMode`:  "spec.devices.requests[$1].exactly.allocationMode",
			`spec\.devices\.requests\[(\d+)\]\.deviceClassName`: "spec.devices.requests[$1].exactly.deviceClassName",
			`spec\.devices\.requests\[(\d+)\]\.count`:           "spec.devices.requests[$1].exactly.count",
		}),
		want: ErrorList{
			Invalid(NewPath("spec", "devices", "requests").Index(0).Child("allocationMode"), nil, ""),
			Invalid(NewPath("spec", "devices", "requests").Index(1).Child("deviceClassName"), nil, ""),
			Invalid(NewPath("spec", "devices", "requests").Index(2).Child("count"), nil, ""),
		},
		got: ErrorList{
			Invalid(NewPath("spec", "devices", "requests").Index(0).Child("exactly", "allocationMode"), "InvalidMode", "unsupported value"),
			Invalid(NewPath("spec", "devices", "requests").Index(1).Child("exactly", "deviceClassName"), "", "required"),
			Invalid(NewPath("spec", "devices", "requests").Index(2).Child("exactly", "count"), 0, "must be greater than zero"),
		},
	}, {
		name: "path translations: no match for different indices",
		matcher: ErrorMatcher{}.ByField(map[string]string{
			`spec\.devices\.requests\[(\d+)\]\.allocationMode`: "spec.devices.requests[$1].exactly.allocationMode",
		}),
		want:           ErrorList{Invalid(NewPath("spec", "devices", "requests").Index(0).Child("allocationMode"), nil, "")},
		got:            ErrorList{Invalid(NewPath("spec", "devices", "requests").Index(1).Child("exactly", "allocationMode"), "InvalidMode", "unsupported value")},
		expectedErrors: []string{"expected an error matching:", "unmatched error:"},
	}, {
		name:    "declarative only: match",
		matcher: ErrorMatcher{}.ByDeclarativeOnly(),
		want:    ErrorList{{DeclarativeOnly: true}},
		got:     ErrorList{{DeclarativeOnly: true}},
	}, {
		name:           "declarative only: no match",
		matcher:        ErrorMatcher{}.ByDeclarativeOnly(),
		want:           ErrorList{{DeclarativeOnly: true}},
		got:            ErrorList{{DeclarativeOnly: false}},
		expectedErrors: []string{"expected an error matching:", "unmatched error:"},
	}, {
		name:    "with origin: single match",
		matcher: ErrorMatcher{}.ByField().ByOrigin(),
		want:    ErrorList{Invalid(NewPath("f"), nil, "").WithOrigin("o")},
		got:     ErrorList{Invalid(NewPath("f"), "v", "d").WithOrigin("o")},
	}, {
		name:    "with origin: multiple matches, different details",
		matcher: ErrorMatcher{}.ByField().ByOrigin(),
		want: ErrorList{
			Invalid(NewPath("f1"), nil, "").WithOrigin("o"),
			Invalid(NewPath("f2"), nil, "").WithOrigin("o"),
		},
		got: ErrorList{
			Invalid(NewPath("f1"), "v", "d1").WithOrigin("o"),
			Invalid(NewPath("f2"), "v", "d1").WithOrigin("o"),
			Invalid(NewPath("f1"), "v", "d2").WithOrigin("o"),
			Invalid(NewPath("f2"), "v", "d2").WithOrigin("o"),
		},
		expectedLogs: []string{"multiple errors matched:", "multiple errors matched:"},
	}, {
		name:    "with origin: multiple matches, same exact error",
		matcher: ErrorMatcher{}.ByField().ByOrigin(),
		want: ErrorList{
			Invalid(NewPath("f1"), nil, "").WithOrigin("o"),
			Invalid(NewPath("f2"), nil, "").WithOrigin("o"),
		},
		got: ErrorList{
			Invalid(NewPath("f1"), "v", "d").WithOrigin("o"),
			Invalid(NewPath("f1"), "v", "d").WithOrigin("o"),
			Invalid(NewPath("f2"), "v", "d").WithOrigin("o"),
			Invalid(NewPath("f2"), "v", "d").WithOrigin("o"),
		},
		expectedLogs: []string{"multiple errors matched:", "multiple errors matched:"},
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fakeT := &fakeTestIntf{}
			tc.matcher.Test(fakeT, tc.want, tc.got)
			if want, got := len(tc.expectedErrors), len(fakeT.errs); got != want {
				if got == 0 {
					t.Errorf("expected %d errors, got %d", want, got)
				} else {
					q := make([]string, len(fakeT.errs))
					for i, err := range fakeT.errs {
						q[i] = fmt.Sprintf("%q", err)
					}
					t.Errorf("expected %d errors, got %d:\n%s", want, got, strings.Join(q, "\n"))
				}
			} else {
				for i := range tc.expectedErrors {
					if !strings.HasPrefix(fakeT.errs[i], tc.expectedErrors[i]) {
						t.Errorf("error %d: expected prefix %q, got %q", i, tc.expectedErrors[i], fakeT.errs[i])
					}
				}
			}
			if want, got := len(tc.expectedLogs), len(fakeT.logs); got != want {
				if got == 0 && want > 0 {
					t.Errorf("expected %d logs, got %d", want, got)
				} else if got > 0 {
					q := make([]string, len(fakeT.logs))
					for i, log := range fakeT.logs {
						q[i] = fmt.Sprintf("%q", log)
					}
					t.Errorf("expected %d logs, got %d:\n%s", want, got, strings.Join(q, "\n"))
				}
			} else {
				for i := range tc.expectedLogs {
					if !strings.HasPrefix(fakeT.logs[i], tc.expectedLogs[i]) {
						t.Errorf("log %d: expected prefix %q, got %q", i, tc.expectedLogs[i], fakeT.logs[i])
					}
				}
			}
		})
	}
}

func TestErrorMatcher_Render(t *testing.T) {
	testCases := []struct {
		name     string
		matcher  ErrorMatcher
		err      *Error
		expected string
	}{
		{
			name:     "empty matcher",
			matcher:  ErrorMatcher{},
			err:      Invalid(NewPath("field"), "value", "detail"),
			expected: "{}",
		},
		{
			name:     "single field - type",
			matcher:  ErrorMatcher{}.ByType(),
			err:      Invalid(NewPath("field"), "value", "detail"),
			expected: `{Type="Invalid value"}`,
		},
		{
			name:     "single field - value with string",
			matcher:  ErrorMatcher{}.ByValue(),
			err:      Invalid(NewPath("field"), "string_value", "detail"),
			expected: `{Value="string_value"}`,
		},
		{
			name:     "single field - value with nil",
			matcher:  ErrorMatcher{}.ByValue(),
			err:      Invalid(NewPath("field"), nil, "detail"),
			expected: `{Value=<nil>}`,
		},
		{
			name:     "multiple fields",
			matcher:  ErrorMatcher{}.ByType().ByField().ByValue(),
			err:      Invalid(NewPath("field"), "value", "detail"),
			expected: `{Type="Invalid value", Field="field", Value="value"}`,
		},
		{
			name:     "all fields",
			matcher:  ErrorMatcher{}.ByType().ByField().ByValue().ByOrigin().ByDetailExact(),
			err:      Invalid(NewPath("field"), "value", "detail").WithOrigin("origin"),
			expected: `{Type="Invalid value", Field="field", Value="value", Origin="origin", Detail="detail"}`,
		},
		{
			name: "with path translation - translated",
			matcher: ErrorMatcher{}.ByField(map[string]string{
				`spec\.devices\.requests\[(\d+)\]\.allocationMode`: "spec.devices.requests[$1].exactly.allocationMode",
			}),
			err:      Invalid(NewPath("spec", "devices", "requests").Index(0).Child("allocationMode"), "value", "detail"),
			expected: `{Field="spec.devices.requests[0].exactly.allocationMode" (translated from "spec.devices.requests[0].allocationMode")}`,
		},
		{
			name: "with path translation - no translation needed",
			matcher: ErrorMatcher{}.ByField(map[string]string{
				`spec\.devices\.requests\[(\d+)\]\.allocationMode`: "spec.devices.requests[$1].exactly.allocationMode",
			}),
			err:      Invalid(NewPath("spec", "other", "field"), "value", "detail"),
			expected: `{Field="spec.other.field"}`,
		},
		{
			name: "with path translation - already canonical",
			matcher: ErrorMatcher{}.ByField(map[string]string{
				`spec\.devices\.requests\[(\d+)\]\.allocationMode`: "spec.devices.requests[$1].exactly.allocationMode",
			}),
			err:      Invalid(NewPath("spec", "devices", "requests").Index(0).Child("exactly", "allocationMode"), "value", "detail"),
			expected: `{Field="spec.devices.requests[0].exactly.allocationMode"}`,
		},
		{
			name:    "with covered by declarative",
			matcher: ErrorMatcher{}.ByDeclarativeOnly(),
			err: func() *Error {
				e := Invalid(NewPath("field"), "value", "detail")
				e.DeclarativeOnly = true
				return e
			}(),
			expected: `{DeclarativeOnly=true}`,
		},
		{
			name:    "all fields with covered by declarative",
			matcher: ErrorMatcher{}.ByType().ByField().ByValue().ByOrigin().ByDetailExact().ByDeclarativeOnly(),
			err: func() *Error {
				e := Invalid(NewPath("field"), "value", "detail").WithOrigin("origin")
				e.DeclarativeOnly = true
				return e
			}(),
			expected: `{Type="Invalid value", Field="field", Value="value", Origin="origin", Detail="detail", DeclarativeOnly=true}`,
		},
		{
			name:     "requireOriginWhenInvalid with origin",
			matcher:  ErrorMatcher{}.ByOrigin().RequireOriginWhenInvalid(),
			err:      Invalid(NewPath("field"), "value", "detail").WithOrigin("origin"),
			expected: `{Origin="origin"}`,
		},
		{
			name:     "different error types",
			matcher:  ErrorMatcher{}.ByType().ByValue(),
			err:      Required(NewPath("field"), "detail"),
			expected: `{Type="Required value", Value=""}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.matcher.Render(tc.err)
			if result != tc.expected {
				t.Errorf("Render() = %v, want %v", result, tc.expected)
			}
		})
	}
}
