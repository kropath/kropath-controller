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

// QuickSightKropathSection holds the QuickSight family governance fields from
// KropathConfig.spec.mandatory.quicksight / .defaults.quicksight (ADR-015 §3.5)
// PLUS the tier-level tags and synced maps from KropathConfig.spec.mandatory.tags
// (populated by the reconciler so that tag cascade flows through MergeQuickSightCascade).
//
// Zero value of each field is the permissive sentinel (not enforced):
// empty string for ImportMode, nil/empty for maps.
type QuickSightKropathSection struct {
	// ImportMode is the enforced dataset import mode.
	// "" (zero value) = not enforced.
	// Allowed values: SPICE, DIRECT_QUERY.
	ImportMode string `json:"importMode,omitempty"`

	// Tags are org-level cloud resource tags from KropathConfig tier-level Tags.
	// nil / empty map = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are K8s labels to propagate to created QuickSight resource CRs.
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are K8s annotations to propagate to created QuickSight resource CRs.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// QuickSightConfigSection holds the QuickSight governance fields from
// QuickSightConfig.spec.mandatory or QuickSightConfig.spec.defaults
// (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type QuickSightConfigSection struct {
	// ImportMode is the enforced dataset import mode.
	// "" (zero value) = not enforced.
	// Allowed values: SPICE, DIRECT_QUERY.
	ImportMode string `json:"importMode,omitempty"`

	// NamingTemplate is the resource naming template (e.g. "{namespace}-{name}").
	// Governed only at QuickSightConfig levels 3-4 (mandatory) and 6-7 (defaults).
	// KropathConfig.quicksight does NOT carry namingTemplate.
	// Empty string = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// Tags are cloud resource tags for this QuickSight config profile.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are K8s labels to propagate to created QuickSight resource CRs.
	// Additive map merge across all QuickSight cascade sources.
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are K8s annotations to propagate to created QuickSight resource CRs.
	// Additive map merge across all QuickSight cascade sources.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveQuickSightSection is one tier (mandatory or defaults) of the merged
// QuickSight governance result written into QuickSightConfig.status.effectiveConfig
// by the controller.
type EffectiveQuickSightSection struct {
	ImportMode        string            `json:"importMode,omitempty"`
	NamingTemplate    string            `json:"namingTemplate,omitempty"`
	Tags              map[string]string `json:"tags,omitempty"`
	SyncedLabels      map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveQuickSightConfig is the merged QuickSight governance result written into
// QuickSightConfig.status.effectiveConfig by the controller.
type EffectiveQuickSightConfig struct {
	Mandatory EffectiveQuickSightSection `json:"mandatory"`
	Defaults  EffectiveQuickSightSection `json:"defaults"`
}

// MergeQuickSightCascade merges QuickSight governance fields from all cascade sources
// and returns the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for QuickSight (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.quicksight)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.quicksight)
//	Level 3 — globalQSCfgMandatory    (QuickSightConfig in kro-system, mandatory)
//	Level 4 — localQSCfgMandatory     (QuickSightConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localQSCfgDefaults      (QuickSightConfig in resource namespace, defaults)
//	Level 7 — globalQSCfgDefaults     (QuickSightConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.quicksight)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.quicksight)
//
// ImportMode: KropathConfig levels 1-2 + QuickSightConfig levels 3-4 (mandatory);
//
//	QuickSightConfig levels 6-7 + KropathConfig levels 8-9 (defaults).
//
// NamingTemplate: QuickSightConfig levels only (3-4 mandatory, 6-7 defaults).
// Tags/SyncedLabels/SyncedAnnotations: additive union merge across all four
// mandatory levels (L4 added first, L1 wins on key conflict) and all four
// defaults levels (L9 added first, L6 wins on key conflict).
func MergeQuickSightCascade(
	globalKropathMandatory QuickSightKropathSection, // level 1
	localKropathMandatory QuickSightKropathSection, // level 2
	globalQSCfgMandatory QuickSightConfigSection, // level 3
	localQSCfgMandatory QuickSightConfigSection, // level 4
	localQSCfgDefaults QuickSightConfigSection, // level 6
	globalQSCfgDefaults QuickSightConfigSection, // level 7
	localKropathDefaults QuickSightKropathSection, // level 8
	globalKropathDefaults QuickSightKropathSection, // level 9
) EffectiveQuickSightConfig {
	return EffectiveQuickSightConfig{
		Mandatory: EffectiveQuickSightSection{
			// ImportMode: KropathConfig levels 1-2 + QuickSightConfig levels 3-4.
			// L1 highest priority, L4 lowest.
			ImportMode: firstNonEmptyString(
				globalKropathMandatory.ImportMode, // level 1
				localKropathMandatory.ImportMode,  // level 2
				globalQSCfgMandatory.ImportMode,   // level 3
				localQSCfgMandatory.ImportMode,    // level 4
			),
			// NamingTemplate: QuickSightConfig levels only (3, 4).
			NamingTemplate: firstNonEmptyString(
				globalQSCfgMandatory.NamingTemplate, // level 3
				localQSCfgMandatory.NamingTemplate,  // level 4
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localQSCfgMandatory.Tags,   // level 4 (lowest priority)
				globalQSCfgMandatory.Tags,  // level 3
				localKropathMandatory.Tags, // level 2
				globalKropathMandatory.Tags, // level 1 (highest priority)
			),
			// SyncedLabels: additive union across all mandatory sources.
			// L4 added first (lowest priority), L1 wins on key conflict.
			SyncedLabels: mergeMaps(
				localQSCfgMandatory.SyncedLabels,   // level 4
				globalQSCfgMandatory.SyncedLabels,  // level 3
				localKropathMandatory.SyncedLabels, // level 2
				globalKropathMandatory.SyncedLabels, // level 1
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localQSCfgMandatory.SyncedAnnotations,   // level 4
				globalQSCfgMandatory.SyncedAnnotations,  // level 3
				localKropathMandatory.SyncedAnnotations, // level 2
				globalKropathMandatory.SyncedAnnotations, // level 1
			),
		},
		Defaults: EffectiveQuickSightSection{
			// ImportMode: QuickSightConfig levels 6-7 + KropathConfig levels 8-9.
			// L6 highest priority, L9 lowest.
			ImportMode: firstNonEmptyString(
				localQSCfgDefaults.ImportMode,   // level 6
				globalQSCfgDefaults.ImportMode,  // level 7
				localKropathDefaults.ImportMode, // level 8
				globalKropathDefaults.ImportMode, // level 9
			),
			// NamingTemplate: QuickSightConfig levels only (6, 7).
			NamingTemplate: firstNonEmptyString(
				localQSCfgDefaults.NamingTemplate,  // level 6
				globalQSCfgDefaults.NamingTemplate, // level 7
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,  // level 9 (lowest priority)
				localKropathDefaults.Tags,   // level 8
				globalQSCfgDefaults.Tags,    // level 7
				localQSCfgDefaults.Tags,     // level 6 (highest priority)
			),
			// SyncedLabels: additive union across all defaults sources.
			// L9 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalKropathDefaults.SyncedLabels,  // level 9
				localKropathDefaults.SyncedLabels,   // level 8
				globalQSCfgDefaults.SyncedLabels,    // level 7
				localQSCfgDefaults.SyncedLabels,     // level 6
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalKropathDefaults.SyncedAnnotations,  // level 9
				localKropathDefaults.SyncedAnnotations,   // level 8
				globalQSCfgDefaults.SyncedAnnotations,    // level 7
				localQSCfgDefaults.SyncedAnnotations,     // level 6
			),
		},
	}
}
