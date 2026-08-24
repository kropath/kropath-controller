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

// EMRKropathSection holds the EMR-family governance fields from
// KropathConfig.spec.mandatory.emr / .defaults.emr (ADR-015 §3.5)
// PLUS the tier-level tags from KropathConfig.spec.mandatory.tags (populated
// by the reconciler so that tag cascade flows through MergeEMRCascade).
//
// Only releaseLabel and architecture appear at the KropathConfig level —
// autoStopIdleTimeoutMinutes, maximumCapacityCPU, maximumCapacityMemory,
// diskEncryptionKeyARN, and namingTemplate are EMR Serverless-specific or
// per-type choices that live only in EMRConfig (family design §8).
//
// Zero value of each field is the permissive sentinel (not enforced).
type EMRKropathSection struct {
	// ReleaseLabel is the enforced EMR release version (e.g. "emr-7.1.0").
	// Empty string = not enforced.
	ReleaseLabel string `json:"releaseLabel,omitempty"`

	// Architecture is the enforced CPU architecture ("X86_64" or "ARM64").
	// Empty string = not enforced.
	Architecture string `json:"architecture,omitempty"`

	// Tags are tier-level cloud resource tags from KropathConfig.spec.mandatory.tags.
	// Populated by the reconciler from the tier-level field, not from spec.mandatory.emr.
	// nil / empty map = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// EMRConfigSection holds the EMR governance fields from EMRConfig.spec.mandatory
// or EMRConfig.spec.defaults (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type EMRConfigSection struct {
	// ReleaseLabel is the enforced EMR release version. Empty = not enforced.
	ReleaseLabel string `json:"releaseLabel,omitempty"`

	// Architecture is the enforced CPU architecture. Empty = not enforced.
	Architecture string `json:"architecture,omitempty"`

	// AutoStopIdleTimeoutMinutes is the enforced idle timeout for EMR Serverless auto-stop.
	// Zero (0) = not enforced — check != 0 before applying (integer sentinel pattern).
	AutoStopIdleTimeoutMinutes int64 `json:"autoStopIdleTimeoutMinutes,omitempty"`

	// MaximumCapacityCPU is the enforced maximum total vCPU for EMR Serverless (e.g. "400").
	// Empty = not enforced.
	MaximumCapacityCPU string `json:"maximumCapacityCPU,omitempty"`

	// MaximumCapacityMemory is the enforced maximum total memory (e.g. "3000g").
	// Empty = not enforced.
	MaximumCapacityMemory string `json:"maximumCapacityMemory,omitempty"`

	// DiskEncryptionKeyARN is the enforced KMS key ARN for local worker disk encryption.
	// Empty = not enforced.
	DiskEncryptionKeyARN string `json:"diskEncryptionKeyARN,omitempty"`

	// NamingTemplate is the EMR resource naming template (e.g. "{namespace}-{name}").
	// Empty = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// Tags are cloud resource tags. nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to EMR cloud resources.
	// nil / empty = no synced labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to EMR cloud resources.
	// nil / empty = no synced annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveEMRSection is one tier (mandatory or defaults) of the merged EMR governance
// result written into EMRConfig.status.effectiveConfig by the controller.
type EffectiveEMRSection struct {
	ReleaseLabel               string            `json:"releaseLabel,omitempty"`
	Architecture               string            `json:"architecture,omitempty"`
	AutoStopIdleTimeoutMinutes int64             `json:"autoStopIdleTimeoutMinutes,omitempty"`
	MaximumCapacityCPU         string            `json:"maximumCapacityCPU,omitempty"`
	MaximumCapacityMemory      string            `json:"maximumCapacityMemory,omitempty"`
	DiskEncryptionKeyARN       string            `json:"diskEncryptionKeyARN,omitempty"`
	NamingTemplate             string            `json:"namingTemplate,omitempty"`
	Tags                       map[string]string `json:"tags,omitempty"`
	SyncedLabels               map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations          map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveEMRConfig is the merged EMR governance result written into
// EMRConfig.status.effectiveConfig by the controller.
type EffectiveEMRConfig struct {
	Mandatory EffectiveEMRSection `json:"mandatory"`
	Defaults  EffectiveEMRSection `json:"defaults"`
}

// MergeEMRCascade merges EMR governance fields from all cascade sources and
// returns the effective configuration to be written to status.effectiveConfig.
//
// The ten-level priority chain (ADR-015 §5.3) for EMR fields:
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace)
//	Level 3 — globalEMRCfgMandatory   (EMRConfig in kro-system)
//	Level 4 — localEMRCfgMandatory    (EMRConfig in resource namespace)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localEMRCfgDefaults     (EMRConfig in resource namespace)
//	Level 7 — globalEMRCfgDefaults    (EMRConfig in kro-system)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system)
//
// For mandatory (levels 1–4): first non-zero value in priority order wins.
// For defaults (levels 6–9): first non-zero value in priority order wins.
// For tags / syncedLabels / syncedAnnotations: additive merge; lower level numbers win on key conflicts.
//
// releaseLabel and architecture appear at all four mandatory/defaults levels.
// autoStopIdleTimeoutMinutes, maximumCapacityCPU, maximumCapacityMemory, diskEncryptionKeyARN,
// and namingTemplate appear only at levels 3–4 (mandatory) and 6–7 (defaults) — they are not
// in KropathConfig (family design §8).
// Tags appear at all four mandatory/defaults levels; KropathSection.Tags carries the
// tier-level KropathConfig.mandatory.tags (populated by the reconciler).
// syncedLabels and syncedAnnotations appear at levels 3–4 (mandatory) and 6–7 (defaults) only.
//
// autoStopIdleTimeoutMinutes uses integer-zero as its "not set" sentinel (same pattern as
// firstNonZeroInt64). A value of 0 means "not enforced" and is skipped in priority resolution.
func MergeEMRCascade(
	// Mandatory inputs (highest → lowest priority)
	globalKropathMandatory EMRKropathSection, // level 1
	localKropathMandatory EMRKropathSection, // level 2
	globalEMRCfgMandatory EMRConfigSection, // level 3
	localEMRCfgMandatory EMRConfigSection, // level 4
	// Defaults inputs (highest → lowest priority)
	localEMRCfgDefaults EMRConfigSection, // level 6
	globalEMRCfgDefaults EMRConfigSection, // level 7
	localKropathDefaults EMRKropathSection, // level 8
	globalKropathDefaults EMRKropathSection, // level 9
) EffectiveEMRConfig {
	return EffectiveEMRConfig{
		Mandatory: EffectiveEMRSection{
			ReleaseLabel: firstNonEmptyString(
				globalKropathMandatory.ReleaseLabel, // level 1
				localKropathMandatory.ReleaseLabel,  // level 2
				globalEMRCfgMandatory.ReleaseLabel,  // level 3
				localEMRCfgMandatory.ReleaseLabel,   // level 4
			),
			Architecture: firstNonEmptyString(
				globalKropathMandatory.Architecture, // level 1
				localKropathMandatory.Architecture,  // level 2
				globalEMRCfgMandatory.Architecture,  // level 3
				localEMRCfgMandatory.Architecture,   // level 4
			),
			// autoStopIdleTimeoutMinutes not in KropathConfig: levels 3 and 4 only.
			AutoStopIdleTimeoutMinutes: firstNonZeroInt64(
				globalEMRCfgMandatory.AutoStopIdleTimeoutMinutes, // level 3
				localEMRCfgMandatory.AutoStopIdleTimeoutMinutes,  // level 4
			),
			// maximumCapacityCPU not in KropathConfig: levels 3 and 4 only.
			MaximumCapacityCPU: firstNonEmptyString(
				globalEMRCfgMandatory.MaximumCapacityCPU, // level 3
				localEMRCfgMandatory.MaximumCapacityCPU,  // level 4
			),
			// maximumCapacityMemory not in KropathConfig: levels 3 and 4 only.
			MaximumCapacityMemory: firstNonEmptyString(
				globalEMRCfgMandatory.MaximumCapacityMemory, // level 3
				localEMRCfgMandatory.MaximumCapacityMemory,  // level 4
			),
			// diskEncryptionKeyARN not in KropathConfig: levels 3 and 4 only.
			DiskEncryptionKeyARN: firstNonEmptyString(
				globalEMRCfgMandatory.DiskEncryptionKeyARN, // level 3
				localEMRCfgMandatory.DiskEncryptionKeyARN,  // level 4
			),
			// namingTemplate not in KropathConfig: levels 3 and 4 only.
			NamingTemplate: firstNonEmptyString(
				globalEMRCfgMandatory.NamingTemplate, // level 3
				localEMRCfgMandatory.NamingTemplate,  // level 4
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflicts.
			Tags: mergeMaps(
				localEMRCfgMandatory.Tags,   // level 4 (lowest priority, set first)
				globalEMRCfgMandatory.Tags,  // level 3
				localKropathMandatory.Tags,  // level 2
				globalKropathMandatory.Tags, // level 1 (highest priority, last to write)
			),
			// syncedLabels not in KropathConfig: levels 3 and 4 only.
			SyncedLabels: mergeMaps(
				localEMRCfgMandatory.SyncedLabels,  // level 4
				globalEMRCfgMandatory.SyncedLabels, // level 3
			),
			// syncedAnnotations not in KropathConfig: levels 3 and 4 only.
			SyncedAnnotations: mergeMaps(
				localEMRCfgMandatory.SyncedAnnotations,  // level 4
				globalEMRCfgMandatory.SyncedAnnotations, // level 3
			),
		},
		Defaults: EffectiveEMRSection{
			ReleaseLabel: firstNonEmptyString(
				localEMRCfgDefaults.ReleaseLabel,  // level 6
				globalEMRCfgDefaults.ReleaseLabel, // level 7
				localKropathDefaults.ReleaseLabel,  // level 8
				globalKropathDefaults.ReleaseLabel, // level 9
			),
			Architecture: firstNonEmptyString(
				localEMRCfgDefaults.Architecture,  // level 6
				globalEMRCfgDefaults.Architecture, // level 7
				localKropathDefaults.Architecture,  // level 8
				globalKropathDefaults.Architecture, // level 9
			),
			// autoStopIdleTimeoutMinutes not in KropathConfig: levels 6 and 7 only.
			AutoStopIdleTimeoutMinutes: firstNonZeroInt64(
				localEMRCfgDefaults.AutoStopIdleTimeoutMinutes,  // level 6
				globalEMRCfgDefaults.AutoStopIdleTimeoutMinutes, // level 7
			),
			// maximumCapacityCPU not in KropathConfig: levels 6 and 7 only.
			MaximumCapacityCPU: firstNonEmptyString(
				localEMRCfgDefaults.MaximumCapacityCPU,  // level 6
				globalEMRCfgDefaults.MaximumCapacityCPU, // level 7
			),
			// maximumCapacityMemory not in KropathConfig: levels 6 and 7 only.
			MaximumCapacityMemory: firstNonEmptyString(
				localEMRCfgDefaults.MaximumCapacityMemory,  // level 6
				globalEMRCfgDefaults.MaximumCapacityMemory, // level 7
			),
			// diskEncryptionKeyARN not in KropathConfig: levels 6 and 7 only.
			DiskEncryptionKeyARN: firstNonEmptyString(
				localEMRCfgDefaults.DiskEncryptionKeyARN,  // level 6
				globalEMRCfgDefaults.DiskEncryptionKeyARN, // level 7
			),
			// namingTemplate not in KropathConfig: levels 6 and 7 only.
			NamingTemplate: firstNonEmptyString(
				localEMRCfgDefaults.NamingTemplate,  // level 6
				globalEMRCfgDefaults.NamingTemplate, // level 7
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflicts.
			Tags: mergeMaps(
				globalKropathDefaults.Tags, // level 9 (lowest priority)
				localKropathDefaults.Tags,  // level 8
				globalEMRCfgDefaults.Tags,  // level 7
				localEMRCfgDefaults.Tags,   // level 6 (highest priority)
			),
			// syncedLabels not in KropathConfig: levels 6 and 7 only.
			SyncedLabels: mergeMaps(
				globalEMRCfgDefaults.SyncedLabels, // level 7
				localEMRCfgDefaults.SyncedLabels,  // level 6 (wins)
			),
			// syncedAnnotations not in KropathConfig: levels 6 and 7 only.
			SyncedAnnotations: mergeMaps(
				globalEMRCfgDefaults.SyncedAnnotations, // level 7
				localEMRCfgDefaults.SyncedAnnotations,  // level 6 (wins)
			),
		},
	}
}
