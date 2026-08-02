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

// CloudWatchLogsKropathSection holds the CloudWatch Logs-family governance fields from
// KropathConfig.spec.mandatory.cloudwatchlogs / .defaults.cloudwatchlogs (ADR-015 §3.5).
//
// Only 2 scalar fields are governed at the KropathConfig level: kmsKeyId and
// retentionDays. namingTemplate, syncedLabels, and syncedAnnotations are
// CloudWatchLogsConfig-only (family design §8).
//
// Zero value of each field is the permissive sentinel (not enforced).
type CloudWatchLogsKropathSection struct {
	// KmsKeyId is the full KMS key ARN to enforce for log group encryption.
	// Empty string = not enforced. CloudWatch Logs requires a full key ARN
	// (no AWS-managed service key alias exists for CloudWatch Logs).
	KmsKeyId string `json:"kmsKeyId,omitempty"`

	// RetentionDays is the log group retention period in days.
	// 0 (zero value) = not enforced; first-non-zero-wins in cascade.
	// Must be one of the CloudWatch Logs allowed values when non-zero.
	RetentionDays int64 `json:"retentionDays,omitempty"`

	// Tags are tier-level cloud resource tags.
	// The reconciler populates this from KropathConfig.spec.mandatory.tags / .defaults.tags
	// so that tag union merge flows through MergeCloudWatchLogsCascade alongside the
	// CloudWatch Logs-specific fields.
	// nil / empty map = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// CloudWatchLogsConfigSection holds the CloudWatch Logs governance fields from
// CloudWatchLogsConfig.spec.mandatory or CloudWatchLogsConfig.spec.defaults
// (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type CloudWatchLogsConfigSection struct {
	// KmsKeyId is the full KMS key ARN. Empty string = not enforced.
	KmsKeyId string `json:"kmsKeyId,omitempty"`

	// RetentionDays is the log group retention period in days. 0 = not enforced.
	RetentionDays int64 `json:"retentionDays,omitempty"`

	// NamingTemplate is the log group naming template (e.g. "{namespace}-{name}").
	// Governed only at CloudWatchLogsConfig levels 3-4 (mandatory) and 6-7 (defaults).
	// KropathConfig.cloudwatchlogs does NOT carry namingTemplate.
	// Empty string = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created log group resources.
	// Additive map merge across CloudWatchLogsConfig tiers only.
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to created log group resources.
	// Additive map merge across CloudWatchLogsConfig tiers only.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`

	// Tags are cloud resource tags for this CloudWatch Logs config profile.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// EffectiveCloudWatchLogsSection is one tier (mandatory or defaults) of the merged
// CloudWatch Logs governance result written into CloudWatchLogsConfig.status.effectiveConfig
// by the controller.
type EffectiveCloudWatchLogsSection struct {
	KmsKeyId          string            `json:"kmsKeyId,omitempty"`
	RetentionDays     int64             `json:"retentionDays,omitempty"`
	NamingTemplate    string            `json:"namingTemplate,omitempty"`
	SyncedLabels      map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
	Tags              map[string]string `json:"tags,omitempty"`
}

// EffectiveCloudWatchLogsConfig is the merged CloudWatch Logs governance result written into
// CloudWatchLogsConfig.status.effectiveConfig by the controller.
type EffectiveCloudWatchLogsConfig struct {
	Mandatory EffectiveCloudWatchLogsSection `json:"mandatory"`
	Defaults  EffectiveCloudWatchLogsSection `json:"defaults"`
}

// MergeCloudWatchLogsCascade merges CloudWatch Logs governance fields from all cascade
// sources and returns the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for CloudWatch Logs (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.cloudwatchlogs)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.cloudwatchlogs)
//	Level 3 — globalCWLCfgMandatory   (CloudWatchLogsConfig in kro-system, mandatory)
//	Level 4 — localCWLCfgMandatory    (CloudWatchLogsConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localCWLCfgDefaults     (CloudWatchLogsConfig in resource namespace, defaults)
//	Level 7 — globalCWLCfgDefaults    (CloudWatchLogsConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.cloudwatchlogs)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.cloudwatchlogs)
//
// Scalar merge: firstNonEmptyString / firstNonZeroInt64 in priority order (lowest number wins).
// Tags: additive union merge across all four mandatory levels, all four defaults levels.
// SyncedLabels/SyncedAnnotations: additive union from CloudWatchLogsConfig levels only (no KropathConfig).
// NamingTemplate: governed only at CloudWatchLogsConfig levels (3-4 mandatory, 6-7 defaults).
func MergeCloudWatchLogsCascade(
	globalKropathMandatory CloudWatchLogsKropathSection, // level 1
	localKropathMandatory CloudWatchLogsKropathSection, // level 2
	globalCWLCfgMandatory CloudWatchLogsConfigSection, // level 3
	localCWLCfgMandatory CloudWatchLogsConfigSection, // level 4
	localCWLCfgDefaults CloudWatchLogsConfigSection, // level 6
	globalCWLCfgDefaults CloudWatchLogsConfigSection, // level 7
	localKropathDefaults CloudWatchLogsKropathSection, // level 8
	globalKropathDefaults CloudWatchLogsKropathSection, // level 9
) EffectiveCloudWatchLogsConfig {
	return EffectiveCloudWatchLogsConfig{
		Mandatory: EffectiveCloudWatchLogsSection{
			KmsKeyId: firstNonEmptyString(
				globalKropathMandatory.KmsKeyId,
				localKropathMandatory.KmsKeyId,
				globalCWLCfgMandatory.KmsKeyId,
				localCWLCfgMandatory.KmsKeyId,
			),
			RetentionDays: firstNonZeroInt64(
				globalKropathMandatory.RetentionDays,
				localKropathMandatory.RetentionDays,
				globalCWLCfgMandatory.RetentionDays,
				localCWLCfgMandatory.RetentionDays,
			),
			// NamingTemplate: CloudWatchLogsConfig levels only (3, 4);
			// KropathConfig has no namingTemplate field for cloudwatchlogs.
			NamingTemplate: firstNonEmptyString(
				globalCWLCfgMandatory.NamingTemplate,
				localCWLCfgMandatory.NamingTemplate,
			),
			// SyncedLabels: additive union from CloudWatchLogsConfig levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localCWLCfgMandatory.SyncedLabels,
				globalCWLCfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localCWLCfgMandatory.SyncedAnnotations,
				globalCWLCfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localCWLCfgMandatory.Tags,
				globalCWLCfgMandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
		},
		Defaults: EffectiveCloudWatchLogsSection{
			KmsKeyId: firstNonEmptyString(
				localCWLCfgDefaults.KmsKeyId,
				globalCWLCfgDefaults.KmsKeyId,
				localKropathDefaults.KmsKeyId,
				globalKropathDefaults.KmsKeyId,
			),
			RetentionDays: firstNonZeroInt64(
				localCWLCfgDefaults.RetentionDays,
				globalCWLCfgDefaults.RetentionDays,
				localKropathDefaults.RetentionDays,
				globalKropathDefaults.RetentionDays,
			),
			// NamingTemplate: CloudWatchLogsConfig levels only (6, 7).
			NamingTemplate: firstNonEmptyString(
				localCWLCfgDefaults.NamingTemplate,
				globalCWLCfgDefaults.NamingTemplate,
			),
			// SyncedLabels: additive union from CloudWatchLogsConfig levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalCWLCfgDefaults.SyncedLabels,
				localCWLCfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalCWLCfgDefaults.SyncedAnnotations,
				localCWLCfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalCWLCfgDefaults.Tags,
				localCWLCfgDefaults.Tags,
			),
		},
	}
}
