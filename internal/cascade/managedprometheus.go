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

// ManagedPrometheusKropathSection holds the Managed Prometheus-family governance fields from
// KropathConfig.spec.mandatory.managedprometheus / .defaults.managedprometheus (ADR-015 §3.5).
//
// Only 2 scalar fields are governed at the KropathConfig level: alias and logGroupARN.
// namingTemplate, syncedLabels, and syncedAnnotations are ManagedPrometheusConfig-only.
//
// Zero value of each field is the permissive sentinel (not enforced).
type ManagedPrometheusKropathSection struct {
	// Alias is the workspace alias to enforce org-wide.
	// Empty string = not enforced.
	Alias string `json:"alias,omitempty"`

	// LogGroupARN is the CloudWatch Logs log group ARN for AMP vended metrics.
	// Empty string = not enforced.
	LogGroupARN string `json:"logGroupARN,omitempty"`

	// Tags are tier-level cloud resource tags.
	// The reconciler populates this from KropathConfig.spec.mandatory.tags / .defaults.tags
	// so that tag union merge flows through MergeManagedPrometheusCascade alongside the
	// Managed Prometheus-specific fields.
	// nil / empty map = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// ManagedPrometheusConfigSection holds the Managed Prometheus governance fields from
// ManagedPrometheusConfig.spec.mandatory or ManagedPrometheusConfig.spec.defaults
// (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type ManagedPrometheusConfigSection struct {
	// Alias is the workspace alias. Empty string = not enforced.
	Alias string `json:"alias,omitempty"`

	// LogGroupARN is the CloudWatch Logs log group ARN. Empty string = not enforced.
	LogGroupARN string `json:"logGroupARN,omitempty"`

	// NamingTemplate is the workspace naming template (e.g. "{namespace}-{name}").
	// Governed only at ManagedPrometheusConfig levels 3-4 (mandatory) and 6-7 (defaults).
	// KropathConfig.managedprometheus does NOT carry namingTemplate.
	// Empty string = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created workspace resources.
	// Additive map merge across ManagedPrometheusConfig tiers only.
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to created workspace resources.
	// Additive map merge across ManagedPrometheusConfig tiers only.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`

	// Tags are cloud resource tags for this Managed Prometheus config profile.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// EffectiveManagedPrometheusSection is one tier (mandatory or defaults) of the merged
// Managed Prometheus governance result written into ManagedPrometheusConfig.status.effectiveConfig
// by the controller.
type EffectiveManagedPrometheusSection struct {
	Alias             string            `json:"alias,omitempty"`
	LogGroupARN       string            `json:"logGroupARN,omitempty"`
	NamingTemplate    string            `json:"namingTemplate,omitempty"`
	SyncedLabels      map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
	Tags              map[string]string `json:"tags,omitempty"`
}

// EffectiveManagedPrometheusConfig is the merged Managed Prometheus governance result written into
// ManagedPrometheusConfig.status.effectiveConfig by the controller.
type EffectiveManagedPrometheusConfig struct {
	Mandatory EffectiveManagedPrometheusSection `json:"mandatory"`
	Defaults  EffectiveManagedPrometheusSection `json:"defaults"`
}

// defaultNamingTemplate is the built-in fallback for defaults.namingTemplate when all
// ManagedPrometheusConfig tier sources are empty.
const defaultNamingTemplate = "{namespace}-{name}"

// MergeManagedPrometheusCascade merges Managed Prometheus governance fields from all cascade
// sources and returns the effective configuration to be written to status.effectiveConfig.
//
// Eight-level priority chain (ADR-015 §5):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.managedprometheus)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.managedprometheus)
//	Level 3 — globalMPCfgMandatory    (ManagedPrometheusConfig in kro-system, mandatory)
//	Level 4 — localMPCfgMandatory     (ManagedPrometheusConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localMPCfgDefaults      (ManagedPrometheusConfig in resource namespace, defaults)
//	Level 7 — globalMPCfgDefaults     (ManagedPrometheusConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.managedprometheus)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.managedprometheus)
//
// Scalar merge: firstNonEmptyString in priority order (lowest number wins).
// Tags: additive union merge across all four mandatory levels, all four defaults levels.
// SyncedLabels/SyncedAnnotations: additive union from ManagedPrometheusConfig levels only.
// NamingTemplate: governed only at ManagedPrometheusConfig levels; defaults falls back to
// "{namespace}-{name}" when all MPConfig defaults levels are empty.
func MergeManagedPrometheusCascade(
	globalKropathMandatory ManagedPrometheusKropathSection, // level 1
	localKropathMandatory ManagedPrometheusKropathSection, // level 2
	globalMPCfgMandatory ManagedPrometheusConfigSection, // level 3
	localMPCfgMandatory ManagedPrometheusConfigSection, // level 4
	localMPCfgDefaults ManagedPrometheusConfigSection, // level 6
	globalMPCfgDefaults ManagedPrometheusConfigSection, // level 7
	localKropathDefaults ManagedPrometheusKropathSection, // level 8
	globalKropathDefaults ManagedPrometheusKropathSection, // level 9
) EffectiveManagedPrometheusConfig {
	defaultsNamingTemplate := firstNonEmptyString(
		localMPCfgDefaults.NamingTemplate,
		globalMPCfgDefaults.NamingTemplate,
	)
	if defaultsNamingTemplate == "" {
		defaultsNamingTemplate = defaultNamingTemplate
	}

	return EffectiveManagedPrometheusConfig{
		Mandatory: EffectiveManagedPrometheusSection{
			Alias: firstNonEmptyString(
				globalKropathMandatory.Alias,
				localKropathMandatory.Alias,
				globalMPCfgMandatory.Alias,
				localMPCfgMandatory.Alias,
			),
			LogGroupARN: firstNonEmptyString(
				globalKropathMandatory.LogGroupARN,
				localKropathMandatory.LogGroupARN,
				globalMPCfgMandatory.LogGroupARN,
				localMPCfgMandatory.LogGroupARN,
			),
			// NamingTemplate: ManagedPrometheusConfig levels only (3, 4).
			NamingTemplate: firstNonEmptyString(
				globalMPCfgMandatory.NamingTemplate,
				localMPCfgMandatory.NamingTemplate,
			),
			// SyncedLabels: additive union from ManagedPrometheusConfig levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localMPCfgMandatory.SyncedLabels,
				globalMPCfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localMPCfgMandatory.SyncedAnnotations,
				globalMPCfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localMPCfgMandatory.Tags,
				globalMPCfgMandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
		},
		Defaults: EffectiveManagedPrometheusSection{
			Alias: firstNonEmptyString(
				localMPCfgDefaults.Alias,
				globalMPCfgDefaults.Alias,
				localKropathDefaults.Alias,
				globalKropathDefaults.Alias,
			),
			LogGroupARN: firstNonEmptyString(
				localMPCfgDefaults.LogGroupARN,
				globalMPCfgDefaults.LogGroupARN,
				localKropathDefaults.LogGroupARN,
				globalKropathDefaults.LogGroupARN,
			),
			// NamingTemplate: ManagedPrometheusConfig levels only (6, 7); falls back to
			// "{namespace}-{name}" when all MPConfig defaults levels are empty.
			NamingTemplate: defaultsNamingTemplate,
			// SyncedLabels: additive union from ManagedPrometheusConfig levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalMPCfgDefaults.SyncedLabels,
				localMPCfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalMPCfgDefaults.SyncedAnnotations,
				localMPCfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalMPCfgDefaults.Tags,
				localMPCfgDefaults.Tags,
			),
		},
	}
}
