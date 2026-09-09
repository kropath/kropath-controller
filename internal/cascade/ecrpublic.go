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

// ECRPublicKropathSection holds the ECR Public governance fields sourced from
// KropathConfig tier-level tags. ECR Public adds no service-specific fields to
// KropathConfig (planning doc T-02); only the global tags/syncedLabels/syncedAnnotations
// flow through this section, populated by the reconciler from the tier-level fields.
//
// Zero value is the permissive sentinel (no enforcement).
type ECRPublicKropathSection struct {
	// Tags are tier-level cloud resource tags from KropathConfig.spec.mandatory.tags
	// or .defaults.tags. Populated by the reconciler; nil/empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are Kubernetes labels from KropathConfig.spec.mandatory.syncedLabels
	// or .defaults.syncedLabels. Populated by the reconciler.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations from
	// KropathConfig.spec.mandatory.syncedAnnotations or .defaults.syncedAnnotations.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// ECRPublicConfigSection holds the ECR Public governance fields from
// ECRPublicConfig.spec.mandatory or ECRPublicConfig.spec.defaults (ADR-015 §3.5).
//
// Zero value is the permissive sentinel (not enforced).
type ECRPublicConfigSection struct {
	// NamingTemplate is the ECR Public repository naming template
	// (e.g. "{namespace}/{name}"). Governed only at ECRPublicConfig levels 3-4
	// (mandatory) and 6-7 (defaults). Empty = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// Tags are cloud resource tags for this ECRPublicConfig profile.
	// nil/empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to ECR Public resource CRs.
	// Additive map merge across ECRPublicConfig tiers only.
	// nil/empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to ECR Public resource CRs.
	// Additive map merge across ECRPublicConfig tiers only.
	// nil/empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveECRPublicSection is one tier (mandatory or defaults) of the merged ECR Public
// governance result written into ECRPublicConfig.status.effectiveConfig by the controller.
type EffectiveECRPublicSection struct {
	NamingTemplate    string            `json:"namingTemplate,omitempty"`
	Tags              map[string]string `json:"tags,omitempty"`
	SyncedLabels      map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveECRPublicConfig is the merged ECR Public governance result written into
// ECRPublicConfig.status.effectiveConfig by the controller.
type EffectiveECRPublicConfig struct {
	Mandatory EffectiveECRPublicSection `json:"mandatory"`
	Defaults  EffectiveECRPublicSection `json:"defaults"`
}

// MergeECRPublicCascade merges ECR Public governance fields from all cascade sources
// and returns the effective configuration to be written to status.effectiveConfig.
//
// Nine-level priority chain for ECR Public (ADR-015 §5.3).
// ECR Public adds no service-specific fields to KropathConfig, so only tier-level
// tags/syncedLabels/syncedAnnotations flow from the KPC levels:
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory tier tags)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory tier tags)
//	Level 3 — globalECRPubCfgMandatory  (ECRPublicConfig in kro-system, mandatory)
//	Level 4 — localECRPubCfgMandatory   (ECRPublicConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localECRPubCfgDefaults    (ECRPublicConfig in resource namespace, defaults)
//	Level 7 — globalECRPubCfgDefaults   (ECRPublicConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults tier tags)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults tier tags)
//
// Scalar merge (NamingTemplate): firstNonEmptyString at ECRPublicConfig levels only (3-4
// mandatory, 6-7 defaults). KropathConfig has no namingTemplate for ECR Public.
// Tags: additive union merge across all four mandatory levels (L1 wins on key conflict)
// and all four defaults levels (L6 wins on key conflict).
// SyncedLabels/SyncedAnnotations: additive union from ECRPublicConfig levels only
// (L3, L4 mandatory; L6, L7 defaults). KropathConfig does not govern synced labels/annotations
// for ECR Public.
func MergeECRPublicCascade(
	globalKropathMandatory ECRPublicKropathSection, // level 1
	localKropathMandatory ECRPublicKropathSection, // level 2
	globalECRPubCfgMandatory ECRPublicConfigSection, // level 3
	localECRPubCfgMandatory ECRPublicConfigSection, // level 4
	localECRPubCfgDefaults ECRPublicConfigSection, // level 6
	globalECRPubCfgDefaults ECRPublicConfigSection, // level 7
	localKropathDefaults ECRPublicKropathSection, // level 8
	globalKropathDefaults ECRPublicKropathSection, // level 9
) EffectiveECRPublicConfig {
	return EffectiveECRPublicConfig{
		Mandatory: EffectiveECRPublicSection{
			// NamingTemplate: ECRPublicConfig levels only (3, 4).
			// KropathConfig has no namingTemplate for ECR Public.
			NamingTemplate: firstNonEmptyString(
				globalECRPubCfgMandatory.NamingTemplate, // level 3
				localECRPubCfgMandatory.NamingTemplate,  // level 4
			),
			// SyncedLabels: additive union from ECRPublicConfig levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localECRPubCfgMandatory.SyncedLabels,
				globalECRPubCfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localECRPubCfgMandatory.SyncedAnnotations,
				globalECRPubCfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localECRPubCfgMandatory.Tags,   // level 4 (lowest priority)
				globalECRPubCfgMandatory.Tags,  // level 3
				localKropathMandatory.Tags,     // level 2
				globalKropathMandatory.Tags,    // level 1 (highest priority)
			),
		},
		Defaults: EffectiveECRPublicSection{
			// NamingTemplate: ECRPublicConfig levels only (6, 7).
			NamingTemplate: firstNonEmptyString(
				localECRPubCfgDefaults.NamingTemplate,  // level 6
				globalECRPubCfgDefaults.NamingTemplate, // level 7
			),
			// SyncedLabels: additive union from ECRPublicConfig levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalECRPubCfgDefaults.SyncedLabels,
				localECRPubCfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalECRPubCfgDefaults.SyncedAnnotations,
				localECRPubCfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,     // level 9 (lowest priority)
				localKropathDefaults.Tags,      // level 8
				globalECRPubCfgDefaults.Tags,   // level 7
				localECRPubCfgDefaults.Tags,    // level 6 (highest priority)
			),
		},
	}
}
