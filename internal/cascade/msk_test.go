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

func TestMergeMSKCascade_KropathMandatoryWinsOnKafkaVersion(t *testing.T) {
	// Level 1 (globalKropathMandatory) wins over all lower levels.
	result := MergeMSKCascade(
		MSKKropathSection{KafkaVersion: "3.7.0"}, // level 1 — wins
		MSKKropathSection{KafkaVersion: "3.6.0"}, // level 2
		MSKConfigSection{KafkaVersion: "3.5.0"},   // level 3
		MSKConfigSection{KafkaVersion: "3.4.0"},   // level 4
		MSKConfigSection{},                         // level 6
		MSKConfigSection{},                         // level 7
		MSKKropathSection{},                        // level 8
		MSKKropathSection{},                        // level 9
	)
	if result.Mandatory.KafkaVersion != "3.7.0" {
		t.Errorf("expected mandatory.kafkaVersion=3.7.0, got %q", result.Mandatory.KafkaVersion)
	}
}

func TestMergeMSKCascade_LocalKropathMandatoryWinsOverMSKCfg(t *testing.T) {
	// Level 2 (localKropathMandatory) wins over MSKConfig mandatory levels (3-4).
	result := MergeMSKCascade(
		MSKKropathSection{},                         // level 1 — empty
		MSKKropathSection{KafkaVersion: "3.7.0"},   // level 2 — wins
		MSKConfigSection{KafkaVersion: "3.5.0"},    // level 3
		MSKConfigSection{KafkaVersion: "3.4.0"},    // level 4
		MSKConfigSection{},                          // level 6
		MSKConfigSection{},                          // level 7
		MSKKropathSection{},                         // level 8
		MSKKropathSection{},                         // level 9
	)
	if result.Mandatory.KafkaVersion != "3.7.0" {
		t.Errorf("expected mandatory.kafkaVersion=3.7.0, got %q", result.Mandatory.KafkaVersion)
	}
}

func TestMergeMSKCascade_GlobalMSKCfgMandatoryWinsOverLocal(t *testing.T) {
	// Level 3 (globalMSKCfgMandatory) wins over level 4 (localMSKCfgMandatory).
	result := MergeMSKCascade(
		MSKKropathSection{},                         // level 1 — empty
		MSKKropathSection{},                         // level 2 — empty
		MSKConfigSection{KafkaVersion: "3.7.0"},    // level 3 — wins
		MSKConfigSection{KafkaVersion: "3.5.0"},    // level 4
		MSKConfigSection{},                          // level 6
		MSKConfigSection{},                          // level 7
		MSKKropathSection{},                         // level 8
		MSKKropathSection{},                         // level 9
	)
	if result.Mandatory.KafkaVersion != "3.7.0" {
		t.Errorf("expected mandatory.kafkaVersion=3.7.0, got %q", result.Mandatory.KafkaVersion)
	}
}

func TestMergeMSKCascade_LocalMSKCfgDefaultsWinsOverGlobal(t *testing.T) {
	// Level 6 (localMSKCfgDefaults) wins over level 7 (globalMSKCfgDefaults).
	result := MergeMSKCascade(
		MSKKropathSection{}, // level 1
		MSKKropathSection{}, // level 2
		MSKConfigSection{},  // level 3
		MSKConfigSection{},  // level 4
		MSKConfigSection{KafkaVersion: "3.7.0"}, // level 6 — wins
		MSKConfigSection{KafkaVersion: "3.6.0"}, // level 7
		MSKKropathSection{},                     // level 8
		MSKKropathSection{},                     // level 9
	)
	if result.Defaults.KafkaVersion != "3.7.0" {
		t.Errorf("expected defaults.kafkaVersion=3.7.0, got %q", result.Defaults.KafkaVersion)
	}
}

