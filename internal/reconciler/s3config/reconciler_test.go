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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func testReconciler(t *testing.T, objs ...runtime.Object) (*Reconciler, client.Client) {
	t.Helper()

	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme: %v", err)
	}

	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(objs...).
		WithStatusSubresource(&v1alpha1.AWSS3Config{}).
		Build()

	return &Reconciler{Client: c, Scheme: scheme}, c
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
	return &v1alpha1.AWSKropathConfig{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "AWSKropathConfig"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.AWSKropathConfigSpec{
			Mandatory: v1alpha1.AWSKropathConfigTier{S3: s3},
		},
	}
}

func globalS3Config(name string, mandatory, defaults cascade.S3Section) *v1alpha1.AWSS3Config {
	return &v1alpha1.AWSS3Config{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "AWSS3Config"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: kroSystemNamespace,
		},
		Spec: v1alpha1.AWSS3ConfigSpec{
			Mandatory: mandatory,
			Defaults:  defaults,
		},
	}
}

func localS3Config(namespace, name string, mandatory, defaults cascade.S3Section) *v1alpha1.AWSS3Config {
	return &v1alpha1.AWSS3Config{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "AWSS3Config"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.AWSS3ConfigSpec{
			Mandatory: mandatory,
			Defaults:  defaults,
		},
	}
}

func getS3Config(t *testing.T, c client.Client, namespace, name string) *v1alpha1.AWSS3Config {
	t.Helper()

	cfg := &v1alpha1.AWSS3Config{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("AWSS3Config"))
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, cfg); err != nil {
		t.Fatalf("Get AWSS3Config %s/%s: %v", namespace, name, err)
	}
	return cfg
}

func TestReconcilerReconcile(t *testing.T) {
	rec, c := testReconciler(t,
		globalKropathConfig("general-policy", cascade.S3Section{EncryptionAlgorithm: "aws:kms"}),
		localKropathConfig("payments-prod", "general-policy", cascade.S3Section{BlockPublicAccess: true}),
		globalS3Config("general-policy", cascade.S3Section{KmsKeyArn: "arn:global-s3"}, cascade.S3Section{Versioning: "Enabled"}),
		localS3Config("payments-prod", "general-policy", cascade.S3Section{Versioning: "Suspended"}, cascade.S3Section{LogDeliveryBucket: "logs"}),
		localS3Config("data-prod", "general-policy", cascade.S3Section{ObjectLockMode: "GOVERNANCE"}, cascade.S3Section{}),
	)

	result, err := rec.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "payments-prod", Name: "general-policy"},
	})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	if result.RequeueAfter != 0 {
		t.Fatalf("unexpected requeue: %v", result.RequeueAfter)
	}

	got := getS3Config(t, c, "payments-prod", "general-policy")
	if got.Status.EffectiveConfig.Mandatory.EncryptionAlgorithm != "aws:kms" {
		t.Fatalf("mandatory.encryptionAlgorithm = %q, want aws:kms", got.Status.EffectiveConfig.Mandatory.EncryptionAlgorithm)
	}
	if !got.Status.EffectiveConfig.Mandatory.BlockPublicAccess {
		t.Fatal("mandatory.blockPublicAccess should be true")
	}
	if got.Status.EffectiveConfig.Mandatory.KmsKeyArn != "arn:global-s3" {
		t.Fatalf("mandatory.kmsKeyArn = %q, want arn:global-s3", got.Status.EffectiveConfig.Mandatory.KmsKeyArn)
	}
	if got.Status.EffectiveConfig.Defaults.Versioning != "Enabled" {
		t.Fatalf("defaults.versioning = %q, want Enabled", got.Status.EffectiveConfig.Defaults.Versioning)
	}
	if got.Status.EffectiveConfig.Defaults.LogDeliveryBucket != "logs" {
		t.Fatalf("defaults.logDeliveryBucket = %q, want logs", got.Status.EffectiveConfig.Defaults.LogDeliveryBucket)
	}
}

func TestRequestsForKropathConfigChangeIncludesGlobalAndLocalMatches(t *testing.T) {
	rec, _ := testReconciler(t,
		globalS3Config("general-policy", cascade.S3Section{}, cascade.S3Section{}),
		localS3Config("payments-prod", "general-policy", cascade.S3Section{}, cascade.S3Section{}),
		localS3Config("other-ns", "general-policy", cascade.S3Section{}, cascade.S3Section{}),
		localS3Config("payments-prod", "other-policy", cascade.S3Section{}, cascade.S3Section{}),
	)

	requests := rec.requestsForKropathConfigChange(context.Background(), &v1alpha1.AWSKropathConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "general-policy", Namespace: kroSystemNamespace},
	})

	if len(requests) != 3 {
		t.Fatalf("requests len = %d, want 3", len(requests))
	}
	got := map[string]bool{}
	for _, req := range requests {
		got[req.Namespace+"/"+req.Name] = true
	}
	for _, want := range []string{
		kroSystemNamespace + "/general-policy",
		"payments-prod/general-policy",
		"other-ns/general-policy",
	} {
		if !got[want] {
			t.Fatalf("missing request %q in %#v", want, requests)
		}
	}
}

func TestRequestsForS3ConfigChangeQueuesNamespaceMatchesForGlobalSource(t *testing.T) {
	rec, _ := testReconciler(t,
		globalS3Config("general-policy", cascade.S3Section{}, cascade.S3Section{}),
		localS3Config("payments-prod", "general-policy", cascade.S3Section{}, cascade.S3Section{}),
		localS3Config("data-prod", "general-policy", cascade.S3Section{}, cascade.S3Section{}),
	)

	requests := rec.requestsForS3ConfigChange(context.Background(), &v1alpha1.AWSS3Config{
		ObjectMeta: metav1.ObjectMeta{Name: "general-policy", Namespace: kroSystemNamespace},
	})

	if len(requests) != 2 {
		t.Fatalf("requests len = %d, want 2", len(requests))
	}
}
