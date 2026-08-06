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

// LambdaKropathSection holds the Lambda-family governance fields from
// KropathConfig.spec.mandatory.lambda / .defaults.lambda (ADR-015 §3.5),
// PLUS the tier-level tags, syncedLabels, and syncedAnnotations from
// KropathConfig.spec.mandatory.tags / .syncedLabels / .syncedAnnotations.
//
// Five governance fields live under spec.mandatory.lambda / spec.defaults.lambda.
// Tags, syncedLabels, and syncedAnnotations are org-wide tier-level fields
// (not under spec.lambda) and are populated by the reconciler so that the
// full tag/label/annotation cascade flows through MergeLambdaCascade.
//
// Zero value of each field is the permissive sentinel (not enforced).
type LambdaKropathSection struct {
	// Runtime is the org-wide enforced Lambda runtime (e.g. "python3.12").
	// Empty string = not enforced.
	Runtime string `json:"runtime,omitempty"`

	// MemorySize is the org-wide mandatory memory ceiling in MB.
	// 0 = not enforced; positive value = hard ceiling applied to all functions.
	MemorySize int64 `json:"memorySize,omitempty"`

	// Timeout is the org-wide mandatory execution timeout ceiling in seconds.
	// 0 = not enforced; positive value = hard ceiling.
	Timeout int64 `json:"timeout,omitempty"`

	// TracingMode is the org-wide enforced AWS X-Ray tracing mode.
	// Empty string = not enforced; "PassThrough" or "Active".
	TracingMode string `json:"tracingMode,omitempty"`

	// KmsKeyArn is the org-wide mandatory KMS key ARN for environment variable encryption.
	// Empty string = not enforced.
	KmsKeyArn string `json:"kmsKeyArn,omitempty"`

	// Tags are tier-level cloud resource tags from KropathConfig.spec.mandatory.tags
	// (org-wide, not Lambda-specific). Populated by the reconciler.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are org-wide Kubernetes labels from KropathConfig.spec.mandatory.syncedLabels.
	// Populated by the reconciler.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are org-wide Kubernetes annotations from KropathConfig.spec.mandatory.syncedAnnotations.
	// Populated by the reconciler.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// LambdaConfigSection holds the Lambda governance fields from LambdaConfig.spec.mandatory
// or LambdaConfig.spec.defaults (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type LambdaConfigSection struct {
	// Runtime is the enforced Lambda runtime for this profile.
	// Empty string = not enforced.
	Runtime string `json:"runtime,omitempty"`

	// MemorySize is the memory ceiling in MB for this profile.
	// 0 = not enforced (mandatory) or no default (defaults).
	MemorySize int64 `json:"memorySize,omitempty"`

	// Timeout is the execution timeout ceiling in seconds for this profile.
	// 0 = not enforced (mandatory) or no default (defaults).
	Timeout int64 `json:"timeout,omitempty"`

	// TracingMode is the AWS X-Ray tracing mode for this profile.
	// Empty string = not enforced.
	TracingMode string `json:"tracingMode,omitempty"`

	// KmsKeyArn is the KMS key ARN for environment variable encryption.
	// Empty string = not enforced.
	KmsKeyArn string `json:"kmsKeyArn,omitempty"`

	// NamingTemplate is the Lambda function naming template (e.g. "{namespace}-{name}").
	// Governed only at LambdaConfig levels (not in KropathConfig).
	// Empty string = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// Tags are cloud resource tags for this Lambda config profile.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to Lambda resources (ADR-015 §6.1).
	// Applied as both K8s labels and cloud tags.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to Lambda resources.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveLambdaSection is one tier (mandatory or defaults) of the merged Lambda governance
// result written into LambdaConfig.status.effectiveConfig by the controller.
type EffectiveLambdaSection struct {
	Runtime           string            `json:"runtime,omitempty"`
	MemorySize        int64             `json:"memorySize,omitempty"`
	Timeout           int64             `json:"timeout,omitempty"`
	TracingMode       string            `json:"tracingMode,omitempty"`
	KmsKeyArn         string            `json:"kmsKeyArn,omitempty"`
	NamingTemplate    string            `json:"namingTemplate,omitempty"`
	Tags              map[string]string `json:"tags,omitempty"`
	SyncedLabels      map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveLambdaConfig is the merged Lambda governance result written into
// LambdaConfig.status.effectiveConfig by the controller.
type EffectiveLambdaConfig struct {
	Mandatory EffectiveLambdaSection `json:"mandatory"`
	Defaults  EffectiveLambdaSection `json:"defaults"`
}

// MergeLambdaCascade merges Lambda governance fields from all cascade sources and
// returns the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for Lambda (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.lambda)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.lambda)
//	Level 3 — globalLambdaCfgMandatory (LambdaConfig in kro-system, mandatory)
//	Level 4 — localLambdaCfgMandatory  (LambdaConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localLambdaCfgDefaults   (LambdaConfig in resource namespace, defaults)
//	Level 7 — globalLambdaCfgDefaults  (LambdaConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults     (KropathConfig in resource namespace, defaults.lambda)
//	Level 9 — globalKropathDefaults    (KropathConfig in kro-system, defaults.lambda)
//
// Scalar string merge (runtime, tracingMode, kmsKeyArn): firstNonEmptyString in priority order.
// Scalar integer merge (memorySize, timeout): firstNonZeroInt64 in priority order.
// NamingTemplate: LambdaConfig levels only (3-4 mandatory, 6-7 defaults); not in KropathConfig.
// Tags: additive union across all four mandatory sources (L1 wins on key conflict) and all four
// defaults sources (L6 wins on key conflict).
// SyncedLabels/SyncedAnnotations: additive union from all sources (mandatory: L1 wins;
// defaults: L6 wins).
func MergeLambdaCascade(
	// Mandatory inputs (highest → lowest priority)
	globalKropathMandatory LambdaKropathSection,    // level 1
	localKropathMandatory LambdaKropathSection,     // level 2
	globalLambdaCfgMandatory LambdaConfigSection,   // level 3
	localLambdaCfgMandatory LambdaConfigSection,    // level 4
	// Defaults inputs (highest → lowest priority)
	localLambdaCfgDefaults LambdaConfigSection,     // level 6
	globalLambdaCfgDefaults LambdaConfigSection,    // level 7
	localKropathDefaults LambdaKropathSection,      // level 8
	globalKropathDefaults LambdaKropathSection,     // level 9
) EffectiveLambdaConfig {
	return EffectiveLambdaConfig{
		Mandatory: EffectiveLambdaSection{
			Runtime: firstNonEmptyString(
				globalKropathMandatory.Runtime,      // level 1
				localKropathMandatory.Runtime,       // level 2
				globalLambdaCfgMandatory.Runtime,    // level 3
				localLambdaCfgMandatory.Runtime,     // level 4
			),
			MemorySize: firstNonZeroInt64(
				globalKropathMandatory.MemorySize,   // level 1
				localKropathMandatory.MemorySize,    // level 2
				globalLambdaCfgMandatory.MemorySize, // level 3
				localLambdaCfgMandatory.MemorySize,  // level 4
			),
			Timeout: firstNonZeroInt64(
				globalKropathMandatory.Timeout,      // level 1
				localKropathMandatory.Timeout,       // level 2
				globalLambdaCfgMandatory.Timeout,    // level 3
				localLambdaCfgMandatory.Timeout,     // level 4
			),
			TracingMode: firstNonEmptyString(
				globalKropathMandatory.TracingMode,      // level 1
				localKropathMandatory.TracingMode,       // level 2
				globalLambdaCfgMandatory.TracingMode,    // level 3
				localLambdaCfgMandatory.TracingMode,     // level 4
			),
			KmsKeyArn: firstNonEmptyString(
				globalKropathMandatory.KmsKeyArn,    // level 1
				localKropathMandatory.KmsKeyArn,     // level 2
				globalLambdaCfgMandatory.KmsKeyArn,  // level 3
				localLambdaCfgMandatory.KmsKeyArn,   // level 4
			),
			// NamingTemplate: LambdaConfig levels only (3, 4). KropathConfig has no namingTemplate.
			NamingTemplate: firstNonEmptyString(
				globalLambdaCfgMandatory.NamingTemplate, // level 3
				localLambdaCfgMandatory.NamingTemplate,  // level 4
			),
			// Tags: union of all mandatory sources; L4 added first (lowest priority), L1 wins on key conflict.
			Tags: mergeMaps(
				localLambdaCfgMandatory.Tags,   // level 4
				globalLambdaCfgMandatory.Tags,  // level 3
				localKropathMandatory.Tags,     // level 2
				globalKropathMandatory.Tags,    // level 1 (highest priority)
			),
			// SyncedLabels: additive union from all mandatory sources; L4 added first, L1 wins on key conflict.
			SyncedLabels: mergeMaps(
				localLambdaCfgMandatory.SyncedLabels,   // level 4
				globalLambdaCfgMandatory.SyncedLabels,  // level 3
				localKropathMandatory.SyncedLabels,     // level 2
				globalKropathMandatory.SyncedLabels,    // level 1
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localLambdaCfgMandatory.SyncedAnnotations,   // level 4
				globalLambdaCfgMandatory.SyncedAnnotations,  // level 3
				localKropathMandatory.SyncedAnnotations,     // level 2
				globalKropathMandatory.SyncedAnnotations,    // level 1
			),
		},
		Defaults: EffectiveLambdaSection{
			Runtime: firstNonEmptyString(
				localLambdaCfgDefaults.Runtime,    // level 6
				globalLambdaCfgDefaults.Runtime,   // level 7
				localKropathDefaults.Runtime,      // level 8
				globalKropathDefaults.Runtime,     // level 9
			),
			MemorySize: firstNonZeroInt64(
				localLambdaCfgDefaults.MemorySize,    // level 6
				globalLambdaCfgDefaults.MemorySize,   // level 7
				localKropathDefaults.MemorySize,      // level 8
				globalKropathDefaults.MemorySize,     // level 9
			),
			Timeout: firstNonZeroInt64(
				localLambdaCfgDefaults.Timeout,    // level 6
				globalLambdaCfgDefaults.Timeout,   // level 7
				localKropathDefaults.Timeout,      // level 8
				globalKropathDefaults.Timeout,     // level 9
			),
			TracingMode: firstNonEmptyString(
				localLambdaCfgDefaults.TracingMode,    // level 6
				globalLambdaCfgDefaults.TracingMode,   // level 7
				localKropathDefaults.TracingMode,      // level 8
				globalKropathDefaults.TracingMode,     // level 9
			),
			KmsKeyArn: firstNonEmptyString(
				localLambdaCfgDefaults.KmsKeyArn,    // level 6
				globalLambdaCfgDefaults.KmsKeyArn,   // level 7
				localKropathDefaults.KmsKeyArn,      // level 8
				globalKropathDefaults.KmsKeyArn,     // level 9
			),
			// NamingTemplate: LambdaConfig levels only (6, 7).
			NamingTemplate: firstNonEmptyString(
				localLambdaCfgDefaults.NamingTemplate,  // level 6
				globalLambdaCfgDefaults.NamingTemplate, // level 7
			),
			// Tags: union of all defaults sources; L9 added first (lowest priority), L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,      // level 9
				localKropathDefaults.Tags,       // level 8
				globalLambdaCfgDefaults.Tags,    // level 7
				localLambdaCfgDefaults.Tags,     // level 6 (highest priority)
			),
			// SyncedLabels: additive union from all defaults sources; L9 added first, L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalKropathDefaults.SyncedLabels,      // level 9
				localKropathDefaults.SyncedLabels,       // level 8
				globalLambdaCfgDefaults.SyncedLabels,    // level 7
				localLambdaCfgDefaults.SyncedLabels,     // level 6
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalKropathDefaults.SyncedAnnotations,      // level 9
				localKropathDefaults.SyncedAnnotations,       // level 8
				globalLambdaCfgDefaults.SyncedAnnotations,    // level 7
				localLambdaCfgDefaults.SyncedAnnotations,     // level 6
			),
		},
	}
}
