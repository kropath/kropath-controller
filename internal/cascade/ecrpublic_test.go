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

var (
	zeroECRPublicKropath = ECRPublicKropathSection{}
	zeroECRPublicConfig  = ECRPublicConfigSection{}
)

func mergeECRPublicAllZero() EffectiveECRPublicConfig {
	return MergeECRPublicCascade(
		zeroECRPublicKropath, zeroECRPublicKropath,
		zeroECRPublicConfig, zeroECRPublicConfig,
		zeroECRPublicConfig, zeroECRPublicConfig,
		zeroECRPublicKropath, zeroECRPublicKropath,
	)
}

func TestMergeECRPublicCascade_AllZero(t *testing.T) {
	got := mergeECRPublicAllZero()
	want := EffectiveECRPublicConfig{}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("all-zero inputs: got %+v, want %+v", got, want)
	}
}

func TestMergeECRPublicCascade_NamingTemplate(t *testing.T) {
	tests := []struct {
		name                    string
		globalECRPubCfgMandatory ECRPublicConfigSection
		localECRPubCfgMandatory  ECRPublicConfigSection
		localECRPubCfgDefaults   ECRPublicConfigSection
		globalECRPubCfgDefaults  ECRPublicConfigSection
		wantMandatoryNaming      string
		wantDefaultsNaming       string
	}{
		{
			// L3 globalECRPublicConfig.mandatory.namingTemplate propagates.
			name:                    "L3-global-mandatory-namingTemplate-propagates",
			globalECRPubCfgMandatory: ECRPublicConfigSection{NamingTemplate: "{configRef}/{namespace}/{name}"},
			wantMandatoryNaming:      "{configRef}/{namespace}/{name}",
		},
		{
			// L3 global mandatory wins over L4 local mandatory.
			name:                    "L3-wins-over-L4-mandatory",
			globalECRPubCfgMandatory: ECRPublicConfigSection{NamingTemplate: "{configRef}/{namespace}/{name}"},
			localECRPubCfgMandatory:  ECRPublicConfigSection{NamingTemplate: "{namespace}/{name}"},
			wantMandatoryNaming:      "{configRef}/{namespace}/{name}",
		},
		{
			// L4 local mandatory propagates when L3 empty.
			name:                   "L4-local-mandatory-propagates-when-L3-empty",
			localECRPubCfgMandatory: ECRPublicConfigSection{NamingTemplate: "{namespace}/{name}"},
			wantMandatoryNaming:    "{namespace}/{name}",
		},
		{
			// L6 localDefaults propagates to defaults tier.
			name:                   "L6-local-defaults-namingTemplate-propagates",
			localECRPubCfgDefaults:  ECRPublicConfigSection{NamingTemplate: "{namespace}/{name}"},
			wantDefaultsNaming:     "{namespace}/{name}",
		},
		{
			// L6 local defaults wins over L7 global defaults.
			name:                   "L6-wins-over-L7-defaults",
			localECRPubCfgDefaults:  ECRPublicConfigSection{NamingTemplate: "{namespace}/{name}"},
			globalECRPubCfgDefaults: ECRPublicConfigSection{NamingTemplate: "{configRef}/{name}"},
			wantDefaultsNaming:     "{namespace}/{name}",
		},
		{
			// L7 global defaults propagates when L6 empty.
			name:                    "L7-global-defaults-propagates-when-L6-empty",
			globalECRPubCfgDefaults:  ECRPublicConfigSection{NamingTemplate: "{namespace}/{name}"},
			wantDefaultsNaming:      "{namespace}/{name}",
		},
		{
			// Mandatory set, defaults empty — only mandatory populated.
			name:                    "mandatory-set-defaults-empty",
			globalECRPubCfgMandatory: ECRPublicConfigSection{NamingTemplate: "{configRef}/{namespace}/{name}"},
			wantMandatoryNaming:      "{configRef}/{namespace}/{name}",
			wantDefaultsNaming:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeECRPublicCascade(
				zeroECRPublicKropath, zeroECRPublicKropath,
				tt.globalECRPubCfgMandatory, tt.localECRPubCfgMandatory,
				tt.localECRPubCfgDefaults, tt.globalECRPubCfgDefaults,
				zeroECRPublicKropath, zeroECRPublicKropath,
			)
			if got.Mandatory.NamingTemplate != tt.wantMandatoryNaming {
				t.Errorf("Mandatory.NamingTemplate: got %q, want %q", got.Mandatory.NamingTemplate, tt.wantMandatoryNaming)
			}
			if got.Defaults.NamingTemplate != tt.wantDefaultsNaming {
				t.Errorf("Defaults.NamingTemplate: got %q, want %q", got.Defaults.NamingTemplate, tt.wantDefaultsNaming)
			}
		})
	}
}

