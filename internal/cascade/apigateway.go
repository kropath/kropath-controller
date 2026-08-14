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

// ApiGatewayKropathSection holds the API Gateway (REST) family governance fields from
// KropathConfig.spec.mandatory.apigateway / .defaults.apigateway (ADR-015 §3.5),
// PLUS the tier-level tags from KropathConfig.spec.mandatory.tags / .defaults.tags.
//
// Two governance fields are promoted to KropathConfig: endpointType and minimumTlsVersion.
// Tags are org-wide tier-level fields populated by the reconciler so the full tag cascade
// flows through MergeApiGatewayCascade.
//
// Zero value of each field is the permissive sentinel (not enforced).
type ApiGatewayKropathSection struct {
	// EndpointType is the org-wide enforced endpoint type for REST APIs.
	// Empty string = not enforced. e.g. "REGIONAL" or "EDGE".
	EndpointType string `json:"endpointType,omitempty"`

	// MinimumTlsVersion is the org-wide enforced minimum TLS version for custom domain names.
	// Empty string = not enforced. e.g. "TLS_1_2".
	MinimumTlsVersion string `json:"minimumTlsVersion,omitempty"`

	// Tags are tier-level cloud resource tags from KropathConfig.spec.mandatory.tags
	// or .defaults.tags. Populated by the reconciler.
	Tags map[string]string `json:"tags,omitempty"`
}

// ApiGatewayConfigSection holds the API Gateway (REST) governance fields from
// ApiGatewayConfig.spec.mandatory or ApiGatewayConfig.spec.defaults
// (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type ApiGatewayConfigSection struct {
	// EndpointType enforces the REST API endpoint type. Empty = not enforced.
	EndpointType string `json:"endpointType,omitempty"`

	// ApiKeySource enforces the API key source for metered APIs.
	// Empty = not enforced. e.g. "HEADER" or "AUTHORIZER".
	ApiKeySource string `json:"apiKeySource,omitempty"`

	// MinimumTlsVersion is the minimum TLS version for custom domain names.
	// Empty string = not enforced. e.g. "TLS_1_2".
	MinimumTlsVersion string `json:"minimumTlsVersion,omitempty"`

	// DisableExecuteApiEndpoint enforces disabling the default execute-api endpoint.
	// true = all REST APIs must disable the endpoint. false = not enforced.
	DisableExecuteApiEndpoint bool `json:"disableExecuteApiEndpoint,omitempty"`

	// NamingTemplate is the API Gateway resource naming template.
	// Empty string = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// Tags are cloud resource tags for this API Gateway config profile.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to API Gateway resources.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to API Gateway resources.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveApiGatewaySection is one tier (mandatory or defaults) of the merged API Gateway
