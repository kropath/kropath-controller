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

// AccessPointPublicAccessBlockSection holds the four S3 Control access-point
// public access block booleans. Used in both the KropathConfig and
// S3AdvancedConfig governance sections.
//
// Zero value (all false) is the permissive sentinel (not enforced).
type AccessPointPublicAccessBlockSection struct {
	// BlockPublicACLs blocks public ACLs on access points when true.
	// false (zero value) = not enforced.
	BlockPublicACLs bool `json:"blockPublicACLs,omitempty"`

	// BlockPublicPolicy blocks public bucket policies on access points when true.
	// false (zero value) = not enforced.
	BlockPublicPolicy bool `json:"blockPublicPolicy,omitempty"`

	// IgnorePublicACLs ignores public ACLs on access points when true.
	// false (zero value) = not enforced.
	IgnorePublicACLs bool `json:"ignorePublicACLs,omitempty"`

	// RestrictPublicBuckets restricts public bucket policies on access points when true.
	// false (zero value) = not enforced.
	RestrictPublicBuckets bool `json:"restrictPublicBuckets,omitempty"`
}

// S3AdvancedKropathSection holds the S3 Advanced governance fields from
// KropathConfig.spec.mandatory.s3Advanced / .defaults.s3Advanced (ADR-015 §3.5).
//
// Only access point public access block and VPC-only enforcement flow through
// KropathConfig; per-service encryption is governed at the S3AdvancedConfig level
// only (spec §KropathConfig Additions).
//
// Zero value of each field is the permissive sentinel (not enforced).
type S3AdvancedKropathSection struct {
	// AccessPointPublicAccessBlock governs public access block settings for S3
	// Control access points at the KropathConfig (org/namespace) level.
	AccessPointPublicAccessBlock AccessPointPublicAccessBlockSection `json:"accessPointPublicAccessBlock,omitempty"`

	// AccessPointVpcOnly enforces VPC-only access for all S3 Control access points
	// when true. false (zero value) = not enforced.
	AccessPointVpcOnly bool `json:"accessPointVpcOnly,omitempty"`
}

// TableBucketEncryptionSection holds S3 Tables encryption governance fields.
type TableBucketEncryptionSection struct {
	// SseAlgorithm is the enforced server-side encryption algorithm.
	// Empty string = not enforced.
	SseAlgorithm string `json:"sseAlgorithm,omitempty"`

	// KmsKeyARN is the enforced KMS key ARN for S3 Tables encryption.
	// Empty string = not enforced.
	KmsKeyARN string `json:"kmsKeyARN,omitempty"`
}

// VectorBucketEncryptionSection holds S3 Vectors vector bucket encryption fields.
type VectorBucketEncryptionSection struct {
	// SseType is the enforced server-side encryption type for vector buckets.
	// Empty string = not enforced.
	SseType string `json:"sseType,omitempty"`

	// KmsKeyARN is the enforced KMS key ARN for vector bucket encryption.
	// Empty string = not enforced.
	KmsKeyARN string `json:"kmsKeyARN,omitempty"`
}

// VectorIndexEncryptionSection holds S3 Vectors index encryption fields.
type VectorIndexEncryptionSection struct {
	// SseType is the enforced server-side encryption type for vector indexes.
	// Empty string = not enforced.
	SseType string `json:"sseType,omitempty"`

	// KmsKeyARN is the enforced KMS key ARN for vector index encryption.
	// Empty string = not enforced.
	KmsKeyARN string `json:"kmsKeyARN,omitempty"`
}

// FilesEncryptionSection holds S3 Files encryption governance fields.
type FilesEncryptionSection struct {
	// KmsKeyID is the enforced KMS key ID for S3 Files encryption.
	// Empty string = not enforced.
	KmsKeyID string `json:"kmsKeyID,omitempty"`
}