func TestMergeECRPublicCascade_Tags(t *testing.T) {
	tests := []struct {
		name                    string
		globalKropathMandatory  ECRPublicKropathSection
		localKropathMandatory   ECRPublicKropathSection
		globalECRPubMandatory   ECRPublicConfigSection
		localECRPubMandatory    ECRPublicConfigSection
		localECRPubDefaults     ECRPublicConfigSection
		globalECRPubDefaults    ECRPublicConfigSection
		localKropathDefaults    ECRPublicKropathSection
		globalKropathDefaults   ECRPublicKropathSection
		wantMandatoryTags       map[string]string
		wantDefaultsTags        map[string]string
	}{
		{
			// L1 globalKropathMandatory.Tags propagate to mandatory.
			name:                   "L1-global-kpc-mandatory-tags-propagate",
			globalKropathMandatory: ECRPublicKropathSection{Tags: map[string]string{"org": "my-company"}},
			wantMandatoryTags:      map[string]string{"org": "my-company"},
		},
		{
			// L1 global KPC mandatory wins over L3 global ECRPublicConfig mandatory on key conflict.
			name:                   "L1-kpc-wins-over-L3-ecrcfg-key-conflict",
			globalKropathMandatory: ECRPublicKropathSection{Tags: map[string]string{"env": "prod"}},
			globalECRPubMandatory:  ECRPublicConfigSection{Tags: map[string]string{"env": "dev"}},
			wantMandatoryTags:      map[string]string{"env": "prod"},
		},
		{
			// Additive merge: KPC and ECRPublicConfig tags combined when keys differ.
			name:                   "additive-merge-kpc-and-ecrcfg-tags",
			globalKropathMandatory: ECRPublicKropathSection{Tags: map[string]string{"org": "my-company"}},
			globalECRPubMandatory:  ECRPublicConfigSection{Tags: map[string]string{"team": "platform"}},
			wantMandatoryTags:      map[string]string{"org": "my-company", "team": "platform"},
		},
		{
			// L2 local KPC mandatory wins over L3 global ECRPublicConfig mandatory.
			name:                  "L2-local-kpc-wins-over-L3",
			localKropathMandatory: ECRPublicKropathSection{Tags: map[string]string{"env": "staging"}},
			globalECRPubMandatory: ECRPublicConfigSection{Tags: map[string]string{"env": "prod"}},
			wantMandatoryTags:     map[string]string{"env": "staging"},
		},
		{
			// L3 global ECRPublicConfig mandatory wins over L4 local on key conflict.
			name:                  "L3-global-ecrcfg-wins-over-L4-local",
			globalECRPubMandatory: ECRPublicConfigSection{Tags: map[string]string{"visibility": "public"}},
			localECRPubMandatory:  ECRPublicConfigSection{Tags: map[string]string{"visibility": "private"}},
			wantMandatoryTags:     map[string]string{"visibility": "public"},
		},
		{
			// L6 local ECRPublicConfig defaults propagate to defaults.
			name:                "L6-local-ecrcfg-defaults-tags-propagate",
			localECRPubDefaults: ECRPublicConfigSection{Tags: map[string]string{"team": "backend"}},
			wantDefaultsTags:    map[string]string{"team": "backend"},
		},
		{
			// L6 local defaults wins over L9 global KPC defaults on key conflict.
			name:                   "L6-local-ecrcfg-wins-over-L9-kpc-defaults",
			localECRPubDefaults:    ECRPublicConfigSection{Tags: map[string]string{"cost-center": "eng"}},
			globalKropathDefaults:  ECRPublicKropathSection{Tags: map[string]string{"cost-center": "default"}},
			wantDefaultsTags:       map[string]string{"cost-center": "eng"},
		},
		{
			// L9 global KPC defaults propagate when all higher defaults empty.
			name:                  "L9-global-kpc-defaults-propagate",
			globalKropathDefaults: ECRPublicKropathSection{Tags: map[string]string{"managed-by": "kropath"}},
			wantDefaultsTags:      map[string]string{"managed-by": "kropath"},
		},
		{
			// All four mandatory levels combined additively.
			name: "all-four-mandatory-levels-additive",
			globalKropathMandatory: ECRPublicKropathSection{Tags: map[string]string{"l1": "kpc-global"}},
			localKropathMandatory:  ECRPublicKropathSection{Tags: map[string]string{"l2": "kpc-local"}},
			globalECRPubMandatory:  ECRPublicConfigSection{Tags: map[string]string{"l3": "ecrcfg-global"}},
			localECRPubMandatory:   ECRPublicConfigSection{Tags: map[string]string{"l4": "ecrcfg-local"}},
			wantMandatoryTags: map[string]string{
				"l1": "kpc-global",
				"l2": "kpc-local",
				"l3": "ecrcfg-global",
				"l4": "ecrcfg-local",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeECRPublicCascade(
				tt.globalKropathMandatory, tt.localKropathMandatory,
				tt.globalECRPubMandatory, tt.localECRPubMandatory,
				tt.localECRPubDefaults, tt.globalECRPubDefaults,
				tt.localKropathDefaults, tt.globalKropathDefaults,
			)
			if !reflect.DeepEqual(got.Mandatory.Tags, tt.wantMandatoryTags) {
				t.Errorf("Mandatory.Tags: got %v, want %v", got.Mandatory.Tags, tt.wantMandatoryTags)
			}
			if !reflect.DeepEqual(got.Defaults.Tags, tt.wantDefaultsTags) {
				t.Errorf("Defaults.Tags: got %v, want %v", got.Defaults.Tags, tt.wantDefaultsTags)
			}
		})
	}
}

