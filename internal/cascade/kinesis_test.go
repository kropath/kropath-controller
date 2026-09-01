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

func TestMergeKinesisCascade_KropathMandatoryWinsOnStreamMode(t *testing.T) {
	// Level 1 (globalKropathMandatory) wins over all lower levels.
	result := MergeKinesisCascade(
		KinesisKropathSection{StreamMode: "on_demand"}, // level 1 — wins
		KinesisKropathSection{StreamMode: "provisioned"}, // level 2
		KinesisConfigSection{StreamMode: "provisioned"},   // level 3
		KinesisConfigSection{StreamMode: "provisioned"},   // level 4
		KinesisConfigSection{},                            // level 6
		KinesisConfigSection{},                            // level 7
		KinesisKropathSection{},                           // level 8
		KinesisKropathSection{},                           // level 9
	)
	if result.Mandatory.StreamMode != "on_demand" {
		t.Errorf("expected mandatory.streamMode=on_demand, got %q", result.Mandatory.StreamMode)
	}
}

func TestMergeKinesisCascade_LocalKropathMandatoryWinsOverKinesisCfg(t *testing.T) {
	// Level 2 (localKropathMandatory) wins over KinesisConfig mandatory levels (3-4).
	result := MergeKinesisCascade(
		KinesisKropathSection{},                           // level 1 — empty
		KinesisKropathSection{StreamMode: "on_demand"},   // level 2 — wins
		KinesisConfigSection{StreamMode: "provisioned"},   // level 3
		KinesisConfigSection{StreamMode: "provisioned"},   // level 4
		KinesisConfigSection{},
		KinesisConfigSection{},
		KinesisKropathSection{},
		KinesisKropathSection{},
	)
	if result.Mandatory.StreamMode != "on_demand" {
		t.Errorf("expected mandatory.streamMode=on_demand, got %q", result.Mandatory.StreamMode)
	}
}

func TestMergeKinesisCascade_GlobalKinesisCfgMandatoryWinsOverLocal(t *testing.T) {
	// Level 3 (globalKinesisCfgMandatory) wins over level 4 (localKinesisCfgMandatory).
	result := MergeKinesisCascade(
		KinesisKropathSection{},                          // level 1
		KinesisKropathSection{},                          // level 2
		KinesisConfigSection{StreamMode: "on_demand"},    // level 3 — wins
		KinesisConfigSection{StreamMode: "provisioned"},  // level 4
		KinesisConfigSection{},
		KinesisConfigSection{},
		KinesisKropathSection{},
		KinesisKropathSection{},
	)
	if result.Mandatory.StreamMode != "on_demand" {
		t.Errorf("expected mandatory.streamMode=on_demand, got %q", result.Mandatory.StreamMode)
	}
}

func TestMergeKinesisCascade_LocalKinesisCfgDefaultsWinsOverGlobal(t *testing.T) {
	// Level 6 (localKinesisCfgDefaults) wins over level 7 (globalKinesisCfgDefaults).
	result := MergeKinesisCascade(
		KinesisKropathSection{},
		KinesisKropathSection{},
		KinesisConfigSection{},
		KinesisConfigSection{},
		KinesisConfigSection{StreamMode: "on_demand"},   // level 6 — wins
		KinesisConfigSection{StreamMode: "provisioned"}, // level 7
		KinesisKropathSection{},
		KinesisKropathSection{},
	)
	if result.Defaults.StreamMode != "on_demand" {
		t.Errorf("expected defaults.streamMode=on_demand, got %q", result.Defaults.StreamMode)
	}
}

func TestMergeKinesisCascade_GlobalKropathDefaultsIsWeakest(t *testing.T) {
	// Level 9 (globalKropathDefaults) only applies when all other defaults levels are empty.
	result := MergeKinesisCascade(
		KinesisKropathSection{},
		KinesisKropathSection{},
		KinesisConfigSection{},
		KinesisConfigSection{},
		KinesisConfigSection{},
		KinesisConfigSection{},
		KinesisKropathSection{},
		KinesisKropathSection{StreamMode: "on_demand"}, // level 9 — weakest, only source
	)
	if result.Defaults.StreamMode != "on_demand" {
		t.Errorf("expected defaults.streamMode=on_demand, got %q", result.Defaults.StreamMode)
	}
}

func TestMergeKinesisCascade_ShardCountMandatory_Level1Wins(t *testing.T) {
	// Level 1 (globalKropathMandatory) shardCount wins over all lower levels.
	result := MergeKinesisCascade(
		KinesisKropathSection{ShardCount: 10}, // level 1 — wins
		KinesisKropathSection{ShardCount: 4},
		KinesisConfigSection{ShardCount: 2},
		KinesisConfigSection{ShardCount: 1},
		KinesisConfigSection{},
		KinesisConfigSection{},
		KinesisKropathSection{},
		KinesisKropathSection{},
	)
	if result.Mandatory.ShardCount != 10 {
		t.Errorf("expected mandatory.shardCount=10, got %d", result.Mandatory.ShardCount)
	}
}

