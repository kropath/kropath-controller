// Copyright 2026 kropath Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cascade

import "fmt"

// ValidateBedrockConfig runs two cross-validations on the merged Bedrock config
// (family design §7 controller validation contract):
//
//  1. allowedModels / foundationModel: when mandatory.foundationModel is non-empty
//     and mandatory.allowedModels is non-empty, foundationModel must be a member of
//     allowedModels. Returns InvalidModelNotInAllowedList on failure.
//
//  2. Guardrail pair consistency: at each tier (mandatory, defaults), guardrailIdentifier
//     and guardrailVersion must either both be non-empty or both be empty. A mismatch
//     in either tier returns InvalidGuardrailConfiguration.
//
// On failure the caller must NOT write status.effectiveConfig.
func ValidateBedrockConfig(eff EffectiveBedrockConfig) (valid bool, reason, message string) {
	// 1. allowedModels / foundationModel cross-validation (mandatory tier only)
	if eff.Mandatory.FoundationModel != "" && len(eff.Mandatory.AllowedModels) > 0 {
		found := false
		for _, m := range eff.Mandatory.AllowedModels {
			if m == eff.Mandatory.FoundationModel {
				found = true
				break
			}
		}
		if !found {
			return false, "InvalidModelNotInAllowedList", fmt.Sprintf(
				"mandatory.foundationModel %q is not in mandatory.allowedModels %v",
				eff.Mandatory.FoundationModel, eff.Mandatory.AllowedModels,
			)
		}
	}

	// 2. Guardrail pair consistency — mandatory tier
	mandatoryHasID := eff.Mandatory.GuardrailIdentifier != ""
	mandatoryHasVer := eff.Mandatory.GuardrailVersion != ""
	if mandatoryHasID != mandatoryHasVer {
		return false, "InvalidGuardrailConfiguration", fmt.Sprintf(
			"mandatory.guardrailIdentifier and mandatory.guardrailVersion must both be set or both be empty; got identifier=%q version=%q",
			eff.Mandatory.GuardrailIdentifier, eff.Mandatory.GuardrailVersion,
		)
	}

	// 2. Guardrail pair consistency — defaults tier
	defaultsHasID := eff.Defaults.GuardrailIdentifier != ""
	defaultsHasVer := eff.Defaults.GuardrailVersion != ""
	if defaultsHasID != defaultsHasVer {
		return false, "InvalidGuardrailConfiguration", fmt.Sprintf(
			"defaults.guardrailIdentifier and defaults.guardrailVersion must both be set or both be empty; got identifier=%q version=%q",
			eff.Defaults.GuardrailIdentifier, eff.Defaults.GuardrailVersion,
		)
	}

	return true, "", ""
}
