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

// S3Section holds the S3-specific governance fields shared by AWSKropathConfig
// and AWSS3Config.
//
// Zero value of each field is the permissive sentinel (not enforced).
type S3Section struct {
	EncryptionAlgorithm     string            `json:"encryptionAlgorithm,omitempty"`
	KmsKeyArn               string            `json:"kmsKeyArn,omitempty"`
	BlockPublicAccess       bool              `json:"blockPublicAccess,omitempty"`
	Versioning              string            `json:"versioning,omitempty"`
	LoggingEnabled          bool              `json:"loggingEnabled,omitempty"`
	LogDeliveryBucket       string            `json:"logDeliveryBucket,omitempty"`
	EnforceHttpsOnly        bool              `json:"enforceHttpsOnly,omitempty"`
	ObjectLockMode          string            `json:"objectLockMode,omitempty"`
	ObjectLockRetentionDays int64             `json:"objectLockRetentionDays,omitempty"`
	NamingTemplate          string            `json:"namingTemplate,omitempty"`
	Tags                    map[string]string `json:"tags,omitempty"`
	SyncedLabels            map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations       map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveS3Cascade is the merged S3 governance result written into
// AWSS3Config.status.effectiveConfig by the controller.
type EffectiveS3Cascade struct {
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
//	Level 3 — globalS3Mandatory       (AWSS3Config in kro-system)
//	Level 4 — localS3Mandatory        (AWSS3Config in resource namespace)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localS3Defaults         (AWSS3Config in resource namespace)
//	Level 7 — globalS3Defaults        (AWSS3Config in kro-system)
//	Level 8 — localKropathDefaults    (AWSKropathConfig in resource namespace)
//	Level 9 — globalKropathDefaults   (AWSKropathConfig in kro-system)
//
// For mandatory/default scalar fields, the first non-zero value in priority
// order wins. For map fields, higher-priority keys overwrite lower-priority
// keys while keeping additive entries from every tier.
func MergeS3Cascade(
	globalKropathMandatory S3Section,
	localKropathMandatory S3Section,
	globalS3Mandatory S3Section,
	localS3Mandatory S3Section,
	localS3Defaults S3Section,
	globalS3Defaults S3Section,
	localKropathDefaults S3Section,
	globalKropathDefaults S3Section,
) EffectiveS3Cascade {
	return EffectiveS3Cascade{
		Mandatory: S3Section{
			EncryptionAlgorithm: firstNonEmptyString(
				globalKropathMandatory.EncryptionAlgorithm,
				localKropathMandatory.EncryptionAlgorithm,
				globalS3Mandatory.EncryptionAlgorithm,
				localS3Mandatory.EncryptionAlgorithm,
			),
			KmsKeyArn: firstNonEmptyString(
				globalKropathMandatory.KmsKeyArn,
				localKropathMandatory.KmsKeyArn,
				globalS3Mandatory.KmsKeyArn,
				localS3Mandatory.KmsKeyArn,
			),
			BlockPublicAccess: firstTrue(
				globalKropathMandatory.BlockPublicAccess,
				localKropathMandatory.BlockPublicAccess,
				globalS3Mandatory.BlockPublicAccess,
				localS3Mandatory.BlockPublicAccess,
			),
			Versioning: firstNonEmptyString(
				globalKropathMandatory.Versioning,
				localKropathMandatory.Versioning,
				globalS3Mandatory.Versioning,
				localS3Mandatory.Versioning,
			),
			LoggingEnabled: firstTrue(
				globalKropathMandatory.LoggingEnabled,
				localKropathMandatory.LoggingEnabled,
				globalS3Mandatory.LoggingEnabled,
				localS3Mandatory.LoggingEnabled,
			),
			LogDeliveryBucket: firstNonEmptyString(
				globalKropathMandatory.LogDeliveryBucket,
				localKropathMandatory.LogDeliveryBucket,
				globalS3Mandatory.LogDeliveryBucket,
				localS3Mandatory.LogDeliveryBucket,
			),
			EnforceHttpsOnly: firstTrue(
				globalKropathMandatory.EnforceHttpsOnly,
				localKropathMandatory.EnforceHttpsOnly,
				globalS3Mandatory.EnforceHttpsOnly,
				localS3Mandatory.EnforceHttpsOnly,
			),
			ObjectLockMode: firstNonEmptyString(
				globalKropathMandatory.ObjectLockMode,
				localKropathMandatory.ObjectLockMode,
				globalS3Mandatory.ObjectLockMode,
				localS3Mandatory.ObjectLockMode,
			),
			ObjectLockRetentionDays: firstNonZeroInt64(
				globalKropathMandatory.ObjectLockRetentionDays,
				localKropathMandatory.ObjectLockRetentionDays,
				globalS3Mandatory.ObjectLockRetentionDays,
				localS3Mandatory.ObjectLockRetentionDays,
			),
			NamingTemplate: firstNonEmptyString(
				globalKropathMandatory.NamingTemplate,
				localKropathMandatory.NamingTemplate,
				globalS3Mandatory.NamingTemplate,
				localS3Mandatory.NamingTemplate,
			),
			Tags: mergeStringMaps(
				localS3Mandatory.Tags,
				globalS3Mandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
			SyncedLabels: mergeStringMaps(
				localS3Mandatory.SyncedLabels,
				globalS3Mandatory.SyncedLabels,
				localKropathMandatory.SyncedLabels,
				globalKropathMandatory.SyncedLabels,
			),
			SyncedAnnotations: mergeStringMaps(
				localS3Mandatory.SyncedAnnotations,
				globalS3Mandatory.SyncedAnnotations,
				localKropathMandatory.SyncedAnnotations,
				globalKropathMandatory.SyncedAnnotations,
			),
		},
		Defaults: S3Section{
			EncryptionAlgorithm: firstNonEmptyString(
				localS3Defaults.EncryptionAlgorithm,
				globalS3Defaults.EncryptionAlgorithm,
				localKropathDefaults.EncryptionAlgorithm,
				globalKropathDefaults.EncryptionAlgorithm,
			),
			KmsKeyArn: firstNonEmptyString(
				localS3Defaults.KmsKeyArn,
				globalS3Defaults.KmsKeyArn,
				localKropathDefaults.KmsKeyArn,
				globalKropathDefaults.KmsKeyArn,
			),
			BlockPublicAccess: firstTrue(
				localS3Defaults.BlockPublicAccess,
				globalS3Defaults.BlockPublicAccess,
				localKropathDefaults.BlockPublicAccess,
				globalKropathDefaults.BlockPublicAccess,
			),
			Versioning: firstNonEmptyString(
				localS3Defaults.Versioning,
				globalS3Defaults.Versioning,
				localKropathDefaults.Versioning,
				globalKropathDefaults.Versioning,
			),
			LoggingEnabled: firstTrue(
				localS3Defaults.LoggingEnabled,
				globalS3Defaults.LoggingEnabled,
				localKropathDefaults.LoggingEnabled,
				globalKropathDefaults.LoggingEnabled,
			),
			LogDeliveryBucket: firstNonEmptyString(
				localS3Defaults.LogDeliveryBucket,
				globalS3Defaults.LogDeliveryBucket,
				localKropathDefaults.LogDeliveryBucket,
				globalKropathDefaults.LogDeliveryBucket,
			),
			EnforceHttpsOnly: firstTrue(
				localS3Defaults.EnforceHttpsOnly,
				globalS3Defaults.EnforceHttpsOnly,
				localKropathDefaults.EnforceHttpsOnly,
				globalKropathDefaults.EnforceHttpsOnly,
			),
			ObjectLockMode: firstNonEmptyString(
				localS3Defaults.ObjectLockMode,
				globalS3Defaults.ObjectLockMode,
				localKropathDefaults.ObjectLockMode,
				globalKropathDefaults.ObjectLockMode,
			),
			ObjectLockRetentionDays: firstNonZeroInt64(
				localS3Defaults.ObjectLockRetentionDays,
				globalS3Defaults.ObjectLockRetentionDays,
				localKropathDefaults.ObjectLockRetentionDays,
				globalKropathDefaults.ObjectLockRetentionDays,
			),
			NamingTemplate: firstNonEmptyString(
				localS3Defaults.NamingTemplate,
				globalS3Defaults.NamingTemplate,
				localKropathDefaults.NamingTemplate,
				globalKropathDefaults.NamingTemplate,
			),
			Tags: mergeStringMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalS3Defaults.Tags,
				localS3Defaults.Tags,
			),
			SyncedLabels: mergeStringMaps(
				globalKropathDefaults.SyncedLabels,
				localKropathDefaults.SyncedLabels,
				globalS3Defaults.SyncedLabels,
				localS3Defaults.SyncedLabels,
			),
			SyncedAnnotations: mergeStringMaps(
				globalKropathDefaults.SyncedAnnotations,
				localKropathDefaults.SyncedAnnotations,
				globalS3Defaults.SyncedAnnotations,
				localS3Defaults.SyncedAnnotations,
			),
		},
	}
}

func mergeStringMaps(maps ...map[string]string) map[string]string {
	out := make(map[string]string)
	for _, input := range maps {
		for key, value := range input {
			out[key] = value
		}
	}
	return out
}
