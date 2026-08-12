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

import (
	"testing"
)

// zeroECRKropath is a convenience zero-value ECRKropathSection for table tests.
var zeroECRKropath = ECRKropathSection{}

// zeroECRConfig is a convenience zero-value ECRConfigSection for table tests.
var zeroECRConfig = ECRConfigSection{}

// mergeAllZero calls MergeECRCascade with all zero inputs.
func mergeECRAllZero() EffectiveECRConfig {
	return MergeECRCascade(
		zeroECRKropath, zeroECRKropath,
		zeroECRConfig, zeroECRConfig,
		zeroECRConfig, zeroECRConfig,
		zeroECRKropath, zeroECRKropath,
	)
}

// TestMergeECRCascade_ImageTagMutability covers AC-1 through AC-4.
func TestMergeECRCascade_ImageTagMutability(t *testing.T) {
	tests := []struct {
		name                       string
		globalKropathMandatory     ECRKropathSection
		localKropathMandatory      ECRKropathSection
		globalECRCfgMandatory      ECRConfigSection
		localECRCfgMandatory       ECRConfigSection
		localECRCfgDefaults        ECRConfigSection
		globalECRCfgDefaults       ECRConfigSection
		localKropathDefaults       ECRKropathSection
		globalKropathDefaults      ECRKropathSection
		wantMandatoryMutability    string
		wantDefaultsMutability     string
	}{
		{
			// AC-1: global KropathConfig.mandatory.ecr.imageTagMutability (L1) propagates.
			name:                    "AC-1: global-kpc-mandatory-L1-propagates",
			globalKropathMandatory:  ECRKropathSection{ImageTagMutability: "IMMUTABLE"},
			wantMandatoryMutability: "IMMUTABLE",
		},
		{
			// AC-2: KropathConfig.mandatory (L1) wins over ECRConfig.mandatory (L3).
			name:                    "AC-2: kpc-L1-wins-over-ecrcfg-L3",
			globalKropathMandatory:  ECRKropathSection{ImageTagMutability: "IMMUTABLE"},
			globalECRCfgMandatory:   ECRConfigSection{ImageTagMutability: "MUTABLE"},
			wantMandatoryMutability: "IMMUTABLE",
		},
		{
			// AC-3: Only ECRConfig.defaults.imageTagMutability set; mandatory stays empty.
			name:                   "AC-3: ecrcfg-defaults-L7-propagates",
			globalECRCfgDefaults:   ECRConfigSection{ImageTagMutability: "MUTABLE"},
			wantMandatoryMutability: "",
			wantDefaultsMutability:  "MUTABLE",
		},
		{
			// AC-4: global ECRConfig.mandatory (L3) wins over local ECRConfig.mandatory (L4).
			name:                    "AC-4: global-ecrcfg-mandatory-L3-wins-local-L4",
			globalECRCfgMandatory:   ECRConfigSection{ImageTagMutability: "MUTABLE"},
			localECRCfgMandatory:    ECRConfigSection{ImageTagMutability: "IMMUTABLE"},
			wantMandatoryMutability: "MUTABLE",
		},
		{
			// L2 KropathConfig.mandatory wins over L3 ECRConfig.mandatory.
			name:                    "local-kpc-mandatory-L2-wins-over-ecrcfg-L3",
			localKropathMandatory:   ECRKropathSection{ImageTagMutability: "IMMUTABLE"},
			globalECRCfgMandatory:   ECRConfigSection{ImageTagMutability: "MUTABLE"},
			wantMandatoryMutability: "IMMUTABLE",
		},
		{
			// L9 globalKropathDefaults propagates when all higher defaults levels empty.
			name:                   "global-kpc-defaults-L9-propagates",
			globalKropathDefaults:  ECRKropathSection{ImageTagMutability: "MUTABLE"},
			wantDefaultsMutability: "MUTABLE",
		},
		{
			// L6 local ECRCfgDefaults wins over L7 global.
			name:                   "local-ecrcfg-defaults-L6-wins-global-L7",
			localECRCfgDefaults:    ECRConfigSection{ImageTagMutability: "IMMUTABLE"},
			globalECRCfgDefaults:   ECRConfigSection{ImageTagMutability: "MUTABLE"},
			wantDefaultsMutability: "IMMUTABLE",
		},
		{
			// L8 local KropathDefaults wins over L9 global.
			name:                   "local-kpc-defaults-L8-wins-global-L9",
			localKropathDefaults:   ECRKropathSection{ImageTagMutability: "IMMUTABLE"},
			globalKropathDefaults:  ECRKropathSection{ImageTagMutability: "MUTABLE"},
			wantDefaultsMutability: "IMMUTABLE",
		},
		{
			// All zero — zero value returned.
			name:                    "all-zero-returns-empty",
			wantMandatoryMutability: "",
			wantDefaultsMutability:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MergeECRCascade(
				tc.globalKropathMandatory,
				tc.localKropathMandatory,
				tc.globalECRCfgMandatory,
				tc.localECRCfgMandatory,
				tc.localECRCfgDefaults,
				tc.globalECRCfgDefaults,
				tc.localKropathDefaults,
				tc.globalKropathDefaults,
			)
			if got.Mandatory.ImageTagMutability != tc.wantMandatoryMutability {
				t.Errorf("Mandatory.ImageTagMutability = %q, want %q", got.Mandatory.ImageTagMutability, tc.wantMandatoryMutability)
			}
			if got.Defaults.ImageTagMutability != tc.wantDefaultsMutability {
				t.Errorf("Defaults.ImageTagMutability = %q, want %q", got.Defaults.ImageTagMutability, tc.wantDefaultsMutability)
			}
		})
	}
}

