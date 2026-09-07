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

// BackupKropathSection holds the Backup-family governance fields from
// KropathConfig.spec.mandatory.backup / .defaults.backup (ADR-015 §3.5).
// PLUS the tier-level tags from KropathConfig.spec.mandatory.tags (populated
// by the reconciler so that tag cascade flows through MergeBackupCascade).
//
// iamRoleARN is intentionally omitted — it is a per-selection choice (family design §8).
// scheduleExpression, targetBackupVaultName are per-rule choices, also omitted.
//
// Zero value of each field is the permissive sentinel (not enforced).
type BackupKropathSection struct {
	// EncryptionKeyARN enforces org-wide KMS key for all vaults when non-empty.
	// Empty string (zero value) = not enforced.
	EncryptionKeyARN string `json:"encryptionKeyARN,omitempty"`

	// VaultLockMinRetentionDays enforces minimum Vault Lock retention when > 0.
	// 0 (zero value) = not enforced.
	VaultLockMinRetentionDays int64 `json:"vaultLockMinRetentionDays,omitempty"`

	// VaultLockMaxRetentionDays enforces maximum Vault Lock retention when > 0.
	// 0 (zero value) = not enforced.
	VaultLockMaxRetentionDays int64 `json:"vaultLockMaxRetentionDays,omitempty"`

	// EnableContinuousBackup enforces PITR org-wide when true.
	// false (zero value) = not enforced.
	EnableContinuousBackup bool `json:"enableContinuousBackup,omitempty"`

	// DefaultLifecycleDeleteAfterDays enforces minimum retention when > 0.
	// 0 (zero value) = not enforced.
	DefaultLifecycleDeleteAfterDays int64 `json:"defaultLifecycleDeleteAfterDays,omitempty"`

	// DefaultLifecycleMoveToColdStorageAfterDays enforces cold storage transition when > 0.
	// 0 (zero value) = not enforced.
	DefaultLifecycleMoveToColdStorageAfterDays int64 `json:"defaultLifecycleMoveToColdStorageAfterDays,omitempty"`

	// EnableMalwareScan enforces malware scanning org-wide when true.
	// false (zero value) = not enforced.
	EnableMalwareScan bool `json:"enableMalwareScan,omitempty"`

	// ScannerRoleARN enforces a specific scanner IAM role when non-empty.
	// Empty string (zero value) = not enforced.
	ScannerRoleARN string `json:"scannerRoleARN,omitempty"`

	// Tags are tier-level cloud resource tags from KropathConfig.spec.mandatory.tags.
	// Populated by the reconciler from the tier-level field, not from spec.mandatory.backup.
	// nil / empty map (zero value) = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// BackupConfigSection holds the Backup governance fields from BackupConfig.spec.mandatory
// or BackupConfig.spec.defaults (per-type ResourceConfig).
//
// Zero value of each field is the permissive sentinel (not enforced).
type BackupConfigSection struct {
	EncryptionKeyARN                           string            `json:"encryptionKeyARN,omitempty"`
	VaultLockMinRetentionDays                  int64             `json:"vaultLockMinRetentionDays,omitempty"`
	VaultLockMaxRetentionDays                  int64             `json:"vaultLockMaxRetentionDays,omitempty"`
	EnableContinuousBackup                     bool              `json:"enableContinuousBackup,omitempty"`
	DefaultLifecycleDeleteAfterDays            int64             `json:"defaultLifecycleDeleteAfterDays,omitempty"`
	DefaultLifecycleMoveToColdStorageAfterDays int64             `json:"defaultLifecycleMoveToColdStorageAfterDays,omitempty"`
	EnableMalwareScan                          bool              `json:"enableMalwareScan,omitempty"`
	ScannerRoleARN                             string            `json:"scannerRoleARN,omitempty"`
	// IAMRoleARN is a per-selection choice; it is NOT in KropathConfig.
	IAMRoleARN     string            `json:"iamRoleARN,omitempty"`
	NamingTemplate string            `json:"namingTemplate,omitempty"`
	Tags           map[string]string `json:"tags,omitempty"`
	SyncedLabels      map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveBackupSection is one tier (mandatory or defaults) of the merged Backup
// governance result written into BackupConfig.status.effectiveConfig by the controller.
type EffectiveBackupSection struct {
	EncryptionKeyARN                           string            `json:"encryptionKeyARN,omitempty"`
	VaultLockMinRetentionDays                  int64             `json:"vaultLockMinRetentionDays,omitempty"`
	VaultLockMaxRetentionDays                  int64             `json:"vaultLockMaxRetentionDays,omitempty"`
	EnableContinuousBackup                     bool              `json:"enableContinuousBackup,omitempty"`
	DefaultLifecycleDeleteAfterDays            int64             `json:"defaultLifecycleDeleteAfterDays,omitempty"`
	DefaultLifecycleMoveToColdStorageAfterDays int64             `json:"defaultLifecycleMoveToColdStorageAfterDays,omitempty"`
	EnableMalwareScan                          bool              `json:"enableMalwareScan,omitempty"`
	ScannerRoleARN                             string            `json:"scannerRoleARN,omitempty"`
	IAMRoleARN                                 string            `json:"iamRoleARN,omitempty"`
	NamingTemplate                             string            `json:"namingTemplate,omitempty"`
	Tags                                       map[string]string `json:"tags,omitempty"`
	SyncedLabels                               map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations                          map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveBackupConfig is the merged Backup governance result written into
// BackupConfig.status.effectiveConfig by the controller.
type EffectiveBackupConfig struct {
	Mandatory EffectiveBackupSection `json:"mandatory"`
	Defaults  EffectiveBackupSection `json:"defaults"`
}

// MergeBackupCascade merges Backup governance fields from all cascade sources and
// returns the effective configuration to be written to status.effectiveConfig.
//
// The ten-level priority chain (ADR-015 §5.3) for Backup fields:
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace)
//	Level 3 — globalBackupCfgMandatory (BackupConfig in kro-system)
//	Level 4 — localBackupCfgMandatory  (BackupConfig in resource namespace)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localBackupCfgDefaults   (BackupConfig in resource namespace)
//	Level 7 — globalBackupCfgDefaults  (BackupConfig in kro-system)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system)
//
// For mandatory (levels 1–4): first non-zero value in priority order wins.
// For defaults (levels 6–9): first non-zero value in priority order wins.
// For tags: union merge across all sources; lower level numbers win on key conflicts.
//
// iamRoleARN and namingTemplate are not in KropathConfig (per-selection/per-rule choices
// per family design §8), so they only appear at levels 3–4 (mandatory) and 6–7 (defaults).
// All other fields appear at all four mandatory/defaults levels.
// BackupKropathSection.Tags carries the tier-level KropathConfig.mandatory.tags
// (populated by the reconciler).
func MergeBackupCascade(
	// Mandatory inputs (highest → lowest priority)
	globalKropathMandatory BackupKropathSection, // level 1
	localKropathMandatory BackupKropathSection,  // level 2
	globalBackupCfgMandatory BackupConfigSection, // level 3
	localBackupCfgMandatory BackupConfigSection,  // level 4
	// Defaults inputs (highest → lowest priority)
	localBackupCfgDefaults BackupConfigSection,   // level 6
	globalBackupCfgDefaults BackupConfigSection,  // level 7
	localKropathDefaults BackupKropathSection,    // level 8
	globalKropathDefaults BackupKropathSection,   // level 9
) EffectiveBackupConfig {
	return EffectiveBackupConfig{
		Mandatory: EffectiveBackupSection{
			EncryptionKeyARN: firstNonEmptyString(
				globalKropathMandatory.EncryptionKeyARN,  // level 1
				localKropathMandatory.EncryptionKeyARN,   // level 2
				globalBackupCfgMandatory.EncryptionKeyARN, // level 3
				localBackupCfgMandatory.EncryptionKeyARN,  // level 4
			),
			VaultLockMinRetentionDays: firstNonZeroInt64(
				globalKropathMandatory.VaultLockMinRetentionDays,  // level 1
				localKropathMandatory.VaultLockMinRetentionDays,   // level 2
				globalBackupCfgMandatory.VaultLockMinRetentionDays, // level 3
				localBackupCfgMandatory.VaultLockMinRetentionDays,  // level 4
			),
			VaultLockMaxRetentionDays: firstNonZeroInt64(
				globalKropathMandatory.VaultLockMaxRetentionDays,  // level 1
				localKropathMandatory.VaultLockMaxRetentionDays,   // level 2
				globalBackupCfgMandatory.VaultLockMaxRetentionDays, // level 3
				localBackupCfgMandatory.VaultLockMaxRetentionDays,  // level 4
			),
			EnableContinuousBackup: firstTrue(
				globalKropathMandatory.EnableContinuousBackup,  // level 1
				localKropathMandatory.EnableContinuousBackup,   // level 2
				globalBackupCfgMandatory.EnableContinuousBackup, // level 3
				localBackupCfgMandatory.EnableContinuousBackup,  // level 4
			),
			DefaultLifecycleDeleteAfterDays: firstNonZeroInt64(
				globalKropathMandatory.DefaultLifecycleDeleteAfterDays,  // level 1
				localKropathMandatory.DefaultLifecycleDeleteAfterDays,   // level 2
				globalBackupCfgMandatory.DefaultLifecycleDeleteAfterDays, // level 3
				localBackupCfgMandatory.DefaultLifecycleDeleteAfterDays,  // level 4
			),
			DefaultLifecycleMoveToColdStorageAfterDays: firstNonZeroInt64(
				globalKropathMandatory.DefaultLifecycleMoveToColdStorageAfterDays,  // level 1
				localKropathMandatory.DefaultLifecycleMoveToColdStorageAfterDays,   // level 2
				globalBackupCfgMandatory.DefaultLifecycleMoveToColdStorageAfterDays, // level 3
				localBackupCfgMandatory.DefaultLifecycleMoveToColdStorageAfterDays,  // level 4
			),
			EnableMalwareScan: firstTrue(
				globalKropathMandatory.EnableMalwareScan,  // level 1
				localKropathMandatory.EnableMalwareScan,   // level 2
				globalBackupCfgMandatory.EnableMalwareScan, // level 3
				localBackupCfgMandatory.EnableMalwareScan,  // level 4
			),
			ScannerRoleARN: firstNonEmptyString(
				globalKropathMandatory.ScannerRoleARN,  // level 1
				localKropathMandatory.ScannerRoleARN,   // level 2
				globalBackupCfgMandatory.ScannerRoleARN, // level 3
				localBackupCfgMandatory.ScannerRoleARN,  // level 4
			),
			// iamRoleARN not in KropathConfig: levels 3 and 4 only.
			IAMRoleARN: firstNonEmptyString(
				globalBackupCfgMandatory.IAMRoleARN, // level 3
				localBackupCfgMandatory.IAMRoleARN,  // level 4
			),
			// namingTemplate not in KropathConfig: levels 3 and 4 only.
			NamingTemplate: firstNonEmptyString(
				globalBackupCfgMandatory.NamingTemplate, // level 3
				localBackupCfgMandatory.NamingTemplate,  // level 4
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflicts.
			Tags: mergeMaps(
				localBackupCfgMandatory.Tags,   // level 4 (lowest priority, set first)
				globalBackupCfgMandatory.Tags,  // level 3
				localKropathMandatory.Tags,     // level 2
				globalKropathMandatory.Tags,    // level 1 (highest priority, last to write)
			),
			SyncedLabels: mergeMaps(
				localBackupCfgMandatory.SyncedLabels,
				globalBackupCfgMandatory.SyncedLabels,
			),
			SyncedAnnotations: mergeMaps(
				localBackupCfgMandatory.SyncedAnnotations,
				globalBackupCfgMandatory.SyncedAnnotations,
			),
		},
		Defaults: EffectiveBackupSection{
			EncryptionKeyARN: firstNonEmptyString(
				localBackupCfgDefaults.EncryptionKeyARN,  // level 6
				globalBackupCfgDefaults.EncryptionKeyARN, // level 7
				localKropathDefaults.EncryptionKeyARN,    // level 8
				globalKropathDefaults.EncryptionKeyARN,   // level 9
			),
			VaultLockMinRetentionDays: firstNonZeroInt64(
				localBackupCfgDefaults.VaultLockMinRetentionDays,  // level 6
				globalBackupCfgDefaults.VaultLockMinRetentionDays, // level 7
				localKropathDefaults.VaultLockMinRetentionDays,    // level 8
				globalKropathDefaults.VaultLockMinRetentionDays,   // level 9
			),
			VaultLockMaxRetentionDays: firstNonZeroInt64(
				localBackupCfgDefaults.VaultLockMaxRetentionDays,  // level 6
				globalBackupCfgDefaults.VaultLockMaxRetentionDays, // level 7
				localKropathDefaults.VaultLockMaxRetentionDays,    // level 8
				globalKropathDefaults.VaultLockMaxRetentionDays,   // level 9
			),
			EnableContinuousBackup: firstTrue(
				localBackupCfgDefaults.EnableContinuousBackup,  // level 6
				globalBackupCfgDefaults.EnableContinuousBackup, // level 7
				localKropathDefaults.EnableContinuousBackup,    // level 8
				globalKropathDefaults.EnableContinuousBackup,   // level 9
			),
			DefaultLifecycleDeleteAfterDays: firstNonZeroInt64(
				localBackupCfgDefaults.DefaultLifecycleDeleteAfterDays,  // level 6
				globalBackupCfgDefaults.DefaultLifecycleDeleteAfterDays, // level 7
				localKropathDefaults.DefaultLifecycleDeleteAfterDays,    // level 8
				globalKropathDefaults.DefaultLifecycleDeleteAfterDays,   // level 9
			),
			DefaultLifecycleMoveToColdStorageAfterDays: firstNonZeroInt64(
				localBackupCfgDefaults.DefaultLifecycleMoveToColdStorageAfterDays,  // level 6
				globalBackupCfgDefaults.DefaultLifecycleMoveToColdStorageAfterDays, // level 7
				localKropathDefaults.DefaultLifecycleMoveToColdStorageAfterDays,    // level 8
				globalKropathDefaults.DefaultLifecycleMoveToColdStorageAfterDays,   // level 9
			),
			EnableMalwareScan: firstTrue(
				localBackupCfgDefaults.EnableMalwareScan,  // level 6
				globalBackupCfgDefaults.EnableMalwareScan, // level 7
				localKropathDefaults.EnableMalwareScan,    // level 8
				globalKropathDefaults.EnableMalwareScan,   // level 9
			),
			ScannerRoleARN: firstNonEmptyString(
				localBackupCfgDefaults.ScannerRoleARN,  // level 6
				globalBackupCfgDefaults.ScannerRoleARN, // level 7
				localKropathDefaults.ScannerRoleARN,    // level 8
				globalKropathDefaults.ScannerRoleARN,   // level 9
			),
			// iamRoleARN not in KropathConfig: levels 6 and 7 only.
			IAMRoleARN: firstNonEmptyString(
				localBackupCfgDefaults.IAMRoleARN,  // level 6
				globalBackupCfgDefaults.IAMRoleARN, // level 7
			),
			// namingTemplate not in KropathConfig: levels 6 and 7 only.
			NamingTemplate: firstNonEmptyString(
				localBackupCfgDefaults.NamingTemplate,  // level 6
				globalBackupCfgDefaults.NamingTemplate, // level 7
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflicts.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,   // level 9 (lowest priority)
				localKropathDefaults.Tags,    // level 8
				globalBackupCfgDefaults.Tags, // level 7
				localBackupCfgDefaults.Tags,  // level 6 (highest priority)
			),
			SyncedLabels: mergeMaps(
				globalBackupCfgDefaults.SyncedLabels,
				localBackupCfgDefaults.SyncedLabels,
			),
			SyncedAnnotations: mergeMaps(
				globalBackupCfgDefaults.SyncedAnnotations,
				localBackupCfgDefaults.SyncedAnnotations,
			),
		},
	}
}
