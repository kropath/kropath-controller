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

// KinesisKropathSection holds the Kinesis-family governance fields from
// KropathConfig.spec.mandatory.kinesis / .defaults.kinesis (ADR-015 §3.5).
//
// Two fields have KropathConfig.kinesis equivalents (org-wide governance):
// streamMode and shardCount.
//
// namingTemplate is profile-specific and governed at the KinesisConfig level
// only — it does NOT appear in KropathConfig.kinesis.
//
// Zero value of each field is the permissive sentinel ("" / 0 = not enforced).
type KinesisKropathSection struct {
	// StreamMode enforces capacity mode org-wide.
	// "" (zero value) = not enforced. Valid values: "" | "on_demand" | "provisioned".
	StreamMode string `json:"streamMode,omitempty"`

	// ShardCount enforces shard count org-wide for provisioned streams.
	// 0 (zero value) = not enforced.
	ShardCount int64 `json:"shardCount,omitempty"`

	// Tags are tier-level cloud resource tags merged from KropathConfig.spec.mandatory.tags /
	// .defaults.tags — the generic cross-family tags, not a kinesis-scoped field.
	// nil / empty map = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// KinesisConfigSection holds the Kinesis governance fields from
// KinesisConfig.spec.mandatory or KinesisConfig.spec.defaults
// (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel ("" / 0 = not enforced).
type KinesisConfigSection struct {
	// StreamMode enforces capacity mode for streams in scope.
	// "" = not enforced. Valid values: "" | "on_demand" | "provisioned".
	StreamMode string `json:"streamMode,omitempty"`

	// ShardCount enforces shard count for provisioned streams.
	// 0 = not enforced.
	ShardCount int64 `json:"shardCount,omitempty"`

	// NamingTemplate is the naming template (e.g. "{namespace}-{name}").
	// "" = not enforced.
	// No KropathConfig.kinesis equivalent — profile-specific governance only.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// Tags are cloud resource tags for this Kinesis config profile.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created resources
	// (applied as both K8s labels and cloud tags per ADR-015 §6.1).
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to created resources.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveKinesisSection is one tier (mandatory or defaults) of the merged Kinesis
// governance result written into KinesisConfig.status.effectiveConfig.
type EffectiveKinesisSection struct {
	StreamMode        string            `json:"streamMode,omitempty"`
	ShardCount        int64             `json:"shardCount,omitempty"`
	NamingTemplate    string            `json:"namingTemplate,omitempty"`
	Tags              map[string]string `json:"tags,omitempty"`
	SyncedLabels      map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveKinesisConfig is the merged Kinesis governance result written into
// KinesisConfig.status.effectiveConfig by the controller.
type EffectiveKinesisConfig struct {
	Mandatory EffectiveKinesisSection `json:"mandatory"`
	Defaults  EffectiveKinesisSection `json:"defaults"`
}

// MergeKinesisCascade merges Kinesis governance fields from all cascade sources and
// returns the effective configuration to be written to status.effectiveConfig.
//
// Nine-level priority chain for Kinesis (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.kinesis)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.kinesis)
//	Level 3 — globalKinesisCfgMandatory   (KinesisConfig in kro-system, mandatory)
//	Level 4 — localKinesisCfgMandatory    (KinesisConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localKinesisCfgDefaults     (KinesisConfig in resource namespace, defaults)
//	Level 7 — globalKinesisCfgDefaults    (KinesisConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.kinesis)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.kinesis)
//
// Two fields have KropathConfig.kinesis equivalents (levels 1-2, 8-9):
//   streamMode, shardCount.
//
// One field has no KropathConfig.kinesis equivalent (levels 3-7 only):
//   namingTemplate.
//
// String merge: firstNonEmptyString in priority order.
// Integer merge: firstNonZeroInt64 in priority order.
// Tags: additive union merge across all mandatory levels, all defaults levels.
// SyncedLabels/SyncedAnnotations: additive union from KinesisConfig levels only (no KropathConfig).
func MergeKinesisCascade(
	globalKropathMandatory KinesisKropathSection, // level 1
	localKropathMandatory KinesisKropathSection, // level 2
	globalKinesisCfgMandatory KinesisConfigSection, // level 3
	localKinesisCfgMandatory KinesisConfigSection, // level 4
	localKinesisCfgDefaults KinesisConfigSection, // level 6
	globalKinesisCfgDefaults KinesisConfigSection, // level 7
	localKropathDefaults KinesisKropathSection, // level 8
	globalKropathDefaults KinesisKropathSection, // level 9
) EffectiveKinesisConfig {
	return EffectiveKinesisConfig{
		Mandatory: EffectiveKinesisSection{
			// KropathConfig levels (1-2) exist for streamMode and shardCount.
			StreamMode: firstNonEmptyString(
				globalKropathMandatory.StreamMode,
				localKropathMandatory.StreamMode,
				globalKinesisCfgMandatory.StreamMode,
				localKinesisCfgMandatory.StreamMode,
			),
			ShardCount: firstNonZeroInt64(
				globalKropathMandatory.ShardCount,
				localKropathMandatory.ShardCount,
				globalKinesisCfgMandatory.ShardCount,
				localKinesisCfgMandatory.ShardCount,
			),
			// No KropathConfig.kinesis level for namingTemplate.
			NamingTemplate: firstNonEmptyString(
				globalKinesisCfgMandatory.NamingTemplate,
				localKinesisCfgMandatory.NamingTemplate,
			),
			// SyncedLabels: additive union from KinesisConfig levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localKinesisCfgMandatory.SyncedLabels,
				globalKinesisCfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localKinesisCfgMandatory.SyncedAnnotations,
				globalKinesisCfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localKinesisCfgMandatory.Tags,
				globalKinesisCfgMandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
		},
		Defaults: EffectiveKinesisSection{
			// KropathConfig levels (8-9) exist for streamMode and shardCount.
			StreamMode: firstNonEmptyString(
				localKinesisCfgDefaults.StreamMode,
				globalKinesisCfgDefaults.StreamMode,
				localKropathDefaults.StreamMode,
				globalKropathDefaults.StreamMode,
			),
			ShardCount: firstNonZeroInt64(
				localKinesisCfgDefaults.ShardCount,
				globalKinesisCfgDefaults.ShardCount,
				localKropathDefaults.ShardCount,
				globalKropathDefaults.ShardCount,
			),
			// No KropathConfig.kinesis level for namingTemplate.
			NamingTemplate: firstNonEmptyString(
				localKinesisCfgDefaults.NamingTemplate,
				globalKinesisCfgDefaults.NamingTemplate,
			),
			// SyncedLabels: additive union from KinesisConfig levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalKinesisCfgDefaults.SyncedLabels,
				localKinesisCfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalKinesisCfgDefaults.SyncedAnnotations,
				localKinesisCfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalKinesisCfgDefaults.Tags,
				localKinesisCfgDefaults.Tags,
			),
		},
	}
}
