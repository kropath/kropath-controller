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

// CognitoPasswordPolicySection holds the password policy sub-fields shared between
// KropathConfig.cognito.passwordPolicy and CognitoConfig.mandatory/defaults.passwordPolicy.
//
// Integer fields use 0 as the "not set" sentinel (first-non-zero-wins cascade).
// Boolean pointer fields use nil = not set; false = explicitly disabled.
type CognitoPasswordPolicySection struct {
	// MinimumLength is the minimum password length. 0 = not enforced; valid range: 6–99.
	MinimumLength int64 `json:"minimumLength,omitempty"`

	// RequireLowercase enforces at least one lowercase letter. nil = not set.
	RequireLowercase *bool `json:"requireLowercase,omitempty"`

	// RequireNumbers enforces at least one digit. nil = not set.
	RequireNumbers *bool `json:"requireNumbers,omitempty"`

	// RequireSymbols enforces at least one symbol. nil = not set.
	RequireSymbols *bool `json:"requireSymbols,omitempty"`

	// RequireUppercase enforces at least one uppercase letter. nil = not set.
	RequireUppercase *bool `json:"requireUppercase,omitempty"`

	// TemporaryPasswordValidityDays sets the temporary-password expiry in days.
	// 0 = not enforced; valid range: 1–365.
	TemporaryPasswordValidityDays int64 `json:"temporaryPasswordValidityDays,omitempty"`
}

