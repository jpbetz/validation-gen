# API Reviewers Guide for Declarative Validation (New APIs)
Starting in v1.36, Declarative Validation is the recommended way to author API validation
logic for Kubernetes (for supported validation cases).

This guide focuses specifically on the practical aspects of reviewing PRs that
**use DV for new APIs and fields**, where DV serves as the authoritative source
of truth from Day 1 with **no fallback handwritten code**.

> **Scope.** This guide covers the "happy path" for brand-new APIs that opt in
> to declarative enforcement. It does **not** cover migrating existing APIs from
> handwritten validation to DV. For migration guidance, see
> [MIGRATION_GUIDE.md](./MIGRATION_GUIDE.md).

---

## 1. The Happy Path: What to Expect in a PR

When a developer adds a new API or field and wants to use DV, the tags in
`types.go` are the *only* validation logic for simple validations (complex/non-standard validations will still need hand-written code in a validations.go file).

Before reviewing the logic, familiarize yourself with the
[Official Declarative Validation Tag Catalog](https://kubernetes.io/docs/reference/using-api/declarative-validation/).
This is your reference for what tags exist, what they do, and their stability
level.

### Step 0: `doc.go` (Enabling Code Generation)
Before tags can be used, the developer must ensure code generation is enabled for the API package.

```go
// +k8s:validation-gen=TypeMeta
// +k8s:validation-gen-input=k8s.io/api/<group>/<version>
package v1
```

### Step 1: `types.go` (The Single Source of Truth)

The developer adds standard DV tags directly to the new field. Verify these tags
make logical sense for the field's purpose and come from the official catalog.

If the API includes a `/status` subresource and validation is needed for it, the root type definition must include the `supportsSubresource` tag.

```go
// +k8s:supportsSubresource="/status"
type MyFeature struct {
    metav1.TypeMeta `json:",inline"`
    // ...
}

type MyNewFeatureSpec struct {
    // +required
    // +k8s:required
    // +k8s:maxLength=256
    // +k8s:format=k8s-short-name
    FeatureName string `json:"featureName"`

    // +required
    // +k8s:required
    // +k8s:minimum=0
    // +k8s:maximum=1000
    Replicas int32 `json:"replicas"`

    // +optional
    // +k8s:maxLength=4096
    Description string `json:"description,omitempty"`
}
```

**What to look for:**

- **Tags come from the official catalog.** Every `+k8s:` tag should correspond
  to an entry in the
  [tag reference](https://kubernetes.io/docs/reference/using-api/declarative-validation/).
  Invented or misspelled tags are silently ignored — this is one of the most
  dangerous failure modes because validation-gen will not error, and the field
  will simply go unvalidated.

- **Tags are applied across all API versions.** If the resource has both `v1`
  and `v1beta1` definitions, the tags must appear on both. Missing tags on one
  version means that version is unvalidated. *(Note: There are tests to ensure that this doesn't happen, but be cautious and manually verify as a best practice).*

- **No handwritten `Validate*` functions for the same constraints.** For a new
  DV-only API, the tags are authoritative. The only handwritten validation that
  should exist is for things DV cannot express (see
  [FAQ: What about cross-field validation?](#q-what-about-cross-field-validation)).

### Step 2: `strategy.go` (The Plumbing)

For new APIs to use these tags authoritatively, the strategy must be explicitly
told to enforce them using `rest.WithDeclarativeEnforcement()`.

```go
func (myStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
    // If there is complex cross-field validation, it still lives here
    allErrs := validation.ValidateMyFeature(obj.(*myapi.MyFeature))

    return rest.ValidateDeclarativelyWithMigrationChecks(
        ctx,
        legacyscheme.Scheme,
        obj,
        nil,
        allErrs,
        operation.Create,
        rest.WithDeclarativeEnforcement(), // <--- Critical for New APIs
    )
}
```

**What to look for:**

- **`rest.WithDeclarativeEnforcement()` is present.** This is the single most
  important line in the PR. Without it, every tag in `types.go` is dead code for
  validation purposes — the API will accept any input. See
  [Mistake 1](#-mistake-1-missing-the-enforcement-flag) below.

- **Cross-field validation coexists correctly.** If the API has cross-field
  constraints, the handwritten `Validate*` function's errors should be passed
  into `ValidateDeclarativelyWithMigrationChecks` via the `allErrs` parameter.
  DV handles per-field tags; cross-field logic stays handwritten.

### Step 3: `validation_test.go` (The Tests)

Tests for DV rely on marking expected errors to confirm they came from the DV
framework, not handwritten code. Because `WithDeclarativeEnforcement()` is used,
errors from standard tags are marked as "Non-Shadowed" (meaning they are actively
enforced).

```go
"feature name too long": {
    obj: mkFeature(func(f *myapi.MyFeature) {
        f.Spec.FeatureName = strings.Repeat("a", 257)
    }),
    expectedErrs: field.ErrorList{
        field.TooLongMaxLength(field.NewPath("spec", "featureName"), 257, 256).MarkNonShadowed(),
    },
},
```

**What to look for:**

- **`.MarkNonShadowed()` is used on expected errors for standard tags.** This marks the error as
  coming from declarative validation in enforcement mode. If the PR involves an alpha or beta validation rule (`+k8s:alpha` or `+k8s:beta`), the tests should use `.MarkAlpha()` or `.MarkBeta()` respectively. If this is missing, the
  test framework may not correctly attribute the error source.

- **Tests verify wiring, not framework logic.** The test should confirm that the
  tag is applied to the right field and produces an error on invalid input. It
  should *not* exhaustively test the framework's implementation of a tag (e.g.,
  50 cases for DNS label validation). One or two cases per tag is sufficient —
  see [Mistake 3](#-mistake-3-over-testing-framework-logic) below.

### Step 4: `zz_generated.validations.go` (The Generated Code)

The PR must include the generated validation code. When a developer adds or
changes `+k8s:` tags in `types.go`, they must run
`hack/update-codegen.sh validation` to regenerate the validation functions. The
generated file is committed alongside the hand-authored changes.

The generated file lives at:

```
pkg/apis/<group>/<version>/zz_generated.validations.go
```

There is **one generated file per API group/version**. For example:

- `pkg/apis/core/v1/zz_generated.validations.go`
- `pkg/apis/admissionregistration/v1/zz_generated.validations.go`
- `pkg/apis/admissionregistration/v1beta1/zz_generated.validations.go`

If the API has multiple versions (e.g., `v1` and `v1beta1`), there will be a
separate generated file for each version, and tags must be present on the
`types.go` for each version to produce the corresponding generated code.

**What to look for:**

- **The generated file is present in the PR.** If a developer adds tags to
  `types.go` but forgets to run `make generate`, the generated code will be
  missing or stale. The PR should include changes to the
  `zz_generated.validations.go` file(s) that correspond to the tagged API
  versions.

- **The generated code reflects the tags.** You do not need to review the
  generated code line-by-line (it is auto-generated), but a quick glance can
  confirm that the expected validation functions are present for the fields that
  were tagged.

- **All relevant versions are regenerated.** If tags were added to both
  `admissionregistration/v1/types.go` and `admissionregistration/v1beta1/types.go`,
  both `pkg/apis/admissionregistration/v1/zz_generated.validations.go` and
  `pkg/apis/admissionregistration/v1beta1/zz_generated.validations.go` should
  show changes.

---

## 2. Example Reviews: Catching Common Mistakes

Here are examples of what you should actively look for and push back on during a
review for a new API.

### ❌ Mistake 1: Missing the Enforcement Flag

**The PR:** A developer adds `+k8s:required` to a new API field, writes tests
using `.MarkNonShadowed()`, but forgets to update `strategy.go`.

**Why it's bad:** Without `rest.WithDeclarativeEnforcement()`, the `+k8s:required`
tag is treated as an implicit shadow migration (the legacy behavior). It will
generate a metric mismatch, but it will *not* reject the invalid request. The API
is effectively unvalidated.

**Your Review Comment:**
> *"I see you're adding DV tags for this new API. As this is a new API and
> there's no handwritten fallback, you need to ensure these tags are actually
> enforced. Please add `rest.WithDeclarativeEnforcement()` to the
> `ValidateDeclarativelyWithMigrationChecks` call in `strategy.go`."*

### ❌ Mistake 2: Writing Handwritten Fallbacks

**The PR:** A developer adds `+k8s:minimum=1` to a new field in `types.go`, but
also writes `if spec.NewField < 1 { ... }` in `validation.go`.

**Why it's bad:** The primary benefit of DV for new APIs is eliminating
boilerplate Go code. Standard tags under `WithDeclarativeEnforcement()` are fully
binding and authoritative. Duplicating them in handwritten code creates two
sources of truth that can drift and doubles the maintenance burden.

**Your Review Comment:**
> *"Since we are using `WithDeclarativeEnforcement()` for this new API, the
> `+k8s:minimum=1` tag is authoritative. You can safely remove the equivalent
> handwritten check from `validation.go`."*

### ❌ Mistake 3: Over-Testing Framework Logic

**The PR:** A developer adds `+k8s:format=k8s-short-name` and proceeds to write
50 test cases in `validation_test.go` checking every conceivable valid and
invalid character combination for a DNS label.

**Why it's bad:** We trust the `validation-gen` framework to implement
`k8s-short-name` correctly (it has its own exhaustive tests). Reviewing an exhaustive matrix of tests on the resource itself isn't strictly necessary, though reviewers are free to ask for specific corner cases they care about.

**Your Review Comment:**
> *"Since we are using the standard `+k8s:format=k8s-short-name` tag, we don't
> necessarily need an exhaustive matrix to test the format itself here — the framework guarantees
> the base cases. Let's ensure we have a few basic valid/invalid examples, plus any specific corner cases you think are important for this specific field."*

---

## 3. FAQ & Common Pitfalls

**Q: What about cross-field validation?**

DV operates on individual fields. It cannot express constraints like "if field A
is set, field B must also be set" or "field X must be less than field Y." These
constraints must still be implemented in handwritten validation code and passed
into `ValidateDeclarativelyWithMigrationChecks` via the `allErrs` parameter.

For a new DV-only API, this means: simple per-field constraints go in tags, and
cross-field logic goes in a handwritten `Validate*` function that **only**
contains the cross-field checks.

**Q: Does DV short-circuit validation?**

Yes. If a field is missing and marked `+k8s:required`, the framework reports the
"required" error and does **not** run further validations on that field (such as
`+k8s:minimum` or `+k8s:format`). This is expected behavior — it avoids
cascading errors on nil/zero values.

**Q: Are there any stability constraints on the tags I can use for a new API?**

All tags can be used regardless of stability level on any API, but there is some
guidance. If a tag is explicitly marked as "Alpha" or "Beta" in the
[Tag Catalog](https://kubernetes.io/docs/reference/using-api/declarative-validation/),
it generally should not be used as the *sole* authoritative validation for a
Stable/GA Kubernetes API (can definitely be used on Alpha/Beta APIs). Standard
(GA) tags can be used authoritatively everywhere.

**Q: What are `+k8s:alpha` and `+k8s:beta` lifecycle prefixes?**

These are part of the validation lifecycle mechanism for graduating validation
rules on **existing** APIs during migration:

- `+k8s:alpha(since:v1.XX)=<tag>`: Runs in shadow mode only (metrics, no
  rejection).
- `+k8s:beta(since:v1.XX)=<tag>`: Enforced by default, disable-able via the
  `DeclarativeValidationTakeover` feature gate.

**For brand-new APIs using `rest.WithDeclarativeEnforcement()`, you typically do
not need these prefixes.** Standard tags (without a lifecycle prefix) are
permanently enforced when the enforcement flag is set. The lifecycle prefixes
exist for migrating validation on existing APIs where existing clients may be
sending data that would fail the new rule.

---

## Summary Checklist for API Reviewers (New APIs)

1. [ ] Are the `+k8s:` tags chosen appropriately from the official catalog?
2. [ ] Are the `zz_generated.validations.go` file(s) updated and included in the PR for each tagged API version?
3. [ ] Are the tags applied consistently across all relevant API versions (v1, v1beta1, etc.)?
4. [ ] Is `rest.WithDeclarativeEnforcement()` present in `strategy.go`?
5. [ ] Is redundant handwritten validation omitted in favor of the standard tags?
6. [ ] Are test expectations correctly identifying the DV errors using `.MarkNonShadowed()`?
7. [ ] Is cross-field logic appropriately left in handwritten code?