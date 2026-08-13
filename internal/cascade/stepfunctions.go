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

// StepFunctionsKropathSection holds the Step Functions-family governance fields
// from KropathConfig.spec.mandatory.stepfunctions / .defaults.stepfunctions
// (ADR-015 §3.5).
//
// Only loggingLevel is governed at the KropathConfig level; tracingEnabled,
// includeExecutionData, and namingTemplate are StepFunctionsConfig-only
// (family design §8, OD-8).
//
// Zero value for LoggingLevel is "" (not enforced / permissive sentinel).
type StepFunctionsKropathSection struct {
	// LoggingLevel is the org-wide Step Functions execution logging level.
	// "" = not enforced (mandatory) or no org-wide default (defaults).
	// Valid non-empty values: "ALL" | "ERROR" | "FATAL" | "OFF".
	LoggingLevel string `json:"loggingLevel,omitempty"`

	// Tags are tier-level cloud resource tags from KropathConfig.spec.mandatory.tags
	// or .defaults.tags. Populated by the reconciler so that tag cascade flows
	// through MergeStepFunctionsCascade alongside Step Functions-specific fields.
	Tags map[string]string `json:"tags,omitempty"`
}

// StepFunctionsConfigSection holds the Step Functions governance fields from
// StepFunctionsConfig.spec.mandatory or StepFunctionsConfig.spec.defaults
// (per-type ResourceConfig, ADR-015 §3.5).
type StepFunctionsConfigSection struct {
	// LoggingLevel is the per-profile execution logging level enforcement.
	// "" = not enforced (mandatory) or no profile-level default (defaults).
	LoggingLevel string `json:"loggingLevel,omitempty"`

	// TracingEnabled controls X-Ray tracing for state machine executions.
	// nil = not set (falls through); true = enable tracing; false = explicitly disabled.
	TracingEnabled *bool `json:"tracingEnabled,omitempty"`

	// IncludeExecutionData controls whether execution data (input/output) is included
	// in CloudWatch Logs. nil = not set; true = include; false = explicitly excluded.
	IncludeExecutionData *bool `json:"includeExecutionData,omitempty"`

	// NamingTemplate is the Step Functions resource naming template
	// (e.g. "{namespace}-{name}"). Governed only at StepFunctionsConfig levels
	// 3-4 (mandatory) and 6-7 (defaults). Empty string = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// Tags are cloud resource tags for this StepFunctionsConfig profile.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created Step Functions resources.
	// Additive map merge across StepFunctionsConfig tiers only.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to created Step Functions resources.
	// Additive map merge across StepFunctionsConfig tiers only.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveStepFunctionsSection is one tier (mandatory or defaults) of the merged
// Step Functions governance result written into StepFunctionsConfig.status.effectiveConfig
// by the controller.
type EffectiveStepFunctionsSection struct {
	LoggingLevel         string            `json:"loggingLevel,omitempty"`
	TracingEnabled       *bool             `json:"tracingEnabled,omitempty"`
	IncludeExecutionData *bool             `json:"includeExecutionData,omitempty"`
	NamingTemplate       string            `json:"namingTemplate,omitempty"`
	Tags                 map[string]string `json:"tags,omitempty"`
	SyncedLabels         map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations    map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveStepFunctionsConfig is the merged Step Functions governance result written
// into StepFunctionsConfig.status.effectiveConfig by the controller.
type EffectiveStepFunctionsConfig struct {
	Mandatory EffectiveStepFunctionsSection `json:"mandatory"`
	Defaults  EffectiveStepFunctionsSection `json:"defaults"`
}

// MergeStepFunctionsCascade merges Step Functions governance fields from all cascade
// sources and returns the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for Step Functions (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.stepfunctions)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.stepfunctions)
//	Level 3 — globalSFNCfgMandatory   (StepFunctionsConfig in kro-system, mandatory)
//	Level 4 — localSFNCfgMandatory    (StepFunctionsConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localSFNCfgDefaults     (StepFunctionsConfig in resource namespace, defaults)
//	Level 7 — globalSFNCfgDefaults    (StepFunctionsConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.stepfunctions)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.stepfunctions)
//
// loggingLevel: firstNonEmptyString in priority order (lowest level number wins).
// tracingEnabled / includeExecutionData: firstNonNilBoolPtr from StepFunctionsConfig
// levels only (3-4 mandatory, 6-7 defaults). KropathConfig.stepfunctions has no *bool fields.
// NamingTemplate: StepFunctionsConfig levels only (3-4 mandatory, 6-7 defaults).
// Tags: additive union across all four levels per tier.
// SyncedLabels/SyncedAnnotations: additive union from StepFunctionsConfig levels only.
func MergeStepFunctionsCascade(
	globalKropathMandatory StepFunctionsKropathSection, // level 1
	localKropathMandatory StepFunctionsKropathSection,  // level 2
	globalSFNCfgMandatory StepFunctionsConfigSection,   // level 3
	localSFNCfgMandatory StepFunctionsConfigSection,    // level 4
	localSFNCfgDefaults StepFunctionsConfigSection,     // level 6
	globalSFNCfgDefaults StepFunctionsConfigSection,    // level 7
	localKropathDefaults StepFunctionsKropathSection,   // level 8
	globalKropathDefaults StepFunctionsKropathSection,  // level 9
) EffectiveStepFunctionsConfig {
	return EffectiveStepFunctionsConfig{
		Mandatory: EffectiveStepFunctionsSection{
			// loggingLevel: KropathConfig levels 1-2 win over StepFunctionsConfig 3-4.
			LoggingLevel: firstNonEmptyString(
				globalKropathMandatory.LoggingLevel,
				localKropathMandatory.LoggingLevel,
				globalSFNCfgMandatory.LoggingLevel,
				localSFNCfgMandatory.LoggingLevel,
			),
			// tracingEnabled: StepFunctionsConfig levels only (3, 4).
			TracingEnabled: firstNonNilBoolPtr(
				globalSFNCfgMandatory.TracingEnabled,
				localSFNCfgMandatory.TracingEnabled,
			),
			// includeExecutionData: StepFunctionsConfig levels only (3, 4).
			IncludeExecutionData: firstNonNilBoolPtr(
				globalSFNCfgMandatory.IncludeExecutionData,
				localSFNCfgMandatory.IncludeExecutionData,
			),
			// NamingTemplate: StepFunctionsConfig levels only (3, 4).
			NamingTemplate: firstNonEmptyString(
				globalSFNCfgMandatory.NamingTemplate,
				localSFNCfgMandatory.NamingTemplate,
			),
			// SyncedLabels: additive union from StepFunctionsConfig mandatory levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localSFNCfgMandatory.SyncedLabels,
				globalSFNCfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localSFNCfgMandatory.SyncedAnnotations,
				globalSFNCfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localSFNCfgMandatory.Tags,
				globalSFNCfgMandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
		},
		Defaults: EffectiveStepFunctionsSection{
			// loggingLevel: StepFunctionsConfig levels 6-7 win over KropathConfig 8-9.
			LoggingLevel: firstNonEmptyString(
				localSFNCfgDefaults.LoggingLevel,
				globalSFNCfgDefaults.LoggingLevel,
				localKropathDefaults.LoggingLevel,
				globalKropathDefaults.LoggingLevel,
			),
			// tracingEnabled: StepFunctionsConfig levels only (6, 7).
			TracingEnabled: firstNonNilBoolPtr(
				localSFNCfgDefaults.TracingEnabled,
				globalSFNCfgDefaults.TracingEnabled,
			),
			// includeExecutionData: StepFunctionsConfig levels only (6, 7).
			IncludeExecutionData: firstNonNilBoolPtr(
				localSFNCfgDefaults.IncludeExecutionData,
				globalSFNCfgDefaults.IncludeExecutionData,
			),
			// NamingTemplate: StepFunctionsConfig levels only (6, 7).
			NamingTemplate: firstNonEmptyString(
				localSFNCfgDefaults.NamingTemplate,
				globalSFNCfgDefaults.NamingTemplate,
			),
			// SyncedLabels: additive union from StepFunctionsConfig defaults levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalSFNCfgDefaults.SyncedLabels,
				localSFNCfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalSFNCfgDefaults.SyncedAnnotations,
				localSFNCfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalSFNCfgDefaults.Tags,
				localSFNCfgDefaults.Tags,
			),
		},
	}
}
