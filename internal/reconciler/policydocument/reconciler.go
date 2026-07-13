// Copyright 2026 The kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package policydocument

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"

	"github.com/kropath/kropath-controller/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

const defaultRequeueAfter = 10 * time.Second

// countInt32 clamps a slice length to int32 range, guarding the status-field
// conversion since a document could theoretically hold more entries than
// int32 can represent.
func countInt32(n int) int32 {
	if n > math.MaxInt32 {
		return math.MaxInt32
	}
	return int32(n) // #nosec G115 -- bounds-checked above, n <= math.MaxInt32
}

type Reconciler struct {
	client.Client
	RequeueAfter time.Duration
	ResolveRefFn func(context.Context, client.Client, string, *v1alpha1.PolicyRef) (string, bool, error)
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	builder := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.AWSPolicyDocument{}).
		Watches(&v1alpha1.AWSPolicyDocument{}, handler.EnqueueRequestsFromMapFunc(r.mapSourceDocumentUpdates))

	for _, kind := range policyRefWatchKinds() {
		prototype := &unstructured.Unstructured{}
		prototype.SetGroupVersionKind(schema.GroupVersionKind{Group: "aws.kropath.run", Version: "v1alpha1", Kind: kind})
		builder = builder.Watches(prototype, handler.EnqueueRequestsFromMapFunc(r.mapReferencedResourceUpdates))
	}

	return builder.Complete(r)
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var doc v1alpha1.AWSPolicyDocument
	if err := r.Get(ctx, req.NamespacedName, &doc); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	originalStatus := snapshotStatus(doc.Status)
	result, err := r.reconcileDocument(ctx, &doc)
	if err != nil {
		return ctrl.Result{}, err
	}

	if reflect.DeepEqual(originalStatus, doc.Status) {
		return result, nil
	}

	if err := r.Status().Update(ctx, &doc); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	return result, nil
}

func snapshotStatus(status v1alpha1.AWSPolicyDocumentStatus) v1alpha1.AWSPolicyDocumentStatus {
	if len(status.Conditions) > 0 {
		status.Conditions = append([]metav1.Condition(nil), status.Conditions...)
	}
	return status
}

