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

// OrganizationsKropathSection holds the Organizations-family governance fields from
// KropathConfig.spec.mandatory.organizations / .defaults.organizations (ADR-015 §3.5).
//
// Three scalar fields are governed at the KropathConfig level:
// iamUserAccessToBilling, roleName, parentID.
// namingTemplate, syncedLabels, and syncedAnnotations are
// OrganizationsConfig-only (family design §4.1).
//
// Zero value of each string field is the permissive sentinel (not enforced).
type OrganizationsKropathSection struct {
	// IamUserAccessToBilling enforces billing access for member accounts.
	// Empty string = not enforced. Valid values when non-empty: "allow" | "deny".
	IamUserAccessToBilling string `json:"iamUserAccessToBilling,omitempty"`

	// RoleName enforces the cross-account IAM role name for member accounts.
	// Empty string = not enforced.
	RoleName string `json:"roleName,omitempty"`

	// ParentID enforces the root or OU ID for OU placement.
	// Empty string = not enforced.
	ParentID string `json:"parentId,omitempty"`

	// Tags are tier-level cloud resource tags.
	// The reconciler populates this from KropathConfig.spec.mandatory.tags /
	// .defaults.tags so that tag union merge flows through
	// MergeOrganizationsCascade alongside the Organizations-specific fields.
	// nil / empty map = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// OrganizationsConfigSection holds the Organizations governance fields from
// OrganizationsConfig.spec.mandatory or OrganizationsConfig.spec.defaults
// (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each string field is the permissive sentinel (not enforced).
type OrganizationsConfigSection struct {
	// IamUserAccessToBilling enforces billing access for member accounts.
	// Empty string = not enforced.
	IamUserAccessToBilling string `json:"iamUserAccessToBilling,omitempty"`

	// RoleName enforces the cross-account IAM role name.
	// Empty string = not enforced.
	RoleName string `json:"roleName,omitempty"`

	// ParentID enforces the root or OU ID for OU placement.
	// Empty string = not enforced.
	ParentID string `json:"parentId,omitempty"`

	// NamingTemplate is the resource naming template (e.g. "{namespace}-{name}").
	// Governed only at OrganizationsConfig levels 3-4 (mandatory) and 6-7 (defaults).
	// KropathConfig.organizations does NOT carry namingTemplate.
	// Empty string = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created Organizations resources.
	// Additive map merge across OrganizationsConfig tiers only.
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to created Organizations resources.
	// Additive map merge across OrganizationsConfig tiers only.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`

	// Tags are cloud resource tags for this Organizations config profile.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// EffectiveOrganizationsSection is one tier (mandatory or defaults) of the merged
// Organizations governance result written into OrganizationsConfig.status.effectiveConfig
// by the controller.
type EffectiveOrganizationsSection struct {
	IamUserAccessToBilling string            `json:"iamUserAccessToBilling,omitempty"`
	RoleName               string            `json:"roleName,omitempty"`
	ParentID               string            `json:"parentId,omitempty"`
	NamingTemplate         string            `json:"namingTemplate,omitempty"`
	SyncedLabels           map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations      map[string]string `json:"syncedAnnotations,omitempty"`
	Tags                   map[string]string `json:"tags,omitempty"`
}

// EffectiveOrganizationsConfig is the merged Organizations governance result written into
// OrganizationsConfig.status.effectiveConfig by the controller.
type EffectiveOrganizationsConfig struct {
	Mandatory EffectiveOrganizationsSection `json:"mandatory"`
	Defaults  EffectiveOrganizationsSection `json:"defaults"`
}

// MergeOrganizationsCascade merges Organizations governance fields from all cascade
// sources and returns the effective configuration to be written to status.effectiveConfig.
//
// Nine-level priority chain for Organizations (ADR-015 §5.3); this controller handles 8 of the 9
// active levels — level 5 (instance spec) is resolved in RGD CEL, not here:
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.organizations)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.organizations)
//	Level 3 — globalOrgCfgMandatory   (OrganizationsConfig in kro-system, mandatory)
//	Level 4 — localOrgCfgMandatory    (OrganizationsConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localOrgCfgDefaults     (OrganizationsConfig in resource namespace, defaults)
//	Level 7 — globalOrgCfgDefaults    (OrganizationsConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.organizations)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.organizations)
//
// Scalar merge: firstNonEmptyString in priority order (lowest number wins).
// Tags: additive union merge across all four mandatory levels, all four defaults levels.
// SyncedLabels/SyncedAnnotations: additive union from OrganizationsConfig levels only
// (no KropathConfig).
// NamingTemplate: governed only at OrganizationsConfig levels (3-4 mandatory, 6-7 defaults).
func MergeOrganizationsCascade(
	globalKropathMandatory OrganizationsKropathSection, // level 1
	localKropathMandatory OrganizationsKropathSection, // level 2
	globalOrgCfgMandatory OrganizationsConfigSection, // level 3
	localOrgCfgMandatory OrganizationsConfigSection, // level 4
	localOrgCfgDefaults OrganizationsConfigSection, // level 6
	globalOrgCfgDefaults OrganizationsConfigSection, // level 7
	localKropathDefaults OrganizationsKropathSection, // level 8
	globalKropathDefaults OrganizationsKropathSection, // level 9
) EffectiveOrganizationsConfig {
	return EffectiveOrganizationsConfig{
		Mandatory: EffectiveOrganizationsSection{
			IamUserAccessToBilling: firstNonEmptyString(
				globalKropathMandatory.IamUserAccessToBilling,
				localKropathMandatory.IamUserAccessToBilling,
				globalOrgCfgMandatory.IamUserAccessToBilling,
				localOrgCfgMandatory.IamUserAccessToBilling,
			),
			RoleName: firstNonEmptyString(
				globalKropathMandatory.RoleName,
				localKropathMandatory.RoleName,
				globalOrgCfgMandatory.RoleName,
				localOrgCfgMandatory.RoleName,
			),
			ParentID: firstNonEmptyString(
				globalKropathMandatory.ParentID,
				localKropathMandatory.ParentID,
				globalOrgCfgMandatory.ParentID,
				localOrgCfgMandatory.ParentID,
			),
			// NamingTemplate: OrganizationsConfig levels only (3, 4);
			// KropathConfig has no namingTemplate field for organizations.
			NamingTemplate: firstNonEmptyString(
				globalOrgCfgMandatory.NamingTemplate,
				localOrgCfgMandatory.NamingTemplate,
			),
			// SyncedLabels: additive union from OrganizationsConfig levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localOrgCfgMandatory.SyncedLabels,
				globalOrgCfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localOrgCfgMandatory.SyncedAnnotations,
				globalOrgCfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localOrgCfgMandatory.Tags,
				globalOrgCfgMandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
		},
		Defaults: EffectiveOrganizationsSection{
			IamUserAccessToBilling: firstNonEmptyString(
				localOrgCfgDefaults.IamUserAccessToBilling,
				globalOrgCfgDefaults.IamUserAccessToBilling,
				localKropathDefaults.IamUserAccessToBilling,
				globalKropathDefaults.IamUserAccessToBilling,
			),
			RoleName: firstNonEmptyString(
				localOrgCfgDefaults.RoleName,
				globalOrgCfgDefaults.RoleName,
				localKropathDefaults.RoleName,
				globalKropathDefaults.RoleName,
			),
			ParentID: firstNonEmptyString(
				localOrgCfgDefaults.ParentID,
				globalOrgCfgDefaults.ParentID,
				localKropathDefaults.ParentID,
				globalKropathDefaults.ParentID,
			),
			// NamingTemplate: OrganizationsConfig levels only (6, 7).
			NamingTemplate: firstNonEmptyString(
				localOrgCfgDefaults.NamingTemplate,
				globalOrgCfgDefaults.NamingTemplate,
			),
			// SyncedLabels: additive union from OrganizationsConfig levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalOrgCfgDefaults.SyncedLabels,
				localOrgCfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalOrgCfgDefaults.SyncedAnnotations,
				localOrgCfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalOrgCfgDefaults.Tags,
				localOrgCfgDefaults.Tags,
			),
		},
	}
}