// TestMergeECRCascade_EncryptionType covers AC-5 through AC-6.
func TestMergeECRCascade_EncryptionType(t *testing.T) {
	tests := []struct {
		name                    string
		globalKropathMandatory  ECRKropathSection
		globalECRCfgDefaults    ECRConfigSection
		wantMandatoryEncType    string
		wantDefaultsEncType     string
	}{
		{
			// AC-5: global KropathConfig.mandatory.ecr.encryptionType (L1) propagates.
			name:                   "AC-5: global-kpc-mandatory-L1-encryptionType",
			globalKropathMandatory: ECRKropathSection{EncryptionType: "KMS"},
			wantMandatoryEncType:   "KMS",
		},
		{
			// AC-6: Only ECRConfig.defaults.encryptionType set.
			name:                 "AC-6: global-ecrcfg-defaults-L7-encryptionType",
			globalECRCfgDefaults: ECRConfigSection{EncryptionType: "AES256"},
			wantDefaultsEncType:  "AES256",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MergeECRCascade(
				tc.globalKropathMandatory, zeroECRKropath,
				zeroECRConfig, zeroECRConfig,
				zeroECRConfig, tc.globalECRCfgDefaults,
				zeroECRKropath, zeroECRKropath,
			)
			if got.Mandatory.EncryptionType != tc.wantMandatoryEncType {
				t.Errorf("Mandatory.EncryptionType = %q, want %q", got.Mandatory.EncryptionType, tc.wantMandatoryEncType)
			}
			if got.Defaults.EncryptionType != tc.wantDefaultsEncType {
				t.Errorf("Defaults.EncryptionType = %q, want %q", got.Defaults.EncryptionType, tc.wantDefaultsEncType)
			}
		})
	}
}

