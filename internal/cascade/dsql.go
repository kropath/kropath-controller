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

// DSQLKropathSection holds the DSQL-family governance fields from
// KropathConfig.spec.mandatory.dsql / .defaults.dsql (ADR-015 §3.x).
//
// Only deletionProtectionEnabled is org-wide for DSQL; kmsEncryptionKey is
// per-profile/per-cluster and lives in DSQLConfigSection only (family design §8).
// Tags are tier-level cloud resource tags from KropathConfig.spec.mandatory.tags,
// populated by the reconciler so they flow through MergeDSQLCascade.
//
// Zero value of each field is the permissive sentinel (not enforced).
type DSQLKropathSection struct {
	// DeletionProtectionEnabled enforces deletion protection org-wide when true.
	// false (zero value) = not enforced.
	DeletionProtectionEnabled bool `json:"deletionProtectionEnabled,omitempty"`

	// Tags are tier-level cloud resource tags from KropathConfig.spec.mandatory.tags.
	// Populated by the reconciler from the tier-level field, not from spec.mandatory.dsql.
	// nil / empty map (zero value) = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// DSQLConfigSection holds the DSQL governance fields from DSQLConfig.spec.mandatory
// or DSQLConfig.spec.defaults (per-type ResourceConfig).
//
// Zero value of each field is the permissive sentinel (not enforced).
type DSQLConfigSection struct {
	// DeletionProtectionEnabled enforces deletion protection when true. false = not enforced.
	DeletionProtectionEnabled bool `json:"deletionProtectionEnabled,omitempty"`

	// KmsEncryptionKey is the ARN of the customer-managed KMS key to enforce.
	// Empty string = not enforced; AWS-owned key is used unless the instance or
	// defaults tier specifies one.
	KmsEncryptionKey string `json:"kmsEncryptionKey,omitempty"`

	// Tags are cloud resource tags for this DSQL config profile.
	// nil / empty map (zero value) = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are Kubernetes labels synced to cloud resource tags.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations synced to cloud resource tags.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveDSQLSection is one tier (mandatory or defaults) of the merged DSQL governance
// result written into DSQLConfig.status.effectiveConfig by the controller.
type EffectiveDSQLSection struct {
	DeletionProtectionEnabled bool              `json:"deletionProtectionEnabled,omitempty"`
	KmsEncryptionKey           string            `json:"kmsEncryptionKey,omitempty"`
	Tags                       map[string]string `json:"tags,omitempty"`
}

// EffectiveDSQLConfig is the merged DSQL governance result written into
// DSQLConfig.status.effectiveConfig by the controller.
type EffectiveDSQLConfig struct {
	Mandatory EffectiveDSQLSection `json:"mandatory"`
	Defaults  EffectiveDSQLSection `json:"defaults"`
}

// MergeDSQLCascade merges DSQL governance fields from all cascade sources and
// returns the effective configuration to be written to status.effectiveConfig.
//
// The ten-level priority chain (ADR-015 §5.3) for DSQL fields:
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace)
//	Level 3 — globalDSQLCfgMandatory  (DSQLConfig in kro-system)
//	Level 4 — localDSQLCfgMandatory   (DSQLConfig in resource namespace)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localDSQLCfgDefaults    (DSQLConfig in resource namespace)
//	Level 7 — globalDSQLCfgDefaults   (DSQLConfig in kro-system)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system)
//
// For mandatory (levels 1–4): first non-zero value in priority order wins.
// For defaults (levels 6–9): first non-zero value in priority order wins.
// For tags: union merge across all sources; lower level numbers win on key conflicts.
//
// deletionProtectionEnabled appears at all four mandatory/defaults levels via both
// KropathSection (levels 1–2, 8–9) and DSQLConfigSection (levels 3–4, 6–7).
// kmsEncryptionKey is NOT in KropathConfig (per-profile/per-cluster per family design §8),
// so it only appears at levels 3–4 (mandatory) and 6–7 (defaults).
// Tags appear at all four mandatory/defaults levels; KropathSection.Tags carries the
// tier-level KropathConfig.mandatory.tags (populated by the reconciler).
func MergeDSQLCascade(
	// Mandatory inputs (highest → lowest priority)
	globalKropathMandatory DSQLKropathSection, // level 1
	localKropathMandatory DSQLKropathSection, // level 2
	globalDSQLCfgMandatory DSQLConfigSection, // level 3
	localDSQLCfgMandatory DSQLConfigSection, // level 4
	// Defaults inputs (highest → lowest priority)
	localDSQLCfgDefaults DSQLConfigSection, // level 6
	globalDSQLCfgDefaults DSQLConfigSection, // level 7
	localKropathDefaults DSQLKropathSection, // level 8
	globalKropathDefaults DSQLKropathSection, // level 9
) EffectiveDSQLConfig {
	return EffectiveDSQLConfig{
		Mandatory: EffectiveDSQLSection{
			DeletionProtectionEnabled: firstTrue(
				globalKropathMandatory.DeletionProtectionEnabled, // level 1
				localKropathMandatory.DeletionProtectionEnabled,  // level 2
				globalDSQLCfgMandatory.DeletionProtectionEnabled, // level 3
				localDSQLCfgMandatory.DeletionProtectionEnabled,  // level 4
			),
			// kmsEncryptionKey not in KropathConfig: levels 3 and 4 only.
			KmsEncryptionKey: firstNonEmptyString(
				globalDSQLCfgMandatory.KmsEncryptionKey, // level 3
				localDSQLCfgMandatory.KmsEncryptionKey,  // level 4
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflicts.
			Tags: mergeMaps(
				localDSQLCfgMandatory.Tags,   // level 4 (lowest priority, set first)
				globalDSQLCfgMandatory.Tags,  // level 3
				localKropathMandatory.Tags,   // level 2
				globalKropathMandatory.Tags,  // level 1 (highest priority, last to write)
			),
		},
		Defaults: EffectiveDSQLSection{
			DeletionProtectionEnabled: firstTrue(
				localDSQLCfgDefaults.DeletionProtectionEnabled,  // level 6
				globalDSQLCfgDefaults.DeletionProtectionEnabled, // level 7
				localKropathDefaults.DeletionProtectionEnabled,   // level 8
				globalKropathDefaults.DeletionProtectionEnabled,  // level 9
			),
			// kmsEncryptionKey not in KropathConfig: levels 6 and 7 only.
			KmsEncryptionKey: firstNonEmptyString(
				localDSQLCfgDefaults.KmsEncryptionKey,  // level 6
				globalDSQLCfgDefaults.KmsEncryptionKey, // level 7
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflicts.
			Tags: mergeMaps(
				globalKropathDefaults.Tags, // level 9 (lowest priority)
				localKropathDefaults.Tags,  // level 8
				globalDSQLCfgDefaults.Tags, // level 7
				localDSQLCfgDefaults.Tags,  // level 6 (highest priority)
			),
		},
	}
}
