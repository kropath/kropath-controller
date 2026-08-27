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

package dsqlconfig

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kropath/kropath-controller/api/v1alpha1"
	"github.com/kropath/kropath-controller/internal/cascade"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestReconcileC1GlobalMandatoryDeleteProtectionPropagates(t *testing.T) {
	rec, cfg := testReconciler(t,
		globalKropathConfig("general-policy", v1alpha1.KropathConfigTier{
			DSQL: cascade.DSQLKropathSection{DeletionProtectionEnabled: true},
		}),
		localDSQLConfig("payments-prod", "general-policy", cascade.DSQLConfigSection{}, cascade.DSQLConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getDSQLConfig(t, rec.Client, "payments-prod", "general-policy")
	if !updated.Status.EffectiveConfig.Mandatory.DeletionProtectionEnabled {
		t.Fatal("mandatory.deletionProtectionEnabled = false, want true")
	}
	if updated.Status.ObservedGeneration != cfg.Generation {
		t.Fatalf("observedGeneration = %d, want %d", updated.Status.ObservedGeneration, cfg.Generation)
	}
}

func TestReconcileC1Level1MandatoryWinsOverLevel4(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfig("general-policy", v1alpha1.KropathConfigTier{
			DSQL: cascade.DSQLKropathSection{DeletionProtectionEnabled: true},
		}),
		localDSQLConfig("payments-prod", "general-policy",
			cascade.DSQLConfigSection{DeletionProtectionEnabled: false},
			cascade.DSQLConfigSection{},
		),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getDSQLConfig(t, rec.Client, "payments-prod", "general-policy")
	if !updated.Status.EffectiveConfig.Mandatory.DeletionProtectionEnabled {
		t.Fatal("mandatory.deletionProtectionEnabled = false, want true (level-1 wins)")
	}
}

func TestReconcileC5GlobalDSQLConfigKmsKeyPropagates(t *testing.T) {
	rec, _ := testReconciler(t,
		globalDSQLConfig("general-policy",
			cascade.DSQLConfigSection{KmsEncryptionKey: "arn:aws:kms:us-east-1:123456789012:key/global"},
			cascade.DSQLConfigSection{},
		),
		localDSQLConfig("payments-prod", "general-policy", cascade.DSQLConfigSection{}, cascade.DSQLConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getDSQLConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Mandatory.KmsEncryptionKey; got != "arn:aws:kms:us-east-1:123456789012:key/global" {
		t.Fatalf("mandatory.kmsEncryptionKey = %q, want global key ARN", got)
	}
}

func TestReconcileC5DefaultsKmsKeyFromGlobalDSQLConfig(t *testing.T) {
	rec, _ := testReconciler(t,
		globalDSQLConfig("general-policy",
			cascade.DSQLConfigSection{},
			cascade.DSQLConfigSection{KmsEncryptionKey: "arn:aws:kms:us-east-1:123456789012:key/default-key"},
		),
		localDSQLConfig("payments-prod", "general-policy", cascade.DSQLConfigSection{}, cascade.DSQLConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getDSQLConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Defaults.KmsEncryptionKey; got != "arn:aws:kms:us-east-1:123456789012:key/default-key" {
		t.Fatalf("defaults.kmsEncryptionKey = %q, want default-key ARN", got)
	}
	if got := updated.Status.EffectiveConfig.Mandatory.KmsEncryptionKey; got != "" {
		t.Fatalf("mandatory.kmsEncryptionKey = %q, want empty", got)
	}
}

func TestReconcileC8TagUnionMerge(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfig("general-policy", v1alpha1.KropathConfigTier{
			Tags: map[string]string{"owner": "platform-team", "shared-key": "from-global-kropath"},
		}),
		localDSQLConfig("payments-prod", "general-policy",
			cascade.DSQLConfigSection{Tags: map[string]string{"team": "payments", "shared-key": "from-local-dsqlconfig"}},
			cascade.DSQLConfigSection{},
		),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getDSQLConfig(t, rec.Client, "payments-prod", "general-policy")
	tags := updated.Status.EffectiveConfig.Mandatory.Tags
	if tags["owner"] != "platform-team" {
		t.Fatalf("tags[owner] = %q, want platform-team", tags["owner"])
	}
	if tags["team"] != "payments" {
		t.Fatalf("tags[team] = %q, want payments", tags["team"])
	}
	// Level 1 (global KropathConfig mandatory) wins over level 4 (local DSQLConfig mandatory) on conflicts.
	if tags["shared-key"] != "from-global-kropath" {
		t.Fatalf("tags[shared-key] = %q, want from-global-kropath (level-1 wins)", tags["shared-key"])
	}
}

func TestReconcileC9ValidConditionSetOnSuccess(t *testing.T) {
	rec, _ := testReconciler(t,
		localDSQLConfig("payments-prod", "general-policy", cascade.DSQLConfigSection{}, cascade.DSQLConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getDSQLConfig(t, rec.Client, "payments-prod", "general-policy")
	cond := findCondition(updated.Status.Conditions, "Valid")
	if cond == nil {
		t.Fatal("missing Valid condition")
	}
	if cond.Status != metav1.ConditionTrue {
		t.Fatalf("Valid condition status = %v, want True", cond.Status)
	}
	if cond.Reason != "ValidationPassed" {
		t.Fatalf("Valid condition reason = %q, want ValidationPassed", cond.Reason)
	}
}

func TestReconcileCopiesAWSIdentity(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfigWithAWS("general-policy", v1alpha1.ProviderIdentity{AccountID: "123456789012", Region: "us-east-1"}),
		localDSQLConfig("payments-prod", "general-policy", cascade.DSQLConfigSection{}, cascade.DSQLConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getDSQLConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.AWS.AccountID; got != "123456789012" {
		t.Fatalf("aws.accountId = %q, want 123456789012", got)
	}
	if got := updated.Status.EffectiveConfig.AWS.Region; got != "us-east-1" {
		t.Fatalf("aws.region = %q, want us-east-1", got)
	}
}

func TestRequestsForKropathConfigChangeGlobal(t *testing.T) {
	rec, _ := testReconciler(t,
		localDSQLConfig("payments-prod", "general-policy", cascade.DSQLConfigSection{}, cascade.DSQLConfigSection{}),
		localDSQLConfig("sandbox", "general-policy", cascade.DSQLConfigSection{}, cascade.DSQLConfigSection{}),
		localDSQLConfig("payments-prod", "other-policy", cascade.DSQLConfigSection{}, cascade.DSQLConfigSection{}),
	)

	got := rec.requestsForKropathConfigChange(context.Background(), &v1alpha1.KropathConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "general-policy", Namespace: "kro-system"},
	})

	want := map[string]bool{
		"payments-prod/general-policy": false,
		"sandbox/general-policy":       false,
	}
	if len(got) != len(want) {
		t.Fatalf("requests len = %d, want %d (%#v)", len(got), len(want), got)
	}
	for _, r := range got {
		key := r.Namespace + "/" + r.Name
		if _, ok := want[key]; !ok {
			t.Fatalf("unexpected request %q", key)
		}
		want[key] = true
	}
	for key, seen := range want {
		if !seen {
			t.Fatalf("missing request %q", key)
		}
	}
}

func TestRequestsForDSQLConfigChangeGlobal(t *testing.T) {
	rec, _ := testReconciler(t,
		localDSQLConfig("payments-prod", "general-policy", cascade.DSQLConfigSection{}, cascade.DSQLConfigSection{}),
		localDSQLConfig("data-prod", "general-policy", cascade.DSQLConfigSection{}, cascade.DSQLConfigSection{}),
	)

	got := rec.requestsForDSQLConfigChange(context.Background(), &v1alpha1.DSQLConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "general-policy", Namespace: "kro-system"},
	})

	if len(got) != 2 {
		t.Fatalf("requests len = %d, want 2 (%#v)", len(got), got)
	}
}

func TestRequestsForDSQLConfigChangeNonGlobalIgnored(t *testing.T) {
	rec, _ := testReconciler(t,
		localDSQLConfig("payments-prod", "general-policy", cascade.DSQLConfigSection{}, cascade.DSQLConfigSection{}),
	)

	got := rec.requestsForDSQLConfigChange(context.Background(), &v1alpha1.DSQLConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "general-policy", Namespace: "payments-prod"},
	})

	if len(got) != 0 {
		t.Fatalf("requests len = %d, want 0 (%#v)", len(got), got)
	}
}

// ─── Test fixtures ──────────────────────────────────────────────────────────

func testReconciler(t *testing.T, objs ...runtime.Object) (*Reconciler, *v1alpha1.DSQLConfig) {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}
	builder := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&v1alpha1.DSQLConfig{})
	for _, obj := range objs {
		builder = builder.WithRuntimeObjects(obj)
	}
	cl := builder.Build()
	cfg := &v1alpha1.DSQLConfig{}
	if err := cl.Get(context.Background(), client.ObjectKey{Namespace: "payments-prod", Name: "general-policy"}, cfg); err != nil {
		t.Fatalf("seed local DSQLConfig: %v", err)
	}
	return &Reconciler{Client: cl, Log: logr.Discard(), Scheme: scheme}, cfg
}

func globalKropathConfig(name string, tier v1alpha1.KropathConfigTier) *v1alpha1.KropathConfig {
	return &v1alpha1.KropathConfig{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "KropathConfig"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "kro-system",
		},
		Spec: v1alpha1.KropathConfigSpec{
			Mandatory: tier,
		},
	}
}

