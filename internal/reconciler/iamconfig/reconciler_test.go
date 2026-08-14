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

package iamconfig

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

func TestReconcileAC1GlobalMandatoryBoundaryWins(t *testing.T) {
	rec, cfg := testReconciler(t,
		globalKropathConfig("general-policy", cascade.IAMSection{PermissionsBoundaryArn: "arn:aws:iam::123:policy/GlobalBlanket"}),
		localIAMConfig("payments-prod", "general-policy"),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getIAMConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Mandatory.PermissionsBoundaryArn; got != "arn:aws:iam::123:policy/GlobalBlanket" {
		t.Fatalf("mandatory.permissionsBoundaryArn = %q, want global boundary", got)
	}
	if updated.Status.ObservedGeneration != cfg.Generation {
		t.Fatalf("observedGeneration = %d, want %d", updated.Status.ObservedGeneration, cfg.Generation)
	}
}

func TestReconcileAC2Level1WinsOverLevel3(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfig("general-policy", cascade.IAMSection{PermissionsBoundaryArn: "arn:aws:iam::123:policy/GlobalBlanket"}),
		globalIAMConfig("general-policy", cascade.IAMSection{PermissionsBoundaryArn: "arn:aws:iam::123:policy/IAMCfgBoundary"}),
		localIAMConfig("payments-prod", "general-policy"),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getIAMConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Mandatory.PermissionsBoundaryArn; got != "arn:aws:iam::123:policy/GlobalBlanket" {
		t.Fatalf("mandatory.permissionsBoundaryArn = %q, want level-1 value", got)
	}
}

func TestReconcileAC3BlockIamUserAccessKeys(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfig("general-policy", cascade.IAMSection{BlockIamUserAccessKeys: true}),
		localIAMConfig("payments-prod", "general-policy"),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getIAMConfig(t, rec.Client, "payments-prod", "general-policy")
	if !updated.Status.EffectiveConfig.Mandatory.BlockIamUserAccessKeys {
		t.Fatal("mandatory.blockIamUserAccessKeys = false, want true")
	}
}

func TestReconcileAC4Level1MaxSessionWins(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfig("general-policy", cascade.IAMSection{MaxSessionDurationSeconds: 3600}),
		globalIAMConfig("general-policy", cascade.IAMSection{MaxSessionDurationSeconds: 7200}),
		localIAMConfig("payments-prod", "general-policy"),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getIAMConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Mandatory.MaxSessionDurationSeconds; got != 3600 {
		t.Fatalf("mandatory.maxSessionDurationSeconds = %d, want 3600", got)
	}
}

func TestReconcileAC5DefaultsOnly(t *testing.T) {
	rec, _ := testReconciler(t,
		globalIAMConfigDefaults("general-policy", cascade.IAMSection{MaxSessionDurationSeconds: 3600}),
		localIAMConfig("payments-prod", "general-policy"),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getIAMConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Defaults.MaxSessionDurationSeconds; got != 3600 {
		t.Fatalf("defaults.maxSessionDurationSeconds = %d, want 3600", got)
	}
	if got := updated.Status.EffectiveConfig.Mandatory.MaxSessionDurationSeconds; got != 0 {
		t.Fatalf("mandatory.maxSessionDurationSeconds = %d, want 0", got)
	}
}

func TestReconcileAC6GlobalMandatoryWinsOverLocal(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfig("general-policy", cascade.IAMSection{PermissionsBoundaryArn: "arn:aws:iam::123:policy/GlobalBlanket"}),
		localKropathConfig("payments-prod", "general-policy", cascade.IAMSection{PermissionsBoundaryArn: "arn:aws:iam::123:policy/NsBlanket"}),
		localIAMConfig("payments-prod", "general-policy"),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getIAMConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Mandatory.PermissionsBoundaryArn; got != "arn:aws:iam::123:policy/GlobalBlanket" {
		t.Fatalf("mandatory.permissionsBoundaryArn = %q, want global value", got)
	}
}

func TestReconcileCopiesAWSIdentity(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfigWithAWS("general-policy", AWSIdentity("123456789012", "us-east-1")),
		localIAMConfig("payments-prod", "general-policy"),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getIAMConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.AWS.AccountID; got != "123456789012" {
		t.Fatalf("aws.accountId = %q, want 123456789012", got)
	}
	if got := updated.Status.EffectiveConfig.AWS.Region; got != "us-east-1" {
		t.Fatalf("aws.region = %q, want us-east-1", got)
	}
}

func TestRequestsForIAMConfigChangeGlobal(t *testing.T) {
	rec, _ := testReconciler(t,
		globalIAMConfigDefaults("general-policy", cascade.IAMSection{MaxSessionDurationSeconds: 3600}),
		localIAMConfig("payments-prod", "general-policy"),
		localIAMConfig("sandbox", "general-policy"),
		localIAMConfig("payments-prod", "other-policy"),
	)

	got := rec.requestsForIAMConfigChange(context.Background(), &v1alpha1.IAMConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "general-policy",
			Namespace: "kro-system",
		},
	})

	if len(got) != 2 {
		t.Fatalf("requests len = %d, want 2", len(got))
	}
	want := map[string]bool{
		"payments-prod/general-policy": false,
		"sandbox/general-policy":       false,
	}
	for _, req := range got {
		key := req.Namespace + "/" + req.Name
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

func TestRequestsForIAMConfigChangeNonGlobalIgnored(t *testing.T) {
	rec, _ := testReconciler(t, localIAMConfig("payments-prod", "general-policy"))

	got := rec.requestsForIAMConfigChange(context.Background(), &v1alpha1.IAMConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "general-policy",
			Namespace: "payments-prod",
		},
	})

	if len(got) != 0 {
		t.Fatalf("requests len = %d, want 0", len(got))
	}
}

