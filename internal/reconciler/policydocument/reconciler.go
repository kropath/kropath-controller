// Copyright 2026 The kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package policydocument

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/kropath/kropath-controller/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const defaultRequeueAfter = 10 * time.Second

type Reconciler struct {
	client.Client
	RequeueAfter time.Duration
	ResolveRefFn func(context.Context, client.Client, string, *v1alpha1.PolicyRef) (string, bool, error)
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.AWSPolicyDocument{}).
		Complete(r)
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var doc v1alpha1.AWSPolicyDocument
	if err := r.Get(ctx, req.NamespacedName, &doc); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	result, err := r.reconcileDocument(ctx, &doc)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := r.Status().Update(ctx, &doc); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	return result, nil
}

func (r *Reconciler) reconcileDocument(ctx context.Context, doc *v1alpha1.AWSPolicyDocument) (ctrl.Result, error) {
	doc.Status.ObservedGeneration = doc.Generation
	resolve := r.ResolveRefFn
	if resolve == nil {
		resolve = resolveRef
	}

	if strings.TrimSpace(doc.Spec.DocumentJSON) != "" {
		if !json.Valid([]byte(doc.Spec.DocumentJSON)) {
			doc.Status.Conditions = setCondition(doc.Status.Conditions, readyCondition(metav1.ConditionFalse, "InvalidDocumentJSON", "spec.documentJSON is not valid JSON", doc.Generation))
			doc.Status.Conditions = setCondition(doc.Status.Conditions, sourceNotReadyCondition(metav1.ConditionFalse, "InvalidDocumentJSON", "spec.documentJSON is not valid JSON", doc.Generation))
			doc.Status.Conditions = setCondition(doc.Status.Conditions, sidConflictCondition(metav1.ConditionFalse, "InvalidDocumentJSON", "spec.documentJSON is not valid JSON", doc.Generation))
			return ctrl.Result{}, nil
		}

		doc.Status.ResolvedDocumentJSON = doc.Spec.DocumentJSON
		doc.Status.StatementCount = 0
		doc.Status.SourceCount = int32(len(doc.Spec.Sources))
		doc.Status.Conditions = setCondition(doc.Status.Conditions, readyCondition(metav1.ConditionTrue, "DocumentResolved", "spec.documentJSON copied to status.resolvedDocumentJSON", doc.Generation))
		doc.Status.Conditions = setCondition(doc.Status.Conditions, sourceNotReadyCondition(metav1.ConditionFalse, "DocumentResolved", "no referenced sources required", doc.Generation))
		doc.Status.Conditions = setCondition(doc.Status.Conditions, sidConflictCondition(metav1.ConditionFalse, "DocumentResolved", "no merged statements", doc.Generation))
		return ctrl.Result{}, nil
	}

	statements := make([]policyStatementJSON, 0, len(doc.Spec.Statements))
	for _, stmt := range doc.Spec.Statements {
		resolved, pending, err := resolveStatement(ctx, r.Client, doc.Namespace, stmt, resolve)
		if err != nil {
			return ctrl.Result{}, err
		}
		if pending {
			doc.Status.Conditions = setCondition(doc.Status.Conditions, readyCondition(metav1.ConditionFalse, "SourceNotReady", "one or more refs are not ready", doc.Generation))
			doc.Status.Conditions = setCondition(doc.Status.Conditions, sourceNotReadyCondition(metav1.ConditionTrue, "SourceNotReady", "one or more refs are not ready", doc.Generation))
			doc.Status.Conditions = setCondition(doc.Status.Conditions, sidConflictCondition(metav1.ConditionFalse, "SourceNotReady", "ref resolution is still pending", doc.Generation))
			return ctrl.Result{RequeueAfter: r.requeueAfter()}, nil
		}
		statements = append(statements, resolved)
	}

	doc.Status.StatementCount = int32(len(statements))
	doc.Status.SourceCount = int32(len(doc.Spec.Sources))

	payload := map[string]any{
		"Version":   "2012-10-17",
		"Statement": statements,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return ctrl.Result{}, err
	}
	doc.Status.ResolvedDocumentJSON = string(raw)
	doc.Status.Conditions = setCondition(doc.Status.Conditions, readyCondition(metav1.ConditionTrue, "DocumentResolved", "structured policy document serialized", doc.Generation))
	doc.Status.Conditions = setCondition(doc.Status.Conditions, sourceNotReadyCondition(metav1.ConditionFalse, "DocumentResolved", "all refs resolved", doc.Generation))
	doc.Status.Conditions = setCondition(doc.Status.Conditions, sidConflictCondition(metav1.ConditionFalse, "DocumentResolved", "no sid conflicts detected", doc.Generation))
	return ctrl.Result{}, nil
}

