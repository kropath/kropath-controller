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

// SESKropathSection holds the standard KropathConfig governance fields that affect
// the SES family. SES requires no SES-specific KropathConfig extension (unlike
// families with encryption mode, capacity, or access control governance fields).
//
// The reconciler populates this from the top-level KropathConfigTier fields
// (Tags, SyncedLabels, SyncedAnnotations) since there is no KropathConfigTier.SES
// section.
//
// Zero value of each field is the permissive sentinel (not enforced).
type SESKropathSection struct {
	// Tags are org-level cloud resource tags.
	// nil / empty map = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are K8s labels to propagate to created SES resource CRs.
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are K8s annotations to propagate to created SES resource CRs.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// SESConfigSection holds the SES governance fields from SESConfig.spec.mandatory
// or SESConfig.spec.defaults (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type SESConfigSection struct {
	// NamingTemplate is the configuration set naming template (e.g. "{namespace}-{name}").
	// Governed only at SESConfig levels 3-4 (mandatory) and 6-7 (defaults).
	// KropathConfig has no SES-specific namingTemplate field.
	// Empty string = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// Tags are cloud resource tags for this SES config profile.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are K8s labels to propagate to created SES resource CRs.
	// Additive map merge across all SES cascade sources.
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are K8s annotations to propagate to created SES resource CRs.
	// Additive map merge across all SES cascade sources.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveSESSection is one tier (mandatory or defaults) of the merged SES governance
// result written into SESConfig.status.effectiveConfig by the controller.
type EffectiveSESSection struct {
	NamingTemplate    string            `json:"namingTemplate,omitempty"`
	Tags              map[string]string `json:"tags,omitempty"`
	SyncedLabels      map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveSESConfig is the merged SES governance result written into
// SESConfig.status.effectiveConfig by the controller.
type EffectiveSESConfig struct {
	Mandatory EffectiveSESSection `json:"mandatory"`
	Defaults  EffectiveSESSection `json:"defaults"`
}

// MergeSESCascade merges SES governance fields from all cascade sources and returns
// the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for SES (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, standard mandatory fields)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, standard mandatory fields)
//	Level 3 — globalSESCfgMandatory   (SESConfig in kro-system, mandatory)
//	Level 4 — localSESCfgMandatory    (SESConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localSESCfgDefaults     (SESConfig in resource namespace, defaults)
//	Level 7 — globalSESCfgDefaults    (SESConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, standard defaults fields)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, standard defaults fields)
//
// NamingTemplate: SESConfig levels only (3-4 mandatory, 6-7 defaults).
// KropathConfig has no SES-specific namingTemplate field.
//
// Tags/SyncedLabels/SyncedAnnotations: additive union merge across all four
// mandatory levels (L4 added first, L1 wins on key conflict) and all four
// defaults levels (L9 added first, L6 wins on key conflict).
func MergeSESCascade(
	globalKropathMandatory SESKropathSection, // level 1
	localKropathMandatory SESKropathSection, // level 2
	globalSESCfgMandatory SESConfigSection, // level 3
	localSESCfgMandatory SESConfigSection, // level 4
	localSESCfgDefaults SESConfigSection, // level 6
	globalSESCfgDefaults SESConfigSection, // level 7
	localKropathDefaults SESKropathSection, // level 8
	globalKropathDefaults SESKropathSection, // level 9
) EffectiveSESConfig {
	return EffectiveSESConfig{
		Mandatory: EffectiveSESSection{
			// NamingTemplate: SESConfig levels only (3, 4); no KropathConfig source.
			NamingTemplate: firstNonEmptyString(
				globalSESCfgMandatory.NamingTemplate,
				localSESCfgMandatory.NamingTemplate,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localSESCfgMandatory.Tags,
				globalSESCfgMandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
			// SyncedLabels: additive union across all mandatory sources.
			// L4 added first (lowest priority), L1 wins on key conflict.
			SyncedLabels: mergeMaps(
				localSESCfgMandatory.SyncedLabels,
				globalSESCfgMandatory.SyncedLabels,
				localKropathMandatory.SyncedLabels,
				globalKropathMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localSESCfgMandatory.SyncedAnnotations,
				globalSESCfgMandatory.SyncedAnnotations,
				localKropathMandatory.SyncedAnnotations,
				globalKropathMandatory.SyncedAnnotations,
			),
		},
		Defaults: EffectiveSESSection{
			// NamingTemplate: SESConfig levels only (6, 7); no KropathConfig source.
			NamingTemplate: firstNonEmptyString(
				localSESCfgDefaults.NamingTemplate,
				globalSESCfgDefaults.NamingTemplate,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalSESCfgDefaults.Tags,
				localSESCfgDefaults.Tags,
			),
			// SyncedLabels: additive union across all defaults sources.
			// L9 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalKropathDefaults.SyncedLabels,
				localKropathDefaults.SyncedLabels,
				globalSESCfgDefaults.SyncedLabels,
				localSESCfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalKropathDefaults.SyncedAnnotations,
				localKropathDefaults.SyncedAnnotations,
				globalSESCfgDefaults.SyncedAnnotations,
				localSESCfgDefaults.SyncedAnnotations,
			),
		},
	}
}