// TestMergeECRCascade_KmsKeyID covers AC-7 through AC-8.
func TestMergeECRCascade_KmsKeyID(t *testing.T) {
	const (
		kmsGlobal = "arn:aws:kms:ap-southeast-2:123456789012:key/mrk-profile"
		kmsAlias  = "alias/default-key"
		kmsLocal  = "arn:aws:kms:ap-southeast-2:123456789012:key/mrk-local"
	)

	tests := []struct {
		name                  string
		globalECRCfgMandatory ECRConfigSection
		localECRCfgMandatory  ECRConfigSection
		localECRCfgDefaults   ECRConfigSection
		globalECRCfgDefaults  ECRConfigSection
		wantMandatoryKmsKeyID string
		wantDefaultsKmsKeyID  string
	}{
		{
			// AC-7: global ECRConfig.mandatory.kmsKeyID (L3) propagates; no KropathConfig entry.
			name:                  "AC-7: global-ecrcfg-mandatory-L3-kmsKeyID",
			globalECRCfgMandatory: ECRConfigSection{KmsKeyID: kmsGlobal},
			wantMandatoryKmsKeyID: kmsGlobal,
		},
		{
			// AC-8: Only ECRConfig.defaults.kmsKeyID set; mandatory stays empty.
			name:                 "AC-8: global-ecrcfg-defaults-L7-kmsKeyID",
			globalECRCfgDefaults: ECRConfigSection{KmsKeyID: kmsAlias},
			wantMandatoryKmsKeyID: "",
			wantDefaultsKmsKeyID:  kmsAlias,
		},
		{
			// L3 global wins over L4 local.
			name:                  "global-ecrcfg-mandatory-L3-wins-local-L4",
			globalECRCfgMandatory: ECRConfigSection{KmsKeyID: kmsGlobal},
			localECRCfgMandatory:  ECRConfigSection{KmsKeyID: kmsLocal},
			wantMandatoryKmsKeyID: kmsGlobal,
		},
		{
			// L6 local wins over L7 global defaults.
			name:                 "local-ecrcfg-defaults-L6-wins-global-L7",
			localECRCfgDefaults:  ECRConfigSection{KmsKeyID: kmsLocal},
			globalECRCfgDefaults: ECRConfigSection{KmsKeyID: kmsGlobal},
			wantDefaultsKmsKeyID: kmsLocal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MergeECRCascade(
				zeroECRKropath, zeroECRKropath,
				tc.globalECRCfgMandatory, tc.localECRCfgMandatory,
				tc.localECRCfgDefaults, tc.globalECRCfgDefaults,
				zeroECRKropath, zeroECRKropath,
			)
			if got.Mandatory.KmsKeyID != tc.wantMandatoryKmsKeyID {
				t.Errorf("Mandatory.KmsKeyID = %q, want %q", got.Mandatory.KmsKeyID, tc.wantMandatoryKmsKeyID)
			}
			if got.Defaults.KmsKeyID != tc.wantDefaultsKmsKeyID {
				t.Errorf("Defaults.KmsKeyID = %q, want %q", got.Defaults.KmsKeyID, tc.wantDefaultsKmsKeyID)
			}
		})
	}
}

// TestMergeECRCascade_LifecyclePolicy covers AC-9 through AC-11.
func TestMergeECRCascade_LifecyclePolicy(t *testing.T) {
	const (
		policyKropath  = `{"rules":[{"rulePriority":1,"description":"kc","selection":{"tagStatus":"any","countType":"imageCountMoreThan","countNumber":10},"action":{"type":"expire"}}]}`
		policyECRCfg   = `{"rules":[{"rulePriority":2,"description":"ecrcfg","selection":{"tagStatus":"any","countType":"imageCountMoreThan","countNumber":20},"action":{"type":"expire"}}]}`
		policyDefaults = `{"rules":[{"rulePriority":3,"description":"defaults","selection":{"tagStatus":"any","countType":"imageCountMoreThan","countNumber":30},"action":{"type":"expire"}}]}`
	)

	tests := []struct {
		name                       string
		globalKropathMandatory     ECRKropathSection
		globalECRCfgMandatory      ECRConfigSection
		globalECRCfgDefaults       ECRConfigSection
		wantMandatoryLifecycle     string
		wantDefaultsLifecycle      string
	}{
		{
			// AC-9: global KropathConfig.mandatory.ecr.lifecyclePolicy (L1) propagates.
			name:                   "AC-9: global-kpc-mandatory-L1-lifecyclePolicy",
			globalKropathMandatory: ECRKropathSection{LifecyclePolicy: policyKropath},
			wantMandatoryLifecycle: policyKropath,
		},
		{
			// AC-10: Only ECRConfig.defaults.lifecyclePolicy (L7) set.
			name:                  "AC-10: global-ecrcfg-defaults-L7-lifecyclePolicy",
			globalECRCfgDefaults:  ECRConfigSection{LifecyclePolicy: policyDefaults},
			wantDefaultsLifecycle: policyDefaults,
		},
		{
			// AC-11: KropathConfig.mandatory (L1) wins over ECRConfig.mandatory (L3).
			name:                   "AC-11: kpc-L1-wins-ecrcfg-L3-lifecyclePolicy",
			globalKropathMandatory: ECRKropathSection{LifecyclePolicy: policyKropath},
			globalECRCfgMandatory:  ECRConfigSection{LifecyclePolicy: policyECRCfg},
			wantMandatoryLifecycle: policyKropath,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MergeECRCascade(
				tc.globalKropathMandatory, zeroECRKropath,
				tc.globalECRCfgMandatory, zeroECRConfig,
				zeroECRConfig, tc.globalECRCfgDefaults,
				zeroECRKropath, zeroECRKropath,
			)
			if got.Mandatory.LifecyclePolicy != tc.wantMandatoryLifecycle {
				t.Errorf("Mandatory.LifecyclePolicy = %q, want %q", got.Mandatory.LifecyclePolicy, tc.wantMandatoryLifecycle)
			}
			if got.Defaults.LifecyclePolicy != tc.wantDefaultsLifecycle {
				t.Errorf("Defaults.LifecyclePolicy = %q, want %q", got.Defaults.LifecyclePolicy, tc.wantDefaultsLifecycle)
			}
		})
	}
}

