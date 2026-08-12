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

// ElastiCacheKropathSection holds the ElastiCache-family governance fields from
// KropathConfig.spec.mandatory.elasticache / .defaults.elasticache (ADR-015 §3.5).
//
// Only two of the eleven ElastiCacheConfig fields have KropathConfig equivalents:
// atRestEncryptionEnabled and transitEncryptionEnabled. The remaining nine fields
// (automaticFailoverEnabled, multiAZEnabled, snapshotRetentionLimit, engine,
// blockNoPasswordUsers, namingTemplate, tags, syncedLabels, syncedAnnotations) have
// no KropathConfig.elasticache equivalent — they are per-profile governance only.
//
// Zero value of each field is the permissive sentinel (false = not enforced).
type ElastiCacheKropathSection struct {
	// AtRestEncryptionEnabled enforces encryption at rest org-wide.
	// false (zero value) = not enforced at this level.
	AtRestEncryptionEnabled bool `json:"atRestEncryptionEnabled,omitempty"`

	// TransitEncryptionEnabled enforces encryption in transit org-wide.
	// false (zero value) = not enforced at this level.
	TransitEncryptionEnabled bool `json:"transitEncryptionEnabled,omitempty"`

	// Tags are tier-level cloud resource tags merged from KropathConfig.spec.mandatory.tags /
	// .defaults.tags — the generic cross-family tags, not an elasticache-scoped field.
	// nil / empty map = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// ElastiCacheConfigSection holds the ElastiCache governance fields from
// ElastiCacheConfig.spec.mandatory or ElastiCacheConfig.spec.defaults
// (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type ElastiCacheConfigSection struct {
	// AtRestEncryptionEnabled enforces encryption at rest. false = not enforced.
	AtRestEncryptionEnabled bool `json:"atRestEncryptionEnabled,omitempty"`

	// TransitEncryptionEnabled enforces encryption in transit. false = not enforced.
	TransitEncryptionEnabled bool `json:"transitEncryptionEnabled,omitempty"`

	// AutomaticFailoverEnabled enforces automatic failover. false = not enforced.
	// No KropathConfig.elasticache equivalent — HA settings are per-profile only.
	AutomaticFailoverEnabled bool `json:"automaticFailoverEnabled,omitempty"`

	// MultiAZEnabled enforces Multi-AZ deployment. false = not enforced.
	// No KropathConfig.elasticache equivalent — HA settings are per-profile only.
	MultiAZEnabled bool `json:"multiAZEnabled,omitempty"`

	// Engine restricts the cache engine ("valkey", "redis", "memcached").
	// Empty string = not enforced.
	// No KropathConfig.elasticache equivalent — engine choice is per-profile governance (OD-2).
	Engine string `json:"engine,omitempty"`

	// BlockNoPasswordUsers prevents creation of users without password authentication.
	// false = not enforced.
	// No KropathConfig.elasticache equivalent — scoped per-profile (OD-4).
	BlockNoPasswordUsers bool `json:"blockNoPasswordUsers,omitempty"`

	// SnapshotRetentionLimit is the minimum snapshot retention in days.
	// 0 (zero value) = not enforced.
	// No KropathConfig.elasticache equivalent — retention is per-profile governance.
	SnapshotRetentionLimit int64 `json:"snapshotRetentionLimit,omitempty"`

	// NamingTemplate is the naming template (e.g. "{namespace}-{name}").
	// Empty string = not enforced.
	// No KropathConfig.elasticache equivalent — consistent with the SQS precedent.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// Tags are cloud resource tags for this ElastiCache config profile.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created resources.
	// Additive map merge across ElastiCacheConfig tiers only.
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to created resources.
	// Additive map merge across ElastiCacheConfig tiers only.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveElastiCacheSection is one tier (mandatory or defaults) of the merged
// ElastiCache governance result written into ElastiCacheConfig.status.effectiveConfig.
type EffectiveElastiCacheSection struct {
	AtRestEncryptionEnabled  bool              `json:"atRestEncryptionEnabled,omitempty"`
	TransitEncryptionEnabled bool              `json:"transitEncryptionEnabled,omitempty"`
	AutomaticFailoverEnabled bool              `json:"automaticFailoverEnabled,omitempty"`
	MultiAZEnabled           bool              `json:"multiAZEnabled,omitempty"`
	Engine                   string            `json:"engine,omitempty"`
	BlockNoPasswordUsers     bool              `json:"blockNoPasswordUsers,omitempty"`
	SnapshotRetentionLimit   int64             `json:"snapshotRetentionLimit,omitempty"`
	NamingTemplate           string            `json:"namingTemplate,omitempty"`
	Tags                     map[string]string `json:"tags,omitempty"`
	SyncedLabels             map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations        map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveElastiCacheConfig is the merged ElastiCache governance result written into
// ElastiCacheConfig.status.effectiveConfig by the controller.
type EffectiveElastiCacheConfig struct {
	Mandatory EffectiveElastiCacheSection `json:"mandatory"`
	Defaults  EffectiveElastiCacheSection `json:"defaults"`
}

// MergeElastiCacheCascade merges ElastiCache governance fields from all cascade sources
// and returns the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for ElastiCache (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.elasticache)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.elasticache)
//	Level 3 — globalECCfgMandatory    (ElastiCacheConfig in kro-system, mandatory)
//	Level 4 — localECCfgMandatory     (ElastiCacheConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localECCfgDefaults      (ElastiCacheConfig in resource namespace, defaults)
//	Level 7 — globalECCfgDefaults     (ElastiCacheConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.elasticache)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.elasticache)
//
// Only atRestEncryptionEnabled and transitEncryptionEnabled have KropathConfig.elasticache
// equivalents (levels 1-2, 8-9). The remaining nine fields fall through to levels 3-7 only.
//
// Boolean merge: firstTrue in priority order (lowest level number wins).
// Integer merge: firstNonZeroInt64 in priority order.
// String merge: firstNonEmptyString in priority order.
// Tags: additive union merge across all mandatory levels, all defaults levels.
// SyncedLabels/SyncedAnnotations: additive union from ElastiCacheConfig levels only (no KropathConfig).
// NamingTemplate: governed only at ElastiCacheConfig levels (3-4 mandatory, 6-7 defaults).
func MergeElastiCacheCascade(
	globalKropathMandatory ElastiCacheKropathSection, // level 1
	localKropathMandatory ElastiCacheKropathSection, // level 2
	globalECCfgMandatory ElastiCacheConfigSection, // level 3
	localECCfgMandatory ElastiCacheConfigSection, // level 4
	localECCfgDefaults ElastiCacheConfigSection, // level 6
	globalECCfgDefaults ElastiCacheConfigSection, // level 7
	localKropathDefaults ElastiCacheKropathSection, // level 8
	globalKropathDefaults ElastiCacheKropathSection, // level 9
) EffectiveElastiCacheConfig {
	return EffectiveElastiCacheConfig{
		Mandatory: EffectiveElastiCacheSection{
			// KropathConfig levels (1-2) exist for atRest and transit only.
			AtRestEncryptionEnabled: firstTrue(
				globalKropathMandatory.AtRestEncryptionEnabled,
				localKropathMandatory.AtRestEncryptionEnabled,
				globalECCfgMandatory.AtRestEncryptionEnabled,
				localECCfgMandatory.AtRestEncryptionEnabled,
			),
			TransitEncryptionEnabled: firstTrue(
				globalKropathMandatory.TransitEncryptionEnabled,
				localKropathMandatory.TransitEncryptionEnabled,
				globalECCfgMandatory.TransitEncryptionEnabled,
				localECCfgMandatory.TransitEncryptionEnabled,
			),
			// No KropathConfig.elasticache levels for the remaining 9 fields.
			AutomaticFailoverEnabled: firstTrue(
				globalECCfgMandatory.AutomaticFailoverEnabled,
				localECCfgMandatory.AutomaticFailoverEnabled,
			),
			MultiAZEnabled: firstTrue(
				globalECCfgMandatory.MultiAZEnabled,
				localECCfgMandatory.MultiAZEnabled,
			),
			Engine: firstNonEmptyString(
				globalECCfgMandatory.Engine,
				localECCfgMandatory.Engine,
			),
			BlockNoPasswordUsers: firstTrue(
				globalECCfgMandatory.BlockNoPasswordUsers,
				localECCfgMandatory.BlockNoPasswordUsers,
			),
			SnapshotRetentionLimit: firstNonZeroInt64(
				globalECCfgMandatory.SnapshotRetentionLimit,
				localECCfgMandatory.SnapshotRetentionLimit,
			),
			// NamingTemplate: ElastiCacheConfig levels only (3, 4).
			NamingTemplate: firstNonEmptyString(
				globalECCfgMandatory.NamingTemplate,
				localECCfgMandatory.NamingTemplate,
			),
			// SyncedLabels: additive union from ElastiCacheConfig levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localECCfgMandatory.SyncedLabels,
				globalECCfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localECCfgMandatory.SyncedAnnotations,
				globalECCfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localECCfgMandatory.Tags,
				globalECCfgMandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
		},
		Defaults: EffectiveElastiCacheSection{
			// KropathConfig levels (8-9) exist for atRest and transit only.
			AtRestEncryptionEnabled: firstTrue(
				localECCfgDefaults.AtRestEncryptionEnabled,
				globalECCfgDefaults.AtRestEncryptionEnabled,
				localKropathDefaults.AtRestEncryptionEnabled,
				globalKropathDefaults.AtRestEncryptionEnabled,
			),
			TransitEncryptionEnabled: firstTrue(
				localECCfgDefaults.TransitEncryptionEnabled,
				globalECCfgDefaults.TransitEncryptionEnabled,
				localKropathDefaults.TransitEncryptionEnabled,
				globalKropathDefaults.TransitEncryptionEnabled,
			),
			// No KropathConfig.elasticache levels for the remaining 9 fields.
			AutomaticFailoverEnabled: firstTrue(
				localECCfgDefaults.AutomaticFailoverEnabled,
				globalECCfgDefaults.AutomaticFailoverEnabled,
			),
			MultiAZEnabled: firstTrue(
				localECCfgDefaults.MultiAZEnabled,
				globalECCfgDefaults.MultiAZEnabled,
			),
			Engine: firstNonEmptyString(
				localECCfgDefaults.Engine,
				globalECCfgDefaults.Engine,
			),
			BlockNoPasswordUsers: firstTrue(
				localECCfgDefaults.BlockNoPasswordUsers,
				globalECCfgDefaults.BlockNoPasswordUsers,
			),
			SnapshotRetentionLimit: firstNonZeroInt64(
				localECCfgDefaults.SnapshotRetentionLimit,
				globalECCfgDefaults.SnapshotRetentionLimit,
			),
			// NamingTemplate: ElastiCacheConfig levels only (6, 7).
			NamingTemplate: firstNonEmptyString(
				localECCfgDefaults.NamingTemplate,
				globalECCfgDefaults.NamingTemplate,
			),
			// SyncedLabels: additive union from ElastiCacheConfig levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalECCfgDefaults.SyncedLabels,
				localECCfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalECCfgDefaults.SyncedAnnotations,
				localECCfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalECCfgDefaults.Tags,
				localECCfgDefaults.Tags,
			),
		},
	}
}