// S3AdvancedConfigSection holds the S3 Advanced governance fields from
// S3AdvancedConfig.spec.mandatory or S3AdvancedConfig.spec.defaults.
//
// All 8 field groups (per-service encryption, access point governance, and
// common fields) appear in both mandatory and defaults tiers (tier symmetry).
//
// Zero value of each field is the permissive sentinel (not enforced).
type S3AdvancedConfigSection struct {
	// TableBucketEncryption governs S3 Tables table bucket SSE settings.
	TableBucketEncryption TableBucketEncryptionSection `json:"tableBucketEncryption,omitempty"`

	// VectorBucketEncryption governs S3 Vectors vector bucket SSE settings.
	VectorBucketEncryption VectorBucketEncryptionSection `json:"vectorBucketEncryption,omitempty"`

	// VectorIndexEncryption governs S3 Vectors index SSE settings.
	VectorIndexEncryption VectorIndexEncryptionSection `json:"vectorIndexEncryption,omitempty"`

	// FilesEncryption governs S3 Files KMS encryption settings.
	FilesEncryption FilesEncryptionSection `json:"filesEncryption,omitempty"`

	// AccessPointPublicAccessBlock governs public access block settings for S3
	// Control access points.
	AccessPointPublicAccessBlock AccessPointPublicAccessBlockSection `json:"accessPointPublicAccessBlock,omitempty"`

	// AccessPointVpcOnly enforces VPC-only access for all S3 Control access points
	// when true. false (zero value) = not enforced.
	AccessPointVpcOnly bool `json:"accessPointVpcOnly,omitempty"`

	// Tags are cloud resource tags for this S3 Advanced config profile.
	// nil / empty = no tags at this level. Tags use additive union merge.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are Kubernetes labels propagated to created resources.
	// Additive map merge across S3AdvancedConfig tiers only.
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations propagated to created resources.
	// Additive map merge across S3AdvancedConfig tiers only.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`

	// NamingTemplate is the cloud resource naming template
	// (e.g. "{namespace}-{name}"). Governed only at S3AdvancedConfig levels;
	// KropathConfig.s3Advanced does NOT carry namingTemplate.
	// Empty string = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`
}

// EffectiveS3AdvancedSection is one tier (mandatory or defaults) of the merged
// S3 Advanced governance result written into
// S3AdvancedConfig.status.effectiveConfig by the controller.
type EffectiveS3AdvancedSection struct {
	TableBucketEncryption        TableBucketEncryptionSection         `json:"tableBucketEncryption,omitempty"`
	VectorBucketEncryption       VectorBucketEncryptionSection        `json:"vectorBucketEncryption,omitempty"`
	VectorIndexEncryption        VectorIndexEncryptionSection         `json:"vectorIndexEncryption,omitempty"`
	FilesEncryption              FilesEncryptionSection               `json:"filesEncryption,omitempty"`
	AccessPointPublicAccessBlock AccessPointPublicAccessBlockSection  `json:"accessPointPublicAccessBlock,omitempty"`
	AccessPointVpcOnly           bool                                 `json:"accessPointVpcOnly,omitempty"`
	Tags                         map[string]string                    `json:"tags,omitempty"`
	SyncedLabels                 map[string]string                    `json:"syncedLabels,omitempty"`
	SyncedAnnotations            map[string]string                    `json:"syncedAnnotations,omitempty"`
	NamingTemplate               string                               `json:"namingTemplate,omitempty"`
}

// EffectiveS3AdvancedConfig is the merged S3 Advanced governance result written
// into S3AdvancedConfig.status.effectiveConfig by the controller.
type EffectiveS3AdvancedConfig struct {
	Mandatory EffectiveS3AdvancedSection `json:"mandatory"`
	Defaults  EffectiveS3AdvancedSection `json:"defaults"`
}

