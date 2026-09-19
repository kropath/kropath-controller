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
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestCheckRoleAccountMapConfigMapMissing(t *testing.T) {
	c := testClient(t).Build()

	findings := checkRoleAccountMap(context.Background(), c, Options{ACKNamespace: "ack-system"}, nil)
	if len(findings) != 1 || findings[0].Severity != SeverityUnknown {
		t.Errorf("expected a single Unknown finding for a missing ConfigMap, got: %+v", findings)
	}
}

func TestCheckRoleAccountMapMatchingEntry(t *testing.T) {
	cm := testConfigMap("ack-system", "ack-role-account-map", map[string]string{
		"111122223333": "arn:aws:iam::111122223333:role/ACKRole",
	})
	c := testClient(t, cm).Build()

	findings := checkRoleAccountMap(context.Background(), c, Options{ACKNamespace: "ack-system"}, nil)
	if len(findings) != 1 || findings[0].Severity != SeverityOK {
		t.Errorf("expected a single OK finding for a matching entry, got: %+v", findings)
	}
}

func TestCheckRoleAccountMapMismatchedEntry(t *testing.T) {
	cm := testConfigMap("ack-system", "ack-role-account-map", map[string]string{
		"111122223333": "arn:aws:iam::999988887777:role/ACKRole",
	})
	c := testClient(t, cm).Build()

	findings := checkRoleAccountMap(context.Background(), c, Options{ACKNamespace: "ack-system"}, nil)
	if len(findings) != 1 || findings[0].Severity != SeverityBlocking {
		t.Errorf("expected a single Blocking finding for a mismatched entry, got: %+v", findings)
	}
}

func TestCheckRoleAccountMapMalformedKey(t *testing.T) {
	cm := testConfigMap("ack-system", "ack-role-account-map", map[string]string{
		"not-an-account-id": "arn:aws:iam::111122223333:role/ACKRole",
	})
	c := testClient(t, cm).Build()

	findings := checkRoleAccountMap(context.Background(), c, Options{ACKNamespace: "ack-system"}, nil)
	if len(findings) != 1 || findings[0].Severity != SeverityWarning {
		t.Errorf("expected a single Warning finding for a malformed key, got: %+v", findings)
	}
}

func TestCheckRoleAccountMapMalformedARN(t *testing.T) {
	cm := testConfigMap("ack-system", "ack-role-account-map", map[string]string{
		"111122223333": "not-an-arn",
	})
	c := testClient(t, cm).Build()

	findings := checkRoleAccountMap(context.Background(), c, Options{ACKNamespace: "ack-system"}, nil)
	if len(findings) != 1 || findings[0].Severity != SeverityWarning {
		t.Errorf("expected a single Warning finding for a malformed ARN, got: %+v", findings)
	}
}

func TestCheckRoleAccountMapCustomName(t *testing.T) {
	cm := testConfigMap("ack-system", "custom-role-map", map[string]string{
		"111122223333": "arn:aws:iam::111122223333:role/ACKRole",
	})
	c := testClient(t, cm).Build()

	findings := checkRoleAccountMap(context.Background(), c, Options{ACKNamespace: "ack-system", RoleAccountMapName: "custom-role-map"}, nil)
	if len(findings) != 1 || findings[0].Severity != SeverityOK {
		t.Errorf("expected a single OK finding using a custom ConfigMap name, got: %+v", findings)
	}
}

func TestCheckRoleAccountMapNamespaceMissingEntry(t *testing.T) {
	cm := testConfigMap("ack-system", "ack-role-account-map", map[string]string{
		"999988887777": "arn:aws:iam::999988887777:role/ACKRole",
	})
	managed := []corev1.Namespace{*namespace("payments-prod", map[string]string{
		"services.k8s.aws/owner-account-id": "111122223333",
	}, nil)}
	c := testClient(t, cm).Build()

	findings := checkRoleAccountMap(context.Background(), c, Options{ACKNamespace: "ack-system"}, managed)

	blocking := 0
	for _, f := range findings {
		if f.Severity == SeverityBlocking && f.Namespace == "payments-prod" {
			blocking++
		}
	}
	if blocking != 1 {
		t.Errorf("expected exactly one blocking finding for payments-prod's missing map entry, got %d in %+v", blocking, findings)
	}
}

func TestCheckRoleAccountMapNamespaceWithoutAnnotationIsIgnored(t *testing.T) {
	cm := testConfigMap("ack-system", "ack-role-account-map", nil)
	managed := []corev1.Namespace{*namespace("payments-prod", nil, nil)}
	c := testClient(t, cm).Build()

	findings := checkRoleAccountMap(context.Background(), c, Options{ACKNamespace: "ack-system"}, managed)
	for _, f := range findings {
		if f.Namespace == "payments-prod" {
			t.Errorf("namespace without owner-account-id annotation should produce no finding, got: %+v", f)
		}
	}
}