func TestMergeMSKCascade_GlobalKropathDefaultsIsWeakest(t *testing.T) {
	// Level 9 (globalKropathDefaults) only applies when all other defaults levels are empty.
	result := MergeMSKCascade(
		MSKKropathSection{}, // level 1
		MSKKropathSection{}, // level 2
		MSKConfigSection{},  // level 3
		MSKConfigSection{},  // level 4
		MSKConfigSection{},  // level 6
		MSKConfigSection{},  // level 7
		MSKKropathSection{}, // level 8
		MSKKropathSection{KafkaVersion: "3.6.0"}, // level 9 — weakest, only source
	)
	if result.Defaults.KafkaVersion != "3.6.0" {
		t.Errorf("expected defaults.kafkaVersion=3.6.0, got %q", result.Defaults.KafkaVersion)
	}
}

func TestMergeMSKCascade_EncryptionInTransitClientBroker_Level1(t *testing.T) {
	result := MergeMSKCascade(
		MSKKropathSection{EncryptionInTransitClientBroker: "TLS"}, // level 1 — wins
		MSKKropathSection{EncryptionInTransitClientBroker: "TLS_PLAINTEXT"},
		MSKConfigSection{EncryptionInTransitClientBroker: "PLAINTEXT"},
		MSKConfigSection{},
		MSKConfigSection{},
		MSKConfigSection{},
		MSKKropathSection{},
		MSKKropathSection{},
	)
	if result.Mandatory.EncryptionInTransitClientBroker != "TLS" {
		t.Errorf("expected TLS, got %q", result.Mandatory.EncryptionInTransitClientBroker)
	}
}

func TestMergeMSKCascade_EncryptionInTransitInCluster_Level1(t *testing.T) {
	result := MergeMSKCascade(
		MSKKropathSection{EncryptionInTransitInCluster: "true"}, // level 1 — wins
		MSKKropathSection{},
		MSKConfigSection{EncryptionInTransitInCluster: "false"},
		MSKConfigSection{},
		MSKConfigSection{},
		MSKConfigSection{},
		MSKKropathSection{},
		MSKKropathSection{},
	)
	if result.Mandatory.EncryptionInTransitInCluster != "true" {
		t.Errorf("expected true, got %q", result.Mandatory.EncryptionInTransitInCluster)
	}
}

func TestMergeMSKCascade_EncryptionAtRestKmsKeyId_Level1(t *testing.T) {
	result := MergeMSKCascade(
		MSKKropathSection{EncryptionAtRestKmsKeyId: "arn:aws:kms:ap-southeast-2:123:key/org-key"}, // level 1
		MSKKropathSection{},
		MSKConfigSection{EncryptionAtRestKmsKeyId: "arn:aws:kms:ap-southeast-2:123:key/profile-key"},
		MSKConfigSection{},
		MSKConfigSection{},
		MSKConfigSection{},
		MSKKropathSection{},
		MSKKropathSection{},
	)
	if result.Mandatory.EncryptionAtRestKmsKeyId != "arn:aws:kms:ap-southeast-2:123:key/org-key" {
		t.Errorf("expected org-key, got %q", result.Mandatory.EncryptionAtRestKmsKeyId)
	}
}

func TestMergeMSKCascade_EnhancedMonitoring_Level1(t *testing.T) {
	result := MergeMSKCascade(
		MSKKropathSection{EnhancedMonitoring: "PER_BROKER"}, // level 1
		MSKKropathSection{},
		MSKConfigSection{EnhancedMonitoring: "DEFAULT"},
		MSKConfigSection{},
		MSKConfigSection{},
		MSKConfigSection{},
		MSKKropathSection{},
		MSKKropathSection{},
	)
	if result.Mandatory.EnhancedMonitoring != "PER_BROKER" {
		t.Errorf("expected PER_BROKER, got %q", result.Mandatory.EnhancedMonitoring)
	}
}

