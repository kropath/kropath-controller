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

// S3Section holds the S3-specific governance fields shared by
// AWSKropathConfig.spec.mandatory.s3 / .defaults.s3 and
// AWSS3Config.spec.mandatory / .spec.defaults.
//
// Zero value of each field is the permissive sentinel (not enforced).
type S3Section struct {
	// EncryptionAlgorithm is the bucket encryption algorithm.
	// Empty string = no enforcement.
	EncryptionAlgorithm string `json:"encryptionAlgorithm,omitempty"`

	// KmsKeyArn is the ARN of the KMS key to use for SSE-KMS.
	// Empty string = no enforcement.
	KmsKeyArn string `json:"kmsKeyArn,omitempty"`

	// BlockPublicAccess blocks public access when true.
	// false (zero value) = not enforced.
	BlockPublicAccess bool `json:"blockPublicAccess,omitempty"`

	// Versioning controls the bucket versioning state.
	// Empty string = no enforcement.
	Versioning string `json:"versioning,omitempty"`

	// LoggingEnabled requires server access logging when true.
	// false (zero value) = not enforced.
	LoggingEnabled bool `json:"loggingEnabled,omitempty"`

	// LogDeliveryBucket is the target bucket for server access logs.
	// Empty string = not enforced.
	LogDeliveryBucket string `json:"logDeliveryBucket,omitempty"`

	// EnforceHttpsOnly denies non-TLS access when true.
	// false (zero value) = not enforced.
	EnforceHttpsOnly bool `json:"enforceHttpsOnly,omitempty"`

	// ObjectLockMode controls S3 Object Lock governance mode.
	// Empty string = not enforced.
	ObjectLockMode string `json:"objectLockMode,omitempty"`

	// ObjectLockRetentionDays controls S3 Object Lock retention.
	// 0 = not enforced.
	ObjectLockRetentionDays int64 `json:"objectLockRetentionDays,omitempty"`
}

// EffectiveS3Config is the merged S3 governance result written into
// AWSS3Config.status.effectiveConfig by the controller.
type EffectiveS3Config struct {
	Mandatory S3Section `json:"mandatory"`
	Defaults  S3Section `json:"defaults"`
}

// MergeS3Cascade merges S3 governance fields from all cascade sources and
// returns the effective configuration to be written to status.effectiveConfig.
//
// The ten-level priority chain (ADR-015 §5.3) for S3 fields:
//
//	Level 1 — globalKropathMandatory  (AWSKropathConfig in kro-system)
//	Level 2 — localKropathMandatory   (AWSKropathConfig in resource namespace)
//	Level 3 — globalS3CfgMandatory    (AWSS3Config in kro-system)
//	Level 4 — localS3CfgMandatory     (AWSS3Config in resource namespace)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localS3CfgDefaults      (AWSS3Config in resource namespace)
//	Level 7 — globalS3CfgDefaults     (AWSS3Config in kro-system)
//	Level 8 — localKropathDefaults    (AWSKropathConfig in resource namespace)
//	Level 9 — globalKropathDefaults   (AWSKropathConfig in kro-system)
//
// For mandatory (levels 1–4): first non-zero value in priority order wins.
// For defaults (levels 6–9): first non-zero value in priority order wins.
// A source that is absent (zero-value struct) is silently skipped.
func MergeS3Cascade(
	globalKropathMandatory S3Section, // level 1
	localKropathMandatory S3Section, // level 2
	globalS3CfgMandatory S3Section, // level 3
	localS3CfgMandatory S3Section, // level 4
	localS3CfgDefaults S3Section, // level 6
	globalS3CfgDefaults S3Section, // level 7
	localKropathDefaults S3Section, // level 8
	globalKropathDefaults S3Section, // level 9
) EffectiveS3Config {
	return EffectiveS3Config{
		Mandatory: S3Section{
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
		},
		Defaults: S3Section{
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
		},
	}
}