// MergeS3AdvancedCascade merges S3 Advanced governance fields from all cascade
// sources and returns the effective configuration to be written to
// status.effectiveConfig.
//
// The ten-level priority chain (ADR-015 §5.3) for S3 Advanced fields:
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace)
//	Level 3 — globalS3AdvCfgMandatory (S3AdvancedConfig in kro-system)
//	Level 4 — localS3AdvCfgMandatory  (S3AdvancedConfig in resource namespace)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localS3AdvCfgDefaults   (S3AdvancedConfig in resource namespace)
//	Level 7 — globalS3AdvCfgDefaults  (S3AdvancedConfig in kro-system)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system)
//
// For mandatory (levels 1–4): first non-zero value in priority order wins.
// For defaults (levels 6–9): first non-zero value in priority order wins.
//
// Field scope:
//   - accessPointPublicAccessBlock, accessPointVpcOnly: full 8-level cascade
//     (KropathConfig levels 1-2/8-9 + S3AdvancedConfig levels 3-4/6-7)
//   - tableBucketEncryption, vectorBucketEncryption, vectorIndexEncryption,
//     filesEncryption: S3AdvancedConfig levels only (3-4/6-7); no KropathConfig
//   - tags: additive union merge across S3AdvancedConfig levels only (no KropathConfig)
//   - namingTemplate, syncedLabels, syncedAnnotations: S3AdvancedConfig only
func MergeS3AdvancedCascade(
	// KropathConfig mandatory inputs (levels 1-2)
	globalKropathMandatory S3AdvancedKropathSection, // level 1
	localKropathMandatory S3AdvancedKropathSection, // level 2
	// S3AdvancedConfig mandatory inputs (levels 3-4)
	globalS3AdvCfgMandatory S3AdvancedConfigSection, // level 3
	localS3AdvCfgMandatory S3AdvancedConfigSection, // level 4
	// S3AdvancedConfig defaults inputs (levels 6-7)
	localS3AdvCfgDefaults S3AdvancedConfigSection, // level 6
	globalS3AdvCfgDefaults S3AdvancedConfigSection, // level 7
	// KropathConfig defaults inputs (levels 8-9)
	localKropathDefaults S3AdvancedKropathSection, // level 8
	globalKropathDefaults S3AdvancedKropathSection, // level 9
) EffectiveS3AdvancedConfig {
	return EffectiveS3AdvancedConfig{
		Mandatory: EffectiveS3AdvancedSection{
			// Per-service encryption: S3AdvancedConfig levels only (3-4)
			TableBucketEncryption: TableBucketEncryptionSection{
				SseAlgorithm: firstNonEmptyString(
					globalS3AdvCfgMandatory.TableBucketEncryption.SseAlgorithm,
					localS3AdvCfgMandatory.TableBucketEncryption.SseAlgorithm,
				),
				KmsKeyARN: firstNonEmptyString(
					globalS3AdvCfgMandatory.TableBucketEncryption.KmsKeyARN,
					localS3AdvCfgMandatory.TableBucketEncryption.KmsKeyARN,
				),
			},
			VectorBucketEncryption: VectorBucketEncryptionSection{
				SseType: firstNonEmptyString(
					globalS3AdvCfgMandatory.VectorBucketEncryption.SseType,
					localS3AdvCfgMandatory.VectorBucketEncryption.SseType,
				),
				KmsKeyARN: firstNonEmptyString(
					globalS3AdvCfgMandatory.VectorBucketEncryption.KmsKeyARN,
					localS3AdvCfgMandatory.VectorBucketEncryption.KmsKeyARN,
				),
			},
			VectorIndexEncryption: VectorIndexEncryptionSection{
				SseType: firstNonEmptyString(
					globalS3AdvCfgMandatory.VectorIndexEncryption.SseType,
					localS3AdvCfgMandatory.VectorIndexEncryption.SseType,
				),
				KmsKeyARN: firstNonEmptyString(
					globalS3AdvCfgMandatory.VectorIndexEncryption.KmsKeyARN,
					localS3AdvCfgMandatory.VectorIndexEncryption.KmsKeyARN,
				),
			},
			FilesEncryption: FilesEncryptionSection{
				KmsKeyID: firstNonEmptyString(
					globalS3AdvCfgMandatory.FilesEncryption.KmsKeyID,
					localS3AdvCfgMandatory.FilesEncryption.KmsKeyID,
				),
			},
			// Access point governance: full 8-level cascade (KropathConfig + S3AdvancedConfig)
			AccessPointPublicAccessBlock: AccessPointPublicAccessBlockSection{
				BlockPublicACLs: firstTrue(
					globalKropathMandatory.AccessPointPublicAccessBlock.BlockPublicACLs,
					localKropathMandatory.AccessPointPublicAccessBlock.BlockPublicACLs,
					globalS3AdvCfgMandatory.AccessPointPublicAccessBlock.BlockPublicACLs,
					localS3AdvCfgMandatory.AccessPointPublicAccessBlock.BlockPublicACLs,
				),
				BlockPublicPolicy: firstTrue(
					globalKropathMandatory.AccessPointPublicAccessBlock.BlockPublicPolicy,
					localKropathMandatory.AccessPointPublicAccessBlock.BlockPublicPolicy,
					globalS3AdvCfgMandatory.AccessPointPublicAccessBlock.BlockPublicPolicy,
					localS3AdvCfgMandatory.AccessPointPublicAccessBlock.BlockPublicPolicy,
				),
				IgnorePublicACLs: firstTrue(
					globalKropathMandatory.AccessPointPublicAccessBlock.IgnorePublicACLs,
					localKropathMandatory.AccessPointPublicAccessBlock.IgnorePublicACLs,
					globalS3AdvCfgMandatory.AccessPointPublicAccessBlock.IgnorePublicACLs,
					localS3AdvCfgMandatory.AccessPointPublicAccessBlock.IgnorePublicACLs,
				),
				RestrictPublicBuckets: firstTrue(
					globalKropathMandatory.AccessPointPublicAccessBlock.RestrictPublicBuckets,
					localKropathMandatory.AccessPointPublicAccessBlock.RestrictPublicBuckets,
					globalS3AdvCfgMandatory.AccessPointPublicAccessBlock.RestrictPublicBuckets,
					localS3AdvCfgMandatory.AccessPointPublicAccessBlock.RestrictPublicBuckets,
				),
			},
			AccessPointVpcOnly: firstTrue(
				globalKropathMandatory.AccessPointVpcOnly,
				localKropathMandatory.AccessPointVpcOnly,
				globalS3AdvCfgMandatory.AccessPointVpcOnly,
				localS3AdvCfgMandatory.AccessPointVpcOnly,
			),
			// Tags: additive union merge across S3AdvancedConfig mandatory levels only
			// (no KropathConfig wrapper for this family — spec §Key notes)
			Tags: mergeMaps(
				localS3AdvCfgMandatory.Tags,
				globalS3AdvCfgMandatory.Tags,
			),
			// SyncedLabels: additive merge across S3AdvancedConfig mandatory levels
			SyncedLabels: mergeMaps(
				localS3AdvCfgMandatory.SyncedLabels,
				globalS3AdvCfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: additive merge across S3AdvancedConfig mandatory levels
			SyncedAnnotations: mergeMaps(
				localS3AdvCfgMandatory.SyncedAnnotations,
				globalS3AdvCfgMandatory.SyncedAnnotations,
			),
			// NamingTemplate: S3AdvancedConfig mandatory levels only
			// (globalS3AdvCfg wins for mandatory tier; no KropathConfig)
			NamingTemplate: firstNonEmptyString(
				globalS3AdvCfgMandatory.NamingTemplate,
				localS3AdvCfgMandatory.NamingTemplate,
			),
		},
		Defaults: EffectiveS3AdvancedSection{
			// Per-service encryption: S3AdvancedConfig levels only (6-7)
			TableBucketEncryption: TableBucketEncryptionSection{
				SseAlgorithm: firstNonEmptyString(
					localS3AdvCfgDefaults.TableBucketEncryption.SseAlgorithm,
					globalS3AdvCfgDefaults.TableBucketEncryption.SseAlgorithm,
				),
				KmsKeyARN: firstNonEmptyString(
					localS3AdvCfgDefaults.TableBucketEncryption.KmsKeyARN,
					globalS3AdvCfgDefaults.TableBucketEncryption.KmsKeyARN,
				),
			},
			VectorBucketEncryption: VectorBucketEncryptionSection{
				SseType: firstNonEmptyString(
					localS3AdvCfgDefaults.VectorBucketEncryption.SseType,
					globalS3AdvCfgDefaults.VectorBucketEncryption.SseType,
				),
				KmsKeyARN: firstNonEmptyString(
					localS3AdvCfgDefaults.VectorBucketEncryption.KmsKeyARN,
					globalS3AdvCfgDefaults.VectorBucketEncryption.KmsKeyARN,
				),
			},
			VectorIndexEncryption: VectorIndexEncryptionSection{
				SseType: firstNonEmptyString(
					localS3AdvCfgDefaults.VectorIndexEncryption.SseType,
					globalS3AdvCfgDefaults.VectorIndexEncryption.SseType,
				),
				KmsKeyARN: firstNonEmptyString(
					localS3AdvCfgDefaults.VectorIndexEncryption.KmsKeyARN,
					globalS3AdvCfgDefaults.VectorIndexEncryption.KmsKeyARN,
				),
			},
			FilesEncryption: FilesEncryptionSection{
				KmsKeyID: firstNonEmptyString(
					localS3AdvCfgDefaults.FilesEncryption.KmsKeyID,
					globalS3AdvCfgDefaults.FilesEncryption.KmsKeyID,
				),
			},
			// Access point governance: full 8-level cascade (S3AdvancedConfig + KropathConfig)
			AccessPointPublicAccessBlock: AccessPointPublicAccessBlockSection{
				BlockPublicACLs: firstTrue(
					localS3AdvCfgDefaults.AccessPointPublicAccessBlock.BlockPublicACLs,
					globalS3AdvCfgDefaults.AccessPointPublicAccessBlock.BlockPublicACLs,
					localKropathDefaults.AccessPointPublicAccessBlock.BlockPublicACLs,
					globalKropathDefaults.AccessPointPublicAccessBlock.BlockPublicACLs,
				),
				BlockPublicPolicy: firstTrue(
					localS3AdvCfgDefaults.AccessPointPublicAccessBlock.BlockPublicPolicy,
					globalS3AdvCfgDefaults.AccessPointPublicAccessBlock.BlockPublicPolicy,
					localKropathDefaults.AccessPointPublicAccessBlock.BlockPublicPolicy,
					globalKropathDefaults.AccessPointPublicAccessBlock.BlockPublicPolicy,
				),
				IgnorePublicACLs: firstTrue(
					localS3AdvCfgDefaults.AccessPointPublicAccessBlock.IgnorePublicACLs,
					globalS3AdvCfgDefaults.AccessPointPublicAccessBlock.IgnorePublicACLs,
					localKropathDefaults.AccessPointPublicAccessBlock.IgnorePublicACLs,
					globalKropathDefaults.AccessPointPublicAccessBlock.IgnorePublicACLs,
				),
				RestrictPublicBuckets: firstTrue(
					localS3AdvCfgDefaults.AccessPointPublicAccessBlock.RestrictPublicBuckets,
					globalS3AdvCfgDefaults.AccessPointPublicAccessBlock.RestrictPublicBuckets,
					localKropathDefaults.AccessPointPublicAccessBlock.RestrictPublicBuckets,
					globalKropathDefaults.AccessPointPublicAccessBlock.RestrictPublicBuckets,
				),
			},
			AccessPointVpcOnly: firstTrue(
				localS3AdvCfgDefaults.AccessPointVpcOnly,
				globalS3AdvCfgDefaults.AccessPointVpcOnly,
				localKropathDefaults.AccessPointVpcOnly,
				globalKropathDefaults.AccessPointVpcOnly,
			),
			// Tags: additive union merge across S3AdvancedConfig defaults levels only
			Tags: mergeMaps(
				globalS3AdvCfgDefaults.Tags,
				localS3AdvCfgDefaults.Tags,
			),
			// SyncedLabels: additive merge across S3AdvancedConfig defaults levels
			SyncedLabels: mergeMaps(
				globalS3AdvCfgDefaults.SyncedLabels,
				localS3AdvCfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: additive merge across S3AdvancedConfig defaults levels
			SyncedAnnotations: mergeMaps(
				globalS3AdvCfgDefaults.SyncedAnnotations,
				localS3AdvCfgDefaults.SyncedAnnotations,
			),
			// NamingTemplate: S3AdvancedConfig defaults levels only
			// (localS3AdvCfg wins for defaults tier; no KropathConfig)
			NamingTemplate: firstNonEmptyString(
				localS3AdvCfgDefaults.NamingTemplate,
				globalS3AdvCfgDefaults.NamingTemplate,
			),
		},
	}
}