func TestMergeKinesisCascade_ShardCountMandatory_KropathTier2Wins(t *testing.T) {
	// Level 2 (localKropathMandatory) shardCount wins over KinesisConfig levels.
	result := MergeKinesisCascade(
		KinesisKropathSection{},                   // level 1 — empty (0)
		KinesisKropathSection{ShardCount: 8},      // level 2 — wins
		KinesisConfigSection{ShardCount: 4},        // level 3
		KinesisConfigSection{ShardCount: 2},        // level 4
		KinesisConfigSection{},
		KinesisConfigSection{},
		KinesisKropathSection{},
		KinesisKropathSection{},
	)
	if result.Mandatory.ShardCount != 8 {
		t.Errorf("expected mandatory.shardCount=8, got %d", result.Mandatory.ShardCount)
	}
}

func TestMergeKinesisCascade_ShardCountDefaults_LocalCfgWinsOverGlobal(t *testing.T) {
	// Level 6 (localKinesisCfgDefaults) shardCount wins over level 7.
	result := MergeKinesisCascade(
		KinesisKropathSection{},
		KinesisKropathSection{},
		KinesisConfigSection{},
		KinesisConfigSection{},
		KinesisConfigSection{ShardCount: 4}, // level 6 — wins
		KinesisConfigSection{ShardCount: 2}, // level 7
		KinesisKropathSection{},
		KinesisKropathSection{},
	)
	if result.Defaults.ShardCount != 4 {
		t.Errorf("expected defaults.shardCount=4, got %d", result.Defaults.ShardCount)
	}
}

func TestMergeKinesisCascade_ShardCountDefaults_KropathDefaultsIsWeakest(t *testing.T) {
	// Level 9 shardCount applies only when all stronger defaults are 0.
	result := MergeKinesisCascade(
		KinesisKropathSection{},
		KinesisKropathSection{},
		KinesisConfigSection{},
		KinesisConfigSection{},
		KinesisConfigSection{},
		KinesisConfigSection{},
		KinesisKropathSection{},
		KinesisKropathSection{ShardCount: 2}, // level 9 — weakest
	)
	if result.Defaults.ShardCount != 2 {
		t.Errorf("expected defaults.shardCount=2, got %d", result.Defaults.ShardCount)
	}
}

func TestMergeKinesisCascade_NamingTemplateNoKropathLevels(t *testing.T) {
	// namingTemplate has no KropathConfig.kinesis equivalent — only KinesisConfig levels 3-4 apply.
	result := MergeKinesisCascade(
		KinesisKropathSection{}, // level 1 — no namingTemplate field
		KinesisKropathSection{}, // level 2
		KinesisConfigSection{NamingTemplate: "corp-{name}"},        // level 3 — wins
		KinesisConfigSection{NamingTemplate: "{namespace}-{name}"}, // level 4
		KinesisConfigSection{},
		KinesisConfigSection{},
		KinesisKropathSection{},
		KinesisKropathSection{},
	)
	if result.Mandatory.NamingTemplate != "corp-{name}" {
		t.Errorf("expected corp-{name}, got %q", result.Mandatory.NamingTemplate)
	}
}

func TestMergeKinesisCascade_NamingTemplateDefaultsFallthrough(t *testing.T) {
	// Level 6 local config defaults.namingTemplate wins over level 7 global.
	result := MergeKinesisCascade(
		KinesisKropathSection{},
		KinesisKropathSection{},
		KinesisConfigSection{},
		KinesisConfigSection{},
		KinesisConfigSection{NamingTemplate: "{namespace}-{name}"}, // level 6
		KinesisConfigSection{NamingTemplate: "global-{name}"},       // level 7
		KinesisKropathSection{},
		KinesisKropathSection{},
	)
	if result.Defaults.NamingTemplate != "{namespace}-{name}" {
		t.Errorf("expected {namespace}-{name}, got %q", result.Defaults.NamingTemplate)
	}
}