func TestMergeECRPublicCascade_SyncedLabels(t *testing.T) {
	tests := []struct {
		name                   string
		globalECRPubMandatory  ECRPublicConfigSection
		localECRPubMandatory   ECRPublicConfigSection
		localECRPubDefaults    ECRPublicConfigSection
		globalECRPubDefaults   ECRPublicConfigSection
		wantMandatoryLabels    map[string]string
		wantDefaultsLabels     map[string]string
	}{
		{
			// L3 global ECRPublicConfig mandatory syncedLabels propagate.
			name:                  "L3-global-mandatory-syncedLabels-propagate",
			globalECRPubMandatory: ECRPublicConfigSection{SyncedLabels: map[string]string{"app": "my-app"}},
			wantMandatoryLabels:   map[string]string{"app": "my-app"},
		},
		{
			// L3 wins over L4 on key conflict.
			name:                  "L3-wins-over-L4-syncedLabels",
			globalECRPubMandatory: ECRPublicConfigSection{SyncedLabels: map[string]string{"tier": "global"}},
			localECRPubMandatory:  ECRPublicConfigSection{SyncedLabels: map[string]string{"tier": "local"}},
			wantMandatoryLabels:   map[string]string{"tier": "global"},
		},
		{
			// Additive merge of SyncedLabels from L3 and L4 when keys differ.
			name:                  "additive-merge-L3-L4-syncedLabels",
			globalECRPubMandatory: ECRPublicConfigSection{SyncedLabels: map[string]string{"from-global": "yes"}},
			localECRPubMandatory:  ECRPublicConfigSection{SyncedLabels: map[string]string{"from-local": "yes"}},
			wantMandatoryLabels:   map[string]string{"from-global": "yes", "from-local": "yes"},
		},
		{
			// L6 local defaults syncedLabels propagate.
			name:                "L6-local-defaults-syncedLabels",
			localECRPubDefaults: ECRPublicConfigSection{SyncedLabels: map[string]string{"default-label": "v1"}},
			wantDefaultsLabels:  map[string]string{"default-label": "v1"},
		},
		{
			// L6 wins over L7 in defaults.
			name:                 "L6-wins-over-L7-syncedLabels",
			localECRPubDefaults:  ECRPublicConfigSection{SyncedLabels: map[string]string{"tier": "local"}},
			globalECRPubDefaults: ECRPublicConfigSection{SyncedLabels: map[string]string{"tier": "global"}},
			wantDefaultsLabels:   map[string]string{"tier": "local"},
		},
		{
			// KropathConfig does NOT contribute to syncedLabels.
			name:                  "kpc-does-not-contribute-syncedLabels",
			globalECRPubMandatory: ECRPublicConfigSection{},
			wantMandatoryLabels:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeECRPublicCascade(
				zeroECRPublicKropath, zeroECRPublicKropath,
				tt.globalECRPubMandatory, tt.localECRPubMandatory,
				tt.localECRPubDefaults, tt.globalECRPubDefaults,
				zeroECRPublicKropath, zeroECRPublicKropath,
			)
			if !reflect.DeepEqual(got.Mandatory.SyncedLabels, tt.wantMandatoryLabels) {
				t.Errorf("Mandatory.SyncedLabels: got %v, want %v", got.Mandatory.SyncedLabels, tt.wantMandatoryLabels)
			}
			if !reflect.DeepEqual(got.Defaults.SyncedLabels, tt.wantDefaultsLabels) {
				t.Errorf("Defaults.SyncedLabels: got %v, want %v", got.Defaults.SyncedLabels, tt.wantDefaultsLabels)
			}
		})
	}
}

