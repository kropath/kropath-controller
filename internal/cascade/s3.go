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

// Package cascade implements the ten-level governance cascade for kropath
// ResourceConfig CRDs, following ADR-010 and ADR-015 §5.3.
package cascade

// S3Section holds the S3-specific governance fields from
// KropathConfig.spec.mandatory.s3 / .defaults.s3 (ADR-015 §3.5).
//
// Tags are the only map field present here; NamingTemplate, SyncedLabels, and
// SyncedAnnotations are S3Config-only (S3ConfigSection). The reconciler populates
// Tags from KropathConfig.spec.mandatory.tags / .defaults.tags so that tag union
// merge flows through MergeS3Cascade alongside the S3-specific fields.
//
// Zero value of each field is the permissive sentinel (not enforced).
type S3Section struct {
	// EncryptionAlgorithm is the bucket encryption mode.
	// Empty string = no enforcement.
	EncryptionAlgorithm string `json:"encryptionAlgorithm,omitempty"`

	// KmsKeyArn is the KMS key ARN to enforce for bucket encryption.
	// Empty string = no enforcement.
	KmsKeyArn string `json:"kmsKeyArn,omitempty"`

	// BlockPublicAccess forces S3 public access blocking when true.
	// false (zero value) = not enforced.
	BlockPublicAccess bool `json:"blockPublicAccess,omitempty"`

	// Versioning is the enforced bucket versioning state.
	// Empty string = no enforcement.
	Versioning string `json:"versioning,omitempty"`

	// LoggingEnabled enforces server access logging when true.
	// false (zero value) = not enforced.
	LoggingEnabled bool `json:"loggingEnabled,omitempty"`

	// LogDeliveryBucket is the target bucket for server access logs.
	// Empty string = no enforcement.
	LogDeliveryBucket string `json:"logDeliveryBucket,omitempty"`

	// EnforceHttpsOnly requires TLS for bucket access when true.
	// false (zero value) = not enforced.
	EnforceHttpsOnly bool `json:"enforceHttpsOnly,omitempty"`

	// ObjectLockMode is the enforced object lock mode.
	// Empty string = no enforcement.
	ObjectLockMode string `json:"objectLockMode,omitempty"`

	// ObjectLockRetentionDays is the enforced retention period in days.
	// 0 (zero value) = no enforcement.
	ObjectLockRetentionDays int64 `json:"objectLockRetentionDays,omitempty"`

	// Tags are tier-level cloud resource tags.
	// The reconciler populates this from KropathConfig.spec.mandatory.tags / .defaults.tags
	// so that tag union merge flows through MergeS3Cascade alongside the S3-specific fields.
	// nil / empty map = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// S3ConfigSection holds the S3 governance fields from S3Config.spec.mandatory
// or S3Config.spec.defaults (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type S3ConfigSection struct {
	// EncryptionAlgorithm is the bucket encryption mode. Empty string = not enforced.
	EncryptionAlgorithm string `json:"encryptionAlgorithm,omitempty"`

	// KmsKeyArn is the KMS key ARN. Empty string = not enforced.
	KmsKeyArn string `json:"kmsKeyArn,omitempty"`

	// BlockPublicAccess forces S3 public access blocking when true. false = not enforced.
	BlockPublicAccess bool `json:"blockPublicAccess,omitempty"`

	// Versioning is the enforced bucket versioning state. Empty string = not enforced.
	Versioning string `json:"versioning,omitempty"`

	// LoggingEnabled enforces server access logging when true. false = not enforced.
	LoggingEnabled bool `json:"loggingEnabled,omitempty"`

	// LogDeliveryBucket is the target bucket for server access logs. Empty string = not enforced.
	LogDeliveryBucket string `json:"logDeliveryBucket,omitempty"`

	// LoggingTargetPrefix is the target prefix for server access log objects.
	// Supports {namespace}/{name}/{account_id}/{region}/{configRef} tokens.
	// Empty string = not enforced; governed at S3Config levels only (not KropathConfig).
	LoggingTargetPrefix string `json:"loggingTargetPrefix,omitempty"`

	// EnforceHttpsOnly requires TLS for bucket access when true. false = not enforced.
	EnforceHttpsOnly bool `json:"enforceHttpsOnly,omitempty"`

	// ObjectLockMode is the enforced object lock mode. Empty string = not enforced.
	ObjectLockMode string `json:"objectLockMode,omitempty"`

	// ObjectLockRetentionDays is the enforced retention period in days. 0 = not enforced.
	ObjectLockRetentionDays int64 `json:"objectLockRetentionDays,omitempty"`

	// NamingTemplate is the bucket naming template (e.g. "{namespace}-{name}").
	// Governed only at S3Config levels 3-4 (mandatory) and 6-7 (defaults).
	// KropathConfig.s3 does NOT carry namingTemplate.
	// Empty string = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created bucket resources.
	// Additive map merge across S3Config tiers only.
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to created bucket resources.
	// Additive map merge across S3Config tiers only.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`

	// Tags are cloud resource tags for this S3 config profile.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// EffectiveS3Section is one tier (mandatory or defaults) of the merged S3 governance
// result written into S3Config.status.effectiveConfig by the controller.
type EffectiveS3Section struct {
	EncryptionAlgorithm     string            `json:"encryptionAlgorithm,omitempty"`
	KmsKeyArn               string            `json:"kmsKeyArn,omitempty"`
	BlockPublicAccess        bool              `json:"blockPublicAccess,omitempty"`
	Versioning              string            `json:"versioning,omitempty"`
	LoggingEnabled          bool              `json:"loggingEnabled,omitempty"`
	LogDeliveryBucket       string            `json:"logDeliveryBucket,omitempty"`
	LoggingTargetPrefix     string            `json:"loggingTargetPrefix,omitempty"`
	EnforceHttpsOnly        bool              `json:"enforceHttpsOnly,omitempty"`
	ObjectLockMode          string            `json:"objectLockMode,omitempty"`
	ObjectLockRetentionDays int64             `json:"objectLockRetentionDays,omitempty"`
	NamingTemplate          string            `json:"namingTemplate,omitempty"`
	SyncedLabels            map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations       map[string]string `json:"syncedAnnotations,omitempty"`
	Tags                    map[string]string `json:"tags,omitempty"`
}

// EffectiveS3Config is the merged S3 governance result written into
// S3Config.status.effectiveConfig by the controller.
type EffectiveS3Config struct {
	Mandatory EffectiveS3Section `json:"mandatory"`
	Defaults  EffectiveS3Section `json:"defaults"`
}

// MergeS3Cascade merges S3 governance fields from all cascade sources and
// returns the effective configuration to be written to status.effectiveConfig.
//
// The ten-level priority chain (ADR-015 §5.3) for S3 fields:
//
//	Level 1 — globalKropathMandatory  (KropathConfig in global namespace)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace)
//	Level 3 — globalS3CfgMandatory    (S3Config in global namespace)
//	Level 4 — localS3CfgMandatory     (S3Config in resource namespace)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localS3CfgDefaults      (S3Config in resource namespace)
//	Level 7 — globalS3CfgDefaults     (S3Config in global namespace)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace)
//	Level 9 — globalKropathDefaults   (KropathConfig in global namespace)
//
// For mandatory (levels 1–4): first non-zero value in priority order wins.
// For defaults (levels 6–9): first non-zero value in priority order wins.
// Tags: union merge across all four levels per tier (additive).
// NamingTemplate, SyncedLabels, SyncedAnnotations: S3Config levels only.
func MergeS3Cascade(
	// KropathConfig mandatory inputs (levels 1-2)
	globalKropathMandatory S3Section, // level 1
	localKropathMandatory S3Section, // level 2
	// S3Config mandatory inputs (levels 3-4)
	globalS3CfgMandatory S3ConfigSection, // level 3
	localS3CfgMandatory S3ConfigSection, // level 4
	// S3Config defaults inputs (levels 6-7)
	localS3CfgDefaults S3ConfigSection, // level 6
	globalS3CfgDefaults S3ConfigSection, // level 7
	// KropathConfig defaults inputs (levels 8-9)
	localKropathDefaults S3Section, // level 8
	globalKropathDefaults S3Section, // level 9
) EffectiveS3Config {
	return EffectiveS3Config{
		Mandatory: EffectiveS3Section{
			EncryptionAlgorithm: firstNonEmptyString(
				globalKropathMandatory.EncryptionAlgorithm,
				localKropathMandatory.EncryptionAlgorithm,
				globalS3CfgMandatory.EncryptionAlgorithm,
				localS3CfgMandatory.EncryptionAlgorithm,
			),
			KmsKeyArn: firstNonEmptyString(
				globalKropathMandatory.KmsKeyArn,
				localKropathMandatory.KmsKeyArn,
				globalS3CfgMandatory.KmsKeyArn,
				localS3CfgMandatory.KmsKeyArn,
			),
			BlockPublicAccess: firstTrue(
				globalKropathMandatory.BlockPublicAccess,
				localKropathMandatory.BlockPublicAccess,
				globalS3CfgMandatory.BlockPublicAccess,
				localS3CfgMandatory.BlockPublicAccess,
			),
			Versioning: firstNonEmptyString(
				globalKropathMandatory.Versioning,
				localKropathMandatory.Versioning,
				globalS3CfgMandatory.Versioning,
				localS3CfgMandatory.Versioning,
			),
			LoggingEnabled: firstTrue(
				globalKropathMandatory.LoggingEnabled,
				localKropathMandatory.LoggingEnabled,
				globalS3CfgMandatory.LoggingEnabled,
				localS3CfgMandatory.LoggingEnabled,
			),
			LogDeliveryBucket: firstNonEmptyString(
				globalKropathMandatory.LogDeliveryBucket,
				localKropathMandatory.LogDeliveryBucket,
				globalS3CfgMandatory.LogDeliveryBucket,
				localS3CfgMandatory.LogDeliveryBucket,
			),
			// LoggingTargetPrefix: S3Config levels only (not KropathConfig)
			LoggingTargetPrefix: firstNonEmptyString(
				globalS3CfgMandatory.LoggingTargetPrefix,
				localS3CfgMandatory.LoggingTargetPrefix,
			),
			EnforceHttpsOnly: firstTrue(
				globalKropathMandatory.EnforceHttpsOnly,
				localKropathMandatory.EnforceHttpsOnly,
				globalS3CfgMandatory.EnforceHttpsOnly,
				localS3CfgMandatory.EnforceHttpsOnly,
			),
			ObjectLockMode: firstNonEmptyString(
				globalKropathMandatory.ObjectLockMode,
				localKropathMandatory.ObjectLockMode,
				globalS3CfgMandatory.ObjectLockMode,
				localS3CfgMandatory.ObjectLockMode,
			),
			ObjectLockRetentionDays: firstNonZeroInt64(
				globalKropathMandatory.ObjectLockRetentionDays,
				localKropathMandatory.ObjectLockRetentionDays,
				globalS3CfgMandatory.ObjectLockRetentionDays,
				localS3CfgMandatory.ObjectLockRetentionDays,
			),
			// NamingTemplate: S3Config levels only (globalS3Cfg wins for mandatory tier)
			NamingTemplate: firstNonEmptyString(
				globalS3CfgMandatory.NamingTemplate,
				localS3CfgMandatory.NamingTemplate,
			),
			// SyncedLabels: additive merge across S3Config mandatory levels
			SyncedLabels: mergeMaps(
				localS3CfgMandatory.SyncedLabels,
				globalS3CfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: additive merge across S3Config mandatory levels
			SyncedAnnotations: mergeMaps(
				localS3CfgMandatory.SyncedAnnotations,
				globalS3CfgMandatory.SyncedAnnotations,
			),
			// Tags: union merge across all four mandatory levels (additive)
			Tags: mergeMaps(
				localS3CfgMandatory.Tags,
				globalS3CfgMandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
		},
		Defaults: EffectiveS3Section{
			EncryptionAlgorithm: firstNonEmptyString(
				localS3CfgDefaults.EncryptionAlgorithm,
				globalS3CfgDefaults.EncryptionAlgorithm,
				localKropathDefaults.EncryptionAlgorithm,
				globalKropathDefaults.EncryptionAlgorithm,
			),
			KmsKeyArn: firstNonEmptyString(
				localS3CfgDefaults.KmsKeyArn,
				globalS3CfgDefaults.KmsKeyArn,
				localKropathDefaults.KmsKeyArn,
				globalKropathDefaults.KmsKeyArn,
			),
			BlockPublicAccess: firstTrue(
				localS3CfgDefaults.BlockPublicAccess,
				globalS3CfgDefaults.BlockPublicAccess,
				localKropathDefaults.BlockPublicAccess,
				globalKropathDefaults.BlockPublicAccess,
			),
			Versioning: firstNonEmptyString(
				localS3CfgDefaults.Versioning,
				globalS3CfgDefaults.Versioning,
				localKropathDefaults.Versioning,
				globalKropathDefaults.Versioning,
			),
			LoggingEnabled: firstTrue(
				localS3CfgDefaults.LoggingEnabled,
				globalS3CfgDefaults.LoggingEnabled,
				localKropathDefaults.LoggingEnabled,
				globalKropathDefaults.LoggingEnabled,
			),
			LogDeliveryBucket: firstNonEmptyString(
				localS3CfgDefaults.LogDeliveryBucket,
				globalS3CfgDefaults.LogDeliveryBucket,
				localKropathDefaults.LogDeliveryBucket,
				globalKropathDefaults.LogDeliveryBucket,
			),
			// LoggingTargetPrefix: S3Config levels only (not KropathConfig)
			LoggingTargetPrefix: firstNonEmptyString(
				localS3CfgDefaults.LoggingTargetPrefix,
				globalS3CfgDefaults.LoggingTargetPrefix,
			),
			EnforceHttpsOnly: firstTrue(
				localS3CfgDefaults.EnforceHttpsOnly,
				globalS3CfgDefaults.EnforceHttpsOnly,
				localKropathDefaults.EnforceHttpsOnly,
				globalKropathDefaults.EnforceHttpsOnly,
			),
			ObjectLockMode: firstNonEmptyString(
				localS3CfgDefaults.ObjectLockMode,
				globalS3CfgDefaults.ObjectLockMode,
				localKropathDefaults.ObjectLockMode,
				globalKropathDefaults.ObjectLockMode,
			),
			ObjectLockRetentionDays: firstNonZeroInt64(
				localS3CfgDefaults.ObjectLockRetentionDays,
				globalS3CfgDefaults.ObjectLockRetentionDays,
				localKropathDefaults.ObjectLockRetentionDays,
				globalKropathDefaults.ObjectLockRetentionDays,
			),
			// NamingTemplate: S3Config levels only (localS3Cfg wins for defaults tier)
			NamingTemplate: firstNonEmptyString(
				localS3CfgDefaults.NamingTemplate,
				globalS3CfgDefaults.NamingTemplate,
			),
			// SyncedLabels: additive merge across S3Config defaults levels
			SyncedLabels: mergeMaps(
				globalS3CfgDefaults.SyncedLabels,
				localS3CfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: additive merge across S3Config defaults levels
			SyncedAnnotations: mergeMaps(
				globalS3CfgDefaults.SyncedAnnotations,
				localS3CfgDefaults.SyncedAnnotations,
			),
			// Tags: union merge across all four defaults levels (additive)
			Tags: mergeMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalS3CfgDefaults.Tags,
				localS3CfgDefaults.Tags,
			),
		},
	}
}
