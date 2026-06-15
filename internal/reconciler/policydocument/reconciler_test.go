// Copyright 2026 The kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package policydocument

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/kropath/kropath-controller/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type recordingClient struct {
	client.Client
	statusUpdates int
}

func (c *recordingClient) Status() client.StatusWriter {
	return &recordingStatusWriter{
		StatusWriter: c.Client.Status(),
		onUpdate: func() {
			c.statusUpdates++
		},
	}
}

type recordingStatusWriter struct {
	client.StatusWriter
	onUpdate func()
}

func (w *recordingStatusWriter) Update(ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption) error {
	w.onUpdate()
	return w.StatusWriter.Update(ctx, obj, opts...)
}

func TestReconcilePassesThroughDocumentJSON(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	doc := &v1alpha1.AWSPolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "kropath.run/v1alpha1", Kind: v1alpha1.AWSPolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "raw", Namespace: "default", Generation: 3},
		Spec: v1alpha1.AWSPolicyDocumentSpec{
			DocumentJSON: `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"s3:GetObject","Resource":"*"}]}`,
		},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(doc).Build()
	r := &Reconciler{Client: c}

	if _, err := r.reconcileDocument(context.Background(), doc); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	if doc.Status.ResolvedDocumentJSON != doc.Spec.DocumentJSON {
		t.Fatalf("resolved json mismatch: got %s", doc.Status.ResolvedDocumentJSON)
	}
	if !conditionHasStatus(doc.Status.Conditions, v1alpha1.ConditionReady, metav1.ConditionTrue) {
		t.Fatalf("ready condition not true: %#v", doc.Status.Conditions)
	}
}

func TestReconcileClearsResolvedDocumentOnInvalidDocumentJSON(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	doc := &v1alpha1.AWSPolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "kropath.run/v1alpha1", Kind: v1alpha1.AWSPolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "raw", Namespace: "default", Generation: 4},
		Spec: v1alpha1.AWSPolicyDocumentSpec{
			DocumentJSON: `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"s3:GetObject","Resource":"*"}]}`,
		},
		Status: v1alpha1.AWSPolicyDocumentStatus{
			ResolvedDocumentJSON: `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"s3:PutObject","Resource":"*"}]}`,
			StatementCount:       1,
		},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(doc).Build()
	r := &Reconciler{Client: c}

	if _, err := r.reconcileDocument(context.Background(), doc); err != nil {
		t.Fatalf("seed reconcile: %v", err)
	}

	doc.Spec.DocumentJSON = `{"Version":"2012-10-17",`
	if _, err := r.reconcileDocument(context.Background(), doc); err != nil {
		t.Fatalf("reconcile invalid json: %v", err)
	}

	if doc.Status.ResolvedDocumentJSON != "" {
		t.Fatalf("expected resolved json cleared, got %s", doc.Status.ResolvedDocumentJSON)
	}
	if doc.Status.StatementCount != 0 {
		t.Fatalf("expected statement count cleared, got %d", doc.Status.StatementCount)
	}
	if !conditionHasStatus(doc.Status.Conditions, v1alpha1.ConditionReady, metav1.ConditionFalse) {
		t.Fatalf("ready condition not false: %#v", doc.Status.Conditions)
	}
}

func TestReconcileSkipsStatusUpdateWhenStatusIsUnchanged(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	doc := &v1alpha1.AWSPolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "kropath.run/v1alpha1", Kind: v1alpha1.AWSPolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "raw", Namespace: "default", Generation: 3},
		Spec: v1alpha1.AWSPolicyDocumentSpec{
			DocumentJSON: `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"s3:GetObject","Resource":"*"}]}`,
		},
	}

	calc := &Reconciler{Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(doc).Build()}
	if _, err := calc.reconcileDocument(context.Background(), doc); err != nil {
		t.Fatalf("seed reconcile: %v", err)
	}

	stored := &v1alpha1.AWSPolicyDocument{
		TypeMeta:   doc.TypeMeta,
		ObjectMeta: doc.ObjectMeta,
		Spec:       doc.Spec,
		Status:     snapshotStatus(doc.Status),
	}

	baseClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(stored).Build()
	r := &Reconciler{Client: &recordingClient{Client: baseClient}}

	if _, err := r.Reconcile(context.Background(), ctrlRequest("default", "raw")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	if rc, ok := r.Client.(*recordingClient); !ok {
		t.Fatalf("unexpected client type %T", r.Client)
	} else if rc.statusUpdates != 0 {
		t.Fatalf("expected no status updates, got %d", rc.statusUpdates)
	}
}

