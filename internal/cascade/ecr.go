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

import "fmt"

// ECRKropathSection holds the ECR governance fields from
// KropathConfig.spec.mandatory.ecr / .defaults.ecr (ADR-015 §3.5).
//
// Fields NOT in KropathConfig: kmsKeyID (per-profile), namingTemplate (v1 exclusion),
// policy, scanOnPush, registryID, imageTagMutabilityExclusionFilters.
//
// Zero value of each field is the permissive sentinel (not enforced).
type ECRKropathSection struct {
	// ImageTagMutability is the org-wide image tag mutability mandate.
	// Empty string = not enforced; "IMMUTABLE" = org-wide mandate.
	ImageTagMutability string `json:"imageTagMutability,omitempty"`

	// EncryptionType is the org-wide encryption type mandate.
	// Empty string = not enforced; "KMS" = org-wide KMS mandate.
	EncryptionType string `json:"encryptionType,omitempty"`

	// LifecyclePolicy is the org-wide lifecycle policy JSON mandate.
	// Empty string = not enforced.
	LifecyclePolicy string `json:"lifecyclePolicy,omitempty"`

	// Tags are tier-level cloud resource tags from KropathConfig.spec.mandatory.tags
	// or .defaults.tags. Populated by the reconciler from the tier-level field so that
	// tag cascade flows through MergeECRCascade alongside the ECR-specific fields.
	// nil / empty map = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// ECRConfigSection holds the ECR governance fields from
// ECRConfig.spec.mandatory or ECRConfig.spec.defaults (ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type ECRConfigSection struct {
	// ImageTagMutability is the image tag mutability setting.
	// Empty = not enforced (mandatory) or "MUTABLE" (defaults).
	ImageTagMutability string `json:"imageTagMutability,omitempty"`

	// EncryptionType is the encryption type.
	// Empty = not enforced (mandatory) or "AES256" (defaults).
	EncryptionType string `json:"encryptionType,omitempty"`

	// KmsKeyID is the KMS key ARN or alias.
	// Empty = not enforced (mandatory) or default AES256 encryption (defaults).
	// Not present in KropathConfig.ecr (per-profile field).
	KmsKeyID string `json:"kmsKeyID,omitempty"`

	// LifecyclePolicy is the lifecycle policy JSON string.
	// Empty = not enforced (mandatory) or no lifecycle policy (defaults).
	LifecyclePolicy string `json:"lifecyclePolicy,omitempty"`

	// NamingTemplate is the ECR repository naming template.
	// Governed only at ECRConfig levels 3-4 (mandatory) and 6-7 (defaults).
	// Not present in KropathConfig.ecr (v1 exclusion).
	// Empty = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to ECR resource CRs.
	// Additive map merge across ECRConfig tiers only.
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to ECR resource CRs.
	// Additive map merge across ECRConfig tiers only.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`

	// Tags are cloud resource tags for this ECRConfig profile.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// EffectiveECRSection is one tier (mandatory or defaults) of the merged ECR
// governance result written into ECRConfig.status.effectiveConfig by the controller.
type EffectiveECRSection struct {
	ImageTagMutability string            `json:"imageTagMutability,omitempty"`
	EncryptionType     string            `json:"encryptionType,omitempty"`
	KmsKeyID           string            `json:"kmsKeyID,omitempty"`
	LifecyclePolicy    string            `json:"lifecyclePolicy,omitempty"`
	NamingTemplate     string            `json:"namingTemplate,omitempty"`
	SyncedLabels       map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations  map[string]string `json:"syncedAnnotations,omitempty"`
	Tags               map[string]string `json:"tags,omitempty"`
}

// EffectiveECRConfig is the merged ECR governance result written into
// ECRConfig.status.effectiveConfig by the controller.
type EffectiveECRConfig struct {
	Mandatory EffectiveECRSection `json:"mandatory"`
	Defaults  EffectiveECRSection `json:"defaults"`
}

// MergeECRCascade merges ECR governance fields from all cascade sources and
// returns the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for ECR (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.ecr)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.ecr)
//	Level 3 — globalECRCfgMandatory   (ECRConfig in kro-system, mandatory)
//	Level 4 — localECRCfgMandatory    (ECRConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localECRCfgDefaults     (ECRConfig in resource namespace, defaults)
//	Level 7 — globalECRCfgDefaults    (ECRConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.ecr)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.ecr)
//
// Scalar merge (all string fields): firstNonEmptyString in priority order.
// imageTagMutability: levels 1-4 mandatory (KropathConfig, ECRConfig), 6-9 defaults.
// encryptionType:     levels 1-4 mandatory, 6-9 defaults.
// kmsKeyID:           levels 3-4 mandatory, 6-7 defaults (not in KropathConfig.ecr).
// lifecyclePolicy:    levels 1-4 mandatory, 6-9 defaults.
// namingTemplate:     levels 3-4 mandatory, 6-7 defaults (not in KropathConfig.ecr).
// Tags: additive union merge across all four mandatory/defaults levels; lower level wins on key conflict.
// SyncedLabels/SyncedAnnotations: additive union from ECRConfig levels only (not KropathConfig.ecr).
//
// Note: callers MUST call ValidateECREncryption on the returned mandatory section
// before writing status.effectiveConfig. If validation fails, effectiveConfig must not be written.
func MergeECRCascade(
	globalKropathMandatory ECRKropathSection, // level 1
	localKropathMandatory ECRKropathSection, // level 2
	globalECRCfgMandatory ECRConfigSection, // level 3
	localECRCfgMandatory ECRConfigSection, // level 4
	localECRCfgDefaults ECRConfigSection, // level 6
	globalECRCfgDefaults ECRConfigSection, // level 7
	localKropathDefaults ECRKropathSection, // level 8
	globalKropathDefaults ECRKropathSection, // level 9
) EffectiveECRConfig {
	return EffectiveECRConfig{
		Mandatory: EffectiveECRSection{
			// imageTagMutability: levels 1, 2, 3, 4 (KropathConfig wins over ECRConfig).
			ImageTagMutability: firstNonEmptyString(
				globalKropathMandatory.ImageTagMutability, // level 1
				localKropathMandatory.ImageTagMutability,  // level 2
				globalECRCfgMandatory.ImageTagMutability,  // level 3
				localECRCfgMandatory.ImageTagMutability,   // level 4
			),
			// encryptionType: levels 1, 2, 3, 4 (KropathConfig wins over ECRConfig).
			EncryptionType: firstNonEmptyString(
				globalKropathMandatory.EncryptionType, // level 1
				localKropathMandatory.EncryptionType,  // level 2
				globalECRCfgMandatory.EncryptionType,  // level 3
				localECRCfgMandatory.EncryptionType,   // level 4
			),
			// kmsKeyID: levels 3, 4 only (not in KropathConfig.ecr).
			KmsKeyID: firstNonEmptyString(
				globalECRCfgMandatory.KmsKeyID, // level 3
				localECRCfgMandatory.KmsKeyID,  // level 4
			),
			// lifecyclePolicy: levels 1, 2, 3, 4 (KropathConfig wins over ECRConfig).
			LifecyclePolicy: firstNonEmptyString(
				globalKropathMandatory.LifecyclePolicy, // level 1
				localKropathMandatory.LifecyclePolicy,  // level 2
				globalECRCfgMandatory.LifecyclePolicy,  // level 3
				localECRCfgMandatory.LifecyclePolicy,   // level 4
			),
			// namingTemplate: levels 3, 4 only (not in KropathConfig.ecr v1).
			NamingTemplate: firstNonEmptyString(
				globalECRCfgMandatory.NamingTemplate, // level 3
				localECRCfgMandatory.NamingTemplate,  // level 4
			),
			// SyncedLabels: additive union from ECRConfig levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localECRCfgMandatory.SyncedLabels,
				globalECRCfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localECRCfgMandatory.SyncedAnnotations,
				globalECRCfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localECRCfgMandatory.Tags,    // level 4 (lowest priority)
				globalECRCfgMandatory.Tags,   // level 3
				localKropathMandatory.Tags,   // level 2
				globalKropathMandatory.Tags,  // level 1 (highest priority)
			),
		},
		Defaults: EffectiveECRSection{
			// imageTagMutability: levels 6, 7, 8, 9.
			ImageTagMutability: firstNonEmptyString(
				localECRCfgDefaults.ImageTagMutability,    // level 6
				globalECRCfgDefaults.ImageTagMutability,   // level 7
				localKropathDefaults.ImageTagMutability,   // level 8
				globalKropathDefaults.ImageTagMutability,  // level 9
			),
			// encryptionType: levels 6, 7, 8, 9.
			EncryptionType: firstNonEmptyString(
				localECRCfgDefaults.EncryptionType,    // level 6
				globalECRCfgDefaults.EncryptionType,   // level 7
				localKropathDefaults.EncryptionType,   // level 8
				globalKropathDefaults.EncryptionType,  // level 9
			),
			// kmsKeyID: levels 6, 7 only (not in KropathConfig.ecr).
			KmsKeyID: firstNonEmptyString(
				localECRCfgDefaults.KmsKeyID,   // level 6
				globalECRCfgDefaults.KmsKeyID,  // level 7
			),
			// lifecyclePolicy: levels 6, 7, 8, 9.
			LifecyclePolicy: firstNonEmptyString(
				localECRCfgDefaults.LifecyclePolicy,    // level 6
				globalECRCfgDefaults.LifecyclePolicy,   // level 7
				localKropathDefaults.LifecyclePolicy,   // level 8
				globalKropathDefaults.LifecyclePolicy,  // level 9
			),
			// namingTemplate: levels 6, 7 only (not in KropathConfig.ecr v1).
			NamingTemplate: firstNonEmptyString(
				localECRCfgDefaults.NamingTemplate,   // level 6
				globalECRCfgDefaults.NamingTemplate,  // level 7
			),
			// SyncedLabels: additive union from ECRConfig levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalECRCfgDefaults.SyncedLabels,
				localECRCfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalECRCfgDefaults.SyncedAnnotations,
				localECRCfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,  // level 9 (lowest priority)
				localKropathDefaults.Tags,   // level 8
				globalECRCfgDefaults.Tags,   // level 7
				localECRCfgDefaults.Tags,    // level 6 (highest priority)
			),
		},
	}
}

// ValidateECREncryption checks the cross-field constraint: encryptionType=AES256 AND
// a non-empty kmsKeyID in the same mandatory tier is an invalid configuration.
//
// KMS key management requires encryptionType=KMS. Setting encryptionType=AES256 while
// also mandating a kmsKeyID is contradictory and must be rejected at reconcile time.
//
// Returns (true, "", "") when the constraint does not apply or passes.
// Returns (false, "InvalidEncryptionConfiguration", <message>) on failure.
// On failure the caller MUST NOT write status.effectiveConfig.
func ValidateECREncryption(mandatory EffectiveECRSection) (valid bool, reason, message string) {
	if mandatory.EncryptionType == "AES256" && mandatory.KmsKeyID != "" {
		return false, "InvalidEncryptionConfiguration", fmt.Sprintf(
			"mandatory.encryptionType is %q but mandatory.kmsKeyID is also set (%q); "+
				"KMS key management requires encryptionType=KMS",
			mandatory.EncryptionType, mandatory.KmsKeyID,
		)
	}
	return true, "", ""
}
