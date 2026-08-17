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

// MemoryDBKropathSection holds the MemoryDB-family governance fields from
// KropathConfig.spec.mandatory.memorydb / .defaults.memorydb (ADR-015 §3.5).
//
// Only three of the twelve MemoryDBConfig fields have KropathConfig equivalents:
// tlsEnabled, allowedNodeTypes, and snapshotRetentionLimit (levels 1–2, 8–9).
// The remaining nine fields (kmsKeyArn, nodeType, engineVersion, numReplicasPerShard,
// autoMinorVersionUpgrade, namingTemplate, tags, syncedLabels, syncedAnnotations) have
// no KropathConfig.memorydb equivalent — they are per-profile governance only.
//
// Zero value of each field is the permissive sentinel (false / nil / 0 = not enforced).
type MemoryDBKropathSection struct {
	// TLSEnabled enforces TLS on all MemoryDB clusters org-wide.
	// false (zero value) = not enforced at this level.
	TLSEnabled bool `json:"tlsEnabled,omitempty"`

	// AllowedNodeTypes restricts node types to an org-wide allowlist.
	// nil / empty slice = no restriction.
	AllowedNodeTypes []string `json:"allowedNodeTypes,omitempty"`

	// SnapshotRetentionLimit is the org-wide minimum snapshot retention in days.
	// 0 (zero value) = not enforced at this level.
	SnapshotRetentionLimit int64 `json:"snapshotRetentionLimit,omitempty"`

	// Tags are tier-level cloud resource tags merged from KropathConfig.spec.mandatory.tags /
	// .defaults.tags — the generic cross-family tags, not a memorydb-scoped field.
	// nil / empty map = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// MemoryDBConfigSection holds the MemoryDB governance fields from
// MemoryDBConfig.spec.mandatory or MemoryDBConfig.spec.defaults
// (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type MemoryDBConfigSection struct {
	// TLSEnabled enforces TLS on all clusters. false = not enforced.
	TLSEnabled bool `json:"tlsEnabled,omitempty"`

	// KMSKeyARN forces a specific KMS key for encryption.
	// Empty string = not enforced.
	// No KropathConfig.memorydb equivalent — key management is per-profile only.
	KMSKeyARN string `json:"kmsKeyArn,omitempty"`

	// NodeType forces a specific node type (e.g. "db.r7g.large").
	// Empty string = not enforced.
	// No KropathConfig.memorydb equivalent — node selection is per-profile only.
	NodeType string `json:"nodeType,omitempty"`

	// EngineVersion forces a specific MemoryDB engine version (e.g. "7.1.0").
	// Empty string = not enforced.
	// No KropathConfig.memorydb equivalent — version selection is per-profile only.
	EngineVersion string `json:"engineVersion,omitempty"`

	// AllowedNodeTypes restricts node types to an allowlist.
	// nil / empty slice = no restriction.
	AllowedNodeTypes []string `json:"allowedNodeTypes,omitempty"`

	// NumReplicasPerShard is the minimum replica count per shard.
	// 0 (zero value) = not enforced.
	// No KropathConfig.memorydb equivalent — HA settings are per-profile only.
	NumReplicasPerShard int64 `json:"numReplicasPerShard,omitempty"`

	// SnapshotRetentionLimit is the minimum snapshot retention in days.
	// 0 (zero value) = not enforced.
	SnapshotRetentionLimit int64 `json:"snapshotRetentionLimit,omitempty"`

	// AutoMinorVersionUpgrade forces automatic minor version upgrades.
	// false = not enforced.
	// No KropathConfig.memorydb equivalent — version policy is per-profile only.
	AutoMinorVersionUpgrade bool `json:"autoMinorVersionUpgrade,omitempty"`

	// NamingTemplate is the naming template (e.g. "{namespace}-{name}").
	// Empty string = not enforced.
	// No KropathConfig.memorydb equivalent — consistent with the SQS/EC precedent.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// Tags are cloud resource tags for this MemoryDB config profile.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created resources.
	// Additive map merge across MemoryDBConfig tiers only.
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to created resources.
	// Additive map merge across MemoryDBConfig tiers only.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveMemoryDBSection is one tier (mandatory or defaults) of the merged
// MemoryDB governance result written into MemoryDBConfig.status.effectiveConfig.
type EffectiveMemoryDBSection struct {
	TLSEnabled              bool              `json:"tlsEnabled,omitempty"`
	KMSKeyARN               string            `json:"kmsKeyArn,omitempty"`
	NodeType                string            `json:"nodeType,omitempty"`
	EngineVersion           string            `json:"engineVersion,omitempty"`
	AllowedNodeTypes        []string          `json:"allowedNodeTypes,omitempty"`
	NumReplicasPerShard     int64             `json:"numReplicasPerShard,omitempty"`
	SnapshotRetentionLimit  int64             `json:"snapshotRetentionLimit,omitempty"`
	AutoMinorVersionUpgrade bool              `json:"autoMinorVersionUpgrade,omitempty"`
	NamingTemplate          string            `json:"namingTemplate,omitempty"`
	Tags                    map[string]string `json:"tags,omitempty"`
	SyncedLabels            map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations       map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveMemoryDBConfig is the merged MemoryDB governance result written into
// MemoryDBConfig.status.effectiveConfig by the controller.
type EffectiveMemoryDBConfig struct {
	Mandatory EffectiveMemoryDBSection `json:"mandatory"`
	Defaults  EffectiveMemoryDBSection `json:"defaults"`
}

// MergeMemoryDBCascade merges MemoryDB governance fields from all cascade sources
// and returns the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for MemoryDB (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.memorydb)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.memorydb)
//	Level 3 — globalMDBCfgMandatory   (MemoryDBConfig in kro-system, mandatory)
//	Level 4 — localMDBCfgMandatory    (MemoryDBConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localMDBCfgDefaults     (MemoryDBConfig in resource namespace, defaults)
//	Level 7 — globalMDBCfgDefaults    (MemoryDBConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.memorydb)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.memorydb)
//
// Only tlsEnabled, allowedNodeTypes, and snapshotRetentionLimit have KropathConfig.memorydb
// equivalents (levels 1-2, 8-9). The remaining nine fields fall through to levels 3-7 only.
//
// Boolean merge: firstTrue in priority order (lowest level number wins).
// Integer merge: firstNonZeroInt64 in priority order.
// String merge: firstNonEmptyString in priority order.
// Slice merge: firstNonEmptyStrings in priority order (for allowedNodeTypes).
// Tags: additive union merge across all mandatory levels, all defaults levels.
// SyncedLabels/SyncedAnnotations: additive union from MemoryDBConfig levels only (no KropathConfig).
// NamingTemplate: governed only at MemoryDBConfig levels (3-4 mandatory, 6-7 defaults).
func MergeMemoryDBCascade(
	globalKropathMandatory MemoryDBKropathSection, // level 1
	localKropathMandatory MemoryDBKropathSection, // level 2
	globalMDBCfgMandatory MemoryDBConfigSection, // level 3
	localMDBCfgMandatory MemoryDBConfigSection, // level 4
	localMDBCfgDefaults MemoryDBConfigSection, // level 6
	globalMDBCfgDefaults MemoryDBConfigSection, // level 7
	localKropathDefaults MemoryDBKropathSection, // level 8
	globalKropathDefaults MemoryDBKropathSection, // level 9
) EffectiveMemoryDBConfig {
	return EffectiveMemoryDBConfig{
		Mandatory: EffectiveMemoryDBSection{
			// tlsEnabled: KropathConfig levels (1-2) exist for this field.
			TLSEnabled: firstTrue(
				globalKropathMandatory.TLSEnabled,
				localKropathMandatory.TLSEnabled,
				globalMDBCfgMandatory.TLSEnabled,
				localMDBCfgMandatory.TLSEnabled,
			),
			// No KropathConfig.memorydb levels for kmsKeyArn.
			KMSKeyARN: firstNonEmptyString(
				globalMDBCfgMandatory.KMSKeyARN,
				localMDBCfgMandatory.KMSKeyARN,
			),
			// No KropathConfig.memorydb levels for nodeType.
			NodeType: firstNonEmptyString(
				globalMDBCfgMandatory.NodeType,
				localMDBCfgMandatory.NodeType,
			),
			// No KropathConfig.memorydb levels for engineVersion.
			EngineVersion: firstNonEmptyString(
				globalMDBCfgMandatory.EngineVersion,
				localMDBCfgMandatory.EngineVersion,
			),
			// allowedNodeTypes: KropathConfig levels (1-2) exist for this field.
			AllowedNodeTypes: firstNonEmptyStrings(
				globalKropathMandatory.AllowedNodeTypes,
				localKropathMandatory.AllowedNodeTypes,
				globalMDBCfgMandatory.AllowedNodeTypes,
				localMDBCfgMandatory.AllowedNodeTypes,
			),
			// No KropathConfig.memorydb levels for numReplicasPerShard.
			NumReplicasPerShard: firstNonZeroInt64(
				globalMDBCfgMandatory.NumReplicasPerShard,
				localMDBCfgMandatory.NumReplicasPerShard,
			),
			// snapshotRetentionLimit: KropathConfig levels (1-2) exist for this field.
			SnapshotRetentionLimit: firstNonZeroInt64(
				globalKropathMandatory.SnapshotRetentionLimit,
				localKropathMandatory.SnapshotRetentionLimit,
				globalMDBCfgMandatory.SnapshotRetentionLimit,
				localMDBCfgMandatory.SnapshotRetentionLimit,
			),
			// No KropathConfig.memorydb levels for autoMinorVersionUpgrade.
			AutoMinorVersionUpgrade: firstTrue(
				globalMDBCfgMandatory.AutoMinorVersionUpgrade,
				localMDBCfgMandatory.AutoMinorVersionUpgrade,
			),
			// NamingTemplate: MemoryDBConfig levels only (3, 4).
			NamingTemplate: firstNonEmptyString(
				globalMDBCfgMandatory.NamingTemplate,
				localMDBCfgMandatory.NamingTemplate,
			),
			// SyncedLabels: additive union from MemoryDBConfig levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localMDBCfgMandatory.SyncedLabels,
				globalMDBCfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localMDBCfgMandatory.SyncedAnnotations,
				globalMDBCfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localMDBCfgMandatory.Tags,
				globalMDBCfgMandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
		},
		Defaults: EffectiveMemoryDBSection{
			// tlsEnabled: KropathConfig levels (8-9) exist for this field.
			TLSEnabled: firstTrue(
				localMDBCfgDefaults.TLSEnabled,
				globalMDBCfgDefaults.TLSEnabled,
				localKropathDefaults.TLSEnabled,
				globalKropathDefaults.TLSEnabled,
			),
			// No KropathConfig.memorydb levels for kmsKeyArn.
			KMSKeyARN: firstNonEmptyString(
				localMDBCfgDefaults.KMSKeyARN,
				globalMDBCfgDefaults.KMSKeyARN,
			),
			// No KropathConfig.memorydb levels for nodeType.
			NodeType: firstNonEmptyString(
				localMDBCfgDefaults.NodeType,
				globalMDBCfgDefaults.NodeType,
			),
			// No KropathConfig.memorydb levels for engineVersion.
			EngineVersion: firstNonEmptyString(
				localMDBCfgDefaults.EngineVersion,
				globalMDBCfgDefaults.EngineVersion,
			),
			// allowedNodeTypes: KropathConfig levels (8-9) exist for this field.
			AllowedNodeTypes: firstNonEmptyStrings(
				localMDBCfgDefaults.AllowedNodeTypes,
				globalMDBCfgDefaults.AllowedNodeTypes,
				localKropathDefaults.AllowedNodeTypes,
				globalKropathDefaults.AllowedNodeTypes,
			),
			// No KropathConfig.memorydb levels for numReplicasPerShard.
			NumReplicasPerShard: firstNonZeroInt64(
				localMDBCfgDefaults.NumReplicasPerShard,
				globalMDBCfgDefaults.NumReplicasPerShard,
			),
			// snapshotRetentionLimit: KropathConfig levels (8-9) exist for this field.
			SnapshotRetentionLimit: firstNonZeroInt64(
				localMDBCfgDefaults.SnapshotRetentionLimit,
				globalMDBCfgDefaults.SnapshotRetentionLimit,
				localKropathDefaults.SnapshotRetentionLimit,
				globalKropathDefaults.SnapshotRetentionLimit,
			),
			// No KropathConfig.memorydb levels for autoMinorVersionUpgrade.
			AutoMinorVersionUpgrade: firstTrue(
				localMDBCfgDefaults.AutoMinorVersionUpgrade,
				globalMDBCfgDefaults.AutoMinorVersionUpgrade,
			),
			// NamingTemplate: MemoryDBConfig levels only (6, 7).
			NamingTemplate: firstNonEmptyString(
				localMDBCfgDefaults.NamingTemplate,
				globalMDBCfgDefaults.NamingTemplate,
			),
			// SyncedLabels: additive union from MemoryDBConfig levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalMDBCfgDefaults.SyncedLabels,
				localMDBCfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalMDBCfgDefaults.SyncedAnnotations,
				localMDBCfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalMDBCfgDefaults.Tags,
				localMDBCfgDefaults.Tags,
			),
		},
	}
}

// ValidateMemoryDBNodeType checks the nodeType/allowedNodeTypes cross-field constraint.
// When mandatory.nodeType is non-empty AND mandatory.allowedNodeTypes is non-empty,
// nodeType must be in the allowedNodeTypes list.
//
// Returns (valid=true, "") when the constraint is satisfied or does not apply.
// Returns (valid=false, message) when nodeType is not in allowedNodeTypes.
func ValidateMemoryDBNodeType(mandatory EffectiveMemoryDBSection) (bool, string) {
	if mandatory.NodeType == "" || len(mandatory.AllowedNodeTypes) == 0 {
		return true, ""
	}
	for _, allowed := range mandatory.AllowedNodeTypes {
		if mandatory.NodeType == allowed {
			return true, ""
		}
	}
	return false, fmt.Sprintf(
		"nodeType %q is not in allowedNodeTypes %v",
		mandatory.NodeType,
		mandatory.AllowedNodeTypes,
	)
}
