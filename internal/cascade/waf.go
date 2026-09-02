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

// WAFKropathSection holds the WAF-family governance fields from
// KropathConfig.spec.mandatory.waf / .defaults.waf (ADR-015 §3.5).
//
// Four scalar fields are governed at the KropathConfig level: scope,
// defaultAction, cloudWatchMetricsEnabled, sampledRequestsEnabled.
// namingTemplate, syncedLabels, and syncedAnnotations are WAFConfig-only.
//
// Zero value of each field is the permissive sentinel (not enforced):
//   - string: "" = not enforced
//   - bool: false = not enforced
type WAFKropathSection struct {
	// Scope is the org-wide WAF scope enforcement ("CLOUDFRONT" | "REGIONAL").
	// Empty string = not enforced.
	Scope string `json:"scope,omitempty"`

	// DefaultAction is the org-wide default action enforcement ("allow" | "block").
	// Empty string = not enforced.
	DefaultAction string `json:"defaultAction,omitempty"`

	// CloudWatchMetricsEnabled enforces org-wide CloudWatch metrics for WAF resources.
	// false = not enforced; true = require CloudWatch metrics.
	CloudWatchMetricsEnabled bool `json:"cloudWatchMetricsEnabled,omitempty"`

	// SampledRequestsEnabled enforces org-wide sampled request recording.
	// false = not enforced; true = require sampled requests.
	SampledRequestsEnabled bool `json:"sampledRequestsEnabled,omitempty"`

	// Tags are tier-level cloud resource tags.
	// The reconciler populates this from KropathConfig.spec.mandatory.tags /
	// .defaults.tags so that tag union merge flows through MergeWAFCascade
	// alongside the WAF-specific fields.
	// nil / empty map = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// WAFConfigSection holds the WAF governance fields from WAFConfig.spec.mandatory
// or WAFConfig.spec.defaults (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type WAFConfigSection struct {
	// Scope is the WAF scope ("CLOUDFRONT" | "REGIONAL"). Empty string = not enforced.
	Scope string `json:"scope,omitempty"`

	// DefaultAction is the default action ("allow" | "block"). Empty string = not enforced.
	DefaultAction string `json:"defaultAction,omitempty"`

	// CloudWatchMetricsEnabled enables CloudWatch metrics for this profile.
	// false = not enforced; true = require CloudWatch metrics.
	CloudWatchMetricsEnabled bool `json:"cloudWatchMetricsEnabled,omitempty"`

	// SampledRequestsEnabled enables sampled request recording for this profile.
	// false = not enforced; true = require sampled requests.
	SampledRequestsEnabled bool `json:"sampledRequestsEnabled,omitempty"`

	// NamingTemplate is the WAF resource naming template.
	// Governed only at WAFConfig levels 3-4 (mandatory) and 6-7 (defaults).
	// Empty string = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// Tags are cloud resource tags for this WAF config profile.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are Kubernetes labels propagated to WAF resources.
	// Additive map merge across WAFConfig tiers only (ADR-015 §6.1).
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations propagated to WAF resources.
	// Additive map merge across WAFConfig tiers only.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveWAFSection is one tier (mandatory or defaults) of the merged
// WAF governance result written into WAFConfig.status.effectiveConfig.
type EffectiveWAFSection struct {
	Scope                    string            `json:"scope,omitempty"`
	DefaultAction            string            `json:"defaultAction,omitempty"`
	CloudWatchMetricsEnabled bool              `json:"cloudWatchMetricsEnabled,omitempty"`
	SampledRequestsEnabled   bool              `json:"sampledRequestsEnabled,omitempty"`
	NamingTemplate           string            `json:"namingTemplate,omitempty"`
	Tags                     map[string]string `json:"tags,omitempty"`
	SyncedLabels             map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations        map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveWAFConfig is the merged WAF governance result written into
// WAFConfig.status.effectiveConfig by the controller.
type EffectiveWAFConfig struct {
	Mandatory EffectiveWAFSection `json:"mandatory"`
	Defaults  EffectiveWAFSection `json:"defaults"`
}

// MergeWAFCascade merges WAF governance fields from all cascade sources and
// returns the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for WAF (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.waf)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.waf)
//	Level 3 — globalWAFCfgMandatory   (WAFConfig in kro-system, mandatory)
//	Level 4 — localWAFCfgMandatory    (WAFConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localWAFCfgDefaults     (WAFConfig in resource namespace, defaults)
//	Level 7 — globalWAFCfgDefaults    (WAFConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.waf)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.waf)
//
// Scalar merge: firstNonEmptyString / firstTrue in priority order (lowest number wins).
// Tags: additive union merge across all four mandatory levels, all four defaults levels.
// SyncedLabels/SyncedAnnotations: additive union from WAFConfig levels only (no KropathConfig).
// NamingTemplate: governed only at WAFConfig levels (3-4 mandatory, 6-7 defaults).
func MergeWAFCascade(
	globalKropathMandatory WAFKropathSection, // level 1
	localKropathMandatory WAFKropathSection,  // level 2
	globalWAFCfgMandatory WAFConfigSection,   // level 3
	localWAFCfgMandatory WAFConfigSection,    // level 4
	localWAFCfgDefaults WAFConfigSection,     // level 6
	globalWAFCfgDefaults WAFConfigSection,    // level 7
	localKropathDefaults WAFKropathSection,   // level 8
	globalKropathDefaults WAFKropathSection,  // level 9
) EffectiveWAFConfig {
	return EffectiveWAFConfig{
		Mandatory: EffectiveWAFSection{
			// scope: KropathConfig levels 1-2 then WAFConfig levels 3-4.
			Scope: firstNonEmptyString(
				globalKropathMandatory.Scope,
				localKropathMandatory.Scope,
				globalWAFCfgMandatory.Scope,
				localWAFCfgMandatory.Scope,
			),
			// defaultAction: KropathConfig levels 1-2 then WAFConfig levels 3-4.
			DefaultAction: firstNonEmptyString(
				globalKropathMandatory.DefaultAction,
				localKropathMandatory.DefaultAction,
				globalWAFCfgMandatory.DefaultAction,
				localWAFCfgMandatory.DefaultAction,
			),
			// cloudWatchMetricsEnabled: KropathConfig levels 1-2 then WAFConfig levels 3-4.
			CloudWatchMetricsEnabled: firstTrue(
				globalKropathMandatory.CloudWatchMetricsEnabled,
				localKropathMandatory.CloudWatchMetricsEnabled,
				globalWAFCfgMandatory.CloudWatchMetricsEnabled,
				localWAFCfgMandatory.CloudWatchMetricsEnabled,
			),
			// sampledRequestsEnabled: KropathConfig levels 1-2 then WAFConfig levels 3-4.
			SampledRequestsEnabled: firstTrue(
				globalKropathMandatory.SampledRequestsEnabled,
				localKropathMandatory.SampledRequestsEnabled,
				globalWAFCfgMandatory.SampledRequestsEnabled,
				localWAFCfgMandatory.SampledRequestsEnabled,
			),
			// namingTemplate: WAFConfig levels only (3, 4).
			NamingTemplate: firstNonEmptyString(
				globalWAFCfgMandatory.NamingTemplate,
				localWAFCfgMandatory.NamingTemplate,
			),
			// syncedLabels: additive union from WAFConfig levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localWAFCfgMandatory.SyncedLabels,
				globalWAFCfgMandatory.SyncedLabels,
			),
			// syncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localWAFCfgMandatory.SyncedAnnotations,
				globalWAFCfgMandatory.SyncedAnnotations,
			),
			// tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localWAFCfgMandatory.Tags,
				globalWAFCfgMandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
		},
		Defaults: EffectiveWAFSection{
			// scope: WAFConfig defaults levels 6-7 then KropathConfig levels 8-9.
			Scope: firstNonEmptyString(
				localWAFCfgDefaults.Scope,
				globalWAFCfgDefaults.Scope,
				localKropathDefaults.Scope,
				globalKropathDefaults.Scope,
			),
			// defaultAction: WAFConfig defaults levels 6-7 then KropathConfig levels 8-9.
			DefaultAction: firstNonEmptyString(
				localWAFCfgDefaults.DefaultAction,
				globalWAFCfgDefaults.DefaultAction,
				localKropathDefaults.DefaultAction,
				globalKropathDefaults.DefaultAction,
			),
			// cloudWatchMetricsEnabled: WAFConfig defaults levels 6-7 then KropathConfig levels 8-9.
			CloudWatchMetricsEnabled: firstTrue(
				localWAFCfgDefaults.CloudWatchMetricsEnabled,
				globalWAFCfgDefaults.CloudWatchMetricsEnabled,
				localKropathDefaults.CloudWatchMetricsEnabled,
				globalKropathDefaults.CloudWatchMetricsEnabled,
			),
			// sampledRequestsEnabled: WAFConfig defaults levels 6-7 then KropathConfig levels 8-9.
			SampledRequestsEnabled: firstTrue(
				localWAFCfgDefaults.SampledRequestsEnabled,
				globalWAFCfgDefaults.SampledRequestsEnabled,
				localKropathDefaults.SampledRequestsEnabled,
				globalKropathDefaults.SampledRequestsEnabled,
			),
			// namingTemplate: WAFConfig levels only (6, 7).
			NamingTemplate: firstNonEmptyString(
				localWAFCfgDefaults.NamingTemplate,
				globalWAFCfgDefaults.NamingTemplate,
			),
			// syncedLabels: additive union from WAFConfig levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalWAFCfgDefaults.SyncedLabels,
				localWAFCfgDefaults.SyncedLabels,
			),
			// syncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalWAFCfgDefaults.SyncedAnnotations,
				localWAFCfgDefaults.SyncedAnnotations,
			),
			// tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalWAFCfgDefaults.Tags,
				localWAFCfgDefaults.Tags,
			),
		},
	}
}