// TestMergeECRCascade_NamingTemplate covers AC-12 through AC-13.
func TestMergeECRCascade_NamingTemplate(t *testing.T) {
	tests := []struct {
		name                   string
		globalECRCfgMandatory  ECRConfigSection
		globalECRCfgDefaults   ECRConfigSection
		wantMandatoryNaming    string
		wantDefaultsNaming     string
	}{
		{
			// AC-12: ECRConfig.defaults.namingTemplate (L7) propagates; no KropathConfig entry.
			name:                 "AC-12: global-ecrcfg-defaults-L7-namingTemplate",
			globalECRCfgDefaults: ECRConfigSection{NamingTemplate: "{namespace}/{name}"},
			wantDefaultsNaming:   "{namespace}/{name}",
		},
		{
			// AC-13: ECRConfig.mandatory.namingTemplate (L3) propagates.
			name:                  "AC-13: global-ecrcfg-mandatory-L3-namingTemplate",
			globalECRCfgMandatory: ECRConfigSection{NamingTemplate: "{namespace}/{configRef}/{name}"},
			wantMandatoryNaming:   "{namespace}/{configRef}/{name}",
		},
		{
			// L3 mandatory wins over L4 local mandatory.
			name:                  "global-ecrcfg-mandatory-L3-wins-local-L4",
			globalECRCfgMandatory: ECRConfigSection{NamingTemplate: "{namespace}/{name}"},
			wantMandatoryNaming:   "{namespace}/{name}",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MergeECRCascade(
				zeroECRKropath, zeroECRKropath,
				tc.globalECRCfgMandatory, zeroECRConfig,
				zeroECRConfig, tc.globalECRCfgDefaults,
				zeroECRKropath, zeroECRKropath,
			)
			if got.Mandatory.NamingTemplate != tc.wantMandatoryNaming {
				t.Errorf("Mandatory.NamingTemplate = %q, want %q", got.Mandatory.NamingTemplate, tc.wantMandatoryNaming)
			}
			if got.Defaults.NamingTemplate != tc.wantDefaultsNaming {
				t.Errorf("Defaults.NamingTemplate = %q, want %q", got.Defaults.NamingTemplate, tc.wantDefaultsNaming)
			}
		})
	}
}