// governance result written into ApiGatewayConfig.status.effectiveConfig by the controller.
type EffectiveApiGatewaySection struct {
	EndpointType              string            `json:"endpointType,omitempty"`
	ApiKeySource              string            `json:"apiKeySource,omitempty"`
	MinimumTlsVersion         string            `json:"minimumTlsVersion,omitempty"`
	DisableExecuteApiEndpoint bool              `json:"disableExecuteApiEndpoint,omitempty"`
	NamingTemplate            string            `json:"namingTemplate,omitempty"`
	Tags                      map[string]string `json:"tags,omitempty"`
	SyncedLabels              map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations         map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveApiGatewayConfig is the merged API Gateway governance result written into
// ApiGatewayConfig.status.effectiveConfig by the controller.
type EffectiveApiGatewayConfig struct {
	Mandatory EffectiveApiGatewaySection `json:"mandatory"`
	Defaults  EffectiveApiGatewaySection `json:"defaults"`
}

// MergeApiGatewayCascade merges API Gateway (REST) governance fields from all cascade sources and
// returns the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for API Gateway REST (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.apigateway)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.apigateway)
//	Level 3 — globalApigwCfgMandatory (ApiGatewayConfig in kro-system, mandatory)
//	Level 4 — localApigwCfgMandatory  (ApiGatewayConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localApigwCfgDefaults   (ApiGatewayConfig in resource namespace, defaults)
//	Level 7 — globalApigwCfgDefaults  (ApiGatewayConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.apigateway)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.apigateway)
//
// String merge (endpointType, minimumTlsVersion): firstNonEmptyString.
// Boolean merge (disableExecuteApiEndpoint): firstTrue in priority order.
//
// KropathConfig fields (endpointType, minimumTlsVersion):
//   - Mandatory: all four sources (L1 → L4).
//   - Defaults: all four sources (L6 → L9).
//
// ApiGatewayConfig-only fields (apiKeySource, disableExecuteApiEndpoint, namingTemplate):
//   - Mandatory: L3 → L4 only.
//   - Defaults: L6 → L7 only.
//
// Tags: additive union across all four mandatory sources (L1 wins on key conflict) and all four
// defaults sources (L6 wins on key conflict).
// SyncedLabels/SyncedAnnotations: additive union from ApiGatewayConfig levels only
// (mandatory: L3 wins; defaults: L6 wins).
func MergeApiGatewayCascade(
	// Mandatory inputs (highest → lowest priority)
	globalKropathMandatory ApiGatewayKropathSection,  // level 1
	localKropathMandatory ApiGatewayKropathSection,   // level 2
	globalApigwCfgMandatory ApiGatewayConfigSection,  // level 3
	localApigwCfgMandatory ApiGatewayConfigSection,   // level 4
	// Defaults inputs (highest → lowest priority)
	localApigwCfgDefaults ApiGatewayConfigSection,    // level 6
	globalApigwCfgDefaults ApiGatewayConfigSection,   // level 7
	localKropathDefaults ApiGatewayKropathSection,    // level 8
	globalKropathDefaults ApiGatewayKropathSection,   // level 9
) EffectiveApiGatewayConfig {
	return EffectiveApiGatewayConfig{
		Mandatory: EffectiveApiGatewaySection{
			// KropathConfig fields: all four mandatory sources (L1 wins).
			EndpointType: firstNonEmptyString(
				globalKropathMandatory.EndpointType,       // level 1
				localKropathMandatory.EndpointType,        // level 2
				globalApigwCfgMandatory.EndpointType,      // level 3
				localApigwCfgMandatory.EndpointType,       // level 4
			),
			MinimumTlsVersion: firstNonEmptyString(
				globalKropathMandatory.MinimumTlsVersion,       // level 1
				localKropathMandatory.MinimumTlsVersion,        // level 2
				globalApigwCfgMandatory.MinimumTlsVersion,      // level 3
				localApigwCfgMandatory.MinimumTlsVersion,       // level 4
			),
			// ApiGatewayConfig-only fields: L3 and L4 only (no KropathConfig equivalent).
			ApiKeySource: firstNonEmptyString(
				globalApigwCfgMandatory.ApiKeySource,  // level 3
				localApigwCfgMandatory.ApiKeySource,   // level 4
			),
			DisableExecuteApiEndpoint: firstTrue(
				globalApigwCfgMandatory.DisableExecuteApiEndpoint,  // level 3
				localApigwCfgMandatory.DisableExecuteApiEndpoint,   // level 4
			),
			// NamingTemplate: ApiGatewayConfig levels only (3, 4).
			NamingTemplate: firstNonEmptyString(
				globalApigwCfgMandatory.NamingTemplate,  // level 3
				localApigwCfgMandatory.NamingTemplate,   // level 4
			),
			// Tags: union of all mandatory sources; L4 added first (lowest priority), L1 wins on key conflict.
			Tags: mergeMaps(
				localApigwCfgMandatory.Tags,    // level 4
				globalApigwCfgMandatory.Tags,   // level 3
				localKropathMandatory.Tags,     // level 2
				globalKropathMandatory.Tags,    // level 1 (highest priority)
			),
			// SyncedLabels: additive union from ApiGatewayConfig mandatory levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localApigwCfgMandatory.SyncedLabels,    // level 4
				globalApigwCfgMandatory.SyncedLabels,   // level 3
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localApigwCfgMandatory.SyncedAnnotations,    // level 4
				globalApigwCfgMandatory.SyncedAnnotations,   // level 3
			),
		},
		Defaults: EffectiveApiGatewaySection{
			// KropathConfig fields: all four defaults sources (L6 wins).
			EndpointType: firstNonEmptyString(
				localApigwCfgDefaults.EndpointType,       // level 6
				globalApigwCfgDefaults.EndpointType,      // level 7
				localKropathDefaults.EndpointType,        // level 8
				globalKropathDefaults.EndpointType,       // level 9
			),
			MinimumTlsVersion: firstNonEmptyString(
				localApigwCfgDefaults.MinimumTlsVersion,       // level 6
				globalApigwCfgDefaults.MinimumTlsVersion,      // level 7
				localKropathDefaults.MinimumTlsVersion,        // level 8
				globalKropathDefaults.MinimumTlsVersion,       // level 9
			),
			// ApiGatewayConfig-only fields: L6 and L7 only.
			ApiKeySource: firstNonEmptyString(
				localApigwCfgDefaults.ApiKeySource,   // level 6
				globalApigwCfgDefaults.ApiKeySource,  // level 7
			),
			DisableExecuteApiEndpoint: firstTrue(
				localApigwCfgDefaults.DisableExecuteApiEndpoint,   // level 6
				globalApigwCfgDefaults.DisableExecuteApiEndpoint,  // level 7
			),
			// NamingTemplate: ApiGatewayConfig levels only (6, 7).
			NamingTemplate: firstNonEmptyString(
				localApigwCfgDefaults.NamingTemplate,   // level 6
				globalApigwCfgDefaults.NamingTemplate,  // level 7
			),
			// Tags: union of all defaults sources; L9 added first (lowest priority), L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,      // level 9
				localKropathDefaults.Tags,       // level 8
				globalApigwCfgDefaults.Tags,     // level 7
				localApigwCfgDefaults.Tags,      // level 6 (highest priority)
			),
			// SyncedLabels: additive union from ApiGatewayConfig defaults levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalApigwCfgDefaults.SyncedLabels,    // level 7
				localApigwCfgDefaults.SyncedLabels,     // level 6
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalApigwCfgDefaults.SyncedAnnotations,    // level 7
				localApigwCfgDefaults.SyncedAnnotations,     // level 6
			),
		},
	}
}
