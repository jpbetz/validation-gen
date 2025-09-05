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

package testing

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	featuregatetesting "k8s.io/component-base/featuregate/testing"
	"k8s.io/kubernetes/pkg/features"
)

// ValidateFunc is a function that runs validation.
type ValidateFunc func(ctx context.Context, obj runtime.Object) field.ErrorList

// ValidateUpdateFunc is a function that runs update validation.
type ValidateUpdateFunc func(ctx context.Context, obj, old runtime.Object) field.ErrorList

// TestDeclarativeValidation runs a test for declarative validation.
// It runs the validation function with and without the DeclarativeValidation and
// DeclarativeValidationTakeover feature gates and compares the results.
func TestDeclarativeValidation(t *testing.T, ctx context.Context, obj runtime.Object, validateFn ValidateFunc, expectedErrs field.ErrorList) {
	t.Helper()
	var declarativeTakeoverErrs field.ErrorList
	var imperativeErrs field.ErrorList
	for _, gateVal := range []bool{true, false} {
		// We only need to test both gate enabled and disabled together, because
		// 1) the DeclarativeValidationTakeover won't take effect if DeclarativeValidation is disabled.
		// 2) the validation output, when only DeclarativeValidation is enabled, is the same as when both gates are disabled.
		featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.DeclarativeValidation, gateVal)
		featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.DeclarativeValidationTakeover, gateVal)

		errs := validateFn(ctx, obj)
		if gateVal {
			declarativeTakeoverErrs = errs
		} else {
			imperativeErrs = errs
		}
		// The errOutputMatcher is used to verify the output matches the expected errors in test cases.
		errOutputMatcher := field.ErrorMatcher{}.ByType().ByField().ByOrigin()
		if len(expectedErrs) > 0 {
			errOutputMatcher.Test(t, expectedErrs, errs)
		} else if len(errs) != 0 {
			t.Errorf("expected no errors, but got: %v", errs)
		}
	}
	// The equivalenceMatcher is used to verify the output errors from hand-written imperative validation
	// are equivalent to the output errors when DeclarativeValidationTakeover is enabled.
	equivalenceMatcher := field.ErrorMatcher{}.ByType().ByField().ByOrigin()
	equivalenceMatcher.Test(t, imperativeErrs, declarativeTakeoverErrs)

	VerifyVersionedValidationEquivalence(t, obj, nil)
}

// TestDeclarativeUpdateValidation runs a test for declarative update validation.
// It runs the validation function with and without the DeclarativeValidation and
// DeclarativeValidationTakeover feature gates and compares the results.
func TestDeclarativeUpdateValidation(t *testing.T, ctx context.Context, obj, old runtime.Object, validateUpdateFn ValidateUpdateFunc, expectedErrs field.ErrorList) {
	t.Helper()
	var declarativeTakeoverErrs field.ErrorList
	var imperativeErrs field.ErrorList
	for _, gateVal := range []bool{true, false} {
		// We only need to test both gate enabled and disabled together, because
		// 1) the DeclarativeValidationTakeover won't take effect if DeclarativeValidation is disabled.
		// 2) the validation output, when only DeclarativeValidation is enabled, is the same as when both gates are disabled.
		featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.DeclarativeValidation, gateVal)
		featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.DeclarativeValidationTakeover, gateVal)
		errs := validateUpdateFn(ctx, obj, old)
		if gateVal {
			declarativeTakeoverErrs = errs
		} else {
			imperativeErrs = errs
		}
		// The errOutputMatcher is used to verify the output matches the expected errors in test cases.
		errOutputMatcher := field.ErrorMatcher{}.ByType().ByField().ByOrigin()

		if len(expectedErrs) > 0 {
			errOutputMatcher.Test(t, expectedErrs, errs)
		} else if len(errs) != 0 {
			t.Errorf("expected no errors, but got: %v", errs)
		}
	}
	// The equivalenceMatcher is used to verify the output errors from hand-written imperative validation
	// are equivalent to the output errors when DeclarativeValidationTakeover is enabled.
	equivalenceMatcher := field.ErrorMatcher{}.ByType().ByField().ByOrigin()
	// TODO: remove this once ErrorMatcher has been extended to handle this form of deduplication.
	dedupedImperativeErrs := field.ErrorList{}
	for _, err := range imperativeErrs {
		found := false
		for _, existingErr := range dedupedImperativeErrs {
			if equivalenceMatcher.Matches(existingErr, err) {
				found = true
				break
			}
		}
		if !found {
			dedupedImperativeErrs = append(dedupedImperativeErrs, err)
		}
	}
	equivalenceMatcher.Test(t, dedupedImperativeErrs, declarativeTakeoverErrs)

	VerifyVersionedValidationEquivalence(t, obj, old)
}