func TestReconcileResolvesPredictedArnRef(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	role := &unstructured.Unstructured{}
	role.SetGroupVersionKind(schema.GroupVersionKind{Group: "kropath.run", Version: "v1alpha1", Kind: "AWSIAMRole"})
	role.SetNamespace("default")
	role.SetName("lambda-exec")
	if err := unstructured.SetNestedField(role.Object, "arn:aws:iam::123456789012:role/lambda-exec", "status", "predictedArn"); err != nil {
		t.Fatalf("seed role status: %v", err)
	}

	doc := &v1alpha1.AWSPolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "kropath.run/v1alpha1", Kind: v1alpha1.AWSPolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "policy", Namespace: "default", Generation: 5},
		Spec: v1alpha1.AWSPolicyDocumentSpec{
			Statements: []v1alpha1.PolicyStatement{
				{
					Sid:     "AllowLambdaInvoke",
					Effect:  "Allow",
					Actions: []string{"lambda:InvokeFunction"},
					Resources: []v1alpha1.PolicyResource{
						{Ref: &v1alpha1.PolicyRef{Kind: "AWSIAMRole", Name: "lambda-exec"}},
					},
				},
			},
		},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(doc, role).Build()
	r := &Reconciler{Client: c}

	if _, err := r.reconcileDocument(context.Background(), doc); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	if !jsonContains(doc.Status.ResolvedDocumentJSON, "arn:aws:iam::123456789012:role/lambda-exec") {
		t.Fatalf("resolved json missing arn: %s", doc.Status.ResolvedDocumentJSON)
	}
	if !conditionHasStatus(doc.Status.Conditions, v1alpha1.ConditionReady, metav1.ConditionTrue) {
		t.Fatalf("ready condition not true: %#v", doc.Status.Conditions)
	}
}

func TestReconcileMarksSourceNotReadyWhenRefPending(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	doc := &v1alpha1.AWSPolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "kropath.run/v1alpha1", Kind: v1alpha1.AWSPolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "policy", Namespace: "default", Generation: 2},
		Spec: v1alpha1.AWSPolicyDocumentSpec{
			Statements: []v1alpha1.PolicyStatement{
				{
					Effect:  "Allow",
					Actions: []string{"lambda:InvokeFunction"},
					Resources: []v1alpha1.PolicyResource{
						{Ref: &v1alpha1.PolicyRef{Kind: "AWSIAMRole", Name: "lambda-exec"}},
					},
				},
			},
		},
		Status: v1alpha1.AWSPolicyDocumentStatus{
			ResolvedDocumentJSON: `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"s3:GetObject","Resource":"*"}]}`,
			StatementCount:       1,
			SourceCount:          1,
		},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(doc).Build()
	r := &Reconciler{Client: c, RequeueAfter: 5 * time.Second}

	result, err := r.reconcileDocument(context.Background(), doc)
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if result.RequeueAfter != 5*time.Second {
		t.Fatalf("expected requeue after 5s, got %s", result.RequeueAfter)
	}

	if !conditionHasStatus(doc.Status.Conditions, v1alpha1.ConditionSourceNotReady, metav1.ConditionTrue) {
		t.Fatalf("source not ready not true: %#v", doc.Status.Conditions)
	}
	if conditionHasStatus(doc.Status.Conditions, v1alpha1.ConditionReady, metav1.ConditionTrue) {
		t.Fatalf("ready should be false: %#v", doc.Status.Conditions)
	}
	if doc.Status.ResolvedDocumentJSON != "" {
		t.Fatalf("expected resolved json cleared, got %s", doc.Status.ResolvedDocumentJSON)
	}
	if doc.Status.StatementCount != 0 {
		t.Fatalf("expected statement count cleared, got %d", doc.Status.StatementCount)
	}
}

