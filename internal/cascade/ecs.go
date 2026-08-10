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

// ECSKropathSection holds the ECS-family governance fields from
// KropathConfig.spec.mandatory.ecs / .defaults.ecs (ADR-015 §3.5)
// PLUS the tier-level tags/syncedLabels/syncedAnnotations from
// KropathConfig.spec.mandatory.tags (populated by the reconciler so that
// tag and synced-label cascade flows through MergeECSCascade).
//
// Only the two cross-type fields are governed here: containerInsights and
// defaultLaunchType. All other ECS fields (defaultPlatformVersion,
// defaultNetworkMode, defaultTaskCPU, defaultTaskMemory, namingTemplate)
// are ECSConfig-only (family design §8).
//
// Zero value of each field is the permissive sentinel (not enforced).
type ECSKropathSection struct {
	// ContainerInsights enforces Container Insights enabled on all clusters
	// org-wide when true. false (zero value) = not enforced.
	ContainerInsights bool `json:"containerInsights,omitempty"`

	// DefaultLaunchType enforces an org-wide launch type (EC2 | FARGATE).
	// Empty string (zero value) = not enforced.
	DefaultLaunchType string `json:"defaultLaunchType,omitempty"`

	// Tags are tier-level cloud resource tags from KropathConfig.spec.mandatory.tags
	// or .defaults.tags. Populated by the reconciler from the tier-level field, not
	// from spec.mandatory.ecs. nil / empty map (zero value) = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are Kubernetes labels from KropathConfig.spec.mandatory.syncedLabels
	// or .defaults.syncedLabels. Populated by the reconciler.
	// nil / empty map (zero value) = no synced labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations from KropathConfig.spec.mandatory.syncedAnnotations
	// or .defaults.syncedAnnotations. Populated by the reconciler.
	// nil / empty map (zero value) = no synced annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// ECSConfigSection holds the ECS governance fields from
// ECSConfig.spec.mandatory or ECSConfig.spec.defaults
// (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type ECSConfigSection struct {
	// ContainerInsights enforces Container Insights when true.
	// false = not enforced.
	ContainerInsights bool `json:"containerInsights,omitempty"`

	// DefaultLaunchType enforces an ECS launch type (EC2 | FARGATE).
	// Empty string = not enforced.
	DefaultLaunchType string `json:"defaultLaunchType,omitempty"`

	// DefaultPlatformVersion enforces a Fargate platform version (e.g. "1.4.0").
	// Governed at ECSConfig levels only (not in KropathConfig).
	// Empty string = not enforced.
	DefaultPlatformVersion string `json:"defaultPlatformVersion,omitempty"`

	// DefaultNetworkMode enforces a task definition network mode
	// (none | bridge | awsvpc | host).
	// Governed at ECSConfig levels only.
	// Empty string = not enforced.
	DefaultNetworkMode string `json:"defaultNetworkMode,omitempty"`

	// DefaultTaskCPU enforces a Fargate task-level CPU allocation (e.g. "256").
	// Governed at ECSConfig levels only.
	// Empty string = not enforced.
	DefaultTaskCPU string `json:"defaultTaskCPU,omitempty"`

	// DefaultTaskMemory enforces a Fargate task-level memory allocation (e.g. "512").
	// Governed at ECSConfig levels only.
	// Empty string = not enforced.
	DefaultTaskMemory string `json:"defaultTaskMemory,omitempty"`

	// NamingTemplate is the ECS resource naming template (e.g. "{namespace}-{name}").
	// Governed at ECSConfig levels only.
	// Empty string = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// Tags are cloud resource tags for this ECS config profile.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created ECS resources.
	// Additive map merge across ECSConfig tiers only.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to created ECS resources.
	// Additive map merge across ECSConfig tiers only.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveECSSection is one tier (mandatory or defaults) of the merged ECS
// governance result written into ECSConfig.status.effectiveConfig by the controller.
type EffectiveECSSection struct {
	ContainerInsights      bool              `json:"containerInsights,omitempty"`
	DefaultLaunchType      string            `json:"defaultLaunchType,omitempty"`
	DefaultPlatformVersion string            `json:"defaultPlatformVersion,omitempty"`
	DefaultNetworkMode     string            `json:"defaultNetworkMode,omitempty"`
	DefaultTaskCPU         string            `json:"defaultTaskCPU,omitempty"`
	DefaultTaskMemory      string            `json:"defaultTaskMemory,omitempty"`
	NamingTemplate         string            `json:"namingTemplate,omitempty"`
	Tags                   map[string]string `json:"tags,omitempty"`
	SyncedLabels           map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations      map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveECSConfig is the merged ECS governance result written into
// ECSConfig.status.effectiveConfig by the controller.
type EffectiveECSConfig struct {
	Mandatory EffectiveECSSection `json:"mandatory"`
	Defaults  EffectiveECSSection `json:"defaults"`
}

// MergeECSCascade merges ECS governance fields from all cascade sources and
// returns the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for ECS (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.ecs)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.ecs)
//	Level 3 — globalECSCfgMandatory   (ECSConfig in kro-system, mandatory)
//	Level 4 — localECSCfgMandatory    (ECSConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localECSCfgDefaults     (ECSConfig in resource namespace, defaults)
//	Level 7 — globalECSCfgDefaults    (ECSConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.ecs)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.ecs)
//
// containerInsights: firstTrue in priority order (level 1 wins; lowest level that is true wins).
// defaultLaunchType: firstNonEmptyString in priority order.
// defaultPlatformVersion, defaultNetworkMode, defaultTaskCPU, defaultTaskMemory, namingTemplate:
//   ECSConfig levels only (3-4 mandatory, 6-7 defaults).
// Tags: additive union across all mandatory sources (L1 wins on key conflict) and all defaults
//   sources (L6 wins on key conflict). ECSKropathSection.Tags carries the tier-level
//   KropathConfig.mandatory.tags (populated by the reconciler).
// SyncedLabels/SyncedAnnotations: additive union from ECSConfig levels only
//   (mandatory: L3 wins; defaults: L6 wins).
func MergeECSCascade(
	globalKropathMandatory ECSKropathSection, // level 1
	localKropathMandatory ECSKropathSection,  // level 2
	globalECSCfgMandatory ECSConfigSection,   // level 3
	localECSCfgMandatory ECSConfigSection,    // level 4
	localECSCfgDefaults ECSConfigSection,     // level 6
	globalECSCfgDefaults ECSConfigSection,    // level 7
	localKropathDefaults ECSKropathSection,   // level 8
	globalKropathDefaults ECSKropathSection,  // level 9
) EffectiveECSConfig {
	return EffectiveECSConfig{
		Mandatory: EffectiveECSSection{
			ContainerInsights: firstTrue(
				globalKropathMandatory.ContainerInsights, // level 1
				localKropathMandatory.ContainerInsights,  // level 2
				globalECSCfgMandatory.ContainerInsights,  // level 3
				localECSCfgMandatory.ContainerInsights,   // level 4
			),
			DefaultLaunchType: firstNonEmptyString(
				globalKropathMandatory.DefaultLaunchType, // level 1
				localKropathMandatory.DefaultLaunchType,  // level 2
				globalECSCfgMandatory.DefaultLaunchType,  // level 3
				localECSCfgMandatory.DefaultLaunchType,   // level 4
			),
			// DefaultPlatformVersion: ECSConfig levels only (3, 4).
			DefaultPlatformVersion: firstNonEmptyString(
				globalECSCfgMandatory.DefaultPlatformVersion, // level 3
				localECSCfgMandatory.DefaultPlatformVersion,  // level 4
			),
			// DefaultNetworkMode: ECSConfig levels only (3, 4).
			DefaultNetworkMode: firstNonEmptyString(
				globalECSCfgMandatory.DefaultNetworkMode, // level 3
				localECSCfgMandatory.DefaultNetworkMode,  // level 4
			),
			// DefaultTaskCPU: ECSConfig levels only (3, 4).
			DefaultTaskCPU: firstNonEmptyString(
				globalECSCfgMandatory.DefaultTaskCPU, // level 3
				localECSCfgMandatory.DefaultTaskCPU,  // level 4
			),
			// DefaultTaskMemory: ECSConfig levels only (3, 4).
			DefaultTaskMemory: firstNonEmptyString(
				globalECSCfgMandatory.DefaultTaskMemory, // level 3
				localECSCfgMandatory.DefaultTaskMemory,  // level 4
			),
			// NamingTemplate: ECSConfig levels only (3, 4).
			NamingTemplate: firstNonEmptyString(
				globalECSCfgMandatory.NamingTemplate, // level 3
				localECSCfgMandatory.NamingTemplate,  // level 4
			),
			// SyncedLabels: additive union from ECSConfig mandatory levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localECSCfgMandatory.SyncedLabels,
				globalECSCfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localECSCfgMandatory.SyncedAnnotations,
				globalECSCfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localECSCfgMandatory.Tags,    // level 4 (lowest priority, set first)
				globalECSCfgMandatory.Tags,   // level 3
				localKropathMandatory.Tags,   // level 2
				globalKropathMandatory.Tags,  // level 1 (highest priority, last to write)
			),
		},
		Defaults: EffectiveECSSection{
			ContainerInsights: firstTrue(
				localECSCfgDefaults.ContainerInsights,   // level 6
				globalECSCfgDefaults.ContainerInsights,  // level 7
				localKropathDefaults.ContainerInsights,  // level 8
				globalKropathDefaults.ContainerInsights, // level 9
			),
			DefaultLaunchType: firstNonEmptyString(
				localECSCfgDefaults.DefaultLaunchType,   // level 6
				globalECSCfgDefaults.DefaultLaunchType,  // level 7
				localKropathDefaults.DefaultLaunchType,  // level 8
				globalKropathDefaults.DefaultLaunchType, // level 9
			),
			// DefaultPlatformVersion: ECSConfig levels only (6, 7).
			DefaultPlatformVersion: firstNonEmptyString(
				localECSCfgDefaults.DefaultPlatformVersion,  // level 6
				globalECSCfgDefaults.DefaultPlatformVersion, // level 7
			),
			// DefaultNetworkMode: ECSConfig levels only (6, 7).
			DefaultNetworkMode: firstNonEmptyString(
				localECSCfgDefaults.DefaultNetworkMode,  // level 6
				globalECSCfgDefaults.DefaultNetworkMode, // level 7
			),
			// DefaultTaskCPU: ECSConfig levels only (6, 7).
			DefaultTaskCPU: firstNonEmptyString(
				localECSCfgDefaults.DefaultTaskCPU,  // level 6
				globalECSCfgDefaults.DefaultTaskCPU, // level 7
			),
			// DefaultTaskMemory: ECSConfig levels only (6, 7).
			DefaultTaskMemory: firstNonEmptyString(
				localECSCfgDefaults.DefaultTaskMemory,  // level 6
				globalECSCfgDefaults.DefaultTaskMemory, // level 7
			),
			// NamingTemplate: ECSConfig levels only (6, 7).
			NamingTemplate: firstNonEmptyString(
				localECSCfgDefaults.NamingTemplate,  // level 6
				globalECSCfgDefaults.NamingTemplate, // level 7
			),
			// SyncedLabels: additive union from ECSConfig defaults levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalECSCfgDefaults.SyncedLabels,
				localECSCfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalECSCfgDefaults.SyncedAnnotations,
				localECSCfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags, // level 9 (lowest priority)
				localKropathDefaults.Tags,  // level 8
				globalECSCfgDefaults.Tags,  // level 7
				localECSCfgDefaults.Tags,   // level 6 (highest priority)
			),
		},
	}
}
