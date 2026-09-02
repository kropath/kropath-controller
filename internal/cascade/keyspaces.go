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

// KeyspacesKropathSection holds the Keyspaces governance fields that appear in
// KropathConfig.spec.mandatory and KropathConfig.spec.defaults (levels 1-2, 8-9).
// namingTemplate is KeyspacesConfig-only and does NOT appear here.
type KeyspacesKropathSection struct {
	ReplicationStrategy string `json:"replicationStrategy,omitempty"`
	ThroughputMode      string `json:"throughputMode,omitempty"`
	EncryptionType      string `json:"encryptionType,omitempty"`
	// Boolean sentinel: false = not enforced.
	PointInTimeRecovery bool `json:"pointInTimeRecovery,omitempty"`
	TTLEnabled          bool `json:"ttlEnabled,omitempty"`
	Tags                map[string]string `json:"tags,omitempty"`
}

// KeyspacesConfigSection holds the Keyspaces governance fields that appear in
// KeyspacesConfig.spec.mandatory and KeyspacesConfig.spec.defaults (levels 3-4, 6-7).
type KeyspacesConfigSection struct {
	ReplicationStrategy string `json:"replicationStrategy,omitempty"`
	ThroughputMode      string `json:"throughputMode,omitempty"`
	EncryptionType      string `json:"encryptionType,omitempty"`
	// Boolean sentinel: false = not enforced.
	PointInTimeRecovery bool `json:"pointInTimeRecovery,omitempty"`
	TTLEnabled          bool `json:"ttlEnabled,omitempty"`
	NamingTemplate      string `json:"namingTemplate,omitempty"`
	Tags                map[string]string `json:"tags,omitempty"`
	SyncedLabels        map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations   map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveKeyspacesSection is the resolved tier within EffectiveKeyspacesConfig.
type EffectiveKeyspacesSection struct {
	ReplicationStrategy string            `json:"replicationStrategy,omitempty"`
	ThroughputMode      string            `json:"throughputMode,omitempty"`
	EncryptionType      string            `json:"encryptionType,omitempty"`
	PointInTimeRecovery bool              `json:"pointInTimeRecovery,omitempty"`
	TTLEnabled          bool              `json:"ttlEnabled,omitempty"`
	NamingTemplate      string            `json:"namingTemplate,omitempty"`
	Tags                map[string]string `json:"tags,omitempty"`
	SyncedLabels        map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations   map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveKeyspacesConfig is written to KeyspacesConfig.status.effectiveConfig.
type EffectiveKeyspacesConfig struct {
	Mandatory EffectiveKeyspacesSection `json:"mandatory,omitempty"`
	Defaults  EffectiveKeyspacesSection `json:"defaults,omitempty"`
}

// MergeKeyspacesCascade resolves the 9-level effectiveConfig cascade for Keyspaces.
//
// Priority order (lower level number wins):
//
//	Mandatory tier:
//	  L1 globalKropathMandatory → L2 localKropathMandatory →
//	  L3 globalKeyspacesCfgMandatory → L4 localKeyspacesCfgMandatory
//
//	Defaults tier:
//	  L6 localKeyspacesCfgDefaults → L7 globalKeyspacesCfgDefaults →
//	  L8 localKropathDefaults → L9 globalKropathDefaults
//
// Tags are merged (union); lower level number wins on key conflict.
func MergeKeyspacesCascade(
	globalKropathMandatory, localKropathMandatory KeyspacesKropathSection,
	globalKeyspacesCfgMandatory, localKeyspacesCfgMandatory KeyspacesConfigSection,
	localKeyspacesCfgDefaults, globalKeyspacesCfgDefaults KeyspacesConfigSection,
	localKropathDefaults, globalKropathDefaults KeyspacesKropathSection,
) EffectiveKeyspacesConfig {
	mandatory := EffectiveKeyspacesSection{
		ReplicationStrategy: firstNonEmptyString(
			globalKropathMandatory.ReplicationStrategy,
			localKropathMandatory.ReplicationStrategy,
			globalKeyspacesCfgMandatory.ReplicationStrategy,
			localKeyspacesCfgMandatory.ReplicationStrategy,
		),
		ThroughputMode: firstNonEmptyString(
			globalKropathMandatory.ThroughputMode,
			localKropathMandatory.ThroughputMode,
			globalKeyspacesCfgMandatory.ThroughputMode,
			localKeyspacesCfgMandatory.ThroughputMode,
		),
		EncryptionType: firstNonEmptyString(
			globalKropathMandatory.EncryptionType,
			localKropathMandatory.EncryptionType,
			globalKeyspacesCfgMandatory.EncryptionType,
			localKeyspacesCfgMandatory.EncryptionType,
		),
		PointInTimeRecovery: firstTrue(
			globalKropathMandatory.PointInTimeRecovery,
			localKropathMandatory.PointInTimeRecovery,
			globalKeyspacesCfgMandatory.PointInTimeRecovery,
			localKeyspacesCfgMandatory.PointInTimeRecovery,
		),
		TTLEnabled: firstTrue(
			globalKropathMandatory.TTLEnabled,
			localKropathMandatory.TTLEnabled,
			globalKeyspacesCfgMandatory.TTLEnabled,
			localKeyspacesCfgMandatory.TTLEnabled,
		),
		NamingTemplate: firstNonEmptyString(
			globalKeyspacesCfgMandatory.NamingTemplate,
			localKeyspacesCfgMandatory.NamingTemplate,
		),
		// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflicts.
		Tags: mergeMaps(
			localKeyspacesCfgMandatory.Tags,  // level 4 (lowest priority)
			globalKeyspacesCfgMandatory.Tags, // level 3
			localKropathMandatory.Tags,       // level 2
			globalKropathMandatory.Tags,      // level 1 (highest priority)
		),
		SyncedLabels: mergeMaps(
			localKeyspacesCfgMandatory.SyncedLabels,
			globalKeyspacesCfgMandatory.SyncedLabels,
		),
		SyncedAnnotations: mergeMaps(
			localKeyspacesCfgMandatory.SyncedAnnotations,
			globalKeyspacesCfgMandatory.SyncedAnnotations,
		),
	}

	defaults := EffectiveKeyspacesSection{
		ReplicationStrategy: firstNonEmptyString(
			localKeyspacesCfgDefaults.ReplicationStrategy,
			globalKeyspacesCfgDefaults.ReplicationStrategy,
			localKropathDefaults.ReplicationStrategy,
			globalKropathDefaults.ReplicationStrategy,
		),
		ThroughputMode: firstNonEmptyString(
			localKeyspacesCfgDefaults.ThroughputMode,
			globalKeyspacesCfgDefaults.ThroughputMode,
			localKropathDefaults.ThroughputMode,
			globalKropathDefaults.ThroughputMode,
		),
		EncryptionType: firstNonEmptyString(
			localKeyspacesCfgDefaults.EncryptionType,
			globalKeyspacesCfgDefaults.EncryptionType,
			localKropathDefaults.EncryptionType,
			globalKropathDefaults.EncryptionType,
		),
		PointInTimeRecovery: firstTrue(
			localKeyspacesCfgDefaults.PointInTimeRecovery,
			globalKeyspacesCfgDefaults.PointInTimeRecovery,
			localKropathDefaults.PointInTimeRecovery,
			globalKropathDefaults.PointInTimeRecovery,
		),
		TTLEnabled: firstTrue(
			localKeyspacesCfgDefaults.TTLEnabled,
			globalKeyspacesCfgDefaults.TTLEnabled,
			localKropathDefaults.TTLEnabled,
			globalKropathDefaults.TTLEnabled,
		),
		NamingTemplate: firstNonEmptyString(
			localKeyspacesCfgDefaults.NamingTemplate,
			globalKeyspacesCfgDefaults.NamingTemplate,
		),
		// Tags: union of all defaults sources; L9 added first, L6 wins on key conflicts.
		Tags: mergeMaps(
			globalKropathDefaults.Tags,       // level 9 (lowest priority)
			localKropathDefaults.Tags,        // level 8
			globalKeyspacesCfgDefaults.Tags,  // level 7
			localKeyspacesCfgDefaults.Tags,   // level 6 (highest priority)
		),
		SyncedLabels: mergeMaps(
			globalKeyspacesCfgDefaults.SyncedLabels,
			localKeyspacesCfgDefaults.SyncedLabels,
		),
		SyncedAnnotations: mergeMaps(
			globalKeyspacesCfgDefaults.SyncedAnnotations,
			localKeyspacesCfgDefaults.SyncedAnnotations,
		),
	}

	return EffectiveKeyspacesConfig{
		Mandatory: mandatory,
		Defaults:  defaults,
	}
}
