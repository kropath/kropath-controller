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

// NetworkFirewallKropathSection holds the Network Firewall-family governance fields from
// KropathConfig.spec.mandatory.networkfirewall / .defaults.networkfirewall (ADR-015 §3.5).
//
// 7 scalar fields are governed at the KropathConfig level.
// namingTemplate, syncedLabels, and syncedAnnotations are NetworkFirewallConfig-only.
// loggingDestination fields are NetworkFirewallConfig-only (no KropathConfig counterpart).
//
// Zero value of each field is the permissive sentinel (not enforced).
type NetworkFirewallKropathSection struct {
	// DeleteProtection enforces deletion protection org-wide.
	// false = not enforced; true = all firewalls must enable deletion protection.
	// Sentinel: false (firstTrue semantics — only true participates in cascade).
	DeleteProtection bool `json:"deleteProtection,omitempty"`

	// FirewallPolicyChangeProtection enforces policy change protection org-wide.
	// false = not enforced; true = all firewalls must enable policy change protection.
	FirewallPolicyChangeProtection bool `json:"firewallPolicyChangeProtection,omitempty"`

	// SubnetChangeProtection enforces subnet change protection org-wide.
	// false = not enforced; true = all firewalls must enable subnet change protection.
	SubnetChangeProtection bool `json:"subnetChangeProtection,omitempty"`

	// EncryptionType is the org-wide encryption enforcement.
	// Empty string = not enforced; CUSTOMER_KMS forces CMK encryption.
	EncryptionType string `json:"encryptionType,omitempty"`

	// StatefulRuleOrder is the org-wide stateful rule evaluation order enforcement.
	// Empty string = not enforced; STRICT_ORDER or DEFAULT_ACTION_ORDER.
	StatefulRuleOrder string `json:"statefulRuleOrder,omitempty"`

	// StatefulDefaultActions is the org-wide enforced set of default stateful actions.
	// nil / empty = not enforced; first-non-empty-wins in cascade.
	StatefulDefaultActions []string `json:"statefulDefaultActions,omitempty"`

	// StreamExceptionPolicy is the org-wide stream exception policy enforcement.
	// Empty string = not enforced; DROP or CONTINUE.
	StreamExceptionPolicy string `json:"streamExceptionPolicy,omitempty"`

	// Tags are tier-level cloud resource tags from KropathConfig.spec.mandatory.tags.
	// Populated by the reconciler from the tier-level field so that tag union merge flows
	// through MergeNetworkFirewallCascade alongside the family-specific fields.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// NetworkFirewallConfigSection holds the Network Firewall governance fields from
// NetworkFirewallConfig.spec.mandatory or NetworkFirewallConfig.spec.defaults
// (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type NetworkFirewallConfigSection struct {
	// DeleteProtection enforces deletion protection for this profile.
	// false = not enforced; true = forces deletion protection.
	DeleteProtection bool `json:"deleteProtection,omitempty"`

	// FirewallPolicyChangeProtection enforces policy change protection for this profile.
	FirewallPolicyChangeProtection bool `json:"firewallPolicyChangeProtection,omitempty"`

	// SubnetChangeProtection enforces subnet change protection for this profile.
	SubnetChangeProtection bool `json:"subnetChangeProtection,omitempty"`

	// EncryptionType is the encryption enforcement for this profile.
	// Empty string = not enforced.
	EncryptionType string `json:"encryptionType,omitempty"`

	// StatefulRuleOrder is the stateful rule evaluation order for this profile.
	// Empty string = not enforced.
	StatefulRuleOrder string `json:"statefulRuleOrder,omitempty"`

	// StatefulDefaultActions is the enforced set of default stateful actions for this profile.
	// nil / empty = not enforced.
	StatefulDefaultActions []string `json:"statefulDefaultActions,omitempty"`

	// StreamExceptionPolicy is the stream exception policy for this profile.
	// Empty string = not enforced.
	StreamExceptionPolicy string `json:"streamExceptionPolicy,omitempty"`

	// NamingTemplate is the firewall naming template (e.g. "{namespace}-{name}").
	// Governed only at NetworkFirewallConfig levels 3-4 (mandatory) and 6-7 (defaults).
	// KropathConfig.networkfirewall does NOT carry namingTemplate.
	// Empty string = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created firewall resources.
	// Additive map merge across NetworkFirewallConfig tiers only.
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to created firewall resources.
	// Additive map merge across NetworkFirewallConfig tiers only.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`

	// Tags are cloud resource tags for this Network Firewall config profile.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// EffectiveNetworkFirewallSection is one tier (mandatory or defaults) of the merged
// Network Firewall governance result written into NetworkFirewallConfig.status.effectiveConfig
// by the controller.
type EffectiveNetworkFirewallSection struct {
	DeleteProtection               bool              `json:"deleteProtection,omitempty"`
	FirewallPolicyChangeProtection bool              `json:"firewallPolicyChangeProtection,omitempty"`
	SubnetChangeProtection         bool              `json:"subnetChangeProtection,omitempty"`
	EncryptionType                 string            `json:"encryptionType,omitempty"`
	StatefulRuleOrder              string            `json:"statefulRuleOrder,omitempty"`
	StatefulDefaultActions         []string          `json:"statefulDefaultActions,omitempty"`
	StreamExceptionPolicy          string            `json:"streamExceptionPolicy,omitempty"`
	NamingTemplate                 string            `json:"namingTemplate,omitempty"`
	SyncedLabels                   map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations              map[string]string `json:"syncedAnnotations,omitempty"`
	Tags                           map[string]string `json:"tags,omitempty"`
}

// EffectiveNetworkFirewallConfig is the merged Network Firewall governance result written into
// NetworkFirewallConfig.status.effectiveConfig by the controller.
type EffectiveNetworkFirewallConfig struct {
	Mandatory EffectiveNetworkFirewallSection `json:"mandatory"`
	Defaults  EffectiveNetworkFirewallSection `json:"defaults"`
}

// MergeNetworkFirewallCascade merges Network Firewall governance fields from all cascade
// sources and returns the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for Network Firewall (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.networkfirewall)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.networkfirewall)
//	Level 3 — globalNFWCfgMandatory   (NetworkFirewallConfig in kro-system, mandatory)
//	Level 4 — localNFWCfgMandatory    (NetworkFirewallConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localNFWCfgDefaults     (NetworkFirewallConfig in resource namespace, defaults)
//	Level 7 — globalNFWCfgDefaults    (NetworkFirewallConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.networkfirewall)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.networkfirewall)
//
// Boolean fields (deleteProtection, firewallPolicyChangeProtection, subnetChangeProtection):
//   firstTrue semantics — false = not enforced at this level; only true participates in cascade.
// String fields: firstNonEmptyString in priority order (lowest-numbered level wins).
// StatefulDefaultActions: firstNonEmptyStrings — nil/empty = not enforced.
// NamingTemplate: governed only at NetworkFirewallConfig levels (3-4 mandatory, 6-7 defaults).
// Tags: additive union merge across all mandatory levels / all defaults levels.
// SyncedLabels/SyncedAnnotations: additive union from NetworkFirewallConfig levels only.
func MergeNetworkFirewallCascade(
	globalKropathMandatory NetworkFirewallKropathSection, // level 1
	localKropathMandatory NetworkFirewallKropathSection, // level 2
	globalNFWCfgMandatory NetworkFirewallConfigSection, // level 3
	localNFWCfgMandatory NetworkFirewallConfigSection, // level 4
	localNFWCfgDefaults NetworkFirewallConfigSection, // level 6
	globalNFWCfgDefaults NetworkFirewallConfigSection, // level 7
	localKropathDefaults NetworkFirewallKropathSection, // level 8
	globalKropathDefaults NetworkFirewallKropathSection, // level 9
) EffectiveNetworkFirewallConfig {
	return EffectiveNetworkFirewallConfig{
		Mandatory: EffectiveNetworkFirewallSection{
			// Boolean protection fields: false = not enforced; firstTrue finds the first true.
			DeleteProtection: firstTrue(
				globalKropathMandatory.DeleteProtection,
				localKropathMandatory.DeleteProtection,
				globalNFWCfgMandatory.DeleteProtection,
				localNFWCfgMandatory.DeleteProtection,
			),
			FirewallPolicyChangeProtection: firstTrue(
				globalKropathMandatory.FirewallPolicyChangeProtection,
				localKropathMandatory.FirewallPolicyChangeProtection,
				globalNFWCfgMandatory.FirewallPolicyChangeProtection,
				localNFWCfgMandatory.FirewallPolicyChangeProtection,
			),
			SubnetChangeProtection: firstTrue(
				globalKropathMandatory.SubnetChangeProtection,
				localKropathMandatory.SubnetChangeProtection,
				globalNFWCfgMandatory.SubnetChangeProtection,
				localNFWCfgMandatory.SubnetChangeProtection,
			),
			EncryptionType: firstNonEmptyString(
				globalKropathMandatory.EncryptionType,
				localKropathMandatory.EncryptionType,
				globalNFWCfgMandatory.EncryptionType,
				localNFWCfgMandatory.EncryptionType,
			),
			StatefulRuleOrder: firstNonEmptyString(
				globalKropathMandatory.StatefulRuleOrder,
				localKropathMandatory.StatefulRuleOrder,
				globalNFWCfgMandatory.StatefulRuleOrder,
				localNFWCfgMandatory.StatefulRuleOrder,
			),
			StatefulDefaultActions: firstNonEmptyStrings(
				globalKropathMandatory.StatefulDefaultActions,
				localKropathMandatory.StatefulDefaultActions,
				globalNFWCfgMandatory.StatefulDefaultActions,
				localNFWCfgMandatory.StatefulDefaultActions,
			),
			StreamExceptionPolicy: firstNonEmptyString(
				globalKropathMandatory.StreamExceptionPolicy,
				localKropathMandatory.StreamExceptionPolicy,
				globalNFWCfgMandatory.StreamExceptionPolicy,
				localNFWCfgMandatory.StreamExceptionPolicy,
			),
			// NamingTemplate: NetworkFirewallConfig levels only (3, 4).
			NamingTemplate: firstNonEmptyString(
				globalNFWCfgMandatory.NamingTemplate,
				localNFWCfgMandatory.NamingTemplate,
			),
			// SyncedLabels: additive union from NetworkFirewallConfig levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localNFWCfgMandatory.SyncedLabels,
				globalNFWCfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localNFWCfgMandatory.SyncedAnnotations,
				globalNFWCfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localNFWCfgMandatory.Tags,
				globalNFWCfgMandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
		},
		Defaults: EffectiveNetworkFirewallSection{
			DeleteProtection: firstTrue(
				localNFWCfgDefaults.DeleteProtection,
				globalNFWCfgDefaults.DeleteProtection,
				localKropathDefaults.DeleteProtection,
				globalKropathDefaults.DeleteProtection,
			),
			FirewallPolicyChangeProtection: firstTrue(
				localNFWCfgDefaults.FirewallPolicyChangeProtection,
				globalNFWCfgDefaults.FirewallPolicyChangeProtection,
				localKropathDefaults.FirewallPolicyChangeProtection,
				globalKropathDefaults.FirewallPolicyChangeProtection,
			),
			SubnetChangeProtection: firstTrue(
				localNFWCfgDefaults.SubnetChangeProtection,
				globalNFWCfgDefaults.SubnetChangeProtection,
				localKropathDefaults.SubnetChangeProtection,
				globalKropathDefaults.SubnetChangeProtection,
			),
			EncryptionType: firstNonEmptyString(
				localNFWCfgDefaults.EncryptionType,
				globalNFWCfgDefaults.EncryptionType,
				localKropathDefaults.EncryptionType,
				globalKropathDefaults.EncryptionType,
			),
			StatefulRuleOrder: firstNonEmptyString(
				localNFWCfgDefaults.StatefulRuleOrder,
				globalNFWCfgDefaults.StatefulRuleOrder,
				localKropathDefaults.StatefulRuleOrder,
				globalKropathDefaults.StatefulRuleOrder,
			),
			StatefulDefaultActions: firstNonEmptyStrings(
				localNFWCfgDefaults.StatefulDefaultActions,
				globalNFWCfgDefaults.StatefulDefaultActions,
				localKropathDefaults.StatefulDefaultActions,
				globalKropathDefaults.StatefulDefaultActions,
			),
			StreamExceptionPolicy: firstNonEmptyString(
				localNFWCfgDefaults.StreamExceptionPolicy,
				globalNFWCfgDefaults.StreamExceptionPolicy,
				localKropathDefaults.StreamExceptionPolicy,
				globalKropathDefaults.StreamExceptionPolicy,
			),
			// NamingTemplate: NetworkFirewallConfig levels only (6, 7).
			NamingTemplate: firstNonEmptyString(
				localNFWCfgDefaults.NamingTemplate,
				globalNFWCfgDefaults.NamingTemplate,
			),
			// SyncedLabels: additive union from NetworkFirewallConfig levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalNFWCfgDefaults.SyncedLabels,
				localNFWCfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalNFWCfgDefaults.SyncedAnnotations,
				localNFWCfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalNFWCfgDefaults.Tags,
				localNFWCfgDefaults.Tags,
			),
		},
	}
}
