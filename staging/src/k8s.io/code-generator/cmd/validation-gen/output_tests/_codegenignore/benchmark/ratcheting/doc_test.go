package ratcheting

import (
	"context"
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/api/operation"
	"k8s.io/apimachinery/pkg/api/validate"
	"k8s.io/apimachinery/pkg/util/validation/field"
	apivalidation "k8s.io/kubernetes/pkg/apis/core/validation"
)

// NoRatcheting - Baseline implementation with no item-level validation ratcheting
func EachSliceValNoRatcheting[T any](ctx context.Context, op operation.Operation, fldPath *field.Path, newSlice, oldSlice []T,
	match, equiv validate.CompareFunc[T], validator validate.ValidateFunc[*T]) field.ErrorList {
	var errs field.ErrorList
	for i, val := range newSlice {
		// Always validate, no ratcheting
		errs = append(errs, validator(ctx, op, fldPath.Index(i), &val, nil)...)
	}
	return errs
}

// MapBasedLookup - Ratcheting using a map for efficient lookups
// This is an alternative implementation that uses map-based lookup instead of linear search
func EachSliceValMapBased[T any](ctx context.Context, op operation.Operation, fldPath *field.Path, newSlice, oldSlice []T,
	match, equiv validate.CompareFunc[T], validator validate.ValidateFunc[*T]) field.ErrorList {
	var errs field.ErrorList

	// Build lookup map for old slice if needed
	var oldMap map[string]*T
	if match != nil && len(oldSlice) > 0 {
		oldMap = make(map[string]*T, len(oldSlice))
		for i := range oldSlice {
			// Use the match function to create a key for the map
			// This is a simplified approach - in practice, you'd need a proper key function
			key := fmt.Sprintf("%v", oldSlice[i]) // Simple string representation as key
			oldMap[key] = &oldSlice[i]
		}
	}

	for i, val := range newSlice {
		var old *T
		if oldMap != nil {
			// Use the same key generation for lookup
			key := fmt.Sprintf("%v", val)
			if found, exists := oldMap[key]; exists {
				old = found
			}
		}

		// If the operation is an update, for validation ratcheting, skip re-validating if the old
		// value exists and either:
		// 1. The match function provides full comparison (equiv is nil)
		// 2. The equiv function confirms the values are equivalent
		if op.Type == operation.Update && old != nil && (equiv == nil || equiv(val, *old)) {
			continue
		}
		errs = append(errs, validator(ctx, op, fldPath.Index(i), &val, old)...)
	}
	return errs
}

// LinearScan - Current implementation using linear search (for comparison)
// This uses the actual lookup function from each.go
func EachSliceValLinearScan[T any](ctx context.Context, op operation.Operation, fldPath *field.Path, newSlice, oldSlice []T,
	match, equiv validate.CompareFunc[T], validator validate.ValidateFunc[*T]) field.ErrorList {
	var errs field.ErrorList
	for i, val := range newSlice {
		var old *T
		if match != nil && len(oldSlice) > 0 {
			old = lookup(oldSlice, val, match)
		}
		// If the operation is an update, for validation ratcheting, skip re-validating if the old
		// value exists and either:
		// 1. The match function provides full comparison (equiv is nil)
		// 2. The equiv function confirms the values are equivalent
		if op.Type == operation.Update && old != nil && (equiv == nil || equiv(val, *old)) {
			continue
		}
		errs = append(errs, validator(ctx, op, fldPath.Index(i), &val, old)...)
	}
	return errs
}

// lookup returns a pointer to the first element in the list that matches the
// target, according to the provided comparison function, or else nil.
// This is copied from each.go to ensure consistency
func lookup[T any](list []T, target T, cmp func(T, T) bool) *T {
	for i := range list {
		if cmp(list[i], target) {
			return &list[i]
		}
	}
	return nil
}

func validateEndpoint(ctx context.Context, op operation.Operation, fldPath *field.Path, obj, oldObj *Endpoint) field.ErrorList {
	return apivalidation.ValidateDNS1123Label(obj.Name, fldPath.Child("name"))
}

// validateEndpointHeavy simulates a moderately heavy validation function that would benefit from ratcheting
func validateEndpointHeavy(ctx context.Context, op operation.Operation, fldPath *field.Path, obj, oldObj *Endpoint) field.ErrorList {
	var errs field.ErrorList

	// 1. DNS1123 validation
	errs = append(errs, apivalidation.ValidateDNS1123Label(obj.Name, fldPath.Child("name"))...)

	// 2. Simulate additional validation work
	if len(obj.Name) > 0 {
		// Simulate some computational work
		sum := 0
		for _, char := range obj.Name {
			sum += int(char)
		}

		// Simulate string operations based on the sum
		for i := 0; i < sum%20; i++ {
			_ = fmt.Sprintf("validation-%s-%d", obj.Name, i)
		}
	}

	return errs
}