func TestMergeMSKCascade_InstanceTypeNoKropathLevels(t *testing.T) {
	// instanceType has no KropathConfig.msk equivalent — only MSKConfig levels 3-4 apply.
	result := MergeMSKCascade(
		MSKKropathSection{}, // level 1 — no instanceType field
		MSKKropathSection{}, // level 2
		MSKConfigSection{InstanceType: "kafka.m5.xlarge"}, // level 3 — wins
		MSKConfigSection{InstanceType: "kafka.m5.large"},  // level 4
		MSKConfigSection{},
		MSKConfigSection{},
		MSKKropathSection{},
		MSKKropathSection{},
	)
	if result.Mandatory.InstanceType != "kafka.m5.xlarge" {
		t.Errorf("expected kafka.m5.xlarge, got %q", result.Mandatory.InstanceType)
	}
}

func TestMergeMSKCascade_InstanceTypeDefaultsNoKropathLevels(t *testing.T) {
	// instanceType in defaults tier: only MSKConfig levels 6-7 apply (no KropathConfig equivalent).
	result := MergeMSKCascade(
		MSKKropathSection{},
		MSKKropathSection{},
		MSKConfigSection{},
		MSKConfigSection{},
		MSKConfigSection{InstanceType: "kafka.m5.large"}, // level 6 — wins
		MSKConfigSection{InstanceType: "kafka.m5.2xlarge"}, // level 7
		MSKKropathSection{},
		MSKKropathSection{},
	)
	if result.Defaults.InstanceType != "kafka.m5.large" {
		t.Errorf("expected kafka.m5.large, got %q", result.Defaults.InstanceType)
	}
}

func TestMergeMSKCascade_NamingTemplateNoKropathLevels(t *testing.T) {
	// namingTemplate has no KropathConfig.msk equivalent.
	result := MergeMSKCascade(
		MSKKropathSection{},
		MSKKropathSection{},
		MSKConfigSection{NamingTemplate: "corp-{name}"}, // level 3 — wins
		MSKConfigSection{NamingTemplate: "{namespace}-{name}"},
		MSKConfigSection{},
		MSKConfigSection{},
		MSKKropathSection{},
		MSKKropathSection{},
	)
	if result.Mandatory.NamingTemplate != "corp-{name}" {
		t.Errorf("expected corp-{name}, got %q", result.Mandatory.NamingTemplate)
	}
}

func TestMergeMSKCascade_NamingTemplateDefaultsFallthrough(t *testing.T) {
	// Level 6 local config defaults.namingTemplate wins over level 7 global.
	result := MergeMSKCascade(
		MSKKropathSection{},
		MSKKropathSection{},
		MSKConfigSection{},
		MSKConfigSection{},
		MSKConfigSection{NamingTemplate: "{namespace}-{name}"}, // level 6
		MSKConfigSection{NamingTemplate: "global-{name}"},       // level 7
		MSKKropathSection{},
		MSKKropathSection{},
	)
	if result.Defaults.NamingTemplate != "{namespace}-{name}" {
		t.Errorf("expected {namespace}-{name}, got %q", result.Defaults.NamingTemplate)
	}
}

func TestMergeMSKCascade_TagsMergeAcrossAllMandatoryLevels(t *testing.T) {
	// Mandatory tags: union of KropathConfig.mandatory.tags (generic) + MSKConfig mandatory tiers.
	// Higher levels win on key conflict.
	result := MergeMSKCascade(
		MSKKropathSection{Tags: map[string]string{"cost-centre": "infra", "shared": "yes"}}, // level 1 — wins on conflict
		MSKKropathSection{Tags: map[string]string{"cost-centre": "dev"}},                    // level 2 — overridden
		MSKConfigSection{Tags: map[string]string{"cluster-type": "streaming"}},              // level 3
		MSKConfigSection{Tags: map[string]string{"team": "platform"}},                       // level 4
		MSKConfigSection{},
		MSKConfigSection{},
		MSKKropathSection{},
		MSKKropathSection{},
	)
	want := map[string]string{
		"cost-centre":  "infra",
		"shared":       "yes",
		"cluster-type": "streaming",
		"team":         "platform",
	}
	for k, v := range want {
		if result.Mandatory.Tags[k] != v {
			t.Errorf("mandatory.tags[%q]: expected %q, got %q", k, v, result.Mandatory.Tags[k])
		}
	}
}

