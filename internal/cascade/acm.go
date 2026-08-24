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

// ACMKropathSection holds the Certificate Manager family governance fields from
// KropathConfig.spec.mandatory.certificateManager / .defaults.certificateManager
// (ADR-015 §3.5), PLUS the tier-level tags from KropathConfig.spec.mandatory.tags /
// .defaults.tags.
//
// Only keyAlgorithm and certificateTransparencyLogging are promoted to KropathConfig
// for the Certificate Manager family. usageMode and keyStorageSecurityStandard are
// per-CA choices that belong in ACMConfig or instance spec (family design §8).
//
// Tags are org-wide tier-level fields populated by the reconciler so the full tag
// cascade flows through MergeACMCascade.
//
// Zero value of each field is the permissive sentinel (not enforced).
type ACMKropathSection struct {
	// KeyAlgorithm is the org-wide enforced key algorithm for certificates and CAs.
	// Empty string = not enforced. e.g. "EC_prime256v1", "RSA_2048".
	KeyAlgorithm string `json:"keyAlgorithm,omitempty"`

	// CertificateTransparencyLogging is the org-wide enforced CT logging preference.
	// Empty string = not enforced. e.g. "ENABLED", "DISABLED".
	CertificateTransparencyLogging string `json:"certificateTransparencyLogging,omitempty"`

	// Tags are tier-level cloud resource tags from KropathConfig.spec.mandatory.tags
	// or .defaults.tags. Populated by the reconciler.
	Tags map[string]string `json:"tags,omitempty"`
}

