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

// zeroSageMakerKropath is a convenience zero-value SageMakerKropathSection for table tests.
var zeroSageMakerKropath = SageMakerKropathSection{}

// zeroSageMakerConfig is a convenience zero-value SageMakerConfigSection for table tests.
var zeroSageMakerConfig = SageMakerConfigSection{}

// mergeSageMakerAllZero calls MergeSageMakerCascade with all zero inputs.
func mergeSageMakerAllZero() EffectiveSageMakerConfig {
	return MergeSageMakerCascade(
		zeroSageMakerKropath, zeroSageMakerKropath,
		zeroSageMakerConfig, zeroSageMakerConfig,
		zeroSageMakerConfig, zeroSageMakerConfig,
		zeroSageMakerKropath, zeroSageMakerKropath,
	)
}

// TestMergeSageMakerCascade_InstanceType covers AC-T02-1 and AC-T02-2.
func TestMergeSageMakerCascade_InstanceType(t *testing.T) {
	tests := []struct {
		name                    string
		globalKropathMandatory  SageMakerKropathSection
		localKropathMandatory   SageMakerKropathSection
		globalSMCfgMandatory    SageMakerConfigSection
		localSMCfgMandatory     SageMakerConfigSection
		localSMCfgDefaults      SageMakerConfigSection
		globalSMCfgDefaults     SageMakerConfigSection
		localKropathDefaults    SageMakerKropathSection
		globalKropathDefaults   SageMakerKropathSection
		wantMandatoryInstance   string
		wantDefaultsInstance    string
	}{
		{
			// AC-T02-1: KropathConfig.mandatory.sagemaker.instanceType (L1) wins over
			// SageMakerConfig.mandatory.instanceType (L3).
			name:                   "AC-T02-1: kpc-L1-wins-over-smcfg-L3",
			globalKropathMandatory: SageMakerKropathSection{InstanceType: "ml.m5.xlarge"},
			globalSMCfgMandatory:   SageMakerConfigSection{InstanceType: "ml.t3.medium"},
			wantMandatoryInstance:  "ml.m5.xlarge",
		},
		{
			// AC-T02-2: SageMakerConfig.mandatory.instanceType (L3) wins over
			// SageMakerConfig.defaults.instanceType (L7).
			name:                  "AC-T02-2: smcfg-mandatory-L3-wins-defaults-L7",
			globalSMCfgMandatory:  SageMakerConfigSection{InstanceType: "ml.m5.xlarge"},
			globalSMCfgDefaults:   SageMakerConfigSection{InstanceType: "ml.t3.medium"},
			wantMandatoryInstance: "ml.m5.xlarge",
			wantDefaultsInstance:  "ml.t3.medium",
		},
		{
			// L2 local KropathConfig.mandatory wins over L3 global SageMakerConfig.mandatory.
			name:                   "local-kpc-mandatory-L2-wins-smcfg-L3",
			localKropathMandatory:  SageMakerKropathSection{InstanceType: "ml.c5.2xlarge"},
			globalSMCfgMandatory:   SageMakerConfigSection{InstanceType: "ml.t3.medium"},
			wantMandatoryInstance:  "ml.c5.2xlarge",
		},
		{
			// L3 global SageMakerConfig.mandatory wins over L4 local.
			name:                  "global-smcfg-mandatory-L3-wins-local-L4",
			globalSMCfgMandatory:  SageMakerConfigSection{InstanceType: "ml.m5.xlarge"},
			localSMCfgMandatory:   SageMakerConfigSection{InstanceType: "ml.t3.medium"},
			wantMandatoryInstance: "ml.m5.xlarge",
		},
		{
			// L6 local SageMakerConfig.defaults wins over L7 global.
			name:                 "local-smcfg-defaults-L6-wins-global-L7",
			localSMCfgDefaults:   SageMakerConfigSection{InstanceType: "ml.c5.xlarge"},
			globalSMCfgDefaults:  SageMakerConfigSection{InstanceType: "ml.t3.medium"},
			wantDefaultsInstance: "ml.c5.xlarge",
		},
		{
			// L9 globalKropathDefaults propagates when all higher defaults levels empty.
			name:                  "global-kpc-defaults-L9-propagates",
			globalKropathDefaults: SageMakerKropathSection{InstanceType: "ml.t3.medium"},
			wantDefaultsInstance:  "ml.t3.medium",
		},
		{
			// All zero — empty string returned.
			name:                  "all-zero-returns-empty",
			wantMandatoryInstance: "",
			wantDefaultsInstance:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MergeSageMakerCascade(
				tc.globalKropathMandatory,
				tc.localKropathMandatory,
				tc.globalSMCfgMandatory,
				tc.localSMCfgMandatory,
				tc.localSMCfgDefaults,
				tc.globalSMCfgDefaults,
				tc.localKropathDefaults,
				tc.globalKropathDefaults,
			)
			if got.Mandatory.InstanceType != tc.wantMandatoryInstance {
				t.Errorf("Mandatory.InstanceType = %q, want %q", got.Mandatory.InstanceType, tc.wantMandatoryInstance)
			}
			if got.Defaults.InstanceType != tc.wantDefaultsInstance {
				t.Errorf("Defaults.InstanceType = %q, want %q", got.Defaults.InstanceType, tc.wantDefaultsInstance)
			}
		})
	}
}

