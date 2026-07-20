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

// SQSKropathSection holds the SQS-family governance fields from
// KropathConfig.spec.mandatory.sqs / .defaults.sqs (ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
// namingTemplate, syncedLabels, and syncedAnnotations are NOT present here;
// they live only in SQSConfig (SQSConfigSection).
type SQSKropathSection struct {
	// EncryptionType is the SQS encryption mode ("kms", "sqs-managed", "").
	// Empty string = not enforced.
	EncryptionType string `json:"encryptionType,omitempty"`

	// KmsMasterKeyId is the KMS key ID/ARN to enforce for SSE-KMS.
	// Empty string = not enforced.
	KmsMasterKeyId string `json:"kmsMasterKeyId,omitempty"`

	// VisibilityTimeout is the message visibility timeout in seconds.
	// 0 (zero value) = not enforced.
	VisibilityTimeout int64 `json:"visibilityTimeout,omitempty"`

	// MessageRetentionPeriod is the message retention period in seconds.
	// 0 (zero value) = not enforced.
	MessageRetentionPeriod int64 `json:"messageRetentionPeriod,omitempty"`

	// DelaySeconds is the delivery delay in seconds.
	// 0 (zero value) = not enforced.
	DelaySeconds int64 `json:"delaySeconds,omitempty"`

	// MaximumMessageSize is the maximum message size in bytes.
	// 0 (zero value) = not enforced.
	MaximumMessageSize int64 `json:"maximumMessageSize,omitempty"`

	// Tags are tier-level cloud resource tags.
	// The reconciler populates this from KropathConfig.spec.mandatory.tags / .defaults.tags
	// so that tag union merge flows through MergeSQSCascade alongside the SQS-specific fields.
	// nil / empty map = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// SQSConfigSection holds the SQS governance fields from SQSConfig.spec.mandatory
// or SQSConfig.spec.defaults (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type SQSConfigSection struct {
	// EncryptionType is the SQS encryption mode. Empty string = not enforced.
	EncryptionType string `json:"encryptionType,omitempty"`

	// KmsMasterKeyId is the KMS key ID/ARN. Empty string = not enforced.
	KmsMasterKeyId string `json:"kmsMasterKeyId,omitempty"`

	// VisibilityTimeout is the message visibility timeout in seconds. 0 = not enforced.
	VisibilityTimeout int64 `json:"visibilityTimeout,omitempty"`

	// MessageRetentionPeriod is the message retention period in seconds. 0 = not enforced.
	MessageRetentionPeriod int64 `json:"messageRetentionPeriod,omitempty"`

	// DelaySeconds is the delivery delay in seconds. 0 = not enforced.
	DelaySeconds int64 `json:"delaySeconds,omitempty"`

	// MaximumMessageSize is the maximum message size in bytes. 0 = not enforced.
	MaximumMessageSize int64 `json:"maximumMessageSize,omitempty"`

	// NamingTemplate is the queue naming template (e.g. "{namespace}-{name}").
	// Governed only at SQSConfig levels 3-4 (mandatory) and 6-7 (defaults).
	// KropathConfig.sqs does NOT carry namingTemplate.
	// Empty string = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created queue resources.
	// Additive map merge across SQSConfig tiers only.
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to created queue resources.
	// Additive map merge across SQSConfig tiers only.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`

	// Tags are cloud resource tags for this SQS config profile.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// EffectiveSQSSection is one tier (mandatory or defaults) of the merged SQS governance
// result written into SQSConfig.status.effectiveConfig by the controller.
type EffectiveSQSSection struct {
	EncryptionType         string            `json:"encryptionType,omitempty"`
	KmsMasterKeyId         string            `json:"kmsMasterKeyId,omitempty"`
	VisibilityTimeout      int64             `json:"visibilityTimeout,omitempty"`
	MessageRetentionPeriod int64             `json:"messageRetentionPeriod,omitempty"`
	DelaySeconds           int64             `json:"delaySeconds,omitempty"`
	MaximumMessageSize     int64             `json:"maximumMessageSize,omitempty"`
	NamingTemplate         string            `json:"namingTemplate,omitempty"`
	SyncedLabels           map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations      map[string]string `json:"syncedAnnotations,omitempty"`
	Tags                   map[string]string `json:"tags,omitempty"`
}

// EffectiveSQSConfig is the merged SQS governance result written into
// SQSConfig.status.effectiveConfig by the controller.
type EffectiveSQSConfig struct {
	Mandatory EffectiveSQSSection `json:"mandatory"`
	Defaults  EffectiveSQSSection `json:"defaults"`
}

// MergeSQSCascade merges SQS governance fields from all cascade sources and
// returns the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for SQS (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.sqs)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.sqs)
//	Level 3 — globalSQSCfgMandatory   (SQSConfig in kro-system, mandatory)
//	Level 4 — localSQSCfgMandatory    (SQSConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localSQSCfgDefaults     (SQSConfig in resource namespace, defaults)
//	Level 7 — globalSQSCfgDefaults    (SQSConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.sqs)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.sqs)
//
// Scalar merge: firstNonEmptyString / firstNonZeroInt64 in priority order (lowest number wins).
// Tags: additive union merge across all four mandatory levels, all four defaults levels.
// SyncedLabels/SyncedAnnotations: additive union from SQSConfig levels only (no KropathConfig).
// NamingTemplate: governed only at SQSConfig levels (3-4 mandatory, 6-7 defaults).
func MergeSQSCascade(
	globalKropathMandatory SQSKropathSection, // level 1
	localKropathMandatory SQSKropathSection, // level 2
	globalSQSCfgMandatory SQSConfigSection, // level 3
	localSQSCfgMandatory SQSConfigSection, // level 4
	localSQSCfgDefaults SQSConfigSection, // level 6
	globalSQSCfgDefaults SQSConfigSection, // level 7
	localKropathDefaults SQSKropathSection, // level 8
	globalKropathDefaults SQSKropathSection, // level 9
) EffectiveSQSConfig {
	return EffectiveSQSConfig{
		Mandatory: EffectiveSQSSection{
			EncryptionType: firstNonEmptyString(
				globalKropathMandatory.EncryptionType,
				localKropathMandatory.EncryptionType,
				globalSQSCfgMandatory.EncryptionType,
				localSQSCfgMandatory.EncryptionType,
			),
			KmsMasterKeyId: firstNonEmptyString(
				globalKropathMandatory.KmsMasterKeyId,
				localKropathMandatory.KmsMasterKeyId,
				globalSQSCfgMandatory.KmsMasterKeyId,
				localSQSCfgMandatory.KmsMasterKeyId,
			),
			VisibilityTimeout: firstNonZeroInt64(
				globalKropathMandatory.VisibilityTimeout,
				localKropathMandatory.VisibilityTimeout,
				globalSQSCfgMandatory.VisibilityTimeout,
				localSQSCfgMandatory.VisibilityTimeout,
			),
			MessageRetentionPeriod: firstNonZeroInt64(
				globalKropathMandatory.MessageRetentionPeriod,
				localKropathMandatory.MessageRetentionPeriod,
				globalSQSCfgMandatory.MessageRetentionPeriod,
				localSQSCfgMandatory.MessageRetentionPeriod,
			),
			DelaySeconds: firstNonZeroInt64(
				globalKropathMandatory.DelaySeconds,
				localKropathMandatory.DelaySeconds,
				globalSQSCfgMandatory.DelaySeconds,
				localSQSCfgMandatory.DelaySeconds,
			),
			MaximumMessageSize: firstNonZeroInt64(
				globalKropathMandatory.MaximumMessageSize,
				localKropathMandatory.MaximumMessageSize,
				globalSQSCfgMandatory.MaximumMessageSize,
				localSQSCfgMandatory.MaximumMessageSize,
			),
			// NamingTemplate: SQSConfig levels only (3, 4); KropathConfig has no namingTemplate.
			NamingTemplate: firstNonEmptyString(
				globalSQSCfgMandatory.NamingTemplate,
				localSQSCfgMandatory.NamingTemplate,
			),
			// SyncedLabels: additive union from SQSConfig levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localSQSCfgMandatory.SyncedLabels,
				globalSQSCfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localSQSCfgMandatory.SyncedAnnotations,
				globalSQSCfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localSQSCfgMandatory.Tags,
				globalSQSCfgMandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
		},
		Defaults: EffectiveSQSSection{
			EncryptionType: firstNonEmptyString(
				localSQSCfgDefaults.EncryptionType,
				globalSQSCfgDefaults.EncryptionType,
				localKropathDefaults.EncryptionType,
				globalKropathDefaults.EncryptionType,
			),
			KmsMasterKeyId: firstNonEmptyString(
				localSQSCfgDefaults.KmsMasterKeyId,
				globalSQSCfgDefaults.KmsMasterKeyId,
				localKropathDefaults.KmsMasterKeyId,
				globalKropathDefaults.KmsMasterKeyId,
			),
			VisibilityTimeout: firstNonZeroInt64(
				localSQSCfgDefaults.VisibilityTimeout,
				globalSQSCfgDefaults.VisibilityTimeout,
				localKropathDefaults.VisibilityTimeout,
				globalKropathDefaults.VisibilityTimeout,
			),
			MessageRetentionPeriod: firstNonZeroInt64(
				localSQSCfgDefaults.MessageRetentionPeriod,
				globalSQSCfgDefaults.MessageRetentionPeriod,
				localKropathDefaults.MessageRetentionPeriod,
				globalKropathDefaults.MessageRetentionPeriod,
			),
			DelaySeconds: firstNonZeroInt64(
				localSQSCfgDefaults.DelaySeconds,
				globalSQSCfgDefaults.DelaySeconds,
				localKropathDefaults.DelaySeconds,
				globalKropathDefaults.DelaySeconds,
			),
			MaximumMessageSize: firstNonZeroInt64(
				localSQSCfgDefaults.MaximumMessageSize,
				globalSQSCfgDefaults.MaximumMessageSize,
				localKropathDefaults.MaximumMessageSize,
				globalKropathDefaults.MaximumMessageSize,
			),
			// NamingTemplate: SQSConfig levels only (6, 7).
			NamingTemplate: firstNonEmptyString(
				localSQSCfgDefaults.NamingTemplate,
				globalSQSCfgDefaults.NamingTemplate,
			),
			// SyncedLabels: additive union from SQSConfig levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalSQSCfgDefaults.SyncedLabels,
				localSQSCfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalSQSCfgDefaults.SyncedAnnotations,
				localSQSCfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalSQSCfgDefaults.Tags,
				localSQSCfgDefaults.Tags,
			),
		},
	}
}
