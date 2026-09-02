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

// BedrockKropathSection holds the Bedrock-family governance fields promoted from
// KropathConfig.spec.mandatory.bedrock / .defaults.bedrock (ADR-015 §3.5).
//
// Only 3 fields are governed at the KropathConfig level: guardrailIdentifier,
// guardrailVersion, and allowedModels. foundationModel, maxIterations, maxTokens,
// timeoutSeconds, and idleSessionTTLInSeconds are per-resource choices (family design §8)
// and are NOT in KropathConfig.
//
// Zero value of each field is the permissive sentinel (not enforced).
type BedrockKropathSection struct {
	// GuardrailIdentifier enforces a guardrail ARN/ID org-wide.
	// Empty string = not enforced.
	GuardrailIdentifier string `json:"guardrailIdentifier,omitempty"`

	// GuardrailVersion enforces a guardrail version org-wide.
	// Empty string = not enforced. Must be paired with GuardrailIdentifier.
	GuardrailVersion string `json:"guardrailVersion,omitempty"`

	// AllowedModels is the org-wide allowlist of permitted foundation model IDs.
	// nil / empty slice = no restriction.
	AllowedModels []string `json:"allowedModels,omitempty"`

	// Tags are tier-level cloud resource tags from KropathConfig.spec.mandatory.tags /
	// .defaults.tags. Populated by the reconciler so they flow through MergeBedrockCascade.
	// nil / empty map = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// BedrockConfigSection holds the Bedrock governance fields from
// BedrockConfig.spec.mandatory or BedrockConfig.spec.defaults
// (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type BedrockConfigSection struct {
	// FoundationModel is the enforced Amazon Bedrock model ID (e.g. "anthropic.claude-3-5-sonnet-20241022-v2:0").
	// Not in KropathConfig — per-resource choice (family design §8). Empty string = not enforced.
	FoundationModel string `json:"foundationModel,omitempty"`

	// GuardrailIdentifier enforces a guardrail ARN/ID for this config profile.
	// Empty string = not enforced. Must be paired with GuardrailVersion.
	GuardrailIdentifier string `json:"guardrailIdentifier,omitempty"`

	// GuardrailVersion enforces a guardrail version for this config profile.
	// Empty string = not enforced. Must be paired with GuardrailIdentifier.
	GuardrailVersion string `json:"guardrailVersion,omitempty"`

	// AllowedModels is the per-profile allowlist of permitted foundation model IDs.
	// nil / empty slice = no restriction.
	AllowedModels []string `json:"allowedModels,omitempty"`

	// MaxIterations caps agent loop iterations. 0 = not enforced.
	MaxIterations int64 `json:"maxIterations,omitempty"`

	// MaxTokens caps token output per invocation. 0 = not enforced.
	MaxTokens int64 `json:"maxTokens,omitempty"`

	// TimeoutSeconds caps the agent session timeout. 0 = not enforced.
	TimeoutSeconds int64 `json:"timeoutSeconds,omitempty"`

	// IdleSessionTTLInSeconds caps idle session TTL. 0 = not enforced.
	IdleSessionTTLInSeconds int64 `json:"idleSessionTTLInSeconds,omitempty"`

	// NamingTemplate is the resource naming template. Empty string = not enforced.
	// Governed only at BedrockConfig levels 3-4 (mandatory) and 6-7 (defaults).
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created resources.
	// Additive map merge across BedrockConfig tiers only.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to created resources.
	// Additive map merge across BedrockConfig tiers only.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`

	// Tags are cloud resource tags for this Bedrock config profile.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// EffectiveBedrockSection is one tier (mandatory or defaults) of the merged
// Bedrock governance result written into BedrockConfig.status.effectiveConfig
// by the controller.
type EffectiveBedrockSection struct {
	FoundationModel         string            `json:"foundationModel,omitempty"`
	GuardrailIdentifier     string            `json:"guardrailIdentifier,omitempty"`
	GuardrailVersion        string            `json:"guardrailVersion,omitempty"`
	AllowedModels           []string          `json:"allowedModels,omitempty"`
	MaxIterations           int64             `json:"maxIterations,omitempty"`
	MaxTokens               int64             `json:"maxTokens,omitempty"`
	TimeoutSeconds          int64             `json:"timeoutSeconds,omitempty"`
	IdleSessionTTLInSeconds int64             `json:"idleSessionTTLInSeconds,omitempty"`
	NamingTemplate          string            `json:"namingTemplate,omitempty"`
	SyncedLabels            map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations       map[string]string `json:"syncedAnnotations,omitempty"`
	Tags                    map[string]string `json:"tags,omitempty"`
}

// EffectiveBedrockConfig is the merged Bedrock governance result written into
// BedrockConfig.status.effectiveConfig by the controller.
type EffectiveBedrockConfig struct {
	Mandatory EffectiveBedrockSection `json:"mandatory"`
	Defaults  EffectiveBedrockSection `json:"defaults"`
}

// MergeBedrockCascade merges Bedrock governance fields from all cascade
// sources and returns the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for Bedrock (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.bedrock)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.bedrock)
//	Level 3 — globalBCCfgMandatory    (BedrockConfig in kro-system, mandatory)
//	Level 4 — localBCCfgMandatory     (BedrockConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localBCCfgDefaults      (BedrockConfig in resource namespace, defaults)
//	Level 7 — globalBCCfgDefaults     (BedrockConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.bedrock)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.bedrock)
//
// Scalar merge: firstNonEmptyString / firstNonZeroInt64 in priority order.
// String slices (allowedModels): firstNonEmptyStrings in priority order.
// Tags: additive union merge across all four mandatory/defaults levels.
// SyncedLabels/SyncedAnnotations: additive union from BedrockConfig levels only.
// NamingTemplate: BedrockConfig levels only (3-4 mandatory, 6-7 defaults).
//
// foundationModel, maxIterations, maxTokens, timeoutSeconds, idleSessionTTLInSeconds
// are NOT in KropathConfig (levels 1-2 and 8-9 are empty for these fields).
func MergeBedrockCascade(
	globalKropathMandatory BedrockKropathSection, // level 1
	localKropathMandatory BedrockKropathSection, // level 2
	globalBCCfgMandatory BedrockConfigSection, // level 3
	localBCCfgMandatory BedrockConfigSection, // level 4
	localBCCfgDefaults BedrockConfigSection, // level 6
	globalBCCfgDefaults BedrockConfigSection, // level 7
	localKropathDefaults BedrockKropathSection, // level 8
	globalKropathDefaults BedrockKropathSection, // level 9
) EffectiveBedrockConfig {
	return EffectiveBedrockConfig{
		Mandatory: EffectiveBedrockSection{
			// guardrailIdentifier: all 4 mandatory levels (KropathConfig at 1-2, BedrockConfig at 3-4)
			GuardrailIdentifier: firstNonEmptyString(
				globalKropathMandatory.GuardrailIdentifier, // level 1
				localKropathMandatory.GuardrailIdentifier,  // level 2
				globalBCCfgMandatory.GuardrailIdentifier,   // level 3
				localBCCfgMandatory.GuardrailIdentifier,    // level 4
			),
			// guardrailVersion: same four-level chain as guardrailIdentifier
			GuardrailVersion: firstNonEmptyString(
				globalKropathMandatory.GuardrailVersion, // level 1
				localKropathMandatory.GuardrailVersion,  // level 2
				globalBCCfgMandatory.GuardrailVersion,   // level 3
				localBCCfgMandatory.GuardrailVersion,    // level 4
			),
			// allowedModels: all 4 mandatory levels
			AllowedModels: firstNonEmptyStrings(
				globalKropathMandatory.AllowedModels, // level 1
				localKropathMandatory.AllowedModels,  // level 2
				globalBCCfgMandatory.AllowedModels,   // level 3
				localBCCfgMandatory.AllowedModels,    // level 4
			),
			// foundationModel: BedrockConfig levels only (3, 4); not in KropathConfig
			FoundationModel: firstNonEmptyString(
				globalBCCfgMandatory.FoundationModel, // level 3
				localBCCfgMandatory.FoundationModel,  // level 4
			),
			// maxIterations: BedrockConfig levels only (3, 4)
			MaxIterations: firstNonZeroInt64(
				globalBCCfgMandatory.MaxIterations, // level 3
				localBCCfgMandatory.MaxIterations,  // level 4
			),
			// maxTokens: BedrockConfig levels only (3, 4)
			MaxTokens: firstNonZeroInt64(
				globalBCCfgMandatory.MaxTokens, // level 3
				localBCCfgMandatory.MaxTokens,  // level 4
			),
			// timeoutSeconds: BedrockConfig levels only (3, 4)
			TimeoutSeconds: firstNonZeroInt64(
				globalBCCfgMandatory.TimeoutSeconds, // level 3
				localBCCfgMandatory.TimeoutSeconds,  // level 4
			),
			// idleSessionTTLInSeconds: BedrockConfig levels only (3, 4)
			IdleSessionTTLInSeconds: firstNonZeroInt64(
				globalBCCfgMandatory.IdleSessionTTLInSeconds, // level 3
				localBCCfgMandatory.IdleSessionTTLInSeconds,  // level 4
			),
			// NamingTemplate: BedrockConfig levels only (3, 4)
			NamingTemplate: firstNonEmptyString(
				globalBCCfgMandatory.NamingTemplate, // level 3
				localBCCfgMandatory.NamingTemplate,  // level 4
			),
			// SyncedLabels: additive union from BedrockConfig levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localBCCfgMandatory.SyncedLabels,
				globalBCCfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localBCCfgMandatory.SyncedAnnotations,
				globalBCCfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localBCCfgMandatory.Tags,
				globalBCCfgMandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
		},
		Defaults: EffectiveBedrockSection{
			// guardrailIdentifier: all 4 defaults levels (BedrockConfig at 6-7, KropathConfig at 8-9)
			GuardrailIdentifier: firstNonEmptyString(
				localBCCfgDefaults.GuardrailIdentifier,    // level 6
				globalBCCfgDefaults.GuardrailIdentifier,   // level 7
				localKropathDefaults.GuardrailIdentifier,  // level 8
				globalKropathDefaults.GuardrailIdentifier, // level 9
			),
			// guardrailVersion: same four-level defaults chain
			GuardrailVersion: firstNonEmptyString(
				localBCCfgDefaults.GuardrailVersion,    // level 6
				globalBCCfgDefaults.GuardrailVersion,   // level 7
				localKropathDefaults.GuardrailVersion,  // level 8
				globalKropathDefaults.GuardrailVersion, // level 9
			),
			// allowedModels: all 4 defaults levels
			AllowedModels: firstNonEmptyStrings(
				localBCCfgDefaults.AllowedModels,    // level 6
				globalBCCfgDefaults.AllowedModels,   // level 7
				localKropathDefaults.AllowedModels,  // level 8
				globalKropathDefaults.AllowedModels, // level 9
			),
			// foundationModel: BedrockConfig levels only (6, 7); not in KropathConfig
			FoundationModel: firstNonEmptyString(
				localBCCfgDefaults.FoundationModel,  // level 6
				globalBCCfgDefaults.FoundationModel, // level 7
			),
			// maxIterations: BedrockConfig levels only (6, 7)
			MaxIterations: firstNonZeroInt64(
				localBCCfgDefaults.MaxIterations,  // level 6
				globalBCCfgDefaults.MaxIterations, // level 7
			),
			// maxTokens: BedrockConfig levels only (6, 7)
			MaxTokens: firstNonZeroInt64(
				localBCCfgDefaults.MaxTokens,  // level 6
				globalBCCfgDefaults.MaxTokens, // level 7
			),
			// timeoutSeconds: BedrockConfig levels only (6, 7)
			TimeoutSeconds: firstNonZeroInt64(
				localBCCfgDefaults.TimeoutSeconds,  // level 6
				globalBCCfgDefaults.TimeoutSeconds, // level 7
			),
			// idleSessionTTLInSeconds: BedrockConfig levels only (6, 7)
			IdleSessionTTLInSeconds: firstNonZeroInt64(
				localBCCfgDefaults.IdleSessionTTLInSeconds,  // level 6
				globalBCCfgDefaults.IdleSessionTTLInSeconds, // level 7
			),
			// NamingTemplate: BedrockConfig levels only (6, 7)
			NamingTemplate: firstNonEmptyString(
				localBCCfgDefaults.NamingTemplate,  // level 6
				globalBCCfgDefaults.NamingTemplate, // level 7
			),
			// SyncedLabels: additive union from BedrockConfig levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalBCCfgDefaults.SyncedLabels,
				localBCCfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalBCCfgDefaults.SyncedAnnotations,
				localBCCfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalBCCfgDefaults.Tags,
				localBCCfgDefaults.Tags,
			),
		},
	}
}