func (r *Reconciler) reconcileDocument(ctx context.Context, doc *v1alpha1.AWSPolicyDocument) (ctrl.Result, error) {
	doc.Status.ObservedGeneration = doc.Generation
	resolve := r.ResolveRefFn
	if resolve == nil {
		resolve = resolveRef
	}

	if strings.TrimSpace(doc.Spec.DocumentJSON) != "" {
		if !json.Valid([]byte(doc.Spec.DocumentJSON)) {
			doc.Status.ResolvedDocumentJSON = ""
			doc.Status.StatementCount = 0
			doc.Status.Conditions = setCondition(doc.Status.Conditions, readyCondition(metav1.ConditionFalse, "InvalidDocumentJSON", "spec.documentJSON is not valid JSON", doc.Generation))
			doc.Status.Conditions = setCondition(doc.Status.Conditions, sourceNotReadyCondition(metav1.ConditionFalse, "InvalidDocumentJSON", "spec.documentJSON is not valid JSON", doc.Generation))
			doc.Status.Conditions = setCondition(doc.Status.Conditions, sidConflictCondition(metav1.ConditionFalse, "InvalidDocumentJSON", "spec.documentJSON is not valid JSON", doc.Generation))
			return ctrl.Result{}, nil
		}

		doc.Status.ResolvedDocumentJSON = doc.Spec.DocumentJSON
		doc.Status.StatementCount = 0
		doc.Status.SourceCount = countInt32(len(doc.Spec.Sources))
		doc.Status.Conditions = setCondition(doc.Status.Conditions, readyCondition(metav1.ConditionTrue, "DocumentResolved", "spec.documentJSON copied to status.resolvedDocumentJSON", doc.Generation))
		doc.Status.Conditions = setCondition(doc.Status.Conditions, sourceNotReadyCondition(metav1.ConditionFalse, "DocumentResolved", "no referenced sources required", doc.Generation))
		doc.Status.Conditions = setCondition(doc.Status.Conditions, sidConflictCondition(metav1.ConditionFalse, "DocumentResolved", "no merged statements", doc.Generation))
		return ctrl.Result{}, nil
	}

	sourceStatements, sourceIssue, err := r.collectSourceStatements(ctx, doc)
	if err != nil {
		return ctrl.Result{}, err
	}
	if sourceIssue != nil {
		if sourceIssue.clearResolvedDocumentJSON {
			doc.Status.ResolvedDocumentJSON = ""
			doc.Status.StatementCount = 0
		}
		doc.Status.SourceCount = countInt32(len(doc.Spec.Sources))
		doc.Status.Conditions = setCondition(doc.Status.Conditions, readyCondition(metav1.ConditionFalse, sourceIssue.readyReason, sourceIssue.message, doc.Generation))
		doc.Status.Conditions = setCondition(doc.Status.Conditions, sourceNotReadyCondition(sourceIssue.sourceNotReadyStatus, sourceIssue.sourceNotReadyReason, sourceIssue.message, doc.Generation))
		doc.Status.Conditions = setCondition(doc.Status.Conditions, sidConflictCondition(metav1.ConditionFalse, sourceIssue.readyReason, sourceIssue.message, doc.Generation))
		return ctrl.Result{RequeueAfter: sourceIssue.requeueAfter}, nil
	}

	statements := make([]policyStatementJSON, 0, len(doc.Spec.Statements))
	for _, stmt := range doc.Spec.Statements {
		resolved, pending, err := resolveStatement(ctx, r.Client, doc.Namespace, stmt, resolve)
		if err != nil {
			return ctrl.Result{}, err
		}
		if pending {
			doc.Status.ResolvedDocumentJSON = ""
			doc.Status.StatementCount = 0
			doc.Status.SourceCount = countInt32(len(doc.Spec.Sources))
			doc.Status.Conditions = setCondition(doc.Status.Conditions, readyCondition(metav1.ConditionFalse, "SourceNotReady", "one or more refs are not ready", doc.Generation))
			doc.Status.Conditions = setCondition(doc.Status.Conditions, sourceNotReadyCondition(metav1.ConditionTrue, "SourceNotReady", "one or more refs are not ready", doc.Generation))
			doc.Status.Conditions = setCondition(doc.Status.Conditions, sidConflictCondition(metav1.ConditionFalse, "SourceNotReady", "ref resolution is still pending", doc.Generation))
			return ctrl.Result{RequeueAfter: r.requeueAfter()}, nil
		}
		statements = append(statements, resolved)
	}

	statements = append(sourceStatements, statements...)
	if sid, conflict := findSidConflict(statements); conflict {
		message := fmt.Sprintf("Sid %q appears more than once across merged statements", sid)
		doc.Status.Conditions = setCondition(doc.Status.Conditions, readyCondition(metav1.ConditionFalse, "SidConflict", message, doc.Generation))
		doc.Status.Conditions = setCondition(doc.Status.Conditions, sourceNotReadyCondition(metav1.ConditionFalse, "SidConflict", message, doc.Generation))
		doc.Status.Conditions = setCondition(doc.Status.Conditions, sidConflictCondition(metav1.ConditionTrue, "SidConflict", message, doc.Generation))
		return ctrl.Result{}, nil
	}

	doc.Status.StatementCount = countInt32(len(statements))
	doc.Status.SourceCount = countInt32(len(doc.Spec.Sources))

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

type sourceDocumentIssue struct {
	requeueAfter              time.Duration
	readyReason               string
	message                   string
	sourceNotReadyStatus      metav1.ConditionStatus
	sourceNotReadyReason      string
	clearResolvedDocumentJSON bool
}

func (r *Reconciler) collectSourceStatements(ctx context.Context, doc *v1alpha1.AWSPolicyDocument) ([]policyStatementJSON, *sourceDocumentIssue, error) {
	if len(doc.Spec.Sources) == 0 {
		return nil, nil, nil
	}

	statements := make([]policyStatementJSON, 0)
	for _, source := range doc.Spec.Sources {
		var sourceDoc v1alpha1.AWSPolicyDocument
		if err := r.Get(ctx, client.ObjectKey{Namespace: doc.Namespace, Name: source.Name}, &sourceDoc); err != nil {
			if apierrors.IsNotFound(err) {
				return nil, &sourceDocumentIssue{
					requeueAfter:              r.requeueAfter(),
					readyReason:               "SourceMissing",
					message:                   fmt.Sprintf("source document %q not found", source.Name),
					sourceNotReadyStatus:      metav1.ConditionTrue,
					sourceNotReadyReason:      "SourceMissing",
					clearResolvedDocumentJSON: true,
				}, nil
			}
			return nil, nil, err
		}

		if strings.TrimSpace(sourceDoc.Spec.DocumentJSON) != "" {
			return nil, &sourceDocumentIssue{
				readyReason:               "MergeFromRawNotSupported",
				message:                   fmt.Sprintf("source document %q uses spec.documentJSON and cannot be merged", source.Name),
				sourceNotReadyStatus:      metav1.ConditionFalse,
				sourceNotReadyReason:      "MergeFromRawNotSupported",
				clearResolvedDocumentJSON: false,
			}, nil
		}

		if strings.TrimSpace(sourceDoc.Status.ResolvedDocumentJSON) == "" {
			return nil, &sourceDocumentIssue{
				requeueAfter:              r.requeueAfter(),
				readyReason:               "SourcePending",
				message:                   fmt.Sprintf("source document %q has not resolved yet", source.Name),
				sourceNotReadyStatus:      metav1.ConditionTrue,
				sourceNotReadyReason:      "SourcePending",
				clearResolvedDocumentJSON: true,
			}, nil
		}

		sourceStatements, err := parseStatements(sourceDoc.Status.ResolvedDocumentJSON)
		if err != nil {
			return nil, nil, fmt.Errorf("parse resolved document for source %q: %w", source.Name, err)
		}
		statements = append(statements, sourceStatements...)
	}

	return statements, nil, nil
}

func parseStatements(raw string) ([]policyStatementJSON, error) {
	var doc policyDocumentJSON
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		return nil, err
	}
	return doc.Statement, nil
}