// generateTestData creates test data with the specified number of endpoints and change rate
// changeRate should be between 0.0 and 1.0, where:
// - 0.0 means no changes (all items are the same)
// - 1.0 means all items changed (all items are different)
// - 0.9 means 90% unchanged (90% of items are the same)
func generateTestData(count int, changeRate float64) ([]Endpoint, []Endpoint) {
	newSlice := make([]Endpoint, count)
	oldSlice := make([]Endpoint, count)

	for i := 0; i < count; i++ {
		newSlice[i] = Endpoint{
			Name: fmt.Sprintf("endpoint-%d", i),
		}

		// Determine if this item should be unchanged based on the rate
		shouldBeUnchanged := float64(i)/float64(count) < (1.0 - changeRate)

		if shouldBeUnchanged {
			// Same data (no change)
			oldSlice[i] = newSlice[i]
		} else {
			// Different data (changed)
			oldSlice[i] = Endpoint{
				Name: fmt.Sprintf("endpoint-%d-different", i),
			}
		}
	}

	return newSlice, oldSlice
}

// BenchmarkRatchetingApproaches - Benchmark different ratcheting approaches
// with 1000 endpoints and 10% changes.
func BenchmarkRatchetingApproaches(b *testing.B) {
	const endpointCount = 1000

	// Generate test data with ~10% changes
	newSlice, oldSlice := generateTestData(endpointCount, 0.1)

	// Create operation context
	ctx := context.Background()
	op := operation.Operation{Type: operation.Update}
	fldPath := field.NewPath("endpoints")

	// Test cases
	testCases := []struct {
		name     string
		function func(context.Context, operation.Operation, *field.Path, []Endpoint, []Endpoint, validate.CompareFunc[Endpoint], validate.CompareFunc[Endpoint], validate.ValidateFunc[*Endpoint]) field.ErrorList
	}{
		{"NoRatcheting", EachSliceValNoRatcheting[Endpoint]},
		{"MapBasedLookup", EachSliceValMapBased[Endpoint]},
		{"LinearScan", EachSliceValLinearScan[Endpoint]},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = tc.function(ctx, op, fldPath, newSlice, oldSlice, validate.DirectEqual[Endpoint], validate.DirectEqual[Endpoint], validateEndpoint)
			}
		})
	}
}

// Benchmark different change scenarios to show ratcheting impact
func BenchmarkChangeScenarios(b *testing.B) {
	const endpointCount = 1000

	scenarios := []struct {
		name       string
		changeRate float64
	}{
		{"LowChange", 0.1},    // 10% changes (90% unchanged)
		{"MediumChange", 0.3}, // 30% changes (70% unchanged)
		{"HighChange", 0.5},   // 50% changes (50% unchanged)
		{"NoChange", 0.0},     // 0% changes (100% unchanged)
	}

	for _, scenario := range scenarios {
		newSlice, oldSlice := generateTestData(endpointCount, scenario.changeRate)
		ctx := context.Background()
		op := operation.Operation{Type: operation.Update}
		fldPath := field.NewPath("endpoints")

		b.Run(fmt.Sprintf("LinearScan_%s", scenario.name), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = EachSliceValLinearScan[Endpoint](ctx, op, fldPath, newSlice, oldSlice, validate.DirectEqual[Endpoint], validate.SemanticDeepEqual[Endpoint], validateEndpoint)
			}
		})

		b.Run(fmt.Sprintf("MapBasedLookup_%s", scenario.name), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = EachSliceValMapBased[Endpoint](ctx, op, fldPath, newSlice, oldSlice, validate.DirectEqual[Endpoint], validate.SemanticDeepEqual[Endpoint], validateEndpoint)
			}
		})

		b.Run(fmt.Sprintf("NoRatcheting_%s", scenario.name), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = EachSliceValNoRatcheting[Endpoint](ctx, op, fldPath, newSlice, oldSlice, validate.DirectEqual[Endpoint], validate.SemanticDeepEqual[Endpoint], validateEndpoint)
			}
		})
	}
}

// Benchmark heavy validation to demonstrate ratcheting benefits
func BenchmarkHeavyValidation(b *testing.B) {
	const endpointCount = 100
	const changeRate = 0.1 // 10% changes (90% unchanged)

	newSlice, oldSlice := generateTestData(endpointCount, changeRate)
	ctx := context.Background()
	op := operation.Operation{Type: operation.Update}
	fldPath := field.NewPath("endpoints")

	b.Run("HeavyValidation_LinearScan", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = EachSliceValLinearScan[Endpoint](ctx, op, fldPath, newSlice, oldSlice, validate.DirectEqual[Endpoint], validate.SemanticDeepEqual[Endpoint], validateEndpointHeavy)
		}
	})

	b.Run("HeavyValidation_MapBasedLookup", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = EachSliceValMapBased[Endpoint](ctx, op, fldPath, newSlice, oldSlice, validate.DirectEqual[Endpoint], validate.SemanticDeepEqual[Endpoint], validateEndpointHeavy)
		}
	})

	b.Run("HeavyValidation_NoRatcheting", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = EachSliceValNoRatcheting[Endpoint](ctx, op, fldPath, newSlice, oldSlice, validate.DirectEqual[Endpoint], validate.SemanticDeepEqual[Endpoint], validateEndpointHeavy)
		}
	})
}
