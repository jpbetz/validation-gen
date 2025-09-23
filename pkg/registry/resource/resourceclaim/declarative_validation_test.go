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

package resourceclaim

import (
	"fmt"
	"strings"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/client-go/kubernetes/fake"
	featuregatetesting "k8s.io/component-base/featuregate/testing"
	apitesting "k8s.io/kubernetes/pkg/api/testing"
	"k8s.io/kubernetes/pkg/apis/resource"
	"k8s.io/kubernetes/pkg/features"
	pointer "k8s.io/utils/ptr"
)

var apiVersions = []string{"v1beta1", "v1beta2", "v1"}

// v1beta1PathTranslations maps field paths from the flat v1beta1 API structure
// to the nested 'exactly' structure of the internal and v1/v1beta2 APIs. This
// allows the ErrorMatcher to correctly compare validation errors across versions.
var v1beta1PathTranslations = map[string]string{
	"spec.devices.requests[0].allocationMode": "spec.devices.requests[0].exactly.allocationMode",
	"spec.devices.requests[0].count":          "spec.devices.requests[0].exactly.count",
	"spec.devices.requests[0].adminAccess":    "spec.devices.requests[0].exactly.adminAccess",
	"spec.devices.requests[0].tolerations":    "spec.devices.requests[0].exactly.tolerations",
	"spec.devices.requests[0].selectors":      "spec.devices.requests[0].exactly.selectors",
}

func TestDeclarativeValidate(t *testing.T) {
	for _, apiVersion := range apiVersions {
		t.Run(apiVersion, func(t *testing.T) {
			testDeclarativeValidate(t, apiVersion)
		})
	}
}