func TestReconcileResolvesPrincipalRefToPredictedArn(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	bucket := &unstructured.Unstructured{}
	bucket.SetGroupVersionKind(schema.GroupVersionKind{Group: "kropath.run", Version: "v1alpha1", Kind: "AWSS3Bucket"})
	bucket.SetNamespace("default")
	bucket.SetName("my-bucket")
	if err := unstructured.SetNestedField(bucket.Object, "arn:aws:s3:::my-bucket", "status", "predictedArn"); err != nil {
		t.Fatalf("seed bucket status: %v", err)
	}

	doc := &v1alpha1.AWSPolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "kropath.run/v1alpha1", Kind: v1alpha1.AWSPolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "bucket-policy", Namespace: "default", Generation: 6},
		Spec: v1alpha1.AWSPolicyDocumentSpec{
			Statements: []v1alpha1.PolicyStatement{
				{
					Effect:  "Allow",
					Actions: []string{"s3:GetObject"},
					Principals: []v1alpha1.PolicyPrincipal{
						{Type: "AWS", Ref: &v1alpha1.PolicyRef{Kind: "AWSS3Bucket", Name: "my-bucket"}},
					},
				},
			},
		},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(doc, bucket).Build()
	r := &Reconciler{Client: c}

	if _, err := r.reconcileDocument(context.Background(), doc); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	if !jsonContains(doc.Status.ResolvedDocumentJSON, "arn:aws:s3:::my-bucket") {
		t.Fatalf("resolved json missing bucket arn: %s", doc.Status.ResolvedDocumentJSON)
	}
	if !jsonContains(doc.Status.ResolvedDocumentJSON, `"Principal":{"AWS":"arn:aws:s3:::my-bucket"}`) {
		t.Fatalf("resolved json missing principal: %s", doc.Status.ResolvedDocumentJSON)
	}
}

func TestReconcileSucceedsAfterRefBecomesReady(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	role := &unstructured.Unstructured{}
	role.SetGroupVersionKind(schema.GroupVersionKind{Group: "kropath.run", Version: "v1alpha1", Kind: "AWSIAMRole"})
	role.SetNamespace("default")
	role.SetName("lambda-exec")

	doc := &v1alpha1.AWSPolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "kropath.run/v1alpha1", Kind: v1alpha1.AWSPolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "policy", Namespace: "default", Generation: 7},
		Spec: v1alpha1.AWSPolicyDocumentSpec{
			Statements: []v1alpha1.PolicyStatement{
				{
					Effect:  "Allow",
					Actions: []string{"lambda:InvokeFunction"},
					Resources: []v1alpha1.PolicyResource{
						{Ref: &v1alpha1.PolicyRef{Kind: "AWSIAMRole", Name: "lambda-exec"}},
					},
				},
			},
		},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(doc, role).Build()
	r := &Reconciler{Client: c}

	if _, err := r.reconcileDocument(context.Background(), doc); err != nil {
		t.Fatalf("first reconcile: %v", err)
	}
	if !conditionHasStatus(doc.Status.Conditions, v1alpha1.ConditionSourceNotReady, metav1.ConditionTrue) {
		t.Fatalf("expected pending state first: %#v", doc.Status.Conditions)
	}

	if err := unstructured.SetNestedField(role.Object, "arn:aws:iam::123456789012:role/lambda-exec", "status", "predictedArn"); err != nil {
		t.Fatalf("seed role status: %v", err)
	}
	if err := c.Update(context.Background(), role); err != nil {
		t.Fatalf("update role: %v", err)
	}

	doc.Status = v1alpha1.AWSPolicyDocumentStatus{}
	if _, err := r.reconcileDocument(context.Background(), doc); err != nil {
		t.Fatalf("second reconcile: %v", err)
	}

	if !conditionHasStatus(doc.Status.Conditions, v1alpha1.ConditionReady, metav1.ConditionTrue) {
		t.Fatalf("expected ready after ref population: %#v", doc.Status.Conditions)
	}
	if !jsonContains(doc.Status.ResolvedDocumentJSON, "arn:aws:iam::123456789012:role/lambda-exec") {
		t.Fatalf("resolved json missing arn after ready: %s", doc.Status.ResolvedDocumentJSON)
	}
	if conditionHasStatus(doc.Status.Conditions, v1alpha1.ConditionSourceNotReady, metav1.ConditionTrue) {
		t.Fatalf("source not ready should clear: %#v", doc.Status.Conditions)
	}
}

func TestResolveRefUsesStatusArn(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "kropath.run", Version: "v1alpha1", Kind: "AWSIAMPolicy"})
	obj.SetNamespace("default")
	obj.SetName("policy")
	if err := unstructured.SetNestedField(obj.Object, "arn:aws:iam::123456789012:policy/example", "status", "arn"); err != nil {
		t.Fatalf("seed arn: %v", err)
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(obj).Build()
	arn, pending, err := resolveRef(context.Background(), c, "default", &v1alpha1.PolicyRef{Kind: "AWSIAMPolicy", Name: "policy", Field: "arn"})
	if err != nil {
		t.Fatalf("resolve ref: %v", err)
	}
	if pending {
		t.Fatalf("expected resolved ref")
	}
	if arn != "arn:aws:iam::123456789012:policy/example" {
		t.Fatalf("unexpected arn %q", arn)
	}
}

