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

// GlueKropathSection holds the Glue-family governance fields from
// KropathConfig.spec.mandatory.glue / .defaults.glue (ADR-015 §3.5)
// PLUS the tier-level tags from KropathConfig.spec.mandatory.tags (populated
// by the reconciler so that tag cascade flows through MergeGlueCascade).
//
// Tags, syncedLabels, and syncedAnnotations are org-wide fields at the
// KropathConfig tier level — they do NOT appear under spec.glue.
//
// Zero value of each scalar field is the permissive sentinel (not enforced).
type GlueKropathSection struct {
	// GlueVersion is the enforced Glue runtime version (e.g. "4.0", "3.0").
	// Empty string = not enforced.
	GlueVersion string `json:"glueVersion,omitempty"`

	// WorkerType is the enforced Glue worker type (G.1X | G.2X | G.4X | G.8X | G.025X | Z.2X).
	// Empty string = not enforced.
	WorkerType string `json:"workerType,omitempty"`

	// NumberOfWorkers is the enforced Glue worker count.
	// Zero (0) = not enforced (integer sentinel pattern).
	NumberOfWorkers int64 `json:"numberOfWorkers,omitempty"`

	// ExecutionClass is the enforced Glue execution class (STANDARD | FLEX).
	// Empty string = not enforced.
	ExecutionClass string `json:"executionClass,omitempty"`

	// Timeout is the enforced Glue job timeout in minutes.
	// Zero (0) = not enforced.
	Timeout int64 `json:"timeout,omitempty"`

	// MaxRetries is the enforced Glue job retry limit.
	// Zero (0) = not enforced.
	MaxRetries int64 `json:"maxRetries,omitempty"`

	// MaxConcurrentRuns is the enforced Glue job concurrency limit.
	// Zero (0) = not enforced.
	MaxConcurrentRuns int64 `json:"maxConcurrentRuns,omitempty"`

	// SecurityConfiguration is the enforced Glue SecurityConfiguration name.
	// Empty string = not enforced.
	SecurityConfiguration string `json:"securityConfiguration,omitempty"`

	// NotifyDelayAfter is the enforced CloudWatch notification delay in minutes.
	// Zero (0) = not enforced.
	NotifyDelayAfter int64 `json:"notifyDelayAfter,omitempty"`

	// Tags are tier-level cloud resource tags from KropathConfig.spec.mandatory.tags
	// or KropathConfig.spec.defaults.tags. Populated by the reconciler from the
	// tier-level field, not from spec.mandatory.glue. nil / empty = no tags.
	Tags map[string]string `json:"tags,omitempty"`
}

