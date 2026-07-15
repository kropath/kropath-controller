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

// Package cascade implements the ten-level governance cascade for kropath
// ResourceConfig CRDs, following ADR-010 and ADR-015 §5.3.
package cascade

// IAMSection holds the three IAM-specific governance fields shared by
// KropathConfig.spec.mandatory.iam / .defaults.iam and
// IAMConfig.spec.mandatory / .spec.defaults.
//
// Zero value of each field is the permissive sentinel (not enforced).
type IAMSection struct {
	// PermissionsBoundaryArn is the ARN of the IAM permissions boundary policy.
	// Empty string = no boundary enforced.
	PermissionsBoundaryArn string `json:"permissionsBoundaryArn,omitempty"`

	// BlockIamUserAccessKeys prevents IAM user access key creation when true.
	// false (zero value) = not enforced.
	BlockIamUserAccessKeys bool `json:"blockIamUserAccessKeys,omitempty"`

	// MaxSessionDurationSeconds caps the maximum role session duration in seconds.
	// 0 (zero value) = no cap enforced.
	MaxSessionDurationSeconds int64 `json:"maxSessionDurationSeconds,omitempty"`
}

// EffectiveIAMConfig is the merged IAM governance result written into
// IAMConfig.status.effectiveConfig by the controller.
type EffectiveIAMConfig struct {
	Mandatory IAMSection `json:"mandatory"`
	Defaults  IAMSection `json:"defaults"`
}

// MergeIAMCascade merges IAM governance fields from all cascade sources and
// returns the effective configuration to be written to status.effectiveConfig.
//
// The ten-level priority chain (ADR-015 §5.3) for IAM fields:
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace)
//	Level 3 — globalIAMCfgMandatory   (IAMConfig in kro-system)
//	Level 4 — localIAMCfgMandatory    (IAMConfig in resource namespace)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localIAMCfgDefaults     (IAMConfig in resource namespace)
//	Level 7 — globalIAMCfgDefaults    (IAMConfig in kro-system)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system)
//
// For mandatory (levels 1–4): first non-zero value in priority order wins.
// For defaults (levels 6–9): first non-zero value in priority order wins.
// A source that is absent (nil or zero-value struct) is silently skipped.
func MergeIAMCascade(
	// Mandatory inputs (highest → lowest priority)
	globalKropathMandatory IAMSection, // level 1
	localKropathMandatory IAMSection, // level 2
	globalIAMCfgMandatory IAMSection, // level 3
	localIAMCfgMandatory IAMSection, // level 4
	// Defaults inputs (highest → lowest priority)
	localIAMCfgDefaults IAMSection, // level 6
	globalIAMCfgDefaults IAMSection, // level 7
	localKropathDefaults IAMSection, // level 8
	globalKropathDefaults IAMSection, // level 9
) EffectiveIAMConfig {
	return EffectiveIAMConfig{
		Mandatory: IAMSection{
			PermissionsBoundaryArn: firstNonEmptyString(
				globalKropathMandatory.PermissionsBoundaryArn,
				localKropathMandatory.PermissionsBoundaryArn,
				globalIAMCfgMandatory.PermissionsBoundaryArn,
				localIAMCfgMandatory.PermissionsBoundaryArn,
			),
			BlockIamUserAccessKeys: firstTrue(
				globalKropathMandatory.BlockIamUserAccessKeys,
				localKropathMandatory.BlockIamUserAccessKeys,
				globalIAMCfgMandatory.BlockIamUserAccessKeys,
				localIAMCfgMandatory.BlockIamUserAccessKeys,
			),
			MaxSessionDurationSeconds: firstNonZeroInt64(
				globalKropathMandatory.MaxSessionDurationSeconds,
				localKropathMandatory.MaxSessionDurationSeconds,
				globalIAMCfgMandatory.MaxSessionDurationSeconds,
				localIAMCfgMandatory.MaxSessionDurationSeconds,
			),
		},
		Defaults: IAMSection{
			PermissionsBoundaryArn: firstNonEmptyString(
				localIAMCfgDefaults.PermissionsBoundaryArn,
				globalIAMCfgDefaults.PermissionsBoundaryArn,
				localKropathDefaults.PermissionsBoundaryArn,
				globalKropathDefaults.PermissionsBoundaryArn,
			),
			BlockIamUserAccessKeys: firstTrue(
				localIAMCfgDefaults.BlockIamUserAccessKeys,
				globalIAMCfgDefaults.BlockIamUserAccessKeys,
				localKropathDefaults.BlockIamUserAccessKeys,
				globalKropathDefaults.BlockIamUserAccessKeys,
			),
			MaxSessionDurationSeconds: firstNonZeroInt64(
				localIAMCfgDefaults.MaxSessionDurationSeconds,
				globalIAMCfgDefaults.MaxSessionDurationSeconds,
				localKropathDefaults.MaxSessionDurationSeconds,
				globalKropathDefaults.MaxSessionDurationSeconds,
			),
		},
	}
}

// firstNonEmptyString returns the first non-empty string from the candidates.
func firstNonEmptyString(candidates ...string) string {
	for _, s := range candidates {
		if s != "" {
			return s
		}
	}
	return ""
}

// firstTrue returns true if any candidate is true; otherwise false.
// Boolean zero value is false, so the first true is a non-zero governance signal.
func firstTrue(candidates ...bool) bool {
	for _, b := range candidates {
		if b {
			return true
		}
	}
	return false
}

// firstNonZeroInt64 returns the first non-zero int64 from the candidates.
func firstNonZeroInt64(candidates ...int64) int64 {
	for _, v := range candidates {
		if v != 0 {
			return v
		}
	}
	return 0
}