func (r *Reconciler) requeueAfter() time.Duration {
	if r.RequeueAfter > 0 {
		return r.RequeueAfter
	}
	return defaultRequeueAfter
}

type policyDocumentJSON struct {
	Version   string                `json:"Version"`
	Statement []policyStatementJSON `json:"Statement"`
}

type policyStatementJSON struct {
	Sid       string         `json:"Sid,omitempty"`
	Effect    string         `json:"Effect"`
	Principal any            `json:"Principal,omitempty"`
	Action    any            `json:"Action"`
	Resource  any            `json:"Resource,omitempty"`
	Condition map[string]any `json:"Condition,omitempty"`
}

func resolveStatement(ctx context.Context, c client.Client, namespace string, stmt v1alpha1.PolicyStatement, resolve func(context.Context, client.Client, string, *v1alpha1.PolicyRef) (string, bool, error)) (policyStatementJSON, bool, error) {
	resolved := policyStatementJSON{
		Sid:    stmt.Sid,
		Effect: stmt.Effect,
		Action: normalizeStringList(stmt.Actions),
	}

	if len(stmt.Principals) > 0 {
		principal, pending, err := resolvePrincipals(ctx, c, namespace, stmt.Principals, resolve)
		if err != nil {
			return policyStatementJSON{}, false, err
		}
		if pending {
			return policyStatementJSON{}, true, nil
		}
		resolved.Principal = principal
	}

	if len(stmt.Resources) > 0 {
		resource, pending, err := resolveResources(ctx, c, namespace, stmt.Resources, resolve)
		if err != nil {
			return policyStatementJSON{}, false, err
		}
		if pending {
			return policyStatementJSON{}, true, nil
		}
		resolved.Resource = resource
	}

	if len(stmt.Conditions) > 0 {
		resolved.Condition = buildConditions(stmt.Conditions)
	}

	return resolved, false, nil
}

func resolvePrincipals(ctx context.Context, c client.Client, namespace string, principals []v1alpha1.PolicyPrincipal, resolve func(context.Context, client.Client, string, *v1alpha1.PolicyRef) (string, bool, error)) (any, bool, error) {
	grouped := map[string][]string{}
	for _, principal := range principals {
		if strings.TrimSpace(principal.ARN) != "" {
			grouped[principal.Type] = append(grouped[principal.Type], principal.ARN)
			continue
		}
		if principal.Ref == nil {
			return nil, false, fmt.Errorf("principal requires either arn or ref")
		}
		arn, pending, err := resolve(ctx, c, namespace, principal.Ref)
		if err != nil {
			return nil, false, err
		}
		if pending {
			return nil, true, nil
		}
		grouped[principal.Type] = append(grouped[principal.Type], arn)
	}
	return normalizePrincipalGroup(grouped), false, nil
}

func resolveResources(ctx context.Context, c client.Client, namespace string, resources []v1alpha1.PolicyResource, resolve func(context.Context, client.Client, string, *v1alpha1.PolicyRef) (string, bool, error)) (any, bool, error) {
	values := make([]string, 0, len(resources))
	for _, resource := range resources {
		if strings.TrimSpace(resource.ARN) != "" {
			values = append(values, resource.ARN)
			continue
		}
		if resource.Ref == nil {
			return nil, false, fmt.Errorf("resource requires either arn or ref")
		}
		arn, pending, err := resolve(ctx, c, namespace, resource.Ref)
		if err != nil {
			return nil, false, err
		}
		if pending {
			return nil, true, nil
		}
		values = append(values, arn)
	}
	return normalizeStringList(values), false, nil
}

func normalizePrincipalGroup(grouped map[string][]string) any {
	if len(grouped) == 1 {
		for key, values := range grouped {
			if len(values) == 1 {
				return map[string]any{key: values[0]}
			}
			return map[string]any{key: values}
		}
	}

	out := make(map[string]any, len(grouped))
	for key, values := range grouped {
		if len(values) == 1 {
			out[key] = values[0]
			continue
		}
		out[key] = values
	}
	return out
}

func normalizeStringList(values []string) any {
	switch len(values) {
	case 0:
		return []string{}
	case 1:
		return values[0]
	default:
		return values
	}
}

func buildConditions(conditions []v1alpha1.PolicyCondition) map[string]any {
	out := make(map[string]any, len(conditions))
	for _, condition := range conditions {
		values := normalizeStringList(condition.Values)
		operatorBlock, ok := out[condition.Operator]
		if !ok {
			out[condition.Operator] = map[string]any{condition.Key: values}
			continue
		}
		block, _ := operatorBlock.(map[string]any)
		block[condition.Key] = values
		out[condition.Operator] = block
	}
	return out
}
