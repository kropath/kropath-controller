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

// ReplicaRegion is a cross-region replication target for a Secrets Manager secret.
// It mirrors the ACK SecretsManager Secret.spec.replicaRegions ReplicaRegionType shape.
type ReplicaRegion struct {
	// Region is the AWS region code for the replica (e.g. "us-west-2").
	Region string `json:"region,omitempty"`
	// KmsKeyID is the KMS key to use in the replica region.
	// Empty string = use the region's aws/secretsmanager default key.
	KmsKeyID string `json:"kmsKeyID,omitempty"`
}

// SMKropathSection holds the Secrets Manager governance fields from
// KropathConfig.spec.mandatory.secretsManager / .defaults.secretsManager (ADR-015 §3.5).
//
// Only kmsKeyID is governed at the KropathConfig level. Fields such as
// replicaRegions, forceOverwriteReplicaSecret, and namingTemplate are not
// in KropathConfig because they vary by compliance tier or are too operational
// for org-wide enforcement (family design §8).
//
// Zero value of each field is the permissive sentinel (not enforced).
type SMKropathSection struct {
	// KmsKeyID is the org-wide KMS key to mandate for all secrets.
	// Empty string = not enforced; non-empty = org-wide encryption key mandate.
	KmsKeyID string `json:"kmsKeyID,omitempty"`

	// Tags are tier-level cloud resource tags from KropathConfig.spec.mandatory.tags
	// or .defaults.tags. Populated by the reconciler from the tier-level field so that
	// tag cascade flows through MergeSecretsManagerCascade alongside the SM-specific fields.
	// nil / empty map = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// SMConfigSection holds the Secrets Manager governance fields from
// SecretsManagerConfig.spec.mandatory or SecretsManagerConfig.spec.defaults
// (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type SMConfigSection struct {
	// KmsKeyID is the KMS key ARN, key ID, or alias to use for encrypting secrets.
	// Empty string = not enforced (mandatory) or use AWS default aws/secretsmanager (defaults).
	KmsKeyID string `json:"kmsKeyID,omitempty"`

	// ReplicaRegions is the set of cross-region replication targets.
	// Zero-value sentinel: nil / empty slice (not enforced / no replication).
	//
	// Merge semantics: PRIORITY REPLACEMENT — the first non-empty array wins.
	// This differs from map fields (tags, syncedLabels) which use additive union merge.
	// Replication targets represent a complete DR strategy; an org-wide policy and a
	// namespace-level policy must not be merged additively (family design OD-1, Option B).
	ReplicaRegions []ReplicaRegion `json:"replicaRegions,omitempty"`

	// ForceOverwriteReplicaSecret controls whether to overwrite a pre-existing secret
	// with the same name in a replica region.
	// Zero-value sentinel: false (permissive — do not overwrite).
	ForceOverwriteReplicaSecret bool `json:"forceOverwriteReplicaSecret,omitempty"`

	// NamingTemplate is the secret naming template (e.g. "{namespace}-{name}").
	// Governed only at SecretsManagerConfig levels 3-4 (mandatory) and 6-7 (defaults).
	// KropathConfig.secretsManager has no namingTemplate field.
	// Empty string = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created secret resources.
	// Additive map merge across SecretsManagerConfig tiers only.
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to created secret resources.
	// Additive map merge across SecretsManagerConfig tiers only.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`

	// Tags are cloud resource tags for this SecretsManagerConfig profile.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// EffectiveSMSection is one tier (mandatory or defaults) of the merged Secrets Manager
// governance result written into SecretsManagerConfig.status.effectiveConfig by the controller.
type EffectiveSMSection struct {
	KmsKeyID                    string            `json:"kmsKeyID,omitempty"`
	ReplicaRegions              []ReplicaRegion   `json:"replicaRegions,omitempty"`
	ForceOverwriteReplicaSecret bool              `json:"forceOverwriteReplicaSecret,omitempty"`
	NamingTemplate              string            `json:"namingTemplate,omitempty"`
	SyncedLabels                map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations           map[string]string `json:"syncedAnnotations,omitempty"`
	Tags                        map[string]string `json:"tags,omitempty"`
}

// EffectiveSMConfig is the merged Secrets Manager governance result written into
// SecretsManagerConfig.status.effectiveConfig by the controller.
type EffectiveSMConfig struct {
	Mandatory EffectiveSMSection `json:"mandatory"`
	Defaults  EffectiveSMSection `json:"defaults"`
}

// MergeSecretsManagerCascade merges Secrets Manager governance fields from all cascade
// sources and returns the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for Secrets Manager (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.secretsManager)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.secretsManager)
//	Level 3 — globalSMCfgMandatory    (SecretsManagerConfig in kro-system, mandatory)
//	Level 4 — localSMCfgMandatory     (SecretsManagerConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localSMCfgDefaults      (SecretsManagerConfig in resource namespace, defaults)
//	Level 7 — globalSMCfgDefaults     (SecretsManagerConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.secretsManager)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.secretsManager)
//
// Scalar merge (kmsKeyID, namingTemplate): firstNonEmptyString in priority order.
// Boolean merge (forceOverwriteReplicaSecret): firstTrue in priority order.
// Array merge (replicaRegions): firstNonEmptyReplicaRegions — PRIORITY REPLACEMENT.
//   The first non-empty array wins; arrays are NOT merged additively. This models a
//   complete DR strategy rather than an additive set (family design OD-1, Option B).
//   kmsKeyID: levels 1–4 mandatory, 6–9 defaults.
//   replicaRegions: levels 3–4 mandatory, 6–7 defaults (not in KropathConfig).
//   forceOverwriteReplicaSecret: levels 3–4 mandatory, 6–7 defaults (not in KropathConfig).
//   namingTemplate: levels 3–4 mandatory, 6–7 defaults (not in KropathConfig).
// Tags: additive union merge across all four mandatory/defaults levels; lower level number wins on key conflict.
// SyncedLabels/SyncedAnnotations: additive union from SecretsManagerConfig levels only (not KropathConfig).
func MergeSecretsManagerCascade(
	globalKropathMandatory SMKropathSection, // level 1
	localKropathMandatory SMKropathSection, // level 2
	globalSMCfgMandatory SMConfigSection, // level 3
	localSMCfgMandatory SMConfigSection, // level 4
	localSMCfgDefaults SMConfigSection, // level 6
	globalSMCfgDefaults SMConfigSection, // level 7
	localKropathDefaults SMKropathSection, // level 8
	globalKropathDefaults SMKropathSection, // level 9
) EffectiveSMConfig {
	return EffectiveSMConfig{
		Mandatory: EffectiveSMSection{
			// kmsKeyID: levels 1, 2, 3, 4 (KropathConfig wins over SecretsManagerConfig).
			KmsKeyID: firstNonEmptyString(
				globalKropathMandatory.KmsKeyID, // level 1
				localKropathMandatory.KmsKeyID,  // level 2
				globalSMCfgMandatory.KmsKeyID,   // level 3
				localSMCfgMandatory.KmsKeyID,    // level 4
			),
			// replicaRegions: levels 3, 4 only (not in KropathConfig).
			// Priority replacement: first non-empty array wins; NOT additive.
			ReplicaRegions: firstNonEmptyReplicaRegions(
				globalSMCfgMandatory.ReplicaRegions, // level 3
				localSMCfgMandatory.ReplicaRegions,  // level 4
			),
			// forceOverwriteReplicaSecret: levels 3, 4 only (not in KropathConfig).
			ForceOverwriteReplicaSecret: firstTrue(
				globalSMCfgMandatory.ForceOverwriteReplicaSecret, // level 3
				localSMCfgMandatory.ForceOverwriteReplicaSecret,  // level 4
			),
			// namingTemplate: levels 3, 4 only (KropathConfig has no namingTemplate for SM).
			NamingTemplate: firstNonEmptyString(
				globalSMCfgMandatory.NamingTemplate, // level 3
				localSMCfgMandatory.NamingTemplate,  // level 4
			),
			// SyncedLabels: additive union from SecretsManagerConfig levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localSMCfgMandatory.SyncedLabels,
				globalSMCfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localSMCfgMandatory.SyncedAnnotations,
				globalSMCfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localSMCfgMandatory.Tags,    // level 4 (lowest priority)
				globalSMCfgMandatory.Tags,   // level 3
				localKropathMandatory.Tags,  // level 2
				globalKropathMandatory.Tags, // level 1 (highest priority)
			),
		},
		Defaults: EffectiveSMSection{
			// kmsKeyID: levels 6, 7, 8, 9.
			KmsKeyID: firstNonEmptyString(
				localSMCfgDefaults.KmsKeyID,    // level 6
				globalSMCfgDefaults.KmsKeyID,   // level 7
				localKropathDefaults.KmsKeyID,  // level 8
				globalKropathDefaults.KmsKeyID, // level 9
			),
			// replicaRegions: levels 6, 7 only (not in KropathConfig).
			// Priority replacement: first non-empty array wins.
			ReplicaRegions: firstNonEmptyReplicaRegions(
				localSMCfgDefaults.ReplicaRegions,  // level 6
				globalSMCfgDefaults.ReplicaRegions, // level 7
			),
			// forceOverwriteReplicaSecret: levels 6, 7 only (not in KropathConfig).
			ForceOverwriteReplicaSecret: firstTrue(
				localSMCfgDefaults.ForceOverwriteReplicaSecret,  // level 6
				globalSMCfgDefaults.ForceOverwriteReplicaSecret, // level 7
			),
			// namingTemplate: levels 6, 7 only (KropathConfig has no namingTemplate for SM).
			NamingTemplate: firstNonEmptyString(
				localSMCfgDefaults.NamingTemplate,  // level 6
				globalSMCfgDefaults.NamingTemplate, // level 7
			),
			// SyncedLabels: additive union from SecretsManagerConfig levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalSMCfgDefaults.SyncedLabels,
				localSMCfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalSMCfgDefaults.SyncedAnnotations,
				localSMCfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags, // level 9 (lowest priority)
				localKropathDefaults.Tags,  // level 8
				globalSMCfgDefaults.Tags,   // level 7
				localSMCfgDefaults.Tags,    // level 6 (highest priority)
			),
		},
	}
}

// firstNonEmptyReplicaRegions returns a defensive copy of the first non-nil, non-empty
// ReplicaRegion slice from candidates.
//
// This implements PRIORITY REPLACEMENT semantics for replicaRegions: the first non-empty
// array wins and is returned as-is. Arrays are never merged additively because replica
// regions represent a complete DR strategy — merging them from multiple governance tiers
// would combine incompatible DR policies (family design OD-1, Option B).
//
// A defensive copy is returned to prevent the caller from aliasing a cached CR slice.
func firstNonEmptyReplicaRegions(candidates ...[]ReplicaRegion) []ReplicaRegion {
	for _, rr := range candidates {
		if len(rr) > 0 {
			out := make([]ReplicaRegion, len(rr))
			copy(out, rr)
			return out
		}
	}
	return nil
}
