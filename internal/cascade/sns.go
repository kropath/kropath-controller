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

// SNSKropathSection holds the SNS-family governance fields from
// KropathConfig.spec.mandatory.sns / .defaults.sns (ADR-015 §3.5).
//
// Only 3 fields are governed at the KropathConfig level: kmsMasterKeyId,
// signatureVersion, and tracingConfig. dataProtectionPolicy, deliveryFeedback,
// and namingTemplate are SNSConfig-only (family design §8).
//
// Zero value of each field is the permissive sentinel (not enforced).
type SNSKropathSection struct {
	// KmsMasterKeyId is the KMS key ID/ARN/alias to enforce for SNS topic encryption.
	KmsMasterKeyId string `json:"kmsMasterKeyId,omitempty"`

	// SignatureVersion is the SNS message signature version to enforce ("1" or "2").
	SignatureVersion string `json:"signatureVersion,omitempty"`

	// TracingConfig is the SNS tracing mode to enforce ("Active" or "PassThrough").
	TracingConfig string `json:"tracingConfig,omitempty"`

	// Tags are tier-level cloud resource tags from KropathConfig.spec.mandatory.tags
	// or .defaults.tags. Populated by the reconciler so that tag cascade flows
	// through MergeSNSCascade alongside the SNS-specific fields.
	Tags map[string]string `json:"tags,omitempty"`
}

// DeliveryFeedbackProtocol holds the delivery feedback settings for one SNS
// delivery protocol.
type DeliveryFeedbackProtocol struct {
	SuccessFeedbackRoleArn    string `json:"successFeedbackRoleArn,omitempty"`
	FailureFeedbackRoleArn    string `json:"failureFeedbackRoleArn,omitempty"`
	SuccessFeedbackSampleRate string `json:"successFeedbackSampleRate,omitempty"`
}

// DeliveryFeedback holds delivery feedback settings for all 5 SNS delivery
// protocols. Pointer fields allow omitempty to suppress absent protocol blocks
// in the serialized effectiveConfig.
//
// Per-key merge semantics: each of the 15 leaf fields (5 protocols × 3 fields)
// is resolved independently using first-non-empty-wins across cascade levels
// 3-4 (mandatory) or 6-7 (defaults). KropathConfig.sns has no deliveryFeedback
// counterpart (levels 1-2 and 8-9 do not participate).
type DeliveryFeedback struct {
	Application *DeliveryFeedbackProtocol `json:"application,omitempty"`
	HTTP        *DeliveryFeedbackProtocol `json:"http,omitempty"`
	Lambda      *DeliveryFeedbackProtocol `json:"lambda,omitempty"`
	SQS         *DeliveryFeedbackProtocol `json:"sqs,omitempty"`
	Firehose    *DeliveryFeedbackProtocol `json:"firehose,omitempty"`
}

// SNSConfigSection holds the SNS governance fields from SNSConfig.spec.mandatory
// or SNSConfig.spec.defaults (per-type ResourceConfig, ADR-015 §3.5).
type SNSConfigSection struct {
	KmsMasterKeyId   string `json:"kmsMasterKeyId,omitempty"`
	SignatureVersion string `json:"signatureVersion,omitempty"`
	TracingConfig    string `json:"tracingConfig,omitempty"`

	// DataProtectionPolicy is a JSON-encoded SNS data protection policy document.
	// Governed only at SNSConfig levels 3-4 (mandatory) and 6-7 (defaults).
	DataProtectionPolicy string `json:"dataProtectionPolicy,omitempty"`

	// DeliveryFeedback holds per-protocol delivery feedback settings.
	// Uses per-key merge semantics across SNSConfig levels only.
	DeliveryFeedback DeliveryFeedback `json:"deliveryFeedback,omitempty"`

	// NamingTemplate is the topic naming template (e.g. "{namespace}-{name}").
	// Governed only at SNSConfig levels 3-4 (mandatory) and 6-7 (defaults).
	NamingTemplate string `json:"namingTemplate,omitempty"`

	SyncedLabels      map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
	Tags              map[string]string `json:"tags,omitempty"`
}