// TestMergeECRCascade_TagMerge covers AC-14.
func TestMergeECRCascade_TagMerge(t *testing.T) {
	t.Run("AC-14: tag-union-merge-kpc-and-ecrcfg", func(t *testing.T) {
		got := MergeECRCascade(
			ECRKropathSection{Tags: map[string]string{"cost-centre": "platform"}}, // L1
			zeroECRKropath,
			ECRConfigSection{Tags: map[string]string{"resource-type": "ecr"}}, // L3
			zeroECRConfig,
			zeroECRConfig, zeroECRConfig,
			zeroECRKropath, zeroECRKropath,
		)
		if got.Mandatory.Tags["cost-centre"] != "platform" {
			t.Errorf("Mandatory.Tags[cost-centre] = %q, want %q", got.Mandatory.Tags["cost-centre"], "platform")
		}
		if got.Mandatory.Tags["resource-type"] != "ecr" {
			t.Errorf("Mandatory.Tags[resource-type] = %q, want %q", got.Mandatory.Tags["resource-type"], "ecr")
		}
	})

	t.Run("kpc-L1-tag-wins-ecrcfg-L3-on-conflict", func(t *testing.T) {
		got := MergeECRCascade(
			ECRKropathSection{Tags: map[string]string{"owner": "platform-team"}}, // L1
			zeroECRKropath,
			ECRConfigSection{Tags: map[string]string{"owner": "ecr-team"}}, // L3
			zeroECRConfig,
			zeroECRConfig, zeroECRConfig,
			zeroECRKropath, zeroECRKropath,
		)
		if got.Mandatory.Tags["owner"] != "platform-team" {
			t.Errorf("Mandatory.Tags[owner] = %q, want %q (L1 must win over L3)", got.Mandatory.Tags["owner"], "platform-team")
		}
	})
}

// TestMergeECRCascade_SyncedLabels covers AC-15.
func TestMergeECRCascade_SyncedLabels(t *testing.T) {
	t.Run("AC-15: global-ecrcfg-mandatory-L3-syncedLabels", func(t *testing.T) {
		got := MergeECRCascade(
			zeroECRKropath, zeroECRKropath,
			ECRConfigSection{SyncedLabels: map[string]string{"data-class": "internal"}}, // L3
			zeroECRConfig,
			zeroECRConfig, zeroECRConfig,
			zeroECRKropath, zeroECRKropath,
		)
		if got.Mandatory.SyncedLabels["data-class"] != "internal" {
			t.Errorf("Mandatory.SyncedLabels[data-class] = %q, want %q", got.Mandatory.SyncedLabels["data-class"], "internal")
		}
	})

	t.Run("synced-labels-union-L3-L4", func(t *testing.T) {
		got := MergeECRCascade(
			zeroECRKropath, zeroECRKropath,
			ECRConfigSection{SyncedLabels: map[string]string{"global-label": "g"}},  // L3
			ECRConfigSection{SyncedLabels: map[string]string{"local-label": "l"}},   // L4
			zeroECRConfig, zeroECRConfig,
			zeroECRKropath, zeroECRKropath,
		)
		if got.Mandatory.SyncedLabels["global-label"] != "g" {
			t.Errorf("Mandatory.SyncedLabels[global-label] = %q, want %q", got.Mandatory.SyncedLabels["global-label"], "g")
		}
		if got.Mandatory.SyncedLabels["local-label"] != "l" {
			t.Errorf("Mandatory.SyncedLabels[local-label] = %q, want %q", got.Mandatory.SyncedLabels["local-label"], "l")
		}
	})

	t.Run("synced-labels-L3-wins-L4-on-conflict", func(t *testing.T) {
		got := MergeECRCascade(
			zeroECRKropath, zeroECRKropath,
			ECRConfigSection{SyncedLabels: map[string]string{"tier": "global"}}, // L3
			ECRConfigSection{SyncedLabels: map[string]string{"tier": "local"}},  // L4
			zeroECRConfig, zeroECRConfig,
			zeroECRKropath, zeroECRKropath,
		)
		if got.Mandatory.SyncedLabels["tier"] != "global" {
			t.Errorf("Mandatory.SyncedLabels[tier] = %q, want %q (L3 wins over L4)", got.Mandatory.SyncedLabels["tier"], "global")
		}
	})
}

