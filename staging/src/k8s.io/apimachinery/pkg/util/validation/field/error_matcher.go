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
	"reflect"
	"regexp"
	"strings"
)

// ErrorMatcher is a helper for comparing Error objects.
type ErrorMatcher struct {
	matchType                bool
	matchField               bool
	matchValue               bool
	matchOrigin              bool
	matchDetail              func(want, got string) bool
	requireOriginWhenInvalid bool
	// pathTranslations stores bidirectional mappings of equivalent field paths.
	pathTranslations map[string]string
}

// Matches returns true if the two Error objects match according to the
// configured criteria.
func (m ErrorMatcher) Matches(want, got *Error) bool {
	if m.matchType && want.Type != got.Type {
		return false
	}
	if m.matchField {
		wantPath := want.Field
		gotPath := got.Field
		isMatch := (wantPath == gotPath)

		// If it's not a direct match, check for a configured translation.
		if !isMatch && m.pathTranslations != nil {
			if translatedWant, ok := m.pathTranslations[wantPath]; ok && translatedWant == gotPath {
				isMatch = true
			}
		}

		if !isMatch {
			return false
		}
	}
	if m.matchValue && !reflect.DeepEqual(want.BadValue, got.BadValue) {
		return false
	}
	if m.matchOrigin {
		if want.Origin != got.Origin {
			return false
		}
		if m.requireOriginWhenInvalid && want.Type == ErrorTypeInvalid {
			if want.Origin == "" || got.Origin == "" {
				return false
			}
		}
	}
	if m.matchDetail != nil && !m.matchDetail(want.Detail, got.Detail) {
		return false
	}
	return true
}

// Render returns a string representation of the specified Error object,
// according to the criteria configured in the ErrorMatcher.
func (m ErrorMatcher) Render(e *Error) string {
	buf := strings.Builder{}

	comma := func() {
		if buf.Len() > 0 {
			buf.WriteString(", ")
		}
	}

	if m.matchType {
		comma()
		buf.WriteString(fmt.Sprintf("Type=%q", e.Type))
	}
	if m.matchField {
		comma()
		fieldStr := fmt.Sprintf("Field=%q", e.Field)
		if m.pathTranslations != nil {
			// Clarify in the output that path translation is active for this matcher.
			fieldStr = fmt.Sprintf("Field(translated)=%q", e.Field)
		}
		buf.WriteString(fieldStr)
	}
	if m.matchValue {
		comma()
		if s, ok := e.BadValue.(string); ok {
			buf.WriteString(fmt.Sprintf("Value=%q", s))
		} else {
			rv := reflect.ValueOf(e.BadValue)
			if rv.Kind() == reflect.Pointer && !rv.IsNil() {
				rv = rv.Elem()
			}
			if rv.IsValid() && rv.CanInterface() {
				buf.WriteString(fmt.Sprintf("Value=%v", rv.Interface()))
			} else {
				buf.WriteString(fmt.Sprintf("Value=%v", e.BadValue))
			}
		}
	}
	if m.matchOrigin || m.requireOriginWhenInvalid && e.Type == ErrorTypeInvalid {
		comma()
		buf.WriteString(fmt.Sprintf("Origin=%q", e.Origin))
	}
	if m.matchDetail != nil {
		comma()
		buf.WriteString(fmt.Sprintf("Detail=%q", e.Detail))
	}
	return "{" + buf.String() + "}"
}

// Exactly returns a derived ErrorMatcher which matches all fields exactly.
func (m ErrorMatcher) Exactly() ErrorMatcher {
	return m.ByType().ByField().ByValue().ByOrigin().ByDetailExact()
}

// ByType returns a derived ErrorMatcher which also matches by type.
func (m ErrorMatcher) ByType() ErrorMatcher {
	m.matchType = true
	return m
}

// ByField returns a derived ErrorMatcher which also matches by field path.
func (m ErrorMatcher) ByField() ErrorMatcher {
	m.matchField = true
	return m
}

// ByValue returns a derived ErrorMatcher which also matches by the errant
// value.
func (m ErrorMatcher) ByValue() ErrorMatcher {
	m.matchValue = true
	return m
}

