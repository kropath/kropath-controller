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
	"reflect"
	"time"

	"github.com/go-logr/logr"
	"github.com/kropath/kropath-controller/api/v1alpha1"
	"github.com/kropath/kropath-controller/internal/cascade"
	"github.com/kropath/kropath-controller/internal/reconciler/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

type Reconciler struct {
	Client client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	cfg := &v1alpha1.S3Config{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("S3Config"))
	if err := r.Client.Get(ctx, req.NamespacedName, cfg); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	updated, result, err := r.reconcile(ctx, cfg)
	if err != nil {
		return ctrl.Result{}, err
	}
	if !updated {
		return result, nil
	}
	if err := r.Client.Status().Update(ctx, cfg); err != nil {
		return ctrl.Result{}, err
	}
	return result, nil
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.S3Config{}).
		Watches(
			&v1alpha1.S3Config{},
			handler.EnqueueRequestsFromMapFunc(r.requestsForS3ConfigChange),
		).
		Watches(
			&v1alpha1.KropathConfig{},
			handler.EnqueueRequestsFromMapFunc(r.requestsForKropathConfigChange),
		).
		Watches(
			&corev1.Namespace{},
			handler.EnqueueRequestsFromMapFunc(r.requestsForNamespaceChange),
		).
		Complete(r)
}

func (r *Reconciler) reconcile(ctx context.Context, cfg *v1alpha1.S3Config) (bool, ctrl.Result, error) {
	// Gap 1: dynamic global namespace resolution via annotation
	globalNS := util.ResolveGlobalNamespace(ctx, r.Client, cfg.Namespace)

	globalKropath, err := r.loadKropathConfig(ctx, globalNS, cfg.Name)
	if err != nil {
		return false, ctrl.Result{}, err
	}
	// Gap 2: local KropathConfig is always named "default"
	localKropath, err := r.loadKropathConfig(ctx, cfg.Namespace, "default")
	if err != nil {
		return false, ctrl.Result{}, err
	}
	globalS3, err := r.loadS3Config(ctx, globalNS, cfg.Name)
	if err != nil {
		return false, ctrl.Result{}, err
	}

	cfg.Status.ObservedGeneration = cfg.Generation

	// Gap 3: augment each S3Section with the tier-level tags so they flow
	// through MergeS3Cascade (mirrors the SQS tag augmentation pattern).
	globalKropathMandatoryS3 := globalKropath.Spec.Mandatory.S3
	globalKropathMandatoryS3.Tags = globalKropath.Spec.Mandatory.Tags
	localKropathMandatoryS3 := localKropath.Spec.Mandatory.S3
	localKropathMandatoryS3.Tags = localKropath.Spec.Mandatory.Tags
	localKropathDefaultsS3 := localKropath.Spec.Defaults.S3
	localKropathDefaultsS3.Tags = localKropath.Spec.Defaults.Tags
	globalKropathDefaultsS3 := globalKropath.Spec.Defaults.S3
	globalKropathDefaultsS3.Tags = globalKropath.Spec.Defaults.Tags

	eff := cascade.MergeS3Cascade(
		globalKropathMandatoryS3,  // level 1 — KropathConfig global mandatory
		localKropathMandatoryS3,   // level 2 — KropathConfig local mandatory
		globalS3.Spec.Mandatory,   // level 3 — S3Config global mandatory
		cfg.Spec.Mandatory,        // level 4 — S3Config local mandatory
		cfg.Spec.Defaults,         // level 6 — S3Config local defaults
		globalS3.Spec.Defaults,    // level 7 — S3Config global defaults
		localKropathDefaultsS3,    // level 8 — KropathConfig local defaults
		globalKropathDefaultsS3,   // level 9 — KropathConfig global defaults
	)

	now := metav1.Now()

	newCond := metav1.Condition{
		Type:               "Reconciled",
		Status:             metav1.ConditionTrue,
		Reason:             "CascadeMerged",
		Message:            "",
		ObservedGeneration: cfg.Generation,
		LastTransitionTime: now,
	}
	newEffConfig := v1alpha1.EffectiveS3Config{
		AWS:       mergeAWSIdentity(localKropath.Spec.AWS, globalKropath.Spec.AWS),
		Mandatory: eff.Mandatory,
		Defaults:  eff.Defaults,
	}

	if !conditionNeedsUpdate(cfg.Status.Conditions, newCond) &&
		reflect.DeepEqual(cfg.Status.EffectiveConfig, newEffConfig) {
		return false, ctrl.Result{}, nil
	}

	cfg.Status.Conditions = setCondition(cfg.Status.Conditions, newCond)
	cfg.Status.EffectiveConfig = newEffConfig
	cfg.Status.SyncedTimestamp = now.UTC().Format(time.RFC3339)

	return true, ctrl.Result{}, nil
}

func (r *Reconciler) loadKropathConfig(ctx context.Context, namespace, name string) (*v1alpha1.KropathConfig, error) {
	cfg := &v1alpha1.KropathConfig{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("KropathConfig"))
	if err := r.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, cfg); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return &v1alpha1.KropathConfig{}, nil
		}
		return nil, err
	}
	return cfg, nil
}

