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

func TestIAMRoleSelectorMatchesNamespaceByName(t *testing.T) {
	sel := testIAMRoleSelector("prod-selector", map[string]interface{}{
		"arn": "arn:aws:iam::111122223333:role/ACK-CrossAccount-Target",
		"namespaceSelector": map[string]interface{}{
			"names": []interface{}{"payments-prod", "data-prod"},
		},
	})
	ns := *namespace("payments-prod", nil, nil)

	if !iamRoleSelectorMatchesNamespace(*sel, ns) {
		t.Error("expected a name match")
	}
}

func TestIAMRoleSelectorMatchesNamespaceByLabel(t *testing.T) {
	sel := testIAMRoleSelector("prod-selector", map[string]interface{}{
		"arn": "arn:aws:iam::111122223333:role/ACK-CrossAccount-Target",
		"namespaceSelector": map[string]interface{}{
			"names": []interface{}{},
			"labelSelector": map[string]interface{}{
				"matchLabels": map[string]interface{}{"tier": "prod"},
			},
		},
	})
	ns := *namespace("payments-prod", nil, map[string]string{"tier": "prod", "team": "payments"})

	if !iamRoleSelectorMatchesNamespace(*sel, ns) {
		t.Error("expected a label subset match")
	}
}

func TestIAMRoleSelectorLabelMatchRequiresAllKeys(t *testing.T) {
	sel := testIAMRoleSelector("prod-selector", map[string]interface{}{
		"arn": "arn:aws:iam::111122223333:role/ACK-CrossAccount-Target",
		"namespaceSelector": map[string]interface{}{
			"names": []interface{}{},
			"labelSelector": map[string]interface{}{
				"matchLabels": map[string]interface{}{"tier": "prod", "team": "payments"},
			},
		},
	})
	ns := *namespace("payments-prod", nil, map[string]string{"tier": "prod"}) // missing "team"

	if iamRoleSelectorMatchesNamespace(*sel, ns) {
		t.Error("expected no match when the namespace lacks one required label")
	}
}

func TestIAMRoleSelectorNoMatch(t *testing.T) {
	sel := testIAMRoleSelector("staging-selector", map[string]interface{}{
		"arn": "arn:aws:iam::111122223333:role/ACK-CrossAccount-Target",
		"namespaceSelector": map[string]interface{}{
			"names": []interface{}{"staging"},
		},
	})
	ns := *namespace("payments-prod", nil, nil)

	if iamRoleSelectorMatchesNamespace(*sel, ns) {
		t.Error("expected no match")
	}
}

func TestCheckIAMRoleSelectorCRDNotInstalled(t *testing.T) {
	// A scheme with no IAMRoleSelector registration reproduces the
	// "CRD absent" case: the client cannot resolve the kind at all.
	sch := testScheme(t)
	c := fakeClientWithoutIAMRoleSelector(t, sch)

	findings := checkIAMRoleSelector(context.Background(), c, nil)
	if len(findings) != 1 {
		t.Fatalf("expected exactly one finding, got %d: %+v", len(findings), findings)
	}
	if findings[0].Severity != SeverityOK && findings[0].Severity != SeverityUnknown {
		t.Errorf("expected OK or Unknown when the CRD cannot be resolved, got: %+v", findings[0])
	}
}

func TestCheckIAMRoleSelectorNoneDefined(t *testing.T) {
	c := testClient(t).Build()

	findings := checkIAMRoleSelector(context.Background(), c, []corev1.Namespace{*namespace("payments-prod", nil, nil)})
	if len(findings) != 1 || findings[0].Severity != SeverityOK {
		t.Errorf("expected a single OK finding with no selectors defined, got: %+v", findings)
	}
}

func TestCheckIAMRoleSelectorMatchingNamespace(t *testing.T) {
	sel := testIAMRoleSelector("prod-selector", map[string]interface{}{
		"arn": "arn:aws:iam::111122223333:role/ACK-CrossAccount-Target",
		"namespaceSelector": map[string]interface{}{
			"names": []interface{}{"payments-prod"},
		},
	})
	c := testClient(t, sel).Build()
	managed := []corev1.Namespace{*namespace("payments-prod", nil, nil)}

	findings := checkIAMRoleSelector(context.Background(), c, managed)
	if len(findings) != 1 || findings[0].Severity != SeverityBlocking || findings[0].Namespace != "payments-prod" {
		t.Errorf("expected a single Blocking finding on payments-prod, got: %+v", findings)
	}
}

func TestCheckIAMRoleSelectorDefinedButNoMatch(t *testing.T) {
	sel := testIAMRoleSelector("staging-selector", map[string]interface{}{
		"arn": "arn:aws:iam::111122223333:role/ACK-CrossAccount-Target",
		"namespaceSelector": map[string]interface{}{
			"names": []interface{}{"staging"},
		},
	})
	c := testClient(t, sel).Build()
	managed := []corev1.Namespace{*namespace("payments-prod", nil, nil)}

	findings := checkIAMRoleSelector(context.Background(), c, managed)
	if len(findings) != 1 || findings[0].Severity != SeverityOK {
		t.Errorf("expected a single OK finding when no selector matches, got: %+v", findings)
	}
}
