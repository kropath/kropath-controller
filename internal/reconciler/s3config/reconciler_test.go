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

package s3config

import (
	"context"
	"testing"

	"github.com/kropath/kropath-controller/api/v1alpha1"
	"github.com/kropath/kropath-controller/internal/cascade"
	"github.com/kropath/kropath-controller/internal/reconciler/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const globalNS = util.DefaultGlobalNamespace // "kro-system"

func testScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme v1alpha1: %v", err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme corev1: %v", err)
	}
	return scheme
}

func testReconciler(t *testing.T, objs ...runtime.Object) (*Reconciler, client.Client) {
	t.Helper()
	scheme := testScheme(t)
	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(objs...).
		WithStatusSubresource(&v1alpha1.S3Config{}).
		Build()
	return &Reconciler{Client: c, Scheme: scheme}, c
}

// --- Object factory helpers ---

func namespace(name string) *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}
}

func namespaceWithAnnotation(name, globalConfigNS string) *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: map[string]string{util.GlobalConfigNamespaceAnnotation: globalConfigNS},
		},
	}
}

func globalKropathConfig(name string, s3 cascade.S3Section) *v1alpha1.KropathConfig {
	return &v1alpha1.KropathConfig{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "KropathConfig"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: globalNS,
		},
		Spec: v1alpha1.KropathConfigSpec{
			Mandatory: v1alpha1.KropathConfigTier{S3: s3},
		},
	}
}

// localKropathConfig creates a KropathConfig in a namespace.
// After Gap 2, the local KPC is always named "default"; callers must pass "default".
func localKropathConfig(ns, name string, s3 cascade.S3Section) *v1alpha1.KropathConfig {
	return &v1alpha1.KropathConfig{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "KropathConfig"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: v1alpha1.KropathConfigSpec{
			Mandatory: v1alpha1.KropathConfigTier{S3: s3},
		},
	}
}

func globalS3Config(name string, mandatory, defaults cascade.S3ConfigSection) *v1alpha1.S3Config {
	return &v1alpha1.S3Config{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "S3Config"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: globalNS,
		},
		Spec: v1alpha1.S3ConfigSpec{
			Mandatory: mandatory,
			Defaults:  defaults,
		},
	}
}

func localS3Config(ns, name string, mandatory, defaults cascade.S3ConfigSection) *v1alpha1.S3Config {
	return &v1alpha1.S3Config{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "S3Config"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: v1alpha1.S3ConfigSpec{
			Mandatory: mandatory,
			Defaults:  defaults,
		},
	}
}

func getS3Config(t *testing.T, c client.Client, ns, name string) *v1alpha1.S3Config {
	t.Helper()
	cfg := &v1alpha1.S3Config{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("S3Config"))
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: ns, Name: name}, cfg); err != nil {
		t.Fatalf("Get S3Config %s/%s: %v", ns, name, err)
	}
	return cfg
}

// --- Core reconcile tests ---

func TestReconcilerReconcile(t *testing.T) {
	rec, c := testReconciler(t,
		namespace("payments-prod"),
		globalKropathConfig("general-policy", cascade.S3Section{EncryptionAlgorithm: "aws:kms"}),
		// Gap 2: local KPC must be named "default"
		localKropathConfig("payments-prod", "default", cascade.S3Section{BlockPublicAccess: true}),
		globalS3Config("general-policy",
			cascade.S3ConfigSection{KmsKeyArn: "arn:global-s3"},
			cascade.S3ConfigSection{Versioning: "Enabled"},
		),
		localS3Config("payments-prod", "general-policy",
			cascade.S3ConfigSection{Versioning: "Suspended"},
			cascade.S3ConfigSection{LogDeliveryBucket: "logs"},
		),
	)

	_, err := rec.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "payments-prod", Name: "general-policy"},
	})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	got := getS3Config(t, c, "payments-prod", "general-policy")
	if got.Status.EffectiveConfig.Mandatory.EncryptionAlgorithm != "aws:kms" {
		t.Errorf("mandatory.encryptionAlgorithm = %q, want aws:kms", got.Status.EffectiveConfig.Mandatory.EncryptionAlgorithm)
	}
	if !got.Status.EffectiveConfig.Mandatory.BlockPublicAccess {
		t.Error("mandatory.blockPublicAccess should be true")
	}
	if got.Status.EffectiveConfig.Mandatory.KmsKeyArn != "arn:global-s3" {
		t.Errorf("mandatory.kmsKeyArn = %q, want arn:global-s3", got.Status.EffectiveConfig.Mandatory.KmsKeyArn)
	}
	if got.Status.EffectiveConfig.Defaults.Versioning != "Enabled" {
		t.Errorf("defaults.versioning = %q, want Enabled", got.Status.EffectiveConfig.Defaults.Versioning)
	}
	if got.Status.EffectiveConfig.Defaults.LogDeliveryBucket != "logs" {
		t.Errorf("defaults.logDeliveryBucket = %q, want logs", got.Status.EffectiveConfig.Defaults.LogDeliveryBucket)
	}
}