func findSidConflict(statements []policyStatementJSON) (string, bool) {
	seen := make(map[string]struct{}, len(statements))
	for _, stmt := range statements {
		sid := strings.TrimSpace(stmt.Sid)
		if sid == "" {
			continue
		}
		if _, ok := seen[sid]; ok {
			return sid, true
		}
		seen[sid] = struct{}{}
	}
	return "", false
}

func (r *Reconciler) mapSourceDocumentUpdates(ctx context.Context, obj client.Object) []ctrl.Request {
	if r == nil || r.Client == nil || obj == nil {
		return nil
	}

	references, err := r.requestsForSourceDocument(ctx, obj.GetNamespace(), obj.GetName())
	if err != nil {
		return nil
	}
	return references
}

func (r *Reconciler) mapReferencedResourceUpdates(ctx context.Context, obj client.Object) []ctrl.Request {
	if r == nil || r.Client == nil || obj == nil {
		return nil
	}

	references, err := r.requestsForReferencedResource(ctx, obj.GetNamespace(), obj.GetName(), obj.GetObjectKind().GroupVersionKind().Kind)
	if err != nil {
		return nil
	}
	return references
}

func (r *Reconciler) requestsForSourceDocument(ctx context.Context, namespace, sourceName string) ([]ctrl.Request, error) {
	var docs v1alpha1.AWSPolicyDocumentList
	if err := r.List(ctx, &docs, client.InNamespace(namespace)); err != nil {
		return nil, err
	}

	requests := make([]ctrl.Request, 0, len(docs.Items))
	for i := range docs.Items {
		doc := &docs.Items[i]
		if !documentReferencesSource(doc, sourceName) {
			continue
		}
		requests = append(requests, ctrl.Request{
			NamespacedName: client.ObjectKey{
				Namespace: doc.Namespace,
				Name:      doc.Name,
			},
		})
	}

	return requests, nil
}

func (r *Reconciler) requestsForReferencedResource(ctx context.Context, namespace, name, kind string) ([]ctrl.Request, error) {
	var docs v1alpha1.AWSPolicyDocumentList
	if err := r.List(ctx, &docs, client.InNamespace(namespace)); err != nil {
		return nil, err
	}

	requests := make([]ctrl.Request, 0, len(docs.Items))
	for i := range docs.Items {
		doc := &docs.Items[i]
		if !documentReferencesResource(doc, kind, name) {
			continue
		}
		requests = append(requests, ctrl.Request{
			NamespacedName: client.ObjectKey{
				Namespace: doc.Namespace,
				Name:      doc.Name,
			},
		})
	}

	return requests, nil
}

func documentReferencesSource(doc *v1alpha1.AWSPolicyDocument, sourceName string) bool {
	for _, source := range doc.Spec.Sources {
		if source.Name == sourceName {
			return true
		}
	}
	return false
}

func documentReferencesResource(doc *v1alpha1.AWSPolicyDocument, kind, name string) bool {
	for _, stmt := range doc.Spec.Statements {
		for _, principal := range stmt.Principals {
			if refMatches(principal.Ref, kind, name) {
				return true
			}
		}
		for _, resource := range stmt.Resources {
			if refMatches(resource.Ref, kind, name) {
				return true
			}
		}
	}
	return false
}

func refMatches(ref *v1alpha1.PolicyRef, kind, name string) bool {
	if ref == nil {
		return false
	}
	return ref.Kind == kind && ref.Name == name
}

func policyRefWatchKinds() []string {
	return []string{
		"AWSIAMRole",
		"AWSS3Bucket",
		"AWSLambdaFunction",
		"AWSSQSQueue",
		"AWSKMSKey",
		"AWSSecretsManagerSecret",
	}
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