func testReconciler(t *testing.T, objs ...runtime.Object) (*Reconciler, *v1alpha1.IAMConfig) {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}
	builder := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&v1alpha1.IAMConfig{})
	for _, obj := range objs {
		builder = builder.WithRuntimeObjects(obj)
	}
	cl := builder.Build()
	cfg := localIAMConfig("payments-prod", "general-policy")
	if err := cl.Get(context.Background(), client.ObjectKeyFromObject(cfg), cfg); err != nil {
		t.Fatalf("seed local IAMConfig: %v", err)
	}
	return &Reconciler{Client: cl, Log: logr.Discard(), Scheme: scheme}, cfg
}

func globalKropathConfig(name string, iam cascade.IAMSection) *v1alpha1.KropathConfig {
	return &v1alpha1.KropathConfig{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "KropathConfig"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "kro-system",
		},
		Spec: v1alpha1.KropathConfigSpec{
			Mandatory: v1alpha1.KropathConfigTier{IAM: iam},
		},
	}
}

func localKropathConfig(namespace, name string, iam cascade.IAMSection) *v1alpha1.KropathConfig {
	cfg := globalKropathConfig(name, cascade.IAMSection{})
	cfg.Namespace = namespace
	cfg.Spec.Mandatory.IAM = iam
	return cfg
}

func globalKropathConfigWithAWS(name string, aws v1alpha1.ProviderIdentity) *v1alpha1.KropathConfig {
	cfg := globalKropathConfig(name, cascade.IAMSection{})
	cfg.Spec.AWS = aws
	return cfg
}

func globalIAMConfig(name string, mandatory cascade.IAMSection) *v1alpha1.IAMConfig {
	return &v1alpha1.IAMConfig{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "IAMConfig"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "kro-system",
		},
		Spec: v1alpha1.IAMConfigSpec{
			Mandatory: mandatory,
		},
	}
}

func globalIAMConfigDefaults(name string, defaults cascade.IAMSection) *v1alpha1.IAMConfig {
	cfg := globalIAMConfig(name, cascade.IAMSection{})
	cfg.Spec.Defaults = defaults
	return cfg
}

func localIAMConfig(namespace, name string) *v1alpha1.IAMConfig {
	return &v1alpha1.IAMConfig{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "IAMConfig"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.IAMConfigSpec{},
	}
}

func AWSIdentity(accountID, region string) v1alpha1.ProviderIdentity {
	return v1alpha1.ProviderIdentity{AccountID: accountID, Region: region}
}

func getIAMConfig(t *testing.T, c client.Client, namespace, name string) *v1alpha1.IAMConfig {
	t.Helper()
	cfg := &v1alpha1.IAMConfig{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("IAMConfig"))
	if err := c.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: name}, cfg); err != nil {
		t.Fatalf("get IAMConfig: %v", err)
	}
	return cfg
}

func req(namespace, name string) ctrl.Request {
	return ctrl.Request{NamespacedName: client.ObjectKey{Namespace: namespace, Name: name}}
}