// TestMergeSageMakerCascade_VolumeSize covers AC-T02-3.
func TestMergeSageMakerCascade_VolumeSize(t *testing.T) {
	tests := []struct {
		name               string
		globalSMCfgMandatory SageMakerConfigSection
		localSMCfgMandatory  SageMakerConfigSection
		localSMCfgDefaults   SageMakerConfigSection
		globalSMCfgDefaults  SageMakerConfigSection
		wantMandatoryVol   int64
		wantDefaultsVol    int64
	}{
		{
			// AC-T02-3: mandatory.volumeSizeInGB (L3) flows to effectiveConfig.mandatory.
			name:                 "AC-T02-3: global-smcfg-mandatory-L3-volumeSize",
			globalSMCfgMandatory: SageMakerConfigSection{VolumeSize: 100},
			wantMandatoryVol:     100,
		},
		{
			// defaults.volumeSizeInGB (L7) flows to effectiveConfig.defaults.
			name:                "global-smcfg-defaults-L7-volumeSize",
			globalSMCfgDefaults: SageMakerConfigSection{VolumeSize: 5},
			wantDefaultsVol:     5,
		},
		{
			// L3 global wins over L4 local on mandatory.
			name:                 "global-smcfg-mandatory-L3-wins-local-L4",
			globalSMCfgMandatory: SageMakerConfigSection{VolumeSize: 100},
			localSMCfgMandatory:  SageMakerConfigSection{VolumeSize: 50},
			wantMandatoryVol:     100,
		},
		{
			// L6 local wins over L7 global on defaults.
			name:                "local-smcfg-defaults-L6-wins-global-L7",
			localSMCfgDefaults:  SageMakerConfigSection{VolumeSize: 20},
			globalSMCfgDefaults: SageMakerConfigSection{VolumeSize: 5},
			wantDefaultsVol:     20,
		},
		{
			// VolumeSize is NOT in KropathConfig.sagemaker — zero value always from kpc levels.
			name:            "all-zero-returns-zero",
			wantMandatoryVol: 0,
			wantDefaultsVol:  0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MergeSageMakerCascade(
				zeroSageMakerKropath, zeroSageMakerKropath,
				tc.globalSMCfgMandatory, tc.localSMCfgMandatory,
				tc.localSMCfgDefaults, tc.globalSMCfgDefaults,
				zeroSageMakerKropath, zeroSageMakerKropath,
			)
			if got.Mandatory.VolumeSize != tc.wantMandatoryVol {
				t.Errorf("Mandatory.VolumeSize = %d, want %d", got.Mandatory.VolumeSize, tc.wantMandatoryVol)
			}
			if got.Defaults.VolumeSize != tc.wantDefaultsVol {
				t.Errorf("Defaults.VolumeSize = %d, want %d", got.Defaults.VolumeSize, tc.wantDefaultsVol)
			}
		})
	}
}