func TestMergeECRPublicCascade_SyncedAnnotations(t *testing.T) {
	tests := []struct {
		name                    string
		globalECRPubMandatory   ECRPublicConfigSection
		localECRPubMandatory    ECRPublicConfigSection
		localECRPubDefaults     ECRPublicConfigSection
		globalECRPubDefaults    ECRPublicConfigSection
		wantMandatoryAnnotations map[string]string
		wantDefaultsAnnotations  map[string]string
	}{
		{
			// L3 global mandatory syncedAnnotations propagate.
			name:                    "L3-global-mandatory-syncedAnnotations-propagate",
			globalECRPubMandatory:   ECRPublicConfigSection{SyncedAnnotations: map[string]string{"team": "platform"}},
			wantMandatoryAnnotations: map[string]string{"team": "platform"},
		},
		{
			// L3 wins over L4 on key conflict.
			name:                    "L3-wins-over-L4-syncedAnnotations",
			globalECRPubMandatory:   ECRPublicConfigSection{SyncedAnnotations: map[string]string{"tier": "global"}},
			localECRPubMandatory:    ECRPublicConfigSection{SyncedAnnotations: map[string]string{"tier": "local"}},
			wantMandatoryAnnotations: map[string]string{"tier": "global"},
		},
		{
			// L6 local defaults syncedAnnotations propagate.
			name:                   "L6-local-defaults-syncedAnnotations",
			localECRPubDefaults:    ECRPublicConfigSection{SyncedAnnotations: map[string]string{"cost": "low"}},
			wantDefaultsAnnotations: map[string]string{"cost": "low"},
		},
		{
			// Additive merge of SyncedAnnotations from L3 and L4 when keys differ.
			name:                    "additive-merge-L3-L4-syncedAnnotations",
			globalECRPubMandatory:   ECRPublicConfigSection{SyncedAnnotations: map[string]string{"from-global": "yes"}},
			localECRPubMandatory:    ECRPublicConfigSection{SyncedAnnotations: map[string]string{"from-local": "yes"}},
			wantMandatoryAnnotations: map[string]string{"from-global": "yes", "from-local": "yes"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeECRPublicCascade(
				zeroECRPublicKropath, zeroECRPublicKropath,
				tt.globalECRPubMandatory, tt.localECRPubMandatory,
				tt.localECRPubDefaults, tt.globalECRPubDefaults,
				zeroECRPublicKropath, zeroECRPublicKropath,
			)
			if !reflect.DeepEqual(got.Mandatory.SyncedAnnotations, tt.wantMandatoryAnnotations) {
				t.Errorf("Mandatory.SyncedAnnotations: got %v, want %v", got.Mandatory.SyncedAnnotations, tt.wantMandatoryAnnotations)
			}
			if !reflect.DeepEqual(got.Defaults.SyncedAnnotations, tt.wantDefaultsAnnotations) {
				t.Errorf("Defaults.SyncedAnnotations: got %v, want %v", got.Defaults.SyncedAnnotations, tt.wantDefaultsAnnotations)
			}
		})
	}
}

