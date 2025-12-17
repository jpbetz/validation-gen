# Declarative Validation (`validation-gen`)

This document provides an overview of the Declarative Validation project in Kubernetes, also known as `validation-gen`. This feature allows developers to define validation logic for native Kubernetes API types using Go comment tags (e.g., `+k8s:minimum=0`).

## Architecture Overview

The declarative validation system consists of two main components:

1.  **Code Generator (`validation-gen`)**: Parses special `+k8s:` comment tags in API type definitions (`types.go`) and generates Go code (`zz_generated.validation.go`) that enforces these rules.
2.  **Runtime Validation Library**: A set of validation functions that the generated code calls to perform the actual validation (e.g., checking minimums, formats, required fields).

## Key Directories

*   **`staging/src/k8s.io/code-generator/cmd/validation-gen/`**: The main package for the code generator.
    *   **`validators/`**: Contains the definitions for the validation tags themselves (e.g., how `+k8s:required` is parsed and what code it generates).
    *   **`output_tests/tags/`**: Contains tests that verify the generated code for each validation tag.
*   **`staging/src/k8s.io/apimachinery/pkg/api/validate/`**: The runtime library containing the actual validation logic called by the generated code.

## Developer Workflows

### 1. Creating a New Simple Declarative Validation Tag

To add a new validation tag (e.g., `+k8s:newRule`), follow this workflow:

1.  **Define the Tag**: Add or modify files in `staging/src/k8s.io/code-generator/cmd/validation-gen/validators/` to define how the new tag is parsed from comments and what Go code it should generate.
2.  **Implement Runtime Logic**: Add the corresponding validation function to `staging/src/k8s.io/apimachinery/pkg/api/validate/`. This is the function that the generated code will call.
3.  **Add Generation Tests**: Create or update test cases in `staging/src/k8s.io/code-generator/cmd/validation-gen/output_tests/tags/` to ensure the code generator produces the correct code for your new tag.
4.  **Run Code Generation**: Regenerate the validation files to apply your changes.
    ```bash
    hack/update-codegen.sh validation
    ```
5.  **Run Tests**: Verify your changes and ensure no regressions.
    ```bash
    go test ./staging/src/k8s.io/code-generator/cmd/validation-gen/...
    ```

### 2. Migrating Handwritten Validation to Declarative Validation

> **Detailed Guide**: For a comprehensive, step-by-step walkthrough, common pitfalls (like short-circuiting), and advanced patterns, please read the **[Migration Guide](MIGRATION_GUIDE.md)**.

Below is a short summary for how to migrate existing imperative validation to the new declarative system, you should consult the comprehensive guide above for more specific and detailed guidance:

1.  **Enable Code Generation**: In the `doc.go` of the API version (e.g., `pkg/apis/<group>/<version>/doc.go`), add the necessary tags:
    ```go
    // +k8s:validation-gen=TypeMeta
    // +k8s:validation-gen-input=k8s.io/api/<group>/<version>
    ```
    For subresources, add `// +k8s:supportsSubresource=/<subresource>` to the type definition.

2.  **Add Validation Tags**: Add the appropriate `+k8s:` tags to the fields in the versioned API's `types.go` file (e.g., located under `./staging/src/k8s.io/api/<group>/<version>/types.go`). Example: `+k8s:minimum=0` on a field. NOTE: the validation tags should be added to the VERSIONED types at under `./staging/src/k8s.io/api/<group>/<version>/types.go`, NOT the internal types.

3.  **Update API Strategy**: In the resource's `strategy.go` file, modify `Validate` and `ValidateUpdate` to call the declarative validation function `rest.ValidateDeclarativelyWithMigrationChecks` and merge their errors with existing ones.

4.  **Add Tests**: Create a `declarative_validation_test.go` in the same directory as `strategy.go` to test the new declarative path. Use the test wrapper `apitesting.VerifyValidationEquivalence` to ensure there is no mismatch between declarative and imperative validation.

5.  **Correlate Errors**: In the old handwritten validation code, mark errors that are now covered by declarative validation using `.MarkCoveredByDeclarative()`. Ensure all errors have a `.WithOrigin()` set.

6.  **Enable Fuzz Testing**: Add the API group/version to `pkg/api/testing/validation_test.go`.

7.  **Run Code Generation & Tests**: Run `hack/update-codegen.sh validation` and the package-specific tests (e.g., `make test WHAT=./pkg/registry/...`).

## Useful Commands

*   **Regenerate Validation Code**:
    ```bash
    hack/update-codegen.sh validation
    ```
*   **Run `validation-gen` Tests**:
    ```bash
    go test ./staging/src/k8s.io/code-generator/cmd/validation-gen/...
    ```
*   **Run `validate/*` Logic Tests**:
    ```bash
    go test ./staging/src/k8s.io/apimachinery/pkg/api/validate/...
    ```
*   **Format Code**:
    ```bash
    hack/update-gofmt.sh
    ```
*   **Run All Verification Checks**:
    ```bash
    hack/verify-all.sh
    ```

## Project Conventions

*   Follow the standard Kubernetes coding style and conventions.
*   Ensure all new files have the correct Apache 2.0 copyright header.
*   All new features and bug fixes must have comprehensive tests.