// TestMergeSageMakerCascade_KmsKeyId covers AC-T02-4.
func TestMergeSageMakerCascade_KmsKeyId(t *testing.T) {
	const (
		kmsKPC    = "arn:aws:kms:us-east-1:123456789012:key/kpc-key"
		kmsSMCfg  = "arn:aws:kms:us-east-1:123456789012:key/smcfg-key"
		kmsLocal  = "arn:aws:kms:us-east-1:123456789012:key/local-key"
	)

	tests := []struct {
		name                   string
		globalKropathMandatory SageMakerKropathSection
		globalSMCfgMandatory   SageMakerConfigSection
		localSMCfgMandatory    SageMakerConfigSection
		localSMCfgDefaults     SageMakerConfigSection
		globalSMCfgDefaults    SageMakerConfigSection
		globalKropathDefaults  SageMakerKropathSection
		wantMandatoryKmsKeyId  string
		wantDefaultsKmsKeyId   string
	}{
		{
			// AC-T02-4: KropathConfig.mandatory.sagemaker.kmsKeyId (L1) flows to effectiveConfig.mandatory.
			name:                   "AC-T02-4: global-kpc-mandatory-L1-kmsKeyId",
			globalKropathMandatory: SageMakerKropathSection{KmsKeyId: kmsKPC},
			wantMandatoryKmsKeyId:  kmsKPC,
		},
		{
			// L1 KropathConfig.mandatory wins over L3 SageMakerConfig.mandatory.
			name:                   "kpc-L1-wins-smcfg-L3-kmsKeyId",
			globalKropathMandatory: SageMakerKropathSection{KmsKeyId: kmsKPC},
			globalSMCfgMandatory:   SageMakerConfigSection{KmsKeyId: kmsSMCfg},
			wantMandatoryKmsKeyId:  kmsKPC,
		},
		{
			// L3 global SageMakerConfig.mandatory kmsKeyId propagates.
			name:                  "global-smcfg-mandatory-L3-kmsKeyId",
			globalSMCfgMandatory:  SageMakerConfigSection{KmsKeyId: kmsSMCfg},
			wantMandatoryKmsKeyId: kmsSMCfg,
		},
		{
			// defaults.kmsKeyId from L7 propagates.
			name:                 "global-smcfg-defaults-L7-kmsKeyId",
			globalSMCfgDefaults:  SageMakerConfigSection{KmsKeyId: kmsSMCfg},
			wantDefaultsKmsKeyId: kmsSMCfg,
		},
		{
			// L9 globalKropathDefaults kmsKeyId propagates.
			name:                  "global-kpc-defaults-L9-kmsKeyId",
			globalKropathDefaults: SageMakerKropathSection{KmsKeyId: kmsKPC},
			wantDefaultsKmsKeyId:  kmsKPC,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MergeSageMakerCascade(
				tc.globalKropathMandatory, zeroSageMakerKropath,
				tc.globalSMCfgMandatory, tc.localSMCfgMandatory,
				tc.localSMCfgDefaults, tc.globalSMCfgDefaults,
				zeroSageMakerKropath, tc.globalKropathDefaults,
			)
			if got.Mandatory.KmsKeyId != tc.wantMandatoryKmsKeyId {
				t.Errorf("Mandatory.KmsKeyId = %q, want %q", got.Mandatory.KmsKeyId, tc.wantMandatoryKmsKeyId)
			}
			if got.Defaults.KmsKeyId != tc.wantDefaultsKmsKeyId {
				t.Errorf("Defaults.KmsKeyId = %q, want %q", got.Defaults.KmsKeyId, tc.wantDefaultsKmsKeyId)
			}
		})
	}
}

