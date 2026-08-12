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

// ApiGatewayV2KropathSection holds the API Gateway V2-family governance fields from
// KropathConfig.spec.mandatory.apigatewayv2 / .defaults.apigatewayv2 (ADR-015 §3.5),
// PLUS the tier-level tags from KropathConfig.spec.mandatory.tags / .defaults.tags.
//
// Three governance fields are promoted to KropathConfig: corsEnabled,
// disableExecuteApiEndpoint, minimumTlsVersion. Tags are org-wide tier-level fields
// populated by the reconciler so the full tag cascade flows through MergeApiGatewayV2Cascade.
//
// Zero value of each field is the permissive sentinel (not enforced).
type ApiGatewayV2KropathSection struct {
	// CorsEnabled is the org-wide enforcement gate for CORS.
	// true = all HTTP APIs must define a corsConfiguration. false = not enforced.
	CorsEnabled bool `json:"corsEnabled,omitempty"`

	// DisableExecuteApiEndpoint is the org-wide enforcement gate for the default execute-api endpoint.
	// true = all APIs must disable the default execute-api endpoint. false = not enforced.
	DisableExecuteApiEndpoint bool `json:"disableExecuteApiEndpoint,omitempty"`

	// MinimumTlsVersion is the org-wide enforced minimum TLS version for custom domain names.
	// Empty string = not enforced. e.g. "TLS_1_2".
	MinimumTlsVersion string `json:"minimumTlsVersion,omitempty"`

	// Tags are tier-level cloud resource tags from KropathConfig.spec.mandatory.tags
	// or .defaults.tags. Populated by the reconciler.
	Tags map[string]string `json:"tags,omitempty"`
}

// ApiGatewayV2ConfigSection holds the API Gateway V2 governance fields from
// ApiGatewayV2Config.spec.mandatory or ApiGatewayV2Config.spec.defaults
// (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type ApiGatewayV2ConfigSection struct {
	// CorsEnabled enforces CORS configuration for HTTP APIs in this profile.
	CorsEnabled bool `json:"corsEnabled,omitempty"`

	// DisableExecuteApiEndpoint enforces disabling the default execute-api endpoint.
	DisableExecuteApiEndpoint bool `json:"disableExecuteApiEndpoint,omitempty"`

	// DefaultThrottlingBurstLimit is the burst throttle limit for all stages in scope.
	// 0 = not enforced (mandatory) or use account default (defaults).
	DefaultThrottlingBurstLimit int64 `json:"defaultThrottlingBurstLimit,omitempty"`

	// DefaultThrottlingRateLimit is the rate throttle limit for all stages in scope.
	// 0.0 = not enforced (mandatory) or use account default (defaults).
	DefaultThrottlingRateLimit float64 `json:"defaultThrottlingRateLimit,omitempty"`

	// MinimumTlsVersion is the minimum TLS version for custom domain names.
	// Empty string = not enforced. e.g. "TLS_1_2".
	MinimumTlsVersion string `json:"minimumTlsVersion,omitempty"`

	// AccessLogDestinationArn is the ARN of the CloudWatch Logs or Kinesis Firehose destination.
	// Empty string = not enforced.
	AccessLogDestinationArn string `json:"accessLogDestinationArn,omitempty"`

	// AccessLogFormat is the access log format string.
	// Empty string = not enforced.
	AccessLogFormat string `json:"accessLogFormat,omitempty"`

	// NamingTemplate is the API Gateway V2 resource naming template.
	// Empty string = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// Tags are cloud resource tags for this API Gateway V2 config profile.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to API Gateway V2 resources.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to API Gateway V2 resources.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveApiGatewayV2Section is one tier (mandatory or defaults) of the merged API Gateway V2
