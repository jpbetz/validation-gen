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

package validators

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/gengo/v2/types"
)

const (
	subfieldTagName = "k8s:subfield"
)

func init() {
	RegisterTagValidator(&subfieldTagValidator{})
}

type subfieldTagValidator struct {
	validator Validator
}

func (stv *subfieldTagValidator) Init(cfg Config) {
	stv.validator = cfg.Validator
}

func (subfieldTagValidator) TagName() string {
	return subfieldTagName
}

var subfieldTagValidScopes = sets.New(ScopeAny)

func (subfieldTagValidator) ValidScopes() sets.Set[Scope] {
	return subfieldTagValidScopes
}

var (
	validateSubfield                = types.Name{Package: libValidationPkg, Name: "Subfield"}
	validateListMapElementByKeyName = types.Name{Package: libValidationPkg, Name: "ListMapElementByKey"}
)

// parseSubfieldConfig parses the subfield configuration string and the chained validation tag.
// It supports two main styles:
// 1. Argument style: +k8s:subfield(CONFIG_ARG)=+CHAINED_TAG
//   - args: ["CONFIG_ARG"]
//   - rawTagValueOrPayload: "+CHAINED_TAG"
//
// 2. equals-separated payload style: +k8s:subfield=CONFIG_PAYLOAD_PREFIX=+CHAINED_TAG
//   - args: []
//   - rawTagValueOrPayload: "CONFIG_PAYLOAD_PREFIX=+CHAINED_TAG"
//
// It returns the config string, the validation tag to apply, whether it's list access, and any error.
func parseSubfieldConfig(args []string, rawTagValueOrPayload string) (string, string, bool, error) {
	var subfieldConfigStr string
	var validationTagToApply string
	var isListAccessByJSON bool

	if len(args) == 1 { // Argument style: +k8s:subfield(CONFIG_ARG)=+CHAINED_TAG
		subfieldConfigStr = args[0]
		validationTagToApply = rawTagValueOrPayload // This is the +CHAINED_TAG part

		if validationTagToApply == "" {
			return "", "", false, fmt.Errorf("%s: payload (chained tag) is required for argument style, config was '(%s)'", subfieldTagName, subfieldConfigStr)
		}
		if strings.HasPrefix(subfieldConfigStr, "{") && strings.HasSuffix(subfieldConfigStr, "}") {
			// Preliminary check if config looks like JSON for list access.
			// Further validation (e.g., actual JSON parsing) happens later.
			isListAccessByJSON = true
		}
		// If not JSON-like, it's assumed to be a direct field name.
	} else if len(args) == 0 { // Colon-separated payload style: +k8s:subfield:CONFIG_PAYLOAD_PREFIX=+CHAINED_TAG
		if rawTagValueOrPayload == "" {
			return "", "", false, fmt.Errorf("%s: payload is required for colon-separated style (e.g., :config=payload)", subfieldTagName)
		}
		payloadString := rawTagValueOrPayload

		// Try to parse as JSON_CONFIG=CHAINED_TAG
		// This attempts to find a valid JSON object at the beginning of the payloadString,
		// followed by an '=' and the chained tag.
		if strings.HasPrefix(payloadString, "{") {
			var p struct{} // Dummy struct for checking JSON syntax
			dec := json.NewDecoder(bytes.NewReader([]byte(payloadString)))
			if err := dec.Decode(&p); err == nil {
				// Successfully decoded a JSON object from the prefix.
				offset := dec.InputOffset() // Position after the decoded JSON object
				potentialJsonConfig := strings.TrimSpace(payloadString[:offset])

				// Ensure the decoded part is indeed the intended config and properly terminated for chaining.
				if strings.HasPrefix(potentialJsonConfig, "{") && strings.HasSuffix(potentialJsonConfig, "}") {
					if offset < int64(len(payloadString)) && payloadString[offset] == '=' {
						subfieldConfigStr = potentialJsonConfig
						validationTagToApply = strings.TrimSpace(payloadString[offset+1:])
						isListAccessByJSON = true
						// Successfully parsed as JSON_CONFIG=CHAINED_TAG, proceed to common checks.
						goto common_checks
					} else if offset == int64(len(payloadString)) {
						return "", "", false, fmt.Errorf("%s: missing chained tag after JSON config '%s' in payload '%s'", subfieldTagName, potentialJsonConfig, payloadString)
					} else {
						return "", "", false, fmt.Errorf("%s: expected '=' after JSON config '%s' for chaining in payload '%s', got '%s'", subfieldTagName, potentialJsonConfig, payloadString, payloadString[offset:])
					}
				}
			}
			// If JSON decoding failed or the structure wasn't right,
			// fall through to treat as simple FIELD_NAME=CHAINED_TAG.
			// A SyntaxError on a very short input (e.g., "{a") might also fall through if not a clear JSON prefix.
		}

		// Fallback or direct path for FIELD_NAME=CHAINED_TAG
		firstEqualsIdx := strings.Index(payloadString, "=")
		if firstEqualsIdx == -1 {
			return "", "", false, fmt.Errorf("%s: missing '=' for chained tag in payload: '%s'", subfieldTagName, payloadString)
		}
		subfieldConfigStr = payloadString[:firstEqualsIdx]
		validationTagToApply = payloadString[firstEqualsIdx+1:]
		isListAccessByJSON = false // Assume direct field access unless subfieldConfigStr is later found to be JSON.

		// Re-check if subfieldConfigStr for FIELD_NAME=CHAINED_TAG is actually JSON.
		// This handles cases like +k8s:subfield:{"type":"foo"}=bar
		if strings.HasPrefix(subfieldConfigStr, "{") && strings.HasSuffix(subfieldConfigStr, "}") {
			var temp map[string]interface{}
			if json.Unmarshal([]byte(subfieldConfigStr), &temp) == nil {
				isListAccessByJSON = true
			}
		}

	} else {
		return "", "", false, fmt.Errorf("%s: invalid arguments. Use '+k8s:subfield(config)=payload' OR '+k8s:subfield:config=payload'. Got %d args: %v", subfieldTagName, len(args), args)
	}

common_checks:
	if subfieldConfigStr == "" {
		return "", "", false, fmt.Errorf("%s: subfield configuration part could not be determined from raw value '%s'", subfieldTagName, rawTagValueOrPayload)
	}
	subfieldConfigStr = strings.TrimSpace(subfieldConfigStr)
	if subfieldConfigStr == "" {
		return "", "", false, fmt.Errorf("%s: subfield configuration part became empty after trim from raw value '%s'", subfieldTagName, rawTagValueOrPayload)
	}

	validationTagToApply = strings.TrimSpace(validationTagToApply)
	if validationTagToApply == "" {
		return "", "", false, fmt.Errorf("%s: chained validation tag part is empty for config '%s'", subfieldTagName, subfieldConfigStr)
	}

	return subfieldConfigStr, validationTagToApply, isListAccessByJSON, nil
}