// TestMergeSageMakerCascade_EnableNetworkIsolation covers AC-T02-5.
func TestMergeSageMakerCascade_EnableNetworkIsolation(t *testing.T) {
	tests := []struct {
		name                    string
		globalKropathMandatory  SageMakerKropathSection
		globalSMCfgMandatory    SageMakerConfigSection
		localSMCfgDefaults      SageMakerConfigSection
		globalSMCfgDefaults     SageMakerConfigSection
		globalKropathDefaults   SageMakerKropathSection
		wantMandatoryNetIso     bool
		wantDefaultsNetIso      bool
	}{
		{
			// AC-T02-5: KropathConfig.mandatory.sagemaker.enableNetworkIsolation (L1) = true.
			name:                    "AC-T02-5: kpc-mandatory-L1-network-isolation",
			globalKropathMandatory:  SageMakerKropathSection{EnableNetworkIsolation: true},
			wantMandatoryNetIso:     true,
		},
		{
			// SageMakerConfig.mandatory.enableNetworkIsolation (L3) = true.
			name:                 "global-smcfg-mandatory-L3-network-isolation",
			globalSMCfgMandatory: SageMakerConfigSection{EnableNetworkIsolation: true},
			wantMandatoryNetIso:  true,
		},
		{
			// defaults.enableNetworkIsolation (L7) = true propagates.
			name:                "global-smcfg-defaults-L7-network-isolation",
			globalSMCfgDefaults: SageMakerConfigSection{EnableNetworkIsolation: true},
			wantDefaultsNetIso:  true,
		},
		{
			// L9 globalKropathDefaults propagates.
			name:                  "global-kpc-defaults-L9-network-isolation",
			globalKropathDefaults: SageMakerKropathSection{EnableNetworkIsolation: true},
			wantDefaultsNetIso:    true,
		},
		{
			// All zero — false (not enforced).
			name:            "all-zero-returns-false",
			wantMandatoryNetIso: false,
			wantDefaultsNetIso:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MergeSageMakerCascade(
				tc.globalKropathMandatory, zeroSageMakerKropath,
				tc.globalSMCfgMandatory, zeroSageMakerConfig,
				tc.localSMCfgDefaults, tc.globalSMCfgDefaults,
				zeroSageMakerKropath, tc.globalKropathDefaults,
			)
			if got.Mandatory.EnableNetworkIsolation != tc.wantMandatoryNetIso {
				t.Errorf("Mandatory.EnableNetworkIsolation = %v, want %v", got.Mandatory.EnableNetworkIsolation, tc.wantMandatoryNetIso)
			}
			if got.Defaults.EnableNetworkIsolation != tc.wantDefaultsNetIso {
				t.Errorf("Defaults.EnableNetworkIsolation = %v, want %v", got.Defaults.EnableNetworkIsolation, tc.wantDefaultsNetIso)
			}
		})
	}
}

// TestMergeSageMakerCascade_EnableInterContainerTrafficEncryption covers booleans from KPC.
func TestMergeSageMakerCascade_EnableInterContainerTrafficEncryption(t *testing.T) {
	tests := []struct {
		name                    string
		globalKropathMandatory  SageMakerKropathSection
		globalSMCfgMandatory    SageMakerConfigSection
		wantMandatoryEncrypt    bool
	}{
		{
			name:                   "kpc-mandatory-L1-intercontainer-encryption",
			globalKropathMandatory: SageMakerKropathSection{EnableInterContainerTrafficEncryption: true},
			wantMandatoryEncrypt:   true,
		},
		{
			name:                 "smcfg-mandatory-L3-intercontainer-encryption",
			globalSMCfgMandatory: SageMakerConfigSection{EnableInterContainerTrafficEncryption: true},
			wantMandatoryEncrypt: true,
		},
		{
			// KropathConfig wins over SageMakerConfig on same tier.
			name:                   "kpc-L1-wins-smcfg-L3-intercontainer",
			globalKropathMandatory: SageMakerKropathSection{EnableInterContainerTrafficEncryption: true},
			globalSMCfgMandatory:   SageMakerConfigSection{EnableInterContainerTrafficEncryption: true},
			wantMandatoryEncrypt:   true,
		},
		{
			name:                 "all-zero-returns-false",
			wantMandatoryEncrypt: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MergeSageMakerCascade(
				tc.globalKropathMandatory, zeroSageMakerKropath,
				tc.globalSMCfgMandatory, zeroSageMakerConfig,
				zeroSageMakerConfig, zeroSageMakerConfig,
				zeroSageMakerKropath, zeroSageMakerKropath,
			)
			if got.Mandatory.EnableInterContainerTrafficEncryption != tc.wantMandatoryEncrypt {
				t.Errorf("Mandatory.EnableInterContainerTrafficEncryption = %v, want %v",
					got.Mandatory.EnableInterContainerTrafficEncryption, tc.wantMandatoryEncrypt)
			}
		})
	}
}

