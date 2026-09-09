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

	"github.com/google/go-cmp/cmp"
)

func TestMergeEBSRecycleBinCascade(t *testing.T) {
	empty := EBSRecycleBinKropathSection{}
	emptyCfg := EBSRecycleBinConfigSection{}

	tests := []struct {
		name                   string
		globalKropathMandatory EBSRecycleBinKropathSection
		localKropathMandatory  EBSRecycleBinKropathSection
		globalRBCfgMandatory   EBSRecycleBinConfigSection
		localRBCfgMandatory    EBSRecycleBinConfigSection
		localRBCfgDefaults     EBSRecycleBinConfigSection
		globalRBCfgDefaults    EBSRecycleBinConfigSection
		localKropathDefaults   EBSRecycleBinKropathSection
		globalKropathDefaults  EBSRecycleBinKropathSection
		want                   EffectiveEBSRecycleBinConfig
	}{
		{
			name: "all empty — zero effective config",
			want: EffectiveEBSRecycleBinConfig{},
		},
		{
			// AC-11: mandatory governance fields written to effectiveConfig.mandatory
			name: "local RecycleBinConfig mandatory sets retentionPeriodValue and retentionPeriodUnit",
			localRBCfgMandatory: EBSRecycleBinConfigSection{
				RetentionPeriodValue: 30,
				RetentionPeriodUnit:  "DAYS",
			},
			want: EffectiveEBSRecycleBinConfig{
				Mandatory: EffectiveEBSRecycleBinSection{
					RetentionPeriodValue: 30,
					RetentionPeriodUnit:  "DAYS",
				},
			},
		},
		{
			// AC-12: lockRule true from global KropathConfig overrides false at RecycleBinConfig level
			name: "global KropathConfig lockRule true wins over RecycleBinConfig false",
			globalKropathMandatory: EBSRecycleBinKropathSection{
				LockRule: true,
			},
			localRBCfgMandatory: EBSRecycleBinConfigSection{
				LockRule: false, // zero value / not enforced at this level
			},
			want: EffectiveEBSRecycleBinConfig{
				Mandatory: EffectiveEBSRecycleBinSection{
					LockRule: true,
				},
			},
		},
		{
			// KropathConfig L1 retentionPeriodValue beats RecycleBinConfig L4
			name: "global KropathConfig mandatory retentionPeriodValue beats local RecycleBinConfig",
			globalKropathMandatory: EBSRecycleBinKropathSection{
				RetentionPeriodValue: 7,
			},
			localRBCfgMandatory: EBSRecycleBinConfigSection{
				RetentionPeriodValue: 30,
			},
			want: EffectiveEBSRecycleBinConfig{
				Mandatory: EffectiveEBSRecycleBinSection{
					RetentionPeriodValue: 7,
				},
			},
		},
		{
			// AC-13: tags merge — KPC mandatory wins on key conflict
			name: "mandatory tags: KPC L1 wins over RecycleBinConfig L3 on key conflict",
			globalKropathMandatory: EBSRecycleBinKropathSection{
				Tags: map[string]string{"env": "prod", "cost-center": "kpc"},
			},
			globalRBCfgMandatory: EBSRecycleBinConfigSection{
				Tags: map[string]string{"env": "staging", "service": "recyclebin"},
			},
			want: EffectiveEBSRecycleBinConfig{
				Mandatory: EffectiveEBSRecycleBinSection{
					Tags: map[string]string{
						"env":         "prod",
						"cost-center": "kpc",
						"service":     "recyclebin",
					},
				},
			},
		},
		{
			// AC-14: syncedLabels from RecycleBinConfig.mandatory propagate
			name: "mandatory syncedLabels from global RecycleBinConfig propagate",
			globalRBCfgMandatory: EBSRecycleBinConfigSection{
				SyncedLabels: map[string]string{"team": "platform"},
			},
			want: EffectiveEBSRecycleBinConfig{
				Mandatory: EffectiveEBSRecycleBinSection{
					SyncedLabels: map[string]string{"team": "platform"},
				},
			},
		},
		{
			// syncedLabels: global RecycleBinConfig mandatory (L3) wins over local (L4) on key conflict
			// mergeMaps uses last-wins; L3 is passed last so it wins.
			name: "mandatory syncedLabels: global RecycleBinConfig wins over local on key conflict",
			localRBCfgMandatory: EBSRecycleBinConfigSection{
				SyncedLabels: map[string]string{"owner": "local-team"},
			},
			globalRBCfgMandatory: EBSRecycleBinConfigSection{
				SyncedLabels: map[string]string{"owner": "global-team", "extra": "value"},
			},
			// L3 wins on "owner"; "extra" merges in from L3 too
			want: EffectiveEBSRecycleBinConfig{
				Mandatory: EffectiveEBSRecycleBinSection{
					SyncedLabels: map[string]string{"owner": "global-team", "extra": "value"},
				},
			},
		},
		{
			// KropathConfig does NOT contribute to syncedLabels (RecycleBinConfig levels only)
			name: "KropathConfig syncedLabels not included in effectiveConfig",
			// KropathConfigTier has no syncedLabels field for RecycleBin, but the KropathSection
			// struct itself has none either. This test verifies nothing from KPC bleeds into syncedLabels.
			localRBCfgMandatory: EBSRecycleBinConfigSection{
				SyncedLabels: map[string]string{"rb-label": "rb-value"},
			},
			want: EffectiveEBSRecycleBinConfig{
				Mandatory: EffectiveEBSRecycleBinSection{
					SyncedLabels: map[string]string{"rb-label": "rb-value"},
				},
			},
		},
		{
			// defaults tier: local RecycleBinConfig (L6) wins over global KropathConfig (L9)
			name: "defaults: local RecycleBinConfig retentionPeriodValue beats global KropathConfig",
			localRBCfgDefaults: EBSRecycleBinConfigSection{
				RetentionPeriodValue: 14,
				RetentionPeriodUnit:  "DAYS",
			},
			globalKropathDefaults: EBSRecycleBinKropathSection{
				RetentionPeriodValue: 90,
			},
			want: EffectiveEBSRecycleBinConfig{
				Defaults: EffectiveEBSRecycleBinSection{
					RetentionPeriodValue: 14,
					RetentionPeriodUnit:  "DAYS",
				},
			},
		},
		{
			// defaults tags: L6 local wins over L9 global KPC on key conflict
			name: "defaults tags: local RecycleBinConfig wins over global KropathConfig on key conflict",
			localRBCfgDefaults: EBSRecycleBinConfigSection{
				Tags: map[string]string{"env": "dev"},
			},
			globalKropathDefaults: EBSRecycleBinKropathSection{
				Tags: map[string]string{"env": "shared", "org": "my-org"},
			},
			want: EffectiveEBSRecycleBinConfig{
				Defaults: EffectiveEBSRecycleBinSection{
					Tags: map[string]string{"env": "dev", "org": "my-org"},
				},
			},
		},
		{
			// unlockDelay fields: governed at RecycleBinConfig levels only
			name: "mandatory unlockDelayValue and unlockDelayUnit from global RecycleBinConfig",
			globalRBCfgMandatory: EBSRecycleBinConfigSection{
				UnlockDelayValue: 7,
				UnlockDelayUnit:  "DAYS",
			},
			want: EffectiveEBSRecycleBinConfig{
				Mandatory: EffectiveEBSRecycleBinSection{
					UnlockDelayValue: 7,
					UnlockDelayUnit:  "DAYS",
				},
			},
		},
		{
			// unlockDelay: local RecycleBinConfig mandatory (L4) beats global (L3)
			name: "mandatory unlockDelayValue: local RecycleBinConfig beats global",
			globalRBCfgMandatory: EBSRecycleBinConfigSection{
				UnlockDelayValue: 14,
				UnlockDelayUnit:  "DAYS",
			},
			localRBCfgMandatory: EBSRecycleBinConfigSection{
				UnlockDelayValue: 3,
				UnlockDelayUnit:  "DAYS",
			},
			// global (L3) = strongest mandatory RecycleBinConfig
			want: EffectiveEBSRecycleBinConfig{
				Mandatory: EffectiveEBSRecycleBinSection{
					UnlockDelayValue: 14,
					UnlockDelayUnit:  "DAYS",
				},
			},
		},
		{
			// lockRule in defaults tier: local RecycleBinConfig (L6) wins
			name: "defaults lockRule: local RecycleBinConfig true wins",
			localRBCfgDefaults: EBSRecycleBinConfigSection{
				LockRule: true,
			},
			want: EffectiveEBSRecycleBinConfig{
				Defaults: EffectiveEBSRecycleBinSection{
					LockRule: true,
				},
			},
		},
		{
			// syncedAnnotations: additive merge across RecycleBinConfig levels for defaults
			name: "defaults syncedAnnotations: global and local RecycleBinConfig merge additively",
			localRBCfgDefaults: EBSRecycleBinConfigSection{
				SyncedAnnotations: map[string]string{"anno-local": "val-local"},
			},
			globalRBCfgDefaults: EBSRecycleBinConfigSection{
				SyncedAnnotations: map[string]string{"anno-global": "val-global"},
			},
			want: EffectiveEBSRecycleBinConfig{
				Defaults: EffectiveEBSRecycleBinSection{
					SyncedAnnotations: map[string]string{
						"anno-local":  "val-local",
						"anno-global": "val-global",
					},
				},
			},
		},
		{
			// Full cascade: all levels populated, verify correct priority wins
			name: "full cascade: correct priority across all levels",
			globalKropathMandatory: EBSRecycleBinKropathSection{
				RetentionPeriodValue: 7,
				LockRule:             true,
				Tags:                 map[string]string{"tier": "kpc-global-mandatory"},
			},
			localKropathMandatory: EBSRecycleBinKropathSection{
				RetentionPeriodValue: 14,
				Tags:                 map[string]string{"local-kpc-tag": "yes"},
			},
			globalRBCfgMandatory: EBSRecycleBinConfigSection{
				RetentionPeriodValue: 30,
				RetentionPeriodUnit:  "DAYS",
				UnlockDelayValue:     5,
				UnlockDelayUnit:      "DAYS",
				SyncedLabels:         map[string]string{"global-rb-label": "v"},
				Tags:                 map[string]string{"global-rb-tag": "v"},
			},
			localRBCfgMandatory: EBSRecycleBinConfigSection{
				RetentionPeriodValue: 60,
				SyncedLabels:         map[string]string{"local-rb-label": "v"},
				Tags:                 map[string]string{"local-rb-tag": "v"},
			},
			localRBCfgDefaults: EBSRecycleBinConfigSection{
				RetentionPeriodValue: 1,
				RetentionPeriodUnit:  "DAYS",
				Tags:                 map[string]string{"local-rb-defaults-tag": "v"},
			},
			globalRBCfgDefaults: EBSRecycleBinConfigSection{
				RetentionPeriodValue: 2,
				Tags:                 map[string]string{"global-rb-defaults-tag": "v"},
			},
			localKropathDefaults: EBSRecycleBinKropathSection{
				RetentionPeriodValue: 3,
				Tags:                 map[string]string{"local-kpc-defaults-tag": "v"},
			},
			globalKropathDefaults: EBSRecycleBinKropathSection{
				RetentionPeriodValue: 4,
				Tags:                 map[string]string{"global-kpc-defaults-tag": "v"},
			},
			want: EffectiveEBSRecycleBinConfig{
				Mandatory: EffectiveEBSRecycleBinSection{
					RetentionPeriodValue: 7, // L1 wins
					RetentionPeriodUnit:  "DAYS",
					LockRule:             true, // L1 wins
					UnlockDelayValue:     5,
					UnlockDelayUnit:      "DAYS",
					SyncedLabels: map[string]string{
						"global-rb-label": "v",
						"local-rb-label":  "v",
					},
					Tags: map[string]string{
						"tier":          "kpc-global-mandatory",
						"local-kpc-tag": "yes",
						"global-rb-tag": "v",
						"local-rb-tag":  "v",
					},
				},
				Defaults: EffectiveEBSRecycleBinSection{
					RetentionPeriodValue: 1, // L6 wins
					RetentionPeriodUnit:  "DAYS",
					Tags: map[string]string{
						"local-rb-defaults-tag":  "v",
						"global-rb-defaults-tag": "v",
						"local-kpc-defaults-tag": "v",
						"global-kpc-defaults-tag": "v",
					},
				},
			},
		},
	}

	// Suppress "declared but not used" for empty variables used in table construction
	_ = empty
	_ = emptyCfg

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeEBSRecycleBinCascade(
				tt.globalKropathMandatory,
				tt.localKropathMandatory,
				tt.globalRBCfgMandatory,
				tt.localRBCfgMandatory,
				tt.localRBCfgDefaults,
				tt.globalRBCfgDefaults,
				tt.localKropathDefaults,
				tt.globalKropathDefaults,
			)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("MergeEBSRecycleBinCascade() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