// EffectiveSNSSection is one tier (mandatory or defaults) of the merged SNS
// governance result written into SNSConfig.status.effectiveConfig.
type EffectiveSNSSection struct {
	KmsMasterKeyId       string            `json:"kmsMasterKeyId,omitempty"`
	SignatureVersion     string            `json:"signatureVersion,omitempty"`
	TracingConfig        string            `json:"tracingConfig,omitempty"`
	DataProtectionPolicy string            `json:"dataProtectionPolicy,omitempty"`
	// DeliveryFeedback pointer so an entirely-empty result is omitted from the
	// serialized effectiveConfig rather than written as an empty object.
	DeliveryFeedback     *DeliveryFeedback `json:"deliveryFeedback,omitempty"`
	NamingTemplate       string            `json:"namingTemplate,omitempty"`
	SyncedLabels         map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations    map[string]string `json:"syncedAnnotations,omitempty"`
	Tags                 map[string]string `json:"tags,omitempty"`
}

// EffectiveSNSConfig is the merged SNS governance result written into
// SNSConfig.status.effectiveConfig by the controller.
type EffectiveSNSConfig struct {
	Mandatory EffectiveSNSSection `json:"mandatory"`
	Defaults  EffectiveSNSSection `json:"defaults"`
}

// MergeSNSCascade merges SNS governance fields from all cascade sources and
// returns the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for SNS (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.sns)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.sns)
//	Level 3 — globalSNSCfgMandatory   (SNSConfig in kro-system, mandatory)
//	Level 4 — localSNSCfgMandatory    (SNSConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localSNSCfgDefaults     (SNSConfig in resource namespace, defaults)
//	Level 7 — globalSNSCfgDefaults    (SNSConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.sns)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.sns)
//
// Scalar merge: firstNonEmptyString in priority order (lowest level number wins).
// deliveryFeedback: per-key merge across SNSConfig levels only (3-4 or 6-7).
// Tags: additive union across all four levels per tier.
// SyncedLabels/SyncedAnnotations: additive union from SNSConfig levels only.
func MergeSNSCascade(
	globalKropathMandatory SNSKropathSection, // level 1
	localKropathMandatory SNSKropathSection,  // level 2
	globalSNSCfgMandatory SNSConfigSection,   // level 3
	localSNSCfgMandatory SNSConfigSection,    // level 4
	localSNSCfgDefaults SNSConfigSection,     // level 6
	globalSNSCfgDefaults SNSConfigSection,    // level 7
	localKropathDefaults SNSKropathSection,   // level 8
	globalKropathDefaults SNSKropathSection,  // level 9
) EffectiveSNSConfig {
	return EffectiveSNSConfig{
		Mandatory: EffectiveSNSSection{
			KmsMasterKeyId: firstNonEmptyString(
				globalKropathMandatory.KmsMasterKeyId,
				localKropathMandatory.KmsMasterKeyId,
				globalSNSCfgMandatory.KmsMasterKeyId,
				localSNSCfgMandatory.KmsMasterKeyId,
			),
			SignatureVersion: firstNonEmptyString(
				globalKropathMandatory.SignatureVersion,
				localKropathMandatory.SignatureVersion,
				globalSNSCfgMandatory.SignatureVersion,
				localSNSCfgMandatory.SignatureVersion,
			),
			TracingConfig: firstNonEmptyString(
				globalKropathMandatory.TracingConfig,
				localKropathMandatory.TracingConfig,
				globalSNSCfgMandatory.TracingConfig,
				localSNSCfgMandatory.TracingConfig,
			),
			// dataProtectionPolicy, deliveryFeedback, namingTemplate:
			// SNSConfig levels only (3-4). KropathConfig.sns has no counterpart.
			DataProtectionPolicy: firstNonEmptyString(
				globalSNSCfgMandatory.DataProtectionPolicy,
				localSNSCfgMandatory.DataProtectionPolicy,
			),
			DeliveryFeedback: mergeDeliveryFeedback(
				globalSNSCfgMandatory.DeliveryFeedback,
				localSNSCfgMandatory.DeliveryFeedback,
			),
			NamingTemplate: firstNonEmptyString(
				globalSNSCfgMandatory.NamingTemplate,
				localSNSCfgMandatory.NamingTemplate,
			),
			// SyncedLabels: additive union from SNSConfig mandatory levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localSNSCfgMandatory.SyncedLabels,
				globalSNSCfgMandatory.SyncedLabels,
			),
			SyncedAnnotations: mergeMaps(
				localSNSCfgMandatory.SyncedAnnotations,
				globalSNSCfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localSNSCfgMandatory.Tags,
				globalSNSCfgMandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
		},
		Defaults: EffectiveSNSSection{
			KmsMasterKeyId: firstNonEmptyString(
				localSNSCfgDefaults.KmsMasterKeyId,
				globalSNSCfgDefaults.KmsMasterKeyId,
				localKropathDefaults.KmsMasterKeyId,
				globalKropathDefaults.KmsMasterKeyId,
			),
			SignatureVersion: firstNonEmptyString(
				localSNSCfgDefaults.SignatureVersion,
				globalSNSCfgDefaults.SignatureVersion,
				localKropathDefaults.SignatureVersion,
				globalKropathDefaults.SignatureVersion,
			),
			TracingConfig: firstNonEmptyString(
				localSNSCfgDefaults.TracingConfig,
				globalSNSCfgDefaults.TracingConfig,
				localKropathDefaults.TracingConfig,
				globalKropathDefaults.TracingConfig,
			),
			DataProtectionPolicy: firstNonEmptyString(
				localSNSCfgDefaults.DataProtectionPolicy,
				globalSNSCfgDefaults.DataProtectionPolicy,
			),
			DeliveryFeedback: mergeDeliveryFeedback(
				localSNSCfgDefaults.DeliveryFeedback,
				globalSNSCfgDefaults.DeliveryFeedback,
			),
			NamingTemplate: firstNonEmptyString(
				localSNSCfgDefaults.NamingTemplate,
				globalSNSCfgDefaults.NamingTemplate,
			),
			// SyncedLabels: additive union from SNSConfig defaults levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalSNSCfgDefaults.SyncedLabels,
				localSNSCfgDefaults.SyncedLabels,
			),
			SyncedAnnotations: mergeMaps(
				globalSNSCfgDefaults.SyncedAnnotations,
				localSNSCfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalSNSCfgDefaults.Tags,
				localSNSCfgDefaults.Tags,
			),
		},
	}
}

