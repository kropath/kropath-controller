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

// SSMKropathSection holds the SSM-family governance fields from
// KropathConfig.spec.mandatory.ssm / .defaults.ssm (ADR-015 §3.5).
//
// Only 4 org-wide governance fields are present here. The remaining SSM
// fields (documentType, operatingSystem, approvedPatchesComplianceLevel,
// rejectedPatchesAction, approvedPatchesEnableNonSecurity, namingTemplate,
// syncedLabels, syncedAnnotations) live in SSMConfig only. The Tags field
// is augmented from the KropathConfig tier-level tags before passing to
// MergeSSMCascade.
type SSMKropathSection struct {
	// ParameterType forces the SSM parameter type (String, StringList, SecureString).
	// Maps to SSMConfig.type. Empty = not enforced.
	ParameterType string `json:"parameterType,omitempty"`

	// ParameterTier forces the SSM parameter tier (Standard, Advanced, Intelligent-Tiering).
	// Maps to SSMConfig.tier. Empty = not enforced.
	ParameterTier string `json:"parameterTier,omitempty"`

	// KeyID is the KMS key ID or ARN for SecureString encryption. Empty = not enforced.
	KeyID string `json:"keyID,omitempty"`

	// AllowedDocumentTypes restricts the allowed SSM document types to an org-wide allowlist.
	// Empty slice = no restriction.
	AllowedDocumentTypes []string `json:"allowedDocumentTypes,omitempty"`

	// Tags are tier-level cloud resource tags. Augmented from KropathConfig.spec.mandatory.tags
	// before calling MergeSSMCascade. nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// SSMConfigSection holds the SSM governance fields from SSMConfig.spec.mandatory
// or SSMConfig.spec.defaults (per-type ResourceConfig, ADR-015 §3.5).
//
// Contains all 13 governance field groups defined in aws-ssm-01-ssmconfig.md.
type SSMConfigSection struct {
	// Type forces the SSM parameter type (String, StringList, SecureString). Empty = not enforced.
	Type string `json:"type,omitempty"`

	// Tier forces the SSM parameter tier (Standard, Advanced, Intelligent-Tiering). Empty = not enforced.
	Tier string `json:"tier,omitempty"`

	// KeyID is the KMS key ID or ARN for SecureString encryption. Empty = not enforced.
	KeyID string `json:"keyID,omitempty"`

	// DocumentType is the SSM document type for Patch Manager. Empty = not enforced.
	DocumentType string `json:"documentType,omitempty"`

	// AllowedDocumentTypes restricts the allowed SSM document types to an allowlist.
	// Empty slice = no restriction.
	AllowedDocumentTypes []string `json:"allowedDocumentTypes,omitempty"`

	// OperatingSystem is the target OS for Patch Manager (e.g. WINDOWS, AMAZON_LINUX_2).
	// Empty = not enforced.
	OperatingSystem string `json:"operatingSystem,omitempty"`

	// ApprovedPatchesComplianceLevel is the compliance level for approved patches
	// (CRITICAL, HIGH, MEDIUM, LOW, INFORMATIONAL, UNSPECIFIED). Empty = not enforced.
	ApprovedPatchesComplianceLevel string `json:"approvedPatchesComplianceLevel,omitempty"`

	// RejectedPatchesAction is the action for rejected patches
	// (ALLOW_AS_DEPENDENCY, BLOCK). Empty = not enforced.
	RejectedPatchesAction string `json:"rejectedPatchesAction,omitempty"`

	// ApprovedPatchesEnableNonSecurity controls whether non-security patches are approved.
	// nil = not enforced; true = enabled; false = explicitly disabled.
	ApprovedPatchesEnableNonSecurity *bool `json:"approvedPatchesEnableNonSecurity,omitempty"`

	// NamingTemplate enforces a naming pattern. Empty = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// Tags are cloud resource tags for this SSM config profile. nil / empty = no tags.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created resources. Additive map merge.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate. Additive map merge.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveSSMSection is one tier (mandatory or defaults) of the merged SSM
// governance result written into SSMConfig.status.effectiveConfig by the controller.
type EffectiveSSMSection struct {
	Type                             string            `json:"type,omitempty"`
	Tier                             string            `json:"tier,omitempty"`
	KeyID                            string            `json:"keyID,omitempty"`
	DocumentType                     string            `json:"documentType,omitempty"`
	AllowedDocumentTypes             []string          `json:"allowedDocumentTypes,omitempty"`
	OperatingSystem                  string            `json:"operatingSystem,omitempty"`
	ApprovedPatchesComplianceLevel   string            `json:"approvedPatchesComplianceLevel,omitempty"`
	RejectedPatchesAction            string            `json:"rejectedPatchesAction,omitempty"`
	ApprovedPatchesEnableNonSecurity *bool             `json:"approvedPatchesEnableNonSecurity,omitempty"`
	NamingTemplate                   string            `json:"namingTemplate,omitempty"`
	Tags                             map[string]string `json:"tags,omitempty"`
	SyncedLabels                     map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations                map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveSSMConfig is the merged SSM governance result written into
// SSMConfig.status.effectiveConfig by the controller.
type EffectiveSSMConfig struct {
	Mandatory EffectiveSSMSection `json:"mandatory"`
	Defaults  EffectiveSSMSection `json:"defaults"`
}

// MergeSSMCascade merges SSM governance fields from all cascade sources and returns
// the effective configuration to be written to SSMConfig.status.effectiveConfig.
//
// Ten-level priority chain for SSM (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.ssm)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.ssm)
//	Level 3 — globalSSMCfgMandatory   (SSMConfig in kro-system, mandatory)
//	Level 4 — localSSMCfgMandatory    (SSMConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localSSMCfgDefaults     (SSMConfig in resource namespace, defaults)
//	Level 7 — globalSSMCfgDefaults    (SSMConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.ssm)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.ssm)
//
// Merge rules:
//   - *bool pointer fields: firstNonNilBoolPtr in priority order (nil = not set, falls through).
//   - string fields: firstNonEmptyString in priority order ("" = not set, falls through).
//   - []string fields: firstNonEmptyStrings in priority order (empty slice = not set, falls through).
//   - Tags: additive union merge across all four mandatory levels, all four defaults levels.
//   - SyncedLabels/SyncedAnnotations: additive union from SSMConfig levels only (no KropathConfig).
//   - DocumentType, OperatingSystem, ApprovedPatchesComplianceLevel, RejectedPatchesAction,
//     ApprovedPatchesEnableNonSecurity, NamingTemplate: SSMConfig levels only (3-4, 6-7).
func MergeSSMCascade(
	globalKropathMandatory SSMKropathSection, // level 1
	localKropathMandatory SSMKropathSection,  // level 2
	globalSSMCfgMandatory SSMConfigSection,   // level 3
	localSSMCfgMandatory SSMConfigSection,    // level 4
	localSSMCfgDefaults SSMConfigSection,     // level 6
	globalSSMCfgDefaults SSMConfigSection,    // level 7
	localKropathDefaults SSMKropathSection,   // level 8
	globalKropathDefaults SSMKropathSection,  // level 9
) EffectiveSSMConfig {
	return EffectiveSSMConfig{
		Mandatory: EffectiveSSMSection{
			// Type (parameterType at KropathConfig): levels 1-4.
			Type: firstNonEmptyString(
				globalKropathMandatory.ParameterType,
				localKropathMandatory.ParameterType,
				globalSSMCfgMandatory.Type,
				localSSMCfgMandatory.Type,
			),
			// Tier (parameterTier at KropathConfig): levels 1-4.
			Tier: firstNonEmptyString(
				globalKropathMandatory.ParameterTier,
				localKropathMandatory.ParameterTier,
				globalSSMCfgMandatory.Tier,
				localSSMCfgMandatory.Tier,
			),
			// KeyID: KropathConfig levels 1-2 participate.
			KeyID: firstNonEmptyString(
				globalKropathMandatory.KeyID,
				localKropathMandatory.KeyID,
				globalSSMCfgMandatory.KeyID,
				localSSMCfgMandatory.KeyID,
			),
			// DocumentType: SSMConfig levels only (3, 4);
			// KropathConfig has no documentType field for ssm.
			DocumentType: firstNonEmptyString(
				globalSSMCfgMandatory.DocumentType,
				localSSMCfgMandatory.DocumentType,
			),
			// AllowedDocumentTypes: KropathConfig levels 1-2 participate.
			AllowedDocumentTypes: firstNonEmptyStrings(
				globalKropathMandatory.AllowedDocumentTypes,
				localKropathMandatory.AllowedDocumentTypes,
				globalSSMCfgMandatory.AllowedDocumentTypes,
				localSSMCfgMandatory.AllowedDocumentTypes,
			),
			// OperatingSystem: SSMConfig levels only (3, 4).
			OperatingSystem: firstNonEmptyString(
				globalSSMCfgMandatory.OperatingSystem,
				localSSMCfgMandatory.OperatingSystem,
			),
			// ApprovedPatchesComplianceLevel: SSMConfig levels only (3, 4).
			ApprovedPatchesComplianceLevel: firstNonEmptyString(
				globalSSMCfgMandatory.ApprovedPatchesComplianceLevel,
				localSSMCfgMandatory.ApprovedPatchesComplianceLevel,
			),
			// RejectedPatchesAction: SSMConfig levels only (3, 4).
			RejectedPatchesAction: firstNonEmptyString(
				globalSSMCfgMandatory.RejectedPatchesAction,
				localSSMCfgMandatory.RejectedPatchesAction,
			),
			// ApprovedPatchesEnableNonSecurity: SSMConfig levels only (3, 4).
			// nil = not enforced; explicit false propagates correctly.
			ApprovedPatchesEnableNonSecurity: firstNonNilBoolPtr(
				globalSSMCfgMandatory.ApprovedPatchesEnableNonSecurity,
				localSSMCfgMandatory.ApprovedPatchesEnableNonSecurity,
			),
			// NamingTemplate: SSMConfig levels only (3, 4).
			NamingTemplate: firstNonEmptyString(
				globalSSMCfgMandatory.NamingTemplate,
				localSSMCfgMandatory.NamingTemplate,
			),
			// SyncedLabels: additive union from SSMConfig levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localSSMCfgMandatory.SyncedLabels,
				globalSSMCfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localSSMCfgMandatory.SyncedAnnotations,
				globalSSMCfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localSSMCfgMandatory.Tags,
				globalSSMCfgMandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
		},
		Defaults: EffectiveSSMSection{
			// Type (parameterType at KropathConfig): levels 6-9.
			Type: firstNonEmptyString(
				localSSMCfgDefaults.Type,
				globalSSMCfgDefaults.Type,
				localKropathDefaults.ParameterType,
				globalKropathDefaults.ParameterType,
			),
			// Tier (parameterTier at KropathConfig): levels 6-9.
			Tier: firstNonEmptyString(
				localSSMCfgDefaults.Tier,
				globalSSMCfgDefaults.Tier,
				localKropathDefaults.ParameterTier,
				globalKropathDefaults.ParameterTier,
			),
			// KeyID: KropathConfig levels 8-9 participate.
			KeyID: firstNonEmptyString(
				localSSMCfgDefaults.KeyID,
				globalSSMCfgDefaults.KeyID,
				localKropathDefaults.KeyID,
				globalKropathDefaults.KeyID,
			),
			// DocumentType: SSMConfig levels only (6, 7).
			DocumentType: firstNonEmptyString(
				localSSMCfgDefaults.DocumentType,
				globalSSMCfgDefaults.DocumentType,
			),
			// AllowedDocumentTypes: KropathConfig levels 8-9 participate.
			AllowedDocumentTypes: firstNonEmptyStrings(
				localSSMCfgDefaults.AllowedDocumentTypes,
				globalSSMCfgDefaults.AllowedDocumentTypes,
				localKropathDefaults.AllowedDocumentTypes,
				globalKropathDefaults.AllowedDocumentTypes,
			),
			// OperatingSystem: SSMConfig levels only (6, 7).
			OperatingSystem: firstNonEmptyString(
				localSSMCfgDefaults.OperatingSystem,
				globalSSMCfgDefaults.OperatingSystem,
			),
			// ApprovedPatchesComplianceLevel: SSMConfig levels only (6, 7).
			ApprovedPatchesComplianceLevel: firstNonEmptyString(
				localSSMCfgDefaults.ApprovedPatchesComplianceLevel,
				globalSSMCfgDefaults.ApprovedPatchesComplianceLevel,
			),
			// RejectedPatchesAction: SSMConfig levels only (6, 7).
			RejectedPatchesAction: firstNonEmptyString(
				localSSMCfgDefaults.RejectedPatchesAction,
				globalSSMCfgDefaults.RejectedPatchesAction,
			),
			// ApprovedPatchesEnableNonSecurity: SSMConfig levels only (6, 7).
			ApprovedPatchesEnableNonSecurity: firstNonNilBoolPtr(
				localSSMCfgDefaults.ApprovedPatchesEnableNonSecurity,
				globalSSMCfgDefaults.ApprovedPatchesEnableNonSecurity,
			),
			// NamingTemplate: SSMConfig levels only (6, 7).
			NamingTemplate: firstNonEmptyString(
				localSSMCfgDefaults.NamingTemplate,
				globalSSMCfgDefaults.NamingTemplate,
			),
			// SyncedLabels: additive union from SSMConfig levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalSSMCfgDefaults.SyncedLabels,
				localSSMCfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalSSMCfgDefaults.SyncedAnnotations,
				localSSMCfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalSSMCfgDefaults.Tags,
				localSSMCfgDefaults.Tags,
			),
		},
	}
}

// ValidateSSMDocumentType checks the documentType/allowedDocumentTypes cross-field
// constraint on the resolved mandatory tier.
//
// When mandatory.documentType is non-empty AND mandatory.allowedDocumentTypes is non-empty,
// documentType must be a member of allowedDocumentTypes.
//
// Returns (valid=true, "") when the constraint is satisfied or does not apply.
// Returns (valid=false, message) when documentType is not in allowedDocumentTypes.
func ValidateSSMDocumentType(mandatory EffectiveSSMSection) (bool, string) {
	if mandatory.DocumentType == "" || len(mandatory.AllowedDocumentTypes) == 0 {
		return true, ""
	}
	for _, allowed := range mandatory.AllowedDocumentTypes {
		if mandatory.DocumentType == allowed {
			return true, ""
		}
	}
	return false, fmt.Sprintf(
		"documentType %q is not in allowedDocumentTypes %v",
		mandatory.DocumentType,
		mandatory.AllowedDocumentTypes,
	)
}
