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
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

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