// mergeDeliveryFeedbackProtocol merges delivery feedback fields for one SNS
// protocol using per-key first-non-empty-wins across the provided sources.
// Returns nil if all three fields remain empty after merging.
func mergeDeliveryFeedbackProtocol(sources ...*DeliveryFeedbackProtocol) *DeliveryFeedbackProtocol {
	var successArn, failureArn, sampleRate string
	for _, s := range sources {
		if s == nil {
			continue
		}
		if successArn == "" {
			successArn = s.SuccessFeedbackRoleArn
		}
		if failureArn == "" {
			failureArn = s.FailureFeedbackRoleArn
		}
		if sampleRate == "" {
			sampleRate = s.SuccessFeedbackSampleRate
		}
	}
	if successArn == "" && failureArn == "" && sampleRate == "" {
		return nil
	}
	return &DeliveryFeedbackProtocol{
		SuccessFeedbackRoleArn:    successArn,
		FailureFeedbackRoleArn:    failureArn,
		SuccessFeedbackSampleRate: sampleRate,
	}
}

// mergeDeliveryFeedback merges delivery feedback from multiple SNSConfigSection
// sources. Each of the 15 leaf fields (5 protocols × 3 fields) is resolved
// independently using first-non-empty-wins across the provided sources.
// Returns nil if all protocol blocks remain entirely empty after merging.
func mergeDeliveryFeedback(sources ...DeliveryFeedback) *DeliveryFeedback {
	apps := make([]*DeliveryFeedbackProtocol, len(sources))
	https := make([]*DeliveryFeedbackProtocol, len(sources))
	lambdas := make([]*DeliveryFeedbackProtocol, len(sources))
	sqss := make([]*DeliveryFeedbackProtocol, len(sources))
	firehoses := make([]*DeliveryFeedbackProtocol, len(sources))
	for i, s := range sources {
		apps[i] = s.Application
		https[i] = s.HTTP
		lambdas[i] = s.Lambda
		sqss[i] = s.SQS
		firehoses[i] = s.Firehose
	}
	app := mergeDeliveryFeedbackProtocol(apps...)
	http := mergeDeliveryFeedbackProtocol(https...)
	lambda := mergeDeliveryFeedbackProtocol(lambdas...)
	sqs := mergeDeliveryFeedbackProtocol(sqss...)
	firehose := mergeDeliveryFeedbackProtocol(firehoses...)
	if app == nil && http == nil && lambda == nil && sqs == nil && firehose == nil {
		return nil
	}
	return &DeliveryFeedback{
		Application: app,
		HTTP:        http,
		Lambda:      lambda,
		SQS:         sqs,
		Firehose:    firehose,
	}
}
