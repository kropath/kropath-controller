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

func TestMergeQuickSightCascade_ImportMode_KropathConfigLevel1Wins(t *testing.T) {
	// AC-1: KropathConfig global mandatory importMode (level 1) wins over all other sources.
	got := MergeQuickSightCascade(
		QuickSightKropathSection{ImportMode: "SPICE"},    // level 1
		QuickSightKropathSection{ImportMode: "DIRECT_QUERY"}, // level 2
		QuickSightConfigSection{ImportMode: "DIRECT_QUERY"},   // level 3
		QuickSightConfigSection{ImportMode: "DIRECT_QUERY"},   // level 4
		QuickSightConfigSection{},                              // level 6
		QuickSightConfigSection{},                              // level 7
		QuickSightKropathSection{},                             // level 8
		QuickSightKropathSection{},                             // level 9
	)
	if got.Mandatory.ImportMode != "SPICE" {
		t.Errorf("expected SPICE, got %q", got.Mandatory.ImportMode)
	}
}

func TestMergeQuickSightCascade_ImportMode_Level3WhenKropathEmpty(t *testing.T) {
	// AC-2: QuickSightConfig global mandatory (level 3) wins when levels 1-2 are empty.
	got := MergeQuickSightCascade(
		QuickSightKropathSection{},                           // level 1 empty
		QuickSightKropathSection{},                           // level 2 empty
		QuickSightConfigSection{ImportMode: "DIRECT_QUERY"}, // level 3
		QuickSightConfigSection{ImportMode: "SPICE"},         // level 4
		QuickSightConfigSection{},
		QuickSightConfigSection{},
		QuickSightKropathSection{},
		QuickSightKropathSection{},
	)
	if got.Mandatory.ImportMode != "DIRECT_QUERY" {
		t.Errorf("expected DIRECT_QUERY, got %q", got.Mandatory.ImportMode)
	}
}

func TestMergeQuickSightCascade_ImportMode_DefaultsLevel6(t *testing.T) {
	// AC-3: all mandatory tiers empty; QuickSightConfig local defaults (level 6) sets importMode.
	got := MergeQuickSightCascade(
		QuickSightKropathSection{}, // level 1
		QuickSightKropathSection{}, // level 2
		QuickSightConfigSection{},  // level 3
		QuickSightConfigSection{},  // level 4
		QuickSightConfigSection{ImportMode: "SPICE"}, // level 6
		QuickSightConfigSection{ImportMode: "DIRECT_QUERY"}, // level 7
		QuickSightKropathSection{ImportMode: "DIRECT_QUERY"}, // level 8
		QuickSightKropathSection{ImportMode: "DIRECT_QUERY"}, // level 9
	)
	if got.Defaults.ImportMode != "SPICE" {
		t.Errorf("expected SPICE, got %q", got.Defaults.ImportMode)
	}
}

func TestMergeQuickSightCascade_NamingTemplate_MandatoryLevel3(t *testing.T) {
	// AC-4: NamingTemplate from QuickSightConfig global mandatory (level 3).
	got := MergeQuickSightCascade(
		QuickSightKropathSection{}, // level 1 — no namingTemplate
		QuickSightKropathSection{},
		QuickSightConfigSection{NamingTemplate: "corp-{namespace}-{name}"},
		QuickSightConfigSection{NamingTemplate: "{namespace}-{name}"},
		QuickSightConfigSection{},
		QuickSightConfigSection{},
		QuickSightKropathSection{},
		QuickSightKropathSection{},
	)
	if got.Mandatory.NamingTemplate != "corp-{namespace}-{name}" {
		t.Errorf("expected corp-{namespace}-{name}, got %q", got.Mandatory.NamingTemplate)
	}
}

func TestMergeQuickSightCascade_NamingTemplate_DefaultsLevel7(t *testing.T) {
	// AC-5: NamingTemplate from QuickSightConfig global defaults (level 7) when level 6 empty.
	got := MergeQuickSightCascade(
		QuickSightKropathSection{},
		QuickSightKropathSection{},
		QuickSightConfigSection{},
		QuickSightConfigSection{},
		QuickSightConfigSection{}, // level 6 empty
		QuickSightConfigSection{NamingTemplate: "{namespace}-{name}"}, // level 7
		QuickSightKropathSection{},
		QuickSightKropathSection{},
	)
	if got.Defaults.NamingTemplate != "{namespace}-{name}" {
		t.Errorf("expected {namespace}-{name}, got %q", got.Defaults.NamingTemplate)
	}
}

