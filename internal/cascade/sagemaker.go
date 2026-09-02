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

// SageMakerKropathSection holds the SageMaker governance fields from
// KropathConfig.spec.mandatory.sagemaker / .defaults.sagemaker (ADR-015 §3.5).
//
// Only 4 fields are present here per OD-5 of the family design:
// instanceType, kmsKeyId, enableNetworkIsolation, enableInterContainerTrafficEncryption.
// Fields volumeSizeInGB, rootAccess, directInternetAccess, namingTemplate live only in
// SageMakerConfig (SageMakerConfigSection).
//
// Zero value of each field is the permissive sentinel (not enforced):
// "" for strings, false for booleans.
type SageMakerKropathSection struct {
	// InstanceType is the org-wide SageMaker instance type mandate.
	// Empty string = not enforced.
	InstanceType string `json:"instanceType,omitempty"`

	// KmsKeyId is the KMS key ARN for SageMaker storage encryption.
	// Empty string = not enforced.
	KmsKeyId string `json:"kmsKeyId,omitempty"`

	// EnableNetworkIsolation forces network isolation on SageMaker resources.
	// false (zero value) = not enforced; true = org-wide mandate.
	EnableNetworkIsolation bool `json:"enableNetworkIsolation,omitempty"`

	// EnableInterContainerTrafficEncryption forces inter-container traffic encryption.
	// false (zero value) = not enforced; true = org-wide mandate.
	EnableInterContainerTrafficEncryption bool `json:"enableInterContainerTrafficEncryption,omitempty"`

	// Tags are tier-level cloud resource tags from KropathConfig.spec.mandatory.tags
	// or .defaults.tags. Populated by the reconciler from the tier-level field so
	// tag cascade flows through MergeSageMakerCascade alongside SageMaker-specific fields.
	// nil / empty map = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// SageMakerConfigSection holds the SageMaker governance fields from
// SageMakerConfig.spec.mandatory or SageMakerConfig.spec.defaults (ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type SageMakerConfigSection struct {
	// InstanceType is the SageMaker instance type.
	// Empty = not enforced (mandatory) or no default (defaults).
	InstanceType string `json:"instanceType,omitempty"`

	// VolumeSize is the EBS volume size in GB for SageMaker resources.
	// 0 = not enforced (mandatory) or default 5 GB (defaults).
	VolumeSize int64 `json:"volumeSizeInGB,omitempty"`

	// KmsKeyId is the KMS key ARN for SageMaker storage encryption.
	// Empty = not enforced.
	KmsKeyId string `json:"kmsKeyId,omitempty"`

	// EnableNetworkIsolation forces network isolation.
	// false (zero value) = not enforced.
	EnableNetworkIsolation bool `json:"enableNetworkIsolation,omitempty"`

	// EnableInterContainerTrafficEncryption forces inter-container traffic encryption.
	// false (zero value) = not enforced.
	EnableInterContainerTrafficEncryption bool `json:"enableInterContainerTrafficEncryption,omitempty"`

	// RootAccess controls notebook root access. "Enabled" | "Disabled" | "" (not enforced).
	RootAccess string `json:"rootAccess,omitempty"`

	// DirectInternetAccess controls notebook internet access.
	// "Enabled" | "Disabled" | "" (not enforced).
	DirectInternetAccess string `json:"directInternetAccess,omitempty"`

	// NamingTemplate is the SageMaker resource naming template.
	// Empty = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// Tags are cloud resource tags for this SageMakerConfig profile.
	// nil / empty = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to SageMaker resource CRs.
	// Additive map merge across SageMakerConfig tiers only.
	// nil / empty = no labels at this level.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to SageMaker resource CRs.
	// Additive map merge across SageMakerConfig tiers only.
	// nil / empty = no annotations at this level.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveSageMakerSection is one tier (mandatory or defaults) of the merged SageMaker
// governance result written into SageMakerConfig.status.effectiveConfig by the controller.
type EffectiveSageMakerSection struct {
	InstanceType                          string            `json:"instanceType,omitempty"`
	VolumeSize                            int64             `json:"volumeSizeInGB,omitempty"`
	KmsKeyId                              string            `json:"kmsKeyId,omitempty"`
	EnableNetworkIsolation                bool              `json:"enableNetworkIsolation,omitempty"`
	EnableInterContainerTrafficEncryption bool              `json:"enableInterContainerTrafficEncryption,omitempty"`
	RootAccess                            string            `json:"rootAccess,omitempty"`
	DirectInternetAccess                  string            `json:"directInternetAccess,omitempty"`
	NamingTemplate                        string            `json:"namingTemplate,omitempty"`
	Tags                                  map[string]string `json:"tags,omitempty"`
	SyncedLabels                          map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations                     map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveSageMakerConfig is the merged SageMaker governance result written into
// SageMakerConfig.status.effectiveConfig by the controller.
type EffectiveSageMakerConfig struct {
	Mandatory EffectiveSageMakerSection `json:"mandatory"`
	Defaults  EffectiveSageMakerSection `json:"defaults"`
}

// MergeSageMakerCascade merges SageMaker governance fields from all cascade sources and
// returns the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for SageMaker (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.sagemaker)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.sagemaker)
//	Level 3 — globalSMCfgMandatory    (SageMakerConfig in kro-system, mandatory)
//	Level 4 — localSMCfgMandatory     (SageMakerConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localSMCfgDefaults      (SageMakerConfig in resource namespace, defaults)
//	Level 7 — globalSMCfgDefaults     (SageMakerConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.sagemaker)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.sagemaker)
//
// Scalar merge rules:
//   - String fields: firstNonEmptyString in priority order.
//   - Boolean fields: firstTrue in priority order (false = zero/not enforced).
//   - Integer fields (VolumeSize): firstNonZeroInt64 in priority order.
//
// Map merge rules:
//   - Tags: additive union across all four mandatory (or defaults) levels; lower index
//     (higher priority) wins on key conflict.
//   - SyncedLabels/SyncedAnnotations: additive union from SageMakerConfig levels only
//     (not present in KropathConfig.sagemaker per OD-5).
//
// Fields only in SageMakerConfig (not KropathConfig.sagemaker):
// VolumeSize, RootAccess, DirectInternetAccess, NamingTemplate, SyncedLabels, SyncedAnnotations.
func MergeSageMakerCascade(
	globalKropathMandatory SageMakerKropathSection, // level 1
	localKropathMandatory SageMakerKropathSection, // level 2
	globalSMCfgMandatory SageMakerConfigSection, // level 3
	localSMCfgMandatory SageMakerConfigSection, // level 4
	localSMCfgDefaults SageMakerConfigSection, // level 6
	globalSMCfgDefaults SageMakerConfigSection, // level 7
	localKropathDefaults SageMakerKropathSection, // level 8
	globalKropathDefaults SageMakerKropathSection, // level 9
) EffectiveSageMakerConfig {
	return EffectiveSageMakerConfig{
		Mandatory: EffectiveSageMakerSection{
			// instanceType: levels 1, 2, 3, 4 (KropathConfig wins over SageMakerConfig).
			InstanceType: firstNonEmptyString(
				globalKropathMandatory.InstanceType, // level 1
				localKropathMandatory.InstanceType,  // level 2
				globalSMCfgMandatory.InstanceType,   // level 3
				localSMCfgMandatory.InstanceType,    // level 4
			),
			// volumeSizeInGB: levels 3, 4 only (not in KropathConfig.sagemaker).
			VolumeSize: firstNonZeroInt64(
				globalSMCfgMandatory.VolumeSize, // level 3
				localSMCfgMandatory.VolumeSize,  // level 4
			),
			// kmsKeyId: levels 1, 2, 3, 4 (KropathConfig wins over SageMakerConfig).
			KmsKeyId: firstNonEmptyString(
				globalKropathMandatory.KmsKeyId, // level 1
				localKropathMandatory.KmsKeyId,  // level 2
				globalSMCfgMandatory.KmsKeyId,   // level 3
				localSMCfgMandatory.KmsKeyId,    // level 4
			),
			// enableNetworkIsolation: levels 1, 2, 3, 4.
			EnableNetworkIsolation: firstTrue(
				globalKropathMandatory.EnableNetworkIsolation, // level 1
				localKropathMandatory.EnableNetworkIsolation,  // level 2
				globalSMCfgMandatory.EnableNetworkIsolation,   // level 3
				localSMCfgMandatory.EnableNetworkIsolation,    // level 4
			),
			// enableInterContainerTrafficEncryption: levels 1, 2, 3, 4.
			EnableInterContainerTrafficEncryption: firstTrue(
				globalKropathMandatory.EnableInterContainerTrafficEncryption, // level 1
				localKropathMandatory.EnableInterContainerTrafficEncryption,  // level 2
				globalSMCfgMandatory.EnableInterContainerTrafficEncryption,   // level 3
				localSMCfgMandatory.EnableInterContainerTrafficEncryption,    // level 4
			),
			// rootAccess: levels 3, 4 only (not in KropathConfig.sagemaker).
			RootAccess: firstNonEmptyString(
				globalSMCfgMandatory.RootAccess, // level 3
				localSMCfgMandatory.RootAccess,  // level 4
			),
			// directInternetAccess: levels 3, 4 only (not in KropathConfig.sagemaker).
			DirectInternetAccess: firstNonEmptyString(
				globalSMCfgMandatory.DirectInternetAccess, // level 3
				localSMCfgMandatory.DirectInternetAccess,  // level 4
			),
			// namingTemplate: levels 3, 4 only (not in KropathConfig.sagemaker).
			NamingTemplate: firstNonEmptyString(
				globalSMCfgMandatory.NamingTemplate, // level 3
				localSMCfgMandatory.NamingTemplate,  // level 4
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localSMCfgMandatory.Tags,            // level 4 (lowest priority)
				globalSMCfgMandatory.Tags,           // level 3
				localKropathMandatory.Tags,          // level 2
				globalKropathMandatory.Tags,         // level 1 (highest priority)
			),
			// SyncedLabels: additive union from SageMakerConfig levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localSMCfgMandatory.SyncedLabels,
				globalSMCfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localSMCfgMandatory.SyncedAnnotations,
				globalSMCfgMandatory.SyncedAnnotations,
			),
		},
		Defaults: EffectiveSageMakerSection{
			// instanceType: levels 6, 7, 8, 9.
			InstanceType: firstNonEmptyString(
				localSMCfgDefaults.InstanceType,    // level 6
				globalSMCfgDefaults.InstanceType,   // level 7
				localKropathDefaults.InstanceType,  // level 8
				globalKropathDefaults.InstanceType, // level 9
			),
			// volumeSizeInGB: levels 6, 7 only (not in KropathConfig.sagemaker).
			VolumeSize: firstNonZeroInt64(
				localSMCfgDefaults.VolumeSize,   // level 6
				globalSMCfgDefaults.VolumeSize,  // level 7
			),
			// kmsKeyId: levels 6, 7, 8, 9.
			KmsKeyId: firstNonEmptyString(
				localSMCfgDefaults.KmsKeyId,    // level 6
				globalSMCfgDefaults.KmsKeyId,   // level 7
				localKropathDefaults.KmsKeyId,  // level 8
				globalKropathDefaults.KmsKeyId, // level 9
			),
			// enableNetworkIsolation: levels 6, 7, 8, 9.
			EnableNetworkIsolation: firstTrue(
				localSMCfgDefaults.EnableNetworkIsolation,    // level 6
				globalSMCfgDefaults.EnableNetworkIsolation,   // level 7
				localKropathDefaults.EnableNetworkIsolation,  // level 8
				globalKropathDefaults.EnableNetworkIsolation, // level 9
			),
			// enableInterContainerTrafficEncryption: levels 6, 7, 8, 9.
			EnableInterContainerTrafficEncryption: firstTrue(
				localSMCfgDefaults.EnableInterContainerTrafficEncryption,    // level 6
				globalSMCfgDefaults.EnableInterContainerTrafficEncryption,   // level 7
				localKropathDefaults.EnableInterContainerTrafficEncryption,  // level 8
				globalKropathDefaults.EnableInterContainerTrafficEncryption, // level 9
			),
			// rootAccess: levels 6, 7 only (not in KropathConfig.sagemaker).
			RootAccess: firstNonEmptyString(
				localSMCfgDefaults.RootAccess,   // level 6
				globalSMCfgDefaults.RootAccess,  // level 7
			),
			// directInternetAccess: levels 6, 7 only (not in KropathConfig.sagemaker).
			DirectInternetAccess: firstNonEmptyString(
				localSMCfgDefaults.DirectInternetAccess,   // level 6
				globalSMCfgDefaults.DirectInternetAccess,  // level 7
			),
			// namingTemplate: levels 6, 7 only (not in KropathConfig.sagemaker).
			NamingTemplate: firstNonEmptyString(
				localSMCfgDefaults.NamingTemplate,   // level 6
				globalSMCfgDefaults.NamingTemplate,  // level 7
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,  // level 9 (lowest priority)
				localKropathDefaults.Tags,   // level 8
				globalSMCfgDefaults.Tags,    // level 7
				localSMCfgDefaults.Tags,     // level 6 (highest priority)
			),
			// SyncedLabels: additive union from SageMakerConfig levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalSMCfgDefaults.SyncedLabels,
				localSMCfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalSMCfgDefaults.SyncedAnnotations,
				localSMCfgDefaults.SyncedAnnotations,
			),
		},
	}
}