func testDeclarativeValidate(t *testing.T, apiVersion string) {
	ctx := genericapirequest.WithRequestInfo(genericapirequest.NewDefaultContext(), &genericapirequest.RequestInfo{
		APIGroup:   "resource.k8s.io",
		APIVersion: apiVersion,
		Resource:   "resourceclaims",
	})
	fakeClient := fake.NewSimpleClientset()
	mockNSClient := fakeClient.CoreV1().Namespaces()
	Strategy := NewStrategy(mockNSClient)
	testCases := map[string]struct {
		input        resource.ResourceClaim
		expectedErrs field.ErrorList
	}{
		"valid": {
			input: mkValidResourceClaim(),
		},
		"valid requests, max allowed": {
			input: mkValidResourceClaim(tweakDevicesConfigs(32)),
		},
		"valid constraints, max allowed": {
			input: mkValidResourceClaim(tweakDevicesConstraints(32)),
		},
		"valid config, max allowed": {
			input: mkValidResourceClaim(tweakDevicesRequests(32)),
		},
		"invalid requests, too many": {
			input: mkValidResourceClaim(tweakDevicesConfigs(33)),
			expectedErrs: field.ErrorList{
				field.TooMany(field.NewPath("spec", "devices", "requests"), 33, 32),
			},
		},
		"invalid constraints, too many": {
			input: mkValidResourceClaim(tweakDevicesConstraints(33)),
			expectedErrs: field.ErrorList{
				field.TooMany(field.NewPath("spec", "devices", "constraints"), 33, 32),
			},
		},
		"invalid config, too many": {
			input: mkValidResourceClaim(tweakDevicesRequests(33)),
			expectedErrs: field.ErrorList{
				field.TooMany(field.NewPath("spec", "devices", "config"), 33, 32),
			},
		},
		"valid DeviceAllocationMode - ExactCount": {
			input: mkValidResourceClaim(tweakAllocationMode(resource.DeviceAllocationModeExactCount, 5)),
		},
		"valid DeviceAllocationMode - All": {
			input: mkValidResourceClaim(tweakAllocationMode(resource.DeviceAllocationModeAll, 0)),
		},
		"invalid DeviceAllocationMode": {
			input: mkValidResourceClaim(tweakAllocationMode("InvalidMode", 0)),
			expectedErrs: field.ErrorList{
				field.NotSupported(
					field.NewPath("spec", "devices", "requests").Index(0).Child("exactly", "allocationMode"),
					"InvalidMode",
					[]string{"All", "ExactCount"},
				).WithOrigin("enum"),
			},
		},
		"invalid DeviceAllocationMode - empty string": {
			input: mkValidResourceClaim(tweakAllocationMode("", 0)),
			expectedErrs: field.ErrorList{
				field.Required(
					field.NewPath("spec", "devices", "requests").Index(0).Child("exactly", "allocationMode"),
					"must be either 'All' or 'ExactCount'",
				),
			},
		},
		"invalid DeviceAllocationMode - case sensitive": {
			input: mkValidResourceClaim(tweakAllocationMode("exactcount", 1)),
			expectedErrs: field.ErrorList{
				field.NotSupported(
					field.NewPath("spec", "devices", "requests").Index(0).Child("exactly", "allocationMode"),
					"exactcount",
					[]string{"All", "ExactCount"},
				).WithOrigin("enum"),
			},
		},
		"valid DeviceAllocationMode in FirstAvailable": {
			input: mkValidResourceClaim(tweakFirstAvailableAllocationMode(resource.DeviceAllocationModeAll)),
		},
		"invalid DeviceAllocationMode in FirstAvailable": {
			input: mkValidResourceClaim(tweakFirstAvailableAllocationMode("BadMode")),
			expectedErrs: field.ErrorList{
				field.NotSupported(
					field.NewPath("spec", "devices", "requests").Index(0).Child("firstAvailable").Index(0).Child("allocationMode"),
					"BadMode",
					[]string{"All", "ExactCount"},
				).WithOrigin("enum"),
			},
		},
	}
	for k, tc := range testCases {
		t.Run(k, func(t *testing.T) {
			var declarativeTakeoverErrs field.ErrorList
			var imperativeErrs field.ErrorList
			for _, gateVal := range []bool{true, false} {
				featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.DeclarativeValidation, gateVal)
				featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.DeclarativeValidationTakeover, gateVal)
				if strings.Contains(k, "FirstAvailable") {
					featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.DRAPrioritizedList, true)
				}

				errs := Strategy.Validate(ctx, &tc.input)
				if gateVal {
					declarativeTakeoverErrs = errs
				} else {
					imperativeErrs = errs
				}
			}

			// Create a matcher that is aware of the path differences between API versions.
			equivalenceMatcher := field.ErrorMatcher{}.ByType().ByField().ByOrigin().WithPathTranslations(v1beta1PathTranslations)
			equivalenceMatcher.Test(t, imperativeErrs, declarativeTakeoverErrs)

			// Now that the imperative validation logic is correct, all tests should pass versioned validation.
			apitesting.VerifyVersionedValidationEquivalence(t, &tc.input, nil)
		})
	}
}

func tweakDevicesConfigs(items int) func(*resource.ResourceClaim) {
	return func(rc *resource.ResourceClaim) {
		for i := 0; i < items; i++ {
			rc.Spec.Devices.Config = append(rc.Spec.Devices.Config, mkDeviceClaimConfiguration())
		}
	}
}

func tweakDevicesConstraints(items int) func(*resource.ResourceClaim) {
	return func(rc *resource.ResourceClaim) {
		for i := 0; i < items; i++ {
			rc.Spec.Devices.Constraints = append(rc.Spec.Devices.Constraints, mkDeviceConstraint())
		}
	}
}

func tweakDevicesRequests(items int) func(*resource.ResourceClaim) {
	return func(rc *resource.ResourceClaim) {
		for i := 0; i < items; i++ {
			rc.Spec.Devices.Requests = append(rc.Spec.Devices.Requests, mkDeviceRequest(fmt.Sprintf("req-%d", i)))
		}
	}
}