// GlueConfigSection holds the Glue governance fields from GlueConfig.spec.mandatory
// or GlueConfig.spec.defaults (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each scalar field is the permissive sentinel (not enforced).
type GlueConfigSection struct {
	// GlueVersion is the enforced Glue runtime version. Empty = not enforced.
	GlueVersion string `json:"glueVersion,omitempty"`

	// WorkerType is the enforced Glue worker type. Empty = not enforced.
	WorkerType string `json:"workerType,omitempty"`

	// NumberOfWorkers is the enforced worker count.
	// Zero (0) = not enforced — check != 0 before applying (integer sentinel pattern).
	NumberOfWorkers int64 `json:"numberOfWorkers,omitempty"`

	// ExecutionClass is the enforced execution class (STANDARD | FLEX). Empty = not enforced.
	ExecutionClass string `json:"executionClass,omitempty"`

	// Timeout is the enforced job timeout in minutes.
	// Zero (0) = not enforced.
	Timeout int64 `json:"timeout,omitempty"`

	// MaxRetries is the enforced retry limit.
	// Zero (0) = not enforced.
	MaxRetries int64 `json:"maxRetries,omitempty"`

	// MaxConcurrentRuns is the enforced concurrency limit.
	// Zero (0) = not enforced.
	MaxConcurrentRuns int64 `json:"maxConcurrentRuns,omitempty"`

	// SecurityConfiguration is the enforced Glue SecurityConfiguration name.
	// Empty = not enforced.
	SecurityConfiguration string `json:"securityConfiguration,omitempty"`

	// NotifyDelayAfter is the enforced CloudWatch notification delay in minutes.
	// Zero (0) = not enforced.
	NotifyDelayAfter int64 `json:"notifyDelayAfter,omitempty"`

	// NamingTemplate is the Glue resource naming template (e.g. "{namespace}-{name}").
	// Empty = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// Tags are cloud resource tags. nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to Glue cloud resources.
	// nil / empty = no synced labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to Glue cloud resources.
	// nil / empty = no synced annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveGlueSection is one tier (mandatory or defaults) of the merged Glue governance
// result written into GlueConfig.status.effectiveConfig by the controller.
type EffectiveGlueSection struct {
	GlueVersion           string            `json:"glueVersion,omitempty"`
	WorkerType            string            `json:"workerType,omitempty"`
	NumberOfWorkers       int64             `json:"numberOfWorkers,omitempty"`
	ExecutionClass        string            `json:"executionClass,omitempty"`
	Timeout               int64             `json:"timeout,omitempty"`
	MaxRetries            int64             `json:"maxRetries,omitempty"`
	MaxConcurrentRuns     int64             `json:"maxConcurrentRuns,omitempty"`
	SecurityConfiguration string            `json:"securityConfiguration,omitempty"`
	NotifyDelayAfter      int64             `json:"notifyDelayAfter,omitempty"`
	NamingTemplate        string            `json:"namingTemplate,omitempty"`
	Tags                  map[string]string `json:"tags,omitempty"`
	SyncedLabels          map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations     map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveGlueConfig is the merged Glue governance result written into
// GlueConfig.status.effectiveConfig by the controller.
type EffectiveGlueConfig struct {
	Mandatory EffectiveGlueSection `json:"mandatory"`
	Defaults  EffectiveGlueSection `json:"defaults"`
}

// MergeGlueCascade merges Glue governance fields from all cascade sources and
// returns the effective configuration to be written to status.effectiveConfig.
//
// The ten-level priority chain (ADR-015 §5.3) for Glue fields:
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace)
//	Level 3 — globalGlueCfgMandatory  (GlueConfig in kro-system)
//	Level 4 — localGlueCfgMandatory   (GlueConfig in resource namespace)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localGlueCfgDefaults    (GlueConfig in resource namespace)
//	Level 7 — globalGlueCfgDefaults   (GlueConfig in kro-system)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system)
//
// For mandatory (levels 1–4): first non-zero value in priority order wins.
// For defaults (levels 6–9): first non-zero value in priority order wins.
// For tags / syncedLabels / syncedAnnotations: additive merge; lower level numbers win on key conflicts.
//
// glueVersion, workerType, executionClass, securityConfiguration appear at all four
// mandatory/defaults levels (both KropathConfig and GlueConfig).
// numberOfWorkers, timeout, maxRetries, maxConcurrentRuns, notifyDelayAfter appear at all
// four levels too — use integer-zero as their "not set" sentinel (firstNonZeroInt64).
// namingTemplate appears only at levels 3–4 (mandatory) and 6–7 (defaults) — not in KropathConfig.
// syncedLabels and syncedAnnotations appear at levels 3–4 (mandatory) and 6–7 (defaults) only.
// Tags appear at all four mandatory/defaults levels; GlueKropathSection.Tags carries the
// tier-level KropathConfig.mandatory.tags / defaults.tags (populated by the reconciler).
func MergeGlueCascade(
	// Mandatory inputs (highest → lowest priority)
	globalKropathMandatory GlueKropathSection, // level 1
	localKropathMandatory GlueKropathSection, // level 2
	globalGlueCfgMandatory GlueConfigSection, // level 3
	localGlueCfgMandatory GlueConfigSection, // level 4
	// Defaults inputs (highest → lowest priority)
	localGlueCfgDefaults GlueConfigSection, // level 6
	globalGlueCfgDefaults GlueConfigSection, // level 7
	localKropathDefaults GlueKropathSection, // level 8
	globalKropathDefaults GlueKropathSection, // level 9
) EffectiveGlueConfig {
	return EffectiveGlueConfig{
		Mandatory: EffectiveGlueSection{
			GlueVersion: firstNonEmptyString(
				globalKropathMandatory.GlueVersion, // level 1
				localKropathMandatory.GlueVersion,  // level 2
				globalGlueCfgMandatory.GlueVersion, // level 3
				localGlueCfgMandatory.GlueVersion,  // level 4
			),
			WorkerType: firstNonEmptyString(
				globalKropathMandatory.WorkerType, // level 1
				localKropathMandatory.WorkerType,  // level 2
				globalGlueCfgMandatory.WorkerType, // level 3
				localGlueCfgMandatory.WorkerType,  // level 4
			),
			NumberOfWorkers: firstNonZeroInt64(
				globalKropathMandatory.NumberOfWorkers, // level 1
				localKropathMandatory.NumberOfWorkers,  // level 2
				globalGlueCfgMandatory.NumberOfWorkers, // level 3
				localGlueCfgMandatory.NumberOfWorkers,  // level 4
			),
			ExecutionClass: firstNonEmptyString(
				globalKropathMandatory.ExecutionClass, // level 1
				localKropathMandatory.ExecutionClass,  // level 2
				globalGlueCfgMandatory.ExecutionClass, // level 3
				localGlueCfgMandatory.ExecutionClass,  // level 4
			),
			Timeout: firstNonZeroInt64(
				globalKropathMandatory.Timeout, // level 1
				localKropathMandatory.Timeout,  // level 2
				globalGlueCfgMandatory.Timeout, // level 3
				localGlueCfgMandatory.Timeout,  // level 4
			),
			MaxRetries: firstNonZeroInt64(
				globalKropathMandatory.MaxRetries, // level 1
				localKropathMandatory.MaxRetries,  // level 2
				globalGlueCfgMandatory.MaxRetries, // level 3
				localGlueCfgMandatory.MaxRetries,  // level 4
			),
			MaxConcurrentRuns: firstNonZeroInt64(
				globalKropathMandatory.MaxConcurrentRuns, // level 1
				localKropathMandatory.MaxConcurrentRuns,  // level 2
				globalGlueCfgMandatory.MaxConcurrentRuns, // level 3
				localGlueCfgMandatory.MaxConcurrentRuns,  // level 4
			),
			SecurityConfiguration: firstNonEmptyString(
				globalKropathMandatory.SecurityConfiguration, // level 1
				localKropathMandatory.SecurityConfiguration,  // level 2
				globalGlueCfgMandatory.SecurityConfiguration, // level 3
				localGlueCfgMandatory.SecurityConfiguration,  // level 4
			),
			NotifyDelayAfter: firstNonZeroInt64(
				globalKropathMandatory.NotifyDelayAfter, // level 1
				localKropathMandatory.NotifyDelayAfter,  // level 2
				globalGlueCfgMandatory.NotifyDelayAfter, // level 3
				localGlueCfgMandatory.NotifyDelayAfter,  // level 4
			),
			// namingTemplate not in KropathConfig: levels 3 and 4 only.
			NamingTemplate: firstNonEmptyString(
				globalGlueCfgMandatory.NamingTemplate, // level 3
				localGlueCfgMandatory.NamingTemplate,  // level 4
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflicts.
			Tags: mergeMaps(
				localGlueCfgMandatory.Tags,   // level 4 (lowest priority, set first)
				globalGlueCfgMandatory.Tags,  // level 3
				localKropathMandatory.Tags,   // level 2
				globalKropathMandatory.Tags,  // level 1 (highest priority, last to write)
			),
			// syncedLabels not in KropathConfig: levels 3 and 4 only.
			SyncedLabels: mergeMaps(
				localGlueCfgMandatory.SyncedLabels,  // level 4
				globalGlueCfgMandatory.SyncedLabels, // level 3
			),
			// syncedAnnotations not in KropathConfig: levels 3 and 4 only.
			SyncedAnnotations: mergeMaps(
				localGlueCfgMandatory.SyncedAnnotations,  // level 4
				globalGlueCfgMandatory.SyncedAnnotations, // level 3
			),
		},
		Defaults: EffectiveGlueSection{
			GlueVersion: firstNonEmptyString(
				localGlueCfgDefaults.GlueVersion,  // level 6
				globalGlueCfgDefaults.GlueVersion, // level 7
				localKropathDefaults.GlueVersion,  // level 8
				globalKropathDefaults.GlueVersion, // level 9
			),
			WorkerType: firstNonEmptyString(
				localGlueCfgDefaults.WorkerType,  // level 6
				globalGlueCfgDefaults.WorkerType, // level 7
				localKropathDefaults.WorkerType,  // level 8
				globalKropathDefaults.WorkerType, // level 9
			),
			NumberOfWorkers: firstNonZeroInt64(
				localGlueCfgDefaults.NumberOfWorkers,  // level 6
				globalGlueCfgDefaults.NumberOfWorkers, // level 7
				localKropathDefaults.NumberOfWorkers,  // level 8
				globalKropathDefaults.NumberOfWorkers, // level 9
			),
			ExecutionClass: firstNonEmptyString(
				localGlueCfgDefaults.ExecutionClass,  // level 6
				globalGlueCfgDefaults.ExecutionClass, // level 7
				localKropathDefaults.ExecutionClass,  // level 8
				globalKropathDefaults.ExecutionClass, // level 9
			),
			Timeout: firstNonZeroInt64(
				localGlueCfgDefaults.Timeout,  // level 6
				globalGlueCfgDefaults.Timeout, // level 7
				localKropathDefaults.Timeout,  // level 8
				globalKropathDefaults.Timeout, // level 9
			),
			MaxRetries: firstNonZeroInt64(
				localGlueCfgDefaults.MaxRetries,  // level 6
				globalGlueCfgDefaults.MaxRetries, // level 7
				localKropathDefaults.MaxRetries,  // level 8
				globalKropathDefaults.MaxRetries, // level 9
			),
			MaxConcurrentRuns: firstNonZeroInt64(
				localGlueCfgDefaults.MaxConcurrentRuns,  // level 6
				globalGlueCfgDefaults.MaxConcurrentRuns, // level 7
				localKropathDefaults.MaxConcurrentRuns,  // level 8
				globalKropathDefaults.MaxConcurrentRuns, // level 9
			),
			SecurityConfiguration: firstNonEmptyString(
				localGlueCfgDefaults.SecurityConfiguration,  // level 6
				globalGlueCfgDefaults.SecurityConfiguration, // level 7
				localKropathDefaults.SecurityConfiguration,  // level 8
				globalKropathDefaults.SecurityConfiguration, // level 9
			),
			NotifyDelayAfter: firstNonZeroInt64(
				localGlueCfgDefaults.NotifyDelayAfter,  // level 6
				globalGlueCfgDefaults.NotifyDelayAfter, // level 7
				localKropathDefaults.NotifyDelayAfter,  // level 8
				globalKropathDefaults.NotifyDelayAfter, // level 9
			),
			// namingTemplate not in KropathConfig: levels 6 and 7 only.
			NamingTemplate: firstNonEmptyString(
				localGlueCfgDefaults.NamingTemplate,  // level 6
				globalGlueCfgDefaults.NamingTemplate, // level 7
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflicts.
			Tags: mergeMaps(
				globalKropathDefaults.Tags, // level 9 (lowest priority)
				localKropathDefaults.Tags,  // level 8
				globalGlueCfgDefaults.Tags, // level 7
				localGlueCfgDefaults.Tags,  // level 6 (highest priority)
			),
			// syncedLabels not in KropathConfig: levels 6 and 7 only.
			SyncedLabels: mergeMaps(
				globalGlueCfgDefaults.SyncedLabels, // level 7
				localGlueCfgDefaults.SyncedLabels,  // level 6 (wins)
			),
			// syncedAnnotations not in KropathConfig: levels 6 and 7 only.
			SyncedAnnotations: mergeMaps(
				globalGlueCfgDefaults.SyncedAnnotations, // level 7
				localGlueCfgDefaults.SyncedAnnotations,  // level 6 (wins)
			),
		},
	}
}