func TestMergeECRPublicCascade_Combined(t *testing.T) {
	// Full end-to-end test with all fields populated.
	got := MergeECRPublicCascade(
		ECRPublicKropathSection{Tags: map[string]string{"org": "my-company"}}, // L1
		ECRPublicKropathSection{Tags: map[string]string{"env": "prod"}},        // L2
		ECRPublicConfigSection{ // L3 global ECRPublicConfig mandatory
			NamingTemplate:    "{configRef}/{namespace}/{name}",
			Tags:              map[string]string{"visibility": "public"},
			SyncedLabels:      map[string]string{"app": "ecrpub"},
			SyncedAnnotations: map[string]string{"team": "platform"},
		},
		ECRPublicConfigSection{ // L4 local ECRPublicConfig mandatory
			Tags:         map[string]string{"local-mandatory": "true"},
			SyncedLabels: map[string]string{"local-label": "yes"},
		},
		ECRPublicConfigSection{ // L6 local ECRPublicConfig defaults
			NamingTemplate: "{namespace}/{name}",
			Tags:           map[string]string{"default-tag": "local"},
		},
		ECRPublicConfigSection{ // L7 global ECRPublicConfig defaults
			Tags: map[string]string{"default-tag": "global"},
		},
		ECRPublicKropathSection{Tags: map[string]string{"kpc-defaults-local": "yes"}}, // L8
		ECRPublicKropathSection{Tags: map[string]string{"managed-by": "kropath"}},     // L9
	)

	// Mandatory checks
	if got.Mandatory.NamingTemplate != "{configRef}/{namespace}/{name}" {
		t.Errorf("Mandatory.NamingTemplate: got %q", got.Mandatory.NamingTemplate)
	}
	wantMandatoryTags := map[string]string{
		"org":             "my-company",  // L1
		"env":             "prod",        // L2
		"visibility":      "public",      // L3
		"local-mandatory": "true",        // L4
	}
	if !reflect.DeepEqual(got.Mandatory.Tags, wantMandatoryTags) {
		t.Errorf("Mandatory.Tags: got %v, want %v", got.Mandatory.Tags, wantMandatoryTags)
	}
	wantMandatoryLabels := map[string]string{
		"app":         "ecrpub", // L3
		"local-label": "yes",   // L4
	}
	if !reflect.DeepEqual(got.Mandatory.SyncedLabels, wantMandatoryLabels) {
		t.Errorf("Mandatory.SyncedLabels: got %v, want %v", got.Mandatory.SyncedLabels, wantMandatoryLabels)
	}
	wantMandatoryAnnotations := map[string]string{"team": "platform"}
	if !reflect.DeepEqual(got.Mandatory.SyncedAnnotations, wantMandatoryAnnotations) {
		t.Errorf("Mandatory.SyncedAnnotations: got %v, want %v", got.Mandatory.SyncedAnnotations, wantMandatoryAnnotations)
	}

	// Defaults checks
	if got.Defaults.NamingTemplate != "{namespace}/{name}" {
		t.Errorf("Defaults.NamingTemplate: got %q", got.Defaults.NamingTemplate)
	}
	wantDefaultsTags := map[string]string{
		"default-tag":        "local",   // L6 wins over L7
		"kpc-defaults-local": "yes",     // L8
		"managed-by":         "kropath", // L9
	}
	if !reflect.DeepEqual(got.Defaults.Tags, wantDefaultsTags) {
		t.Errorf("Defaults.Tags: got %v, want %v", got.Defaults.Tags, wantDefaultsTags)
	}
}
