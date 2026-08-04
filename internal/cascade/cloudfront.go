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

// CloudFrontKropathSection holds the CloudFront-family governance fields from
// KropathConfig.spec.mandatory.cloudfront / .defaults.cloudfront (ADR-015 §3.5).
//
// Only 4 scalar fields are governed at the KropathConfig level:
// viewerProtocolPolicy, minimumProtocolVersion, webACLRequired, and loggingEnabled.
// The remaining CloudFront-specific fields (httpVersion, sslSupportMethod, loggingBucket,
// priceClass, geoRestrictionType, oacSigningBehavior, namingTemplate) are
// CloudFrontConfig-only (family design §8).
//
// Zero value of each field is the permissive sentinel (not enforced).
type CloudFrontKropathSection struct {
	// ViewerProtocolPolicy enforces the viewer protocol policy org-wide.
	// Empty string = not enforced; "https-only" | "redirect-to-https".
	ViewerProtocolPolicy string `json:"viewerProtocolPolicy,omitempty"`

	// MinimumProtocolVersion enforces the minimum TLS version org-wide.
	// Empty string = not enforced; e.g. "TLSv1.2_2021".
	MinimumProtocolVersion string `json:"minimumProtocolVersion,omitempty"`

	// WebACLRequired enforces WAF association on all distributions org-wide when true.
	// false (zero value) = not enforced.
	WebACLRequired bool `json:"webACLRequired,omitempty"`

	// LoggingEnabled enforces access logging on all distributions org-wide when true.
	// false (zero value) = not enforced.
	LoggingEnabled bool `json:"loggingEnabled,omitempty"`

	// Tags are tier-level cloud resource tags.
	// The reconciler populates this from KropathConfig.spec.mandatory.tags /
	// .defaults.tags so that tag union merge flows through MergeCloudFrontCascade
	// alongside the CloudFront-specific fields.
	// nil / empty map = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// CloudFrontConfigSection holds the CloudFront governance fields from
// CloudFrontConfig.spec.mandatory or CloudFrontConfig.spec.defaults
// (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type CloudFrontConfigSection struct {
	// ViewerProtocolPolicy enforces viewer protocol policy for this profile.
	// Empty string = not enforced; "https-only" | "redirect-to-https".
	ViewerProtocolPolicy string `json:"viewerProtocolPolicy,omitempty"`

	// MinimumProtocolVersion enforces minimum TLS version for this profile.
	// Empty string = not enforced; e.g. "TLSv1.2_2021".
	MinimumProtocolVersion string `json:"minimumProtocolVersion,omitempty"`

	// HttpVersion enforces the HTTP version for this profile.
	// Empty string = not enforced; "http2" | "http2and3" | "http3".
	// Governed only at CloudFrontConfig levels 3-4 (mandatory) and 6-7 (defaults).
	HttpVersion string `json:"httpVersion,omitempty"`

	// SslSupportMethod enforces the SSL support method for this profile.
	// Empty string = not enforced; "sni-only" | "vip".
	// Governed only at CloudFrontConfig levels 3-4 (mandatory) and 6-7 (defaults).
	SslSupportMethod string `json:"sslSupportMethod,omitempty"`

	// LoggingEnabled enforces access logging for distributions in this profile.
	// false (zero value) = not enforced.
	LoggingEnabled bool `json:"loggingEnabled,omitempty"`

	// LoggingBucket specifies the S3 bucket domain for access log delivery.
	// Empty string = not enforced.
	// Governed only at CloudFrontConfig levels 3-4 (mandatory) and 6-7 (defaults).
	LoggingBucket string `json:"loggingBucket,omitempty"`

	// PriceClass enforces the CloudFront edge tier for this profile.
	// Empty string = not enforced; "PriceClass_100" | "PriceClass_200" | "PriceClass_All".
	// Governed only at CloudFrontConfig levels 3-4 (mandatory) and 6-7 (defaults).
	PriceClass string `json:"priceClass,omitempty"`

	// WebACLRequired enforces WAF association for distributions in this profile.
	// false (zero value) = not enforced.
	WebACLRequired bool `json:"webACLRequired,omitempty"`

	// GeoRestrictionType enforces geographic restriction for this profile.
	// Empty string = not enforced; "whitelist" | "blacklist" | "none".
	// Governed only at CloudFrontConfig levels 3-4 (mandatory) and 6-7 (defaults).
	GeoRestrictionType string `json:"geoRestrictionType,omitempty"`

	// OacSigningBehavior enforces OAC signing behavior for this profile.
	// Empty string = not enforced; "always" | "never" | "no-override".
	// Governed only at CloudFrontConfig levels 3-4 (mandatory) and 6-7 (defaults).
	OacSigningBehavior string `json:"oacSigningBehavior,omitempty"`

	// NamingTemplate is the cloud resource naming template for this profile.
	// Empty string = not enforced.
	// Governed only at CloudFrontConfig levels 3-4 (mandatory) and 6-7 (defaults).
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// Tags are cloud resource tags for this CloudFront config profile.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created CloudFront resources.
	// Additive map merge across CloudFrontConfig tiers only.
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to created CloudFront resources.
	// Additive map merge across CloudFrontConfig tiers only.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveCloudFrontSection is one tier (mandatory or defaults) of the merged CloudFront
// governance result written into CloudFrontConfig.status.effectiveConfig by the controller.
type EffectiveCloudFrontSection struct {
	ViewerProtocolPolicy   string            `json:"viewerProtocolPolicy,omitempty"`
	MinimumProtocolVersion string            `json:"minimumProtocolVersion,omitempty"`
	HttpVersion            string            `json:"httpVersion,omitempty"`
	SslSupportMethod       string            `json:"sslSupportMethod,omitempty"`
	LoggingEnabled         bool              `json:"loggingEnabled,omitempty"`
	LoggingBucket          string            `json:"loggingBucket,omitempty"`
	PriceClass             string            `json:"priceClass,omitempty"`
	WebACLRequired         bool              `json:"webACLRequired,omitempty"`
	GeoRestrictionType     string            `json:"geoRestrictionType,omitempty"`
	OacSigningBehavior     string            `json:"oacSigningBehavior,omitempty"`
	NamingTemplate         string            `json:"namingTemplate,omitempty"`
	Tags                   map[string]string `json:"tags,omitempty"`
	SyncedLabels           map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations      map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveCloudFrontConfig is the merged CloudFront governance result written into
// CloudFrontConfig.status.effectiveConfig by the controller.
type EffectiveCloudFrontConfig struct {
	Mandatory EffectiveCloudFrontSection `json:"mandatory"`
	Defaults  EffectiveCloudFrontSection `json:"defaults"`
}

// MergeCloudFrontCascade merges CloudFront governance fields from all cascade sources and
// returns the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for CloudFront (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.cloudfront)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.cloudfront)
//	Level 3 — globalCFCfgMandatory    (CloudFrontConfig in kro-system, mandatory)
//	Level 4 — localCFCfgMandatory     (CloudFrontConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localCFCfgDefaults      (CloudFrontConfig in resource namespace, defaults)
//	Level 7 — globalCFCfgDefaults     (CloudFrontConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.cloudfront)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.cloudfront)
//
// Scalar merge: firstNonEmptyString / firstTrue in priority order (lowest number wins).
// Tags: additive union merge across all four mandatory levels, all four defaults levels.
// SyncedLabels/SyncedAnnotations: additive union from CloudFrontConfig levels only (no KropathConfig).
// HttpVersion, SslSupportMethod, LoggingBucket, PriceClass, GeoRestrictionType,
// OacSigningBehavior, NamingTemplate: governed only at CloudFrontConfig levels 3-4 (mandatory)
// and 6-7 (defaults).
func MergeCloudFrontCascade(
	globalKropathMandatory CloudFrontKropathSection, // level 1
	localKropathMandatory CloudFrontKropathSection, // level 2
	globalCFCfgMandatory CloudFrontConfigSection, // level 3
	localCFCfgMandatory CloudFrontConfigSection, // level 4
	localCFCfgDefaults CloudFrontConfigSection, // level 6
	globalCFCfgDefaults CloudFrontConfigSection, // level 7
	localKropathDefaults CloudFrontKropathSection, // level 8
	globalKropathDefaults CloudFrontKropathSection, // level 9
) EffectiveCloudFrontConfig {
	return EffectiveCloudFrontConfig{
		Mandatory: EffectiveCloudFrontSection{
			// viewerProtocolPolicy: KropathConfig levels 1-2 + CloudFrontConfig levels 3-4.
			ViewerProtocolPolicy: firstNonEmptyString(
				globalKropathMandatory.ViewerProtocolPolicy, // level 1
				localKropathMandatory.ViewerProtocolPolicy,  // level 2
				globalCFCfgMandatory.ViewerProtocolPolicy,   // level 3
				localCFCfgMandatory.ViewerProtocolPolicy,    // level 4
			),
			// minimumProtocolVersion: KropathConfig levels 1-2 + CloudFrontConfig levels 3-4.
			MinimumProtocolVersion: firstNonEmptyString(
				globalKropathMandatory.MinimumProtocolVersion, // level 1
				localKropathMandatory.MinimumProtocolVersion,  // level 2
				globalCFCfgMandatory.MinimumProtocolVersion,   // level 3
				localCFCfgMandatory.MinimumProtocolVersion,    // level 4
			),
			// httpVersion: CloudFrontConfig levels only (3, 4).
			HttpVersion: firstNonEmptyString(
				globalCFCfgMandatory.HttpVersion, // level 3
				localCFCfgMandatory.HttpVersion,  // level 4
			),
			// sslSupportMethod: CloudFrontConfig levels only (3, 4).
			SslSupportMethod: firstNonEmptyString(
				globalCFCfgMandatory.SslSupportMethod, // level 3
				localCFCfgMandatory.SslSupportMethod,  // level 4
			),
			// loggingEnabled: KropathConfig levels 1-2 + CloudFrontConfig levels 3-4.
			LoggingEnabled: firstTrue(
				globalKropathMandatory.LoggingEnabled, // level 1
				localKropathMandatory.LoggingEnabled,  // level 2
				globalCFCfgMandatory.LoggingEnabled,   // level 3
				localCFCfgMandatory.LoggingEnabled,    // level 4
			),
			// loggingBucket: CloudFrontConfig levels only (3, 4).
			LoggingBucket: firstNonEmptyString(
				globalCFCfgMandatory.LoggingBucket, // level 3
				localCFCfgMandatory.LoggingBucket,  // level 4
			),
			// priceClass: CloudFrontConfig levels only (3, 4).
			PriceClass: firstNonEmptyString(
				globalCFCfgMandatory.PriceClass, // level 3
				localCFCfgMandatory.PriceClass,  // level 4
			),
			// webACLRequired: KropathConfig levels 1-2 + CloudFrontConfig levels 3-4.
			WebACLRequired: firstTrue(
				globalKropathMandatory.WebACLRequired, // level 1
				localKropathMandatory.WebACLRequired,  // level 2
				globalCFCfgMandatory.WebACLRequired,   // level 3
				localCFCfgMandatory.WebACLRequired,    // level 4
			),
			// geoRestrictionType: CloudFrontConfig levels only (3, 4).
			GeoRestrictionType: firstNonEmptyString(
				globalCFCfgMandatory.GeoRestrictionType, // level 3
				localCFCfgMandatory.GeoRestrictionType,  // level 4
			),
			// oacSigningBehavior: CloudFrontConfig levels only (3, 4).
			OacSigningBehavior: firstNonEmptyString(
				globalCFCfgMandatory.OacSigningBehavior, // level 3
				localCFCfgMandatory.OacSigningBehavior,  // level 4
			),
			// namingTemplate: CloudFrontConfig levels only (3, 4).
			NamingTemplate: firstNonEmptyString(
				globalCFCfgMandatory.NamingTemplate, // level 3
				localCFCfgMandatory.NamingTemplate,  // level 4
			),
			// SyncedLabels: additive union from CloudFrontConfig levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localCFCfgMandatory.SyncedLabels,
				globalCFCfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localCFCfgMandatory.SyncedAnnotations,
				globalCFCfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localCFCfgMandatory.Tags,   // level 4 (lowest priority)
				globalCFCfgMandatory.Tags,  // level 3
				localKropathMandatory.Tags, // level 2
				globalKropathMandatory.Tags, // level 1 (highest priority)
			),
		},
		Defaults: EffectiveCloudFrontSection{
			// viewerProtocolPolicy: CloudFrontConfig levels 6-7 + KropathConfig levels 8-9.
			ViewerProtocolPolicy: firstNonEmptyString(
				localCFCfgDefaults.ViewerProtocolPolicy,  // level 6
				globalCFCfgDefaults.ViewerProtocolPolicy, // level 7
				localKropathDefaults.ViewerProtocolPolicy,  // level 8
				globalKropathDefaults.ViewerProtocolPolicy, // level 9
			),
			// minimumProtocolVersion: CloudFrontConfig levels 6-7 + KropathConfig levels 8-9.
			MinimumProtocolVersion: firstNonEmptyString(
				localCFCfgDefaults.MinimumProtocolVersion,  // level 6
				globalCFCfgDefaults.MinimumProtocolVersion, // level 7
				localKropathDefaults.MinimumProtocolVersion,  // level 8
				globalKropathDefaults.MinimumProtocolVersion, // level 9
			),
			// httpVersion: CloudFrontConfig levels only (6, 7).
			HttpVersion: firstNonEmptyString(
				localCFCfgDefaults.HttpVersion,  // level 6
				globalCFCfgDefaults.HttpVersion, // level 7
			),
			// sslSupportMethod: CloudFrontConfig levels only (6, 7).
			SslSupportMethod: firstNonEmptyString(
				localCFCfgDefaults.SslSupportMethod,  // level 6
				globalCFCfgDefaults.SslSupportMethod, // level 7
			),
			// loggingEnabled: CloudFrontConfig levels 6-7 + KropathConfig levels 8-9.
			LoggingEnabled: firstTrue(
				localCFCfgDefaults.LoggingEnabled,  // level 6
				globalCFCfgDefaults.LoggingEnabled, // level 7
				localKropathDefaults.LoggingEnabled,  // level 8
				globalKropathDefaults.LoggingEnabled, // level 9
			),
			// loggingBucket: CloudFrontConfig levels only (6, 7).
			LoggingBucket: firstNonEmptyString(
				localCFCfgDefaults.LoggingBucket,  // level 6
				globalCFCfgDefaults.LoggingBucket, // level 7
			),
			// priceClass: CloudFrontConfig levels only (6, 7).
			PriceClass: firstNonEmptyString(
				localCFCfgDefaults.PriceClass,  // level 6
				globalCFCfgDefaults.PriceClass, // level 7
			),
			// webACLRequired: CloudFrontConfig levels 6-7 + KropathConfig levels 8-9.
			WebACLRequired: firstTrue(
				localCFCfgDefaults.WebACLRequired,  // level 6
				globalCFCfgDefaults.WebACLRequired, // level 7
				localKropathDefaults.WebACLRequired,  // level 8
				globalKropathDefaults.WebACLRequired, // level 9
			),
			// geoRestrictionType: CloudFrontConfig levels only (6, 7).
			GeoRestrictionType: firstNonEmptyString(
				localCFCfgDefaults.GeoRestrictionType,  // level 6
				globalCFCfgDefaults.GeoRestrictionType, // level 7
			),
			// oacSigningBehavior: CloudFrontConfig levels only (6, 7).
			OacSigningBehavior: firstNonEmptyString(
				localCFCfgDefaults.OacSigningBehavior,  // level 6
				globalCFCfgDefaults.OacSigningBehavior, // level 7
			),
			// namingTemplate: CloudFrontConfig levels only (6, 7).
			NamingTemplate: firstNonEmptyString(
				localCFCfgDefaults.NamingTemplate,  // level 6
				globalCFCfgDefaults.NamingTemplate, // level 7
			),
			// SyncedLabels: additive union from CloudFrontConfig levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalCFCfgDefaults.SyncedLabels,
				localCFCfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalCFCfgDefaults.SyncedAnnotations,
				localCFCfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags, // level 9 (lowest priority)
				localKropathDefaults.Tags,  // level 8
				globalCFCfgDefaults.Tags,   // level 7
				localCFCfgDefaults.Tags,    // level 6 (highest priority)
			),
		},
	}
}
