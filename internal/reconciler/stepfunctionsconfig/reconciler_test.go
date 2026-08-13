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

package stepfunctionsconfig

import (
	"testing"
)

func TestMergeTierTags(t *testing.T) {
	tests := []struct {
		name   string
		tier   map[string]string
		family map[string]string
		want   map[string]string
	}{
		{
			name:   "family-only — tier-level is absent",
			tier:   nil,
			family: map[string]string{"env": "prod"},
			want:   map[string]string{"env": "prod"},
		},
		{
			name:   "tier-only — family-level is absent",
			tier:   map[string]string{"cost-centre": "platform"},
			family: nil,
			want:   map[string]string{"cost-centre": "platform"},
		},
		{
			name:   "additive union — no overlap",
			tier:   map[string]string{"cost-centre": "platform"},
			family: map[string]string{"env": "prod"},
			want:   map[string]string{"cost-centre": "platform", "env": "prod"},
		},
		{
			name:   "key conflict — family wins over tier",
			tier:   map[string]string{"env": "org-wide", "cost-centre": "platform"},
			family: map[string]string{"env": "prod"},
			want:   map[string]string{"env": "prod", "cost-centre": "platform"},
		},
		{
			name:   "both nil — result is nil",
			tier:   nil,
			family: nil,
			want:   nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeTierTags(tc.tier, tc.family)
			if len(got) != len(tc.want) {
				t.Fatalf("mergeTierTags() len = %d, want %d: got %v", len(got), len(tc.want), got)
			}
			for k, want := range tc.want {
				if v, ok := got[k]; !ok || v != want {
					t.Errorf("mergeTierTags()[%q] = %q, want %q", k, v, want)
				}
			}
		})
	}
}
