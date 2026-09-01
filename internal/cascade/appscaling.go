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

// AppScalingKropathSection holds the Application Auto Scaling-family governance
// fields from KropathConfig.spec.mandatory.appScaling / .defaults.appScaling
// (ADR-015 §3.5) PLUS the tier-level tags from KropathConfig.spec.mandatory.tags
// (populated by the reconciler so that tag cascade flows through
// MergeAppScalingCascade).
//
// Zero value of each field is the permissive sentinel (not enforced).
type AppScalingKropathSection struct {
	// MinCapacity enforces an org-wide minimum capacity floor when > 0.
	// 0 = not enforced.
	MinCapacity int64 `json:"minCapacity,omitempty"`

	// MaxCapacity enforces an org-wide maximum capacity ceiling when > 0.
	// 0 = not enforced.
	MaxCapacity int64 `json:"maxCapacity,omitempty"`

	// DisableScaleIn enforces scale-in protection org-wide when true.
	// false (zero value) = not enforced.
	DisableScaleIn bool `json:"disableScaleIn,omitempty"`

	// ScaleInCooldown enforces a minimum scale-in cooldown (seconds) when > 0.
	// 0 = not enforced.
	ScaleInCooldown int64 `json:"scaleInCooldown,omitempty"`

	// ScaleOutCooldown enforces a minimum scale-out cooldown (seconds) when > 0.
	// 0 = not enforced.
	ScaleOutCooldown int64 `json:"scaleOutCooldown,omitempty"`

	// Tags are tier-level cloud resource tags from KropathConfig.spec.mandatory.tags.
	// Populated by the reconciler from the tier-level field, not from
	// spec.mandatory.appScaling. nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// AppScalingConfigSection holds the Application Auto Scaling governance fields
// from AppScalingConfig.spec.mandatory or AppScalingConfig.spec.defaults
// (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type AppScalingConfigSection struct {
	// MinCapacity enforces a minimum capacity floor when > 0. 0 = not enforced.
	MinCapacity int64 `json:"minCapacity,omitempty"`

	// MaxCapacity enforces a maximum capacity ceiling when > 0. 0 = not enforced.
	MaxCapacity int64 `json:"maxCapacity,omitempty"`

	// DisableScaleIn enforces scale-in protection when true. false = not enforced.
	DisableScaleIn bool `json:"disableScaleIn,omitempty"`

	// ScaleInCooldown enforces a minimum scale-in cooldown (seconds) when > 0.
	// 0 = not enforced.
	ScaleInCooldown int64 `json:"scaleInCooldown,omitempty"`

	// ScaleOutCooldown enforces a minimum scale-out cooldown (seconds) when > 0.
	// 0 = not enforced.
	ScaleOutCooldown int64 `json:"scaleOutCooldown,omitempty"`

	// Tags are cloud resource tags for this AppScaling config profile.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created AppScaling resources.
	// Additive map merge across AppScalingConfig tiers only. nil / empty = none.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to created AppScaling resources.
	// Additive map merge across AppScalingConfig tiers only. nil / empty = none.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveAppScalingSection is one tier (mandatory or defaults) of the merged
// Application Auto Scaling governance result written into
// AppScalingConfig.status.effectiveConfig by the controller.
type EffectiveAppScalingSection struct {
	MinCapacity      int64             `json:"minCapacity,omitempty"`
	MaxCapacity      int64             `json:"maxCapacity,omitempty"`
	DisableScaleIn   bool              `json:"disableScaleIn,omitempty"`
	ScaleInCooldown  int64             `json:"scaleInCooldown,omitempty"`
	ScaleOutCooldown int64             `json:"scaleOutCooldown,omitempty"`
	Tags             map[string]string `json:"tags,omitempty"`
	SyncedLabels     map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveAppScalingConfig is the merged Application Auto Scaling governance
// result written into AppScalingConfig.status.effectiveConfig by the controller.
type EffectiveAppScalingConfig struct {
	Mandatory EffectiveAppScalingSection `json:"mandatory"`
	Defaults  EffectiveAppScalingSection `json:"defaults"`
}

// MergeAppScalingCascade merges Application Auto Scaling governance fields from
// all cascade sources and returns the effective configuration to be written to
// status.effectiveConfig.
//
// The ten-level priority chain (ADR-015 §5.3) for AppScaling fields:
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.appScaling)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.appScaling)
//	Level 3 — globalASCfgMandatory    (AppScalingConfig in kro-system, mandatory)
//	Level 4 — localASCfgMandatory     (AppScalingConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localASCfgDefaults      (AppScalingConfig in resource namespace, defaults)
//	Level 7 — globalASCfgDefaults     (AppScalingConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.appScaling)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.appScaling)
//
// For mandatory (levels 1–4): first non-zero value in priority order wins.
// For defaults (levels 6–9): first non-zero value in priority order wins.
// For tags: union merge across all sources; lower level numbers win on key conflicts.
//
// minCapacity, maxCapacity, scaleInCooldown, scaleOutCooldown appear at all four
// mandatory/defaults levels (KropathConfig levels 1-2, 8-9 and AppScalingConfig
// levels 3-4, 6-7).
// disableScaleIn appears at all four mandatory/defaults levels.
// SyncedLabels and SyncedAnnotations are AppScalingConfig-only (levels 3-4, 6-7).
// Tags appear at all four mandatory/defaults levels; AppScalingKropathSection.Tags
// carries the tier-level KropathConfig.mandatory.tags (populated by the reconciler).
func MergeAppScalingCascade(
	// Mandatory inputs (highest → lowest priority)
	globalKropathMandatory AppScalingKropathSection, // level 1
	localKropathMandatory AppScalingKropathSection,  // level 2
	globalASCfgMandatory AppScalingConfigSection,    // level 3
	localASCfgMandatory AppScalingConfigSection,     // level 4
	// Defaults inputs (highest → lowest priority)
	localASCfgDefaults AppScalingConfigSection,     // level 6
	globalASCfgDefaults AppScalingConfigSection,    // level 7
	localKropathDefaults AppScalingKropathSection,  // level 8
	globalKropathDefaults AppScalingKropathSection, // level 9
) EffectiveAppScalingConfig {
	return EffectiveAppScalingConfig{
		Mandatory: EffectiveAppScalingSection{
			MinCapacity: firstNonZeroInt64(
				globalKropathMandatory.MinCapacity, // level 1
				localKropathMandatory.MinCapacity,  // level 2
				globalASCfgMandatory.MinCapacity,   // level 3
				localASCfgMandatory.MinCapacity,    // level 4
			),
			MaxCapacity: firstNonZeroInt64(
				globalKropathMandatory.MaxCapacity, // level 1
				localKropathMandatory.MaxCapacity,  // level 2
				globalASCfgMandatory.MaxCapacity,   // level 3
				localASCfgMandatory.MaxCapacity,    // level 4
			),
			DisableScaleIn: firstTrue(
				globalKropathMandatory.DisableScaleIn, // level 1
				localKropathMandatory.DisableScaleIn,  // level 2
				globalASCfgMandatory.DisableScaleIn,   // level 3
				localASCfgMandatory.DisableScaleIn,    // level 4
			),
			ScaleInCooldown: firstNonZeroInt64(
				globalKropathMandatory.ScaleInCooldown, // level 1
				localKropathMandatory.ScaleInCooldown,  // level 2
				globalASCfgMandatory.ScaleInCooldown,   // level 3
				localASCfgMandatory.ScaleInCooldown,    // level 4
			),
			ScaleOutCooldown: firstNonZeroInt64(
				globalKropathMandatory.ScaleOutCooldown, // level 1
				localKropathMandatory.ScaleOutCooldown,  // level 2
				globalASCfgMandatory.ScaleOutCooldown,   // level 3
				localASCfgMandatory.ScaleOutCooldown,    // level 4
			),
			// SyncedLabels: additive union from AppScalingConfig levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localASCfgMandatory.SyncedLabels,
				globalASCfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localASCfgMandatory.SyncedAnnotations,
				globalASCfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localASCfgMandatory.Tags,    // level 4 (lowest priority, set first)
				globalASCfgMandatory.Tags,   // level 3
				localKropathMandatory.Tags,  // level 2
				globalKropathMandatory.Tags, // level 1 (highest priority, last to write)
			),
		},
		Defaults: EffectiveAppScalingSection{
			MinCapacity: firstNonZeroInt64(
				localASCfgDefaults.MinCapacity,   // level 6
				globalASCfgDefaults.MinCapacity,  // level 7
				localKropathDefaults.MinCapacity, // level 8
				globalKropathDefaults.MinCapacity, // level 9
			),
			MaxCapacity: firstNonZeroInt64(
				localASCfgDefaults.MaxCapacity,   // level 6
				globalASCfgDefaults.MaxCapacity,  // level 7
				localKropathDefaults.MaxCapacity, // level 8
				globalKropathDefaults.MaxCapacity, // level 9
			),
			DisableScaleIn: firstTrue(
				localASCfgDefaults.DisableScaleIn,   // level 6
				globalASCfgDefaults.DisableScaleIn,  // level 7
				localKropathDefaults.DisableScaleIn, // level 8
				globalKropathDefaults.DisableScaleIn, // level 9
			),
			ScaleInCooldown: firstNonZeroInt64(
				localASCfgDefaults.ScaleInCooldown,   // level 6
				globalASCfgDefaults.ScaleInCooldown,  // level 7
				localKropathDefaults.ScaleInCooldown, // level 8
				globalKropathDefaults.ScaleInCooldown, // level 9
			),
			ScaleOutCooldown: firstNonZeroInt64(
				localASCfgDefaults.ScaleOutCooldown,   // level 6
				globalASCfgDefaults.ScaleOutCooldown,  // level 7
				localKropathDefaults.ScaleOutCooldown, // level 8
				globalKropathDefaults.ScaleOutCooldown, // level 9
			),
			// SyncedLabels: additive union from AppScalingConfig levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalASCfgDefaults.SyncedLabels,
				localASCfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalASCfgDefaults.SyncedAnnotations,
				localASCfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags, // level 9 (lowest priority)
				localKropathDefaults.Tags,  // level 8
				globalASCfgDefaults.Tags,   // level 7
				localASCfgDefaults.Tags,    // level 6 (highest priority)
			),
		},
	}
}
