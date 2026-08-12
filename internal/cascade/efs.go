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

// EFSKropathSection holds the EFS-family governance fields that flow from
// KropathConfig into the EFS cascade.
//
// No EFS-specific scalar fields exist in KropathConfig (spec §KropathConfig additions).
// Only org-wide tags are plumbed here; the reconciler populates Tags from
// KropathConfig.spec.mandatory.tags / .defaults.tags before calling MergeEFSCascade.
type EFSKropathSection struct {
	// Tags are tier-level cloud resource tags sourced from KropathConfig org-wide tags.
	// nil / empty map = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// EFSConfigSection holds the EFS governance fields from EFSConfig.spec.mandatory or
// EFSConfig.spec.defaults (per-type ResourceConfig, ADR-015 §3.5).
//
// Boolean pointer fields use nil = not set (falls through); false = explicitly disabled.
// String fields use "" = not set (zero-value sentinel); non-empty = enforced.
type EFSConfigSection struct {
	// Encrypted controls SSE for all EFS file systems in mandatory tier.
	// nil = not set; true = enforce encryption; false = explicitly disabled.
	Encrypted *bool `json:"encrypted,omitempty"`

	// KmsKeyId is the KMS key ARN for SSE-KMS. Empty string = not enforced (AWS-managed).
	KmsKeyId string `json:"kmsKeyId,omitempty"`

	// PerformanceMode is the EFS performance mode ("generalPurpose" or "maxIO").
	// Empty string = not enforced.
	PerformanceMode string `json:"performanceMode,omitempty"`

	// ThroughputMode is the EFS throughput mode ("bursting", "elastic", or "provisioned").
	// Empty string = not enforced.
	ThroughputMode string `json:"throughputMode,omitempty"`

	// BackupEnabled controls AWS Backup for EFS file systems.
	// nil = not set; true = enforce backups; false = explicitly disabled.
	BackupEnabled *bool `json:"backupEnabled,omitempty"`

	// TransitionToIA is the lifecycle policy for transitioning files to IA storage class.
	// Empty string = no IA transition enforced.
	TransitionToIA string `json:"transitionToIA,omitempty"`

	// TransitionToArchive is the lifecycle policy for transitioning files to Archive storage class.
	// Empty string = no Archive transition enforced.
	TransitionToArchive string `json:"transitionToArchive,omitempty"`

	// TransitionToPrimaryStorage is the lifecycle policy for transitioning files back to primary storage.
	// Empty string = no back-transition enforced.
	TransitionToPrimaryStorage string `json:"transitionToPrimaryStorage,omitempty"`

	// ReplicationOverwriteProtection controls whether replication overwrite is enabled.
	// Empty string = not enforced.
	ReplicationOverwriteProtection string `json:"replicationOverwriteProtection,omitempty"`

	// Tags are cloud resource tags for this EFS config profile.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created EFS resources.
	// Additive map merge across EFSConfig tiers only.
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to created EFS resources.
	// Additive map merge across EFSConfig tiers only.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveEFSSection is one tier (mandatory or defaults) of the merged EFS governance
// result written into EFSConfig.status.effectiveConfig by the controller.
type EffectiveEFSSection struct {
	Encrypted                      *bool             `json:"encrypted,omitempty"`
	KmsKeyId                       string            `json:"kmsKeyId,omitempty"`
	PerformanceMode                string            `json:"performanceMode,omitempty"`
	ThroughputMode                 string            `json:"throughputMode,omitempty"`
	BackupEnabled                  *bool             `json:"backupEnabled,omitempty"`
	TransitionToIA                 string            `json:"transitionToIA,omitempty"`
	TransitionToArchive            string            `json:"transitionToArchive,omitempty"`
	TransitionToPrimaryStorage     string            `json:"transitionToPrimaryStorage,omitempty"`
	ReplicationOverwriteProtection string            `json:"replicationOverwriteProtection,omitempty"`
	Tags                           map[string]string `json:"tags,omitempty"`
	SyncedLabels                   map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations              map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveEFSConfig is the merged EFS governance result written into
// EFSConfig.status.effectiveConfig by the controller.
type EffectiveEFSConfig struct {
	Mandatory EffectiveEFSSection `json:"mandatory"`
	Defaults  EffectiveEFSSection `json:"defaults"`
}

// MergeEFSCascade merges EFS governance fields from all cascade sources and
// returns the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for EFS (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory org-wide tags)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory org-wide tags)
//	Level 3 — globalEFSCfgMandatory   (EFSConfig in kro-system, mandatory)
//	Level 4 — localEFSCfgMandatory    (EFSConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localEFSCfgDefaults     (EFSConfig in resource namespace, defaults)
//	Level 7 — globalEFSCfgDefaults    (EFSConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults org-wide tags)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults org-wide tags)
//
// *bool merge: firstNonNilBoolPtr in priority order — nil = not set (falls through).
// String merge: firstNonEmptyString in priority order (lowest level number wins).
// Tags: additive union merge across all four mandatory levels, all four defaults levels.
// SyncedLabels/SyncedAnnotations: additive union from EFSConfig levels only (no KropathConfig).
//
// NOTE: No EFS-specific scalar fields exist in KropathConfig. The caller populates
// EFSKropathSection.Tags from KropathConfig.spec.mandatory.tags / .defaults.tags
// so that org-wide tags flow through the merge alongside EFS-specific fields.
func MergeEFSCascade(
	globalKropathMandatory EFSKropathSection, // level 1
	localKropathMandatory EFSKropathSection, // level 2
	globalEFSCfgMandatory EFSConfigSection, // level 3
	localEFSCfgMandatory EFSConfigSection, // level 4
	localEFSCfgDefaults EFSConfigSection, // level 6
	globalEFSCfgDefaults EFSConfigSection, // level 7
	localKropathDefaults EFSKropathSection, // level 8
	globalKropathDefaults EFSKropathSection, // level 9
) EffectiveEFSConfig {
	return EffectiveEFSConfig{
		Mandatory: EffectiveEFSSection{
			Encrypted: firstNonNilBoolPtr(
				globalEFSCfgMandatory.Encrypted,
				localEFSCfgMandatory.Encrypted,
			),
			KmsKeyId: firstNonEmptyString(
				globalEFSCfgMandatory.KmsKeyId,
				localEFSCfgMandatory.KmsKeyId,
			),
			PerformanceMode: firstNonEmptyString(
				globalEFSCfgMandatory.PerformanceMode,
				localEFSCfgMandatory.PerformanceMode,
			),
			ThroughputMode: firstNonEmptyString(
				globalEFSCfgMandatory.ThroughputMode,
				localEFSCfgMandatory.ThroughputMode,
			),
			BackupEnabled: firstNonNilBoolPtr(
				globalEFSCfgMandatory.BackupEnabled,
				localEFSCfgMandatory.BackupEnabled,
			),
			TransitionToIA: firstNonEmptyString(
				globalEFSCfgMandatory.TransitionToIA,
				localEFSCfgMandatory.TransitionToIA,
			),
			TransitionToArchive: firstNonEmptyString(
				globalEFSCfgMandatory.TransitionToArchive,
				localEFSCfgMandatory.TransitionToArchive,
			),
			TransitionToPrimaryStorage: firstNonEmptyString(
				globalEFSCfgMandatory.TransitionToPrimaryStorage,
				localEFSCfgMandatory.TransitionToPrimaryStorage,
			),
			ReplicationOverwriteProtection: firstNonEmptyString(
				globalEFSCfgMandatory.ReplicationOverwriteProtection,
				localEFSCfgMandatory.ReplicationOverwriteProtection,
			),
			// SyncedLabels: additive union from EFSConfig levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localEFSCfgMandatory.SyncedLabels,
				globalEFSCfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localEFSCfgMandatory.SyncedAnnotations,
				globalEFSCfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localEFSCfgMandatory.Tags,
				globalEFSCfgMandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
		},
		Defaults: EffectiveEFSSection{
			Encrypted: firstNonNilBoolPtr(
				localEFSCfgDefaults.Encrypted,
				globalEFSCfgDefaults.Encrypted,
			),
			KmsKeyId: firstNonEmptyString(
				localEFSCfgDefaults.KmsKeyId,
				globalEFSCfgDefaults.KmsKeyId,
			),
			PerformanceMode: firstNonEmptyString(
				localEFSCfgDefaults.PerformanceMode,
				globalEFSCfgDefaults.PerformanceMode,
			),
			ThroughputMode: firstNonEmptyString(
				localEFSCfgDefaults.ThroughputMode,
				globalEFSCfgDefaults.ThroughputMode,
			),
			BackupEnabled: firstNonNilBoolPtr(
				localEFSCfgDefaults.BackupEnabled,
				globalEFSCfgDefaults.BackupEnabled,
			),
			TransitionToIA: firstNonEmptyString(
				localEFSCfgDefaults.TransitionToIA,
				globalEFSCfgDefaults.TransitionToIA,
			),
			TransitionToArchive: firstNonEmptyString(
				localEFSCfgDefaults.TransitionToArchive,
				globalEFSCfgDefaults.TransitionToArchive,
			),
			TransitionToPrimaryStorage: firstNonEmptyString(
				localEFSCfgDefaults.TransitionToPrimaryStorage,
				globalEFSCfgDefaults.TransitionToPrimaryStorage,
			),
			ReplicationOverwriteProtection: firstNonEmptyString(
				localEFSCfgDefaults.ReplicationOverwriteProtection,
				globalEFSCfgDefaults.ReplicationOverwriteProtection,
			),
			// SyncedLabels: additive union from EFSConfig levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalEFSCfgDefaults.SyncedLabels,
				localEFSCfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalEFSCfgDefaults.SyncedAnnotations,
				localEFSCfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalEFSCfgDefaults.Tags,
				localEFSCfgDefaults.Tags,
			),
		},
	}
}