func TestRequestsForSourceDocumentReturnsParentsReferencingSource(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	source := &v1alpha1.AWSPolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "kropath.run/v1alpha1", Kind: v1alpha1.AWSPolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "base-policy", Namespace: "default"},
	}
	related := &v1alpha1.AWSPolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "kropath.run/v1alpha1", Kind: v1alpha1.AWSPolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "app-policy", Namespace: "default"},
		Spec: v1alpha1.AWSPolicyDocumentSpec{
			Sources: []v1alpha1.PolicyDocumentSource{{Name: "base-policy"}},
		},
	}
	unrelated := &v1alpha1.AWSPolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "kropath.run/v1alpha1", Kind: v1alpha1.AWSPolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "standalone", Namespace: "default"},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(source, related, unrelated).Build()
	r := &Reconciler{Client: c}

	requests, err := r.requestsForSourceDocument(context.Background(), "default", "base-policy")
	if err != nil {
		t.Fatalf("requests for source document: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("expected one parent request, got %d: %#v", len(requests), requests)
	}
	if requests[0].NamespacedName.Namespace != "default" || requests[0].NamespacedName.Name != "app-policy" {
		t.Fatalf("unexpected request: %#v", requests[0])
	}

	requests, err = r.requestsForSourceDocument(context.Background(), "default", "missing")
	if err != nil {
		t.Fatalf("requests for missing source document: %v", err)
	}
	if len(requests) != 0 {
		t.Fatalf("expected no parent requests for missing source, got %#v", requests)
	}
}

func TestRequestsForReferencedResourceReturnsMatchingDocuments(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	match := &v1alpha1.AWSPolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "kropath.run/v1alpha1", Kind: v1alpha1.AWSPolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "match", Namespace: "default"},
		Spec: v1alpha1.AWSPolicyDocumentSpec{
			Statements: []v1alpha1.PolicyStatement{
				{
					Effect:  "Allow",
					Actions: []string{"s3:GetObject"},
					Resources: []v1alpha1.PolicyResource{
						{Ref: &v1alpha1.PolicyRef{Kind: "AWSIAMRole", Name: "lambda-exec"}},
					},
				},
			},
		},
	}
	unrelated := &v1alpha1.AWSPolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "kropath.run/v1alpha1", Kind: v1alpha1.AWSPolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "unrelated", Namespace: "default"},
		Spec: v1alpha1.AWSPolicyDocumentSpec{
			Statements: []v1alpha1.PolicyStatement{
				{
					Effect:  "Allow",
					Actions: []string{"s3:GetObject"},
					Resources: []v1alpha1.PolicyResource{
						{Ref: &v1alpha1.PolicyRef{Kind: "AWSS3Bucket", Name: "my-bucket"}},
					},
				},
			},
		},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(match, unrelated).Build()
	r := &Reconciler{Client: c}

	requests, err := r.requestsForReferencedResource(context.Background(), "default", "lambda-exec", "AWSIAMRole")
	if err != nil {
		t.Fatalf("requests for referenced resource: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("expected one matching request, got %d: %#v", len(requests), requests)
	}
	if requests[0].NamespacedName.Name != "match" {
		t.Fatalf("unexpected request: %#v", requests[0])
	}

	requests, err = r.requestsForReferencedResource(context.Background(), "default", "lambda-exec", "AWSS3Bucket")
	if err != nil {
		t.Fatalf("requests for mismatched kind: %v", err)
	}
	if len(requests) != 0 {
		t.Fatalf("expected no requests for mismatched kind, got %#v", requests)
	}
}

func jsonContains(raw, needle string) bool {
	return strings.Contains(raw, needle)
}

func conditionHasStatus(conditions []metav1.Condition, conditionType string, status metav1.ConditionStatus) bool {
	for _, cond := range conditions {
		if cond.Type == conditionType && cond.Status == status {
			return true
		}
	}
	return false
}

func ctrlRequest(namespace, name string) ctrl.Request {
	return ctrl.Request{NamespacedName: client.ObjectKey{Namespace: namespace, Name: name}}
}