func tweakAllocationMode(mode resource.DeviceAllocationMode, count int64) func(*resource.ResourceClaim) {
	return func(rc *resource.ResourceClaim) {
		if len(rc.Spec.Devices.Requests) > 0 && rc.Spec.Devices.Requests[0].Exactly != nil {
			rc.Spec.Devices.Requests[0].Exactly.AllocationMode = mode
			rc.Spec.Devices.Requests[0].Exactly.Count = count
		}
	}
}

func tweakFirstAvailableAllocationMode(mode resource.DeviceAllocationMode) func(*resource.ResourceClaim) {
	return func(rc *resource.ResourceClaim) {
		if len(rc.Spec.Devices.Requests) > 0 {
			// Clear Exactly and set FirstAvailable
			rc.Spec.Devices.Requests[0].Exactly = nil
			rc.Spec.Devices.Requests[0].FirstAvailable = []resource.DeviceSubRequest{
				{
					Name:            "sub1",
					DeviceClassName: "class1",
					AllocationMode:  mode,
				},
			}
		}
	}
}

func mkDeviceClaimConfiguration() resource.DeviceClaimConfiguration {
	return resource.DeviceClaimConfiguration{
		Requests: []string{"req-0"},
	}
}

func mkDeviceConstraint() resource.DeviceConstraint {
	return resource.DeviceConstraint{
		Requests:       []string{"req-0"},
		MatchAttribute: pointer.To(resource.FullyQualifiedName("a")),
	}
}

func mkDeviceRequest(name string) resource.DeviceRequest {
	return resource.DeviceRequest{
		Name: name,
		Exactly: &resource.ExactDeviceRequest{
			DeviceClassName: "class",
			AllocationMode:  resource.DeviceAllocationModeAll,
		},
	}
}

func TestDeclarativeValidateUpdate(t *testing.T) {
	for _, apiVersion := range apiVersions {
		t.Run(apiVersion, func(t *testing.T) {
			testDeclarativeValidateUpdate(t, apiVersion)
		})
	}
}

func testDeclarativeValidateUpdate(t *testing.T, apiVersion string) {
	ctx := genericapirequest.WithRequestInfo(genericapirequest.NewDefaultContext(), &genericapirequest.RequestInfo{
		APIGroup:   "resource.k8s.io",
		APIVersion: apiVersion,
		Resource:   "resourceclaims",
	})
	fakeClient := fake.NewSimpleClientset()
	mockNSClient := fakeClient.CoreV1().Namespaces()
	Strategy := NewStrategy(mockNSClient)
	validClaim := mkValidResourceClaim()
	testCases := map[string]struct {
		update               resource.ResourceClaim
		old                  resource.ResourceClaim
		expectedErrs         field.ErrorList
		skipEquivalenceCheck bool
	}{
		"valid": {
			update: validClaim,
			old:    validClaim,
		},
		"valid DeviceAllocationMode update - same value": {
			update: mkValidResourceClaim(tweakAllocationMode(resource.DeviceAllocationModeExactCount, 3)),
			old:    mkValidResourceClaim(tweakAllocationMode(resource.DeviceAllocationModeExactCount, 3)),
		},
		"invalid DeviceAllocationMode update - invalid new value": {
			update: mkValidResourceClaim(tweakAllocationMode("NewBadMode", 0)),
			old:    mkValidResourceClaim(),
			// Spec immutability check may prevent enum validation from running
			skipEquivalenceCheck: true,
		},
	}
	for k, tc := range testCases {
		t.Run(k, func(t *testing.T) {
			// Set resource versions for updates
			tc.old.ObjectMeta.ResourceVersion = "1"
			tc.update.ObjectMeta.ResourceVersion = "1"

			var declarativeTakeoverErrs field.ErrorList
			var imperativeErrs field.ErrorList
			for _, gateVal := range []bool{true, false} {
				featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.DeclarativeValidation, gateVal)
				featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.DeclarativeValidationTakeover, gateVal)

				errs := Strategy.ValidateUpdate(ctx, &tc.update, &tc.old)
				if gateVal {
					declarativeTakeoverErrs = errs
				} else {
					imperativeErrs = errs
				}
			}

			if tc.skipEquivalenceCheck {
				// For update tests where spec immutability prevents enum validation,
				// just ensure both produce errors (the exact errors may differ)
				if len(imperativeErrs) == 0 {
					t.Errorf("expected imperative validation errors but got none")
				}
				if len(declarativeTakeoverErrs) == 0 {
					t.Errorf("expected declarative validation errors but got none")
				}
			} else {
				equivalenceMatcher := field.ErrorMatcher{}.ByType().ByField().ByOrigin().WithPathTranslations(v1beta1PathTranslations)
				equivalenceMatcher.Test(t, imperativeErrs, declarativeTakeoverErrs)

				if !tc.skipEquivalenceCheck {
					apitesting.VerifyVersionedValidationEquivalence(t, &tc.update, &tc.old)
				}
			}
		})
	}
}

