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

import "fmt"

// DocumentDBKropathSection holds the DocumentDB-family governance fields from
// KropathConfig.spec.mandatory.documentdb / .defaults.documentdb (ADR-015 §3.5).
//
// Only the 4 org-wide governance fields are present here. The remaining DocumentDB
// fields (kmsKeyArn, performanceInsightsEnabled, dbInstanceClass, engineVersion,
// storageType, enableCloudwatchLogsExports, namingTemplate) live in DocumentDBConfig
// only. The Tags field is augmented from the KropathConfig tier-level tags before
// passing to MergeDocumentDBCascade.
type DocumentDBKropathSection struct {
	// StorageEncrypted forces encryption at rest. nil = not enforced; true = forced on.
	StorageEncrypted *bool `json:"storageEncrypted,omitempty"`

	// DeletionProtection prevents cluster/instance deletion. nil = not enforced; true = forced on.
	DeletionProtection *bool `json:"deletionProtection,omitempty"`

	// AllowedInstanceClasses restricts the allowed DocumentDB instance classes to an org-wide allowlist.
	// Empty slice = no restriction.
	AllowedInstanceClasses []string `json:"allowedInstanceClasses,omitempty"`

	// BackupRetentionPeriod is the minimum backup retention floor in days. 0 = not enforced.
	// Minimum-floor semantics: instance value must be >= this value.
	BackupRetentionPeriod int64 `json:"backupRetentionPeriod,omitempty"`

	// Tags are tier-level cloud resource tags. Augmented from KropathConfig.spec.mandatory.tags
	// before calling MergeDocumentDBCascade. nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// DocumentDBConfigSection holds the DocumentDB governance fields from DocumentDBConfig.spec.mandatory
// or DocumentDBConfig.spec.defaults (per-type ResourceConfig, ADR-015 §3.5).
//
// Contains all 13 governance field groups defined in aws-docdb-01.
type DocumentDBConfigSection struct {
	// StorageEncrypted forces encryption at rest. nil = not enforced; false = explicitly disabled.
	StorageEncrypted *bool `json:"storageEncrypted,omitempty"`

	// DeletionProtection prevents cluster/instance deletion. nil = not enforced.
	DeletionProtection *bool `json:"deletionProtection,omitempty"`

	// KmsKeyArn is the KMS key ARN for encryption at rest. Empty = not enforced.
	KmsKeyArn string `json:"kmsKeyArn,omitempty"`

	// PerformanceInsightsEnabled forces Performance Insights. nil = not enforced.
	PerformanceInsightsEnabled *bool `json:"performanceInsightsEnabled,omitempty"`

	// DbInstanceClass forces a specific instance class (e.g. "db.r6g.large"). Empty = not enforced.
	DbInstanceClass string `json:"dbInstanceClass,omitempty"`

	// EngineVersion forces a specific DocumentDB engine version. Empty = not enforced.
	EngineVersion string `json:"engineVersion,omitempty"`

	// AllowedInstanceClasses restricts the allowed instance classes to an allowlist.
	// Empty = no restriction.
	AllowedInstanceClasses []string `json:"allowedInstanceClasses,omitempty"`

	// StorageType forces a specific storage type (e.g. "standard", "iopt1"). Empty = not enforced.
	StorageType string `json:"storageType,omitempty"`

	// BackupRetentionPeriod is the minimum backup retention floor in days. 0 = not enforced.
	BackupRetentionPeriod int64 `json:"backupRetentionPeriod,omitempty"`

	// EnableCloudwatchLogsExports forces specific CloudWatch log export types. Empty = not enforced.
	EnableCloudwatchLogsExports []string `json:"enableCloudwatchLogsExports,omitempty"`

	// NamingTemplate enforces a naming pattern. Empty = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// Tags are cloud resource tags for this DocumentDB config profile. nil / empty = no tags.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created resources. Additive map merge.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate. Additive map merge.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveDocumentDBSection is one tier (mandatory or defaults) of the merged DocumentDB
// governance result written into DocumentDBConfig.status.effectiveConfig by the controller.
type EffectiveDocumentDBSection struct {
	StorageEncrypted            *bool             `json:"storageEncrypted,omitempty"`
	DeletionProtection          *bool             `json:"deletionProtection,omitempty"`
	KmsKeyArn                   string            `json:"kmsKeyArn,omitempty"`
	PerformanceInsightsEnabled  *bool             `json:"performanceInsightsEnabled,omitempty"`
	DbInstanceClass             string            `json:"dbInstanceClass,omitempty"`
	EngineVersion               string            `json:"engineVersion,omitempty"`
	AllowedInstanceClasses      []string          `json:"allowedInstanceClasses,omitempty"`
	StorageType                 string            `json:"storageType,omitempty"`
	BackupRetentionPeriod       int64             `json:"backupRetentionPeriod,omitempty"`
	EnableCloudwatchLogsExports []string          `json:"enableCloudwatchLogsExports,omitempty"`
	NamingTemplate              string            `json:"namingTemplate,omitempty"`
	Tags                        map[string]string `json:"tags,omitempty"`
	SyncedLabels                map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations           map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveDocumentDBConfig is the merged DocumentDB governance result written into
// DocumentDBConfig.status.effectiveConfig by the controller.
type EffectiveDocumentDBConfig struct {
	Mandatory EffectiveDocumentDBSection `json:"mandatory"`
	Defaults  EffectiveDocumentDBSection `json:"defaults"`
}

// MergeDocumentDBCascade merges DocumentDB governance fields from all cascade sources and returns
// the effective configuration to be written to DocumentDBConfig.status.effectiveConfig.
//
// Nine-level priority chain for DocumentDB (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.documentdb)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.documentdb)
//	Level 3 — globalDocDBCfgMandatory (DocumentDBConfig in kro-system, mandatory)
//	Level 4 — localDocDBCfgMandatory  (DocumentDBConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localDocDBCfgDefaults   (DocumentDBConfig in resource namespace, defaults)
//	Level 7 — globalDocDBCfgDefaults  (DocumentDBConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.documentdb)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.documentdb)
//
// Merge rules:
//   - *bool pointer fields: firstNonNilBoolPtr in priority order (nil = not set, falls through).
//   - string fields: firstNonEmptyString in priority order ("" = not set, falls through).
//   - int64 fields: firstNonZeroInt64 in priority order (0 = not enforced, falls through).
//   - []string fields: firstNonEmptyStrings in priority order (empty slice = not set, falls through).
//   - Tags: additive union merge across all four mandatory levels, all four defaults levels.
//   - SyncedLabels/SyncedAnnotations: additive union from DocumentDBConfig levels only.
//
// BackupRetentionPeriod minimum-floor semantics (OD-3): the mandatory tier carries the floor;
// RGD CEL computes max(effCfg.mandatory.backupRetentionPeriod, spec.backupRetentionPeriod).
func MergeDocumentDBCascade(
	globalKropathMandatory DocumentDBKropathSection, // level 1
	localKropathMandatory DocumentDBKropathSection,  // level 2
	globalDocDBCfgMandatory DocumentDBConfigSection, // level 3
	localDocDBCfgMandatory DocumentDBConfigSection,  // level 4
	localDocDBCfgDefaults DocumentDBConfigSection,   // level 6
	globalDocDBCfgDefaults DocumentDBConfigSection,  // level 7
	localKropathDefaults DocumentDBKropathSection,   // level 8
	globalKropathDefaults DocumentDBKropathSection,  // level 9
) EffectiveDocumentDBConfig {
	return EffectiveDocumentDBConfig{
		Mandatory: EffectiveDocumentDBSection{
			StorageEncrypted: firstNonNilBoolPtr(
				globalKropathMandatory.StorageEncrypted,
				localKropathMandatory.StorageEncrypted,
				globalDocDBCfgMandatory.StorageEncrypted,
				localDocDBCfgMandatory.StorageEncrypted,
			),
			DeletionProtection: firstNonNilBoolPtr(
				globalKropathMandatory.DeletionProtection,
				localKropathMandatory.DeletionProtection,
				globalDocDBCfgMandatory.DeletionProtection,
				localDocDBCfgMandatory.DeletionProtection,
			),
			// KmsKeyArn: KropathConfig levels have no documentdb.kmsKeyArn equivalent.
			KmsKeyArn: firstNonEmptyString(
				globalDocDBCfgMandatory.KmsKeyArn,
				localDocDBCfgMandatory.KmsKeyArn,
			),
			PerformanceInsightsEnabled: firstNonNilBoolPtr(
				globalDocDBCfgMandatory.PerformanceInsightsEnabled,
				localDocDBCfgMandatory.PerformanceInsightsEnabled,
			),
			DbInstanceClass: firstNonEmptyString(
				globalDocDBCfgMandatory.DbInstanceClass,
				localDocDBCfgMandatory.DbInstanceClass,
			),
			EngineVersion: firstNonEmptyString(
				globalDocDBCfgMandatory.EngineVersion,
				localDocDBCfgMandatory.EngineVersion,
			),
			// AllowedInstanceClasses: KropathConfig levels 1-2 exist for this field.
			AllowedInstanceClasses: firstNonEmptyStrings(
				globalKropathMandatory.AllowedInstanceClasses,
				localKropathMandatory.AllowedInstanceClasses,
				globalDocDBCfgMandatory.AllowedInstanceClasses,
				localDocDBCfgMandatory.AllowedInstanceClasses,
			),
			StorageType: firstNonEmptyString(
				globalDocDBCfgMandatory.StorageType,
				localDocDBCfgMandatory.StorageType,
			),
			// BackupRetentionPeriod: KropathConfig levels 1-2 participate.
			BackupRetentionPeriod: firstNonZeroInt64(
				globalKropathMandatory.BackupRetentionPeriod,
				localKropathMandatory.BackupRetentionPeriod,
				globalDocDBCfgMandatory.BackupRetentionPeriod,
				localDocDBCfgMandatory.BackupRetentionPeriod,
			),
			EnableCloudwatchLogsExports: firstNonEmptyStrings(
				globalDocDBCfgMandatory.EnableCloudwatchLogsExports,
				localDocDBCfgMandatory.EnableCloudwatchLogsExports,
			),
			NamingTemplate: firstNonEmptyString(
				globalDocDBCfgMandatory.NamingTemplate,
				localDocDBCfgMandatory.NamingTemplate,
			),
			// SyncedLabels/SyncedAnnotations: additive union from DocumentDBConfig levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localDocDBCfgMandatory.SyncedLabels,
				globalDocDBCfgMandatory.SyncedLabels,
			),
			SyncedAnnotations: mergeMaps(
				localDocDBCfgMandatory.SyncedAnnotations,
				globalDocDBCfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localDocDBCfgMandatory.Tags,
				globalDocDBCfgMandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
		},
		Defaults: EffectiveDocumentDBSection{
			StorageEncrypted: firstNonNilBoolPtr(
				localDocDBCfgDefaults.StorageEncrypted,
				globalDocDBCfgDefaults.StorageEncrypted,
				localKropathDefaults.StorageEncrypted,
				globalKropathDefaults.StorageEncrypted,
			),
			DeletionProtection: firstNonNilBoolPtr(
				localDocDBCfgDefaults.DeletionProtection,
				globalDocDBCfgDefaults.DeletionProtection,
				localKropathDefaults.DeletionProtection,
				globalKropathDefaults.DeletionProtection,
			),
			KmsKeyArn: firstNonEmptyString(
				localDocDBCfgDefaults.KmsKeyArn,
				globalDocDBCfgDefaults.KmsKeyArn,
			),
			PerformanceInsightsEnabled: firstNonNilBoolPtr(
				localDocDBCfgDefaults.PerformanceInsightsEnabled,
				globalDocDBCfgDefaults.PerformanceInsightsEnabled,
			),
			DbInstanceClass: firstNonEmptyString(
				localDocDBCfgDefaults.DbInstanceClass,
				globalDocDBCfgDefaults.DbInstanceClass,
			),
			EngineVersion: firstNonEmptyString(
				localDocDBCfgDefaults.EngineVersion,
				globalDocDBCfgDefaults.EngineVersion,
			),
			// AllowedInstanceClasses: KropathConfig levels 8-9 exist for this field.
			AllowedInstanceClasses: firstNonEmptyStrings(
				localDocDBCfgDefaults.AllowedInstanceClasses,
				globalDocDBCfgDefaults.AllowedInstanceClasses,
				localKropathDefaults.AllowedInstanceClasses,
				globalKropathDefaults.AllowedInstanceClasses,
			),
			StorageType: firstNonEmptyString(
				localDocDBCfgDefaults.StorageType,
				globalDocDBCfgDefaults.StorageType,
			),
			// BackupRetentionPeriod: KropathConfig levels 8-9 participate.
			BackupRetentionPeriod: firstNonZeroInt64(
				localDocDBCfgDefaults.BackupRetentionPeriod,
				globalDocDBCfgDefaults.BackupRetentionPeriod,
				localKropathDefaults.BackupRetentionPeriod,
				globalKropathDefaults.BackupRetentionPeriod,
			),
			EnableCloudwatchLogsExports: firstNonEmptyStrings(
				localDocDBCfgDefaults.EnableCloudwatchLogsExports,
				globalDocDBCfgDefaults.EnableCloudwatchLogsExports,
			),
			NamingTemplate: firstNonEmptyString(
				localDocDBCfgDefaults.NamingTemplate,
				globalDocDBCfgDefaults.NamingTemplate,
			),
			// SyncedLabels/SyncedAnnotations: additive union from DocumentDBConfig levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalDocDBCfgDefaults.SyncedLabels,
				localDocDBCfgDefaults.SyncedLabels,
			),
			SyncedAnnotations: mergeMaps(
				globalDocDBCfgDefaults.SyncedAnnotations,
				localDocDBCfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalDocDBCfgDefaults.Tags,
				localDocDBCfgDefaults.Tags,
			),
		},
	}
}

// ValidateDocumentDBInstanceClass checks the dbInstanceClass/allowedInstanceClasses cross-field
// constraint on the resolved mandatory tier.
//
// When mandatory.dbInstanceClass is non-empty AND mandatory.allowedInstanceClasses is non-empty,
// dbInstanceClass must be a member of allowedInstanceClasses.
//
// Returns (valid=true, "") when the constraint is satisfied or does not apply.
// Returns (valid=false, message) when dbInstanceClass is not in allowedInstanceClasses.
func ValidateDocumentDBInstanceClass(mandatory EffectiveDocumentDBSection) (bool, string) {
	if mandatory.DbInstanceClass == "" || len(mandatory.AllowedInstanceClasses) == 0 {
		return true, ""
	}
	for _, allowed := range mandatory.AllowedInstanceClasses {
		if mandatory.DbInstanceClass == allowed {
			return true, ""
		}
	}
	return false, fmt.Sprintf(
		"dbInstanceClass %q is not in allowedInstanceClasses %v",
		mandatory.DbInstanceClass,
		mandatory.AllowedInstanceClasses,
	)
}