func globalKropathConfigWithAWS(name string, aws v1alpha1.ProviderIdentity) *v1alpha1.KropathConfig {
	cfg := globalKropathConfig(name, v1alpha1.KropathConfigTier{})
	cfg.Spec.AWS = aws
	return cfg
}

func globalDSQLConfig(name string, mandatory, defaults cascade.DSQLConfigSection) *v1alpha1.DSQLConfig {
	return &v1alpha1.DSQLConfig{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "DSQLConfig"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "kro-system",
		},
		Spec: v1alpha1.DSQLConfigSpec{
			Mandatory: mandatory,
			Defaults:  defaults,
		},
	}
}

func localDSQLConfig(namespace, name string, mandatory, defaults cascade.DSQLConfigSection) *v1alpha1.DSQLConfig {
	cfg := globalDSQLConfig(name, mandatory, defaults)
	cfg.Namespace = namespace
	return cfg
}

func getDSQLConfig(t *testing.T, c client.Client, namespace, name string) *v1alpha1.DSQLConfig {
	t.Helper()
	cfg := &v1alpha1.DSQLConfig{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("DSQLConfig"))
	if err := c.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: name}, cfg); err != nil {
		t.Fatalf("get DSQLConfig %s/%s: %v", namespace, name, err)
	}
	return cfg
}

func findCondition(conditions []metav1.Condition, condType string) *metav1.Condition {
	for i := range conditions {
		if conditions[i].Type == condType {
			return &conditions[i]
		}
	}
	return nil
}

func req(namespace, name string) ctrl.Request {
	return ctrl.Request{NamespacedName: client.ObjectKey{Namespace: namespace, Name: name}}
}
