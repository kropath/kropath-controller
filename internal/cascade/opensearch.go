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

// OpenSearchKropathSection holds the OpenSearch-family governance fields from
// KropathConfig.spec.mandatory.opensearch / .defaults.opensearch (ADR-015 §3.5).
//
// Only four of the twelve OpenSearchConfig fields have KropathConfig equivalents:
// encryptionAtRestEnabled, nodeToNodeEncryptionEnabled, enforceHTTPS, tlsSecurityPolicy.
// The remaining eight fields (advancedSecurityEnabled, engineVersion, autoTuneDesiredState,
// standbyReplicas, namingTemplate, tags, syncedLabels, syncedAnnotations) have
// no KropathConfig.opensearch equivalent — they are per-profile governance only.
//
// Zero value of each field is the permissive sentinel (false = not enforced, "" = not enforced).
type OpenSearchKropathSection struct {
	// EncryptionAtRestEnabled enforces encryption at rest org-wide.
	// false (zero value) = not enforced at this level.
	EncryptionAtRestEnabled bool `json:"encryptionAtRestEnabled,omitempty"`

	// NodeToNodeEncryptionEnabled enforces node-to-node encryption org-wide.
	// false (zero value) = not enforced at this level.
	NodeToNodeEncryptionEnabled bool `json:"nodeToNodeEncryptionEnabled,omitempty"`

	// EnforceHTTPS enforces HTTPS-only access org-wide.
	// false (zero value) = not enforced at this level.
	EnforceHTTPS bool `json:"enforceHTTPS,omitempty"`

	// TLSSecurityPolicy is the org-wide minimum TLS policy (e.g. "Policy-Min-TLS-1-2-2019-07").
	// Empty string (zero value) = not enforced at this level.
	TLSSecurityPolicy string `json:"tlsSecurityPolicy,omitempty"`

	// Tags are tier-level cloud resource tags merged from KropathConfig.spec.mandatory.tags /
	// .defaults.tags — the generic cross-family tags, not an opensearch-scoped field.
	// nil / empty map = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// OpenSearchConfigSection holds the OpenSearch governance fields from
// OpenSearchConfig.spec.mandatory or OpenSearchConfig.spec.defaults
// (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type OpenSearchConfigSection struct {
	// EncryptionAtRestEnabled enforces encryption at rest. false = not enforced.
	EncryptionAtRestEnabled bool `json:"encryptionAtRestEnabled,omitempty"`

	// NodeToNodeEncryptionEnabled enforces node-to-node encryption. false = not enforced.
	NodeToNodeEncryptionEnabled bool `json:"nodeToNodeEncryptionEnabled,omitempty"`

	// EnforceHTTPS enforces HTTPS-only access. false = not enforced.
	EnforceHTTPS bool `json:"enforceHTTPS,omitempty"`

	// TLSSecurityPolicy is the minimum TLS policy. Empty string = not enforced.
	TLSSecurityPolicy string `json:"tlsSecurityPolicy,omitempty"`

	// AdvancedSecurityEnabled enforces fine-grained access control (FGAC). false = not enforced.
	// No KropathConfig.opensearch equivalent — opt-in security feature, per-profile only.
	AdvancedSecurityEnabled bool `json:"advancedSecurityEnabled,omitempty"`

	// EngineVersion restricts the OpenSearch engine version (e.g. "OpenSearch_2.9").
	// Empty string = not enforced.
	// No KropathConfig.opensearch equivalent — engine version is per-profile governance.
	EngineVersion string `json:"engineVersion,omitempty"`

	// AutoTuneDesiredState restricts the auto-tune state ("ENABLED" or "DISABLED").
	// Empty string = not enforced.
	// No KropathConfig.opensearch equivalent — auto-tune is per-profile governance.
	AutoTuneDesiredState string `json:"autoTuneDesiredState,omitempty"`

	// StandbyReplicas restricts standby replicas for serverless ("ENABLED" or "DISABLED").
	// Empty string = not enforced.
	// No KropathConfig.opensearch equivalent — serverless standby is per-profile governance.
	StandbyReplicas string `json:"standbyReplicas,omitempty"`

	// NamingTemplate is the naming template (e.g. "{namespace}-{name}").
	// Empty string = not enforced.
	// No KropathConfig.opensearch equivalent — consistent with the SQS/ElastiCache precedent.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// Tags are cloud resource tags for this OpenSearch config profile.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created resources.
	// Additive map merge across OpenSearchConfig tiers only.
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to created resources.
	// Additive map merge across OpenSearchConfig tiers only.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveOpenSearchSection is one tier (mandatory or defaults) of the merged
// OpenSearch governance result written into OpenSearchConfig.status.effectiveConfig.
type EffectiveOpenSearchSection struct {
	EncryptionAtRestEnabled     bool              `json:"encryptionAtRestEnabled,omitempty"`
	NodeToNodeEncryptionEnabled bool              `json:"nodeToNodeEncryptionEnabled,omitempty"`
	EnforceHTTPS                bool              `json:"enforceHTTPS,omitempty"`
	TLSSecurityPolicy           string            `json:"tlsSecurityPolicy,omitempty"`
	AdvancedSecurityEnabled     bool              `json:"advancedSecurityEnabled,omitempty"`
	EngineVersion               string            `json:"engineVersion,omitempty"`
	AutoTuneDesiredState        string            `json:"autoTuneDesiredState,omitempty"`
	StandbyReplicas             string            `json:"standbyReplicas,omitempty"`
	NamingTemplate              string            `json:"namingTemplate,omitempty"`
	Tags                        map[string]string `json:"tags,omitempty"`
	SyncedLabels                map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations           map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveOpenSearchConfig is the merged OpenSearch governance result written into
// OpenSearchConfig.status.effectiveConfig by the controller.
type EffectiveOpenSearchConfig struct {
	Mandatory EffectiveOpenSearchSection `json:"mandatory"`
	Defaults  EffectiveOpenSearchSection `json:"defaults"`
}

// MergeOpenSearchCascade merges OpenSearch governance fields from all cascade sources
// and returns the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for OpenSearch (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.opensearch)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.opensearch)
//	Level 3 — globalOSCfgMandatory    (OpenSearchConfig in kro-system, mandatory)
//	Level 4 — localOSCfgMandatory     (OpenSearchConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localOSCfgDefaults      (OpenSearchConfig in resource namespace, defaults)
//	Level 7 — globalOSCfgDefaults     (OpenSearchConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.opensearch)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.opensearch)
//
// Only encryptionAtRestEnabled, nodeToNodeEncryptionEnabled, enforceHTTPS, and tlsSecurityPolicy
// have KropathConfig.opensearch equivalents (levels 1-2, 8-9). The remaining eight fields fall
// through to levels 3-7 only.
//
// Boolean merge: firstTrue in priority order (lowest level number wins).
// String merge: firstNonEmptyString in priority order.
// Tags: additive union merge across all mandatory levels, all defaults levels.
// SyncedLabels/SyncedAnnotations: additive union from OpenSearchConfig levels only (no KropathConfig).
// NamingTemplate: governed only at OpenSearchConfig levels (3-4 mandatory, 6-7 defaults).
func MergeOpenSearchCascade(
	globalKropathMandatory OpenSearchKropathSection, // level 1
	localKropathMandatory OpenSearchKropathSection, // level 2
	globalOSCfgMandatory OpenSearchConfigSection, // level 3
	localOSCfgMandatory OpenSearchConfigSection, // level 4
	localOSCfgDefaults OpenSearchConfigSection, // level 6
	globalOSCfgDefaults OpenSearchConfigSection, // level 7
	localKropathDefaults OpenSearchKropathSection, // level 8
	globalKropathDefaults OpenSearchKropathSection, // level 9
) EffectiveOpenSearchConfig {
	return EffectiveOpenSearchConfig{
		Mandatory: EffectiveOpenSearchSection{
			// KropathConfig levels (1-2) exist for the four security fields.
			EncryptionAtRestEnabled: firstTrue(
				globalKropathMandatory.EncryptionAtRestEnabled,
				localKropathMandatory.EncryptionAtRestEnabled,
				globalOSCfgMandatory.EncryptionAtRestEnabled,
				localOSCfgMandatory.EncryptionAtRestEnabled,
			),
			NodeToNodeEncryptionEnabled: firstTrue(
				globalKropathMandatory.NodeToNodeEncryptionEnabled,
				localKropathMandatory.NodeToNodeEncryptionEnabled,
				globalOSCfgMandatory.NodeToNodeEncryptionEnabled,
				localOSCfgMandatory.NodeToNodeEncryptionEnabled,
			),
			EnforceHTTPS: firstTrue(
				globalKropathMandatory.EnforceHTTPS,
				localKropathMandatory.EnforceHTTPS,
				globalOSCfgMandatory.EnforceHTTPS,
				localOSCfgMandatory.EnforceHTTPS,
			),
			TLSSecurityPolicy: firstNonEmptyString(
				globalKropathMandatory.TLSSecurityPolicy,
				localKropathMandatory.TLSSecurityPolicy,
				globalOSCfgMandatory.TLSSecurityPolicy,
				localOSCfgMandatory.TLSSecurityPolicy,
			),
			// No KropathConfig.opensearch levels for the remaining eight fields.
			AdvancedSecurityEnabled: firstTrue(
				globalOSCfgMandatory.AdvancedSecurityEnabled,
				localOSCfgMandatory.AdvancedSecurityEnabled,
			),
			EngineVersion: firstNonEmptyString(
				globalOSCfgMandatory.EngineVersion,
				localOSCfgMandatory.EngineVersion,
			),
			AutoTuneDesiredState: firstNonEmptyString(
				globalOSCfgMandatory.AutoTuneDesiredState,
				localOSCfgMandatory.AutoTuneDesiredState,
			),
			StandbyReplicas: firstNonEmptyString(
				globalOSCfgMandatory.StandbyReplicas,
				localOSCfgMandatory.StandbyReplicas,
			),
			// NamingTemplate: OpenSearchConfig levels only (3, 4).
			NamingTemplate: firstNonEmptyString(
				globalOSCfgMandatory.NamingTemplate,
				localOSCfgMandatory.NamingTemplate,
			),
			// SyncedLabels: additive union from OpenSearchConfig levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localOSCfgMandatory.SyncedLabels,
				globalOSCfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localOSCfgMandatory.SyncedAnnotations,
				globalOSCfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localOSCfgMandatory.Tags,
				globalOSCfgMandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
		},
		Defaults: EffectiveOpenSearchSection{
			// KropathConfig levels (8-9) exist for the four security fields.
			EncryptionAtRestEnabled: firstTrue(
				localOSCfgDefaults.EncryptionAtRestEnabled,
				globalOSCfgDefaults.EncryptionAtRestEnabled,
				localKropathDefaults.EncryptionAtRestEnabled,
				globalKropathDefaults.EncryptionAtRestEnabled,
			),
			NodeToNodeEncryptionEnabled: firstTrue(
				localOSCfgDefaults.NodeToNodeEncryptionEnabled,
				globalOSCfgDefaults.NodeToNodeEncryptionEnabled,
				localKropathDefaults.NodeToNodeEncryptionEnabled,
				globalKropathDefaults.NodeToNodeEncryptionEnabled,
			),
			EnforceHTTPS: firstTrue(
				localOSCfgDefaults.EnforceHTTPS,
				globalOSCfgDefaults.EnforceHTTPS,
				localKropathDefaults.EnforceHTTPS,
				globalKropathDefaults.EnforceHTTPS,
			),
			TLSSecurityPolicy: firstNonEmptyString(
				localOSCfgDefaults.TLSSecurityPolicy,
				globalOSCfgDefaults.TLSSecurityPolicy,
				localKropathDefaults.TLSSecurityPolicy,
				globalKropathDefaults.TLSSecurityPolicy,
			),
			// No KropathConfig.opensearch levels for the remaining eight fields.
			AdvancedSecurityEnabled: firstTrue(
				localOSCfgDefaults.AdvancedSecurityEnabled,
				globalOSCfgDefaults.AdvancedSecurityEnabled,
			),
			EngineVersion: firstNonEmptyString(
				localOSCfgDefaults.EngineVersion,
				globalOSCfgDefaults.EngineVersion,
			),
			AutoTuneDesiredState: firstNonEmptyString(
				localOSCfgDefaults.AutoTuneDesiredState,
				globalOSCfgDefaults.AutoTuneDesiredState,
			),
			StandbyReplicas: firstNonEmptyString(
				localOSCfgDefaults.StandbyReplicas,
				globalOSCfgDefaults.StandbyReplicas,
			),
			// NamingTemplate: OpenSearchConfig levels only (6, 7).
			NamingTemplate: firstNonEmptyString(
				localOSCfgDefaults.NamingTemplate,
				globalOSCfgDefaults.NamingTemplate,
			),
			// SyncedLabels: additive union from OpenSearchConfig levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalOSCfgDefaults.SyncedLabels,
				localOSCfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalOSCfgDefaults.SyncedAnnotations,
				localOSCfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalOSCfgDefaults.Tags,
				localOSCfgDefaults.Tags,
			),
		},
	}
}
