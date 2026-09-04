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

// CodeArtifactKropathSection holds the CodeArtifact-family governance fields from
// KropathConfig.spec.mandatory.codeartifact / .defaults.codeartifact (ADR-015 §3.5).
//
// Only 1 scalar field is governed at the KropathConfig level: encryptionKey.
// namingTemplate, syncedLabels, and syncedAnnotations are CodeArtifactConfig-only
// (family design §8).
//
// Zero value of each field is the permissive sentinel (not enforced).
type CodeArtifactKropathSection struct {
	// EncryptionKey is the full KMS key ARN to enforce for CodeArtifact domain
	// encryption at org-wide scope. Empty string = not enforced.
	EncryptionKey string `json:"encryptionKey,omitempty"`

	// Tags are tier-level cloud resource tags.
	// The reconciler populates this from KropathConfig.spec.mandatory.tags /
	// .defaults.tags so that tag union merge flows through MergeCodeArtifactCascade
	// alongside the CodeArtifact-specific fields.
	// nil / empty map = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// CodeArtifactConfigSection holds the CodeArtifact governance fields from
// CodeArtifactConfig.spec.mandatory or CodeArtifactConfig.spec.defaults
// (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type CodeArtifactConfigSection struct {
	// EncryptionKey is the full KMS key ARN for CodeArtifact domain encryption.
	// Empty string = not enforced (mandatory tier) / use AWS-managed key (defaults tier).
	EncryptionKey string `json:"encryptionKey,omitempty"`

	// NamingTemplate is the domain naming template (e.g. "{namespace}-{name}").
	// Governed only at CodeArtifactConfig levels 3-4 (mandatory) and 6-7 (defaults).
	// KropathConfig.codeartifact does NOT carry namingTemplate.
	// Empty string = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created domain resources.
	// Additive map merge across CodeArtifactConfig tiers only.
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to created domain resources.
	// Additive map merge across CodeArtifactConfig tiers only.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`

	// Tags are cloud resource tags for this CodeArtifact config profile.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// EffectiveCodeArtifactSection is one tier (mandatory or defaults) of the merged
// CodeArtifact governance result written into CodeArtifactConfig.status.effectiveConfig
// by the controller.
type EffectiveCodeArtifactSection struct {
	EncryptionKey     string            `json:"encryptionKey,omitempty"`
	NamingTemplate    string            `json:"namingTemplate,omitempty"`
	SyncedLabels      map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
	Tags              map[string]string `json:"tags,omitempty"`
}

// EffectiveCodeArtifactConfig is the merged CodeArtifact governance result written into
// CodeArtifactConfig.status.effectiveConfig by the controller.
type EffectiveCodeArtifactConfig struct {
	Mandatory EffectiveCodeArtifactSection `json:"mandatory"`
	Defaults  EffectiveCodeArtifactSection `json:"defaults"`
}

// MergeCodeArtifactCascade merges CodeArtifact governance fields from all cascade
// sources and returns the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for CodeArtifact (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.codeartifact)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.codeartifact)
//	Level 3 — globalCACfgMandatory    (CodeArtifactConfig in kro-system, mandatory)
//	Level 4 — localCACfgMandatory     (CodeArtifactConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localCACfgDefaults      (CodeArtifactConfig in resource namespace, defaults)
//	Level 7 — globalCACfgDefaults     (CodeArtifactConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.codeartifact)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.codeartifact)
//
// Scalar merge: firstNonEmptyString in priority order (lowest number wins).
// Tags: additive union merge across all four mandatory levels, all four defaults levels.
// SyncedLabels/SyncedAnnotations: additive union from CodeArtifactConfig levels only (no KropathConfig).
// NamingTemplate: governed only at CodeArtifactConfig levels (3-4 mandatory, 6-7 defaults).
func MergeCodeArtifactCascade(
	globalKropathMandatory CodeArtifactKropathSection, // level 1
	localKropathMandatory CodeArtifactKropathSection, // level 2
	globalCACfgMandatory CodeArtifactConfigSection, // level 3
	localCACfgMandatory CodeArtifactConfigSection, // level 4
	localCACfgDefaults CodeArtifactConfigSection, // level 6
	globalCACfgDefaults CodeArtifactConfigSection, // level 7
	localKropathDefaults CodeArtifactKropathSection, // level 8
	globalKropathDefaults CodeArtifactKropathSection, // level 9
) EffectiveCodeArtifactConfig {
	return EffectiveCodeArtifactConfig{
		Mandatory: EffectiveCodeArtifactSection{
			EncryptionKey: firstNonEmptyString(
				globalKropathMandatory.EncryptionKey,
				localKropathMandatory.EncryptionKey,
				globalCACfgMandatory.EncryptionKey,
				localCACfgMandatory.EncryptionKey,
			),
			// NamingTemplate: CodeArtifactConfig levels only (3, 4);
			// KropathConfig has no namingTemplate field for codeartifact.
			NamingTemplate: firstNonEmptyString(
				globalCACfgMandatory.NamingTemplate,
				localCACfgMandatory.NamingTemplate,
			),
			// SyncedLabels: additive union from CodeArtifactConfig levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localCACfgMandatory.SyncedLabels,
				globalCACfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localCACfgMandatory.SyncedAnnotations,
				globalCACfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localCACfgMandatory.Tags,
				globalCACfgMandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
		},
		Defaults: EffectiveCodeArtifactSection{
			EncryptionKey: firstNonEmptyString(
				localCACfgDefaults.EncryptionKey,
				globalCACfgDefaults.EncryptionKey,
				localKropathDefaults.EncryptionKey,
				globalKropathDefaults.EncryptionKey,
			),
			// NamingTemplate: CodeArtifactConfig levels only (6, 7).
			NamingTemplate: firstNonEmptyString(
				localCACfgDefaults.NamingTemplate,
				globalCACfgDefaults.NamingTemplate,
			),
			// SyncedLabels: additive union from CodeArtifactConfig levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalCACfgDefaults.SyncedLabels,
				localCACfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalCACfgDefaults.SyncedAnnotations,
				localCACfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalCACfgDefaults.Tags,
				localCACfgDefaults.Tags,
			),
		},
	}
}
