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

// PipesKropathSection holds the EventBridge Pipes governance fields from
// KropathConfig.spec.mandatory.pipes / .defaults.pipes (ADR-015 §3.5)
// PLUS the tier-level tags, syncedLabels, and syncedAnnotations from
// KropathConfig.spec.mandatory / .defaults (populated by the reconciler so
// that the full org-wide field set flows through MergePipesCascade).
//
// Tags, syncedLabels, and syncedAnnotations are org-wide fields at the
// KropathConfig tier level — they do NOT appear under spec.pipes.
//
// Zero value of each scalar field is the permissive sentinel (not enforced).
type PipesKropathSection struct {
	// DesiredState is the enforced EventBridge Pipes desired state (RUNNING | STOPPED).
	// Empty string = not enforced.
	DesiredState string `json:"desiredState,omitempty"`

	// Tags are tier-level cloud resource tags from KropathConfig.spec.mandatory.tags
	// or KropathConfig.spec.defaults.tags. Populated by the reconciler from the
	// tier-level field, not from spec.mandatory.pipes. nil / empty = no tags.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are tier-level Kubernetes labels to propagate from
	// KropathConfig.spec.mandatory.syncedLabels or .defaults.syncedLabels.
	// Populated by the reconciler. nil / empty = no synced labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are tier-level Kubernetes annotations to propagate from
	// KropathConfig.spec.mandatory.syncedAnnotations or .defaults.syncedAnnotations.
	// Populated by the reconciler. nil / empty = no synced annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// PipesConfigSection holds the EventBridge Pipes governance fields from
// PipesConfig.spec.mandatory or PipesConfig.spec.defaults (per-type ResourceConfig,
// ADR-015 §3.5).
//
// Zero value of each scalar field is the permissive sentinel (not enforced).
type PipesConfigSection struct {
	// DesiredState is the enforced EventBridge Pipes desired state (RUNNING | STOPPED).
	// Empty = not enforced.
	DesiredState string `json:"desiredState,omitempty"`

	// NamingTemplate is the Pipes resource naming template (e.g. "{namespace}-{name}").
	// Empty = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// Tags are cloud resource tags. nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to Pipes cloud resources.
	// nil / empty = no synced labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to Pipes cloud resources.
	// nil / empty = no synced annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectivePipesSection is one tier (mandatory or defaults) of the merged Pipes governance
// result written into PipesConfig.status.effectiveConfig by the controller.
type EffectivePipesSection struct {
	DesiredState      string            `json:"desiredState,omitempty"`
	NamingTemplate    string            `json:"namingTemplate,omitempty"`
	Tags              map[string]string `json:"tags,omitempty"`
	SyncedLabels      map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectivePipesConfig is the merged Pipes governance result written into
// PipesConfig.status.effectiveConfig by the controller.
type EffectivePipesConfig struct {
	Mandatory EffectivePipesSection `json:"mandatory"`
	Defaults  EffectivePipesSection `json:"defaults"`
}

// MergePipesCascade merges EventBridge Pipes governance fields from all cascade sources
// and returns the effective configuration to be written to status.effectiveConfig.
//
// The ten-level priority chain (ADR-015 §5.3) for Pipes fields:
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace)
//	Level 3 — globalPipesCfgMandatory (PipesConfig in kro-system)
//	Level 4 — localPipesCfgMandatory  (PipesConfig in resource namespace)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localPipesCfgDefaults   (PipesConfig in resource namespace)
//	Level 7 — globalPipesCfgDefaults  (PipesConfig in kro-system)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system)
//
// For mandatory (levels 1–4): first non-zero value in priority order wins.
// For defaults (levels 6–9): first non-zero value in priority order wins.
// For tags / syncedLabels / syncedAnnotations: additive merge; lower level numbers win on key conflicts.
//
// desiredState appears at all four mandatory/defaults levels (both KropathConfig and PipesConfig).
// namingTemplate appears only at levels 3–4 (mandatory) and 6–7 (defaults) — not in KropathConfig.
// Tags, syncedLabels, and syncedAnnotations appear at all four mandatory/defaults levels;
// PipesKropathSection.Tags/.SyncedLabels/.SyncedAnnotations carry the tier-level fields from
// KropathConfig.mandatory / defaults (populated by the reconciler).
func MergePipesCascade(
	// Mandatory inputs (highest → lowest priority)
	globalKropathMandatory PipesKropathSection, // level 1
	localKropathMandatory PipesKropathSection, // level 2
	globalPipesCfgMandatory PipesConfigSection, // level 3
	localPipesCfgMandatory PipesConfigSection, // level 4
	// Defaults inputs (highest → lowest priority)
	localPipesCfgDefaults PipesConfigSection, // level 6
	globalPipesCfgDefaults PipesConfigSection, // level 7
	localKropathDefaults PipesKropathSection, // level 8
	globalKropathDefaults PipesKropathSection, // level 9
) EffectivePipesConfig {
	return EffectivePipesConfig{
		Mandatory: EffectivePipesSection{
			DesiredState: firstNonEmptyString(
				globalKropathMandatory.DesiredState,  // level 1
				localKropathMandatory.DesiredState,   // level 2
				globalPipesCfgMandatory.DesiredState, // level 3
				localPipesCfgMandatory.DesiredState,  // level 4
			),
			// namingTemplate not in KropathConfig: levels 3 and 4 only.
			NamingTemplate: firstNonEmptyString(
				globalPipesCfgMandatory.NamingTemplate, // level 3
				localPipesCfgMandatory.NamingTemplate,  // level 4
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflicts.
			Tags: mergeMaps(
				localPipesCfgMandatory.Tags,   // level 4 (lowest priority, set first)
				globalPipesCfgMandatory.Tags,  // level 3
				localKropathMandatory.Tags,    // level 2
				globalKropathMandatory.Tags,   // level 1 (highest priority, last to write)
			),
			// syncedLabels: all mandatory sources; L4 added first, L1 wins on key conflicts.
			SyncedLabels: mergeMaps(
				localPipesCfgMandatory.SyncedLabels,  // level 4
				globalPipesCfgMandatory.SyncedLabels, // level 3
				localKropathMandatory.SyncedLabels,   // level 2
				globalKropathMandatory.SyncedLabels,  // level 1
			),
			// syncedAnnotations: all mandatory sources; L4 added first, L1 wins on key conflicts.
			SyncedAnnotations: mergeMaps(
				localPipesCfgMandatory.SyncedAnnotations,  // level 4
				globalPipesCfgMandatory.SyncedAnnotations, // level 3
				localKropathMandatory.SyncedAnnotations,   // level 2
				globalKropathMandatory.SyncedAnnotations,  // level 1
			),
		},
		Defaults: EffectivePipesSection{
			DesiredState: firstNonEmptyString(
				localPipesCfgDefaults.DesiredState,  // level 6
				globalPipesCfgDefaults.DesiredState, // level 7
				localKropathDefaults.DesiredState,   // level 8
				globalKropathDefaults.DesiredState,  // level 9
			),
			// namingTemplate not in KropathConfig: levels 6 and 7 only.
			NamingTemplate: firstNonEmptyString(
				localPipesCfgDefaults.NamingTemplate,  // level 6
				globalPipesCfgDefaults.NamingTemplate, // level 7
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflicts.
			Tags: mergeMaps(
				globalKropathDefaults.Tags, // level 9 (lowest priority)
				localKropathDefaults.Tags,  // level 8
				globalPipesCfgDefaults.Tags, // level 7
				localPipesCfgDefaults.Tags,  // level 6 (highest priority)
			),
			// syncedLabels: all defaults sources; L9 added first, L6 wins on key conflicts.
			SyncedLabels: mergeMaps(
				globalKropathDefaults.SyncedLabels,  // level 9
				localKropathDefaults.SyncedLabels,   // level 8
				globalPipesCfgDefaults.SyncedLabels, // level 7
				localPipesCfgDefaults.SyncedLabels,  // level 6 (wins)
			),
			// syncedAnnotations: all defaults sources; L9 added first, L6 wins on key conflicts.
			SyncedAnnotations: mergeMaps(
				globalKropathDefaults.SyncedAnnotations,  // level 9
				localKropathDefaults.SyncedAnnotations,   // level 8
				globalPipesCfgDefaults.SyncedAnnotations, // level 7
				localPipesCfgDefaults.SyncedAnnotations,  // level 6 (wins)
			),
		},
	}
}