// TestMergeECRCascade_AllZero verifies the zero-value case returns empty structs.
func TestMergeECRCascade_AllZero(t *testing.T) {
	got := mergeECRAllZero()
	if got.Mandatory.ImageTagMutability != "" {
		t.Errorf("expected empty ImageTagMutability, got %q", got.Mandatory.ImageTagMutability)
	}
	if got.Mandatory.EncryptionType != "" {
		t.Errorf("expected empty EncryptionType, got %q", got.Mandatory.EncryptionType)
	}
	if got.Mandatory.KmsKeyID != "" {
		t.Errorf("expected empty KmsKeyID, got %q", got.Mandatory.KmsKeyID)
	}
	if got.Mandatory.LifecyclePolicy != "" {
		t.Errorf("expected empty LifecyclePolicy, got %q", got.Mandatory.LifecyclePolicy)
	}
	if got.Defaults.ImageTagMutability != "" {
		t.Errorf("expected empty Defaults.ImageTagMutability, got %q", got.Defaults.ImageTagMutability)
	}
	if got.Mandatory.Tags != nil {
		t.Errorf("expected nil Tags, got %v", got.Mandatory.Tags)
	}
}

// TestValidateECREncryption covers AC-17 (negative: encryption misconfiguration).
func TestValidateECREncryption(t *testing.T) {
	tests := []struct {
		name      string
		mandatory EffectiveECRSection
		wantValid bool
		wantReason string
	}{
		{
			// AC-17 (negative): AES256 + kmsKeyID in same mandatory tier → invalid.
			name: "AC-17: AES256-plus-kmsKeyID-invalid",
			mandatory: EffectiveECRSection{
				EncryptionType: "AES256",
				KmsKeyID:       "arn:aws:kms:ap-southeast-2:123456789012:key/mrk-1",
			},
			wantValid:  false,
			wantReason: "InvalidEncryptionConfiguration",
		},
		{
			// KMS + kmsKeyID is valid.
			name: "KMS-plus-kmsKeyID-valid",
			mandatory: EffectiveECRSection{
				EncryptionType: "KMS",
				KmsKeyID:       "arn:aws:kms:ap-southeast-2:123456789012:key/mrk-1",
			},
			wantValid: true,
		},
		{
			// AES256 without kmsKeyID is valid.
			name: "AES256-no-kmsKeyID-valid",
			mandatory: EffectiveECRSection{
				EncryptionType: "AES256",
			},
			wantValid: true,
		},
		{
			// kmsKeyID without encryptionType is valid (encryptionType not yet set).
			name: "kmsKeyID-no-encryptionType-valid",
			mandatory: EffectiveECRSection{
				KmsKeyID: "arn:aws:kms:ap-southeast-2:123456789012:key/mrk-1",
			},
			wantValid: true,
		},
		{
			// Both empty is valid.
			name:      "all-zero-valid",
			mandatory: EffectiveECRSection{},
			wantValid: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			valid, reason, _ := ValidateECREncryption(tc.mandatory)
			if valid != tc.wantValid {
				t.Errorf("valid = %v, want %v", valid, tc.wantValid)
			}
			if !tc.wantValid && reason != tc.wantReason {
				t.Errorf("reason = %q, want %q", reason, tc.wantReason)
			}
		})
	}
}

// TestMergeECRCascade_DefaultsTagPriority verifies defaults tag merge order.
func TestMergeECRCascade_DefaultsTagPriority(t *testing.T) {
	t.Run("L6-wins-over-L9-on-defaults-tag-conflict", func(t *testing.T) {
		got := MergeECRCascade(
			zeroECRKropath, zeroECRKropath,
			zeroECRConfig, zeroECRConfig,
			ECRConfigSection{Tags: map[string]string{"env": "local-profile"}},   // L6
			zeroECRConfig,
			zeroECRKropath,
			ECRKropathSection{Tags: map[string]string{"env": "global-kpc"}}, // L9
		)
		if got.Defaults.Tags["env"] != "local-profile" {
			t.Errorf("Defaults.Tags[env] = %q, want %q (L6 must win over L9)", got.Defaults.Tags["env"], "local-profile")
		}
	})
}
