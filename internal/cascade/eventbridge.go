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

// EventBridgeKropathSection holds the EventBridge-family governance fields from
// KropathConfig.spec.mandatory.eventbridge / .defaults.eventbridge (ADR-015 §3.5).
//
// Only archiveRetentionDays is governed at the KropathConfig level; namingTemplate,
// syncedLabels, and syncedAnnotations are EventBridgeConfig-only (family design §8).
//
// Zero value for ArchiveRetentionDays is 0 (not enforced / permissive sentinel).
type EventBridgeKropathSection struct {
	// ArchiveRetentionDays is the org-wide archive retention floor.
	// 0 = not enforced (mandatory) or indefinite retention (defaults).
	ArchiveRetentionDays int64 `json:"archiveRetentionDays,omitempty"`

	// Tags are tier-level cloud resource tags from KropathConfig.spec.mandatory.tags
	// or .defaults.tags. Populated by the reconciler so that tag cascade flows
	// through MergeEventBridgeCascade alongside EventBridge-specific fields.
	Tags map[string]string `json:"tags,omitempty"`
}

// EventBridgeConfigSection holds the EventBridge governance fields from
// EventBridgeConfig.spec.mandatory or EventBridgeConfig.spec.defaults
// (per-type ResourceConfig, ADR-015 §3.5).
type EventBridgeConfigSection struct {
	// ArchiveRetentionDays is the archive retention setting.
	// Mandatory non-zero = enforced minimum retention; defaults non-zero = applied as default.
	// 0 = not enforced (mandatory) or indefinite retention (defaults).
	ArchiveRetentionDays int64 `json:"archiveRetentionDays,omitempty"`

	// NamingTemplate is the EventBridge resource naming template (e.g. "{namespace}-{name}").
	// Governed only at EventBridgeConfig levels 3-4 (mandatory) and 6-7 (defaults).
	// Empty string = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// Tags are cloud resource tags for this EventBridge config profile.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created EventBridge resources.
	// Additive map merge across EventBridgeConfig tiers only.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to created EventBridge resources.
	// Additive map merge across EventBridgeConfig tiers only.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveEventBridgeSection is one tier (mandatory or defaults) of the merged EventBridge
// governance result written into EventBridgeConfig.status.effectiveConfig by the controller.
type EffectiveEventBridgeSection struct {
	ArchiveRetentionDays int64             `json:"archiveRetentionDays,omitempty"`
	NamingTemplate       string            `json:"namingTemplate,omitempty"`
	Tags                 map[string]string `json:"tags,omitempty"`
	SyncedLabels         map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations    map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveEventBridgeConfig is the merged EventBridge governance result written into
// EventBridgeConfig.status.effectiveConfig by the controller.
type EffectiveEventBridgeConfig struct {
	Mandatory EffectiveEventBridgeSection `json:"mandatory"`
	Defaults  EffectiveEventBridgeSection `json:"defaults"`
}

// MergeEventBridgeCascade merges EventBridge governance fields from all cascade sources and
// returns the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for EventBridge (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.eventbridge)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.eventbridge)
//	Level 3 — globalEBCfgMandatory    (EventBridgeConfig in kro-system, mandatory)
//	Level 4 — localEBCfgMandatory     (EventBridgeConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localEBCfgDefaults      (EventBridgeConfig in resource namespace, defaults)
//	Level 7 — globalEBCfgDefaults     (EventBridgeConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.eventbridge)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.eventbridge)
//
// Scalar integer merge: firstNonZeroInt64 in priority order (lowest level wins).
// String merge: firstNonEmptyString in priority order.
// NamingTemplate: EventBridgeConfig levels only (3-4 mandatory, 6-7 defaults).
// Tags: additive union across all four mandatory sources (L1 wins on key conflict) and all four
// defaults sources (L6 wins on key conflict).
// SyncedLabels/SyncedAnnotations: additive union from EventBridgeConfig levels only
// (mandatory: L3 wins; defaults: L6 wins).
func MergeEventBridgeCascade(
	globalKropathMandatory EventBridgeKropathSection, // level 1
	localKropathMandatory EventBridgeKropathSection,  // level 2
	globalEBCfgMandatory EventBridgeConfigSection,    // level 3
	localEBCfgMandatory EventBridgeConfigSection,     // level 4
	localEBCfgDefaults EventBridgeConfigSection,      // level 6
	globalEBCfgDefaults EventBridgeConfigSection,     // level 7
	localKropathDefaults EventBridgeKropathSection,   // level 8
	globalKropathDefaults EventBridgeKropathSection,  // level 9
) EffectiveEventBridgeConfig {
	return EffectiveEventBridgeConfig{
		Mandatory: EffectiveEventBridgeSection{
			ArchiveRetentionDays: firstNonZeroInt64(
				globalKropathMandatory.ArchiveRetentionDays,
				localKropathMandatory.ArchiveRetentionDays,
				globalEBCfgMandatory.ArchiveRetentionDays,
				localEBCfgMandatory.ArchiveRetentionDays,
			),
			// NamingTemplate: EventBridgeConfig levels only (3, 4). KropathConfig has no namingTemplate.
			NamingTemplate: firstNonEmptyString(
				globalEBCfgMandatory.NamingTemplate,
				localEBCfgMandatory.NamingTemplate,
			),
			// SyncedLabels: additive union from EventBridgeConfig mandatory levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localEBCfgMandatory.SyncedLabels,
				globalEBCfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localEBCfgMandatory.SyncedAnnotations,
				globalEBCfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localEBCfgMandatory.Tags,
				globalEBCfgMandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
		},
		Defaults: EffectiveEventBridgeSection{
			ArchiveRetentionDays: firstNonZeroInt64(
				localEBCfgDefaults.ArchiveRetentionDays,
				globalEBCfgDefaults.ArchiveRetentionDays,
				localKropathDefaults.ArchiveRetentionDays,
				globalKropathDefaults.ArchiveRetentionDays,
			),
			// NamingTemplate: EventBridgeConfig levels only (6, 7).
			NamingTemplate: firstNonEmptyString(
				localEBCfgDefaults.NamingTemplate,
				globalEBCfgDefaults.NamingTemplate,
			),
			// SyncedLabels: additive union from EventBridgeConfig defaults levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalEBCfgDefaults.SyncedLabels,
				localEBCfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalEBCfgDefaults.SyncedAnnotations,
				localEBCfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalEBCfgDefaults.Tags,
				localEBCfgDefaults.Tags,
			),
		},
	}
}