// ByOrigin returns a derived ErrorMatcher which also matches by the origin.
// When this is used and an origin is set in the error, the matcher will
// consider all expected errors with the same origin to be a match. The only
// expception to this is when it finds two errors which are exactly identical,
// which is too suspicious to ignore. This multi-matching allows tests to
// express a single expectation ("I set the X field to an invalid value, and I
// expect an error from origin Y") without having to know exactly how many
// errors might be returned, or in what order, or with what wording.
func (m ErrorMatcher) ByOrigin() ErrorMatcher {
	m.matchOrigin = true
	return m
}

// WithPathTranslations returns a derived ErrorMatcher which considers field paths
// in the provided map to be equivalent. This is useful when testing across API
// versions where a field may have moved. For example, providing `{"a.b": "c.d"}`
// will cause the matcher to treat an error on field "a.b" and an error on field
// "c.d" as having the same path. Mappings are bidirectional.
func (m ErrorMatcher) WithPathTranslations(translations map[string]string) ErrorMatcher {
	// Create a new map that contains both forward and reverse mappings to
	// simplify the logic in the Matches() function.
	bidirectionalMap := make(map[string]string, len(translations)*2)
	for from, to := range translations {
		bidirectionalMap[from] = to
		bidirectionalMap[to] = from
	}
	m.pathTranslations = bidirectionalMap
	return m
}

// RequireOriginWhenInvalid returns a derived ErrorMatcher which also requires
// the Origin field to be set when the Type is Invalid and the matcher is
// matching by Origin.
func (m ErrorMatcher) RequireOriginWhenInvalid() ErrorMatcher {
	m.requireOriginWhenInvalid = true
	return m
}

// ByDetailExact returns a derived ErrorMatcher which also matches errors by
// the exact detail string.
func (m ErrorMatcher) ByDetailExact() ErrorMatcher {
	m.matchDetail = func(want, got string) bool {
		return got == want
	}
	return m
}

// ByDetailSubstring returns a derived ErrorMatcher which also matches errors
// by a substring of the detail string.
func (m ErrorMatcher) ByDetailSubstring() ErrorMatcher {
	m.matchDetail = func(want, got string) bool {
		return strings.Contains(got, want)
	}
	return m
}

// ByDetailRegexp returns a derived ErrorMatcher which also matches errors by a
// regular expression of the detail string, where the "want" string is assumed
// to be a valid regular expression.
func (m ErrorMatcher) ByDetailRegexp() ErrorMatcher {
	m.matchDetail = func(want, got string) bool {
		return regexp.MustCompile(want).MatchString(got)
	}
	return m
}

// TestIntf lets users pass a testing.T while not coupling this package to Go's
// testing package.
type TestIntf interface {
	Helper()
	Errorf(format string, args ...any)
}

// Test compares two ErrorLists by the criteria configured in this matcher, and
// fails the test if they don't match. If matching by origin is enabled and the
// error has a non-empty origin, a given "want" error can match multiple
// "got" errors, and they will all be consumed. The only exception to this is
// if the matcher got multiple identical (in every way, even those not being
// matched on) errors, which is likely to indicate a bug.
func (m ErrorMatcher) Test(tb TestIntf, want, got ErrorList) {
	tb.Helper()

	exactly := m.Exactly() // makes a copy

	// If we ever find an EXACT duplicate error, it's almost certainly a bug
	// worth reporting. If we ever find a use-case where this is not a bug, we
	// can revisit this assumption.
	seen := map[string]bool{}
	for _, g := range got {
		key := exactly.Render(g)
		if seen[key] {
			tb.Errorf("exact duplicate error:\n%s", key)
		}
		seen[key] = true
	}

	remaining := got
	for _, w := range want {
		tmp := make(ErrorList, 0, len(remaining))
		matched := false
		for i, g := range remaining {
			if m.Matches(w, g) {
				matched = true
				if m.matchOrigin && w.Origin != "" {
					// When origin is included in the match, we allow multiple
					// matches against the same wanted error, so that tests
					// can be insulated from the exact number, order, and
					// wording of cases that might return more than one error.
					continue
				} else {
					// Single-match, save the rest of the "got" errors and move
					// on to the next "want" error.
					tmp = append(tmp, remaining[i+1:]...)
					break
				}
			} else {
				tmp = append(tmp, g)
			}
		}
		if !matched {
			tb.Errorf("expected an error matching:\n%s", m.Render(w))
		}
		remaining = tmp
	}
	if len(remaining) > 0 {
		for _, e := range remaining {
			tb.Errorf("unmatched error:\n%s", exactly.Render(e))
		}
	}
}
