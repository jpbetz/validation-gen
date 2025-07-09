# Kubernetes Validation Tags Documentation

This document lists the supported validation tags and their related information.

## Tags Overview

| Tag | Usage | Description | Scopes |
|-----|-------------|-------------|----------|
| [`k8s:eachKey`](#k8seachkey) | k8s:eachKey=\<validation-tag\> | Declares a validation for each value in a map or list. | anywhere |
| [`k8s:eachVal`](#k8seachval) | k8s:eachVal=\<validation-tag\> | Declares a validation for each value in a map or list. | anywhere |
| [`k8s:enum`](#k8senum) | k8s:enum | Indicates that a string type is an enum. All const values of this type are considered values in the enum. | type definitions |
| [`k8s:forbidden`](#k8sforbidden) | k8s:forbidden | Indicates that a field may not be specified. | struct fields |
| [`k8s:format`](#k8sformat) | k8s:format=\<payload\> | Indicates that a string field has a particular format. | anywhere |
| [`k8s:ifDisabled`](#k8sifdisabled) | k8s:ifDisabled(\<option\>)=\<validation-tag\> | Declares a validation that only applies when an option is disabled. | anywhere |
| [`k8s:ifEnabled`](#k8sifenabled) | k8s:ifEnabled(\<option\>)=\<validation-tag\> | Declares a validation that only applies when an option is enabled. | anywhere |
| [`k8s:immutable`](#k8simmutable) | k8s:immutable | Indicates that a field may not be updated. | list values, map values, struct fields, type definitions |
| [`k8s:item`](#k8sitem) | +k8s:item(stringKey: "value", intKey: 42, boolKey: true)=\<validation-tag\> | Declares a validation for an item of a slice declared as a +k8s:listType=map. The item to match is declared by providing field-value pair arguments. All key fields must be specified. | anywhere |
| [`k8s:listMapKey`](#k8slistmapkey) | k8s:listMapKey=\<field-json-name\> | Declares a named sub-field of a list's value-type to be part of the list-map key. | anywhere |
| [`k8s:listType`](#k8slisttype) | k8s:listType=\<type\> | Declares a list field's semantic type. | anywhere |
| [`k8s:maxItems`](#k8smaxitems) | k8s:maxItems=\<non-negative integer\> | Indicates that a list field has a limit on its size. | list values, map values, struct fields, type definitions |
| [`k8s:maxLength`](#k8smaxlength) | k8s:maxLength=\<non-negative integer\> | Indicates that a string field has a limit on its length. | anywhere |
| [`k8s:minimum`](#k8sminimum) | k8s:minimum=\<integer\> | Indicates that a numeric field has a minimum value. | anywhere |
| [`k8s:neq`](#k8sneq) | k8s:neq=\<value\> | Verifies the field's value is not equal to a specific disallowed value. Supports string, integer, and boolean types. | anywhere |
| [`k8s:opaqueType`](#k8sopaquetype) | k8s:opaqueType | Indicates that any validations declared on the referenced type will be ignored. If a referenced type's package is not included in the generator's current flags, this tag must be set, or code generation will fail (preventing silent mistakes). If the validations should not be ignored, add the type's package to the generator using the --readonly-pkg flag. | struct fields |
| [`k8s:optional`](#k8soptional) | k8s:optional | Indicates that a field is optional to clients. | struct fields |
| [`k8s:required`](#k8srequired) | k8s:required | Indicates that a field must be specified by clients. | struct fields |
| [`k8s:subfield`](#k8ssubfield) | k8s:subfield(\<field-json-name\>)=\<validation-tag\> | Declares a validation for a subfield of a struct. | anywhere |
| [`k8s:unionDiscriminator`](#k8suniondiscriminator) | k8s:unionDiscriminator(\<string\>) | Indicates that this field is the discriminator for a union. | list values, struct fields |
| [`k8s:unionMember`](#k8sunionmember) | k8s:unionMember(\<string\>, \<string\>) | Indicates that this field is a member of a union. | list values, struct fields |
| [`k8s:validateError`](#k8svalidateerror) | k8s:validateError=\<string\> | Always fails code generation (useful for testing). | anywhere |
| [`k8s:validateFalse`](#k8svalidatefalse) | k8s:validateFalse(\<comma-separated-list-of-flag-string\>, \<string\>, \<string\>)=\<payload\> | Always fails validation (useful for testing). | anywhere |
| [`k8s:validateTrue`](#k8svalidatetrue) | k8s:validateTrue(\<comma-separated-list-of-flag-string\>, \<string\>, \<string\>)=\<payload\> | Always passes validation (useful for testing). | anywhere |
| [`k8s:zeroOrOneOfMember`](#k8szerooroneofmember) | k8s:zeroOrOneOfMember(\<string\>, \<string\>) | Indicates that this field is a member of a zero-or-one-of union. | list values, struct fields |

## Tag Details

### k8s:eachKey

#### Args

No args

#### Payloads

**Type:** tag | **Required:** true

| Description | Docs |
|-------------|------|
| \<validation-tag\> | The tag to evaluate for each value. |

### k8s:eachVal

#### Args

No args

#### Payloads

**Type:** tag | **Required:** true

| Description | Docs |
|-------------|------|
| \<validation-tag\> | The tag to evaluate for each value. |

### k8s:enum

#### Args

No args

#### Payloads

No payloads

### k8s:forbidden

#### Args

No args

#### Payloads

No payloads

### k8s:format

#### Args

No args

#### Payloads

**Type:** string | **Required:** true

| Description | Docs |
|-------------|------|
| k8s-ip | This field holds an IPv4 or IPv6 address value. IPv4 octets may have leading zeros. |
| k8s-long-name | This field holds a Kubernetes "long name", aka a "DNS subdomain" value. |
| k8s-short-name | This field holds a Kubernetes "short name", aka a "DNS label" value. |

### k8s:ifDisabled

#### Args

| Name | Description | Type | Required | Default | Docs |
|------|-------------|------|----------|---------|------|
| N/A | \<option\> | string | Yes | N/A | N/A |

#### Payloads

**Type:** tag | **Required:** true

| Description | Docs |
|-------------|------|
| \<validation-tag\> | This validation tag will be evaluated only if the validation option is disabled. |

### k8s:ifEnabled

#### Args

| Name | Description | Type | Required | Default | Docs |
|------|-------------|------|----------|---------|------|
| N/A | \<option\> | string | Yes | N/A | N/A |

#### Payloads

**Type:** tag | **Required:** true

| Description | Docs |
|-------------|------|
| \<validation-tag\> | This validation tag will be evaluated only if the validation option is enabled. |

### k8s:immutable

#### Args

No args

#### Payloads

No payloads

### k8s:item

Arguments must be named with the JSON names of the list-map key fields. Values can be strings, integers, or booleans. For example: +k8s:item(name: "myname", priority: 10, enabled: true)=<chained-validation-tag>

#### Args

No args

#### Payloads

**Type:** tag | **Required:** true

| Description | Docs |
|-------------|------|
| \<validation-tag\> | The tag to evaluate for the matching list item. |

### k8s:listMapKey

#### Args

No args

#### Payloads

**Type:** string | **Required:** true

| Description | Docs |
|-------------|------|
| \<field-json-name\> | The name of the field. |

### k8s:listType

#### Args

No args

#### Payloads

**Type:** string | **Required:** true

| Description | Docs |
|-------------|------|
| \<type\> | atomic | map | set |

### k8s:maxItems

#### Args

No args

#### Payloads

**Type:** int | **Required:** true

| Description | Docs |
|-------------|------|
| \<non-negative integer\> | This field must be no more than X items long. |

### k8s:maxLength

#### Args

No args

#### Payloads

**Type:** int | **Required:** true

| Description | Docs |
|-------------|------|
| \<non-negative integer\> | This field must be no more than X characters long. |

### k8s:minimum

#### Args

No args

#### Payloads

**Type:** int | **Required:** true

| Description | Docs |
|-------------|------|
| \<integer\> | This field must be greater than or equal to x. |

### k8s:neq

#### Args

No args

#### Payloads

**Type:** raw | **Required:** true

| Description | Docs |
|-------------|------|
| \<value\> | The disallowed value. The parser will infer the type (string, int, bool). |

### k8s:opaqueType

#### Args

No args

#### Payloads

No payloads

### k8s:optional

#### Args

No args

#### Payloads

No payloads

### k8s:required

#### Args

No args

#### Payloads

No payloads

### k8s:subfield

The named subfield must be a direct field of the struct, or of an embedded struct.

#### Args

| Name | Description | Type | Required | Default | Docs |
|------|-------------|------|----------|---------|------|
| N/A | \<field-json-name\> | string | Yes | N/A | N/A |

#### Payloads

**Type:** tag | **Required:** true

| Description | Docs |
|-------------|------|
| \<validation-tag\> | The tag to evaluate for the subfield. |

### k8s:unionDiscriminator

#### Args

| Name | Description | Type | Required | Default | Docs |
|------|-------------|------|----------|---------|------|
| union | \<string\> | string | No | N/A | the name of the union, if more than one exists |

#### Payloads

No payloads

### k8s:unionMember

#### Args

| Name | Description | Type | Required | Default | Docs |
|------|-------------|------|----------|---------|------|
| union | \<string\> | string | No | N/A | the name of the union, if more than one exists |
| memberName | \<string\> | string | No | the field's name | the discriminator value for this member |

#### Payloads

No payloads

### k8s:validateError

#### Args

No args

#### Payloads

**Type:** string | **Required:** false

| Description | Docs |
|-------------|------|
| \<string\> | This string will be included in the error message. |

### k8s:validateFalse

#### Args

| Name | Description | Type | Required | Default | Docs |
|------|-------------|------|----------|---------|------|
| flags | \<comma-separated-list-of-flag-string\> | string | No | N/A | values: ShortCircuit, NonError |
| typeArg | \<string\> | string | No | N/A | The type arg in generated code (must be the value-type, not pointer). |
| cohort | \<string\> | string | No | N/A | An optional cohort name to group multiple validations. |

#### Payloads

**Type:** string | **Required:** false

| Description | Docs |
|-------------|------|
| \<none\> | N/A |
| \<string\> | The generated code will include this string. |

### k8s:validateTrue

#### Args

| Name | Description | Type | Required | Default | Docs |
|------|-------------|------|----------|---------|------|
| flags | \<comma-separated-list-of-flag-string\> | string | No | N/A | values: ShortCircuit, NonError |
| typeArg | \<string\> | string | No | N/A | The type arg in generated code (must be the value-type, not pointer). |
| cohort | \<string\> | string | No | N/A | An optional cohort name to group multiple validations. |

#### Payloads

**Type:** string | **Required:** false

| Description | Docs |
|-------------|------|
| \<none\> | N/A |
| \<string\> | The generated code will include this string. |

### k8s:zeroOrOneOfMember

A zero-or-one-of union allows at most one member to be set. Unlike regular unions, having no members set is valid.

#### Args

| Name | Description | Type | Required | Default | Docs |
|------|-------------|------|----------|---------|------|
| union | \<string\> | string | No | N/A | the name of the union, if more than one exists |
| memberName | \<string\> | string | No | the field's name | the custom member name for this member |

#### Payloads

No payloads

