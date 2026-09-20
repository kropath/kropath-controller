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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestDeploymentFlag(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		flag      string
		wantValue string
		wantOK    bool
	}{
		{"equals form", []string{"--enable-carm=true"}, "--enable-carm", "true", true},
		{"space-separated form", []string{"--enable-carm", "true"}, "--enable-carm", "true", true},
		{"bare boolean flag at end of args", []string{"--enable-carm"}, "--enable-carm", "true", true},
		{"bare boolean flag followed by another flag", []string{"--enable-carm", "--aws-region=us-east-1"}, "--enable-carm", "true", true},
		{"flag absent", []string{"--aws-region=us-east-1"}, "--enable-carm", "", false},
		{"empty args", nil, "--enable-carm", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := *testDeployment("ack-system", "ack-iam-controller", tt.args)
			value, ok := deploymentFlag(d, tt.flag)
			if ok != tt.wantOK || value != tt.wantValue {
				t.Errorf("deploymentFlag(%v, %q) = (%q, %v), want (%q, %v)", tt.args, tt.flag, value, ok, tt.wantValue, tt.wantOK)
			}
		})
	}
}

func TestDeploymentFlagChecksAllContainers(t *testing.T) {
	d := appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "sidecar", Args: []string{"--metrics-bind-addr=:8080"}},
						{Name: "controller", Args: []string{"--enable-carm=true"}},
					},
				},
			},
		},
	}
	value, ok := deploymentFlag(d, "--enable-carm")
	if !ok || value != "true" {
		t.Errorf("deploymentFlag across containers = (%q, %v), want (\"true\", true)", value, ok)
	}
}

func TestCheckEnableCARM(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		noDeployment bool
		wantSeverity Severity
	}{
		{"no deployments", nil, true, SeverityUnknown},
		{"flag true", []string{"--enable-carm=true"}, false, SeverityOK},
		{"flag false", []string{"--enable-carm=false"}, false, SeverityBlocking},
		{"flag space-separated true", []string{"--enable-carm", "true"}, false, SeverityOK},
		{"flag absent", []string{"--aws-region=us-east-1"}, false, SeverityUnknown},
		{"flag unrecognized value", []string{"--enable-carm=maybe"}, false, SeverityUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c client.Client
			if tt.noDeployment {
				c = testClient(t).Build()
			} else {
				c = testClient(t, testDeployment("ack-system", "ack-iam-controller", tt.args)).Build()
			}

			findings := checkEnableCARM(context.Background(), c, "ack-system")
			if len(findings) == 0 {
				t.Fatal("checkEnableCARM returned no findings")
			}
			for _, f := range findings {
				if f.Check != "enable-carm" {
					t.Errorf("finding.Check = %q, want \"enable-carm\"", f.Check)
				}
				if f.Severity != tt.wantSeverity {
					t.Errorf("finding severity = %q, want %q (message: %s)", f.Severity, tt.wantSeverity, f.Message)
				}
			}
		})
	}
}

func TestCheckWatchNamespaceScope(t *testing.T) {
	managed := []corev1.Namespace{*namespace("payments-prod", nil, nil), *namespace("data-prod", nil, nil)}

	tests := []struct {
		name         string
		args         []string
		wantSeverity Severity
	}{
		{"unset watches everything", []string{"--enable-carm=true"}, SeverityOK},
		{"covers all managed namespaces", []string{"--watch-namespace=payments-prod,data-prod"}, SeverityOK},
		{"covers all plus extra", []string{"--watch-namespace=payments-prod,data-prod,other-prod"}, SeverityOK},
		{"missing one namespace", []string{"--watch-namespace=payments-prod"}, SeverityBlocking},
		{"missing all namespaces", []string{"--watch-namespace=other-prod"}, SeverityBlocking},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := testClient(t, testDeployment("ack-system", "ack-iam-controller", tt.args)).Build()

			findings := checkWatchNamespaceScope(context.Background(), c, "ack-system", managed)
			if len(findings) != 1 {
				t.Fatalf("got %d findings, want 1: %+v", len(findings), findings)
			}
			if findings[0].Severity != tt.wantSeverity {
				t.Errorf("severity = %q, want %q (message: %s)", findings[0].Severity, tt.wantSeverity, findings[0].Message)
			}
		})
	}
}

func TestCheckWatchNamespaceScopeNoManagedNamespaces(t *testing.T) {
	c := testClient(t, testDeployment("ack-system", "ack-iam-controller", []string{"--watch-namespace=other-prod"})).Build()

	findings := checkWatchNamespaceScope(context.Background(), c, "ack-system", nil)
	if len(findings) != 1 || findings[0].Severity != SeverityOK {
		t.Errorf("expected a single OK finding with zero managed namespaces, got: %+v", findings)
	}
}

func TestCheckWatchNamespaceScopeNoDeployments(t *testing.T) {
	c := testClient(t).Build()

	findings := checkWatchNamespaceScope(context.Background(), c, "ack-system", nil)
	if len(findings) != 1 || findings[0].Severity != SeverityUnknown {
		t.Errorf("expected a single Unknown finding with zero Deployments, got: %+v", findings)
	}
}

func TestCheckNamespaceIgnoreListClean(t *testing.T) {
	managed := []corev1.Namespace{*namespace("payments-prod", nil, nil), *namespace("data-prod", nil, nil)}

	findings := checkNamespaceIgnoreList(managed)
	if len(findings) != 1 || findings[0].Severity != SeverityOK {
		t.Errorf("expected a single OK finding, got: %+v", findings)
	}
}

func TestCheckNamespaceIgnoreListNoManagedNamespaces(t *testing.T) {
	findings := checkNamespaceIgnoreList(nil)
	if len(findings) != 1 || findings[0].Severity != SeverityOK {
		t.Errorf("expected a single OK finding with zero managed namespaces, got: %+v", findings)
	}
}

func TestCheckNamespaceIgnoreListOneMatch(t *testing.T) {
	managed := []corev1.Namespace{*namespace("payments-prod", nil, nil), *namespace("kube-system", nil, nil)}

	findings := checkNamespaceIgnoreList(managed)
	if len(findings) != 1 {
		t.Fatalf("expected exactly one finding, got %d: %+v", len(findings), findings)
	}
	if findings[0].Severity != SeverityBlocking || findings[0].Namespace != "kube-system" {
		t.Errorf("finding = %+v, want Blocking on kube-system", findings[0])
	}
}

func TestCheckNamespaceIgnoreListAllThreeMatch(t *testing.T) {
	managed := []corev1.Namespace{
		*namespace("kube-system", nil, nil),
		*namespace("kube-public", nil, nil),
		*namespace("kube-node-lease", nil, nil),
	}

	findings := checkNamespaceIgnoreList(managed)
	if len(findings) != 3 {
		t.Fatalf("expected three findings, got %d: %+v", len(findings), findings)
	}
	for _, f := range findings {
		if f.Severity != SeverityBlocking {
			t.Errorf("finding = %+v, want Blocking", f)
		}
	}
}

func TestCheckNamespaceIgnoreListFindingsAreCheckNamedWatchNamespaceScope(t *testing.T) {
	// This check is the second half of precondition 1, so it must share the
	// same Check name as checkWatchNamespaceScope's findings.
	findings := checkNamespaceIgnoreList([]corev1.Namespace{*namespace("kube-system", nil, nil)})
	for _, f := range findings {
		if f.Check != "watch-namespace-scope" {
			t.Errorf("finding.Check = %q, want %q", f.Check, "watch-namespace-scope")
		}
	}
}