// TestMergeSageMakerCascade_RootAccess covers AC-T02-6.
func TestMergeSageMakerCascade_RootAccess(t *testing.T) {
	tests := []struct {
		name                  string
		globalSMCfgMandatory  SageMakerConfigSection
		localSMCfgDefaults    SageMakerConfigSection
		globalSMCfgDefaults   SageMakerConfigSection
		wantMandatoryRoot     string
		wantDefaultsRoot      string
	}{
		{
			// AC-T02-6: SageMakerConfig.defaults.rootAccess (L7) flows to effectiveConfig.defaults.
			name:                "AC-T02-6: smcfg-defaults-L7-rootAccess-enabled",
			globalSMCfgDefaults: SageMakerConfigSection{RootAccess: "Enabled"},
			wantDefaultsRoot:    "Enabled",
		},
		{
			// mandatory.rootAccess (L3) flows to effectiveConfig.mandatory.
			name:                 "smcfg-mandatory-L3-rootAccess-disabled",
			globalSMCfgMandatory: SageMakerConfigSection{RootAccess: "Disabled"},
			wantMandatoryRoot:    "Disabled",
		},
		{
			// L6 local wins over L7 global defaults.
			name:                "local-smcfg-defaults-L6-wins-global-L7",
			localSMCfgDefaults:  SageMakerConfigSection{RootAccess: "Disabled"},
			globalSMCfgDefaults: SageMakerConfigSection{RootAccess: "Enabled"},
			wantDefaultsRoot:    "Disabled",
		},
		{
			// RootAccess is NOT in KropathConfig.sagemaker — cannot propagate from kpc levels.
			name:             "all-zero-returns-empty",
			wantMandatoryRoot: "",
			wantDefaultsRoot:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MergeSageMakerCascade(
				zeroSageMakerKropath, zeroSageMakerKropath,
				tc.globalSMCfgMandatory, zeroSageMakerConfig,
				tc.localSMCfgDefaults, tc.globalSMCfgDefaults,
				zeroSageMakerKropath, zeroSageMakerKropath,
			)
			if got.Mandatory.RootAccess != tc.wantMandatoryRoot {
				t.Errorf("Mandatory.RootAccess = %q, want %q", got.Mandatory.RootAccess, tc.wantMandatoryRoot)
			}
			if got.Defaults.RootAccess != tc.wantDefaultsRoot {
				t.Errorf("Defaults.RootAccess = %q, want %q", got.Defaults.RootAccess, tc.wantDefaultsRoot)
			}
		})
	}
}

