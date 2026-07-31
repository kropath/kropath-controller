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
	"reflect"
	"testing"
)

// zeroSMKropath is a convenience zero-value SMKropathSection for table tests.
var zeroSMKropath = SMKropathSection{}

// zeroSMConfig is a convenience zero-value SMConfigSection for table tests.
var zeroSMConfig = SMConfigSection{}

func TestMergeSecretsManagerCascade_KmsKeyID(t *testing.T) {
	const (
		kmsOrg    = "arn:aws:kms:ap-southeast-2:123456789012:key/mrk-org"
		kmsNs     = "arn:aws:kms:ap-southeast-2:123456789012:key/mrk-ns"
		kmsGlobal = "arn:aws:kms:ap-southeast-2:123456789012:key/mrk-profile-global"
		kmsLocal  = "arn:aws:kms:ap-southeast-2:123456789012:key/mrk-profile-local"
		kmsAlias  = "alias/default-key"
	)

	tests := []struct {
		name                   string
		globalKropathMandatory SMKropathSection
		localKropathMandatory  SMKropathSection
		globalSMCfgMandatory   SMConfigSection
		localSMCfgMandatory    SMConfigSection
		localSMCfgDefaults     SMConfigSection
		globalSMCfgDefaults    SMConfigSection
		localKropathDefaults   SMKropathSection
		globalKropathDefaults  SMKropathSection
		wantMandatoryKmsKeyID  string
		wantDefaultsKmsKeyID   string
	}{
		{
			// AC-1: global KropathConfig.mandatory.secretsManager.kmsKeyID (L1) propagates.
			name:                   "AC-1: global-kpc-mandatory-L1-propagates",
			globalKropathMandatory: SMKropathSection{KmsKeyID: kmsOrg},
			wantMandatoryKmsKeyID:  kmsOrg,
		},
		{
			// AC-2: KropathConfig L1 wins over SecretsManagerConfig L3.
			name:                   "AC-2: kpc-L1-wins-over-smcfg-L3",
			globalKropathMandatory: SMKropathSection{KmsKeyID: kmsOrg},
			globalSMCfgMandatory:   SMConfigSection{KmsKeyID: kmsGlobal},
			wantMandatoryKmsKeyID:  kmsOrg,
		},
		{
			// AC-3: Only globalSMCfg.defaults.kmsKeyID set; mandatory stays empty.
			name:                  "AC-3: global-smcfg-defaults-L7-kmsKeyID",
			globalSMCfgDefaults:   SMConfigSection{KmsKeyID: kmsAlias},
			wantMandatoryKmsKeyID: "",
			wantDefaultsKmsKeyID:  kmsAlias,
		},
		{
			// AC-4: global SecretsManagerConfig.mandatory (L3) wins over local (L4).
			name:                  "AC-4: global-smcfg-mandatory-L3-wins-local-L4",
			globalSMCfgMandatory:  SMConfigSection{KmsKeyID: kmsGlobal},
			localSMCfgMandatory:   SMConfigSection{KmsKeyID: kmsLocal},
			wantMandatoryKmsKeyID: kmsGlobal,
		},
		{
			// L2 KropathConfig.mandatory wins over L3 SecretsManagerConfig.mandatory.
			name:                  "local-kpc-mandatory-L2-wins-over-smcfg-L3",
			localKropathMandatory: SMKropathSection{KmsKeyID: kmsNs},
			globalSMCfgMandatory:  SMConfigSection{KmsKeyID: kmsGlobal},
			wantMandatoryKmsKeyID: kmsNs,
		},
		{
			// L9 globalKropathDefaults propagates when all higher defaults levels empty.
			name:                  "global-kpc-defaults-L9-propagates",
			globalKropathDefaults: SMKropathSection{KmsKeyID: kmsAlias},
			wantDefaultsKmsKeyID:  kmsAlias,
		},
		{
			// L6 local SMCfgDefaults wins over L7 global.
			name:                 "local-smcfg-defaults-L6-wins-global-L7",
			localSMCfgDefaults:   SMConfigSection{KmsKeyID: kmsLocal},
			globalSMCfgDefaults:  SMConfigSection{KmsKeyID: kmsGlobal},
			wantDefaultsKmsKeyID: kmsLocal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MergeSecretsManagerCascade(
				tc.globalKropathMandatory,
				tc.localKropathMandatory,
				tc.globalSMCfgMandatory,
				tc.localSMCfgMandatory,
				tc.localSMCfgDefaults,
				tc.globalSMCfgDefaults,
				tc.localKropathDefaults,
				tc.globalKropathDefaults,
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

func TestMergeSecretsManagerCascade_ReplicaRegions(t *testing.T) {
	usWest2 := []ReplicaRegion{{Region: "us-west-2"}}
	euWest1 := []ReplicaRegion{{Region: "eu-west-1"}}
	multiRegion := []ReplicaRegion{{Region: "us-west-2"}, {Region: "eu-west-1"}}

	tests := []struct {
		name                  string
		globalSMCfgMandatory  SMConfigSection
		localSMCfgMandatory   SMConfigSection
		localSMCfgDefaults    SMConfigSection
		globalSMCfgDefaults   SMConfigSection
		wantMandatoryReplicas []ReplicaRegion
		wantDefaultsReplicas  []ReplicaRegion
	}{
		{
			// AC-5: global SecretsManagerConfig.mandatory.replicaRegions (L3) propagates.
			name:                  "AC-5: global-smcfg-mandatory-L3-replicaRegions",
			globalSMCfgMandatory:  SMConfigSection{ReplicaRegions: usWest2},
			wantMandatoryReplicas: usWest2,
		},
		{
			// AC-6: L3 wins over L4 — priority replacement, NOT additive.
			// L3 = [{us-west-2}], L4 = [{eu-west-1}] => result must be [{us-west-2}], not both.
			name:                  "AC-6: priority-replacement-L3-wins-L4",
			globalSMCfgMandatory:  SMConfigSection{ReplicaRegions: usWest2},
			localSMCfgMandatory:   SMConfigSection{ReplicaRegions: euWest1},
			wantMandatoryReplicas: usWest2,
		},
		{
			// AC-7: all mandatory empty; global defaults (L7) propagates.
			name:                 "AC-7: global-smcfg-defaults-L7-replicaRegions",
			globalSMCfgDefaults:  SMConfigSection{ReplicaRegions: usWest2},
			wantDefaultsReplicas: usWest2,
		},
		{
			// L4 only: local mandatory propagates when global mandatory empty.
			name:                  "local-mandatory-L4-propagates-when-L3-empty",
			localSMCfgMandatory:   SMConfigSection{ReplicaRegions: euWest1},
			wantMandatoryReplicas: euWest1,
		},
		{
			// L6 local defaults wins over L7 global (priority replacement in defaults too).
			name:                 "local-defaults-L6-wins-global-L7",
			localSMCfgDefaults:  SMConfigSection{ReplicaRegions: euWest1},
			globalSMCfgDefaults: SMConfigSection{ReplicaRegions: usWest2},
			wantDefaultsReplicas: euWest1,
		},
		{
			// Multi-region slice is copied correctly.
			name:                  "multi-region-slice-copied",
			globalSMCfgMandatory:  SMConfigSection{ReplicaRegions: multiRegion},
			wantMandatoryReplicas: multiRegion,
		},
		{
			// Priority replacement does NOT union-merge arrays.
			name:                  "arrays-not-union-merged",
			globalSMCfgMandatory:  SMConfigSection{ReplicaRegions: usWest2},
			localSMCfgMandatory:   SMConfigSection{ReplicaRegions: euWest1},
			wantMandatoryReplicas: usWest2, // NOT [{us-west-2},{eu-west-1}]
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MergeSecretsManagerCascade(
				zeroSMKropath, zeroSMKropath,
				tc.globalSMCfgMandatory,
				tc.localSMCfgMandatory,
				tc.localSMCfgDefaults,
				tc.globalSMCfgDefaults,
				zeroSMKropath, zeroSMKropath,
			)
			if !reflect.DeepEqual(got.Mandatory.ReplicaRegions, tc.wantMandatoryReplicas) {
				t.Errorf("Mandatory.ReplicaRegions = %v, want %v",
					got.Mandatory.ReplicaRegions, tc.wantMandatoryReplicas)
			}
			if !reflect.DeepEqual(got.Defaults.ReplicaRegions, tc.wantDefaultsReplicas) {
				t.Errorf("Defaults.ReplicaRegions = %v, want %v",
					got.Defaults.ReplicaRegions, tc.wantDefaultsReplicas)
			}
		})
	}
}

func TestMergeSecretsManagerCascade_ForceOverwriteReplicaSecret(t *testing.T) {
	tests := []struct {
		name                        string
		globalSMCfgMandatory        SMConfigSection
		localSMCfgMandatory         SMConfigSection
		localSMCfgDefaults          SMConfigSection
		globalSMCfgDefaults         SMConfigSection
		wantMandatoryForceOverwrite bool
		wantDefaultsForceOverwrite  bool
	}{
		{
			// AC-8: global mandatory forceOverwrite=true (L3) propagates.
			name:                        "AC-8: global-mandatory-L3-forceOverwrite-true",
			globalSMCfgMandatory:        SMConfigSection{ForceOverwriteReplicaSecret: true},
			wantMandatoryForceOverwrite: true,
		},
		{
			// L3=true wins over L4=false (firstTrue).
			name:                        "L3-true-wins-over-L4-false",
			globalSMCfgMandatory:        SMConfigSection{ForceOverwriteReplicaSecret: true},
			localSMCfgMandatory:         SMConfigSection{ForceOverwriteReplicaSecret: false},
			wantMandatoryForceOverwrite: true,
		},
		{
			// Only L4=true: propagates.
			name:                        "L4-true-propagates-when-L3-false",
			localSMCfgMandatory:         SMConfigSection{ForceOverwriteReplicaSecret: true},
			wantMandatoryForceOverwrite: true,
		},
		{
			// L6=true defaults propagates.
			name:                       "local-defaults-L6-true",
			localSMCfgDefaults:         SMConfigSection{ForceOverwriteReplicaSecret: true},
			wantDefaultsForceOverwrite: true,
		},
		{
			// All false: result is false.
			name:                        "all-false-result-false",
			wantMandatoryForceOverwrite: false,
			wantDefaultsForceOverwrite:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MergeSecretsManagerCascade(
				zeroSMKropath, zeroSMKropath,
				tc.globalSMCfgMandatory,
				tc.localSMCfgMandatory,
				tc.localSMCfgDefaults,
				tc.globalSMCfgDefaults,
				zeroSMKropath, zeroSMKropath,
			)
			if got.Mandatory.ForceOverwriteReplicaSecret != tc.wantMandatoryForceOverwrite {
				t.Errorf("Mandatory.ForceOverwriteReplicaSecret = %v, want %v",
					got.Mandatory.ForceOverwriteReplicaSecret, tc.wantMandatoryForceOverwrite)
			}
			if got.Defaults.ForceOverwriteReplicaSecret != tc.wantDefaultsForceOverwrite {
				t.Errorf("Defaults.ForceOverwriteReplicaSecret = %v, want %v",
					got.Defaults.ForceOverwriteReplicaSecret, tc.wantDefaultsForceOverwrite)
			}
		})
	}
}

func TestMergeSecretsManagerCascade_NamingTemplate(t *testing.T) {
	const (
		tmplDefault   = "{namespace}-{name}"
		tmplMandatory = "{namespace}-{configRef}-{name}"
		tmplLocal     = "{namespace}-local-{name}"
	)

	tests := []struct {
		name                  string
		globalSMCfgMandatory  SMConfigSection
		localSMCfgMandatory   SMConfigSection
		localSMCfgDefaults    SMConfigSection
		globalSMCfgDefaults   SMConfigSection
		wantMandatoryTemplate string
		wantDefaultsTemplate  string
	}{
		{
			// AC-10: global defaults namingTemplate (L7) propagates.
			name:                 "AC-10: global-defaults-L7-naming",
			globalSMCfgDefaults:  SMConfigSection{NamingTemplate: tmplDefault},
			wantDefaultsTemplate: tmplDefault,
		},
		{
			// AC-11: global mandatory namingTemplate (L3) propagates.
			name:                  "AC-11: global-mandatory-L3-naming",
			globalSMCfgMandatory:  SMConfigSection{NamingTemplate: tmplMandatory},
			wantMandatoryTemplate: tmplMandatory,
		},
		{
			// L3 mandatory wins over L4.
			name:                  "L3-mandatory-wins-L4",
			globalSMCfgMandatory:  SMConfigSection{NamingTemplate: tmplMandatory},
			localSMCfgMandatory:   SMConfigSection{NamingTemplate: tmplLocal},
			wantMandatoryTemplate: tmplMandatory,
		},
		{
			// L6 defaults wins over L7 global.
			name:                 "L6-defaults-wins-L7",
			localSMCfgDefaults:   SMConfigSection{NamingTemplate: tmplLocal},
			globalSMCfgDefaults:  SMConfigSection{NamingTemplate: tmplDefault},
			wantDefaultsTemplate: tmplLocal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MergeSecretsManagerCascade(
				zeroSMKropath, zeroSMKropath,
				tc.globalSMCfgMandatory,
				tc.localSMCfgMandatory,
				tc.localSMCfgDefaults,
				tc.globalSMCfgDefaults,
				zeroSMKropath, zeroSMKropath,
			)
			if got.Mandatory.NamingTemplate != tc.wantMandatoryTemplate {
				t.Errorf("Mandatory.NamingTemplate = %q, want %q",
					got.Mandatory.NamingTemplate, tc.wantMandatoryTemplate)
			}
			if got.Defaults.NamingTemplate != tc.wantDefaultsTemplate {
				t.Errorf("Defaults.NamingTemplate = %q, want %q",
					got.Defaults.NamingTemplate, tc.wantDefaultsTemplate)
			}
		})
	}
}

func TestMergeSecretsManagerCascade_Tags(t *testing.T) {
	tests := []struct {
		name                   string
		globalKropathMandatory SMKropathSection
		localKropathMandatory  SMKropathSection
		globalSMCfgMandatory   SMConfigSection
		localSMCfgMandatory    SMConfigSection
		localSMCfgDefaults     SMConfigSection
		globalSMCfgDefaults    SMConfigSection
		localKropathDefaults   SMKropathSection
		globalKropathDefaults  SMKropathSection
		wantMandatoryTags      map[string]string
		wantDefaultsTags       map[string]string
	}{
		{
			// AC-12: KropathConfig.mandatory.tags union-merged with SecretsManagerConfig.mandatory.tags.
			name: "AC-12: tag-union-merge-mandatory",
			globalKropathMandatory: SMKropathSection{
				Tags: map[string]string{"cost-centre": "security"},
			},
			globalSMCfgMandatory: SMConfigSection{
				Tags: map[string]string{"resource-type": "secret"},
			},
			wantMandatoryTags: map[string]string{
				"cost-centre":   "security",
				"resource-type": "secret",
			},
		},
		{
			// L1 KropathConfig.mandatory.tags win on key conflict with L3 SMCfg.mandatory.tags.
			name: "kpc-L1-tags-win-on-conflict",
			globalKropathMandatory: SMKropathSection{
				Tags: map[string]string{"env": "production"},
			},
			globalSMCfgMandatory: SMConfigSection{
				Tags: map[string]string{"env": "dev"},
			},
			wantMandatoryTags: map[string]string{"env": "production"},
		},
		{
			// Defaults tags union merge: L6 wins over L7.
			name: "defaults-tags-L6-wins-L7",
			localSMCfgDefaults: SMConfigSection{
				Tags: map[string]string{"tier": "premium"},
			},
			globalSMCfgDefaults: SMConfigSection{
				Tags: map[string]string{"tier": "standard", "region": "ap-southeast-2"},
			},
			wantDefaultsTags: map[string]string{
				"tier":   "premium",
				"region": "ap-southeast-2",
			},
		},
		{
			// Global KropathConfig defaults tags (L9) propagate.
			name: "global-kpc-defaults-L9-tags",
			globalKropathDefaults: SMKropathSection{
				Tags: map[string]string{"managed-by": "kropath"},
			},
			wantDefaultsTags: map[string]string{"managed-by": "kropath"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MergeSecretsManagerCascade(
				tc.globalKropathMandatory,
				tc.localKropathMandatory,
				tc.globalSMCfgMandatory,
				tc.localSMCfgMandatory,
				tc.localSMCfgDefaults,
				tc.globalSMCfgDefaults,
				tc.localKropathDefaults,
				tc.globalKropathDefaults,
			)
			if !reflect.DeepEqual(got.Mandatory.Tags, tc.wantMandatoryTags) {
				t.Errorf("Mandatory.Tags = %v, want %v", got.Mandatory.Tags, tc.wantMandatoryTags)
			}
			if !reflect.DeepEqual(got.Defaults.Tags, tc.wantDefaultsTags) {
				t.Errorf("Defaults.Tags = %v, want %v", got.Defaults.Tags, tc.wantDefaultsTags)
			}
		})
	}
}

func TestMergeSecretsManagerCascade_SyncedLabels(t *testing.T) {
	tests := []struct {
		name                 string
		globalSMCfgMandatory SMConfigSection
		localSMCfgMandatory  SMConfigSection
		globalSMCfgDefaults  SMConfigSection
		localSMCfgDefaults   SMConfigSection
		wantMandatoryLabels  map[string]string
		wantDefaultsLabels   map[string]string
	}{
		{
			// AC-13: SecretsManagerConfig.mandatory.syncedLabels propagates.
			name: "AC-13: global-mandatory-syncedLabels",
			globalSMCfgMandatory: SMConfigSection{
				SyncedLabels: map[string]string{"data-class": "confidential"},
			},
			wantMandatoryLabels: map[string]string{"data-class": "confidential"},
		},
		{
			// L3 SyncedLabels win over L4 on key conflict.
			name: "L3-syncedLabels-win-L4",
			globalSMCfgMandatory: SMConfigSection{
				SyncedLabels: map[string]string{"team": "security"},
			},
			localSMCfgMandatory: SMConfigSection{
				SyncedLabels: map[string]string{"team": "platform"},
			},
			wantMandatoryLabels: map[string]string{"team": "security"},
		},
		{
			// Defaults syncedLabels: L6 wins over L7.
			name: "defaults-syncedLabels-L6-wins-L7",
			localSMCfgDefaults: SMConfigSection{
				SyncedLabels: map[string]string{"tier": "premium"},
			},
			globalSMCfgDefaults: SMConfigSection{
				SyncedLabels: map[string]string{"tier": "standard"},
			},
			wantDefaultsLabels: map[string]string{"tier": "premium"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MergeSecretsManagerCascade(
				zeroSMKropath, zeroSMKropath,
				tc.globalSMCfgMandatory,
				tc.localSMCfgMandatory,
				tc.localSMCfgDefaults,
				tc.globalSMCfgDefaults,
				zeroSMKropath, zeroSMKropath,
			)
			if !reflect.DeepEqual(got.Mandatory.SyncedLabels, tc.wantMandatoryLabels) {
				t.Errorf("Mandatory.SyncedLabels = %v, want %v",
					got.Mandatory.SyncedLabels, tc.wantMandatoryLabels)
			}
			if !reflect.DeepEqual(got.Defaults.SyncedLabels, tc.wantDefaultsLabels) {
				t.Errorf("Defaults.SyncedLabels = %v, want %v",
					got.Defaults.SyncedLabels, tc.wantDefaultsLabels)
			}
		})
	}
}

func TestMergeSecretsManagerCascade_AllZero(t *testing.T) {
	got := MergeSecretsManagerCascade(
		zeroSMKropath, zeroSMKropath,
		zeroSMConfig, zeroSMConfig,
		zeroSMConfig, zeroSMConfig,
		zeroSMKropath, zeroSMKropath,
	)
	if got.Mandatory.KmsKeyID != "" {
		t.Errorf("Mandatory.KmsKeyID = %q, want empty", got.Mandatory.KmsKeyID)
	}
	if len(got.Mandatory.ReplicaRegions) != 0 {
		t.Errorf("Mandatory.ReplicaRegions = %v, want nil/empty", got.Mandatory.ReplicaRegions)
	}
	if got.Mandatory.ForceOverwriteReplicaSecret {
		t.Error("Mandatory.ForceOverwriteReplicaSecret = true, want false")
	}
	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("Mandatory.NamingTemplate = %q, want empty", got.Mandatory.NamingTemplate)
	}
	if got.Defaults.KmsKeyID != "" {
		t.Errorf("Defaults.KmsKeyID = %q, want empty", got.Defaults.KmsKeyID)
	}
	if len(got.Defaults.ReplicaRegions) != 0 {
		t.Errorf("Defaults.ReplicaRegions = %v, want nil/empty", got.Defaults.ReplicaRegions)
	}
}

