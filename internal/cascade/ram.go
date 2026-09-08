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

// RAMKropathSection holds the RAM-family governance fields from
// KropathConfig.spec.mandatory.ram / .defaults.ram (ADR-015 §3.5)
// PLUS the tier-level tags from KropathConfig.spec.mandatory.tags (populated
// by the reconciler so that tag cascade flows through MergeRAMCascade).
//
// Zero value of each field is the permissive sentinel (not enforced).
type RAMKropathSection struct {
	// AllowExternalPrincipals controls org-wide external principal enforcement.
	// nil = not enforced at this KropathConfig tier; false = block; true = allow.
	AllowExternalPrincipals *bool `json:"allowExternalPrincipals,omitempty"`

	// AllowedResourceTypes is the org-wide allowed shareable resource type list.
	// Treated as a scalar (first-non-empty wins); nil / empty = no restriction.
	AllowedResourceTypes []string `json:"allowedResourceTypes,omitempty"`

	// Tags are tier-level cloud resource tags from KropathConfig.spec.mandatory.tags.
	// Populated by the reconciler from the tier-level field, not from spec.mandatory.ram.
	// nil / empty map = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are K8s labels to propagate to created RAM resource CRs.
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are K8s annotations to propagate to created RAM resource CRs.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// RAMConfigSection holds the RAM governance fields from RAMConfig.spec.mandatory
// or RAMConfig.spec.defaults (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type RAMConfigSection struct {
	// AllowExternalPrincipals controls whether external principals may be added to shares.
	// nil = not set (falls through); false = explicitly disallow; true = explicitly allow.
	AllowExternalPrincipals *bool `json:"allowExternalPrincipals,omitempty"`

	// AllowedResourceTypes is the allowed shareable resource type list for this profile.
	// Treated as a scalar (first-non-empty wins); nil / empty = not enforced.
	AllowedResourceTypes []string `json:"allowedResourceTypes,omitempty"`

	// NamingTemplate is the resource share naming template (e.g. "{namespace}-{name}").
	// Governed only at RAMConfig levels 3-4 (mandatory) and 6-7 (defaults).
	// KropathConfig has no RAM-specific namingTemplate field.
	// Empty string = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// Tags are cloud resource tags for this RAM config profile.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are K8s labels to propagate to created RAM resource CRs.
	// Additive map merge across all RAM cascade sources.
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are K8s annotations to propagate to created RAM resource CRs.
	// Additive map merge across all RAM cascade sources.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveRAMSection is one tier (mandatory or defaults) of the merged RAM governance
// result written into RAMConfig.status.effectiveConfig by the controller.
type EffectiveRAMSection struct {
	AllowExternalPrincipals *bool             `json:"allowExternalPrincipals,omitempty"`
	AllowedResourceTypes    []string          `json:"allowedResourceTypes,omitempty"`
	NamingTemplate          string            `json:"namingTemplate,omitempty"`
	Tags                    map[string]string `json:"tags,omitempty"`
	SyncedLabels            map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations       map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveRAMConfig is the merged RAM governance result written into
// RAMConfig.status.effectiveConfig by the controller.
type EffectiveRAMConfig struct {
	Mandatory EffectiveRAMSection `json:"mandatory"`
	Defaults  EffectiveRAMSection `json:"defaults"`
}

// MergeRAMCascade merges RAM governance fields from all cascade sources and returns
// the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for RAM (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory)
//	Level 3 — globalRAMCfgMandatory   (RAMConfig in kro-system, mandatory)
//	Level 4 — localRAMCfgMandatory    (RAMConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localRAMCfgDefaults     (RAMConfig in resource namespace, defaults)
//	Level 7 — globalRAMCfgDefaults    (RAMConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults)
//
// AllowExternalPrincipals: *bool — nil = not enforced (falls through); firstNonNilBoolPtr priority.
// AllowedResourceTypes: first-non-empty list wins; no union merge.
// NamingTemplate: RAMConfig levels only (3-4 mandatory, 6-7 defaults).
// Tags/SyncedLabels/SyncedAnnotations: additive union merge across all four
// mandatory levels (L4 added first, L1 wins on key conflict) and all four
// defaults levels (L9 added first, L6 wins on key conflict).
func MergeRAMCascade(
	globalKropathMandatory RAMKropathSection, // level 1
	localKropathMandatory RAMKropathSection, // level 2
	globalRAMCfgMandatory RAMConfigSection, // level 3
	localRAMCfgMandatory RAMConfigSection, // level 4
	localRAMCfgDefaults RAMConfigSection, // level 6
	globalRAMCfgDefaults RAMConfigSection, // level 7
	localKropathDefaults RAMKropathSection, // level 8
	globalKropathDefaults RAMKropathSection, // level 9
) EffectiveRAMConfig {
	return EffectiveRAMConfig{
		Mandatory: EffectiveRAMSection{
			// AllowExternalPrincipals: *bool — L1 wins; nil = not enforced.
			AllowExternalPrincipals: firstNonNilBoolPtr(
				globalKropathMandatory.AllowExternalPrincipals,
				localKropathMandatory.AllowExternalPrincipals,
				globalRAMCfgMandatory.AllowExternalPrincipals,
				localRAMCfgMandatory.AllowExternalPrincipals,
			),
			// AllowedResourceTypes: first-non-empty list wins.
			AllowedResourceTypes: firstNonEmptyStrings(
				globalKropathMandatory.AllowedResourceTypes,
				localKropathMandatory.AllowedResourceTypes,
				globalRAMCfgMandatory.AllowedResourceTypes,
				localRAMCfgMandatory.AllowedResourceTypes,
			),
			// NamingTemplate: RAMConfig levels only (L3, L4); no KropathConfig source.
			NamingTemplate: firstNonEmptyString(
				globalRAMCfgMandatory.NamingTemplate,
				localRAMCfgMandatory.NamingTemplate,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localRAMCfgMandatory.Tags,
				globalRAMCfgMandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
			// SyncedLabels: additive union across all mandatory sources.
			// L4 added first (lowest priority), L1 wins on key conflict.
			SyncedLabels: mergeMaps(
				localRAMCfgMandatory.SyncedLabels,
				globalRAMCfgMandatory.SyncedLabels,
				localKropathMandatory.SyncedLabels,
				globalKropathMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localRAMCfgMandatory.SyncedAnnotations,
				globalRAMCfgMandatory.SyncedAnnotations,
				localKropathMandatory.SyncedAnnotations,
				globalKropathMandatory.SyncedAnnotations,
			),
		},
		Defaults: EffectiveRAMSection{
			// AllowExternalPrincipals: *bool — L6 wins; nil = not enforced.
			AllowExternalPrincipals: firstNonNilBoolPtr(
				localRAMCfgDefaults.AllowExternalPrincipals,
				globalRAMCfgDefaults.AllowExternalPrincipals,
				localKropathDefaults.AllowExternalPrincipals,
				globalKropathDefaults.AllowExternalPrincipals,
			),
			// AllowedResourceTypes: first-non-empty list wins (L6 first).
			AllowedResourceTypes: firstNonEmptyStrings(
				localRAMCfgDefaults.AllowedResourceTypes,
				globalRAMCfgDefaults.AllowedResourceTypes,
				localKropathDefaults.AllowedResourceTypes,
				globalKropathDefaults.AllowedResourceTypes,
			),
			// NamingTemplate: RAMConfig levels only (L6, L7); no KropathConfig source.
			NamingTemplate: firstNonEmptyString(
				localRAMCfgDefaults.NamingTemplate,
				globalRAMCfgDefaults.NamingTemplate,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalRAMCfgDefaults.Tags,
				localRAMCfgDefaults.Tags,
			),
			// SyncedLabels: additive union across all defaults sources.
			// L9 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalKropathDefaults.SyncedLabels,
				localKropathDefaults.SyncedLabels,
				globalRAMCfgDefaults.SyncedLabels,
				localRAMCfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalKropathDefaults.SyncedAnnotations,
				localKropathDefaults.SyncedAnnotations,
				globalRAMCfgDefaults.SyncedAnnotations,
				localRAMCfgDefaults.SyncedAnnotations,
			),
		},
	}
}