// Gap 2: reconciler resolves local KropathConfig by name "default", not by cfg.Name.
func TestLocalKropathConfigLooksUpDefault(t *testing.T) {
	rec, c := testReconciler(t,
		namespace("payments-prod"),
		// KPC named "default" (local) — the one the reconciler should pick up
		localKropathConfig("payments-prod", "default", cascade.S3Section{BlockPublicAccess: true}),
		// KPC named "general-policy" — the reconciler should NOT pick this up as local KPC
		localKropathConfig("payments-prod", "general-policy", cascade.S3Section{EnforceHttpsOnly: true}),
		localS3Config("payments-prod", "general-policy",
			cascade.S3ConfigSection{}, cascade.S3ConfigSection{},
		),
	)

	_, err := rec.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "payments-prod", Name: "general-policy"},
	})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	got := getS3Config(t, c, "payments-prod", "general-policy")
	if !got.Status.EffectiveConfig.Mandatory.BlockPublicAccess {
		t.Error("mandatory.blockPublicAccess should be true — set only on KPC/default")
	}
	// EnforceHttpsOnly is from KPC/"general-policy" which is NOT used as local KPC
	if got.Status.EffectiveConfig.Mandatory.EnforceHttpsOnly {
		t.Error("mandatory.enforceHttpsOnly should be false — KPC/'general-policy' is not used as local KPC")
	}
}

// Gap 1: when namespace carries aws.kropath.run/global-config-namespace annotation,
// the reconciler resolves the global KropathConfig from the annotated namespace.
func TestResolveGlobalNamespaceFromAnnotation(t *testing.T) {
	const customGlobalNS = "platform-config"

	rec, c := testReconciler(t,
		// Annotate the resource namespace to point to a custom global namespace
		namespaceWithAnnotation("payments-prod", customGlobalNS),
		namespace(customGlobalNS),
		// Global KPC in the custom namespace (NOT kro-system)
		&v1alpha1.KropathConfig{
			TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "KropathConfig"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "general-policy",
				Namespace: customGlobalNS,
			},
			Spec: v1alpha1.KropathConfigSpec{
				Mandatory: v1alpha1.KropathConfigTier{S3: cascade.S3Section{EncryptionAlgorithm: "aws:kms"}},
			},
		},
		localKropathConfig("payments-prod", "default", cascade.S3Section{}),
		localS3Config("payments-prod", "general-policy",
			cascade.S3ConfigSection{}, cascade.S3ConfigSection{},
		),
	)

	_, err := rec.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "payments-prod", Name: "general-policy"},
	})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	got := getS3Config(t, c, "payments-prod", "general-policy")
	if got.Status.EffectiveConfig.Mandatory.EncryptionAlgorithm != "aws:kms" {
		t.Errorf("mandatory.encryptionAlgorithm = %q, want aws:kms — global KPC in annotated namespace should apply",
			got.Status.EffectiveConfig.Mandatory.EncryptionAlgorithm)
	}
}

// Gap 1: when no annotation is present on the namespace, the reconciler falls back to kro-system.
func TestGlobalNamespaceDefaultsToKroSystem(t *testing.T) {
	rec, c := testReconciler(t,
		namespace("payments-prod"),
		// KPC in kro-system (default global namespace)
		globalKropathConfig("general-policy", cascade.S3Section{EncryptionAlgorithm: "aws:kms"}),
		localKropathConfig("payments-prod", "default", cascade.S3Section{}),
		localS3Config("payments-prod", "general-policy",
			cascade.S3ConfigSection{}, cascade.S3ConfigSection{},
		),
	)

	_, err := rec.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "payments-prod", Name: "general-policy"},
	})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	got := getS3Config(t, c, "payments-prod", "general-policy")
	if got.Status.EffectiveConfig.Mandatory.EncryptionAlgorithm != "aws:kms" {
		t.Errorf("mandatory.encryptionAlgorithm = %q, want aws:kms — global KPC in kro-system should apply",
			got.Status.EffectiveConfig.Mandatory.EncryptionAlgorithm)
	}
}

