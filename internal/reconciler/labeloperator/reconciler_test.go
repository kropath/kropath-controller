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

package labeloperator

import (
	"context"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	awsGVK      = schema.GroupVersionKind{Group: "aws.kropath.run", Version: "v1alpha1", Kind: "S3Config"}
	awsLabelKey = "aws.kropath.run/resource-name"
	gcpGVK      = schema.GroupVersionKind{Group: "gcp.kropath.run", Version: "v1alpha1", Kind: "CloudStorageBucketConfig"}
	gcpLabelKey = "gcp.kropath.run/resource-name"
)

func makeResource(gvk schema.GroupVersionKind, name, ns string, labels map[string]string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(gvk)
	obj.SetName(name)
	obj.SetNamespace(ns)
	if labels != nil {
		obj.SetLabels(labels)
	}
	return obj
}

func buildReconciler(gvk schema.GroupVersionKind, labelKey string, objs ...client.Object) (*Reconciler, client.Client) {
	c := fake.NewClientBuilder().WithObjects(objs...).Build()
	return &Reconciler{
		Client:   c,
		Log:      log.Log,
		GVK:      gvk,
		LabelKey: labelKey,
	}, c
}

func getLabel(t *testing.T, c client.Client, gvk schema.GroupVersionKind, ns, name, key string) string {
	t.Helper()
	got := &unstructured.Unstructured{}
	got.SetGroupVersionKind(gvk)
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: ns, Name: name}, got); err != nil {
		t.Fatalf("Get %s/%s: %v", ns, name, err)
	}
	return got.GetLabels()[key]
}

// ── needsPatch ────────────────────────────────────────────────────────────────

func TestNeedsPatch_NilLabels_ReturnsTrue(t *testing.T) {
	obj := makeResource(awsGVK, "bucket", "ns", nil)
	if !needsPatch(obj, awsLabelKey) {
		t.Fatal("expected needsPatch=true when labels are nil")
	}
}

func TestNeedsPatch_LabelAbsent_ReturnsTrue(t *testing.T) {
	obj := makeResource(awsGVK, "bucket", "ns", map[string]string{"other": "val"})
	if !needsPatch(obj, awsLabelKey) {
		t.Fatal("expected needsPatch=true when label key is absent")
	}
}

func TestNeedsPatch_LabelWrongValue_ReturnsTrue(t *testing.T) {
	obj := makeResource(awsGVK, "correct-name", "ns", map[string]string{awsLabelKey: "wrong-value"})
	if !needsPatch(obj, awsLabelKey) {
		t.Fatal("expected needsPatch=true when label value doesn't match name")
	}
}

func TestNeedsPatch_LabelCorrect_ReturnsFalse(t *testing.T) {
	obj := makeResource(awsGVK, "general-policy", "ns", map[string]string{awsLabelKey: "general-policy"})
	if needsPatch(obj, awsLabelKey) {
		t.Fatal("expected needsPatch=false when label already correct")
	}
}

// ── buildLabelPatch ───────────────────────────────────────────────────────────

func TestBuildLabelPatch_ProducesValidJSON(t *testing.T) {
	got := string(buildLabelPatch("aws.kropath.run/resource-name", "my-bucket"))
	want := `{"metadata":{"labels":{"aws.kropath.run/resource-name":"my-bucket"}}}`
	if got != want {
		t.Fatalf("buildLabelPatch = %q, want %q", got, want)
	}
}

func TestBuildLabelPatch_EscapesQuotes(t *testing.T) {
	got := string(buildLabelPatch("k", `val"ue`))
	if !strings.Contains(got, `val\"ue`) {
		t.Fatalf("buildLabelPatch did not escape inner quote: %q", got)
	}
}

// ── Reconcile ─────────────────────────────────────────────────────────────────

func TestReconcile_LabelAbsent_PatchApplied(t *testing.T) {
	obj := makeResource(awsGVK, "general-policy", "payments-prod", nil)
	r, c := buildReconciler(awsGVK, awsLabelKey, obj)

	if _, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "payments-prod", Name: "general-policy"},
	}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	if v := getLabel(t, c, awsGVK, "payments-prod", "general-policy", awsLabelKey); v != "general-policy" {
		t.Fatalf("label = %q, want general-policy", v)
	}
}

func TestReconcile_LabelWrongValue_PatchApplied(t *testing.T) {
	obj := makeResource(awsGVK, "correct-name", "ns", map[string]string{awsLabelKey: "wrong-value"})
	r, c := buildReconciler(awsGVK, awsLabelKey, obj)

	if _, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "ns", Name: "correct-name"},
	}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	if v := getLabel(t, c, awsGVK, "ns", "correct-name", awsLabelKey); v != "correct-name" {
		t.Fatalf("label = %q, want correct-name", v)
	}
}

func TestReconcile_LabelCorrect_NoOp(t *testing.T) {
	obj := makeResource(awsGVK, "general-policy", "ns", map[string]string{awsLabelKey: "general-policy"})
	r, c := buildReconciler(awsGVK, awsLabelKey, obj)

	if _, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "ns", Name: "general-policy"},
	}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	// Label must remain correct after the no-op reconcile.
	if v := getLabel(t, c, awsGVK, "ns", "general-policy", awsLabelKey); v != "general-policy" {
		t.Fatalf("label = %q after no-op reconcile, want general-policy", v)
	}
}

func TestReconcile_ResourceNotFound_NoError(t *testing.T) {
	r, _ := buildReconciler(awsGVK, awsLabelKey) // empty store
	_, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "ns", Name: "missing"},
	})
	if err != nil {
		t.Fatalf("expected no error for missing resource, got: %v", err)
	}
}

func TestReconcile_GCPLabelKey_Correct(t *testing.T) {
	obj := makeResource(gcpGVK, "standard-tier", "kro-system", nil)
	r, c := buildReconciler(gcpGVK, gcpLabelKey, obj)

	if _, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "kro-system", Name: "standard-tier"},
	}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	if v := getLabel(t, c, gcpGVK, "kro-system", "standard-tier", gcpLabelKey); v != "standard-tier" {
		t.Fatalf("gcp label = %q, want standard-tier", v)
	}
}