func TestMergeKinesisCascade_TagsMergeAcrossAllMandatoryLevels(t *testing.T) {
	// Mandatory tags: union of KropathConfig.mandatory.tags (generic) + KinesisConfig mandatory tiers.
	// Higher priority levels win on key conflict.
	result := MergeKinesisCascade(
		KinesisKropathSection{Tags: map[string]string{"cost-centre": "infra", "shared": "yes"}}, // level 1 — wins on conflict
		KinesisKropathSection{Tags: map[string]string{"cost-centre": "dev"}},                    // level 2 — overridden
		KinesisConfigSection{Tags: map[string]string{"stream-type": "kinesis"}},                 // level 3
		KinesisConfigSection{Tags: map[string]string{"team": "platform"}},                       // level 4
		KinesisConfigSection{},
		KinesisConfigSection{},
		KinesisKropathSection{},
		KinesisKropathSection{},
	)
	want := map[string]string{
		"cost-centre": "infra",
		"shared":      "yes",
		"stream-type": "kinesis",
		"team":        "platform",
	}
	for k, v := range want {
		if result.Mandatory.Tags[k] != v {
			t.Errorf("mandatory.tags[%q]: expected %q, got %q", k, v, result.Mandatory.Tags[k])
		}
	}
}

func TestMergeKinesisCascade_TagsMergeAcrossAllDefaultsLevels(t *testing.T) {
	// Defaults tags: union of KinesisConfig defaults + KropathConfig.defaults.tags (generic).
	// Lower level number (stronger) wins on key conflict.
	result := MergeKinesisCascade(
		KinesisKropathSection{},
		KinesisKropathSection{},
		KinesisConfigSection{},
		KinesisConfigSection{},
		KinesisConfigSection{Tags: map[string]string{"team": "streaming", "env": "prod"}}, // level 6 — wins
		KinesisConfigSection{Tags: map[string]string{"team": "platform"}},                  // level 7
		KinesisKropathSection{Tags: map[string]string{"org": "acme"}},                      // level 8
		KinesisKropathSection{Tags: map[string]string{"org": "corp"}},                      // level 9 — weakest
	)
	if result.Defaults.Tags["team"] != "streaming" {
		t.Errorf("defaults.tags[team]: expected streaming, got %q", result.Defaults.Tags["team"])
	}
	if result.Defaults.Tags["env"] != "prod" {
		t.Errorf("defaults.tags[env]: expected prod, got %q", result.Defaults.Tags["env"])
	}
	if result.Defaults.Tags["org"] != "acme" {
		t.Errorf("defaults.tags[org]: expected acme (level 8 wins), got %q", result.Defaults.Tags["org"])
	}
}

func TestMergeKinesisCascade_SyncedLabelsMandatory_KinesisCfgLevelsOnly(t *testing.T) {
	// SyncedLabels for mandatory: union of globalKinesisCfgMandatory (L3) + localKinesisCfgMandatory (L4).
	// L3 wins on key conflict. KropathConfig levels do not contribute to syncedLabels.
	result := MergeKinesisCascade(
		KinesisKropathSection{}, // level 1 — no syncedLabels field
		KinesisKropathSection{}, // level 2
		KinesisConfigSection{SyncedLabels: map[string]string{"data-class": "streaming", "region": "ap"}}, // level 3 wins
		KinesisConfigSection{SyncedLabels: map[string]string{"data-class": "internal", "team": "platform"}}, // level 4
		KinesisConfigSection{},
		KinesisConfigSection{},
		KinesisKropathSection{},
		KinesisKropathSection{},
	)
	if result.Mandatory.SyncedLabels["data-class"] != "streaming" {
		t.Errorf("expected data-class=streaming (L3 wins), got %q", result.Mandatory.SyncedLabels["data-class"])
	}
	if result.Mandatory.SyncedLabels["team"] != "platform" {
		t.Errorf("expected team=platform, got %q", result.Mandatory.SyncedLabels["team"])
	}
	if result.Mandatory.SyncedLabels["region"] != "ap" {
		t.Errorf("expected region=ap, got %q", result.Mandatory.SyncedLabels["region"])
	}
}

func TestMergeKinesisCascade_SyncedLabelsDefaults_KinesisCfgLevelsOnly(t *testing.T) {
	// SyncedLabels for defaults: union of globalKinesisCfgDefaults (L7) + localKinesisCfgDefaults (L6).
	// L6 wins on key conflict.
	result := MergeKinesisCascade(
		KinesisKropathSection{},
		KinesisKropathSection{},
		KinesisConfigSection{},
		KinesisConfigSection{},
		KinesisConfigSection{SyncedLabels: map[string]string{"data-class": "internal", "env": "prod"}}, // level 6 — wins
		KinesisConfigSection{SyncedLabels: map[string]string{"data-class": "public"}},                  // level 7
		KinesisKropathSection{},
		KinesisKropathSection{},
	)
	if result.Defaults.SyncedLabels["data-class"] != "internal" {
		t.Errorf("expected data-class=internal (L6 wins), got %q", result.Defaults.SyncedLabels["data-class"])
	}
	if result.Defaults.SyncedLabels["env"] != "prod" {
		t.Errorf("expected env=prod, got %q", result.Defaults.SyncedLabels["env"])
	}
}