func TestValidateStatusUpdateForDeclarative(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()
	mockNSClient := fakeClient.CoreV1().Namespaces()
	Strategy := NewStrategy(mockNSClient)
	strategy := NewStatusStrategy(Strategy)

	ctx := genericapirequest.WithRequestInfo(genericapirequest.NewDefaultContext(), &genericapirequest.RequestInfo{
		APIGroup:    "resource.k8s.io",
		APIVersion:  "v1",
		Subresource: "status",
	})
	poolPath := field.NewPath("status", "allocation", "devices", "results").Index(0).Child("pool")
	testCases := map[string]struct {
		old          resource.ResourceClaim
		update       resource.ResourceClaim
		expectedErrs field.ErrorList
	}{
		"valid pool name": {
			old:    mkValidResourceClaim(),
			update: mkResourceClaimWithStatus(tweakStatusDeviceRequestAllocationResultPool("dra.example.com/pool-a")),
		},
		"valid pool name, max length": {
			old:    mkValidResourceClaim(),
			update: mkResourceClaimWithStatus(tweakStatusDeviceRequestAllocationResultPool(strings.Repeat("a", 63) + "." + strings.Repeat("b", 63) + "." + strings.Repeat("c", 63) + "." + strings.Repeat("d", 55))),
		},
		"invalid pool name, required": {
			old:    mkValidResourceClaim(),
			update: mkResourceClaimWithStatus(tweakStatusDeviceRequestAllocationResultPool("")),
			expectedErrs: field.ErrorList{
				field.Required(poolPath, ""),
			},
		},
		"invalid pool name, too long": {
			old:    mkValidResourceClaim(),
			update: mkResourceClaimWithStatus(tweakStatusDeviceRequestAllocationResultPool(strings.Repeat("a", 253) + "/" + strings.Repeat("a", 253))),
			expectedErrs: field.ErrorList{
				field.TooLong(poolPath, "", 253).WithOrigin("format=k8s-resource-pool-name"),
			},
		},
		"invalid pool name, format": {
			old:    mkValidResourceClaim(),
			update: mkResourceClaimWithStatus(tweakStatusDeviceRequestAllocationResultPool("a/Not_Valid")),
			expectedErrs: field.ErrorList{
				field.Invalid(poolPath, "Not_Valid", "").WithOrigin("format=k8s-resource-pool-name"),
			},
		},
		"invalid pool name, leading slash": {
			old:    mkValidResourceClaim(),
			update: mkResourceClaimWithStatus(tweakStatusDeviceRequestAllocationResultPool("/a")),
			expectedErrs: field.ErrorList{
				field.Invalid(poolPath, "", "").WithOrigin("format=k8s-resource-pool-name"),
			},
		},
		"invalid pool name, trailing slash": {
			old:    mkValidResourceClaim(),
			update: mkResourceClaimWithStatus(tweakStatusDeviceRequestAllocationResultPool("a/")),
			expectedErrs: field.ErrorList{
				field.Invalid(poolPath, "", "").WithOrigin("format=k8s-resource-pool-name"),
			},
		},
		"invalid pool name, double slash": {
			old:    mkValidResourceClaim(),
			update: mkResourceClaimWithStatus(tweakStatusDeviceRequestAllocationResultPool("a//b")),
			expectedErrs: field.ErrorList{
				field.Invalid(poolPath, "", "").WithOrigin("format=k8s-resource-pool-name"),
			},
		},
		"valid status.allocation unchanged": {
			old:    mkResourceClaimWithStatus(),
			update: mkResourceClaimWithStatus(),
		},
		"valid status.allocation set from nil": {
			old:    mkValidResourceClaim(),
			update: mkResourceClaimWithStatus(),
		},
		"valid status.allocation cleared (Unset is allowed)": {
			old:    mkResourceClaimWithStatus(),
			update: mkValidResourceClaim(),
		},
		"invalid status.allocation changed device (NoModify)": {
			old:    mkResourceClaimWithStatus(),
			update: tweakStatusAllocationDevice(mkResourceClaimWithStatus(), "device-different"),
			expectedErrs: field.ErrorList{
				field.Invalid(field.NewPath("status", "allocation"), nil, "field is immutable").WithOrigin("update"),
			},
		},
		"invalid status.allocation changed driver (NoModify)": {
			old:    mkResourceClaimWithStatus(),
			update: tweakStatusAllocationDriver(mkResourceClaimWithStatus(), "different.example.com"),
			expectedErrs: field.ErrorList{
				field.Invalid(field.NewPath("status", "allocation"), nil, "field is immutable").WithOrigin("update"),
			},
		},
		"invalid status.allocation changed pool (NoModify)": {
			old:    mkResourceClaimWithStatus(),
			update: tweakStatusAllocationPool(mkResourceClaimWithStatus(), "different-pool"),
			expectedErrs: field.ErrorList{
				field.Invalid(field.NewPath("status", "allocation"), nil, "field is immutable").WithOrigin("update"),
			},
		},
		"invalid status.allocation added result (NoModify)": {
			old:    mkResourceClaimWithStatus(),
			update: addStatusAllocationResult(mkResourceClaimWithStatus()),
			expectedErrs: field.ErrorList{
				field.Invalid(field.NewPath("status", "allocation"), nil, "field is immutable").WithOrigin("update"),
			},
		},
		"invalid status.allocation removed result (NoModify)": {
			old:    addStatusAllocationResult(mkResourceClaimWithStatus()),
			update: mkResourceClaimWithStatus(),
			expectedErrs: field.ErrorList{
				field.Invalid(field.NewPath("status", "allocation"), nil, "field is immutable").WithOrigin("update"),
			},
		},
	}
	for k, tc := range testCases {
		t.Run(k, func(t *testing.T) {
			tc.old.ObjectMeta.ResourceVersion = "1"
			tc.update.ObjectMeta.ResourceVersion = "1"
			var declarativeTakeoverErrs field.ErrorList
			var imperativeErrs field.ErrorList
			for _, gateVal := range []bool{true, false} {
				t.Run(fmt.Sprintf("gate=%v", gateVal), func(t *testing.T) {
					featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.DeclarativeValidation, gateVal)
					featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.DeclarativeValidationTakeover, gateVal)
					errs := strategy.ValidateUpdate(ctx, &tc.update, &tc.old)
					if gateVal {
						declarativeTakeoverErrs = errs
					} else {
						imperativeErrs = errs
					}
					errOutputMatcher := field.ErrorMatcher{}.ByType().ByField().ByOrigin()

					if len(tc.expectedErrs) > 0 {
						errOutputMatcher.Test(t, tc.expectedErrs, errs)
					} else if len(errs) != 0 {
						t.Errorf("expected no errors, but got: %v", errs)
					}
				})
			}
			equivalenceMatcher := field.ErrorMatcher{}.ByType().ByField().ByOrigin()
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

			apitesting.VerifyVersionedValidationEquivalence(t, &tc.update, &tc.old)
		})
	}
}

