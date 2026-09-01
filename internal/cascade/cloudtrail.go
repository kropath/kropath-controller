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

// CloudTrailKropathSection holds the CloudTrail-family governance fields from
// KropathConfig.spec.mandatory.cloudtrail / .defaults.cloudtrail (ADR-015 §3.5).
//
// Only 3 scalar fields are governed at the KropathConfig level: isMultiRegionTrail,
// enableLogFileValidation, and retentionPeriod. All other CloudTrailConfig fields
// are per-trail/per-EDS choices (family design §8).
//
// Zero value of each field is the permissive sentinel (not enforced).
type CloudTrailKropathSection struct {
	// IsMultiRegionTrail enforces org-wide multi-region trail setting.
	// false (zero value) = not enforced; true = all trails must be multi-region.
	IsMultiRegionTrail bool `json:"isMultiRegionTrail,omitempty"`

	// EnableLogFileValidation enforces log file integrity validation org-wide.
	// false (zero value) = not enforced; true = log file validation required.
	EnableLogFileValidation bool `json:"enableLogFileValidation,omitempty"`

	// RetentionPeriod is the org-wide mandatory event data store retention in days.
	// 0 (zero value) = not enforced; first-non-zero-wins in cascade.
	RetentionPeriod int64 `json:"retentionPeriod,omitempty"`

	// Tags are tier-level cloud resource tags.
	// The reconciler populates this from KropathConfig.spec.mandatory.tags / .defaults.tags
	// so that tag union merge flows through MergeCloudTrailCascade alongside the
	// CloudTrail-specific fields.
	// nil / empty map = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// CloudTrailConfigSection holds the CloudTrail governance fields from
// CloudTrailConfig.spec.mandatory or CloudTrailConfig.spec.defaults
// (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type CloudTrailConfigSection struct {
	// IsMultiRegionTrail enforces multi-region trail setting for this profile.
	// false (zero value) = not enforced; true = multi-region required.
	IsMultiRegionTrail bool `json:"isMultiRegionTrail,omitempty"`

	// EnableLogFileValidation enforces log file integrity validation for this profile.
	// false (zero value) = not enforced; true = validation required.
	EnableLogFileValidation bool `json:"enableLogFileValidation,omitempty"`

	// IncludeGlobalServiceEvents enforces inclusion of global service events for this profile.
	// false (zero value) = not enforced; true = include global service events.
	IncludeGlobalServiceEvents bool `json:"includeGlobalServiceEvents,omitempty"`

	// IsOrganizationTrail enforces org-level trail setting for this profile.
	// false (zero value) = not enforced; true = organization trail required.
	IsOrganizationTrail bool `json:"isOrganizationTrail,omitempty"`

	// S3BucketName is the S3 bucket name for trail log delivery.
	// Empty string = not enforced.
	S3BucketName string `json:"s3BucketName,omitempty"`

	// S3KeyPrefix is the S3 key prefix for trail log delivery.
	// Empty string = not enforced.
	S3KeyPrefix string `json:"s3KeyPrefix,omitempty"`

	// KmsKeyID is the KMS key ARN for CloudTrail log encryption.
	// Empty string = not enforced.
	KmsKeyID string `json:"kmsKeyID,omitempty"`

	// MultiRegionEnabled enforces multi-region event data store for this profile.
	// false (zero value) = not enforced; true = multi-region EDS required.
	MultiRegionEnabled bool `json:"multiRegionEnabled,omitempty"`

	// OrganizationEnabled enforces org-level event data store for this profile.
	// false (zero value) = not enforced; true = organization EDS required.
	OrganizationEnabled bool `json:"organizationEnabled,omitempty"`

	// RetentionPeriod is the event data store retention period in days.
	// 0 (zero value) = not enforced; first-non-zero-wins in cascade.
	RetentionPeriod int64 `json:"retentionPeriod,omitempty"`

	// TerminationProtectionEnabled enforces deletion protection for event data stores.
	// false (zero value) = not enforced; true = termination protection required.
	TerminationProtectionEnabled bool `json:"terminationProtectionEnabled,omitempty"`

	// NamingTemplate is the CloudTrail resource naming template (e.g. "{namespace}-{name}").
	// Governed only at CloudTrailConfig levels 3-4 (mandatory) and 6-7 (defaults).
	// KropathConfig.cloudtrail does NOT carry namingTemplate.
	// Empty string = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created CloudTrail resources.
	// Additive map merge across CloudTrailConfig tiers only.
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to created CloudTrail resources.
	// Additive map merge across CloudTrailConfig tiers only.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`

	// Tags are cloud resource tags for this CloudTrail config profile.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// EffectiveCloudTrailSection is one tier (mandatory or defaults) of the merged
// CloudTrail governance result written into CloudTrailConfig.status.effectiveConfig
// by the controller.
type EffectiveCloudTrailSection struct {
	IsMultiRegionTrail           bool              `json:"isMultiRegionTrail,omitempty"`
	EnableLogFileValidation      bool              `json:"enableLogFileValidation,omitempty"`
	IncludeGlobalServiceEvents   bool              `json:"includeGlobalServiceEvents,omitempty"`
	IsOrganizationTrail          bool              `json:"isOrganizationTrail,omitempty"`
	S3BucketName                 string            `json:"s3BucketName,omitempty"`
	S3KeyPrefix                  string            `json:"s3KeyPrefix,omitempty"`
	KmsKeyID                     string            `json:"kmsKeyID,omitempty"`
	MultiRegionEnabled           bool              `json:"multiRegionEnabled,omitempty"`
	OrganizationEnabled          bool              `json:"organizationEnabled,omitempty"`
	RetentionPeriod              int64             `json:"retentionPeriod,omitempty"`
	TerminationProtectionEnabled bool              `json:"terminationProtectionEnabled,omitempty"`
	NamingTemplate               string            `json:"namingTemplate,omitempty"`
	SyncedLabels                 map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations            map[string]string `json:"syncedAnnotations,omitempty"`
	Tags                         map[string]string `json:"tags,omitempty"`
}

// EffectiveCloudTrailConfig is the merged CloudTrail governance result written into
// CloudTrailConfig.status.effectiveConfig by the controller.
type EffectiveCloudTrailConfig struct {
	Mandatory EffectiveCloudTrailSection `json:"mandatory"`
	Defaults  EffectiveCloudTrailSection `json:"defaults"`
}

// MergeCloudTrailCascade merges CloudTrail governance fields from all cascade
// sources and returns the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for CloudTrail (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.cloudtrail)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.cloudtrail)
//	Level 3 — globalCTCfgMandatory    (CloudTrailConfig in kro-system, mandatory)
//	Level 4 — localCTCfgMandatory     (CloudTrailConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localCTCfgDefaults      (CloudTrailConfig in resource namespace, defaults)
//	Level 7 — globalCTCfgDefaults     (CloudTrailConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.cloudtrail)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.cloudtrail)
//
// Scalar merge: firstNonEmptyString / firstNonZeroInt64 / firstTrue in priority order
// (lowest number wins).
// Tags: additive union merge across all four mandatory levels, all four defaults levels.
// SyncedLabels/SyncedAnnotations: additive union from CloudTrailConfig levels only (no KropathConfig).
// NamingTemplate: governed only at CloudTrailConfig levels (3-4 mandatory, 6-7 defaults).
// KropathConfig fields: isMultiRegionTrail, enableLogFileValidation, retentionPeriod only.
func MergeCloudTrailCascade(
	globalKropathMandatory CloudTrailKropathSection, // level 1
	localKropathMandatory CloudTrailKropathSection, // level 2
	globalCTCfgMandatory CloudTrailConfigSection, // level 3
	localCTCfgMandatory CloudTrailConfigSection, // level 4
	localCTCfgDefaults CloudTrailConfigSection, // level 6
	globalCTCfgDefaults CloudTrailConfigSection, // level 7
	localKropathDefaults CloudTrailKropathSection, // level 8
	globalKropathDefaults CloudTrailKropathSection, // level 9
) EffectiveCloudTrailConfig {
	return EffectiveCloudTrailConfig{
		Mandatory: EffectiveCloudTrailSection{
			// isMultiRegionTrail: KropathConfig levels 1-2 then CloudTrailConfig levels 3-4.
			IsMultiRegionTrail: firstTrue(
				globalKropathMandatory.IsMultiRegionTrail,
				localKropathMandatory.IsMultiRegionTrail,
				globalCTCfgMandatory.IsMultiRegionTrail,
				localCTCfgMandatory.IsMultiRegionTrail,
			),
			// enableLogFileValidation: KropathConfig levels 1-2 then CloudTrailConfig levels 3-4.
			EnableLogFileValidation: firstTrue(
				globalKropathMandatory.EnableLogFileValidation,
				localKropathMandatory.EnableLogFileValidation,
				globalCTCfgMandatory.EnableLogFileValidation,
				localCTCfgMandatory.EnableLogFileValidation,
			),
			// includeGlobalServiceEvents: CloudTrailConfig levels only (3, 4).
			IncludeGlobalServiceEvents: firstTrue(
				globalCTCfgMandatory.IncludeGlobalServiceEvents,
				localCTCfgMandatory.IncludeGlobalServiceEvents,
			),
			// isOrganizationTrail: CloudTrailConfig levels only (3, 4).
			IsOrganizationTrail: firstTrue(
				globalCTCfgMandatory.IsOrganizationTrail,
				localCTCfgMandatory.IsOrganizationTrail,
			),
			// s3BucketName: CloudTrailConfig levels only (3, 4).
			S3BucketName: firstNonEmptyString(
				globalCTCfgMandatory.S3BucketName,
				localCTCfgMandatory.S3BucketName,
			),
			// s3KeyPrefix: CloudTrailConfig levels only (3, 4).
			S3KeyPrefix: firstNonEmptyString(
				globalCTCfgMandatory.S3KeyPrefix,
				localCTCfgMandatory.S3KeyPrefix,
			),
			// kmsKeyID: CloudTrailConfig levels only (3, 4).
			KmsKeyID: firstNonEmptyString(
				globalCTCfgMandatory.KmsKeyID,
				localCTCfgMandatory.KmsKeyID,
			),
			// multiRegionEnabled: CloudTrailConfig levels only (3, 4).
			MultiRegionEnabled: firstTrue(
				globalCTCfgMandatory.MultiRegionEnabled,
				localCTCfgMandatory.MultiRegionEnabled,
			),
			// organizationEnabled: CloudTrailConfig levels only (3, 4).
			OrganizationEnabled: firstTrue(
				globalCTCfgMandatory.OrganizationEnabled,
				localCTCfgMandatory.OrganizationEnabled,
			),
			// retentionPeriod: KropathConfig levels 1-2 then CloudTrailConfig levels 3-4.
			RetentionPeriod: firstNonZeroInt64(
				globalKropathMandatory.RetentionPeriod,
				localKropathMandatory.RetentionPeriod,
				globalCTCfgMandatory.RetentionPeriod,
				localCTCfgMandatory.RetentionPeriod,
			),
			// terminationProtectionEnabled: CloudTrailConfig levels only (3, 4).
			TerminationProtectionEnabled: firstTrue(
				globalCTCfgMandatory.TerminationProtectionEnabled,
				localCTCfgMandatory.TerminationProtectionEnabled,
			),
			// NamingTemplate: CloudTrailConfig levels only (3, 4);
			// KropathConfig has no namingTemplate field for cloudtrail.
			NamingTemplate: firstNonEmptyString(
				globalCTCfgMandatory.NamingTemplate,
				localCTCfgMandatory.NamingTemplate,
			),
			// SyncedLabels: additive union from CloudTrailConfig levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localCTCfgMandatory.SyncedLabels,
				globalCTCfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localCTCfgMandatory.SyncedAnnotations,
				globalCTCfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localCTCfgMandatory.Tags,
				globalCTCfgMandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
		},
		Defaults: EffectiveCloudTrailSection{
			// isMultiRegionTrail: CloudTrailConfig defaults levels 6-7 then KropathConfig levels 8-9.
			IsMultiRegionTrail: firstTrue(
				localCTCfgDefaults.IsMultiRegionTrail,
				globalCTCfgDefaults.IsMultiRegionTrail,
				localKropathDefaults.IsMultiRegionTrail,
				globalKropathDefaults.IsMultiRegionTrail,
			),
			// enableLogFileValidation: CloudTrailConfig defaults levels 6-7 then KropathConfig levels 8-9.
			EnableLogFileValidation: firstTrue(
				localCTCfgDefaults.EnableLogFileValidation,
				globalCTCfgDefaults.EnableLogFileValidation,
				localKropathDefaults.EnableLogFileValidation,
				globalKropathDefaults.EnableLogFileValidation,
			),
			// includeGlobalServiceEvents: CloudTrailConfig levels only (6, 7).
			IncludeGlobalServiceEvents: firstTrue(
				localCTCfgDefaults.IncludeGlobalServiceEvents,
				globalCTCfgDefaults.IncludeGlobalServiceEvents,
			),
			// isOrganizationTrail: CloudTrailConfig levels only (6, 7).
			IsOrganizationTrail: firstTrue(
				localCTCfgDefaults.IsOrganizationTrail,
				globalCTCfgDefaults.IsOrganizationTrail,
			),
			// s3BucketName: CloudTrailConfig levels only (6, 7).
			S3BucketName: firstNonEmptyString(
				localCTCfgDefaults.S3BucketName,
				globalCTCfgDefaults.S3BucketName,
			),
			// s3KeyPrefix: CloudTrailConfig levels only (6, 7).
			S3KeyPrefix: firstNonEmptyString(
				localCTCfgDefaults.S3KeyPrefix,
				globalCTCfgDefaults.S3KeyPrefix,
			),
			// kmsKeyID: CloudTrailConfig levels only (6, 7).
			KmsKeyID: firstNonEmptyString(
				localCTCfgDefaults.KmsKeyID,
				globalCTCfgDefaults.KmsKeyID,
			),
			// multiRegionEnabled: CloudTrailConfig levels only (6, 7).
			MultiRegionEnabled: firstTrue(
				localCTCfgDefaults.MultiRegionEnabled,
				globalCTCfgDefaults.MultiRegionEnabled,
			),
			// organizationEnabled: CloudTrailConfig levels only (6, 7).
			OrganizationEnabled: firstTrue(
				localCTCfgDefaults.OrganizationEnabled,
				globalCTCfgDefaults.OrganizationEnabled,
			),
			// retentionPeriod: CloudTrailConfig defaults levels 6-7 then KropathConfig levels 8-9.
			RetentionPeriod: firstNonZeroInt64(
				localCTCfgDefaults.RetentionPeriod,
				globalCTCfgDefaults.RetentionPeriod,
				localKropathDefaults.RetentionPeriod,
				globalKropathDefaults.RetentionPeriod,
			),
			// terminationProtectionEnabled: CloudTrailConfig levels only (6, 7).
			TerminationProtectionEnabled: firstTrue(
				localCTCfgDefaults.TerminationProtectionEnabled,
				globalCTCfgDefaults.TerminationProtectionEnabled,
			),
			// NamingTemplate: CloudTrailConfig levels only (6, 7).
			NamingTemplate: firstNonEmptyString(
				localCTCfgDefaults.NamingTemplate,
				globalCTCfgDefaults.NamingTemplate,
			),
			// SyncedLabels: additive union from CloudTrailConfig levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalCTCfgDefaults.SyncedLabels,
				localCTCfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalCTCfgDefaults.SyncedAnnotations,
				localCTCfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalCTCfgDefaults.Tags,
				localCTCfgDefaults.Tags,
			),
		},
	}
}