// governance result written into ApiGatewayV2Config.status.effectiveConfig by the controller.
type EffectiveApiGatewayV2Section struct {
	CorsEnabled                 bool              `json:"corsEnabled,omitempty"`
	DisableExecuteApiEndpoint   bool              `json:"disableExecuteApiEndpoint,omitempty"`
	DefaultThrottlingBurstLimit int64             `json:"defaultThrottlingBurstLimit,omitempty"`
	DefaultThrottlingRateLimit  float64           `json:"defaultThrottlingRateLimit,omitempty"`
	MinimumTlsVersion           string            `json:"minimumTlsVersion,omitempty"`
	AccessLogDestinationArn     string            `json:"accessLogDestinationArn,omitempty"`
	AccessLogFormat             string            `json:"accessLogFormat,omitempty"`
	NamingTemplate              string            `json:"namingTemplate,omitempty"`
	Tags                        map[string]string `json:"tags,omitempty"`
	SyncedLabels                map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations           map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveApiGatewayV2Config is the merged API Gateway V2 governance result written into
// ApiGatewayV2Config.status.effectiveConfig by the controller.
type EffectiveApiGatewayV2Config struct {
	Mandatory EffectiveApiGatewayV2Section `json:"mandatory"`
	Defaults  EffectiveApiGatewayV2Section `json:"defaults"`
}

// MergeApiGatewayV2Cascade merges API Gateway V2 governance fields from all cascade sources and
// returns the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for API Gateway V2 (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.apigatewayv2)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.apigatewayv2)
//	Level 3 — globalApigwv2CfgMandatory (ApiGatewayV2Config in kro-system, mandatory)
//	Level 4 — localApigwv2CfgMandatory  (ApiGatewayV2Config in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localApigwv2CfgDefaults   (ApiGatewayV2Config in resource namespace, defaults)
//	Level 7 — globalApigwv2CfgDefaults  (ApiGatewayV2Config in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.apigatewayv2)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.apigatewayv2)
//
// Boolean merge (corsEnabled, disableExecuteApiEndpoint): firstTrue in priority order.
// String merge (minimumTlsVersion, accessLogDestinationArn, accessLogFormat): firstNonEmptyString.
// Integer merge (defaultThrottlingBurstLimit): firstNonZeroInt64.
// Float merge (defaultThrottlingRateLimit): firstNonZeroFloat64.
// KropathConfig fields (corsEnabled, disableExecuteApiEndpoint, minimumTlsVersion):
//   - Mandatory: all four sources (L1 → L4).
//   - Defaults: all four sources (L6 → L9).
//
// ApiGatewayV2Config-only fields (throttling, accessLog, namingTemplate):
//   - Mandatory: L3 → L4 only.
//   - Defaults: L6 → L7 only.
//
// Tags: additive union across all four mandatory sources (L1 wins on key conflict) and all four
// defaults sources (L6 wins on key conflict).
// SyncedLabels/SyncedAnnotations: additive union from ApiGatewayV2Config levels only
// (mandatory: L3 wins; defaults: L6 wins).
func MergeApiGatewayV2Cascade(
	// Mandatory inputs (highest → lowest priority)
	globalKropathMandatory ApiGatewayV2KropathSection,     // level 1
	localKropathMandatory ApiGatewayV2KropathSection,      // level 2
	globalApigwv2CfgMandatory ApiGatewayV2ConfigSection,   // level 3
	localApigwv2CfgMandatory ApiGatewayV2ConfigSection,    // level 4
	// Defaults inputs (highest → lowest priority)
	localApigwv2CfgDefaults ApiGatewayV2ConfigSection,     // level 6
	globalApigwv2CfgDefaults ApiGatewayV2ConfigSection,    // level 7
	localKropathDefaults ApiGatewayV2KropathSection,       // level 8
	globalKropathDefaults ApiGatewayV2KropathSection,      // level 9
) EffectiveApiGatewayV2Config {
	return EffectiveApiGatewayV2Config{
		Mandatory: EffectiveApiGatewayV2Section{
			// KropathConfig fields: all four mandatory sources (L1 wins).
			CorsEnabled: firstTrue(
				globalKropathMandatory.CorsEnabled,         // level 1
				localKropathMandatory.CorsEnabled,          // level 2
				globalApigwv2CfgMandatory.CorsEnabled,      // level 3
				localApigwv2CfgMandatory.CorsEnabled,       // level 4
			),
			DisableExecuteApiEndpoint: firstTrue(
				globalKropathMandatory.DisableExecuteApiEndpoint,         // level 1
				localKropathMandatory.DisableExecuteApiEndpoint,          // level 2
				globalApigwv2CfgMandatory.DisableExecuteApiEndpoint,      // level 3
				localApigwv2CfgMandatory.DisableExecuteApiEndpoint,       // level 4
			),
			MinimumTlsVersion: firstNonEmptyString(
				globalKropathMandatory.MinimumTlsVersion,         // level 1
				localKropathMandatory.MinimumTlsVersion,          // level 2
				globalApigwv2CfgMandatory.MinimumTlsVersion,      // level 3
				localApigwv2CfgMandatory.MinimumTlsVersion,       // level 4
			),
			// ApiGatewayV2Config-only fields: L3 and L4 only (no KropathConfig equivalent).
			DefaultThrottlingBurstLimit: firstNonZeroInt64(
				globalApigwv2CfgMandatory.DefaultThrottlingBurstLimit,    // level 3
				localApigwv2CfgMandatory.DefaultThrottlingBurstLimit,     // level 4
			),
			DefaultThrottlingRateLimit: firstNonZeroFloat64(
				globalApigwv2CfgMandatory.DefaultThrottlingRateLimit,     // level 3
				localApigwv2CfgMandatory.DefaultThrottlingRateLimit,      // level 4
			),
			AccessLogDestinationArn: firstNonEmptyString(
				globalApigwv2CfgMandatory.AccessLogDestinationArn,        // level 3
				localApigwv2CfgMandatory.AccessLogDestinationArn,         // level 4
			),
			AccessLogFormat: firstNonEmptyString(
				globalApigwv2CfgMandatory.AccessLogFormat,                // level 3
				localApigwv2CfgMandatory.AccessLogFormat,                 // level 4
			),
			// NamingTemplate: ApiGatewayV2Config levels only (3, 4).
			NamingTemplate: firstNonEmptyString(
				globalApigwv2CfgMandatory.NamingTemplate,                 // level 3
				localApigwv2CfgMandatory.NamingTemplate,                  // level 4
			),
			// Tags: union of all mandatory sources; L4 added first (lowest priority), L1 wins on key conflict.
			Tags: mergeMaps(
				localApigwv2CfgMandatory.Tags,    // level 4
				globalApigwv2CfgMandatory.Tags,   // level 3
				localKropathMandatory.Tags,       // level 2
				globalKropathMandatory.Tags,      // level 1 (highest priority)
			),
			// SyncedLabels: additive union from ApiGatewayV2Config mandatory levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localApigwv2CfgMandatory.SyncedLabels,    // level 4
				globalApigwv2CfgMandatory.SyncedLabels,   // level 3
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localApigwv2CfgMandatory.SyncedAnnotations,    // level 4
				globalApigwv2CfgMandatory.SyncedAnnotations,   // level 3
			),
		},
		Defaults: EffectiveApiGatewayV2Section{
			// KropathConfig fields: all four defaults sources (L6 wins).
			CorsEnabled: firstTrue(
				localApigwv2CfgDefaults.CorsEnabled,       // level 6
				globalApigwv2CfgDefaults.CorsEnabled,      // level 7
				localKropathDefaults.CorsEnabled,          // level 8
				globalKropathDefaults.CorsEnabled,         // level 9
			),
			DisableExecuteApiEndpoint: firstTrue(
				localApigwv2CfgDefaults.DisableExecuteApiEndpoint,       // level 6
				globalApigwv2CfgDefaults.DisableExecuteApiEndpoint,      // level 7
				localKropathDefaults.DisableExecuteApiEndpoint,          // level 8
				globalKropathDefaults.DisableExecuteApiEndpoint,         // level 9
			),
			MinimumTlsVersion: firstNonEmptyString(
				localApigwv2CfgDefaults.MinimumTlsVersion,       // level 6
				globalApigwv2CfgDefaults.MinimumTlsVersion,      // level 7
				localKropathDefaults.MinimumTlsVersion,          // level 8
				globalKropathDefaults.MinimumTlsVersion,         // level 9
			),
			// ApiGatewayV2Config-only fields: L6 and L7 only.
			DefaultThrottlingBurstLimit: firstNonZeroInt64(
				localApigwv2CfgDefaults.DefaultThrottlingBurstLimit,     // level 6
				globalApigwv2CfgDefaults.DefaultThrottlingBurstLimit,    // level 7
			),
			DefaultThrottlingRateLimit: firstNonZeroFloat64(
				localApigwv2CfgDefaults.DefaultThrottlingRateLimit,      // level 6
				globalApigwv2CfgDefaults.DefaultThrottlingRateLimit,     // level 7
			),
			AccessLogDestinationArn: firstNonEmptyString(
				localApigwv2CfgDefaults.AccessLogDestinationArn,         // level 6
				globalApigwv2CfgDefaults.AccessLogDestinationArn,        // level 7
			),
			AccessLogFormat: firstNonEmptyString(
				localApigwv2CfgDefaults.AccessLogFormat,                 // level 6
				globalApigwv2CfgDefaults.AccessLogFormat,                // level 7
			),
			// NamingTemplate: ApiGatewayV2Config levels only (6, 7).
			NamingTemplate: firstNonEmptyString(
				localApigwv2CfgDefaults.NamingTemplate,                  // level 6
				globalApigwv2CfgDefaults.NamingTemplate,                 // level 7
			),
			// Tags: union of all defaults sources; L9 added first (lowest priority), L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,       // level 9
				localKropathDefaults.Tags,        // level 8
				globalApigwv2CfgDefaults.Tags,    // level 7
				localApigwv2CfgDefaults.Tags,     // level 6 (highest priority)
			),
			// SyncedLabels: additive union from ApiGatewayV2Config defaults levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalApigwv2CfgDefaults.SyncedLabels,    // level 7
				localApigwv2CfgDefaults.SyncedLabels,     // level 6
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalApigwv2CfgDefaults.SyncedAnnotations,    // level 7
				localApigwv2CfgDefaults.SyncedAnnotations,     // level 6
			),
		},
	}
}
