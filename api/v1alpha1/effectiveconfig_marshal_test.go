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

package v1alpha1

import (
	"encoding/json"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// jsonHasKey reports whether marshaled reports a top-level "effectiveConfig"
// key -- the check that reflect.DeepEqual-against-zero-value cannot make,
// because it only ever inspects the in-memory Go struct (see KRO-1199).
func jsonHasKey(t *testing.T, marshaled []byte, key string) bool {
	t.Helper()
	var m map[string]json.RawMessage
	if err := json.Unmarshal(marshaled, &m); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	_, ok := m[key]
	return ok
}

func TestS3ConfigStatusMarshalJSON_OmitsZeroEffectiveConfig(t *testing.T) {
	status := S3ConfigStatus{
		ObservedGeneration: 2,
		SyncedTimestamp:    "2026-09-22T00:00:00Z",
		Conditions: []metav1.Condition{
			{Type: "Reconciled", Status: metav1.ConditionFalse, Reason: "MissingAccountAnnotation"},
		},
	}

	b, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if jsonHasKey(t, b, "effectiveConfig") {
		t.Fatalf("marshaled status carries an \"effectiveConfig\" key for a zero-value EffectiveConfig; want it absent: %s", b)
	}
	if !jsonHasKey(t, b, "observedGeneration") || !jsonHasKey(t, b, "syncedTimestamp") || !jsonHasKey(t, b, "conditions") {
		t.Fatalf("marshaled status is missing a non-zero field: %s", b)
	}
}

func TestS3ConfigStatusMarshalJSON_KeepsNonZeroEffectiveConfig(t *testing.T) {
	status := S3ConfigStatus{
		EffectiveConfig: EffectiveS3Config{
			AWS: ProviderIdentity{AccountID: "111122223333", Region: "ap-southeast-2", Partition: "aws"},
		},
	}

	b, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if !jsonHasKey(t, b, "effectiveConfig") {
		t.Fatalf("marshaled status is missing \"effectiveConfig\" for a non-zero value: %s", b)
	}

	var roundTripped S3ConfigStatus
	if err := json.Unmarshal(b, &roundTripped); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if roundTripped.EffectiveConfig.AWS.AccountID != "111122223333" {
		t.Fatalf("round-tripped AccountID = %q, want %q", roundTripped.EffectiveConfig.AWS.AccountID, "111122223333")
	}
}

func TestS3ConfigStatusMarshalJSON_EmptyStatusMarshalsToEmptyObject(t *testing.T) {
	b, err := json.Marshal(S3ConfigStatus{})
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if string(b) != "{}" {
		t.Fatalf("marshaled zero-value status = %s, want {}", b)
	}
}

// TestAllConfigStatusTypesOmitZeroEffectiveConfig spot-checks a sample of the
// ~57 <ResourceFamily>ConfigStatus types beyond S3Config to confirm the
// generated MarshalJSON methods all delegate to marshalConfigStatus
// correctly, not just the type exercised by the other tests in this file.
func TestAllConfigStatusTypesOmitZeroEffectiveConfig(t *testing.T) {
	cases := []struct {
		name string
		got  interface{ MarshalJSON() ([]byte, error) }
	}{
		{"EC2Config", EC2ConfigStatus{ObservedGeneration: 1}},
		{"IAMConfig", IAMConfigStatus{ObservedGeneration: 1}},
		{"SQSConfig", SQSConfigStatus{ObservedGeneration: 1}},
		{"BedrockConfig", BedrockConfigStatus{ObservedGeneration: 1}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := tc.got.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON: %v", err)
			}
			if jsonHasKey(t, b, "effectiveConfig") {
				t.Fatalf("%s: marshaled status carries \"effectiveConfig\" for a zero value; want it absent: %s", tc.name, b)
			}
		})
	}
}