func TestMergeQuickSightCascade_Tags_MandatoryUnionMerge(t *testing.T) {
	// AC-6: KropathConfig + QuickSightConfig mandatory tags union-merge; L1 wins on conflict.
	got := MergeQuickSightCascade(
		QuickSightKropathSection{Tags: map[string]string{"cost-centre": "platform", "env": "prod"}}, // level 1
		QuickSightKropathSection{},
		QuickSightConfigSection{Tags: map[string]string{"bi-team": "analytics", "env": "staging"}}, // level 3
		QuickSightConfigSection{},
		QuickSightConfigSection{},
		QuickSightConfigSection{},
		QuickSightKropathSection{},
		QuickSightKropathSection{},
	)
	if got.Mandatory.Tags["cost-centre"] != "platform" {
		t.Errorf("expected cost-centre=platform, got %q", got.Mandatory.Tags["cost-centre"])
	}
	if got.Mandatory.Tags["bi-team"] != "analytics" {
		t.Errorf("expected bi-team=analytics, got %q", got.Mandatory.Tags["bi-team"])
	}
	// level 1 wins on key conflict
	if got.Mandatory.Tags["env"] != "prod" {
		t.Errorf("expected env=prod (L1 wins), got %q", got.Mandatory.Tags["env"])
	}
}

func TestMergeQuickSightCascade_SyncedLabels_AllLevelsMerge(t *testing.T) {
	// AC-7: SyncedLabels from all four mandatory levels union-merged.
	got := MergeQuickSightCascade(
		QuickSightKropathSection{SyncedLabels: map[string]string{"data-class": "confidential"}}, // level 1
		QuickSightKropathSection{},
		QuickSightConfigSection{SyncedLabels: map[string]string{"team": "data"}}, // level 3
		QuickSightConfigSection{},
		QuickSightConfigSection{},
		QuickSightConfigSection{},
		QuickSightKropathSection{},
		QuickSightKropathSection{},
	)
	if got.Mandatory.SyncedLabels["data-class"] != "confidential" {
		t.Errorf("expected data-class=confidential, got %q", got.Mandatory.SyncedLabels["data-class"])
	}
	if got.Mandatory.SyncedLabels["team"] != "data" {
		t.Errorf("expected team=data, got %q", got.Mandatory.SyncedLabels["team"])
	}
}

func TestMergeQuickSightCascade_AllEmpty(t *testing.T) {
	got := MergeQuickSightCascade(
		QuickSightKropathSection{},
		QuickSightKropathSection{},
		QuickSightConfigSection{},
		QuickSightConfigSection{},
		QuickSightConfigSection{},
		QuickSightConfigSection{},
		QuickSightKropathSection{},
		QuickSightKropathSection{},
	)
	if got.Mandatory.ImportMode != "" {
		t.Errorf("expected empty ImportMode, got %q", got.Mandatory.ImportMode)
	}
	if got.Defaults.ImportMode != "" {
		t.Errorf("expected empty ImportMode, got %q", got.Defaults.ImportMode)
	}
}

func TestMergeQuickSightCascade_DefaultsImportMode_KropathLevel9(t *testing.T) {
	// Level 9 (globalKropathDefaults) sets defaults importMode when all higher levels empty.
	got := MergeQuickSightCascade(
		QuickSightKropathSection{},
		QuickSightKropathSection{},
		QuickSightConfigSection{},
		QuickSightConfigSection{},
		QuickSightConfigSection{},
		QuickSightConfigSection{},
		QuickSightKropathSection{},
		QuickSightKropathSection{ImportMode: "SPICE"}, // level 9
	)
	if got.Defaults.ImportMode != "SPICE" {
		t.Errorf("expected SPICE, got %q", got.Defaults.ImportMode)
	}
}

func TestMergeQuickSightCascade_DefaultsTags_UnionMerge(t *testing.T) {
	// Defaults tags: L9 added first, L6 wins on key conflict.
	got := MergeQuickSightCascade(
		QuickSightKropathSection{},
		QuickSightKropathSection{},
		QuickSightConfigSection{},
		QuickSightConfigSection{},
		QuickSightConfigSection{Tags: map[string]string{"env": "dev", "team": "bi"}}, // level 6 (wins)
		QuickSightConfigSection{},
		QuickSightKropathSection{},
		QuickSightKropathSection{Tags: map[string]string{"env": "prod", "org": "kropath"}}, // level 9
	)
	if got.Defaults.Tags["env"] != "dev" {
		t.Errorf("expected env=dev (L6 wins), got %q", got.Defaults.Tags["env"])
	}
	if got.Defaults.Tags["team"] != "bi" {
		t.Errorf("expected team=bi, got %q", got.Defaults.Tags["team"])
	}
	if got.Defaults.Tags["org"] != "kropath" {
		t.Errorf("expected org=kropath, got %q", got.Defaults.Tags["org"])
	}
}