func mkValidResourceClaim(tweaks ...func(rc *resource.ResourceClaim)) resource.ResourceClaim {
	rc := resource.ResourceClaim{
		ObjectMeta: v1.ObjectMeta{
			Name:      "valid-claim",
			Namespace: "default",
		},
		Spec: resource.ResourceClaimSpec{
			Devices: resource.DeviceClaim{
				Requests: []resource.DeviceRequest{
					{
						Name: "req-0",
						Exactly: &resource.ExactDeviceRequest{
							DeviceClassName: "class",
							AllocationMode:  resource.DeviceAllocationModeAll,
						},
					},
				},
			},
		},
	}

	for _, tweak := range tweaks {
		tweak(&rc)
	}
	return rc
}

func mkResourceClaimWithStatus(tweaks ...func(rc *resource.ResourceClaim)) resource.ResourceClaim {
	rc := resource.ResourceClaim{
		ObjectMeta: v1.ObjectMeta{
			Name:      "valid-claim",
			Namespace: "default",
		},
		Spec: resource.ResourceClaimSpec{
			Devices: resource.DeviceClaim{
				Requests: []resource.DeviceRequest{
					{
						Name: "req-0",
						Exactly: &resource.ExactDeviceRequest{
							DeviceClassName: "class",
							AllocationMode:  resource.DeviceAllocationModeAll,
						},
					},
				},
			},
		},
		Status: resource.ResourceClaimStatus{
			Allocation: &resource.AllocationResult{
				Devices: resource.DeviceAllocationResult{
					Results: []resource.DeviceRequestAllocationResult{
						{
							Request: "req-0",
							Driver:  "dra.example.com",
							Pool:    "pool-0",
							Device:  "device-0",
						},
					},
				},
			},
		},
	}
	for _, tweak := range tweaks {
		tweak(&rc)
	}
	return rc
}