// ACMConfigSection holds the Certificate Manager governance fields from
// ACMConfig.spec.mandatory or ACMConfig.spec.defaults (per-type ResourceConfig,
// ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type ACMConfigSection struct {
	// KeyAlgorithm enforces the key algorithm. Empty = not enforced.
	// e.g. "RSA_2048", "EC_prime256v1", "EC_secp384r1".
	KeyAlgorithm string `json:"keyAlgorithm,omitempty"`

	// CertificateTransparencyLogging enforces CT logging preference.
	// Empty = not enforced. e.g. "ENABLED", "DISABLED".
	CertificateTransparencyLogging string `json:"certificateTransparencyLogging,omitempty"`

	// UsageMode enforces the private CA usage mode.
	// Empty = not enforced. e.g. "GENERAL_PURPOSE", "SHORT_LIVED_CERTIFICATE".
	// ACMConfig-only (no KropathConfig equivalent).
	UsageMode string `json:"usageMode,omitempty"`

	// KeyStorageSecurityStandard enforces the FIPS level for private CAs.
	// Empty = not enforced.
	// e.g. "FIPS_140_2_LEVEL_2_OR_HIGHER", "FIPS_140_2_LEVEL_3_OR_HIGHER".
	// ACMConfig-only (no KropathConfig equivalent).
	KeyStorageSecurityStandard string `json:"keyStorageSecurityStandard,omitempty"`

	// Tags are cloud resource tags for this Certificate Manager config profile.
	Tags map[string]string `json:"tags,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to Certificate Manager resources.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to Certificate Manager resources.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveACMSection is one tier (mandatory or defaults) of the merged Certificate
// Manager governance result written into ACMConfig.status.effectiveConfig by the controller.
type EffectiveACMSection struct {
	KeyAlgorithm                   string            `json:"keyAlgorithm,omitempty"`
	CertificateTransparencyLogging string            `json:"certificateTransparencyLogging,omitempty"`
	UsageMode                      string            `json:"usageMode,omitempty"`
	KeyStorageSecurityStandard     string            `json:"keyStorageSecurityStandard,omitempty"`
	Tags                           map[string]string `json:"tags,omitempty"`
	SyncedLabels                   map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations              map[string]string `json:"syncedAnnotations,omitempty"`
}

// EffectiveACMConfig is the merged Certificate Manager governance result written into
// ACMConfig.status.effectiveConfig by the controller.
type EffectiveACMConfig struct {
	Mandatory EffectiveACMSection `json:"mandatory"`
	Defaults  EffectiveACMSection `json:"defaults"`
}

// MergeACMCascade merges Certificate Manager governance fields from all cascade sources
// and returns the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for Certificate Manager (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.certificateManager)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.certificateManager)
//	Level 3 — globalACMCfgMandatory   (ACMConfig in kro-system, mandatory)
//	Level 4 — localACMCfgMandatory    (ACMConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localACMCfgDefaults     (ACMConfig in resource namespace, defaults)
//	Level 7 — globalACMCfgDefaults    (ACMConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.certificateManager)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.certificateManager)
//
// KropathConfig fields (keyAlgorithm, certificateTransparencyLogging):
//   - Mandatory: all four sources (L1 wins).
//   - Defaults: all four sources (L6 wins).
//
// ACMConfig-only fields (usageMode, keyStorageSecurityStandard):
//   - Mandatory: L3 and L4 only.
//   - Defaults: L6 and L7 only.
//
// Tags: additive union across all four mandatory sources (L1 wins on key conflict) and all
// four defaults sources (L6 wins on key conflict). ACMKropathSection.Tags carries the
// tier-level KropathConfig.mandatory.tags (populated by the reconciler).
//
// SyncedLabels/SyncedAnnotations: additive union from ACMConfig levels only
// (mandatory: L3 wins; defaults: L6 wins).
func MergeACMCascade(
	// Mandatory inputs (highest → lowest priority)
	globalKropathMandatory ACMKropathSection,  // level 1
	localKropathMandatory ACMKropathSection,   // level 2
	globalACMCfgMandatory ACMConfigSection,    // level 3
	localACMCfgMandatory ACMConfigSection,     // level 4
	// Defaults inputs (highest → lowest priority)
	localACMCfgDefaults ACMConfigSection,      // level 6
	globalACMCfgDefaults ACMConfigSection,     // level 7
	localKropathDefaults ACMKropathSection,    // level 8
	globalKropathDefaults ACMKropathSection,   // level 9
) EffectiveACMConfig {
	return EffectiveACMConfig{
		Mandatory: EffectiveACMSection{
			// KropathConfig fields: all four mandatory sources (L1 wins).
			KeyAlgorithm: firstNonEmptyString(
				globalKropathMandatory.KeyAlgorithm,       // level 1
				localKropathMandatory.KeyAlgorithm,        // level 2
				globalACMCfgMandatory.KeyAlgorithm,        // level 3
				localACMCfgMandatory.KeyAlgorithm,         // level 4
			),
			CertificateTransparencyLogging: firstNonEmptyString(
				globalKropathMandatory.CertificateTransparencyLogging, // level 1
				localKropathMandatory.CertificateTransparencyLogging,  // level 2
				globalACMCfgMandatory.CertificateTransparencyLogging,  // level 3
				localACMCfgMandatory.CertificateTransparencyLogging,   // level 4
			),
			// ACMConfig-only fields: L3 and L4 only (no KropathConfig equivalent).
			UsageMode: firstNonEmptyString(
				globalACMCfgMandatory.UsageMode,  // level 3
				localACMCfgMandatory.UsageMode,   // level 4
			),
			KeyStorageSecurityStandard: firstNonEmptyString(
				globalACMCfgMandatory.KeyStorageSecurityStandard,  // level 3
				localACMCfgMandatory.KeyStorageSecurityStandard,   // level 4
			),
			// Tags: union of all mandatory sources; L4 added first (lowest priority), L1 wins on key conflict.
			Tags: mergeMaps(
				localACMCfgMandatory.Tags,    // level 4
				globalACMCfgMandatory.Tags,   // level 3
				localKropathMandatory.Tags,   // level 2
				globalKropathMandatory.Tags,  // level 1 (highest priority)
			),
			// SyncedLabels: additive union from ACMConfig mandatory levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localACMCfgMandatory.SyncedLabels,   // level 4
				globalACMCfgMandatory.SyncedLabels,  // level 3
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localACMCfgMandatory.SyncedAnnotations,   // level 4
				globalACMCfgMandatory.SyncedAnnotations,  // level 3
			),
		},
		Defaults: EffectiveACMSection{
			// KropathConfig fields: all four defaults sources (L6 wins).
			KeyAlgorithm: firstNonEmptyString(
				localACMCfgDefaults.KeyAlgorithm,       // level 6
				globalACMCfgDefaults.KeyAlgorithm,      // level 7
				localKropathDefaults.KeyAlgorithm,      // level 8
				globalKropathDefaults.KeyAlgorithm,     // level 9
			),
			CertificateTransparencyLogging: firstNonEmptyString(
				localACMCfgDefaults.CertificateTransparencyLogging,   // level 6
				globalACMCfgDefaults.CertificateTransparencyLogging,  // level 7
				localKropathDefaults.CertificateTransparencyLogging,  // level 8
				globalKropathDefaults.CertificateTransparencyLogging, // level 9
			),
			// ACMConfig-only fields: L6 and L7 only.
			UsageMode: firstNonEmptyString(
				localACMCfgDefaults.UsageMode,   // level 6
				globalACMCfgDefaults.UsageMode,  // level 7
			),
			KeyStorageSecurityStandard: firstNonEmptyString(
				localACMCfgDefaults.KeyStorageSecurityStandard,   // level 6
				globalACMCfgDefaults.KeyStorageSecurityStandard,  // level 7
			),
			// Tags: union of all defaults sources; L9 added first (lowest priority), L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,    // level 9
				localKropathDefaults.Tags,     // level 8
				globalACMCfgDefaults.Tags,     // level 7
				localACMCfgDefaults.Tags,      // level 6 (highest priority)
			),
			// SyncedLabels: additive union from ACMConfig defaults levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalACMCfgDefaults.SyncedLabels,   // level 7
				localACMCfgDefaults.SyncedLabels,    // level 6
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalACMCfgDefaults.SyncedAnnotations,   // level 7
				localACMCfgDefaults.SyncedAnnotations,    // level 6
			),
		},
	}
}
