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

// KMSKropathSection holds the KMS-family governance fields from
// KropathConfig.spec.mandatory.kms / .defaults.kms (ADR-015 §3.5)
// PLUS the tier-level tags from KropathConfig.spec.mandatory.tags (populated
// by the reconciler so that tag cascade flows through MergeKMSCascade).
//
// Only the cross-type KMS fields live here: enableKeyRotation and allowedKeySpecs.
// keySpec and keyUsage are per-key choices and are not in KropathConfig (family design §8).
//
// Zero value of each field is the permissive sentinel (not enforced).
type KMSKropathSection struct {
	// EnableKeyRotation enforces key rotation org-wide when true.
	// false (zero value) = not enforced.
	EnableKeyRotation bool `json:"enableKeyRotation,omitempty"`

	// AllowedKeySpecs is the org-wide allowlist of permitted KMS key specifications.
	// nil / empty slice (zero value) = no restriction.
	AllowedKeySpecs []string `json:"allowedKeySpecs,omitempty"`

	// Tags are tier-level cloud resource tags from KropathConfig.spec.mandatory.tags.
	// Populated by the reconciler from the tier-level field, not from spec.mandatory.kms.
	// nil / empty map (zero value) = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// KMSConfigSection holds the KMS governance fields from KMSConfig.spec.mandatory
// or KMSConfig.spec.defaults (per-type ResourceConfig).
//
// Zero value of each field is the permissive sentinel (not enforced).
type KMSConfigSection struct {
	// EnableKeyRotation enforces key rotation when true. false = not enforced.
	EnableKeyRotation bool `json:"enableKeyRotation,omitempty"`

	// KeySpec is the enforced KMS key spec (e.g. "SYMMETRIC_DEFAULT", "RSA_4096").
	// Empty string = not enforced.
	KeySpec string `json:"keySpec,omitempty"`

	// KeyUsage is the enforced KMS key usage (e.g. "ENCRYPT_DECRYPT", "SIGN_VERIFY").
	// Empty string = not enforced.
	KeyUsage string `json:"keyUsage,omitempty"`

	// AllowedKeySpecs is the allowlist of permitted key specs for this profile.
	// nil / empty slice = no restriction.
	AllowedKeySpecs []string `json:"allowedKeySpecs,omitempty"`

	// Tags are cloud resource tags for this KMS config profile.
	// nil / empty map (zero value) = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// EffectiveKMSSection is one tier (mandatory or defaults) of the merged KMS governance
// result written into KMSConfig.status.effectiveConfig by the controller.
type EffectiveKMSSection struct {
	EnableKeyRotation bool              `json:"enableKeyRotation,omitempty"`
	KeySpec           string            `json:"keySpec,omitempty"`
	KeyUsage          string            `json:"keyUsage,omitempty"`
	AllowedKeySpecs   []string          `json:"allowedKeySpecs,omitempty"`
	Tags              map[string]string `json:"tags,omitempty"`
}

// EffectiveKMSConfig is the merged KMS governance result written into
// KMSConfig.status.effectiveConfig by the controller.
type EffectiveKMSConfig struct {
	Mandatory EffectiveKMSSection `json:"mandatory"`
	Defaults  EffectiveKMSSection `json:"defaults"`
}

// MergeKMSCascade merges KMS governance fields from all cascade sources and
// returns the effective configuration to be written to status.effectiveConfig.
//
// The ten-level priority chain (ADR-015 §5.3) for KMS fields:
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace)
//	Level 3 — globalKMSCfgMandatory   (KMSConfig in kro-system)
//	Level 4 — localKMSCfgMandatory    (KMSConfig in resource namespace)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localKMSCfgDefaults     (KMSConfig in resource namespace)
//	Level 7 — globalKMSCfgDefaults    (KMSConfig in kro-system)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system)
//
// For mandatory (levels 1–4): first non-zero value in priority order wins.
// For defaults (levels 6–9): first non-zero value in priority order wins.
// For tags: union merge across all sources; lower level numbers win on key conflicts.
//
// keySpec and keyUsage are not in KropathConfig (per-key choices per family design §8),
// so they only appear at levels 3–4 (mandatory) and 6–7 (defaults).
// enableKeyRotation and allowedKeySpecs appear at all four mandatory/defaults levels.
// Tags appear at all four mandatory/defaults levels; KropathSection.Tags carries the
// tier-level KropathConfig.mandatory.tags (populated by the reconciler).
func MergeKMSCascade(
	// Mandatory inputs (highest → lowest priority)
	globalKropathMandatory KMSKropathSection, // level 1
	localKropathMandatory KMSKropathSection, // level 2
	globalKMSCfgMandatory KMSConfigSection, // level 3
	localKMSCfgMandatory KMSConfigSection, // level 4
	// Defaults inputs (highest → lowest priority)
	localKMSCfgDefaults KMSConfigSection, // level 6
	globalKMSCfgDefaults KMSConfigSection, // level 7
	localKropathDefaults KMSKropathSection, // level 8
	globalKropathDefaults KMSKropathSection, // level 9
) EffectiveKMSConfig {
	return EffectiveKMSConfig{
		Mandatory: EffectiveKMSSection{
			EnableKeyRotation: firstTrue(
				globalKropathMandatory.EnableKeyRotation, // level 1
				localKropathMandatory.EnableKeyRotation,  // level 2
				globalKMSCfgMandatory.EnableKeyRotation,  // level 3
				localKMSCfgMandatory.EnableKeyRotation,   // level 4
			),
			// keySpec not in KropathConfig: levels 3 and 4 only.
			KeySpec: firstNonEmptyString(
				globalKMSCfgMandatory.KeySpec, // level 3
				localKMSCfgMandatory.KeySpec,  // level 4
			),
			// keyUsage not in KropathConfig: levels 3 and 4 only.
			KeyUsage: firstNonEmptyString(
				globalKMSCfgMandatory.KeyUsage, // level 3
				localKMSCfgMandatory.KeyUsage,  // level 4
			),
			AllowedKeySpecs: firstNonEmptyStrings(
				globalKropathMandatory.AllowedKeySpecs, // level 1
				localKropathMandatory.AllowedKeySpecs,  // level 2
				globalKMSCfgMandatory.AllowedKeySpecs,  // level 3
				localKMSCfgMandatory.AllowedKeySpecs,   // level 4
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflicts.
			Tags: mergeMaps(
				localKMSCfgMandatory.Tags,   // level 4 (lowest priority, set first)
				globalKMSCfgMandatory.Tags,  // level 3
				localKropathMandatory.Tags,  // level 2
				globalKropathMandatory.Tags, // level 1 (highest priority, last to write)
			),
		},
		Defaults: EffectiveKMSSection{
			EnableKeyRotation: firstTrue(
				localKMSCfgDefaults.EnableKeyRotation,  // level 6
				globalKMSCfgDefaults.EnableKeyRotation, // level 7
				localKropathDefaults.EnableKeyRotation,  // level 8
				globalKropathDefaults.EnableKeyRotation, // level 9
			),
			// keySpec not in KropathConfig: levels 6 and 7 only.
			KeySpec: firstNonEmptyString(
				localKMSCfgDefaults.KeySpec,  // level 6
				globalKMSCfgDefaults.KeySpec, // level 7
			),
			// keyUsage not in KropathConfig: levels 6 and 7 only.
			KeyUsage: firstNonEmptyString(
				localKMSCfgDefaults.KeyUsage,  // level 6
				globalKMSCfgDefaults.KeyUsage, // level 7
			),
			AllowedKeySpecs: firstNonEmptyStrings(
				localKMSCfgDefaults.AllowedKeySpecs,  // level 6
				globalKMSCfgDefaults.AllowedKeySpecs, // level 7
				localKropathDefaults.AllowedKeySpecs,  // level 8
				globalKropathDefaults.AllowedKeySpecs, // level 9
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflicts.
			Tags: mergeMaps(
				globalKropathDefaults.Tags, // level 9 (lowest priority)
				localKropathDefaults.Tags,  // level 8
				globalKMSCfgDefaults.Tags,  // level 7
				localKMSCfgDefaults.Tags,   // level 6 (highest priority)
			),
		},
	}
}

// ValidateKMSKeySpec validates the allowedKeySpecs / keySpec cross-constraint on the
// resolved mandatory tier of an KMSConfig (family design OD-3, §7).
//
// Constraint: when mandatory.keySpec is non-empty and mandatory.allowedKeySpecs is
// non-empty, keySpec must be a member of allowedKeySpecs.
//
// Returns (true, "", "") when the constraint does not apply or when it passes.
// Returns (false, "InvalidKeySpecNotInAllowedList", <message>) on failure.
// On failure the caller must NOT write status.effectiveConfig.
func ValidateKMSKeySpec(mandatory EffectiveKMSSection) (valid bool, reason, message string) {
	if mandatory.KeySpec == "" || len(mandatory.AllowedKeySpecs) == 0 {
		return true, "", ""
	}
	for _, allowed := range mandatory.AllowedKeySpecs {
		if allowed == mandatory.KeySpec {
			return true, "", ""
		}
	}
	return false, "InvalidKeySpecNotInAllowedList", fmt.Sprintf(
		"mandatory.keySpec %q is not in mandatory.allowedKeySpecs %v",
		mandatory.KeySpec, mandatory.AllowedKeySpecs,
	)
}

// firstNonEmptyStrings returns a defensive copy of the first non-nil, non-empty slice
// from candidates, preventing the caller from aliasing a cached CR slice.
func firstNonEmptyStrings(candidates ...[]string) []string {
	for _, s := range candidates {
		if len(s) > 0 {
			out := make([]string, len(s))
			copy(out, s)
			return out
		}
	}
	return nil
}

// mergeMaps returns a new map that is the union of all input maps.
// Later maps in the argument list override earlier maps for the same key,
// so pass sources in ascending priority order (lowest priority first).
// Returns nil when all inputs are empty.
func mergeMaps(maps ...map[string]string) map[string]string {
	var out map[string]string
	for _, m := range maps {
		for k, v := range m {
			if out == nil {
				out = make(map[string]string)
			}
			out[k] = v
		}
	}
	return out
}
