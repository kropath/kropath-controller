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

// MSKKropathSection holds the MSK-family governance fields from
// KropathConfig.spec.mandatory.msk / .defaults.msk (ADR-015 §3.5).
//
// Five of the MSKConfig fields have KropathConfig equivalents (org-wide blanket
// governance): kafkaVersion, enhancedMonitoring, encryptionInTransitClientBroker,
// encryptionInTransitInCluster, and encryptionAtRestKmsKeyId.
//
// instanceType and namingTemplate are profile-specific and governed at the MSKConfig
// level only — they do NOT appear in KropathConfig.msk.
//
// Zero value of each string field is the permissive sentinel ("" = not enforced).
type MSKKropathSection struct {
	// KafkaVersion enforces a minimum Kafka version org-wide.
	// "" (zero value) = not enforced at this level.
	KafkaVersion string `json:"kafkaVersion,omitempty"`

	// EnhancedMonitoring enforces a CloudWatch monitoring level org-wide.
	// "" = not enforced. Valid values: "DEFAULT" | "PER_BROKER" | "PER_TOPIC_PER_BROKER" | "PER_TOPIC_PER_PARTITION".
	EnhancedMonitoring string `json:"enhancedMonitoring,omitempty"`

	// EncryptionInTransitClientBroker enforces client-broker encryption mode org-wide.
	// "" = not enforced. Valid values: "TLS" | "TLS_PLAINTEXT" | "PLAINTEXT".
	EncryptionInTransitClientBroker string `json:"encryptionInTransitClientBroker,omitempty"`

	// EncryptionInTransitInCluster enforces inter-broker encryption org-wide (three-state string).
	// "" = not enforced. Valid values: "true" | "false".
	EncryptionInTransitInCluster string `json:"encryptionInTransitInCluster,omitempty"`

	// EncryptionAtRestKmsKeyId enforces an org-wide KMS key for EBS encryption.
	// "" = not enforced.
	EncryptionAtRestKmsKeyId string `json:"encryptionAtRestKmsKeyId,omitempty"`

	// Tags are tier-level cloud resource tags merged from KropathConfig.spec.mandatory.tags /
	// .defaults.tags — the generic cross-family tags, not an msk-scoped field.
	// nil / empty map = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// MSKConfigSection holds the MSK governance fields from
// MSKConfig.spec.mandatory or MSKConfig.spec.defaults
// (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each string field is the permissive sentinel ("" = not enforced).
type MSKConfigSection struct {
	// KafkaVersion enforces the Kafka version for all clusters in scope.
	// "" = not enforced.
	KafkaVersion string `json:"kafkaVersion,omitempty"`

	// InstanceType enforces the broker EC2 instance type.
	// "" = not enforced. Example: "kafka.m5.large".
	// No KropathConfig.msk equivalent — profile-specific governance only.
	InstanceType string `json:"instanceType,omitempty"`

	// EncryptionAtRestKmsKeyId enforces the KMS key for EBS encryption.
	// "" = not enforced.
	EncryptionAtRestKmsKeyId string `json:"encryptionAtRestKmsKeyId,omitempty"`

	// EncryptionInTransitClientBroker enforces client-broker encryption mode.
	// "" = not enforced. Valid values: "TLS" | "TLS_PLAINTEXT" | "PLAINTEXT".
	EncryptionInTransitClientBroker string `json:"encryptionInTransitClientBroker,omitempty"`

	// EncryptionInTransitInCluster enforces inter-broker encryption (three-state string).
	// "" = not enforced. Valid values: "true" | "false".
	EncryptionInTransitInCluster string `json:"encryptionInTransitInCluster,omitempty"`

	// EnhancedMonitoring enforces the CloudWatch monitoring level.
	// "" = not enforced. Valid values: "DEFAULT" | "PER_BROKER" | "PER_TOPIC_PER_BROKER" | "PER_TOPIC_PER_PARTITION".
	EnhancedMonitoring string `json:"enhancedMonitoring,omitempty"`

	// NamingTemplate is the naming template (e.g. "{namespace}-{name}").
	// "" = not enforced.
	// No KropathConfig.msk equivalent — profile-specific governance only.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// Tags are cloud resource tags for this MSK config profile.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created resources
	// (applied as both K8s labels and cloud tags per ADR-015 §6.1).
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to created resources.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveMSKSection is one tier (mandatory or defaults) of the merged MSK
// governance result written into MSKConfig.status.effectiveConfig.
type EffectiveMSKSection struct {
	KafkaVersion                    string            `json:"kafkaVersion,omitempty"`
	InstanceType                    string            `json:"instanceType,omitempty"`
	EncryptionAtRestKmsKeyId        string            `json:"encryptionAtRestKmsKeyId,omitempty"`
	EncryptionInTransitClientBroker string            `json:"encryptionInTransitClientBroker,omitempty"`
	EncryptionInTransitInCluster    string            `json:"encryptionInTransitInCluster,omitempty"`
	EnhancedMonitoring              string            `json:"enhancedMonitoring,omitempty"`
	NamingTemplate                  string            `json:"namingTemplate,omitempty"`
	Tags                            map[string]string `json:"tags,omitempty"`
	SyncedLabels                    map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations               map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveMSKConfig is the merged MSK governance result written into
// MSKConfig.status.effectiveConfig by the controller.
type EffectiveMSKConfig struct {
	Mandatory EffectiveMSKSection `json:"mandatory"`
	Defaults  EffectiveMSKSection `json:"defaults"`
}

// MergeMSKCascade merges MSK governance fields from all cascade sources and
// returns the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for MSK (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.msk)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.msk)
//	Level 3 — globalMSKCfgMandatory   (MSKConfig in kro-system, mandatory)
//	Level 4 — localMSKCfgMandatory    (MSKConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localMSKCfgDefaults     (MSKConfig in resource namespace, defaults)
//	Level 7 — globalMSKCfgDefaults    (MSKConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.msk)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.msk)
//
// Five fields have KropathConfig.msk equivalents (levels 1-2, 8-9):
//   kafkaVersion, enhancedMonitoring, encryptionInTransitClientBroker,
//   encryptionInTransitInCluster, encryptionAtRestKmsKeyId.
//
// Two fields have no KropathConfig.msk equivalent (levels 3-7 only):
//   instanceType, namingTemplate.
//
// String merge: firstNonEmptyString in priority order.
// Tags: additive union merge across all mandatory levels, all defaults levels.
// SyncedLabels/SyncedAnnotations: additive union from MSKConfig levels only (no KropathConfig).
func MergeMSKCascade(
	globalKropathMandatory MSKKropathSection, // level 1
	localKropathMandatory MSKKropathSection, // level 2
	globalMSKCfgMandatory MSKConfigSection, // level 3
	localMSKCfgMandatory MSKConfigSection, // level 4
	localMSKCfgDefaults MSKConfigSection, // level 6
	globalMSKCfgDefaults MSKConfigSection, // level 7
	localKropathDefaults MSKKropathSection, // level 8
	globalKropathDefaults MSKKropathSection, // level 9
) EffectiveMSKConfig {
	return EffectiveMSKConfig{
		Mandatory: EffectiveMSKSection{
			// KropathConfig levels (1-2) exist for 5 fields.
			KafkaVersion: firstNonEmptyString(
				globalKropathMandatory.KafkaVersion,
				localKropathMandatory.KafkaVersion,
				globalMSKCfgMandatory.KafkaVersion,
				localMSKCfgMandatory.KafkaVersion,
			),
			EncryptionAtRestKmsKeyId: firstNonEmptyString(
				globalKropathMandatory.EncryptionAtRestKmsKeyId,
				localKropathMandatory.EncryptionAtRestKmsKeyId,
				globalMSKCfgMandatory.EncryptionAtRestKmsKeyId,
				localMSKCfgMandatory.EncryptionAtRestKmsKeyId,
			),
			EncryptionInTransitClientBroker: firstNonEmptyString(
				globalKropathMandatory.EncryptionInTransitClientBroker,
				localKropathMandatory.EncryptionInTransitClientBroker,
				globalMSKCfgMandatory.EncryptionInTransitClientBroker,
				localMSKCfgMandatory.EncryptionInTransitClientBroker,
			),
			EncryptionInTransitInCluster: firstNonEmptyString(
				globalKropathMandatory.EncryptionInTransitInCluster,
				localKropathMandatory.EncryptionInTransitInCluster,
				globalMSKCfgMandatory.EncryptionInTransitInCluster,
				localMSKCfgMandatory.EncryptionInTransitInCluster,
			),
			EnhancedMonitoring: firstNonEmptyString(
				globalKropathMandatory.EnhancedMonitoring,
				localKropathMandatory.EnhancedMonitoring,
				globalMSKCfgMandatory.EnhancedMonitoring,
				localMSKCfgMandatory.EnhancedMonitoring,
			),
			// No KropathConfig.msk levels for instanceType and namingTemplate.
			InstanceType: firstNonEmptyString(
				globalMSKCfgMandatory.InstanceType,
				localMSKCfgMandatory.InstanceType,
			),
			NamingTemplate: firstNonEmptyString(
				globalMSKCfgMandatory.NamingTemplate,
				localMSKCfgMandatory.NamingTemplate,
			),
			// SyncedLabels: additive union from MSKConfig levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localMSKCfgMandatory.SyncedLabels,
				globalMSKCfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localMSKCfgMandatory.SyncedAnnotations,
				globalMSKCfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localMSKCfgMandatory.Tags,
				globalMSKCfgMandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
		},
		Defaults: EffectiveMSKSection{
			// KropathConfig levels (8-9) exist for 5 fields.
			KafkaVersion: firstNonEmptyString(
				localMSKCfgDefaults.KafkaVersion,
				globalMSKCfgDefaults.KafkaVersion,
				localKropathDefaults.KafkaVersion,
				globalKropathDefaults.KafkaVersion,
			),
			EncryptionAtRestKmsKeyId: firstNonEmptyString(
				localMSKCfgDefaults.EncryptionAtRestKmsKeyId,
				globalMSKCfgDefaults.EncryptionAtRestKmsKeyId,
				localKropathDefaults.EncryptionAtRestKmsKeyId,
				globalKropathDefaults.EncryptionAtRestKmsKeyId,
			),
			EncryptionInTransitClientBroker: firstNonEmptyString(
				localMSKCfgDefaults.EncryptionInTransitClientBroker,
				globalMSKCfgDefaults.EncryptionInTransitClientBroker,
				localKropathDefaults.EncryptionInTransitClientBroker,
				globalKropathDefaults.EncryptionInTransitClientBroker,
			),
			EncryptionInTransitInCluster: firstNonEmptyString(
				localMSKCfgDefaults.EncryptionInTransitInCluster,
				globalMSKCfgDefaults.EncryptionInTransitInCluster,
				localKropathDefaults.EncryptionInTransitInCluster,
				globalKropathDefaults.EncryptionInTransitInCluster,
			),
			EnhancedMonitoring: firstNonEmptyString(
				localMSKCfgDefaults.EnhancedMonitoring,
				globalMSKCfgDefaults.EnhancedMonitoring,
				localKropathDefaults.EnhancedMonitoring,
				globalKropathDefaults.EnhancedMonitoring,
			),
			// No KropathConfig.msk levels for instanceType and namingTemplate.
			InstanceType: firstNonEmptyString(
				localMSKCfgDefaults.InstanceType,
				globalMSKCfgDefaults.InstanceType,
			),
			NamingTemplate: firstNonEmptyString(
				localMSKCfgDefaults.NamingTemplate,
				globalMSKCfgDefaults.NamingTemplate,
			),
			// SyncedLabels: additive union from MSKConfig levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalMSKCfgDefaults.SyncedLabels,
				localMSKCfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalMSKCfgDefaults.SyncedAnnotations,
				localMSKCfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalMSKCfgDefaults.Tags,
				localMSKCfgDefaults.Tags,
			),
		},
	}
}