func tweakStatusDeviceRequestAllocationResultPool(pool string) func(rc *resource.ResourceClaim) {
	return func(rc *resource.ResourceClaim) {
		for i := range rc.Status.Allocation.Devices.Results {
			rc.Status.Allocation.Devices.Results[i].Pool = pool
		}
	}
}

func tweakStatusAllocationDevice(obj resource.ResourceClaim, device string) resource.ResourceClaim {
	if obj.Status.Allocation != nil && len(obj.Status.Allocation.Devices.Results) > 0 {
		obj.Status.Allocation.Devices.Results[0].Device = device
	}
	return obj
}

func tweakStatusAllocationDriver(obj resource.ResourceClaim, driver string) resource.ResourceClaim {
	if obj.Status.Allocation != nil && len(obj.Status.Allocation.Devices.Results) > 0 {
		obj.Status.Allocation.Devices.Results[0].Driver = driver
	}
	return obj
}

func tweakStatusAllocationPool(obj resource.ResourceClaim, pool string) resource.ResourceClaim {
	if obj.Status.Allocation != nil && len(obj.Status.Allocation.Devices.Results) > 0 {
		obj.Status.Allocation.Devices.Results[0].Pool = pool
	}
	return obj
}

func addStatusAllocationResult(obj resource.ResourceClaim) resource.ResourceClaim {
	if obj.Status.Allocation != nil {
		obj.Status.Allocation.Devices.Results = append(obj.Status.Allocation.Devices.Results,
			resource.DeviceRequestAllocationResult{
				Request: "req-0",
				Driver:  "another.example.com",
				Pool:    "pool-1",
				Device:  "device-1",
			})
	}
	return obj
}
