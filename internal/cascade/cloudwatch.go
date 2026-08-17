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

// CloudWatchKropathSection holds the CloudWatch-family governance fields from
// KropathConfig.spec.mandatory.cloudwatch / .defaults.cloudwatch (ADR-015 §3.5).
//
// Only 3 scalar fields are governed at the KropathConfig level: actionsEnabled,
// treatMissingData, and outputFormat. namingTemplate is CloudWatchConfig-only
// (cascade levels 3-4 and 6-7).
//
// Zero value of each field is the permissive sentinel (not enforced).
type CloudWatchKropathSection struct {
	// ActionsEnabled enforces alarm actions org-wide.
	// nil = not enforced; true = forced enabled; false = forced disabled.
	// Distinct from false — nil means this level does not participate in the cascade.
	ActionsEnabled *bool `json:"actionsEnabled,omitempty"`

	// TreatMissingData is the org-wide missing data treatment enforcement.
	// Empty string = not enforced; allowed values: breaching|notBreaching|ignore|missing.
	TreatMissingData string `json:"treatMissingData,omitempty"`

	// OutputFormat is the org-wide metric stream output format enforcement.
	// Empty string = not enforced; allowed values: json|opentelemetry1.0|opentelemetry0.7.
	OutputFormat string `json:"outputFormat,omitempty"`

	// Tags are tier-level cloud resource tags.
	// The reconciler populates this from KropathConfig.spec.mandatory.tags / .defaults.tags
	// so that tag union merge flows through MergeCloudWatchCascade.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// CloudWatchConfigSection holds the CloudWatch governance fields from
// CloudWatchConfig.spec.mandatory or CloudWatchConfig.spec.defaults
// (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type CloudWatchConfigSection struct {
	// ActionsEnabled enforces alarm actions for this profile.
	// nil = not enforced; true = forced enabled; false = forced disabled.
	ActionsEnabled *bool `json:"actionsEnabled,omitempty"`

	// TreatMissingData is the missing data treatment for this profile.
	// Empty string = not enforced; allowed values: breaching|notBreaching|ignore|missing.
	TreatMissingData string `json:"treatMissingData,omitempty"`

	// OutputFormat is the metric stream output format for this profile.
	// Empty string = not enforced; allowed values: json|opentelemetry1.0|opentelemetry0.7.
	OutputFormat string `json:"outputFormat,omitempty"`

	// NamingTemplate is the naming template (e.g. "{namespace}-{name}").
	// Governed only at CloudWatchConfig levels 3-4 (mandatory) and 6-7 (defaults).
	// KropathConfig.cloudwatch does NOT carry namingTemplate.
	// Empty string = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created CloudWatch resources.
	// Additive map merge across CloudWatchConfig tiers only.
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to created CloudWatch resources.
	// Additive map merge across CloudWatchConfig tiers only.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`

	// Tags are cloud resource tags for this CloudWatch config profile.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// EffectiveCloudWatchSection is one tier (mandatory or defaults) of the merged
// CloudWatch governance result written into CloudWatchConfig.status.effectiveConfig
// by the controller.
type EffectiveCloudWatchSection struct {
	ActionsEnabled    *bool             `json:"actionsEnabled,omitempty"`
	TreatMissingData  string            `json:"treatMissingData,omitempty"`
	OutputFormat      string            `json:"outputFormat,omitempty"`
	NamingTemplate    string            `json:"namingTemplate,omitempty"`
	SyncedLabels      map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
	Tags              map[string]string `json:"tags,omitempty"`
}

// EffectiveCloudWatchConfig is the merged CloudWatch governance result written into
// CloudWatchConfig.status.effectiveConfig by the controller.
type EffectiveCloudWatchConfig struct {
	Mandatory EffectiveCloudWatchSection `json:"mandatory"`
	Defaults  EffectiveCloudWatchSection `json:"defaults"`
}

// MergeCloudWatchCascade merges CloudWatch governance fields from all cascade
// sources and returns the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for CloudWatch (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.cloudwatch)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.cloudwatch)
//	Level 3 — globalCWCfgMandatory    (CloudWatchConfig in kro-system, mandatory)
//	Level 4 — localCWCfgMandatory     (CloudWatchConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localCWCfgDefaults      (CloudWatchConfig in resource namespace, defaults)
//	Level 7 — globalCWCfgDefaults     (CloudWatchConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.cloudwatch)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.cloudwatch)
//
// Scalar merge: firstNonNilBoolPtr for actionsEnabled; firstNonEmptyString for string fields.
// NamingTemplate: governed only at CloudWatchConfig levels (3-4 mandatory, 6-7 defaults).
// Tags: additive union merge across all mandatory levels / all defaults levels.
// SyncedLabels/SyncedAnnotations: additive union from CloudWatchConfig levels only.
func MergeCloudWatchCascade(
	globalKropathMandatory CloudWatchKropathSection, // level 1
	localKropathMandatory CloudWatchKropathSection, // level 2
	globalCWCfgMandatory CloudWatchConfigSection, // level 3
	localCWCfgMandatory CloudWatchConfigSection, // level 4
	localCWCfgDefaults CloudWatchConfigSection, // level 6
	globalCWCfgDefaults CloudWatchConfigSection, // level 7
	localKropathDefaults CloudWatchKropathSection, // level 8
	globalKropathDefaults CloudWatchKropathSection, // level 9
) EffectiveCloudWatchConfig {
	return EffectiveCloudWatchConfig{
		Mandatory: EffectiveCloudWatchSection{
			// actionsEnabled: boolean pointer — nil means level not participating.
			// Level 1 wins; nil is skipped (firstNonNilBoolPtr semantics).
			ActionsEnabled: firstNonNilBoolPtr(
				globalKropathMandatory.ActionsEnabled,
				localKropathMandatory.ActionsEnabled,
				globalCWCfgMandatory.ActionsEnabled,
				localCWCfgMandatory.ActionsEnabled,
			),
			TreatMissingData: firstNonEmptyString(
				globalKropathMandatory.TreatMissingData,
				localKropathMandatory.TreatMissingData,
				globalCWCfgMandatory.TreatMissingData,
				localCWCfgMandatory.TreatMissingData,
			),
			OutputFormat: firstNonEmptyString(
				globalKropathMandatory.OutputFormat,
				localKropathMandatory.OutputFormat,
				globalCWCfgMandatory.OutputFormat,
				localCWCfgMandatory.OutputFormat,
			),
			// NamingTemplate: CloudWatchConfig levels only (3, 4);
			// KropathConfig has no namingTemplate field for cloudwatch.
			NamingTemplate: firstNonEmptyString(
				globalCWCfgMandatory.NamingTemplate,
				localCWCfgMandatory.NamingTemplate,
			),
			// SyncedLabels: additive union from CloudWatchConfig levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localCWCfgMandatory.SyncedLabels,
				globalCWCfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localCWCfgMandatory.SyncedAnnotations,
				globalCWCfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localCWCfgMandatory.Tags,
				globalCWCfgMandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
		},
		Defaults: EffectiveCloudWatchSection{
			ActionsEnabled: firstNonNilBoolPtr(
				localCWCfgDefaults.ActionsEnabled,
				globalCWCfgDefaults.ActionsEnabled,
				localKropathDefaults.ActionsEnabled,
				globalKropathDefaults.ActionsEnabled,
			),
			TreatMissingData: firstNonEmptyString(
				localCWCfgDefaults.TreatMissingData,
				globalCWCfgDefaults.TreatMissingData,
				localKropathDefaults.TreatMissingData,
				globalKropathDefaults.TreatMissingData,
			),
			OutputFormat: firstNonEmptyString(
				localCWCfgDefaults.OutputFormat,
				globalCWCfgDefaults.OutputFormat,
				localKropathDefaults.OutputFormat,
				globalKropathDefaults.OutputFormat,
			),
			// NamingTemplate: CloudWatchConfig levels only (6, 7).
			NamingTemplate: firstNonEmptyString(
				localCWCfgDefaults.NamingTemplate,
				globalCWCfgDefaults.NamingTemplate,
			),
			// SyncedLabels: additive union from CloudWatchConfig levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalCWCfgDefaults.SyncedLabels,
				localCWCfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalCWCfgDefaults.SyncedAnnotations,
				localCWCfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalCWCfgDefaults.Tags,
				localCWCfgDefaults.Tags,
			),
		},
	}
}
