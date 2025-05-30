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

package validate

import (
	"context"

	"k8s.io/apimachinery/pkg/api/operation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// MatchFn takes a pointer to an item and returns true if it matches the criteria.
type MatchFn[T any] func(*T) bool

// ListMapItemByKeyValues finds the first item in oldList (if any) and the first
// item in newList (if any) that satisfy the 'matches' predicate. It then invokes
// 'itemValidator' on these items (if items are found).
// The fldPath passed to itemValidator is indexed to the matched item's position
// using the index from newList if a match is found there, otherwise the root list index.
// This function processes only the *first* matching item found in each list.
// It assumes that the 'matches' predicate, targets a unique identifier (PK) and
// will match at most one element per list.
// If this assumption is violated, changes in list order can lead this function
// to have inconsistent behavior.
func ListMapItemByKeyValues[TList ~[]TItem, TItem any](
	ctx context.Context, op operation.Operation, fldPath *field.Path,
	newList, oldList TList,
	matches MatchFn[TItem],
	itemValidator func(ctx context.Context, op operation.Operation, fldPath *field.Path, newObj, oldObj *TItem) field.ErrorList,
) field.ErrorList {
	var matchedNew, matchedOld *TItem
	path := fldPath

	for i := range oldList {
		if matches(&oldList[i]) {
			matchedOld = &oldList[i]
			break
		}
	}
	for i := range newList {
		if matches(&newList[i]) {
			matchedNew = &newList[i]
			// Use newList index when available.
			// For deleted items (only in oldList), use root path as there's no meaningful index in current input.
			path = fldPath.Index(i)
			break
		}
	}
	if matchedNew == nil && matchedOld == nil {
		return nil
	}
	return itemValidator(ctx, op, path, matchedNew, matchedOld)
}
