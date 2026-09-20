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

package carmcheck

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestReportWriteTextEmpty(t *testing.T) {
	var buf bytes.Buffer
	r := &Report{}
	if err := r.WriteText(&buf); err != nil {
		t.Fatalf("WriteText() error = %v", err)
	}
	if got := buf.String(); got != "no findings\n" {
		t.Errorf("WriteText() = %q, want %q", got, "no findings\n")
	}
}

func TestReportWriteTextIncludesSeverityCheckAndMessage(t *testing.T) {
	var buf bytes.Buffer
	r := &Report{Findings: []Finding{
		{Check: "team-id", Severity: SeverityBlocking, Namespace: "payments-prod", Message: "carries team-id"},
	}}
	if err := r.WriteText(&buf); err != nil {
		t.Fatalf("WriteText() error = %v", err)
	}
	got := buf.String()
	for _, want := range []string{"BLOCKING", "team-id[payments-prod]", "carries team-id"} {
		if !strings.Contains(got, want) {
			t.Errorf("WriteText() = %q, want it to contain %q", got, want)
		}
	}
}

func TestReportWriteTextOmitsBracketsWithoutNamespace(t *testing.T) {
	var buf bytes.Buffer
	r := &Report{Findings: []Finding{
		{Check: "enable-carm", Severity: SeverityOK, Message: "flag set"},
	}}
	if err := r.WriteText(&buf); err != nil {
		t.Fatalf("WriteText() error = %v", err)
	}
	if got := buf.String(); strings.Contains(got, "enable-carm[") {
		t.Errorf("WriteText() = %q, want no bracketed namespace for an install-wide finding", got)
	}
}

func TestReportWriteJSONRoundTrips(t *testing.T) {
	var buf bytes.Buffer
	r := &Report{Findings: []Finding{
		{Check: "role-account-map", Severity: SeverityBlocking, Namespace: "payments-prod", Message: "mismatch"},
	}}
	if err := r.WriteJSON(&buf); err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}

	var got Report
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if len(got.Findings) != 1 || got.Findings[0] != r.Findings[0] {
		t.Errorf("round-tripped report = %+v, want %+v", got, r)
	}
}