// CognitoKropathSection holds the Cognito-family governance fields from
// KropathConfig.spec.mandatory.cognito / .defaults.cognito (ADR-015 §3.5).
//
// namingTemplate, syncedLabels, and syncedAnnotations are NOT present here;
// they live only in CognitoConfig (CognitoConfigSection).
type CognitoKropathSection struct {
	// MfaConfiguration enforces the MFA level. Empty = not enforced.
	// Valid values: "OFF", "OPTIONAL", "ON".
	MfaConfiguration string `json:"mfaConfiguration,omitempty"`

	// DeletionProtection enforces deletion protection state. Empty = not enforced.
	// Valid values: "ACTIVE", "INACTIVE".
	DeletionProtection string `json:"deletionProtection,omitempty"`

	// AdvancedSecurityMode enforces the advanced security mode. Empty = not enforced.
	// Valid values: "OFF", "AUDIT", "ENFORCED".
	AdvancedSecurityMode string `json:"advancedSecurityMode,omitempty"`

	// PasswordPolicy contains governance sub-fields for the Cognito password policy.
	PasswordPolicy CognitoPasswordPolicySection `json:"passwordPolicy,omitempty"`

	// Tags are tier-level cloud resource tags augmented from KropathConfig.spec.mandatory.tags
	// / .defaults.tags so that tag union merge flows through MergeCognitoCascade.
	// nil / empty map = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// CognitoConfigSection holds the Cognito governance fields from CognitoConfig.spec.mandatory
// or CognitoConfig.spec.defaults (per-type ResourceConfig, ADR-015 §3.5).
//
// Contains the same governance fields as CognitoKropathSection plus namingTemplate,
// syncedLabels, and syncedAnnotations (CognitoConfig-only fields, per family design §4.1).
type CognitoConfigSection struct {
	// MfaConfiguration enforces the MFA level. Empty = not enforced.
	MfaConfiguration string `json:"mfaConfiguration,omitempty"`

	// DeletionProtection enforces deletion protection state. Empty = not enforced.
	DeletionProtection string `json:"deletionProtection,omitempty"`

	// AdvancedSecurityMode enforces the advanced security mode. Empty = not enforced.
	AdvancedSecurityMode string `json:"advancedSecurityMode,omitempty"`

	// PasswordPolicy contains governance sub-fields for the Cognito password policy.
	PasswordPolicy CognitoPasswordPolicySection `json:"passwordPolicy,omitempty"`

	// NamingTemplate is the user pool naming template (e.g. "{namespace}-{name}").
	// Governed only at CognitoConfig levels 3-4 (mandatory) and 6-7 (defaults).
	// KropathConfig.cognito does NOT carry namingTemplate.
	// Empty string = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created user pool resources.
	// Additive map merge across CognitoConfig tiers only.
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to created user pool resources.
	// Additive map merge across CognitoConfig tiers only.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`

	// Tags are cloud resource tags for this Cognito config profile.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// EffectiveCognitoPasswordPolicySection is the merged password policy sub-result for one tier.
type EffectiveCognitoPasswordPolicySection struct {
	MinimumLength                 int64  `json:"minimumLength,omitempty"`
	RequireLowercase              *bool  `json:"requireLowercase,omitempty"`
	RequireNumbers                *bool  `json:"requireNumbers,omitempty"`
	RequireSymbols                *bool  `json:"requireSymbols,omitempty"`
	RequireUppercase              *bool  `json:"requireUppercase,omitempty"`
	TemporaryPasswordValidityDays int64  `json:"temporaryPasswordValidityDays,omitempty"`
}

// EffectiveCognitoSection is one tier (mandatory or defaults) of the merged Cognito governance
// result written into CognitoConfig.status.effectiveConfig by the controller.
type EffectiveCognitoSection struct {
	MfaConfiguration     string                                `json:"mfaConfiguration,omitempty"`
	DeletionProtection   string                                `json:"deletionProtection,omitempty"`
	AdvancedSecurityMode string                                `json:"advancedSecurityMode,omitempty"`
	PasswordPolicy       EffectiveCognitoPasswordPolicySection `json:"passwordPolicy,omitempty"`
	NamingTemplate       string                                `json:"namingTemplate,omitempty"`
	SyncedLabels         map[string]string                     `json:"syncedLabels,omitempty"`
	SyncedAnnotations    map[string]string                     `json:"syncedAnnotations,omitempty"`
	Tags                 map[string]string                     `json:"tags,omitempty"`
}

// EffectiveCognitoConfig is the merged Cognito governance result written into
// CognitoConfig.status.effectiveConfig by the controller.
type EffectiveCognitoConfig struct {
	Mandatory EffectiveCognitoSection `json:"mandatory"`
	Defaults  EffectiveCognitoSection `json:"defaults"`
}

// MergeCognitoCascade merges Cognito governance fields from all cascade sources and
// returns the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for Cognito (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.cognito)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.cognito)
//	Level 3 — globalCognitoCfgMandatory   (CognitoConfig in kro-system, mandatory)
//	Level 4 — localCognitoCfgMandatory    (CognitoConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localCognitoCfgDefaults     (CognitoConfig in resource namespace, defaults)
//	Level 7 — globalCognitoCfgDefaults    (CognitoConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.cognito)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.cognito)
//
// String merge: firstNonEmptyString in priority order (lowest level number wins).
// Integer merge: firstNonZeroInt64 in priority order.
// *bool merge: firstNonNilBoolPtr — nil = not set (falls through); false = explicitly disabled.
// Tags: additive union merge across all four mandatory levels, all four defaults levels.
// SyncedLabels/SyncedAnnotations: additive union from CognitoConfig levels only (no KropathConfig).
// NamingTemplate: governed only at CognitoConfig levels (3-4 mandatory, 6-7 defaults).
// Password policy sub-fields each resolve independently through the same cascade.
func MergeCognitoCascade(
	globalKropathMandatory CognitoKropathSection, // level 1
	localKropathMandatory CognitoKropathSection, // level 2
	globalCognitoCfgMandatory CognitoConfigSection, // level 3
	localCognitoCfgMandatory CognitoConfigSection, // level 4
	localCognitoCfgDefaults CognitoConfigSection, // level 6
	globalCognitoCfgDefaults CognitoConfigSection, // level 7
	localKropathDefaults CognitoKropathSection, // level 8
	globalKropathDefaults CognitoKropathSection, // level 9
) EffectiveCognitoConfig {
	return EffectiveCognitoConfig{
		Mandatory: EffectiveCognitoSection{
			MfaConfiguration: firstNonEmptyString(
				globalKropathMandatory.MfaConfiguration,
				localKropathMandatory.MfaConfiguration,
				globalCognitoCfgMandatory.MfaConfiguration,
				localCognitoCfgMandatory.MfaConfiguration,
			),
			DeletionProtection: firstNonEmptyString(
				globalKropathMandatory.DeletionProtection,
				localKropathMandatory.DeletionProtection,
				globalCognitoCfgMandatory.DeletionProtection,
				localCognitoCfgMandatory.DeletionProtection,
			),
			AdvancedSecurityMode: firstNonEmptyString(
				globalKropathMandatory.AdvancedSecurityMode,
				localKropathMandatory.AdvancedSecurityMode,
				globalCognitoCfgMandatory.AdvancedSecurityMode,
				localCognitoCfgMandatory.AdvancedSecurityMode,
			),
			PasswordPolicy: EffectiveCognitoPasswordPolicySection{
				MinimumLength: firstNonZeroInt64(
					globalKropathMandatory.PasswordPolicy.MinimumLength,
					localKropathMandatory.PasswordPolicy.MinimumLength,
					globalCognitoCfgMandatory.PasswordPolicy.MinimumLength,
					localCognitoCfgMandatory.PasswordPolicy.MinimumLength,
				),
				RequireLowercase: firstNonNilBoolPtr(
					globalKropathMandatory.PasswordPolicy.RequireLowercase,
					localKropathMandatory.PasswordPolicy.RequireLowercase,
					globalCognitoCfgMandatory.PasswordPolicy.RequireLowercase,
					localCognitoCfgMandatory.PasswordPolicy.RequireLowercase,
				),
				RequireNumbers: firstNonNilBoolPtr(
					globalKropathMandatory.PasswordPolicy.RequireNumbers,
					localKropathMandatory.PasswordPolicy.RequireNumbers,
					globalCognitoCfgMandatory.PasswordPolicy.RequireNumbers,
					localCognitoCfgMandatory.PasswordPolicy.RequireNumbers,
				),
				RequireSymbols: firstNonNilBoolPtr(
					globalKropathMandatory.PasswordPolicy.RequireSymbols,
					localKropathMandatory.PasswordPolicy.RequireSymbols,
					globalCognitoCfgMandatory.PasswordPolicy.RequireSymbols,
					localCognitoCfgMandatory.PasswordPolicy.RequireSymbols,
				),
				RequireUppercase: firstNonNilBoolPtr(
					globalKropathMandatory.PasswordPolicy.RequireUppercase,
					localKropathMandatory.PasswordPolicy.RequireUppercase,
					globalCognitoCfgMandatory.PasswordPolicy.RequireUppercase,
					localCognitoCfgMandatory.PasswordPolicy.RequireUppercase,
				),
				TemporaryPasswordValidityDays: firstNonZeroInt64(
					globalKropathMandatory.PasswordPolicy.TemporaryPasswordValidityDays,
					localKropathMandatory.PasswordPolicy.TemporaryPasswordValidityDays,
					globalCognitoCfgMandatory.PasswordPolicy.TemporaryPasswordValidityDays,
					localCognitoCfgMandatory.PasswordPolicy.TemporaryPasswordValidityDays,
				),
			},
			// NamingTemplate: CognitoConfig levels only (3, 4); KropathConfig.cognito has no namingTemplate.
			NamingTemplate: firstNonEmptyString(
				globalCognitoCfgMandatory.NamingTemplate,
				localCognitoCfgMandatory.NamingTemplate,
			),
			// SyncedLabels: additive union from CognitoConfig levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localCognitoCfgMandatory.SyncedLabels,
				globalCognitoCfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localCognitoCfgMandatory.SyncedAnnotations,
				globalCognitoCfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localCognitoCfgMandatory.Tags,
				globalCognitoCfgMandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
		},
		Defaults: EffectiveCognitoSection{
			MfaConfiguration: firstNonEmptyString(
				localCognitoCfgDefaults.MfaConfiguration,
				globalCognitoCfgDefaults.MfaConfiguration,
				localKropathDefaults.MfaConfiguration,
				globalKropathDefaults.MfaConfiguration,
			),
			DeletionProtection: firstNonEmptyString(
				localCognitoCfgDefaults.DeletionProtection,
				globalCognitoCfgDefaults.DeletionProtection,
				localKropathDefaults.DeletionProtection,
				globalKropathDefaults.DeletionProtection,
			),
			AdvancedSecurityMode: firstNonEmptyString(
				localCognitoCfgDefaults.AdvancedSecurityMode,
				globalCognitoCfgDefaults.AdvancedSecurityMode,
				localKropathDefaults.AdvancedSecurityMode,
				globalKropathDefaults.AdvancedSecurityMode,
			),
			PasswordPolicy: EffectiveCognitoPasswordPolicySection{
				MinimumLength: firstNonZeroInt64(
					localCognitoCfgDefaults.PasswordPolicy.MinimumLength,
					globalCognitoCfgDefaults.PasswordPolicy.MinimumLength,
					localKropathDefaults.PasswordPolicy.MinimumLength,
					globalKropathDefaults.PasswordPolicy.MinimumLength,
				),
				RequireLowercase: firstNonNilBoolPtr(
					localCognitoCfgDefaults.PasswordPolicy.RequireLowercase,
					globalCognitoCfgDefaults.PasswordPolicy.RequireLowercase,
					localKropathDefaults.PasswordPolicy.RequireLowercase,
					globalKropathDefaults.PasswordPolicy.RequireLowercase,
				),
				RequireNumbers: firstNonNilBoolPtr(
					localCognitoCfgDefaults.PasswordPolicy.RequireNumbers,
					globalCognitoCfgDefaults.PasswordPolicy.RequireNumbers,
					localKropathDefaults.PasswordPolicy.RequireNumbers,
					globalKropathDefaults.PasswordPolicy.RequireNumbers,
				),
				RequireSymbols: firstNonNilBoolPtr(
					localCognitoCfgDefaults.PasswordPolicy.RequireSymbols,
					globalCognitoCfgDefaults.PasswordPolicy.RequireSymbols,
					localKropathDefaults.PasswordPolicy.RequireSymbols,
					globalKropathDefaults.PasswordPolicy.RequireSymbols,
				),
				RequireUppercase: firstNonNilBoolPtr(
					localCognitoCfgDefaults.PasswordPolicy.RequireUppercase,
					globalCognitoCfgDefaults.PasswordPolicy.RequireUppercase,
					localKropathDefaults.PasswordPolicy.RequireUppercase,
					globalKropathDefaults.PasswordPolicy.RequireUppercase,
				),
				TemporaryPasswordValidityDays: firstNonZeroInt64(
					localCognitoCfgDefaults.PasswordPolicy.TemporaryPasswordValidityDays,
					globalCognitoCfgDefaults.PasswordPolicy.TemporaryPasswordValidityDays,
					localKropathDefaults.PasswordPolicy.TemporaryPasswordValidityDays,
					globalKropathDefaults.PasswordPolicy.TemporaryPasswordValidityDays,
				),
			},
			// NamingTemplate: CognitoConfig levels only (6, 7).
			NamingTemplate: firstNonEmptyString(
				localCognitoCfgDefaults.NamingTemplate,
				globalCognitoCfgDefaults.NamingTemplate,
			),
			// SyncedLabels: additive union from CognitoConfig levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalCognitoCfgDefaults.SyncedLabels,
				localCognitoCfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalCognitoCfgDefaults.SyncedAnnotations,
				localCognitoCfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalCognitoCfgDefaults.Tags,
				localCognitoCfgDefaults.Tags,
			),
		},
	}
}