func (r *Reconciler) loadS3Config(ctx context.Context, namespace, name string) (*v1alpha1.S3Config, error) {
	cfg := &v1alpha1.S3Config{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("S3Config"))
	if err := r.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, cfg); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return &v1alpha1.S3Config{}, nil
		}
		return nil, err
	}
	return cfg, nil
}

// requestsForKropathConfigChange re-enqueues S3Configs when a KropathConfig changes.
// - A KPC named "default" in namespace X is a local KPC: enqueue all S3Configs in X.
// - Any other KPC is a global KPC: enqueue S3Configs whose resolved global namespace
//   matches kpc.Namespace and whose name matches kpc.Name.
func (r *Reconciler) requestsForKropathConfigChange(ctx context.Context, obj client.Object) []ctrl.Request {
	kpc, ok := obj.(*v1alpha1.KropathConfig)
	if !ok {
		return nil
	}

	var list v1alpha1.S3ConfigList
	if err := r.Client.List(ctx, &list); err != nil {
		r.Log.Error(err, "unable to list S3 configs for KropathConfig change")
		return nil
	}

	requests := make([]ctrl.Request, 0, len(list.Items))
	for _, item := range list.Items {
		// Local KPC: named "default" in the resource's own namespace
		if kpc.Name == "default" && kpc.Namespace == item.Namespace {
			requests = append(requests, ctrl.Request{
				NamespacedName: types.NamespacedName{Namespace: item.Namespace, Name: item.Name},
			})
			continue
		}
		// Global KPC: in the resolved global namespace for this item
		globalNS := util.ResolveGlobalNamespace(ctx, r.Client, item.Namespace)
		if kpc.Namespace == globalNS && kpc.Name == item.Name {
			requests = append(requests, ctrl.Request{
				NamespacedName: types.NamespacedName{Namespace: item.Namespace, Name: item.Name},
			})
		}
	}
	return requests
}

// requestsForS3ConfigChange re-enqueues namespaced S3Configs when a global S3Config changes.
// A global S3Config is one residing in the resolved global namespace of the target resource.
func (r *Reconciler) requestsForS3ConfigChange(ctx context.Context, obj client.Object) []ctrl.Request {
	trigger, ok := obj.(*v1alpha1.S3Config)
	if !ok {
		return nil
	}

	var list v1alpha1.S3ConfigList
	if err := r.Client.List(ctx, &list); err != nil {
		r.Log.Error(err, "unable to list S3 configs for S3Config change")
		return nil
	}

	requests := make([]ctrl.Request, 0, len(list.Items))
	for _, item := range list.Items {
		if item.Namespace == trigger.Namespace {
			// Same namespace — this is the local config itself, already reconciled directly.
			continue
		}
		globalNS := util.ResolveGlobalNamespace(ctx, r.Client, item.Namespace)
		if trigger.Namespace == globalNS && trigger.Name == item.Name {
			requests = append(requests, ctrl.Request{
				NamespacedName: types.NamespacedName{Namespace: item.Namespace, Name: item.Name},
			})
		}
	}
	return requests
}

// requestsForNamespaceChange re-enqueues all S3Configs in a namespace when the namespace
// annotation changes (affecting global namespace resolution).
func (r *Reconciler) requestsForNamespaceChange(ctx context.Context, obj client.Object) []ctrl.Request {
	ns, ok := obj.(*corev1.Namespace)
	if !ok {
		return nil
	}

	var list v1alpha1.S3ConfigList
	if err := r.Client.List(ctx, &list, client.InNamespace(ns.Name)); err != nil {
		r.Log.Error(err, "unable to list S3 configs for namespace change", "namespace", ns.Name)
		return nil
	}

	requests := make([]ctrl.Request, 0, len(list.Items))
	for _, item := range list.Items {
		requests = append(requests, ctrl.Request{
			NamespacedName: types.NamespacedName{Namespace: item.Namespace, Name: item.Name},
		})
	}
	return requests
}

func mergeAWSIdentity(local, global v1alpha1.ProviderIdentity) v1alpha1.ProviderIdentity {
	out := global
	if local.AccountID != "" {
		out.AccountID = local.AccountID
	}
	if local.Region != "" {
		out.Region = local.Region
	}
	return out
}

// conditionNeedsUpdate returns true when no existing condition matches Type+Status+Reason+Message.
func conditionNeedsUpdate(conditions []metav1.Condition, new metav1.Condition) bool {
	for _, c := range conditions {
		if c.Type == new.Type {
			return c.Status != new.Status || c.Reason != new.Reason || c.Message != new.Message
		}
	}
	return true
}

// setCondition upserts a condition by Type, preserving LastTransitionTime when status is unchanged.
func setCondition(conditions []metav1.Condition, new metav1.Condition) []metav1.Condition {
	for i, c := range conditions {
		if c.Type == new.Type {
			if c.Status == new.Status {
				new.LastTransitionTime = c.LastTransitionTime
			}
			conditions[i] = new
			return conditions
		}
	}
	return append(conditions, new)
}
