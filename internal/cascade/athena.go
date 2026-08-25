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

// AthenaKropathSection holds the Athena-family governance fields from
// KropathConfig.spec.mandatory.athena / .defaults.athena (ADR-015 §3.5)
// PLUS the tier-level tags from KropathConfig.spec.mandatory.tags (populated
// by the reconciler so that tag cascade flows through MergeAthenaCascade).
//
// Only three WorkGroup-specific booleans appear at the KropathConfig level —
// bytesScannedCutoffPerQuery, requesterPaysEnabled, resultEncryptionOption,
// resultOutputLocation, engineVersion, tags, syncedLabels, syncedAnnotations,
// and namingTemplate are AthenaConfig-only (spec §KropathConfig Additions).
//
// Zero value of each field is the permissive sentinel (not enforced):
// false for booleans, nil/empty for maps.
type AthenaKropathSection struct {
	// EnforceWorkGroupConfiguration forces workgroup-level settings to override client-level ones.
	// false (zero value) = not enforced; true = enforcement required.
	EnforceWorkGroupConfiguration bool `json:"enforceWorkGroupConfiguration,omitempty"`

	// EnableMinimumEncryptionConfiguration requires a minimum encryption configuration.
	// false (zero value) = not enforced; true = encryption required.
	EnableMinimumEncryptionConfiguration bool `json:"enableMinimumEncryptionConfiguration,omitempty"`

	// PublishCloudWatchMetricsEnabled requires CloudWatch metrics publication.
	// false (zero value) = not enforced; true = metrics required.
	PublishCloudWatchMetricsEnabled bool `json:"publishCloudWatchMetricsEnabled,omitempty"`

	// Tags are tier-level cloud resource tags from KropathConfig.spec.mandatory.tags.
	// Populated by the reconciler from the tier-level field, not from spec.mandatory.athena.
	// nil / empty map = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// AthenaConfigSection holds the Athena governance fields from AthenaConfig.spec.mandatory
// or AthenaConfig.spec.defaults (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type AthenaConfigSection struct {
	// EnforceWorkGroupConfiguration forces workgroup-level settings to override client-level ones.
	// false (zero value) = not enforced.
	EnforceWorkGroupConfiguration bool `json:"enforceWorkGroupConfiguration,omitempty"`

	// EnableMinimumEncryptionConfiguration requires a minimum encryption configuration.
	// false (zero value) = not enforced.
	EnableMinimumEncryptionConfiguration bool `json:"enableMinimumEncryptionConfiguration,omitempty"`

	// PublishCloudWatchMetricsEnabled requires CloudWatch metrics publication.
	// false (zero value) = not enforced.
	PublishCloudWatchMetricsEnabled bool `json:"publishCloudWatchMetricsEnabled,omitempty"`

	// BytesScannedCutoffPerQuery is the maximum bytes scanned per query.
	// 0 (zero value) = not enforced; >0 forces a max bytes limit.
	BytesScannedCutoffPerQuery int64 `json:"bytesScannedCutoffPerQuery,omitempty"`

	// RequesterPaysEnabled enables requester-pays for cross-account query data access.
	// false (zero value) = not enforced.
	RequesterPaysEnabled bool `json:"requesterPaysEnabled,omitempty"`

	// ResultEncryptionOption is the enforced result encryption option.
	// Empty string = not enforced; allowed values: SSE_S3, SSE_KMS, CSE_KMS.
	ResultEncryptionOption string `json:"resultEncryptionOption,omitempty"`

	// ResultOutputLocation is the enforced S3 location for query results.
	// Empty string = not enforced; non-empty forces the output location.
	ResultOutputLocation string `json:"resultOutputLocation,omitempty"`

	// EngineVersion is the enforced Athena engine version.
	// Empty string = not enforced; non-empty forces the engine version.
	EngineVersion string `json:"engineVersion,omitempty"`

	// NamingTemplate is the resource naming template (e.g. "{namespace}-{name}").
	// Governed only at AthenaConfig levels 3–4 (mandatory) and 6–7 (defaults).
	// KropathConfig.athena does NOT carry namingTemplate.
	// Empty string = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// Tags are cloud resource tags for this Athena config profile.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created Athena resources.
	// Additive map merge across AthenaConfig tiers only.
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to created Athena resources.
	// Additive map merge across AthenaConfig tiers only.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveAthenaSection is one tier (mandatory or defaults) of the merged Athena
// governance result written into AthenaConfig.status.effectiveConfig by the controller.
type EffectiveAthenaSection struct {
	EnforceWorkGroupConfiguration        bool              `json:"enforceWorkGroupConfiguration,omitempty"`
	EnableMinimumEncryptionConfiguration bool              `json:"enableMinimumEncryptionConfiguration,omitempty"`
	PublishCloudWatchMetricsEnabled      bool              `json:"publishCloudWatchMetricsEnabled,omitempty"`
	BytesScannedCutoffPerQuery           int64             `json:"bytesScannedCutoffPerQuery,omitempty"`
	RequesterPaysEnabled                 bool              `json:"requesterPaysEnabled,omitempty"`
	ResultEncryptionOption               string            `json:"resultEncryptionOption,omitempty"`
	ResultOutputLocation                 string            `json:"resultOutputLocation,omitempty"`
	EngineVersion                        string            `json:"engineVersion,omitempty"`
	NamingTemplate                       string            `json:"namingTemplate,omitempty"`
	Tags                                 map[string]string `json:"tags,omitempty"`
	SyncedLabels                         map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations                    map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveAthenaConfig is the merged Athena governance result written into
// AthenaConfig.status.effectiveConfig by the controller.
type EffectiveAthenaConfig struct {
	Mandatory EffectiveAthenaSection `json:"mandatory"`
	Defaults  EffectiveAthenaSection `json:"defaults"`
}

// MergeAthenaCascade merges Athena governance fields from all cascade sources and
// returns the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for Athena (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.athena)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.athena)
//	Level 3 — globalAthenaCfgMandatory (AthenaConfig in kro-system, mandatory)
//	Level 4 — localAthenaCfgMandatory  (AthenaConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localAthenaCfgDefaults   (AthenaConfig in resource namespace, defaults)
//	Level 7 — globalAthenaCfgDefaults  (AthenaConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.athena)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.athena)
//
// Boolean fields: firstTrue semantics (false = not enforced; true = enforcement active).
// Integer fields: firstNonZeroInt64 (0 = not enforced).
// String fields: firstNonEmptyString ("" = not enforced).
// NamingTemplate: governed only at AthenaConfig levels (3–4 mandatory, 6–7 defaults).
// Tags: additive union merge across all mandatory levels / all defaults levels.
// SyncedLabels/SyncedAnnotations: additive union from AthenaConfig levels only.
func MergeAthenaCascade(
	globalKropathMandatory AthenaKropathSection, // level 1
	localKropathMandatory AthenaKropathSection, // level 2
	globalAthenaCfgMandatory AthenaConfigSection, // level 3
	localAthenaCfgMandatory AthenaConfigSection, // level 4
	localAthenaCfgDefaults AthenaConfigSection, // level 6
	globalAthenaCfgDefaults AthenaConfigSection, // level 7
	localKropathDefaults AthenaKropathSection, // level 8
	globalKropathDefaults AthenaKropathSection, // level 9
) EffectiveAthenaConfig {
	return EffectiveAthenaConfig{
		Mandatory: EffectiveAthenaSection{
			// enforceWorkGroupConfiguration: boolean — false = not enforced; true = active.
			// KropathConfig levels 1–2 + AthenaConfig levels 3–4 participate.
			EnforceWorkGroupConfiguration: firstTrue(
				globalKropathMandatory.EnforceWorkGroupConfiguration,
				localKropathMandatory.EnforceWorkGroupConfiguration,
				globalAthenaCfgMandatory.EnforceWorkGroupConfiguration,
				localAthenaCfgMandatory.EnforceWorkGroupConfiguration,
			),
			// enableMinimumEncryptionConfiguration: boolean sentinel.
			EnableMinimumEncryptionConfiguration: firstTrue(
				globalKropathMandatory.EnableMinimumEncryptionConfiguration,
				localKropathMandatory.EnableMinimumEncryptionConfiguration,
				globalAthenaCfgMandatory.EnableMinimumEncryptionConfiguration,
				localAthenaCfgMandatory.EnableMinimumEncryptionConfiguration,
			),
			// publishCloudWatchMetricsEnabled: boolean sentinel.
			PublishCloudWatchMetricsEnabled: firstTrue(
				globalKropathMandatory.PublishCloudWatchMetricsEnabled,
				localKropathMandatory.PublishCloudWatchMetricsEnabled,
				globalAthenaCfgMandatory.PublishCloudWatchMetricsEnabled,
				localAthenaCfgMandatory.PublishCloudWatchMetricsEnabled,
			),
			// bytesScannedCutoffPerQuery: not in KropathConfig; levels 3 and 4 only.
			BytesScannedCutoffPerQuery: firstNonZeroInt64(
				globalAthenaCfgMandatory.BytesScannedCutoffPerQuery, // level 3
				localAthenaCfgMandatory.BytesScannedCutoffPerQuery,  // level 4
			),
			// requesterPaysEnabled: not in KropathConfig; levels 3 and 4 only.
			RequesterPaysEnabled: firstTrue(
				globalAthenaCfgMandatory.RequesterPaysEnabled, // level 3
				localAthenaCfgMandatory.RequesterPaysEnabled,  // level 4
			),
			// resultEncryptionOption: not in KropathConfig; levels 3 and 4 only.
			ResultEncryptionOption: firstNonEmptyString(
				globalAthenaCfgMandatory.ResultEncryptionOption, // level 3
				localAthenaCfgMandatory.ResultEncryptionOption,  // level 4
			),
			// resultOutputLocation: not in KropathConfig; levels 3 and 4 only.
			ResultOutputLocation: firstNonEmptyString(
				globalAthenaCfgMandatory.ResultOutputLocation, // level 3
				localAthenaCfgMandatory.ResultOutputLocation,  // level 4
			),
			// engineVersion: not in KropathConfig; levels 3 and 4 only.
			EngineVersion: firstNonEmptyString(
				globalAthenaCfgMandatory.EngineVersion, // level 3
				localAthenaCfgMandatory.EngineVersion,  // level 4
			),
			// namingTemplate: not in KropathConfig; AthenaConfig levels only (3, 4).
			NamingTemplate: firstNonEmptyString(
				globalAthenaCfgMandatory.NamingTemplate, // level 3
				localAthenaCfgMandatory.NamingTemplate,  // level 4
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflicts.
			Tags: mergeMaps(
				localAthenaCfgMandatory.Tags,   // level 4 (lowest priority)
				globalAthenaCfgMandatory.Tags,  // level 3
				localKropathMandatory.Tags,     // level 2
				globalKropathMandatory.Tags,    // level 1 (highest priority)
			),
			// SyncedLabels: additive union from AthenaConfig levels only (3, 4).
			// L4 added first (lower priority); L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localAthenaCfgMandatory.SyncedLabels,  // level 4
				globalAthenaCfgMandatory.SyncedLabels, // level 3
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localAthenaCfgMandatory.SyncedAnnotations,  // level 4
				globalAthenaCfgMandatory.SyncedAnnotations, // level 3
			),
		},
		Defaults: EffectiveAthenaSection{
			// enforceWorkGroupConfiguration: AthenaConfig levels 6–7 + KropathConfig levels 8–9.
			EnforceWorkGroupConfiguration: firstTrue(
				localAthenaCfgDefaults.EnforceWorkGroupConfiguration,  // level 6
				globalAthenaCfgDefaults.EnforceWorkGroupConfiguration, // level 7
				localKropathDefaults.EnforceWorkGroupConfiguration,    // level 8
				globalKropathDefaults.EnforceWorkGroupConfiguration,   // level 9
			),
			EnableMinimumEncryptionConfiguration: firstTrue(
				localAthenaCfgDefaults.EnableMinimumEncryptionConfiguration,  // level 6
				globalAthenaCfgDefaults.EnableMinimumEncryptionConfiguration, // level 7
				localKropathDefaults.EnableMinimumEncryptionConfiguration,    // level 8
				globalKropathDefaults.EnableMinimumEncryptionConfiguration,   // level 9
			),
			PublishCloudWatchMetricsEnabled: firstTrue(
				localAthenaCfgDefaults.PublishCloudWatchMetricsEnabled,  // level 6
				globalAthenaCfgDefaults.PublishCloudWatchMetricsEnabled, // level 7
				localKropathDefaults.PublishCloudWatchMetricsEnabled,    // level 8
				globalKropathDefaults.PublishCloudWatchMetricsEnabled,   // level 9
			),
			// bytesScannedCutoffPerQuery: not in KropathConfig; levels 6 and 7 only.
			BytesScannedCutoffPerQuery: firstNonZeroInt64(
				localAthenaCfgDefaults.BytesScannedCutoffPerQuery,  // level 6
				globalAthenaCfgDefaults.BytesScannedCutoffPerQuery, // level 7
			),
			// requesterPaysEnabled: not in KropathConfig; levels 6 and 7 only.
			RequesterPaysEnabled: firstTrue(
				localAthenaCfgDefaults.RequesterPaysEnabled,  // level 6
				globalAthenaCfgDefaults.RequesterPaysEnabled, // level 7
			),
			// resultEncryptionOption: not in KropathConfig; levels 6 and 7 only.
			ResultEncryptionOption: firstNonEmptyString(
				localAthenaCfgDefaults.ResultEncryptionOption,  // level 6
				globalAthenaCfgDefaults.ResultEncryptionOption, // level 7
			),
			// resultOutputLocation: not in KropathConfig; levels 6 and 7 only.
			ResultOutputLocation: firstNonEmptyString(
				localAthenaCfgDefaults.ResultOutputLocation,  // level 6
				globalAthenaCfgDefaults.ResultOutputLocation, // level 7
			),
			// engineVersion: not in KropathConfig; levels 6 and 7 only.
			EngineVersion: firstNonEmptyString(
				localAthenaCfgDefaults.EngineVersion,  // level 6
				globalAthenaCfgDefaults.EngineVersion, // level 7
			),
			// namingTemplate: not in KropathConfig; AthenaConfig levels only (6, 7).
			NamingTemplate: firstNonEmptyString(
				localAthenaCfgDefaults.NamingTemplate,  // level 6
				globalAthenaCfgDefaults.NamingTemplate, // level 7
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflicts.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,    // level 9 (lowest priority)
				localKropathDefaults.Tags,     // level 8
				globalAthenaCfgDefaults.Tags,  // level 7
				localAthenaCfgDefaults.Tags,   // level 6 (highest priority)
			),
			// SyncedLabels: additive union from AthenaConfig levels only (6, 7).
			// L7 added first (lower priority); L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalAthenaCfgDefaults.SyncedLabels, // level 7
				localAthenaCfgDefaults.SyncedLabels,  // level 6 (wins)
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalAthenaCfgDefaults.SyncedAnnotations, // level 7
				localAthenaCfgDefaults.SyncedAnnotations,  // level 6 (wins)
			),
		},
	}
}
