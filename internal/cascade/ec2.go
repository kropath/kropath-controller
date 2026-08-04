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

// EC2KropathSection holds the EC2-family governance fields promoted from
// KropathConfig.spec.mandatory.ec2 / .defaults.ec2 (ADR-015 §3.5).
//
// Only the three fields consolidatable at KropathConfig scope are here:
// flowLogsRequired, imdsv2Required, ebsEncryptionRequired. Tags,
// SyncedLabels, and SyncedAnnotations are tier-level fields populated
// by the reconciler from KropathConfig.spec.mandatory.tags / .syncedLabels /
// .syncedAnnotations respectively.
//
// Zero value = permissive sentinel (false for all booleans, nil for maps).
type EC2KropathSection struct {
	// FlowLogsRequired forces VPC flow logs enabled on all resources. false = not enforced.
	FlowLogsRequired bool `json:"flowLogsRequired,omitempty"`

	// IMDSv2Required forces IMDSv2 (token-required) on all EC2 instances. false = not enforced.
	IMDSv2Required bool `json:"imdsv2Required,omitempty"`

	// EBSEncryptionRequired forces EBS encryption by default. false = not enforced.
	EBSEncryptionRequired bool `json:"ebsEncryptionRequired,omitempty"`

	// Tags are tier-level cloud resource tags. Populated by reconciler from KropathConfig.mandatory.tags.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are tier-level Kubernetes labels. Populated by reconciler from KropathConfig.mandatory.syncedLabels.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are tier-level Kubernetes annotations. Populated by reconciler from KropathConfig.mandatory.syncedAnnotations.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EC2ConfigSection holds the EC2 governance fields from EC2Config.spec.mandatory
// or EC2Config.spec.defaults (per-type ResourceConfig, ADR-015 §3.5).
//
// Contains all 9 scalar governance fields plus namingTemplate and three map fields.
// Zero value = permissive sentinel: false for booleans, "" for strings, 0 for int64.
type EC2ConfigSection struct {
	// FlowLogsRequired forces VPC flow logs enabled. false = not enforced.
	FlowLogsRequired bool `json:"flowLogsRequired,omitempty"`

	// FlowLogTrafficType specifies the traffic type to log (ALL, ACCEPT, REJECT). "" = not enforced.
	FlowLogTrafficType string `json:"flowLogTrafficType,omitempty"`

	// FlowLogMaxAggregationInterval is the flow log aggregation interval in seconds. 0 = not enforced.
	FlowLogMaxAggregationInterval int64 `json:"flowLogMaxAggregationInterval,omitempty"`

	// RestrictPublicIpOnLaunch prevents assigning public IPs on launch. false = not enforced.
	RestrictPublicIpOnLaunch bool `json:"restrictPublicIpOnLaunch,omitempty"`

	// IMDSv2Required forces IMDSv2 (token-required) on EC2 instances. false = not enforced.
	IMDSv2Required bool `json:"imdsv2Required,omitempty"`

	// EBSEncryptionRequired forces EBS encryption by default. false = not enforced.
	EBSEncryptionRequired bool `json:"ebsEncryptionRequired,omitempty"`

	// EBSDefaultKMSKeyId is the KMS key ARN for default EBS encryption. "" = not enforced.
	EBSDefaultKMSKeyId string `json:"ebsDefaultKmsKeyId,omitempty"`

	// PublicIpRestricted blocks public IP assignment. false = not enforced.
	PublicIpRestricted bool `json:"publicIpRestricted,omitempty"`

	// AllowSourceDestCheckDisable permits disabling source/destination checking. false = not enforced.
	AllowSourceDestCheckDisable bool `json:"allowSourceDestCheckDisable,omitempty"`

	// NamingTemplate enforces a naming pattern. Governed only at EC2Config levels.
	// "" = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// Tags are cloud resource tags for this EC2 config profile. nil / empty = no tags.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created resources.
	// Additive map merge from all sources. nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate.
	// Additive map merge from all sources. nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveEC2Section is one tier (mandatory or defaults) of the merged EC2 governance
// result written into EC2Config.status.effectiveConfig by the controller.
type EffectiveEC2Section struct {
	FlowLogsRequired              bool              `json:"flowLogsRequired,omitempty"`
	FlowLogTrafficType            string            `json:"flowLogTrafficType,omitempty"`
	FlowLogMaxAggregationInterval int64             `json:"flowLogMaxAggregationInterval,omitempty"`
	RestrictPublicIpOnLaunch      bool              `json:"restrictPublicIpOnLaunch,omitempty"`
	IMDSv2Required                bool              `json:"imdsv2Required,omitempty"`
	EBSEncryptionRequired         bool              `json:"ebsEncryptionRequired,omitempty"`
	EBSDefaultKMSKeyId            string            `json:"ebsDefaultKmsKeyId,omitempty"`
	PublicIpRestricted            bool              `json:"publicIpRestricted,omitempty"`
	AllowSourceDestCheckDisable   bool              `json:"allowSourceDestCheckDisable,omitempty"`
	NamingTemplate                string            `json:"namingTemplate,omitempty"`
	Tags                          map[string]string `json:"tags,omitempty"`
	SyncedLabels                  map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations             map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveEC2Config is the merged EC2 governance result written into
// EC2Config.status.effectiveConfig by the controller.
type EffectiveEC2Config struct {
	Mandatory EffectiveEC2Section `json:"mandatory"`
	Defaults  EffectiveEC2Section `json:"defaults"`
}

// MergeEC2Cascade merges EC2 governance fields from all cascade sources and returns
// the effective configuration to be written to EC2Config.status.effectiveConfig.
//
// Nine-level priority chain for EC2 (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.ec2)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.ec2)
//	Level 3 — globalEC2CfgMandatory   (EC2Config in kro-system, mandatory)
//	Level 4 — localEC2CfgMandatory    (EC2Config in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localEC2CfgDefaults     (EC2Config in resource namespace, defaults)
//	Level 7 — globalEC2CfgDefaults    (EC2Config in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.ec2)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.ec2)
//
// Merge rules:
//   - bool fields: firstTrue in priority order (false = not set / permissive, falls through).
//   - string fields: firstNonEmptyString in priority order ("" = not set, falls through).
//   - int64 fields: firstNonZeroInt64 in priority order (0 = not set, falls through).
//   - Tags: additive union merge across all four mandatory levels, all four defaults levels.
//   - SyncedLabels/SyncedAnnotations: additive union from all sources (KropathConfig tier-level included).
//   - NamingTemplate: governed only at EC2Config levels (3-4 mandatory, 6-7 defaults).
//   - FlowLogsRequired, IMDSv2Required, EBSEncryptionRequired: all levels (1-4 mandatory, 6-9 defaults).
//   - EC2Config-only fields (e.g. RestrictPublicIpOnLaunch): EC2Config levels only.
func MergeEC2Cascade(
	globalKropathMandatory EC2KropathSection, // level 1
	localKropathMandatory EC2KropathSection, // level 2
	globalEC2CfgMandatory EC2ConfigSection, // level 3
	localEC2CfgMandatory EC2ConfigSection, // level 4
	localEC2CfgDefaults EC2ConfigSection, // level 6
	globalEC2CfgDefaults EC2ConfigSection, // level 7
	localKropathDefaults EC2KropathSection, // level 8
	globalKropathDefaults EC2KropathSection, // level 9
) EffectiveEC2Config {
	return EffectiveEC2Config{
		Mandatory: EffectiveEC2Section{
			// Three KropathConfig-promoted bool fields: all four mandatory levels.
			FlowLogsRequired: firstTrue(
				globalKropathMandatory.FlowLogsRequired,
				localKropathMandatory.FlowLogsRequired,
				globalEC2CfgMandatory.FlowLogsRequired,
				localEC2CfgMandatory.FlowLogsRequired,
			),
			IMDSv2Required: firstTrue(
				globalKropathMandatory.IMDSv2Required,
				localKropathMandatory.IMDSv2Required,
				globalEC2CfgMandatory.IMDSv2Required,
				localEC2CfgMandatory.IMDSv2Required,
			),
			EBSEncryptionRequired: firstTrue(
				globalKropathMandatory.EBSEncryptionRequired,
				localKropathMandatory.EBSEncryptionRequired,
				globalEC2CfgMandatory.EBSEncryptionRequired,
				localEC2CfgMandatory.EBSEncryptionRequired,
			),
			// EC2Config-only bool fields: governed at EC2Config levels 3-4 only.
			RestrictPublicIpOnLaunch: firstTrue(
				globalEC2CfgMandatory.RestrictPublicIpOnLaunch,
				localEC2CfgMandatory.RestrictPublicIpOnLaunch,
			),
			PublicIpRestricted: firstTrue(
				globalEC2CfgMandatory.PublicIpRestricted,
				localEC2CfgMandatory.PublicIpRestricted,
			),
			AllowSourceDestCheckDisable: firstTrue(
				globalEC2CfgMandatory.AllowSourceDestCheckDisable,
				localEC2CfgMandatory.AllowSourceDestCheckDisable,
			),
			// EC2Config-only string fields.
			FlowLogTrafficType: firstNonEmptyString(
				globalEC2CfgMandatory.FlowLogTrafficType,
				localEC2CfgMandatory.FlowLogTrafficType,
			),
			EBSDefaultKMSKeyId: firstNonEmptyString(
				globalEC2CfgMandatory.EBSDefaultKMSKeyId,
				localEC2CfgMandatory.EBSDefaultKMSKeyId,
			),
			// EC2Config-only int64 field.
			FlowLogMaxAggregationInterval: firstNonZeroInt64(
				globalEC2CfgMandatory.FlowLogMaxAggregationInterval,
				localEC2CfgMandatory.FlowLogMaxAggregationInterval,
			),
			// NamingTemplate: EC2Config levels only (no KropathConfig.ec2.namingTemplate).
			NamingTemplate: firstNonEmptyString(
				globalEC2CfgMandatory.NamingTemplate,
				localEC2CfgMandatory.NamingTemplate,
			),
			// Tags: union of all mandatory sources; L4 added first (lowest priority), L1 wins on key conflict.
			Tags: mergeMaps(
				localEC2CfgMandatory.Tags,
				globalEC2CfgMandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
			// SyncedLabels: union from all mandatory sources (KropathConfig tier-level included per AC-13).
			SyncedLabels: mergeMaps(
				localEC2CfgMandatory.SyncedLabels,
				globalEC2CfgMandatory.SyncedLabels,
				localKropathMandatory.SyncedLabels,
				globalKropathMandatory.SyncedLabels,
			),
			SyncedAnnotations: mergeMaps(
				localEC2CfgMandatory.SyncedAnnotations,
				globalEC2CfgMandatory.SyncedAnnotations,
				localKropathMandatory.SyncedAnnotations,
				globalKropathMandatory.SyncedAnnotations,
			),
		},
		Defaults: EffectiveEC2Section{
			// Three KropathConfig-promoted bool fields: all four defaults levels.
			FlowLogsRequired: firstTrue(
				localEC2CfgDefaults.FlowLogsRequired,
				globalEC2CfgDefaults.FlowLogsRequired,
				localKropathDefaults.FlowLogsRequired,
				globalKropathDefaults.FlowLogsRequired,
			),
			IMDSv2Required: firstTrue(
				localEC2CfgDefaults.IMDSv2Required,
				globalEC2CfgDefaults.IMDSv2Required,
				localKropathDefaults.IMDSv2Required,
				globalKropathDefaults.IMDSv2Required,
			),
			EBSEncryptionRequired: firstTrue(
				localEC2CfgDefaults.EBSEncryptionRequired,
				globalEC2CfgDefaults.EBSEncryptionRequired,
				localKropathDefaults.EBSEncryptionRequired,
				globalKropathDefaults.EBSEncryptionRequired,
			),
			// EC2Config-only bool fields: governed at EC2Config levels 6-7 only.
			RestrictPublicIpOnLaunch: firstTrue(
				localEC2CfgDefaults.RestrictPublicIpOnLaunch,
				globalEC2CfgDefaults.RestrictPublicIpOnLaunch,
			),
			PublicIpRestricted: firstTrue(
				localEC2CfgDefaults.PublicIpRestricted,
				globalEC2CfgDefaults.PublicIpRestricted,
			),
			AllowSourceDestCheckDisable: firstTrue(
				localEC2CfgDefaults.AllowSourceDestCheckDisable,
				globalEC2CfgDefaults.AllowSourceDestCheckDisable,
			),
			// EC2Config-only string fields.
			FlowLogTrafficType: firstNonEmptyString(
				localEC2CfgDefaults.FlowLogTrafficType,
				globalEC2CfgDefaults.FlowLogTrafficType,
			),
			EBSDefaultKMSKeyId: firstNonEmptyString(
				localEC2CfgDefaults.EBSDefaultKMSKeyId,
				globalEC2CfgDefaults.EBSDefaultKMSKeyId,
			),
			// EC2Config-only int64 field.
			FlowLogMaxAggregationInterval: firstNonZeroInt64(
				localEC2CfgDefaults.FlowLogMaxAggregationInterval,
				globalEC2CfgDefaults.FlowLogMaxAggregationInterval,
			),
			// NamingTemplate: EC2Config levels only.
			NamingTemplate: firstNonEmptyString(
				localEC2CfgDefaults.NamingTemplate,
				globalEC2CfgDefaults.NamingTemplate,
			),
			// Tags: union of all defaults sources; L9 added first (lowest priority), L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalEC2CfgDefaults.Tags,
				localEC2CfgDefaults.Tags,
			),
			// SyncedLabels: union from all defaults sources (KropathConfig tier-level included per AC-13).
			SyncedLabels: mergeMaps(
				globalKropathDefaults.SyncedLabels,
				localKropathDefaults.SyncedLabels,
				globalEC2CfgDefaults.SyncedLabels,
				localEC2CfgDefaults.SyncedLabels,
			),
			SyncedAnnotations: mergeMaps(
				globalKropathDefaults.SyncedAnnotations,
				localKropathDefaults.SyncedAnnotations,
				globalEC2CfgDefaults.SyncedAnnotations,
				localEC2CfgDefaults.SyncedAnnotations,
			),
		},
	}
}
