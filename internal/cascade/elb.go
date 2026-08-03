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

// ELBKropathSection holds the ELB-family governance fields from
// KropathConfig.spec.mandatory.elb / .defaults.elb (ADR-015 §3.5).
//
// Only the three cross-type ELB fields live here: deletionProtection,
// accessLogsEnabled, and internalOnly. accessLogsS3Bucket, crossZoneEnabled,
// idleTimeoutSeconds, sslPolicy, and namingTemplate are per-profile choices
// and are not in KropathConfig (family design §8).
//
// The Tags field is NOT part of the KropathConfig elb schema; it is populated
// by the reconciler from the tier-level KropathConfig.spec.mandatory.tags /
// .defaults.tags so that tier-level cloud tags flow through MergeELBCascade
// alongside the ELB-specific fields.
//
// Zero value of each field is the permissive sentinel (not enforced).
type ELBKropathSection struct {
	// DeletionProtection enforces load balancer deletion protection org-wide when true.
	// false (zero value) = not enforced.
	DeletionProtection bool `json:"deletionProtection,omitempty"`

	// AccessLogsEnabled enforces access logging for all load balancers when true.
	// false (zero value) = not enforced.
	AccessLogsEnabled bool `json:"accessLogsEnabled,omitempty"`

	// InternalOnly forces load balancer scheme to "internal" when true.
	// false (zero value) = not enforced (both schemes allowed).
	InternalOnly bool `json:"internalOnly,omitempty"`

	// Tags are tier-level cloud resource tags from KropathConfig.spec.mandatory.tags.
	// Populated by the reconciler; not exposed in the KropathConfig elb CRD schema.
	// nil / empty map (zero value) = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// ELBConfigSection holds the ELB governance fields from AWSELBConfig.spec.mandatory
// or AWSELBConfig.spec.defaults (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type ELBConfigSection struct {
	// DeletionProtection enforces load balancer deletion protection. false = not enforced.
	DeletionProtection bool `json:"deletionProtection,omitempty"`

	// AccessLogsEnabled enforces access logging. false = not enforced.
	AccessLogsEnabled bool `json:"accessLogsEnabled,omitempty"`

	// AccessLogsS3Bucket is the S3 bucket for access log delivery.
	// Empty string = not enforced.
	AccessLogsS3Bucket string `json:"accessLogsS3Bucket,omitempty"`

	// CrossZoneEnabled enforces cross-zone load balancing. false = not enforced.
	// ALB always has cross-zone enabled; this field affects NLB/GWLB.
	CrossZoneEnabled bool `json:"crossZoneEnabled,omitempty"`

	// InternalOnly forces load balancer scheme to "internal". false = not enforced.
	InternalOnly bool `json:"internalOnly,omitempty"`

	// IdleTimeoutSeconds is the enforced idle timeout in seconds for ALBs.
	// 0 (zero value) = not enforced (sentinel). Range: 1–4000.
	IdleTimeoutSeconds int64 `json:"idleTimeoutSeconds,omitempty"`

	// SslPolicy is the enforced TLS security policy for HTTPS/TLS listeners.
	// Empty string = not enforced (sentinel).
	SslPolicy string `json:"sslPolicy,omitempty"`

	// NamingTemplate is the template for cloud resource names.
	// Tokens: {name}, {namespace}, {account_id}, {region}, {configRef}, {tag.<key>}.
	// Empty string = not enforced (sentinel).
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// Tags are cloud resource tags for this ELB config profile.
	// nil / empty map (zero value) = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created ELB resources.
	// Additive map merge across ELBConfig tiers only.
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to created ELB resources.
	// Additive map merge across ELBConfig tiers only.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveELBSection is one tier (mandatory or defaults) of the merged ELB governance
// result written into AWSELBConfig.status.effectiveConfig by the controller.
type EffectiveELBSection struct {
	DeletionProtection bool              `json:"deletionProtection,omitempty"`
	AccessLogsEnabled  bool              `json:"accessLogsEnabled,omitempty"`
	AccessLogsS3Bucket string            `json:"accessLogsS3Bucket,omitempty"`
	CrossZoneEnabled   bool              `json:"crossZoneEnabled,omitempty"`
	InternalOnly       bool              `json:"internalOnly,omitempty"`
	IdleTimeoutSeconds int64             `json:"idleTimeoutSeconds,omitempty"`
	SslPolicy          string            `json:"sslPolicy,omitempty"`
	NamingTemplate     string            `json:"namingTemplate,omitempty"`
	Tags               map[string]string `json:"tags,omitempty"`
	SyncedLabels       map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations  map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveELBConfig is the merged ELB governance result written into
// AWSELBConfig.status.effectiveConfig by the controller.
type EffectiveELBConfig struct {
	Mandatory EffectiveELBSection `json:"mandatory"`
	Defaults  EffectiveELBSection `json:"defaults"`
}

// MergeELBCascade merges ELB governance fields from all cascade sources and
// returns the effective configuration to be written to status.effectiveConfig.
//
// The ten-level priority chain (ADR-015 §5.3) for ELB fields:
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.elb)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.elb)
//	Level 3 — globalELBCfgMandatory   (AWSELBConfig in kro-system, mandatory)
//	Level 4 — localELBCfgMandatory    (AWSELBConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localELBCfgDefaults     (AWSELBConfig in resource namespace, defaults)
//	Level 7 — globalELBCfgDefaults    (AWSELBConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.elb)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.elb)
//
// Scalar merge: firstTrue / firstNonEmptyString / firstNonZeroInt64 in priority order.
// Tags: additive union merge across all four mandatory levels, all four defaults levels.
// SyncedLabels/SyncedAnnotations: additive union from AWSELBConfig levels only (no KropathConfig).
// NamingTemplate, SslPolicy, AccessLogsS3Bucket, CrossZoneEnabled, IdleTimeoutSeconds:
//
//	governed only at AWSELBConfig levels (3-4 mandatory, 6-7 defaults).
func MergeELBCascade(
	globalKropathMandatory ELBKropathSection, // level 1
	localKropathMandatory ELBKropathSection, // level 2
	globalELBCfgMandatory ELBConfigSection, // level 3
	localELBCfgMandatory ELBConfigSection, // level 4
	localELBCfgDefaults ELBConfigSection, // level 6
	globalELBCfgDefaults ELBConfigSection, // level 7
	localKropathDefaults ELBKropathSection, // level 8
	globalKropathDefaults ELBKropathSection, // level 9
) EffectiveELBConfig {
	return EffectiveELBConfig{
		Mandatory: EffectiveELBSection{
			DeletionProtection: firstTrue(
				globalKropathMandatory.DeletionProtection, // level 1
				localKropathMandatory.DeletionProtection,  // level 2
				globalELBCfgMandatory.DeletionProtection,  // level 3
				localELBCfgMandatory.DeletionProtection,   // level 4
			),
			AccessLogsEnabled: firstTrue(
				globalKropathMandatory.AccessLogsEnabled, // level 1
				localKropathMandatory.AccessLogsEnabled,  // level 2
				globalELBCfgMandatory.AccessLogsEnabled,  // level 3
				localELBCfgMandatory.AccessLogsEnabled,   // level 4
			),
			// AccessLogsS3Bucket: not in KropathConfig; AWSELBConfig levels only.
			AccessLogsS3Bucket: firstNonEmptyString(
				globalELBCfgMandatory.AccessLogsS3Bucket, // level 3
				localELBCfgMandatory.AccessLogsS3Bucket,  // level 4
			),
			// CrossZoneEnabled: not in KropathConfig; AWSELBConfig levels only.
			CrossZoneEnabled: firstTrue(
				globalELBCfgMandatory.CrossZoneEnabled, // level 3
				localELBCfgMandatory.CrossZoneEnabled,  // level 4
			),
			InternalOnly: firstTrue(
				globalKropathMandatory.InternalOnly, // level 1
				localKropathMandatory.InternalOnly,  // level 2
				globalELBCfgMandatory.InternalOnly,  // level 3
				localELBCfgMandatory.InternalOnly,   // level 4
			),
			// IdleTimeoutSeconds: not in KropathConfig; AWSELBConfig levels only.
			IdleTimeoutSeconds: firstNonZeroInt64(
				globalELBCfgMandatory.IdleTimeoutSeconds, // level 3
				localELBCfgMandatory.IdleTimeoutSeconds,  // level 4
			),
			// SslPolicy: not in KropathConfig; AWSELBConfig levels only.
			SslPolicy: firstNonEmptyString(
				globalELBCfgMandatory.SslPolicy, // level 3
				localELBCfgMandatory.SslPolicy,  // level 4
			),
			// NamingTemplate: not in KropathConfig; AWSELBConfig levels only.
			NamingTemplate: firstNonEmptyString(
				globalELBCfgMandatory.NamingTemplate, // level 3
				localELBCfgMandatory.NamingTemplate,  // level 4
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localELBCfgMandatory.Tags,   // level 4 (lowest priority)
				globalELBCfgMandatory.Tags,  // level 3
				localKropathMandatory.Tags,  // level 2
				globalKropathMandatory.Tags, // level 1 (highest priority)
			),
			// SyncedLabels: additive union from AWSELBConfig levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localELBCfgMandatory.SyncedLabels,
				globalELBCfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localELBCfgMandatory.SyncedAnnotations,
				globalELBCfgMandatory.SyncedAnnotations,
			),
		},
		Defaults: EffectiveELBSection{
			DeletionProtection: firstTrue(
				localELBCfgDefaults.DeletionProtection,  // level 6
				globalELBCfgDefaults.DeletionProtection, // level 7
				localKropathDefaults.DeletionProtection,  // level 8
				globalKropathDefaults.DeletionProtection, // level 9
			),
			AccessLogsEnabled: firstTrue(
				localELBCfgDefaults.AccessLogsEnabled,  // level 6
				globalELBCfgDefaults.AccessLogsEnabled, // level 7
				localKropathDefaults.AccessLogsEnabled,  // level 8
				globalKropathDefaults.AccessLogsEnabled, // level 9
			),
			// AccessLogsS3Bucket: not in KropathConfig; AWSELBConfig levels only.
			AccessLogsS3Bucket: firstNonEmptyString(
				localELBCfgDefaults.AccessLogsS3Bucket,  // level 6
				globalELBCfgDefaults.AccessLogsS3Bucket, // level 7
			),
			// CrossZoneEnabled: not in KropathConfig; AWSELBConfig levels only.
			CrossZoneEnabled: firstTrue(
				localELBCfgDefaults.CrossZoneEnabled,  // level 6
				globalELBCfgDefaults.CrossZoneEnabled, // level 7
			),
			InternalOnly: firstTrue(
				localELBCfgDefaults.InternalOnly,  // level 6
				globalELBCfgDefaults.InternalOnly, // level 7
				localKropathDefaults.InternalOnly,  // level 8
				globalKropathDefaults.InternalOnly, // level 9
			),
			// IdleTimeoutSeconds: not in KropathConfig; AWSELBConfig levels only.
			IdleTimeoutSeconds: firstNonZeroInt64(
				localELBCfgDefaults.IdleTimeoutSeconds,  // level 6
				globalELBCfgDefaults.IdleTimeoutSeconds, // level 7
			),
			// SslPolicy: not in KropathConfig; AWSELBConfig levels only.
			SslPolicy: firstNonEmptyString(
				localELBCfgDefaults.SslPolicy,  // level 6
				globalELBCfgDefaults.SslPolicy, // level 7
			),
			// NamingTemplate: not in KropathConfig; AWSELBConfig levels only.
			NamingTemplate: firstNonEmptyString(
				localELBCfgDefaults.NamingTemplate,  // level 6
				globalELBCfgDefaults.NamingTemplate, // level 7
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags, // level 9 (lowest priority)
				localKropathDefaults.Tags,  // level 8
				globalELBCfgDefaults.Tags,  // level 7
				localELBCfgDefaults.Tags,   // level 6 (highest priority)
			),
			// SyncedLabels: additive union from AWSELBConfig levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalELBCfgDefaults.SyncedLabels,
				localELBCfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalELBCfgDefaults.SyncedAnnotations,
				localELBCfgDefaults.SyncedAnnotations,
			),
		},
	}
}