func (stv *subfieldTagValidator) GetValidations(context Context, args []string, rawTagValueOrPayload string) (Validations, error) {
	var result Validations

	subfieldConfigStr, validationTagToApply, isListAccessByJSON, err := parseSubfieldConfig(args, rawTagValueOrPayload)
	if err != nil {
		return result, err
	}

	if isListAccessByJSON {
		var listSelector map[string]string
		if err := json.Unmarshal([]byte(subfieldConfigStr), &listSelector); err != nil {
			return result, fmt.Errorf("%s: error parsing JSON selector from config '%s': %w", subfieldTagName, subfieldConfigStr, err)
		}
		if len(listSelector) != 1 {
			return result, fmt.Errorf("%s: JSON selector in config '%s' must be a single key-value map", subfieldTagName, subfieldConfigStr)
		}

		// These are the correctly named variables from the loop
		var parsedKeyNameFromJSON, parsedKeyValueFromJSON string
		for k, v := range listSelector {
			parsedKeyNameFromJSON = k
			parsedKeyValueFromJSON = v
			break
		}

		currentFieldType := nonPointer(nativeType(context.Type))
		if currentFieldType.Kind != types.Slice && currentFieldType.Kind != types.Array {
			return result, fmt.Errorf("%s: list access (selector '%s') can only be used on slice/array types, but tag is on field %s of type %s", subfieldTagName, subfieldConfigStr, context.Path.String(), currentFieldType.Name.String())
		}
		elemType := nonPointer(nativeType(currentFieldType.Elem))
		if elemType.Kind != types.Struct {
			return result, fmt.Errorf("%s: elements of slice/array (selector '%s') must be structs, but elements of field %s are %s", subfieldTagName, subfieldConfigStr, context.Path.String(), elemType.Name.String())
		}
		// Use the parsed key name to find the member
		keyFieldMemb := getMemberByJSON(elemType, parsedKeyNameFromJSON)
		if keyFieldMemb == nil {
			return result, fmt.Errorf("%s: element type %s (of list %s) has no field with JSON name %q (from selector '%s')", subfieldTagName, elemType.Name.String(), context.Path.String(), parsedKeyNameFromJSON, subfieldConfigStr)
		}
		if context.Parent == nil || context.Member == nil {
			return result, fmt.Errorf("%s: list access (selector '%s') can only be used on a field of a struct (tag on %s)", subfieldTagName, subfieldConfigStr, context.Path.String())
		}

		subContextForPayload := Context{
			Scope:  ScopeField,
			Type:   elemType,
			Parent: context.Parent,
			// Use the correctly parsed variables here
			Path:   context.Path.Key(parsedKeyNameFromJSON + "=" + parsedKeyValueFromJSON),
			Member: keyFieldMemb,
		}

		payloadValidations, errExtract := stv.validator.ExtractValidations(subContextForPayload, []string{validationTagToApply})
		if errExtract != nil {
			return result, fmt.Errorf("failed to extract chained validations for %s list access (selector '%s', applying to element type %s) on %s: %w", subfieldTagName, subfieldConfigStr, elemType.Name.String(), context.Path.String(), errExtract)
		}
		result.Variables = append(result.Variables, payloadValidations.Variables...)

		for _, vfn := range payloadValidations.Functions {
			f := Function(
				subfieldTagName,
				vfn.Flags,
				validateListMapElementByKeyName,
				parsedKeyNameFromJSON,  // Pass the correct variable
				parsedKeyValueFromJSON, // Pass the correct variable
				WrapperFunction{vfn, elemType},
			)
			result.Functions = append(result.Functions, f)
		}
		return result, nil

	} else { // Direct struct field access
		// ... (this part was okay)
		subname := subfieldConfigStr
		t := nonPointer(nativeType(context.Type))
		if t.Kind != types.Struct {
			return result, fmt.Errorf("%s: direct field access ('%s') can only be used on struct types, got %s for field %s", subfieldTagName, subname, t.Kind, context.Path.String())
		}
		submemb := getMemberByJSON(t, subname)
		if submemb == nil {
			return result, fmt.Errorf("%s: no sub-field with JSON name %q in type %s for field %s", subfieldTagName, subname, t.Name.String(), context.Path.String())
		}

		subContextForPayload := Context{
			Scope:  ScopeField,
			Type:   submemb.Type,
			Parent: t,
			Path:   context.Path.Child(subname),
			Member: submemb,
		}
		payloadValidations, errExtract := stv.validator.ExtractValidations(subContextForPayload, []string{validationTagToApply})
		if errExtract != nil {
			return result, fmt.Errorf("failed to extract chained validations for %s field access on %s for sub-field %s: %w", subfieldTagName, context.Path.String(), subname, errExtract)
		}
		result.Variables = append(result.Variables, payloadValidations.Variables...)

		for _, vfn := range payloadValidations.Functions {
			accessorParentType := types.PointerTo(t)
			actualSubFieldType := submemb.Type
			returnedSubFieldType := actualSubFieldType
			fieldExprPrefix := ""
			if !isNilableType(actualSubFieldType) {
				returnedSubFieldType = types.PointerTo(actualSubFieldType)
				fieldExprPrefix = "&"
			}
			getFn := FunctionLiteral{
				Parameters: []ParamResult{{"o", accessorParentType}},
				Results:    []ParamResult{{"", returnedSubFieldType}},
			}
			getFn.Body = fmt.Sprintf("return %so.%s", fieldExprPrefix, submemb.Name)
			f := Function(
				subfieldTagName,
				vfn.Flags,
				validateSubfield,
				subname,
				getFn,
				WrapperFunction{vfn, submemb.Type},
			)
			result.Functions = append(result.Functions, f)
		}
		return result, nil
	}
}

func (stv subfieldTagValidator) Docs() TagDoc {
	doc := TagDoc{
		Tag:         stv.TagName(),
		Scopes:      stv.ValidScopes().UnsortedList(),
		Description: "Declares a validation for a subfield of a struct.",
		Args: []TagArgDoc{{
			Description: "<field-json-name>",
		}},
		Docs: "The named subfield must be a direct field of the struct, or of an embedded struct.",
		Payloads: []TagPayloadDoc{{
			Description: "<validation-tag>",
			Docs:        "The tag to evaluate for the subfield.",
		}},
	}
	return doc
}