// Gap 3: NamingTemplate from S3Config (not KropathConfig) is carried through to effective config.
func TestNamingTemplateFromS3ConfigApplied(t *testing.T) {
	rec, c := testReconciler(t,
		namespace("payments-prod"),
		localKropathConfig("payments-prod", "default", cascade.S3Section{}),
		localS3Config("payments-prod", "general-policy",
			cascade.S3ConfigSection{NamingTemplate: "{account_id}-{namespace}-{name}"},
			cascade.S3ConfigSection{},
		),
	)

	_, err := rec.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "payments-prod", Name: "general-policy"},
	})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	got := getS3Config(t, c, "payments-prod", "general-policy")
	want := "{account_id}-{namespace}-{name}"
	if got.Status.EffectiveConfig.Mandatory.NamingTemplate != want {
		t.Errorf("mandatory.namingTemplate = %q, want %q", got.Status.EffectiveConfig.Mandatory.NamingTemplate, want)
	}
}

// --- Fan-out tests ---

func TestRequestsForKropathConfigChangeIncludesGlobalAndLocalMatches(t *testing.T) {
	rec, _ := testReconciler(t,
		namespace("payments-prod"),
		namespace("other-ns"),
		globalS3Config("general-policy", cascade.S3ConfigSection{}, cascade.S3ConfigSection{}),
		localS3Config("payments-prod", "general-policy", cascade.S3ConfigSection{}, cascade.S3ConfigSection{}),
		localS3Config("other-ns", "general-policy", cascade.S3ConfigSection{}, cascade.S3ConfigSection{}),
		localS3Config("payments-prod", "other-policy", cascade.S3ConfigSection{}, cascade.S3ConfigSection{}),
	)

	// A global KPC (named "general-policy", in kro-system) should enqueue all S3Configs
	// whose resolved global namespace is kro-system AND whose name matches.
	requests := rec.requestsForKropathConfigChange(context.Background(), &v1alpha1.KropathConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "general-policy", Namespace: globalNS},
	})

	// kro-system/general-policy, payments-prod/general-policy, other-ns/general-policy
	if len(requests) != 3 {
		t.Fatalf("requests len = %d, want 3; got %#v", len(requests), requests)
	}
	got := map[string]bool{}
	for _, req := range requests {
		got[req.Namespace+"/"+req.Name] = true
	}
	for _, want := range []string{
		globalNS + "/general-policy",
		"payments-prod/general-policy",
		"other-ns/general-policy",
	} {
		if !got[want] {
			t.Errorf("missing request %q in %#v", want, requests)
		}
	}
}

func TestRequestsForKropathConfigChangeLocalDefaultEnqueuesNamespace(t *testing.T) {
	rec, _ := testReconciler(t,
		namespace("payments-prod"),
		localS3Config("payments-prod", "general-policy", cascade.S3ConfigSection{}, cascade.S3ConfigSection{}),
		localS3Config("payments-prod", "other-policy", cascade.S3ConfigSection{}, cascade.S3ConfigSection{}),
		// Different namespace — should NOT be enqueued by this local KPC change
		localS3Config("other-ns", "general-policy", cascade.S3ConfigSection{}, cascade.S3ConfigSection{}),
	)

	// Local KPC named "default" in "payments-prod" — should enqueue all S3Configs in that namespace
	requests := rec.requestsForKropathConfigChange(context.Background(), &v1alpha1.KropathConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "default", Namespace: "payments-prod"},
	})

	if len(requests) != 2 {
		t.Fatalf("requests len = %d, want 2; got %#v", len(requests), requests)
	}
}

func TestRequestsForS3ConfigChangeQueuesNamespaceMatchesForGlobalSource(t *testing.T) {
	rec, _ := testReconciler(t,
		namespace("payments-prod"),
		namespace("data-prod"),
		globalS3Config("general-policy", cascade.S3ConfigSection{}, cascade.S3ConfigSection{}),
		localS3Config("payments-prod", "general-policy", cascade.S3ConfigSection{}, cascade.S3ConfigSection{}),
		localS3Config("data-prod", "general-policy", cascade.S3ConfigSection{}, cascade.S3ConfigSection{}),
	)

	requests := rec.requestsForS3ConfigChange(context.Background(), &v1alpha1.S3Config{
		ObjectMeta: metav1.ObjectMeta{Name: "general-policy", Namespace: globalNS},
	})

	if len(requests) != 2 {
		t.Fatalf("requests len = %d, want 2; got %#v", len(requests), requests)
	}
}