// TestMergeSageMakerCascade_DirectInternetAccess covers AC-T02-7.
func TestMergeSageMakerCascade_DirectInternetAccess(t *testing.T) {
	tests := []struct {
		name                   string
		globalSMCfgMandatory   SageMakerConfigSection
		globalSMCfgDefaults    SageMakerConfigSection
		wantMandatoryDIA       string
		wantDefaultsDIA        string
	}{
		{
			// AC-T02-7: mandatory.directInternetAccess = "Disabled" flows to effectiveConfig.mandatory.
			name:                  "AC-T02-7: smcfg-mandatory-L3-directInternetAccess-disabled",
			globalSMCfgMandatory:  SageMakerConfigSection{DirectInternetAccess: "Disabled"},
			wantMandatoryDIA:      "Disabled",
		},
		{
			// defaults.directInternetAccess = "Enabled" flows to effectiveConfig.defaults.
			name:                 "smcfg-defaults-L7-directInternetAccess-enabled",
			globalSMCfgDefaults:  SageMakerConfigSection{DirectInternetAccess: "Enabled"},
			wantDefaultsDIA:      "Enabled",
		},
		{
			// DirectInternetAccess is NOT in KropathConfig.sagemaker.
			name:            "all-zero-returns-empty",
			wantMandatoryDIA: "",
			wantDefaultsDIA:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MergeSageMakerCascade(
				zeroSageMakerKropath, zeroSageMakerKropath,
				tc.globalSMCfgMandatory, zeroSageMakerConfig,
				zeroSageMakerConfig, tc.globalSMCfgDefaults,
				zeroSageMakerKropath, zeroSageMakerKropath,
			)
			if got.Mandatory.DirectInternetAccess != tc.wantMandatoryDIA {
				t.Errorf("Mandatory.DirectInternetAccess = %q, want %q", got.Mandatory.DirectInternetAccess, tc.wantMandatoryDIA)
			}
			if got.Defaults.DirectInternetAccess != tc.wantDefaultsDIA {
				t.Errorf("Defaults.DirectInternetAccess = %q, want %q", got.Defaults.DirectInternetAccess, tc.wantDefaultsDIA)
			}
		})
	}
}