func TestFirstNonEmptyReplicaRegions(t *testing.T) {
	r1 := []ReplicaRegion{{Region: "us-east-1"}}
	r2 := []ReplicaRegion{{Region: "eu-west-1"}}

	tests := []struct {
		name       string
		candidates [][]ReplicaRegion
		want       []ReplicaRegion
	}{
		{
			name:       "all-nil-returns-nil",
			candidates: [][]ReplicaRegion{nil, nil},
			want:       nil,
		},
		{
			name:       "first-non-empty-wins",
			candidates: [][]ReplicaRegion{r1, r2},
			want:       r1,
		},
		{
			name:       "skips-nil-reaches-second",
			candidates: [][]ReplicaRegion{nil, r2},
			want:       r2,
		},
		{
			name:       "skips-empty-slice-reaches-second",
			candidates: [][]ReplicaRegion{{}, r2},
			want:       r2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := firstNonEmptyReplicaRegions(tc.candidates...)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("firstNonEmptyReplicaRegions = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestFirstNonEmptyReplicaRegions_DefensiveCopy(t *testing.T) {
	original := []ReplicaRegion{{Region: "us-east-1"}}
	got := firstNonEmptyReplicaRegions(original)

	// Mutate the original; the returned slice must be unaffected.
	original[0].Region = "mutated"
	if got[0].Region == "mutated" {
		t.Error("firstNonEmptyReplicaRegions did not return a defensive copy")
	}
}
