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

	"github.com/go-logr/logr"
	"github.com/kropath/kropath-controller/api/v1alpha1"
	"github.com/kropath/kropath-controller/internal/cascade"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestReconcileAC1GlobalMandatoryEncryptionWins(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfig("general-policy", cascade.S3Section{EncryptionAlgorithm: "aws:kms"}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getS3Config(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Mandatory.EncryptionAlgorithm; got != "aws:kms" {
		t.Fatalf("mandatory.encryptionAlgorithm = %q, want aws:kms", got)
	}
}

func TestReconcileAC2Level1WinsOverLevel3(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfig("general-policy", cascade.S3Section{EncryptionAlgorithm: "aws:kms"}),
		globalS3Config("general-policy", cascade.S3Section{EncryptionAlgorithm: "AES256"}, cascade.S3Section{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getS3Config(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Mandatory.EncryptionAlgorithm; got != "aws:kms" {
		t.Fatalf("mandatory.encryptionAlgorithm = %q, want level-1 value", got)
	}
}

func TestReconcileAC3BlockPublicAccessPropagates(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfig("general-policy", cascade.S3Section{BlockPublicAccess: true}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getS3Config(t, rec.Client, "payments-prod", "general-policy")
	if !updated.Status.EffectiveConfig.Mandatory.BlockPublicAccess {
		t.Fatal("mandatory.blockPublicAccess = false, want true")
	}
}

func TestReconcileAC4EnforceHttpsOnlyWins(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfig("general-policy", cascade.S3Section{EnforceHttpsOnly: true}),
		globalS3Config("general-policy", cascade.S3Section{EnforceHttpsOnly: false}, cascade.S3Section{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getS3Config(t, rec.Client, "payments-prod", "general-policy")
	if !updated.Status.EffectiveConfig.Mandatory.EnforceHttpsOnly {
		t.Fatal("mandatory.enforceHttpsOnly = false, want true")
	}
}

func TestReconcileAC5DefaultsOnly(t *testing.T) {
	rec, _ := testReconciler(t,
		globalS3Config("general-policy", cascade.S3Section{}, cascade.S3Section{EncryptionAlgorithm: "aws:kms"}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getS3Config(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Defaults.EncryptionAlgorithm; got != "aws:kms" {
		t.Fatalf("defaults.encryptionAlgorithm = %q, want aws:kms", got)
	}
	if got := updated.Status.EffectiveConfig.Mandatory.EncryptionAlgorithm; got != "" {
		t.Fatalf("mandatory.encryptionAlgorithm = %q, want empty", got)
	}
}

func TestReconcileAC6GlobalMandatoryKmsKeyWins(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfig("general-policy", cascade.S3Section{KmsKeyArn: "arn:aws:kms:us-east-1:123:key/global"}),
		localKropathConfig("payments-prod", "general-policy", cascade.S3Section{KmsKeyArn: "arn:aws:kms:us-east-1:123:key/local"}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getS3Config(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Mandatory.KmsKeyArn; got != "arn:aws:kms:us-east-1:123:key/global" {
		t.Fatalf("mandatory.kmsKeyArn = %q, want global value", got)
	}
}

func TestReconcileAC8LoggingPair(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfig("general-policy", cascade.S3Section{
			LoggingEnabled:    true,
			LogDeliveryBucket: "org-access-logs",
		}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getS3Config(t, rec.Client, "payments-prod", "general-policy")
	if !updated.Status.EffectiveConfig.Mandatory.LoggingEnabled {
		t.Fatal("mandatory.loggingEnabled = false, want true")
	}
	if got := updated.Status.EffectiveConfig.Mandatory.LogDeliveryBucket; got != "org-access-logs" {
		t.Fatalf("mandatory.logDeliveryBucket = %q, want org-access-logs", got)
	}
}

func TestReconcileAC9ObjectLockZeroValuePassThrough(t *testing.T) {
	rec, _ := testReconciler(t,
		globalS3Config("general-policy", cascade.S3Section{}, cascade.S3Section{ObjectLockMode: ""}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getS3Config(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Defaults.ObjectLockMode; got != "" {
		t.Fatalf("defaults.objectLockMode = %q, want empty", got)
	}
}

func TestReconcileMergesNamingTemplateAndMetadataMaps(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfig("general-policy", cascade.S3Section{
			NamingTemplate: "global-{namespace}",
			Tags:           map[string]string{"global": "one"},
		}),
		localKropathConfig("payments-prod", "general-policy", cascade.S3Section{
			NamingTemplate:    "local-{namespace}",
			Tags:              map[string]string{"local": "two", "shared": "local"},
			SyncedLabels:      map[string]string{"kpc": "local"},
			SyncedAnnotations: map[string]string{"anno": "local"},
		}),
		globalS3Config("general-policy", cascade.S3Section{
			Tags:              map[string]string{"shared": "s3"},
			SyncedLabels:      map[string]string{"label": "s3"},
			SyncedAnnotations: map[string]string{"anno": "s3"},
		}, cascade.S3Section{
			Tags:              map[string]string{"defaults": "global"},
			SyncedLabels:      map[string]string{"label": "global"},
			SyncedAnnotations: map[string]string{"anno": "global"},
		}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getS3Config(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Mandatory.NamingTemplate; got != "global-{namespace}" {
		t.Fatalf("mandatory.namingTemplate = %q, want global-{namespace}", got)
	}
	if got := updated.Status.EffectiveConfig.Mandatory.Tags["shared"]; got != "local" {
		t.Fatalf("mandatory.tags[shared] = %q, want local", got)
	}
	if got := updated.Status.EffectiveConfig.Mandatory.Tags["global"]; got != "one" {
		t.Fatalf("mandatory.tags[global] = %q, want one", got)
	}
	if got := updated.Status.EffectiveConfig.Mandatory.SyncedLabels["kpc"]; got != "local" {
		t.Fatalf("mandatory.syncedLabels[kpc] = %q, want local", got)
	}
	if got := updated.Status.EffectiveConfig.Mandatory.SyncedLabels["label"]; got != "s3" {
		t.Fatalf("mandatory.syncedLabels[label] = %q, want s3", got)
	}
	if got := updated.Status.EffectiveConfig.Defaults.SyncedAnnotations["anno"]; got != "global" {
		t.Fatalf("defaults.syncedAnnotations[anno] = %q, want global", got)
	}
}

func TestRequestsForKropathConfigChangeGlobal(t *testing.T) {
	rec, _ := testReconciler(t,
		localS3Config("sandbox", "general-policy"),
		localS3Config("payments-prod", "other-policy"),
	)

	got := rec.requestsForKropathConfigChange(context.Background(), &v1alpha1.AWSKropathConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "general-policy",
			Namespace: kroSystemNamespace,
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

func TestRequestsForKropathConfigChangeLocalNamespace(t *testing.T) {
	rec, _ := testReconciler(t)

	got := rec.requestsForKropathConfigChange(context.Background(), &v1alpha1.AWSKropathConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "general-policy",
			Namespace: "payments-prod",
		},
	})

	if len(got) != 1 {
		t.Fatalf("requests len = %d, want 1", len(got))
	}
}

func TestRequestsForS3ConfigChangeGlobal(t *testing.T) {
	rec, _ := testReconciler(t,
		localS3Config("sandbox", "general-policy"),
		localS3Config("payments-prod", "other-policy"),
	)

	got := rec.requestsForS3ConfigChange(context.Background(), &v1alpha1.AWSS3Config{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "general-policy",
			Namespace: kroSystemNamespace,
		},
	})

	if len(got) != 2 {
		t.Fatalf("requests len = %d, want 2", len(got))
	}
}

func testReconciler(t *testing.T, objs ...runtime.Object) (*Reconciler, *v1alpha1.AWSS3Config) {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}
	builder := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&v1alpha1.AWSKropathConfig{}, &v1alpha1.AWSS3Config{})
	builder = builder.WithRuntimeObjects(localS3Config("payments-prod", "general-policy"))
	for _, obj := range objs {
		builder = builder.WithRuntimeObjects(obj)
	}
	cl := builder.Build()
	cfg := &v1alpha1.AWSS3Config{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("AWSS3Config"))
	if err := cl.Get(context.Background(), client.ObjectKey{Namespace: "payments-prod", Name: "general-policy"}, cfg); err != nil {
		t.Fatalf("seed local S3Config: %v", err)
	}
	return &Reconciler{Client: cl, Log: logr.Discard(), Scheme: scheme}, cfg
}

func globalKropathConfig(name string, s3 cascade.S3Section) *v1alpha1.AWSKropathConfig {
	return &v1alpha1.AWSKropathConfig{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "AWSKropathConfig"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: kroSystemNamespace,
		},
		Spec: v1alpha1.AWSKropathConfigSpec{
			Mandatory: v1alpha1.AWSKropathConfigTier{S3: s3},
		},
	}
}

func localKropathConfig(namespace, name string, s3 cascade.S3Section) *v1alpha1.AWSKropathConfig {
	cfg := globalKropathConfig(name, cascade.S3Section{})
	cfg.Namespace = namespace
	cfg.Spec.Mandatory.S3 = s3
	return cfg
}

func localS3Config(namespace, name string) *v1alpha1.AWSS3Config {
	return &v1alpha1.AWSS3Config{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "AWSS3Config"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.AWSS3ConfigSpec{},
	}
}

func globalS3Config(name string, mandatory, defaults cascade.S3Section) *v1alpha1.AWSS3Config {
	cfg := localS3Config(kroSystemNamespace, name)
	cfg.Spec.Mandatory = mandatory
	cfg.Spec.Defaults = defaults
	return cfg
}

func getS3Config(t *testing.T, c client.Client, namespace, name string) *v1alpha1.AWSS3Config {
	t.Helper()
	cfg := &v1alpha1.AWSS3Config{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("AWSS3Config"))
	if err := c.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: name}, cfg); err != nil {
		t.Fatalf("get S3Config: %v", err)
	}
	return cfg
}

func req(namespace, name string) ctrl.Request {
	return ctrl.Request{NamespacedName: client.ObjectKey{Namespace: namespace, Name: name}}
}