// TestMergeSageMakerCascade_NamingTemplate covers AC-T02-8.
func TestMergeSageMakerCascade_NamingTemplate(t *testing.T) {
	tests := []struct {
		name                 string
		globalSMCfgMandatory SageMakerConfigSection
		localSMCfgMandatory  SageMakerConfigSection
		localSMCfgDefaults   SageMakerConfigSection
		globalSMCfgDefaults  SageMakerConfigSection
		wantMandatoryNaming  string
		wantDefaultsNaming   string
	}{
		{
			// AC-T02-8: SageMakerConfig.defaults.namingTemplate (L7) flows to effectiveConfig.defaults.
			name:                "AC-T02-8: smcfg-defaults-L7-namingTemplate",
			globalSMCfgDefaults: SageMakerConfigSection{NamingTemplate: "{namespace}-{name}"},
			wantDefaultsNaming:  "{namespace}-{name}",
		},
		{
			// mandatory.namingTemplate (L3) propagates.
			name:                 "smcfg-mandatory-L3-namingTemplate",
			globalSMCfgMandatory: SageMakerConfigSection{NamingTemplate: "{namespace}/{name}"},
			wantMandatoryNaming:  "{namespace}/{name}",
		},
		{
			// L3 global wins over L4 local mandatory.
			name:                 "global-smcfg-mandatory-L3-wins-local-L4",
			globalSMCfgMandatory: SageMakerConfigSection{NamingTemplate: "{namespace}-{name}"},
			localSMCfgMandatory:  SageMakerConfigSection{NamingTemplate: "{namespace}/{name}"},
			wantMandatoryNaming:  "{namespace}-{name}",
		},
		{
			// NamingTemplate is NOT in KropathConfig.sagemaker.
			name:            "all-zero-returns-empty",
			wantMandatoryNaming: "",
			wantDefaultsNaming:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MergeSageMakerCascade(
				zeroSageMakerKropath, zeroSageMakerKropath,
				tc.globalSMCfgMandatory, tc.localSMCfgMandatory,
				tc.localSMCfgDefaults, tc.globalSMCfgDefaults,
				zeroSageMakerKropath, zeroSageMakerKropath,
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

// TestMergeSageMakerCascade_TagMerge covers AC-T02-9.
func TestMergeSageMakerCascade_TagMerge(t *testing.T) {
	t.Run("AC-T02-9: tag-union-merge-kpc-and-smcfg-mandatory", func(t *testing.T) {
		got := MergeSageMakerCascade(
			SageMakerKropathSection{Tags: map[string]string{"cost-centre": "platform"}}, // L1
			zeroSageMakerKropath,
			SageMakerConfigSection{Tags: map[string]string{"resource-type": "sagemaker"}}, // L3
			zeroSageMakerConfig,
			zeroSageMakerConfig, zeroSageMakerConfig,
			zeroSageMakerKropath, zeroSageMakerKropath,
		)
		if got.Mandatory.Tags["cost-centre"] != "platform" {
			t.Errorf("Mandatory.Tags[cost-centre] = %q, want %q", got.Mandatory.Tags["cost-centre"], "platform")
		}
		if got.Mandatory.Tags["resource-type"] != "sagemaker" {
			t.Errorf("Mandatory.Tags[resource-type] = %q, want %q", got.Mandatory.Tags["resource-type"], "sagemaker")
		}
	})

	t.Run("kpc-L1-tag-wins-smcfg-L3-on-conflict", func(t *testing.T) {
		got := MergeSageMakerCascade(
			SageMakerKropathSection{Tags: map[string]string{"owner": "platform-team"}}, // L1
			zeroSageMakerKropath,
			SageMakerConfigSection{Tags: map[string]string{"owner": "ml-team"}}, // L3
			zeroSageMakerConfig,
			zeroSageMakerConfig, zeroSageMakerConfig,
			zeroSageMakerKropath, zeroSageMakerKropath,
		)
		if got.Mandatory.Tags["owner"] != "platform-team" {
			t.Errorf("Mandatory.Tags[owner] = %q, want %q (L1 must win over L3)", got.Mandatory.Tags["owner"], "platform-team")
		}
	})

	t.Run("defaults-L6-wins-over-L9-on-tag-conflict", func(t *testing.T) {
		got := MergeSageMakerCascade(
			zeroSageMakerKropath, zeroSageMakerKropath,
			zeroSageMakerConfig, zeroSageMakerConfig,
			SageMakerConfigSection{Tags: map[string]string{"env": "local-smcfg"}},    // L6
			zeroSageMakerConfig,
			zeroSageMakerKropath,
			SageMakerKropathSection{Tags: map[string]string{"env": "global-kpc"}}, // L9
		)
		if got.Defaults.Tags["env"] != "local-smcfg" {
			t.Errorf("Defaults.Tags[env] = %q, want %q (L6 must win over L9)", got.Defaults.Tags["env"], "local-smcfg")
		}
	})
}

// TestMergeSageMakerCascade_SyncedLabels covers AC-T02-10.
func TestMergeSageMakerCascade_SyncedLabels(t *testing.T) {
	t.Run("AC-T02-10: global-smcfg-mandatory-L3-syncedLabels", func(t *testing.T) {
		got := MergeSageMakerCascade(
			zeroSageMakerKropath, zeroSageMakerKropath,
			SageMakerConfigSection{SyncedLabels: map[string]string{"data-class": "sensitive"}}, // L3
			zeroSageMakerConfig,
			zeroSageMakerConfig, zeroSageMakerConfig,
			zeroSageMakerKropath, zeroSageMakerKropath,
		)
		if got.Mandatory.SyncedLabels["data-class"] != "sensitive" {
			t.Errorf("Mandatory.SyncedLabels[data-class] = %q, want %q", got.Mandatory.SyncedLabels["data-class"], "sensitive")
		}
	})

	t.Run("synced-labels-union-L3-L4", func(t *testing.T) {
		got := MergeSageMakerCascade(
			zeroSageMakerKropath, zeroSageMakerKropath,
			SageMakerConfigSection{SyncedLabels: map[string]string{"global-label": "g"}}, // L3
			SageMakerConfigSection{SyncedLabels: map[string]string{"local-label": "l"}},  // L4
			zeroSageMakerConfig, zeroSageMakerConfig,
			zeroSageMakerKropath, zeroSageMakerKropath,
		)
		if got.Mandatory.SyncedLabels["global-label"] != "g" {
			t.Errorf("Mandatory.SyncedLabels[global-label] = %q, want %q", got.Mandatory.SyncedLabels["global-label"], "g")
		}
		if got.Mandatory.SyncedLabels["local-label"] != "l" {
			t.Errorf("Mandatory.SyncedLabels[local-label] = %q, want %q", got.Mandatory.SyncedLabels["local-label"], "l")
		}
	})

	t.Run("synced-labels-L3-wins-L4-on-conflict", func(t *testing.T) {
		got := MergeSageMakerCascade(
			zeroSageMakerKropath, zeroSageMakerKropath,
			SageMakerConfigSection{SyncedLabels: map[string]string{"tier": "global"}}, // L3
			SageMakerConfigSection{SyncedLabels: map[string]string{"tier": "local"}},  // L4
			zeroSageMakerConfig, zeroSageMakerConfig,
			zeroSageMakerKropath, zeroSageMakerKropath,
		)
		if got.Mandatory.SyncedLabels["tier"] != "global" {
			t.Errorf("Mandatory.SyncedLabels[tier] = %q, want %q (L3 wins over L4)", got.Mandatory.SyncedLabels["tier"], "global")
		}
	})

	t.Run("synced-annotations-defaults-L6-L7", func(t *testing.T) {
		got := MergeSageMakerCascade(
			zeroSageMakerKropath, zeroSageMakerKropath,
			zeroSageMakerConfig, zeroSageMakerConfig,
			SageMakerConfigSection{SyncedAnnotations: map[string]string{"local-ann": "l"}}, // L6
			SageMakerConfigSection{SyncedAnnotations: map[string]string{"global-ann": "g"}}, // L7
			zeroSageMakerKropath, zeroSageMakerKropath,
		)
		if got.Defaults.SyncedAnnotations["local-ann"] != "l" {
			t.Errorf("Defaults.SyncedAnnotations[local-ann] = %q, want %q", got.Defaults.SyncedAnnotations["local-ann"], "l")
		}
		if got.Defaults.SyncedAnnotations["global-ann"] != "g" {
			t.Errorf("Defaults.SyncedAnnotations[global-ann] = %q, want %q", got.Defaults.SyncedAnnotations["global-ann"], "g")
		}
	})
}

// TestMergeSageMakerCascade_AllZero verifies the zero-value case returns empty structs.
func TestMergeSageMakerCascade_AllZero(t *testing.T) {
	got := mergeSageMakerAllZero()
	if got.Mandatory.InstanceType != "" {
		t.Errorf("expected empty Mandatory.InstanceType, got %q", got.Mandatory.InstanceType)
	}
	if got.Mandatory.VolumeSize != 0 {
		t.Errorf("expected zero Mandatory.VolumeSize, got %d", got.Mandatory.VolumeSize)
	}
	if got.Mandatory.KmsKeyId != "" {
		t.Errorf("expected empty Mandatory.KmsKeyId, got %q", got.Mandatory.KmsKeyId)
	}
	if got.Mandatory.EnableNetworkIsolation != false {
		t.Errorf("expected false Mandatory.EnableNetworkIsolation, got %v", got.Mandatory.EnableNetworkIsolation)
	}
	if got.Mandatory.RootAccess != "" {
		t.Errorf("expected empty Mandatory.RootAccess, got %q", got.Mandatory.RootAccess)
	}
	if got.Mandatory.Tags != nil {
		t.Errorf("expected nil Mandatory.Tags, got %v", got.Mandatory.Tags)
	}
	if got.Defaults.InstanceType != "" {
		t.Errorf("expected empty Defaults.InstanceType, got %q", got.Defaults.InstanceType)
	}
	if got.Defaults.VolumeSize != 0 {
		t.Errorf("expected zero Defaults.VolumeSize, got %d", got.Defaults.VolumeSize)
	}
}
