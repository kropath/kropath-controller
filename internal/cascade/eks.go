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

// EKSKropathSection holds the EKS-family governance fields from
// KropathConfig.spec.mandatory.eks / .defaults.eks (ADR-015 §3.5)
// PLUS the tier-level tags from KropathConfig.spec.mandatory.tags (populated
// by the reconciler so that tag cascade flows through MergeEKSCascade).
//
// Only version, authenticationMode, and loggingTypes appear at the KropathConfig
// level — endpoint, encryption, supportType, and namingTemplate are per-cluster
// choices that live only in EKSConfig (family design §8).
//
// Zero value of each field is the permissive sentinel (not enforced).
type EKSKropathSection struct {
	// Version is the enforced Kubernetes control-plane version (e.g. "1.31").
	// Empty string = not enforced.
	Version string `json:"version,omitempty"`

	// AuthenticationMode is the enforced cluster auth mode ("API", "CONFIG_MAP", or "API_AND_CONFIG_MAP").
	// Empty string = not enforced.
	AuthenticationMode string `json:"authenticationMode,omitempty"`

	// LoggingTypes is the enforced set of EKS control-plane log types.
	// Treated as a scalar (first-non-empty wins); nil / empty = not enforced.
	LoggingTypes []string `json:"loggingTypes,omitempty"`

	// Tags are tier-level cloud resource tags from KropathConfig.spec.mandatory.tags.
	// Populated by the reconciler from the tier-level field, not from spec.mandatory.eks.
	// nil / empty map = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// EKSConfigSection holds the EKS governance fields from EKSConfig.spec.mandatory
// or EKSConfig.spec.defaults (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type EKSConfigSection struct {
	// Version is the enforced Kubernetes version. Empty = not enforced.
	Version string `json:"version,omitempty"`

	// AuthenticationMode is the enforced cluster auth mode. Empty = not enforced.
	AuthenticationMode string `json:"authenticationMode,omitempty"`

	// LoggingTypes is the enforced set of EKS control-plane log types.
	// Treated as a scalar (first-non-empty wins); nil / empty = not enforced.
	LoggingTypes []string `json:"loggingTypes,omitempty"`

	// EncryptionKeyArn is the KMS key ARN for EKS secrets encryption.
	// Empty = not enforced.
	EncryptionKeyArn string `json:"encryptionKeyArn,omitempty"`

	// EndpointPublicAccess controls whether the cluster API endpoint is publicly reachable.
	// nil = not set (falls through); true = enforce enabled; false = enforce disabled.
	EndpointPublicAccess *bool `json:"endpointPublicAccess,omitempty"`

	// EndpointPrivateAccess controls whether the cluster API endpoint is privately reachable.
	// nil = not set (falls through); true = enforce enabled; false = enforce disabled.
	EndpointPrivateAccess *bool `json:"endpointPrivateAccess,omitempty"`

	// SupportType is the enforced EKS support tier ("STANDARD" or "EXTENDED").
	// Empty = not enforced.
	SupportType string `json:"supportType,omitempty"`

	// NamingTemplate is the cluster naming template (e.g. "{namespace}-{name}").
	// Empty = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// Tags are cloud resource tags. nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to the EKS cluster resource.
	// nil / empty = no synced labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to the EKS cluster resource.
	// nil / empty = no synced annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveEKSSection is one tier (mandatory or defaults) of the merged EKS governance
// result written into EKSConfig.status.effectiveConfig by the controller.
type EffectiveEKSSection struct {
	Version               string            `json:"version,omitempty"`
	AuthenticationMode    string            `json:"authenticationMode,omitempty"`
	LoggingTypes          []string          `json:"loggingTypes,omitempty"`
	EncryptionKeyArn      string            `json:"encryptionKeyArn,omitempty"`
	EndpointPublicAccess  *bool             `json:"endpointPublicAccess,omitempty"`
	EndpointPrivateAccess *bool             `json:"endpointPrivateAccess,omitempty"`
	SupportType           string            `json:"supportType,omitempty"`
	NamingTemplate        string            `json:"namingTemplate,omitempty"`
	Tags                  map[string]string `json:"tags,omitempty"`
	SyncedLabels          map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations     map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveEKSConfig is the merged EKS governance result written into
// EKSConfig.status.effectiveConfig by the controller.
type EffectiveEKSConfig struct {
	Mandatory EffectiveEKSSection `json:"mandatory"`
	Defaults  EffectiveEKSSection `json:"defaults"`
}

// MergeEKSCascade merges EKS governance fields from all cascade sources and
// returns the effective configuration to be written to status.effectiveConfig.
//
// The ten-level priority chain (ADR-015 §5.3) for EKS fields:
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace)
//	Level 3 — globalEKSCfgMandatory   (EKSConfig in kro-system)
//	Level 4 — localEKSCfgMandatory    (EKSConfig in resource namespace)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localEKSCfgDefaults     (EKSConfig in resource namespace)
//	Level 7 — globalEKSCfgDefaults    (EKSConfig in kro-system)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system)
//
// For mandatory (levels 1–4): first non-zero value in priority order wins.
// For defaults (levels 6–9): first non-zero value in priority order wins.
// For tags / syncedLabels / syncedAnnotations: additive merge; lower level numbers win on key conflicts.
//
// version, authenticationMode, and loggingTypes appear at all four mandatory/defaults levels.
// encryptionKeyArn, endpointPublicAccess/PrivateAccess, supportType, and namingTemplate appear
// only at levels 3–4 (mandatory) and 6–7 (defaults) — they are not in KropathConfig.
// Tags appear at all four mandatory/defaults levels; KropathSection.Tags carries the
// tier-level KropathConfig.mandatory.tags (populated by the reconciler).
// syncedLabels and syncedAnnotations appear at levels 3–4 (mandatory) and 6–7 (defaults) only.
func MergeEKSCascade(
	// Mandatory inputs (highest → lowest priority)
	globalKropathMandatory EKSKropathSection, // level 1
	localKropathMandatory EKSKropathSection, // level 2
	globalEKSCfgMandatory EKSConfigSection, // level 3
	localEKSCfgMandatory EKSConfigSection, // level 4
	// Defaults inputs (highest → lowest priority)
	localEKSCfgDefaults EKSConfigSection, // level 6
	globalEKSCfgDefaults EKSConfigSection, // level 7
	localKropathDefaults EKSKropathSection, // level 8
	globalKropathDefaults EKSKropathSection, // level 9
) EffectiveEKSConfig {
	return EffectiveEKSConfig{
		Mandatory: EffectiveEKSSection{
			Version: firstNonEmptyString(
				globalKropathMandatory.Version, // level 1
				localKropathMandatory.Version,  // level 2
				globalEKSCfgMandatory.Version,  // level 3
				localEKSCfgMandatory.Version,   // level 4
			),
			AuthenticationMode: firstNonEmptyString(
				globalKropathMandatory.AuthenticationMode, // level 1
				localKropathMandatory.AuthenticationMode,  // level 2
				globalEKSCfgMandatory.AuthenticationMode,  // level 3
				localEKSCfgMandatory.AuthenticationMode,   // level 4
			),
			LoggingTypes: firstNonEmptyStrings(
				globalKropathMandatory.LoggingTypes, // level 1
				localKropathMandatory.LoggingTypes,  // level 2
				globalEKSCfgMandatory.LoggingTypes,  // level 3
				localEKSCfgMandatory.LoggingTypes,   // level 4
			),
			// encryptionKeyArn not in KropathConfig: levels 3 and 4 only.
			EncryptionKeyArn: firstNonEmptyString(
				globalEKSCfgMandatory.EncryptionKeyArn, // level 3
				localEKSCfgMandatory.EncryptionKeyArn,  // level 4
			),
			// endpoint flags not in KropathConfig: levels 3 and 4 only.
			EndpointPublicAccess: firstNonNilBoolPtr(
				globalEKSCfgMandatory.EndpointPublicAccess, // level 3
				localEKSCfgMandatory.EndpointPublicAccess,  // level 4
			),
			EndpointPrivateAccess: firstNonNilBoolPtr(
				globalEKSCfgMandatory.EndpointPrivateAccess, // level 3
				localEKSCfgMandatory.EndpointPrivateAccess,  // level 4
			),
			// supportType not in KropathConfig: levels 3 and 4 only.
			SupportType: firstNonEmptyString(
				globalEKSCfgMandatory.SupportType, // level 3
				localEKSCfgMandatory.SupportType,  // level 4
			),
			// namingTemplate not in KropathConfig: levels 3 and 4 only.
			NamingTemplate: firstNonEmptyString(
				globalEKSCfgMandatory.NamingTemplate, // level 3
				localEKSCfgMandatory.NamingTemplate,  // level 4
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflicts.
			Tags: mergeMaps(
				localEKSCfgMandatory.Tags,   // level 4 (lowest priority, set first)
				globalEKSCfgMandatory.Tags,  // level 3
				localKropathMandatory.Tags,  // level 2
				globalKropathMandatory.Tags, // level 1 (highest priority, last to write)
			),
			// syncedLabels not in KropathConfig: levels 3 and 4 only.
			SyncedLabels: mergeMaps(
				localEKSCfgMandatory.SyncedLabels,  // level 4
				globalEKSCfgMandatory.SyncedLabels, // level 3
			),
			// syncedAnnotations not in KropathConfig: levels 3 and 4 only.
			SyncedAnnotations: mergeMaps(
				localEKSCfgMandatory.SyncedAnnotations,  // level 4
				globalEKSCfgMandatory.SyncedAnnotations, // level 3
			),
		},
		Defaults: EffectiveEKSSection{
			Version: firstNonEmptyString(
				localEKSCfgDefaults.Version,  // level 6
				globalEKSCfgDefaults.Version, // level 7
				localKropathDefaults.Version,  // level 8
				globalKropathDefaults.Version, // level 9
			),
			AuthenticationMode: firstNonEmptyString(
				localEKSCfgDefaults.AuthenticationMode,  // level 6
				globalEKSCfgDefaults.AuthenticationMode, // level 7
				localKropathDefaults.AuthenticationMode,  // level 8
				globalKropathDefaults.AuthenticationMode, // level 9
			),
			LoggingTypes: firstNonEmptyStrings(
				localEKSCfgDefaults.LoggingTypes,  // level 6
				globalEKSCfgDefaults.LoggingTypes, // level 7
				localKropathDefaults.LoggingTypes,  // level 8
				globalKropathDefaults.LoggingTypes, // level 9
			),
			// encryptionKeyArn not in KropathConfig: levels 6 and 7 only.
			EncryptionKeyArn: firstNonEmptyString(
				localEKSCfgDefaults.EncryptionKeyArn,  // level 6
				globalEKSCfgDefaults.EncryptionKeyArn, // level 7
			),
			// endpoint flags not in KropathConfig: levels 6 and 7 only.
			EndpointPublicAccess: firstNonNilBoolPtr(
				localEKSCfgDefaults.EndpointPublicAccess,  // level 6
				globalEKSCfgDefaults.EndpointPublicAccess, // level 7
			),
			EndpointPrivateAccess: firstNonNilBoolPtr(
				localEKSCfgDefaults.EndpointPrivateAccess,  // level 6
				globalEKSCfgDefaults.EndpointPrivateAccess, // level 7
			),
			// supportType not in KropathConfig: levels 6 and 7 only.
			SupportType: firstNonEmptyString(
				localEKSCfgDefaults.SupportType,  // level 6
				globalEKSCfgDefaults.SupportType, // level 7
			),
			// namingTemplate not in KropathConfig: levels 6 and 7 only.
			NamingTemplate: firstNonEmptyString(
				localEKSCfgDefaults.NamingTemplate,  // level 6
				globalEKSCfgDefaults.NamingTemplate, // level 7
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflicts.
			Tags: mergeMaps(
				globalKropathDefaults.Tags, // level 9 (lowest priority)
				localKropathDefaults.Tags,  // level 8
				globalEKSCfgDefaults.Tags,  // level 7
				localEKSCfgDefaults.Tags,   // level 6 (highest priority)
			),
			// syncedLabels not in KropathConfig: levels 6 and 7 only.
			SyncedLabels: mergeMaps(
				globalEKSCfgDefaults.SyncedLabels, // level 7
				localEKSCfgDefaults.SyncedLabels,  // level 6 (wins)
			),
			// syncedAnnotations not in KropathConfig: levels 6 and 7 only.
			SyncedAnnotations: mergeMaps(
				globalEKSCfgDefaults.SyncedAnnotations, // level 7
				localEKSCfgDefaults.SyncedAnnotations,  // level 6 (wins)
			),
		},
	}
}
