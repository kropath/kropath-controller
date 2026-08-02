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

// DynamoDBKropathSection holds the DynamoDB-family governance fields from
// KropathConfig.spec.mandatory.dynamodb / .defaults.dynamodb (ADR-015 §3.5).
//
// Zero value for string fields is "" (not enforced).
// Boolean pointer fields use nil = not set (falls through); false = explicitly disabled.
// This *bool pointer sentinel is new for DynamoDB — prior families (IAM, KMS, S3, SQS, SNS,
// SecretsManager) used only string and integer sentinels.
//
// namingTemplate, syncedLabels, and syncedAnnotations are NOT present here;
// they live only in DynamoDBConfig (DynamoDBConfigSection).
type DynamoDBKropathSection struct {
	// EncryptionEnabled controls SSE (AWS-owned or KMS). nil = not set; false = explicitly disabled.
	EncryptionEnabled *bool `json:"encryptionEnabled,omitempty"`

	// KmsMasterKeyId is the KMS key ID/ARN for SSE-KMS. Empty string = not enforced.
	KmsMasterKeyId string `json:"kmsMasterKeyId,omitempty"`

	// DeletionProtectionEnabled prevents table deletion when true. nil = not set.
	DeletionProtectionEnabled *bool `json:"deletionProtectionEnabled,omitempty"`

	// PointInTimeRecoveryEnabled enables PITR. nil = not set.
	PointInTimeRecoveryEnabled *bool `json:"pointInTimeRecoveryEnabled,omitempty"`

	// BillingMode is the capacity billing mode ("PROVISIONED" or "PAY_PER_REQUEST").
	// Empty string = not enforced.
	BillingMode string `json:"billingMode,omitempty"`

	// TableClass is the DynamoDB table class ("STANDARD" or "STANDARD_INFREQUENT_ACCESS").
	// Empty string = not enforced.
	TableClass string `json:"tableClass,omitempty"`

	// ContributorInsights is the CloudWatch Contributor Insights action ("ENABLE" or "DISABLE").
	// Empty string = not enforced.
	ContributorInsights string `json:"contributorInsights,omitempty"`

	// Tags are tier-level cloud resource tags.
	// The reconciler populates this from KropathConfig.spec.mandatory.tags / .defaults.tags
	// so that tag union merge flows through MergeDynamoDBCascade alongside DynamoDB-specific fields.
	// nil / empty map = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// DynamoDBConfigSection holds the DynamoDB governance fields from DynamoDBConfig.spec.mandatory
// or DynamoDBConfig.spec.defaults (per-type ResourceConfig, ADR-015 §3.5).
//
// Boolean pointer fields use the same nil-aware sentinel as DynamoDBKropathSection.
type DynamoDBConfigSection struct {
	// EncryptionEnabled controls SSE. nil = not set; false = explicitly disabled.
	EncryptionEnabled *bool `json:"encryptionEnabled,omitempty"`

	// KmsMasterKeyId is the KMS key ID/ARN for SSE-KMS. Empty string = not enforced.
	KmsMasterKeyId string `json:"kmsMasterKeyId,omitempty"`

	// DeletionProtectionEnabled prevents table deletion when true. nil = not set.
	DeletionProtectionEnabled *bool `json:"deletionProtectionEnabled,omitempty"`

	// PointInTimeRecoveryEnabled enables PITR. nil = not set.
	PointInTimeRecoveryEnabled *bool `json:"pointInTimeRecoveryEnabled,omitempty"`

	// BillingMode is the capacity billing mode. Empty string = not enforced.
	BillingMode string `json:"billingMode,omitempty"`

	// TableClass is the DynamoDB table class. Empty string = not enforced.
	TableClass string `json:"tableClass,omitempty"`

	// ContributorInsights is the CloudWatch Contributor Insights action. Empty string = not enforced.
	ContributorInsights string `json:"contributorInsights,omitempty"`

	// NamingTemplate is the table naming template (e.g. "{namespace}-{name}").
	// Governed only at DynamoDBConfig levels 3-4 (mandatory) and 6-7 (defaults).
	// KropathConfig.dynamodb does NOT carry namingTemplate.
	// Empty string = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created table resources.
	// Additive map merge across DynamoDBConfig tiers only.
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to created table resources.
	// Additive map merge across DynamoDBConfig tiers only.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`

	// Tags are cloud resource tags for this DynamoDB config profile.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// EffectiveDynamoDBSection is one tier (mandatory or defaults) of the merged DynamoDB governance
// result written into DynamoDBConfig.status.effectiveConfig by the controller.
type EffectiveDynamoDBSection struct {
	EncryptionEnabled          *bool             `json:"encryptionEnabled,omitempty"`
	KmsMasterKeyId             string            `json:"kmsMasterKeyId,omitempty"`
	DeletionProtectionEnabled  *bool             `json:"deletionProtectionEnabled,omitempty"`
	PointInTimeRecoveryEnabled *bool             `json:"pointInTimeRecoveryEnabled,omitempty"`
	BillingMode                string            `json:"billingMode,omitempty"`
	TableClass                 string            `json:"tableClass,omitempty"`
	ContributorInsights        string            `json:"contributorInsights,omitempty"`
	NamingTemplate             string            `json:"namingTemplate,omitempty"`
	SyncedLabels               map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations          map[string]string `json:"syncedAnnotations,omitempty"`
	Tags                       map[string]string `json:"tags,omitempty"`
}

// EffectiveDynamoDBConfig is the merged DynamoDB governance result written into
// DynamoDBConfig.status.effectiveConfig by the controller.
type EffectiveDynamoDBConfig struct {
	Mandatory EffectiveDynamoDBSection `json:"mandatory"`
	Defaults  EffectiveDynamoDBSection `json:"defaults"`
}

// MergeDynamoDBCascade merges DynamoDB governance fields from all cascade sources and
// returns the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for DynamoDB (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.dynamodb)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.dynamodb)
//	Level 3 — globalDDBCfgMandatory   (DynamoDBConfig in kro-system, mandatory)
//	Level 4 — localDDBCfgMandatory    (DynamoDBConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localDDBCfgDefaults     (DynamoDBConfig in resource namespace, defaults)
//	Level 7 — globalDDBCfgDefaults    (DynamoDBConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.dynamodb)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.dynamodb)
//
// Scalar string merge: firstNonEmptyString in priority order (lowest number wins).
// *bool pointer merge: firstNonNilBoolPtr in priority order — nil = not set (falls through).
//
//	NOTE: This is a new sentinel type not present in IAM, KMS, S3, SQS, SNS, or SecretsManager
//	families. The pointer allows distinguishing "not set" (nil) from "explicitly disabled" (false).
//
// Tags: additive union merge across all four mandatory levels, all four defaults levels.
// SyncedLabels/SyncedAnnotations: additive union from DynamoDBConfig levels only (no KropathConfig).
// NamingTemplate: governed only at DynamoDBConfig levels (3-4 mandatory, 6-7 defaults).
func MergeDynamoDBCascade(
	globalKropathMandatory DynamoDBKropathSection, // level 1
	localKropathMandatory DynamoDBKropathSection, // level 2
	globalDDBCfgMandatory DynamoDBConfigSection, // level 3
	localDDBCfgMandatory DynamoDBConfigSection, // level 4
	localDDBCfgDefaults DynamoDBConfigSection, // level 6
	globalDDBCfgDefaults DynamoDBConfigSection, // level 7
	localKropathDefaults DynamoDBKropathSection, // level 8
	globalKropathDefaults DynamoDBKropathSection, // level 9
) EffectiveDynamoDBConfig {
	return EffectiveDynamoDBConfig{
		Mandatory: EffectiveDynamoDBSection{
			// *bool fields: nil = not set, false = explicitly disabled. firstNonNilBoolPtr picks
			// the first non-nil value in cascade order.
			EncryptionEnabled: firstNonNilBoolPtr(
				globalKropathMandatory.EncryptionEnabled,
				localKropathMandatory.EncryptionEnabled,
				globalDDBCfgMandatory.EncryptionEnabled,
				localDDBCfgMandatory.EncryptionEnabled,
			),
			KmsMasterKeyId: firstNonEmptyString(
				globalKropathMandatory.KmsMasterKeyId,
				localKropathMandatory.KmsMasterKeyId,
				globalDDBCfgMandatory.KmsMasterKeyId,
				localDDBCfgMandatory.KmsMasterKeyId,
			),
			DeletionProtectionEnabled: firstNonNilBoolPtr(
				globalKropathMandatory.DeletionProtectionEnabled,
				localKropathMandatory.DeletionProtectionEnabled,
				globalDDBCfgMandatory.DeletionProtectionEnabled,
				localDDBCfgMandatory.DeletionProtectionEnabled,
			),
			PointInTimeRecoveryEnabled: firstNonNilBoolPtr(
				globalKropathMandatory.PointInTimeRecoveryEnabled,
				localKropathMandatory.PointInTimeRecoveryEnabled,
				globalDDBCfgMandatory.PointInTimeRecoveryEnabled,
				localDDBCfgMandatory.PointInTimeRecoveryEnabled,
			),
			BillingMode: firstNonEmptyString(
				globalKropathMandatory.BillingMode,
				localKropathMandatory.BillingMode,
				globalDDBCfgMandatory.BillingMode,
				localDDBCfgMandatory.BillingMode,
			),
			TableClass: firstNonEmptyString(
				globalKropathMandatory.TableClass,
				localKropathMandatory.TableClass,
				globalDDBCfgMandatory.TableClass,
				localDDBCfgMandatory.TableClass,
			),
			ContributorInsights: firstNonEmptyString(
				globalKropathMandatory.ContributorInsights,
				localKropathMandatory.ContributorInsights,
				globalDDBCfgMandatory.ContributorInsights,
				localDDBCfgMandatory.ContributorInsights,
			),
			// NamingTemplate: DynamoDBConfig levels only (3, 4); KropathConfig has no namingTemplate.
			NamingTemplate: firstNonEmptyString(
				globalDDBCfgMandatory.NamingTemplate,
				localDDBCfgMandatory.NamingTemplate,
			),
			// SyncedLabels: additive union from DynamoDBConfig levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localDDBCfgMandatory.SyncedLabels,
				globalDDBCfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localDDBCfgMandatory.SyncedAnnotations,
				globalDDBCfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localDDBCfgMandatory.Tags,
				globalDDBCfgMandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
		},
		Defaults: EffectiveDynamoDBSection{
			EncryptionEnabled: firstNonNilBoolPtr(
				localDDBCfgDefaults.EncryptionEnabled,
				globalDDBCfgDefaults.EncryptionEnabled,
				localKropathDefaults.EncryptionEnabled,
				globalKropathDefaults.EncryptionEnabled,
			),
			KmsMasterKeyId: firstNonEmptyString(
				localDDBCfgDefaults.KmsMasterKeyId,
				globalDDBCfgDefaults.KmsMasterKeyId,
				localKropathDefaults.KmsMasterKeyId,
				globalKropathDefaults.KmsMasterKeyId,
			),
			DeletionProtectionEnabled: firstNonNilBoolPtr(
				localDDBCfgDefaults.DeletionProtectionEnabled,
				globalDDBCfgDefaults.DeletionProtectionEnabled,
				localKropathDefaults.DeletionProtectionEnabled,
				globalKropathDefaults.DeletionProtectionEnabled,
			),
			PointInTimeRecoveryEnabled: firstNonNilBoolPtr(
				localDDBCfgDefaults.PointInTimeRecoveryEnabled,
				globalDDBCfgDefaults.PointInTimeRecoveryEnabled,
				localKropathDefaults.PointInTimeRecoveryEnabled,
				globalKropathDefaults.PointInTimeRecoveryEnabled,
			),
			BillingMode: firstNonEmptyString(
				localDDBCfgDefaults.BillingMode,
				globalDDBCfgDefaults.BillingMode,
				localKropathDefaults.BillingMode,
				globalKropathDefaults.BillingMode,
			),
			TableClass: firstNonEmptyString(
				localDDBCfgDefaults.TableClass,
				globalDDBCfgDefaults.TableClass,
				localKropathDefaults.TableClass,
				globalKropathDefaults.TableClass,
			),
			ContributorInsights: firstNonEmptyString(
				localDDBCfgDefaults.ContributorInsights,
				globalDDBCfgDefaults.ContributorInsights,
				localKropathDefaults.ContributorInsights,
				globalKropathDefaults.ContributorInsights,
			),
			// NamingTemplate: DynamoDBConfig levels only (6, 7).
			NamingTemplate: firstNonEmptyString(
				localDDBCfgDefaults.NamingTemplate,
				globalDDBCfgDefaults.NamingTemplate,
			),
			// SyncedLabels: additive union from DynamoDBConfig levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalDDBCfgDefaults.SyncedLabels,
				localDDBCfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalDDBCfgDefaults.SyncedAnnotations,
				localDDBCfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalDDBCfgDefaults.Tags,
				localDDBCfgDefaults.Tags,
			),
		},
	}
}

// firstNonNilBoolPtr returns the first non-nil *bool from the candidates.
// This sentinel function is introduced for DynamoDB to distinguish nil (not set)
// from false (explicitly disabled). Prior families used string and int64 sentinels only.
func firstNonNilBoolPtr(candidates ...*bool) *bool {
	for _, b := range candidates {
		if b != nil {
			return b
		}
	}
	return nil
}
