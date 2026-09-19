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
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestCheckTeamIDNoneCarryAnnotation(t *testing.T) {
	managed := []corev1.Namespace{
		*namespace("payments-prod", nil, nil),
		*namespace("data-prod", map[string]string{"some-other-annotation": "x"}, nil),
	}

	findings := checkTeamID(managed)
	if len(findings) != 1 || findings[0].Severity != SeverityOK {
		t.Errorf("expected a single OK finding, got: %+v", findings)
	}
}

func TestCheckTeamIDOneCarriesAnnotation(t *testing.T) {
	managed := []corev1.Namespace{
		*namespace("payments-prod", nil, nil),
		*namespace("data-prod", map[string]string{TeamIDAnnotation: "platform-team"}, nil),
	}

	findings := checkTeamID(managed)
	if len(findings) != 1 {
		t.Fatalf("expected exactly one finding, got %d: %+v", len(findings), findings)
	}
	if findings[0].Severity != SeverityBlocking || findings[0].Namespace != "data-prod" {
		t.Errorf("finding = %+v, want Blocking on data-prod", findings[0])
	}
}

func TestCheckTeamIDEmptyAnnotationValueIsIgnored(t *testing.T) {
	managed := []corev1.Namespace{*namespace("payments-prod", map[string]string{TeamIDAnnotation: ""}, nil)}

	findings := checkTeamID(managed)
	if len(findings) != 1 || findings[0].Severity != SeverityOK {
		t.Errorf("expected an empty annotation value to be treated as absent, got: %+v", findings)
	}
}

func TestCheckTeamIDMultipleViolations(t *testing.T) {
	managed := []corev1.Namespace{
		*namespace("payments-prod", map[string]string{TeamIDAnnotation: "team-a"}, nil),
		*namespace("data-prod", map[string]string{TeamIDAnnotation: "team-b"}, nil),
	}

	findings := checkTeamID(managed)
	if len(findings) != 2 {
		t.Fatalf("expected two findings, got %d: %+v", len(findings), findings)
	}
	for _, f := range findings {
		if f.Severity != SeverityBlocking {
			t.Errorf("finding = %+v, want Blocking", f)
		}
	}
}
