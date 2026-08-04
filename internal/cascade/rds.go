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

// RDSKropathSection holds the RDS-family governance fields from
// KropathConfig.spec.mandatory.rds / .defaults.rds (ADR-015 §3.5).
//
// Contains only the 13 instance-applicable fields. Cluster-only fields
// (serverlessV2ScalingMinCapacity, serverlessV2ScalingMaxCapacity, backtrackWindow)
// are NOT present here — they live in RDSConfig (RDSConfigSection) only.
//
// Boolean pointer fields use nil = not set (falls through); false = explicitly disabled.
// String fields use "" = not enforced. Integer fields use 0 = not enforced.
type RDSKropathSection struct {
	// StorageEncrypted forces encryption at rest. nil = not enforced; true = forced on.
	StorageEncrypted *bool `json:"storageEncrypted,omitempty"`

	// KmsKeyID is the KMS key ARN for encryption. Empty = not enforced.
	KmsKeyID string `json:"kmsKeyID,omitempty"`

	// DeletionProtection prevents instance/cluster deletion. nil = not enforced; true = forced on.
	DeletionProtection *bool `json:"deletionProtection,omitempty"`

	// BackupRetentionPeriod is the minimum backup retention floor in days. 0 = not enforced.
	// Minimum-floor semantics: instance value must be >= this; controller writes max(mandatory, instance).
	BackupRetentionPeriod int64 `json:"backupRetentionPeriod,omitempty"`

	// MultiAZ forces Multi-AZ deployment. nil = not enforced; true = forced on.
	MultiAZ *bool `json:"multiAZ,omitempty"`

	// PubliclyAccessible controls public endpoint exposure.
	// nil = not enforced; false = forced private-only (pointer sentinel per OD-2 resolution).
	PubliclyAccessible *bool `json:"publiclyAccessible,omitempty"`

	// StorageType specifies the forced storage type. Empty = not enforced.
	StorageType string `json:"storageType,omitempty"`

	// AutoMinorVersionUpgrade forces auto minor version upgrades. nil = not enforced.
	AutoMinorVersionUpgrade *bool `json:"autoMinorVersionUpgrade,omitempty"`

	// CopyTagsToSnapshot forces tag copying to snapshots. nil = not enforced.
	CopyTagsToSnapshot *bool `json:"copyTagsToSnapshot,omitempty"`

	// PerformanceInsightsEnabled forces Performance Insights on. nil = not enforced.
	PerformanceInsightsEnabled *bool `json:"performanceInsightsEnabled,omitempty"`

	// EnableIAMDatabaseAuthentication forces IAM database authentication. nil = not enforced.
	EnableIAMDatabaseAuthentication *bool `json:"enableIAMDatabaseAuthentication,omitempty"`

	// ManageMasterUserPassword forces Secrets Manager password management. nil = not enforced.
	ManageMasterUserPassword *bool `json:"manageMasterUserPassword,omitempty"`

	// NamingTemplate enforces a naming pattern (e.g. "{namespace}-{name}"). Empty = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// Tags are tier-level cloud resource tags. nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// RDSConfigSection holds the RDS governance fields from RDSConfig.spec.mandatory
// or RDSConfig.spec.defaults (per-type ResourceConfig, ADR-015 §3.5).
//
// Contains all 16 governance fields including the 3 cluster-only fields.
// Cluster-only fields (ServerlessV2ScalingMinCapacity, ServerlessV2ScalingMaxCapacity,
// BacktrackWindow) are ignored by the controller when reconciling RDSInstance resources.
type RDSConfigSection struct {
	// StorageEncrypted forces encryption at rest. nil = not enforced; false = explicitly disabled.
	StorageEncrypted *bool `json:"storageEncrypted,omitempty"`

	// KmsKeyID is the KMS key ARN for encryption. Empty = not enforced.
	KmsKeyID string `json:"kmsKeyID,omitempty"`

	// DeletionProtection prevents instance/cluster deletion. nil = not enforced.
	DeletionProtection *bool `json:"deletionProtection,omitempty"`

	// BackupRetentionPeriod is the minimum backup retention floor in days. 0 = not enforced.
	BackupRetentionPeriod int64 `json:"backupRetentionPeriod,omitempty"`

	// MultiAZ forces Multi-AZ deployment. nil = not enforced.
	MultiAZ *bool `json:"multiAZ,omitempty"`

	// PubliclyAccessible controls public endpoint exposure. nil = not enforced; false = forced private-only.
	PubliclyAccessible *bool `json:"publiclyAccessible,omitempty"`

	// StorageType specifies the forced storage type. Empty = not enforced.
	StorageType string `json:"storageType,omitempty"`

	// AutoMinorVersionUpgrade forces auto minor version upgrades. nil = not enforced.
	AutoMinorVersionUpgrade *bool `json:"autoMinorVersionUpgrade,omitempty"`

	// CopyTagsToSnapshot forces tag copying to snapshots. nil = not enforced.
	CopyTagsToSnapshot *bool `json:"copyTagsToSnapshot,omitempty"`

	// PerformanceInsightsEnabled forces Performance Insights on. nil = not enforced.
	PerformanceInsightsEnabled *bool `json:"performanceInsightsEnabled,omitempty"`

	// EnableIAMDatabaseAuthentication forces IAM database authentication. nil = not enforced.
	EnableIAMDatabaseAuthentication *bool `json:"enableIAMDatabaseAuthentication,omitempty"`

	// ManageMasterUserPassword forces Secrets Manager password management. nil = not enforced.
	ManageMasterUserPassword *bool `json:"manageMasterUserPassword,omitempty"`

	// NamingTemplate enforces a naming pattern. Governed only at RDSConfig levels.
	// Empty = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// ServerlessV2ScalingMinCapacity is the minimum ACU floor for Serverless v2 clusters.
	// Cluster-only: ignored for RDSInstance reconciliation. 0 = not enforced.
	ServerlessV2ScalingMinCapacity float64 `json:"serverlessV2ScalingMinCapacity,omitempty"`

	// ServerlessV2ScalingMaxCapacity is the maximum ACU cap for Serverless v2 clusters.
	// Cluster-only: ignored for RDSInstance reconciliation. 0 = not enforced.
	ServerlessV2ScalingMaxCapacity float64 `json:"serverlessV2ScalingMaxCapacity,omitempty"`

	// BacktrackWindow is the mandatory backtrack window in seconds for Aurora MySQL clusters.
	// Cluster-only: ignored for RDSInstance reconciliation. 0 = not enforced.
	BacktrackWindow int64 `json:"backtrackWindow,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created resources.
	// Additive map merge from RDSConfig tiers only. nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate. Additive map merge from RDSConfig only.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`

	// Tags are cloud resource tags for this RDS config profile. nil / empty = no tags.
	Tags map[string]string `json:"tags,omitempty"`
}

// EffectiveRDSSection is one tier (mandatory or defaults) of the merged RDS governance
// result written into RDSConfig.status.effectiveConfig by the controller.
type EffectiveRDSSection struct {
	StorageEncrypted                *bool             `json:"storageEncrypted,omitempty"`
	KmsKeyID                        string            `json:"kmsKeyID,omitempty"`
	DeletionProtection              *bool             `json:"deletionProtection,omitempty"`
	BackupRetentionPeriod           int64             `json:"backupRetentionPeriod,omitempty"`
	MultiAZ                         *bool             `json:"multiAZ,omitempty"`
	PubliclyAccessible              *bool             `json:"publiclyAccessible,omitempty"`
	StorageType                     string            `json:"storageType,omitempty"`
	AutoMinorVersionUpgrade         *bool             `json:"autoMinorVersionUpgrade,omitempty"`
	CopyTagsToSnapshot              *bool             `json:"copyTagsToSnapshot,omitempty"`
	PerformanceInsightsEnabled      *bool             `json:"performanceInsightsEnabled,omitempty"`
	EnableIAMDatabaseAuthentication *bool             `json:"enableIAMDatabaseAuthentication,omitempty"`
	ManageMasterUserPassword        *bool             `json:"manageMasterUserPassword,omitempty"`
	NamingTemplate                  string            `json:"namingTemplate,omitempty"`
	ServerlessV2ScalingMinCapacity  float64           `json:"serverlessV2ScalingMinCapacity,omitempty"`
	ServerlessV2ScalingMaxCapacity  float64           `json:"serverlessV2ScalingMaxCapacity,omitempty"`
	BacktrackWindow                 int64             `json:"backtrackWindow,omitempty"`
	SyncedLabels                    map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations               map[string]string `json:"syncedAnnotations,omitempty"`
	Tags                            map[string]string `json:"tags,omitempty"`
}

// EffectiveRDSConfig is the merged RDS governance result written into
// RDSConfig.status.effectiveConfig by the controller.
type EffectiveRDSConfig struct {
	Mandatory EffectiveRDSSection `json:"mandatory"`
	Defaults  EffectiveRDSSection `json:"defaults"`
}

// MergeRDSCascade merges RDS governance fields from all cascade sources and returns
// the effective configuration to be written to RDSConfig.status.effectiveConfig.
//
// Nine-level priority chain for RDS (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.rds)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.rds)
//	Level 3 — globalRDSCfgMandatory   (RDSConfig in kro-system, mandatory)
//	Level 4 — localRDSCfgMandatory    (RDSConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localRDSCfgDefaults     (RDSConfig in resource namespace, defaults)
//	Level 7 — globalRDSCfgDefaults    (RDSConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.rds)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.rds)
//
// Merge rules:
//   - *bool pointer fields: firstNonNilBoolPtr in priority order (nil = not set, falls through).
//   - string fields: firstNonEmptyString in priority order ("" = not set, falls through).
//   - int64/float64 fields: firstNonZeroInt64 / firstNonZeroFloat64 in priority order.
//   - Tags: additive union merge across all four mandatory levels, all four defaults levels.
//   - SyncedLabels/SyncedAnnotations: additive union from RDSConfig levels only (no KropathConfig).
//   - NamingTemplate: governed only at RDSConfig levels (3-4 mandatory, 6-7 defaults);
//     KropathConfig.rds carries namingTemplate but it participates at levels 1-2 and 8-9.
//
// BackupRetentionPeriod minimum-floor semantics (OD-3): mandatory carries the floor value;
// RGD CEL computes max(effCfg.mandatory.backupRetentionPeriod, spec.backupRetentionPeriod)
// to allow instances to exceed but not go below the mandatory minimum.
//
// Cluster-only fields (ServerlessV2ScalingMinCapacity, ServerlessV2ScalingMaxCapacity,
// BacktrackWindow): populated from RDSConfig levels only; the controller propagates them
// only when reconciling RDSCluster resources (not RDSInstance).
func MergeRDSCascade(
	globalKropathMandatory RDSKropathSection, // level 1
	localKropathMandatory RDSKropathSection, // level 2
	globalRDSCfgMandatory RDSConfigSection, // level 3
	localRDSCfgMandatory RDSConfigSection, // level 4
	localRDSCfgDefaults RDSConfigSection, // level 6
	globalRDSCfgDefaults RDSConfigSection, // level 7
	localKropathDefaults RDSKropathSection, // level 8
	globalKropathDefaults RDSKropathSection, // level 9
) EffectiveRDSConfig {
	return EffectiveRDSConfig{
		Mandatory: EffectiveRDSSection{
			// *bool pointer fields: nil = not set, falls through to next level.
			StorageEncrypted: firstNonNilBoolPtr(
				globalKropathMandatory.StorageEncrypted,
				localKropathMandatory.StorageEncrypted,
				globalRDSCfgMandatory.StorageEncrypted,
				localRDSCfgMandatory.StorageEncrypted,
			),
			KmsKeyID: firstNonEmptyString(
				globalKropathMandatory.KmsKeyID,
				localKropathMandatory.KmsKeyID,
				globalRDSCfgMandatory.KmsKeyID,
				localRDSCfgMandatory.KmsKeyID,
			),
			DeletionProtection: firstNonNilBoolPtr(
				globalKropathMandatory.DeletionProtection,
				localKropathMandatory.DeletionProtection,
				globalRDSCfgMandatory.DeletionProtection,
				localRDSCfgMandatory.DeletionProtection,
			),
			// BackupRetentionPeriod: first non-zero value in priority order.
			// The mandatory tier carries the minimum floor; RGD CEL computes max(mandatory, instance).
			BackupRetentionPeriod: firstNonZeroInt64(
				globalKropathMandatory.BackupRetentionPeriod,
				localKropathMandatory.BackupRetentionPeriod,
				globalRDSCfgMandatory.BackupRetentionPeriod,
				localRDSCfgMandatory.BackupRetentionPeriod,
			),
			MultiAZ: firstNonNilBoolPtr(
				globalKropathMandatory.MultiAZ,
				localKropathMandatory.MultiAZ,
				globalRDSCfgMandatory.MultiAZ,
				localRDSCfgMandatory.MultiAZ,
			),
			// PubliclyAccessible: pointer sentinel. nil = not enforced; false = forced private-only.
			// globalKropathMandatory.PubliclyAccessible == *false → all instances forced private-only.
			PubliclyAccessible: firstNonNilBoolPtr(
				globalKropathMandatory.PubliclyAccessible,
				localKropathMandatory.PubliclyAccessible,
				globalRDSCfgMandatory.PubliclyAccessible,
				localRDSCfgMandatory.PubliclyAccessible,
			),
			StorageType: firstNonEmptyString(
				globalKropathMandatory.StorageType,
				localKropathMandatory.StorageType,
				globalRDSCfgMandatory.StorageType,
				localRDSCfgMandatory.StorageType,
			),
			AutoMinorVersionUpgrade: firstNonNilBoolPtr(
				globalKropathMandatory.AutoMinorVersionUpgrade,
				localKropathMandatory.AutoMinorVersionUpgrade,
				globalRDSCfgMandatory.AutoMinorVersionUpgrade,
				localRDSCfgMandatory.AutoMinorVersionUpgrade,
			),
			CopyTagsToSnapshot: firstNonNilBoolPtr(
				globalKropathMandatory.CopyTagsToSnapshot,
				localKropathMandatory.CopyTagsToSnapshot,
				globalRDSCfgMandatory.CopyTagsToSnapshot,
				localRDSCfgMandatory.CopyTagsToSnapshot,
			),
			PerformanceInsightsEnabled: firstNonNilBoolPtr(
				globalKropathMandatory.PerformanceInsightsEnabled,
				localKropathMandatory.PerformanceInsightsEnabled,
				globalRDSCfgMandatory.PerformanceInsightsEnabled,
				localRDSCfgMandatory.PerformanceInsightsEnabled,
			),
			EnableIAMDatabaseAuthentication: firstNonNilBoolPtr(
				globalKropathMandatory.EnableIAMDatabaseAuthentication,
				localKropathMandatory.EnableIAMDatabaseAuthentication,
				globalRDSCfgMandatory.EnableIAMDatabaseAuthentication,
				localRDSCfgMandatory.EnableIAMDatabaseAuthentication,
			),
			ManageMasterUserPassword: firstNonNilBoolPtr(
				globalKropathMandatory.ManageMasterUserPassword,
				localKropathMandatory.ManageMasterUserPassword,
				globalRDSCfgMandatory.ManageMasterUserPassword,
				localRDSCfgMandatory.ManageMasterUserPassword,
			),
			// NamingTemplate: KropathConfig levels 1-2 participate alongside RDSConfig levels 3-4.
			NamingTemplate: firstNonEmptyString(
				globalKropathMandatory.NamingTemplate,
				localKropathMandatory.NamingTemplate,
				globalRDSCfgMandatory.NamingTemplate,
				localRDSCfgMandatory.NamingTemplate,
			),
			// Cluster-only fields: governed at RDSConfig levels only (no KropathConfig equivalent).
			ServerlessV2ScalingMinCapacity: firstNonZeroFloat64(
				globalRDSCfgMandatory.ServerlessV2ScalingMinCapacity,
				localRDSCfgMandatory.ServerlessV2ScalingMinCapacity,
			),
			ServerlessV2ScalingMaxCapacity: firstNonZeroFloat64(
				globalRDSCfgMandatory.ServerlessV2ScalingMaxCapacity,
				localRDSCfgMandatory.ServerlessV2ScalingMaxCapacity,
			),
			BacktrackWindow: firstNonZeroInt64(
				globalRDSCfgMandatory.BacktrackWindow,
				localRDSCfgMandatory.BacktrackWindow,
			),
			// SyncedLabels: additive union from RDSConfig levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localRDSCfgMandatory.SyncedLabels,
				globalRDSCfgMandatory.SyncedLabels,
			),
			SyncedAnnotations: mergeMaps(
				localRDSCfgMandatory.SyncedAnnotations,
				globalRDSCfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localRDSCfgMandatory.Tags,
				globalRDSCfgMandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
		},
		Defaults: EffectiveRDSSection{
			StorageEncrypted: firstNonNilBoolPtr(
				localRDSCfgDefaults.StorageEncrypted,
				globalRDSCfgDefaults.StorageEncrypted,
				localKropathDefaults.StorageEncrypted,
				globalKropathDefaults.StorageEncrypted,
			),
			KmsKeyID: firstNonEmptyString(
				localRDSCfgDefaults.KmsKeyID,
				globalRDSCfgDefaults.KmsKeyID,
				localKropathDefaults.KmsKeyID,
				globalKropathDefaults.KmsKeyID,
			),
			DeletionProtection: firstNonNilBoolPtr(
				localRDSCfgDefaults.DeletionProtection,
				globalRDSCfgDefaults.DeletionProtection,
				localKropathDefaults.DeletionProtection,
				globalKropathDefaults.DeletionProtection,
			),
			BackupRetentionPeriod: firstNonZeroInt64(
				localRDSCfgDefaults.BackupRetentionPeriod,
				globalRDSCfgDefaults.BackupRetentionPeriod,
				localKropathDefaults.BackupRetentionPeriod,
				globalKropathDefaults.BackupRetentionPeriod,
			),
			MultiAZ: firstNonNilBoolPtr(
				localRDSCfgDefaults.MultiAZ,
				globalRDSCfgDefaults.MultiAZ,
				localKropathDefaults.MultiAZ,
				globalKropathDefaults.MultiAZ,
			),
			PubliclyAccessible: firstNonNilBoolPtr(
				localRDSCfgDefaults.PubliclyAccessible,
				globalRDSCfgDefaults.PubliclyAccessible,
				localKropathDefaults.PubliclyAccessible,
				globalKropathDefaults.PubliclyAccessible,
			),
			StorageType: firstNonEmptyString(
				localRDSCfgDefaults.StorageType,
				globalRDSCfgDefaults.StorageType,
				localKropathDefaults.StorageType,
				globalKropathDefaults.StorageType,
			),
			AutoMinorVersionUpgrade: firstNonNilBoolPtr(
				localRDSCfgDefaults.AutoMinorVersionUpgrade,
				globalRDSCfgDefaults.AutoMinorVersionUpgrade,
				localKropathDefaults.AutoMinorVersionUpgrade,
				globalKropathDefaults.AutoMinorVersionUpgrade,
			),
			CopyTagsToSnapshot: firstNonNilBoolPtr(
				localRDSCfgDefaults.CopyTagsToSnapshot,
				globalRDSCfgDefaults.CopyTagsToSnapshot,
				localKropathDefaults.CopyTagsToSnapshot,
				globalKropathDefaults.CopyTagsToSnapshot,
			),
			PerformanceInsightsEnabled: firstNonNilBoolPtr(
				localRDSCfgDefaults.PerformanceInsightsEnabled,
				globalRDSCfgDefaults.PerformanceInsightsEnabled,
				localKropathDefaults.PerformanceInsightsEnabled,
				globalKropathDefaults.PerformanceInsightsEnabled,
			),
			EnableIAMDatabaseAuthentication: firstNonNilBoolPtr(
				localRDSCfgDefaults.EnableIAMDatabaseAuthentication,
				globalRDSCfgDefaults.EnableIAMDatabaseAuthentication,
				localKropathDefaults.EnableIAMDatabaseAuthentication,
				globalKropathDefaults.EnableIAMDatabaseAuthentication,
			),
			ManageMasterUserPassword: firstNonNilBoolPtr(
				localRDSCfgDefaults.ManageMasterUserPassword,
				globalRDSCfgDefaults.ManageMasterUserPassword,
				localKropathDefaults.ManageMasterUserPassword,
				globalKropathDefaults.ManageMasterUserPassword,
			),
			NamingTemplate: firstNonEmptyString(
				localRDSCfgDefaults.NamingTemplate,
				globalRDSCfgDefaults.NamingTemplate,
				localKropathDefaults.NamingTemplate,
				globalKropathDefaults.NamingTemplate,
			),
			// Cluster-only fields: governed at RDSConfig levels only.
			ServerlessV2ScalingMinCapacity: firstNonZeroFloat64(
				localRDSCfgDefaults.ServerlessV2ScalingMinCapacity,
				globalRDSCfgDefaults.ServerlessV2ScalingMinCapacity,
			),
			ServerlessV2ScalingMaxCapacity: firstNonZeroFloat64(
				localRDSCfgDefaults.ServerlessV2ScalingMaxCapacity,
				globalRDSCfgDefaults.ServerlessV2ScalingMaxCapacity,
			),
			BacktrackWindow: firstNonZeroInt64(
				localRDSCfgDefaults.BacktrackWindow,
				globalRDSCfgDefaults.BacktrackWindow,
			),
			// SyncedLabels: additive union from RDSConfig levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalRDSCfgDefaults.SyncedLabels,
				localRDSCfgDefaults.SyncedLabels,
			),
			SyncedAnnotations: mergeMaps(
				globalRDSCfgDefaults.SyncedAnnotations,
				localRDSCfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalRDSCfgDefaults.Tags,
				localRDSCfgDefaults.Tags,
			),
		},
	}
}

// firstNonZeroFloat64 returns the first non-zero float64 from the candidates.
// Zero is the sentinel value for "not enforced" in float governance fields (e.g. Serverless v2 ACUs).
func firstNonZeroFloat64(candidates ...float64) float64 {
	for _, v := range candidates {
		if v != 0 {
			return v
		}
	}
	return 0
}
