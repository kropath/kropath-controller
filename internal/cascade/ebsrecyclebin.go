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

// EBSRecycleBinKropathSection holds the EBS Recycle Bin governance fields from
// KropathConfig.spec.mandatory.ebsrecyclebin / .defaults.ebsrecyclebin (ADR-015 §3.5).
//
// Only 2 scalar fields are governed at the KropathConfig level: retentionPeriodValue
// and lockRule. All other fields (retentionPeriodUnit, unlockDelayValue,
// unlockDelayUnit, syncedLabels, syncedAnnotations) are RecycleBinConfig-only.
//
// Zero value of each field is the permissive sentinel (not enforced).
type EBSRecycleBinKropathSection struct {
	// RetentionPeriodValue is the number of retention period units.
	// 0 (zero value) = not enforced; first-non-zero-wins in cascade.
	RetentionPeriodValue int64 `json:"retentionPeriodValue,omitempty"`

	// LockRule enforces that the Recycle Bin rule is locked (cannot be modified or deleted
	// without going through a multi-step unlock process).
	// false (zero value) = not enforced; firstTrue wins in cascade.
	LockRule bool `json:"lockRule,omitempty"`

	// Tags are tier-level cloud resource tags.
	// nil / empty map = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// EBSRecycleBinConfigSection holds the EBS Recycle Bin governance fields from
// RecycleBinConfig.spec.mandatory or RecycleBinConfig.spec.defaults
// (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type EBSRecycleBinConfigSection struct {
	// RetentionPeriodValue is the number of retention period units. 0 = not enforced.
	RetentionPeriodValue int64 `json:"retentionPeriodValue,omitempty"`

	// RetentionPeriodUnit is the unit of the retention period (e.g. "DAYS").
	// Empty string = not enforced.
	RetentionPeriodUnit string `json:"retentionPeriodUnit,omitempty"`

	// LockRule enforces that the Recycle Bin rule is locked.
	// false (zero value) = not enforced.
	LockRule bool `json:"lockRule,omitempty"`

	// UnlockDelayValue is the number of unlock delay units.
	// 0 = not enforced.
	UnlockDelayValue int64 `json:"unlockDelayValue,omitempty"`

	// UnlockDelayUnit is the unit of the unlock delay period (e.g. "DAYS").
	// Empty string = not enforced.
	UnlockDelayUnit string `json:"unlockDelayUnit,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created Recycle Bin rule resources.
	// Additive map merge across RecycleBinConfig tiers only.
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to created Recycle Bin rule resources.
	// Additive map merge across RecycleBinConfig tiers only.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`

	// Tags are cloud resource tags for this Recycle Bin config profile.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// EffectiveEBSRecycleBinSection is one tier (mandatory or defaults) of the merged
// EBS Recycle Bin governance result written into RecycleBinConfig.status.effectiveConfig
// by the controller.
type EffectiveEBSRecycleBinSection struct {
	RetentionPeriodValue int64             `json:"retentionPeriodValue,omitempty"`
	RetentionPeriodUnit  string            `json:"retentionPeriodUnit,omitempty"`
	LockRule             bool              `json:"lockRule,omitempty"`
	UnlockDelayValue     int64             `json:"unlockDelayValue,omitempty"`
	UnlockDelayUnit      string            `json:"unlockDelayUnit,omitempty"`
	SyncedLabels         map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations    map[string]string `json:"syncedAnnotations,omitempty"`
	Tags                 map[string]string `json:"tags,omitempty"`
}

// EffectiveEBSRecycleBinConfig is the merged EBS Recycle Bin governance result written into
// RecycleBinConfig.status.effectiveConfig by the controller.
type EffectiveEBSRecycleBinConfig struct {
	Mandatory EffectiveEBSRecycleBinSection `json:"mandatory"`
	Defaults  EffectiveEBSRecycleBinSection `json:"defaults"`
}

// MergeEBSRecycleBinCascade merges EBS Recycle Bin governance fields from all cascade
// sources and returns the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for EBS Recycle Bin (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.ebsrecyclebin)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.ebsrecyclebin)
//	Level 3 — globalRBCfgMandatory    (RecycleBinConfig in kro-system, mandatory)
//	Level 4 — localRBCfgMandatory     (RecycleBinConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localRBCfgDefaults      (RecycleBinConfig in resource namespace, defaults)
//	Level 7 — globalRBCfgDefaults     (RecycleBinConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.ebsrecyclebin)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.ebsrecyclebin)
//
// Scalar merge: firstNonZeroInt64 / firstNonEmptyString in priority order (lowest number wins).
// LockRule: firstTrue (false = zero value / sentinel, true = enforced).
// Tags: additive union merge across all four mandatory levels, all four defaults levels.
// SyncedLabels/SyncedAnnotations: additive union from RecycleBinConfig levels only (no KropathConfig).
// RetentionPeriodUnit / UnlockDelayUnit / UnlockDelayValue: governed at RecycleBinConfig levels only.
func MergeEBSRecycleBinCascade(
	globalKropathMandatory EBSRecycleBinKropathSection, // level 1
	localKropathMandatory EBSRecycleBinKropathSection, // level 2
	globalRBCfgMandatory EBSRecycleBinConfigSection, // level 3
	localRBCfgMandatory EBSRecycleBinConfigSection, // level 4
	localRBCfgDefaults EBSRecycleBinConfigSection, // level 6
	globalRBCfgDefaults EBSRecycleBinConfigSection, // level 7
	localKropathDefaults EBSRecycleBinKropathSection, // level 8
	globalKropathDefaults EBSRecycleBinKropathSection, // level 9
) EffectiveEBSRecycleBinConfig {
	return EffectiveEBSRecycleBinConfig{
		Mandatory: EffectiveEBSRecycleBinSection{
			RetentionPeriodValue: firstNonZeroInt64(
				globalKropathMandatory.RetentionPeriodValue,
				localKropathMandatory.RetentionPeriodValue,
				globalRBCfgMandatory.RetentionPeriodValue,
				localRBCfgMandatory.RetentionPeriodValue,
			),
			// RetentionPeriodUnit: RecycleBinConfig levels only (3, 4).
			RetentionPeriodUnit: firstNonEmptyString(
				globalRBCfgMandatory.RetentionPeriodUnit,
				localRBCfgMandatory.RetentionPeriodUnit,
			),
			// LockRule: firstTrue — false is the zero-value sentinel meaning "not enforced".
			LockRule: firstTrue(
				globalKropathMandatory.LockRule,
				localKropathMandatory.LockRule,
				globalRBCfgMandatory.LockRule,
				localRBCfgMandatory.LockRule,
			),
			// UnlockDelayValue: RecycleBinConfig levels only (3, 4).
			UnlockDelayValue: firstNonZeroInt64(
				globalRBCfgMandatory.UnlockDelayValue,
				localRBCfgMandatory.UnlockDelayValue,
			),
			// UnlockDelayUnit: RecycleBinConfig levels only (3, 4).
			UnlockDelayUnit: firstNonEmptyString(
				globalRBCfgMandatory.UnlockDelayUnit,
				localRBCfgMandatory.UnlockDelayUnit,
			),
			// SyncedLabels: additive union from RecycleBinConfig levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localRBCfgMandatory.SyncedLabels,
				globalRBCfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localRBCfgMandatory.SyncedAnnotations,
				globalRBCfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localRBCfgMandatory.Tags,
				globalRBCfgMandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
		},
		Defaults: EffectiveEBSRecycleBinSection{
			RetentionPeriodValue: firstNonZeroInt64(
				localRBCfgDefaults.RetentionPeriodValue,
				globalRBCfgDefaults.RetentionPeriodValue,
				localKropathDefaults.RetentionPeriodValue,
				globalKropathDefaults.RetentionPeriodValue,
			),
			// RetentionPeriodUnit: RecycleBinConfig levels only (6, 7).
			RetentionPeriodUnit: firstNonEmptyString(
				localRBCfgDefaults.RetentionPeriodUnit,
				globalRBCfgDefaults.RetentionPeriodUnit,
			),
			// LockRule: firstTrue — L6 wins (strongest for defaults).
			LockRule: firstTrue(
				localRBCfgDefaults.LockRule,
				globalRBCfgDefaults.LockRule,
				localKropathDefaults.LockRule,
				globalKropathDefaults.LockRule,
			),
			// UnlockDelayValue: RecycleBinConfig levels only (6, 7).
			UnlockDelayValue: firstNonZeroInt64(
				localRBCfgDefaults.UnlockDelayValue,
				globalRBCfgDefaults.UnlockDelayValue,
			),
			// UnlockDelayUnit: RecycleBinConfig levels only (6, 7).
			UnlockDelayUnit: firstNonEmptyString(
				localRBCfgDefaults.UnlockDelayUnit,
				globalRBCfgDefaults.UnlockDelayUnit,
			),
			// SyncedLabels: additive union from RecycleBinConfig levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalRBCfgDefaults.SyncedLabels,
				localRBCfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalRBCfgDefaults.SyncedAnnotations,
				localRBCfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalRBCfgDefaults.Tags,
				localRBCfgDefaults.Tags,
			),
		},
	}
}
