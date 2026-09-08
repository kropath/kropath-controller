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

// MQKropathSection holds the MQ-family governance fields from
// KropathConfig.spec.mandatory.mq / .defaults.mq (ADR-015 §3.5).
//
// Nine fields are governed at the KropathConfig level (the org-wide blanket fields).
// Two fields — hostInstanceType and namingTemplate — are profile-specific and
// governed only at MQConfig levels 3-4 and 6-7; they are NOT present here.
//
// Boolean pointer fields use nil as the non-enforced sentinel:
// nil = not enforced; true/false = explicitly enforced.
type MQKropathSection struct {
	// EngineType enforces the broker engine type org-wide.
	// Empty string = not enforced; ACTIVEMQ | RABBITMQ.
	EngineType string `json:"engineType,omitempty"`

	// AuthenticationStrategy enforces the authentication strategy org-wide.
	// Empty string = not enforced; SIMPLE | LDAP (ActiveMQ only).
	AuthenticationStrategy string `json:"authenticationStrategy,omitempty"`

	// DeploymentMode enforces the deployment mode org-wide.
	// Empty string = not enforced; SINGLE_INSTANCE | ACTIVE_STANDBY_MULTI_AZ | CLUSTER_MULTI_AZ.
	DeploymentMode string `json:"deploymentMode,omitempty"`

	// PubliclyAccessible enforces public accessibility org-wide.
	// nil = not enforced; false = mandate private; true = allow public.
	PubliclyAccessible *bool `json:"publiclyAccessible,omitempty"`

	// AutoMinorVersionUpgrade enforces automatic minor version upgrades org-wide.
	// nil = not enforced; true = mandate auto-upgrade; false = prohibit.
	AutoMinorVersionUpgrade *bool `json:"autoMinorVersionUpgrade,omitempty"`

	// LogsGeneral enforces general CloudWatch logging org-wide.
	// nil = not enforced; true = mandate logging; false = prohibit.
	LogsGeneral *bool `json:"logsGeneral,omitempty"`

	// LogsAudit enforces audit logging org-wide (ActiveMQ only).
	// nil = not enforced; true = mandate audit logging; false = prohibit.
	LogsAudit *bool `json:"logsAudit,omitempty"`

	// EncryptionUseAWSOwnedKey enforces the encryption key type org-wide.
	// nil = not enforced; false = mandate customer-managed KMS; true = allow AWS-owned key.
	EncryptionUseAWSOwnedKey *bool `json:"encryptionUseAWSOwnedKey,omitempty"`

	// StorageType enforces the storage type org-wide (ActiveMQ only).
	// Empty string = not enforced; efs | io1.
	StorageType string `json:"storageType,omitempty"`

	// Tags are tier-level cloud resource tags merged from KropathConfig.spec.mandatory.tags /
	// .defaults.tags — the generic cross-family tags, not an mq-scoped field.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// MQConfigSection holds the MQ governance fields from
// MQConfig.spec.mandatory or MQConfig.spec.defaults
// (per-type ResourceConfig, ADR-015 §3.5).
//
// All nine org-wide fields plus two MQConfig-only fields
// (hostInstanceType, namingTemplate) are present here.
type MQConfigSection struct {
	// EngineType restricts the broker engine type for this profile.
	// Empty string = not enforced.
	EngineType string `json:"engineType,omitempty"`

	// AuthenticationStrategy restricts the authentication strategy for this profile.
	// Empty string = not enforced.
	AuthenticationStrategy string `json:"authenticationStrategy,omitempty"`

	// DeploymentMode restricts the deployment mode for this profile.
	// Empty string = not enforced.
	DeploymentMode string `json:"deploymentMode,omitempty"`

	// HostInstanceType restricts the broker instance type for this profile.
	// MQConfig-only (no KropathConfig.mq equivalent — per D-4 decision rule).
	// Empty string = not enforced.
	HostInstanceType string `json:"hostInstanceType,omitempty"`

	// PubliclyAccessible governs public accessibility for this profile.
	// nil = not enforced.
	PubliclyAccessible *bool `json:"publiclyAccessible,omitempty"`

	// AutoMinorVersionUpgrade governs automatic minor version upgrades for this profile.
	// nil = not enforced.
	AutoMinorVersionUpgrade *bool `json:"autoMinorVersionUpgrade,omitempty"`

	// LogsGeneral governs general CloudWatch logging for this profile.
	// nil = not enforced.
	LogsGeneral *bool `json:"logsGeneral,omitempty"`

	// LogsAudit governs audit logging for this profile (ActiveMQ only).
	// nil = not enforced.
	LogsAudit *bool `json:"logsAudit,omitempty"`

	// EncryptionUseAWSOwnedKey governs the encryption key type for this profile.
	// nil = not enforced.
	EncryptionUseAWSOwnedKey *bool `json:"encryptionUseAWSOwnedKey,omitempty"`

	// StorageType restricts the storage type for this profile (ActiveMQ only).
	// Empty string = not enforced.
	StorageType string `json:"storageType,omitempty"`

	// NamingTemplate is the naming template (e.g. "{namespace}-{name}").
	// MQConfig-only (no KropathConfig.mq equivalent — per D-4 decision rule).
	// Empty string = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// Tags are cloud resource tags for this MQ config profile.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created MQ resources.
	// Additive map merge across MQConfig tiers only.
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to created MQ resources.
	// Additive map merge across MQConfig tiers only.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveMQSection is one tier (mandatory or defaults) of the merged
// MQ governance result written into MQConfig.status.effectiveConfig by the controller.
type EffectiveMQSection struct {
	EngineType               string            `json:"engineType,omitempty"`
	AuthenticationStrategy   string            `json:"authenticationStrategy,omitempty"`
	DeploymentMode           string            `json:"deploymentMode,omitempty"`
	HostInstanceType         string            `json:"hostInstanceType,omitempty"`
	PubliclyAccessible       *bool             `json:"publiclyAccessible,omitempty"`
	AutoMinorVersionUpgrade  *bool             `json:"autoMinorVersionUpgrade,omitempty"`
	LogsGeneral              *bool             `json:"logsGeneral,omitempty"`
	LogsAudit                *bool             `json:"logsAudit,omitempty"`
	EncryptionUseAWSOwnedKey *bool             `json:"encryptionUseAWSOwnedKey,omitempty"`
	StorageType              string            `json:"storageType,omitempty"`
	NamingTemplate           string            `json:"namingTemplate,omitempty"`
	Tags                     map[string]string `json:"tags,omitempty"`
	SyncedLabels             map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations        map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveMQConfig is the merged MQ governance result written into
// MQConfig.status.effectiveConfig by the controller.
type EffectiveMQConfig struct {
	Mandatory EffectiveMQSection `json:"mandatory"`
	Defaults  EffectiveMQSection `json:"defaults"`
}

// MergeMQCascade merges MQ governance fields from all cascade sources
// and returns the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for MQ (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.mq)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.mq)
//	Level 3 — globalMQCfgMandatory    (MQConfig in kro-system, mandatory)
//	Level 4 — localMQCfgMandatory     (MQConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localMQCfgDefaults      (MQConfig in resource namespace, defaults)
//	Level 7 — globalMQCfgDefaults     (MQConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.mq)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.mq)
//
// KropathConfig.mq carries 9 fields (levels 1-2, 8-9): engineType, authenticationStrategy,
// deploymentMode, publiclyAccessible, autoMinorVersionUpgrade, logsGeneral, logsAudit,
// encryptionUseAWSOwnedKey, storageType.
//
// hostInstanceType and namingTemplate are MQConfig-only (levels 3-4, 6-7).
//
// String merge: firstNonEmptyString in priority order.
// Boolean pointer merge: firstNonNilBoolPtr in priority order (nil = not participating).
// Tags: additive union merge across all mandatory levels / all defaults levels.
// SyncedLabels/SyncedAnnotations: additive union from MQConfig levels only (no KropathConfig).
func MergeMQCascade(
	globalKropathMandatory MQKropathSection, // level 1
	localKropathMandatory MQKropathSection, // level 2
	globalMQCfgMandatory MQConfigSection, // level 3
	localMQCfgMandatory MQConfigSection, // level 4
	localMQCfgDefaults MQConfigSection, // level 6
	globalMQCfgDefaults MQConfigSection, // level 7
	localKropathDefaults MQKropathSection, // level 8
	globalKropathDefaults MQKropathSection, // level 9
) EffectiveMQConfig {
	return EffectiveMQConfig{
		Mandatory: EffectiveMQSection{
			// KropathConfig levels (1-2) exist for the 9 org-wide fields.
			EngineType: firstNonEmptyString(
				globalKropathMandatory.EngineType,
				localKropathMandatory.EngineType,
				globalMQCfgMandatory.EngineType,
				localMQCfgMandatory.EngineType,
			),
			AuthenticationStrategy: firstNonEmptyString(
				globalKropathMandatory.AuthenticationStrategy,
				localKropathMandatory.AuthenticationStrategy,
				globalMQCfgMandatory.AuthenticationStrategy,
				localMQCfgMandatory.AuthenticationStrategy,
			),
			DeploymentMode: firstNonEmptyString(
				globalKropathMandatory.DeploymentMode,
				localKropathMandatory.DeploymentMode,
				globalMQCfgMandatory.DeploymentMode,
				localMQCfgMandatory.DeploymentMode,
			),
			// HostInstanceType: MQConfig levels only (3, 4).
			HostInstanceType: firstNonEmptyString(
				globalMQCfgMandatory.HostInstanceType,
				localMQCfgMandatory.HostInstanceType,
			),
			// Boolean pointer fields: nil = not enforced at this level.
			PubliclyAccessible: firstNonNilBoolPtr(
				globalKropathMandatory.PubliclyAccessible,
				localKropathMandatory.PubliclyAccessible,
				globalMQCfgMandatory.PubliclyAccessible,
				localMQCfgMandatory.PubliclyAccessible,
			),
			AutoMinorVersionUpgrade: firstNonNilBoolPtr(
				globalKropathMandatory.AutoMinorVersionUpgrade,
				localKropathMandatory.AutoMinorVersionUpgrade,
				globalMQCfgMandatory.AutoMinorVersionUpgrade,
				localMQCfgMandatory.AutoMinorVersionUpgrade,
			),
			LogsGeneral: firstNonNilBoolPtr(
				globalKropathMandatory.LogsGeneral,
				localKropathMandatory.LogsGeneral,
				globalMQCfgMandatory.LogsGeneral,
				localMQCfgMandatory.LogsGeneral,
			),
			LogsAudit: firstNonNilBoolPtr(
				globalKropathMandatory.LogsAudit,
				localKropathMandatory.LogsAudit,
				globalMQCfgMandatory.LogsAudit,
				localMQCfgMandatory.LogsAudit,
			),
			EncryptionUseAWSOwnedKey: firstNonNilBoolPtr(
				globalKropathMandatory.EncryptionUseAWSOwnedKey,
				localKropathMandatory.EncryptionUseAWSOwnedKey,
				globalMQCfgMandatory.EncryptionUseAWSOwnedKey,
				localMQCfgMandatory.EncryptionUseAWSOwnedKey,
			),
			StorageType: firstNonEmptyString(
				globalKropathMandatory.StorageType,
				localKropathMandatory.StorageType,
				globalMQCfgMandatory.StorageType,
				localMQCfgMandatory.StorageType,
			),
			// NamingTemplate: MQConfig levels only (3, 4).
			NamingTemplate: firstNonEmptyString(
				globalMQCfgMandatory.NamingTemplate,
				localMQCfgMandatory.NamingTemplate,
			),
			// SyncedLabels: additive union from MQConfig levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localMQCfgMandatory.SyncedLabels,
				globalMQCfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localMQCfgMandatory.SyncedAnnotations,
				globalMQCfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localMQCfgMandatory.Tags,
				globalMQCfgMandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
		},
		Defaults: EffectiveMQSection{
			// KropathConfig levels (8-9) exist for the 9 org-wide fields.
			EngineType: firstNonEmptyString(
				localMQCfgDefaults.EngineType,
				globalMQCfgDefaults.EngineType,
				localKropathDefaults.EngineType,
				globalKropathDefaults.EngineType,
			),
			AuthenticationStrategy: firstNonEmptyString(
				localMQCfgDefaults.AuthenticationStrategy,
				globalMQCfgDefaults.AuthenticationStrategy,
				localKropathDefaults.AuthenticationStrategy,
				globalKropathDefaults.AuthenticationStrategy,
			),
			DeploymentMode: firstNonEmptyString(
				localMQCfgDefaults.DeploymentMode,
				globalMQCfgDefaults.DeploymentMode,
				localKropathDefaults.DeploymentMode,
				globalKropathDefaults.DeploymentMode,
			),
			// HostInstanceType: MQConfig levels only (6, 7).
			HostInstanceType: firstNonEmptyString(
				localMQCfgDefaults.HostInstanceType,
				globalMQCfgDefaults.HostInstanceType,
			),
			PubliclyAccessible: firstNonNilBoolPtr(
				localMQCfgDefaults.PubliclyAccessible,
				globalMQCfgDefaults.PubliclyAccessible,
				localKropathDefaults.PubliclyAccessible,
				globalKropathDefaults.PubliclyAccessible,
			),
			AutoMinorVersionUpgrade: firstNonNilBoolPtr(
				localMQCfgDefaults.AutoMinorVersionUpgrade,
				globalMQCfgDefaults.AutoMinorVersionUpgrade,
				localKropathDefaults.AutoMinorVersionUpgrade,
				globalKropathDefaults.AutoMinorVersionUpgrade,
			),
			LogsGeneral: firstNonNilBoolPtr(
				localMQCfgDefaults.LogsGeneral,
				globalMQCfgDefaults.LogsGeneral,
				localKropathDefaults.LogsGeneral,
				globalKropathDefaults.LogsGeneral,
			),
			LogsAudit: firstNonNilBoolPtr(
				localMQCfgDefaults.LogsAudit,
				globalMQCfgDefaults.LogsAudit,
				localKropathDefaults.LogsAudit,
				globalKropathDefaults.LogsAudit,
			),
			EncryptionUseAWSOwnedKey: firstNonNilBoolPtr(
				localMQCfgDefaults.EncryptionUseAWSOwnedKey,
				globalMQCfgDefaults.EncryptionUseAWSOwnedKey,
				localKropathDefaults.EncryptionUseAWSOwnedKey,
				globalKropathDefaults.EncryptionUseAWSOwnedKey,
			),
			StorageType: firstNonEmptyString(
				localMQCfgDefaults.StorageType,
				globalMQCfgDefaults.StorageType,
				localKropathDefaults.StorageType,
				globalKropathDefaults.StorageType,
			),
			// NamingTemplate: MQConfig levels only (6, 7).
			NamingTemplate: firstNonEmptyString(
				localMQCfgDefaults.NamingTemplate,
				globalMQCfgDefaults.NamingTemplate,
			),
			// SyncedLabels: additive union from MQConfig levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalMQCfgDefaults.SyncedLabels,
				localMQCfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalMQCfgDefaults.SyncedAnnotations,
				localMQCfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalMQCfgDefaults.Tags,
				localMQCfgDefaults.Tags,
			),
		},
	}
}