func TestMergeMSKCascade_TagsMergeAcrossAllDefaultsLevels(t *testing.T) {
	// Defaults tags: union of MSKConfig defaults + KropathConfig.defaults.tags (generic).
	// Lower level number (stronger) wins on key conflict.
	result := MergeMSKCascade(
		MSKKropathSection{},
		MSKKropathSection{},
		MSKConfigSection{},
		MSKConfigSection{},
		MSKConfigSection{Tags: map[string]string{"team": "streaming", "env": "prod"}}, // level 6 — wins
		MSKConfigSection{Tags: map[string]string{"team": "platform"}},                  // level 7
		MSKKropathSection{Tags: map[string]string{"org": "acme"}},                      // level 8
		MSKKropathSection{Tags: map[string]string{"org": "corp"}},                      // level 9 — weakest
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

func TestMergeMSKCascade_SyncedLabelsMandatory_MSKCfgLevelsOnly(t *testing.T) {
	// SyncedLabels for mandatory: union of globalMSKCfgMandatory (L3) + localMSKCfgMandatory (L4).
	// L3 wins on key conflict. KropathConfig levels do not contribute to syncedLabels.
	result := MergeMSKCascade(
		MSKKropathSection{}, // level 1 — no syncedLabels
		MSKKropathSection{}, // level 2
		MSKConfigSection{SyncedLabels: map[string]string{"data-class": "confidential", "region": "ap"}}, // level 3 wins
		MSKConfigSection{SyncedLabels: map[string]string{"data-class": "internal", "team": "platform"}}, // level 4
		MSKConfigSection{},
		MSKConfigSection{},
		MSKKropathSection{},
		MSKKropathSection{},
	)
	if result.Mandatory.SyncedLabels["data-class"] != "confidential" {
		t.Errorf("expected data-class=confidential (L3 wins), got %q", result.Mandatory.SyncedLabels["data-class"])
	}
	if result.Mandatory.SyncedLabels["team"] != "platform" {
		t.Errorf("expected team=platform, got %q", result.Mandatory.SyncedLabels["team"])
	}
	if result.Mandatory.SyncedLabels["region"] != "ap" {
		t.Errorf("expected region=ap, got %q", result.Mandatory.SyncedLabels["region"])
	}
}

func TestMergeMSKCascade_SyncedLabelsDefaults_MSKCfgLevelsOnly(t *testing.T) {
	// SyncedLabels for defaults: union of globalMSKCfgDefaults (L7) + localMSKCfgDefaults (L6).
	// L6 wins on key conflict.
	result := MergeMSKCascade(
		MSKKropathSection{},
		MSKKropathSection{},
		MSKConfigSection{},
		MSKConfigSection{},
		MSKConfigSection{SyncedLabels: map[string]string{"data-class": "internal", "env": "prod"}}, // level 6 — wins
		MSKConfigSection{SyncedLabels: map[string]string{"data-class": "public"}},                  // level 7
		MSKKropathSection{},
		MSKKropathSection{},
	)
	if result.Defaults.SyncedLabels["data-class"] != "internal" {
		t.Errorf("expected data-class=internal (L6 wins), got %q", result.Defaults.SyncedLabels["data-class"])
	}
	if result.Defaults.SyncedLabels["env"] != "prod" {
		t.Errorf("expected env=prod, got %q", result.Defaults.SyncedLabels["env"])
	}
}

func TestMergeMSKCascade_SyncedAnnotationsMandatory(t *testing.T) {
	result := MergeMSKCascade(
		MSKKropathSection{},
		MSKKropathSection{},
		MSKConfigSection{SyncedAnnotations: map[string]string{"kropath.io/team": "streaming"}}, // level 3
		MSKConfigSection{SyncedAnnotations: map[string]string{"kropath.io/env": "prod"}},       // level 4
		MSKConfigSection{},
		MSKConfigSection{},
		MSKKropathSection{},
		MSKKropathSection{},
	)
	if result.Mandatory.SyncedAnnotations["kropath.io/team"] != "streaming" {
		t.Errorf("expected kropath.io/team=streaming, got %q", result.Mandatory.SyncedAnnotations["kropath.io/team"])
	}
	if result.Mandatory.SyncedAnnotations["kropath.io/env"] != "prod" {
		t.Errorf("expected kropath.io/env=prod, got %q", result.Mandatory.SyncedAnnotations["kropath.io/env"])
	}
}

func TestMergeMSKCascade_EmptyInputsProduceEmptyOutput(t *testing.T) {
	result := MergeMSKCascade(
		MSKKropathSection{},
		MSKKropathSection{},
		MSKConfigSection{},
		MSKConfigSection{},
		MSKConfigSection{},
		MSKConfigSection{},
		MSKKropathSection{},
		MSKKropathSection{},
	)
	empty := EffectiveMSKConfig{}
	if result.Mandatory.KafkaVersion != empty.Mandatory.KafkaVersion {
		t.Errorf("expected empty mandatory, got non-empty")
	}
	if result.Defaults.InstanceType != empty.Defaults.InstanceType {
		t.Errorf("expected empty defaults, got non-empty")
	}
}

func TestMergeMSKCascade_KropathDefaultsLosesToMSKCfgDefaults(t *testing.T) {
	// KropathConfig defaults (L8-9) lose to MSKConfig defaults (L6-7) for KropathConfig fields.
	result := MergeMSKCascade(
		MSKKropathSection{},
		MSKKropathSection{},
		MSKConfigSection{},
		MSKConfigSection{},
		MSKConfigSection{KafkaVersion: "3.7.0"}, // level 6 — wins
		MSKConfigSection{},
		MSKKropathSection{KafkaVersion: "3.6.0"}, // level 8
		MSKKropathSection{KafkaVersion: "3.5.0"}, // level 9
	)
	if result.Defaults.KafkaVersion != "3.7.0" {
		t.Errorf("expected 3.7.0 (L6 wins over L8-9), got %q", result.Defaults.KafkaVersion)
	}
}

func TestMergeMSKCascade_MandatoryAndDefaultsAreIndependent(t *testing.T) {
	// A field set in mandatory does not affect defaults.
	result := MergeMSKCascade(
		MSKKropathSection{EncryptionInTransitClientBroker: "TLS"}, // mandatory level 1
		MSKKropathSection{},
		MSKConfigSection{},
		MSKConfigSection{},
		MSKConfigSection{EncryptionInTransitClientBroker: "TLS_PLAINTEXT"}, // defaults level 6
		MSKConfigSection{},
		MSKKropathSection{},
		MSKKropathSection{},
	)
	if result.Mandatory.EncryptionInTransitClientBroker != "TLS" {
		t.Errorf("expected mandatory TLS, got %q", result.Mandatory.EncryptionInTransitClientBroker)
	}
	if result.Defaults.EncryptionInTransitClientBroker != "TLS_PLAINTEXT" {
		t.Errorf("expected defaults TLS_PLAINTEXT, got %q", result.Defaults.EncryptionInTransitClientBroker)
	}
}

func TestMergeMSKCascade_EncryptionInTransitInClusterThreeState(t *testing.T) {
	// "false" is a valid value (not the zero/empty sentinel) — it should propagate.
	result := MergeMSKCascade(
		MSKKropathSection{},
		MSKKropathSection{},
		MSKConfigSection{},
		MSKConfigSection{},
		MSKConfigSection{EncryptionInTransitInCluster: "false"}, // level 6 — valid non-empty value
		MSKConfigSection{},
		MSKKropathSection{},
		MSKKropathSection{},
	)
	if result.Defaults.EncryptionInTransitInCluster != "false" {
		t.Errorf("expected false (valid three-state value), got %q", result.Defaults.EncryptionInTransitInCluster)
	}
}