func TestMergeKinesisCascade_SyncedAnnotationsMandatory(t *testing.T) {
	result := MergeKinesisCascade(
		KinesisKropathSection{},
		KinesisKropathSection{},
		KinesisConfigSection{SyncedAnnotations: map[string]string{"kropath.io/team": "streaming"}}, // level 3
		KinesisConfigSection{SyncedAnnotations: map[string]string{"kropath.io/env": "prod"}},       // level 4
		KinesisConfigSection{},
		KinesisConfigSection{},
		KinesisKropathSection{},
		KinesisKropathSection{},
	)
	if result.Mandatory.SyncedAnnotations["kropath.io/team"] != "streaming" {
		t.Errorf("expected kropath.io/team=streaming, got %q", result.Mandatory.SyncedAnnotations["kropath.io/team"])
	}
	if result.Mandatory.SyncedAnnotations["kropath.io/env"] != "prod" {
		t.Errorf("expected kropath.io/env=prod, got %q", result.Mandatory.SyncedAnnotations["kropath.io/env"])
	}
}

func TestMergeKinesisCascade_EmptyInputsProduceEmptyOutput(t *testing.T) {
	result := MergeKinesisCascade(
		KinesisKropathSection{},
		KinesisKropathSection{},
		KinesisConfigSection{},
		KinesisConfigSection{},
		KinesisConfigSection{},
		KinesisConfigSection{},
		KinesisKropathSection{},
		KinesisKropathSection{},
	)
	empty := EffectiveKinesisConfig{}
	if result.Mandatory.StreamMode != empty.Mandatory.StreamMode {
		t.Errorf("expected empty mandatory streamMode")
	}
	if result.Mandatory.ShardCount != empty.Mandatory.ShardCount {
		t.Errorf("expected empty mandatory shardCount")
	}
	if result.Defaults.NamingTemplate != empty.Defaults.NamingTemplate {
		t.Errorf("expected empty defaults namingTemplate")
	}
}

func TestMergeKinesisCascade_KropathDefaultsLosesToKinesisCfgDefaults(t *testing.T) {
	// KropathConfig defaults (L8-9) lose to KinesisConfig defaults (L6-7) for KropathConfig fields.
	result := MergeKinesisCascade(
		KinesisKropathSection{},
		KinesisKropathSection{},
		KinesisConfigSection{},
		KinesisConfigSection{},
		KinesisConfigSection{StreamMode: "on_demand"}, // level 6 — wins
		KinesisConfigSection{},
		KinesisKropathSection{StreamMode: "provisioned"}, // level 8
		KinesisKropathSection{StreamMode: "provisioned"}, // level 9
	)
	if result.Defaults.StreamMode != "on_demand" {
		t.Errorf("expected on_demand (L6 wins over L8-9), got %q", result.Defaults.StreamMode)
	}
}

func TestMergeKinesisCascade_MandatoryAndDefaultsAreIndependent(t *testing.T) {
	// A field set in mandatory does not affect defaults.
	result := MergeKinesisCascade(
		KinesisKropathSection{StreamMode: "on_demand"}, // mandatory level 1
		KinesisKropathSection{},
		KinesisConfigSection{},
		KinesisConfigSection{},
		KinesisConfigSection{StreamMode: "provisioned"}, // defaults level 6
		KinesisConfigSection{},
		KinesisKropathSection{},
		KinesisKropathSection{},
	)
	if result.Mandatory.StreamMode != "on_demand" {
		t.Errorf("expected mandatory on_demand, got %q", result.Mandatory.StreamMode)
	}
	if result.Defaults.StreamMode != "provisioned" {
		t.Errorf("expected defaults provisioned, got %q", result.Defaults.StreamMode)
	}
}

func TestMergeKinesisCascade_ShardCountZeroIsNotEnforced(t *testing.T) {
	// shardCount=0 is the sentinel "not enforced" — a non-zero value from a weaker
	// level must propagate if all stronger levels are 0.
	result := MergeKinesisCascade(
		KinesisKropathSection{ShardCount: 0}, // level 1 — not enforced
		KinesisKropathSection{ShardCount: 0}, // level 2
		KinesisConfigSection{ShardCount: 0},  // level 3
		KinesisConfigSection{ShardCount: 4},  // level 4 — enforced
		KinesisConfigSection{},
		KinesisConfigSection{},
		KinesisKropathSection{},
		KinesisKropathSection{},
	)
	if result.Mandatory.ShardCount != 4 {
		t.Errorf("expected shardCount=4 (L4 is first non-zero), got %d", result.Mandatory.ShardCount)
	}
}
