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

// AutoScalingKropathSection holds the Auto Scaling-family governance fields from
// KropathConfig.spec.mandatory.autoscaling / .defaults.autoscaling (ADR-015 §3.5)
// PLUS the tier-level tags from KropathConfig.spec.mandatory.tags (populated
// by the reconciler so that tag cascade flows through MergeAutoScalingCascade).
//
// Only two cross-type fields live here: newInstancesProtectedFromScaleIn and
// capacityRebalance. All other fields are per-profile in AutoScalingConfig.
//
// Zero value of each field is the permissive sentinel (not enforced).
type AutoScalingKropathSection struct {
	// NewInstancesProtectedFromScaleIn enforces instance protection from scale-in
	// org-wide when true. false (zero value) = not enforced.
	NewInstancesProtectedFromScaleIn bool `json:"newInstancesProtectedFromScaleIn,omitempty"`

	// CapacityRebalance enforces Spot capacity rebalancing org-wide when true.
	// false (zero value) = not enforced.
	CapacityRebalance bool `json:"capacityRebalance,omitempty"`

	// Tags are tier-level cloud resource tags from KropathConfig.spec.mandatory.tags.
	// Populated by the reconciler from the tier-level field, not from spec.mandatory.autoscaling.
	// nil / empty map (zero value) = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// AutoScalingConfigSection holds the Auto Scaling governance fields from
// AutoScalingConfig.spec.mandatory or AutoScalingConfig.spec.defaults
// (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type AutoScalingConfigSection struct {
	// NewInstancesProtectedFromScaleIn enforces instance protection when true.
	// false = not enforced.
	NewInstancesProtectedFromScaleIn bool `json:"newInstancesProtectedFromScaleIn,omitempty"`

	// CapacityRebalance enforces Spot capacity rebalancing when true.
	// false = not enforced.
	CapacityRebalance bool `json:"capacityRebalance,omitempty"`

	// HealthCheckType is the enforced ASG health check type (e.g. "EC2", "ELB", "EC2,ELB").
	// Governed at AutoScalingConfig levels only (not in KropathConfig).
	// Empty string = not enforced.
	HealthCheckType string `json:"healthCheckType,omitempty"`

	// HealthCheckGracePeriod is the health check grace period in seconds.
	// Governed at AutoScalingConfig levels only.
	// 0 (zero value) = not enforced.
	HealthCheckGracePeriod int64 `json:"healthCheckGracePeriod,omitempty"`

	// MaxInstanceLifetime is the maximum instance lifetime in seconds.
	// Governed at AutoScalingConfig levels only.
	// 0 (zero value) = disabled / not enforced.
	MaxInstanceLifetime int64 `json:"maxInstanceLifetime,omitempty"`

	// NamingTemplate is the cloud resource naming template.
	// Governed at AutoScalingConfig levels only.
	// Empty string = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created ASG resources.
	// Additive map merge across AutoScalingConfig tiers only.
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to created ASG resources.
	// Additive map merge across AutoScalingConfig tiers only.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`

	// Tags are cloud resource tags for this Auto Scaling config profile.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// EffectiveAutoScalingSection is one tier (mandatory or defaults) of the merged
// Auto Scaling governance result written into AutoScalingConfig.status.effectiveConfig
// by the controller.
type EffectiveAutoScalingSection struct {
	NewInstancesProtectedFromScaleIn bool              `json:"newInstancesProtectedFromScaleIn,omitempty"`
	CapacityRebalance                bool              `json:"capacityRebalance,omitempty"`
	HealthCheckType                  string            `json:"healthCheckType,omitempty"`
	HealthCheckGracePeriod           int64             `json:"healthCheckGracePeriod,omitempty"`
	MaxInstanceLifetime              int64             `json:"maxInstanceLifetime,omitempty"`
	NamingTemplate                   string            `json:"namingTemplate,omitempty"`
	SyncedLabels                     map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations                map[string]string `json:"syncedAnnotations,omitempty"`
	Tags                             map[string]string `json:"tags,omitempty"`
}

// EffectiveAutoScalingConfig is the merged Auto Scaling governance result written into
// AutoScalingConfig.status.effectiveConfig by the controller.
type EffectiveAutoScalingConfig struct {
	Mandatory EffectiveAutoScalingSection `json:"mandatory"`
	Defaults  EffectiveAutoScalingSection `json:"defaults"`
}

// MergeAutoScalingCascade merges Auto Scaling governance fields from all cascade
// sources and returns the effective configuration to be written to status.effectiveConfig.
//
// The ten-level priority chain (ADR-015 §5.3) for Auto Scaling fields:
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.autoscaling)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.autoscaling)
//	Level 3 — globalASCfgMandatory    (AutoScalingConfig in kro-system, mandatory)
//	Level 4 — localASCfgMandatory     (AutoScalingConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localASCfgDefaults      (AutoScalingConfig in resource namespace, defaults)
//	Level 7 — globalASCfgDefaults     (AutoScalingConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.autoscaling)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.autoscaling)
//
// For mandatory (levels 1–4): first non-zero value in priority order wins.
// For defaults (levels 6–9): first non-zero value in priority order wins.
// For tags: union merge across all sources; lower level numbers win on key conflicts.
//
// newInstancesProtectedFromScaleIn and capacityRebalance appear at all four
// mandatory/defaults levels (KropathConfig levels 1-2, 8-9 and AutoScalingConfig levels 3-4, 6-7).
// healthCheckType, healthCheckGracePeriod, maxInstanceLifetime, and namingTemplate are
// AutoScalingConfig-only (levels 3-4 mandatory, 6-7 defaults).
// SyncedLabels and SyncedAnnotations are AutoScalingConfig-only (levels 3-4, 6-7).
// Tags appear at all four mandatory/defaults levels; AutoScalingKropathSection.Tags carries the
// tier-level KropathConfig.mandatory.tags (populated by the reconciler).
func MergeAutoScalingCascade(
	// Mandatory inputs (highest → lowest priority)
	globalKropathMandatory AutoScalingKropathSection, // level 1
	localKropathMandatory AutoScalingKropathSection, // level 2
	globalASCfgMandatory AutoScalingConfigSection, // level 3
	localASCfgMandatory AutoScalingConfigSection, // level 4
	// Defaults inputs (highest → lowest priority)
	localASCfgDefaults AutoScalingConfigSection, // level 6
	globalASCfgDefaults AutoScalingConfigSection, // level 7
	localKropathDefaults AutoScalingKropathSection, // level 8
	globalKropathDefaults AutoScalingKropathSection, // level 9
) EffectiveAutoScalingConfig {
	return EffectiveAutoScalingConfig{
		Mandatory: EffectiveAutoScalingSection{
			NewInstancesProtectedFromScaleIn: firstTrue(
				globalKropathMandatory.NewInstancesProtectedFromScaleIn, // level 1
				localKropathMandatory.NewInstancesProtectedFromScaleIn,  // level 2
				globalASCfgMandatory.NewInstancesProtectedFromScaleIn,   // level 3
				localASCfgMandatory.NewInstancesProtectedFromScaleIn,    // level 4
			),
			CapacityRebalance: firstTrue(
				globalKropathMandatory.CapacityRebalance, // level 1
				localKropathMandatory.CapacityRebalance,  // level 2
				globalASCfgMandatory.CapacityRebalance,   // level 3
				localASCfgMandatory.CapacityRebalance,    // level 4
			),
			// HealthCheckType: AutoScalingConfig levels only (3, 4).
			HealthCheckType: firstNonEmptyString(
				globalASCfgMandatory.HealthCheckType, // level 3
				localASCfgMandatory.HealthCheckType,  // level 4
			),
			// HealthCheckGracePeriod: AutoScalingConfig levels only (3, 4).
			HealthCheckGracePeriod: firstNonZeroInt64(
				globalASCfgMandatory.HealthCheckGracePeriod, // level 3
				localASCfgMandatory.HealthCheckGracePeriod,  // level 4
			),
			// MaxInstanceLifetime: AutoScalingConfig levels only (3, 4).
			MaxInstanceLifetime: firstNonZeroInt64(
				globalASCfgMandatory.MaxInstanceLifetime, // level 3
				localASCfgMandatory.MaxInstanceLifetime,  // level 4
			),
			// NamingTemplate: AutoScalingConfig levels only (3, 4).
			NamingTemplate: firstNonEmptyString(
				globalASCfgMandatory.NamingTemplate, // level 3
				localASCfgMandatory.NamingTemplate,  // level 4
			),
			// SyncedLabels: additive union from AutoScalingConfig levels only.
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
		Defaults: EffectiveAutoScalingSection{
			NewInstancesProtectedFromScaleIn: firstTrue(
				localASCfgDefaults.NewInstancesProtectedFromScaleIn,   // level 6
				globalASCfgDefaults.NewInstancesProtectedFromScaleIn,  // level 7
				localKropathDefaults.NewInstancesProtectedFromScaleIn, // level 8
				globalKropathDefaults.NewInstancesProtectedFromScaleIn, // level 9
			),
			CapacityRebalance: firstTrue(
				localASCfgDefaults.CapacityRebalance,   // level 6
				globalASCfgDefaults.CapacityRebalance,  // level 7
				localKropathDefaults.CapacityRebalance, // level 8
				globalKropathDefaults.CapacityRebalance, // level 9
			),
			// HealthCheckType: AutoScalingConfig levels only (6, 7).
			HealthCheckType: firstNonEmptyString(
				localASCfgDefaults.HealthCheckType,  // level 6
				globalASCfgDefaults.HealthCheckType, // level 7
			),
			// HealthCheckGracePeriod: AutoScalingConfig levels only (6, 7).
			HealthCheckGracePeriod: firstNonZeroInt64(
				localASCfgDefaults.HealthCheckGracePeriod,  // level 6
				globalASCfgDefaults.HealthCheckGracePeriod, // level 7
			),
			// MaxInstanceLifetime: AutoScalingConfig levels only (6, 7).
			MaxInstanceLifetime: firstNonZeroInt64(
				localASCfgDefaults.MaxInstanceLifetime,  // level 6
				globalASCfgDefaults.MaxInstanceLifetime, // level 7
			),
			// NamingTemplate: AutoScalingConfig levels only (6, 7).
			NamingTemplate: firstNonEmptyString(
				localASCfgDefaults.NamingTemplate,  // level 6
				globalASCfgDefaults.NamingTemplate, // level 7
			),
			// SyncedLabels: additive union from AutoScalingConfig levels only.
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
