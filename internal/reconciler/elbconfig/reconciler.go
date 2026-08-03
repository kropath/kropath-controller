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

package elbconfig

import (
	"context"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	"github.com/kropath/kropath-controller/api/v1alpha1"
	"github.com/kropath/kropath-controller/internal/cascade"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

const kroSystemNamespace = "kro-system"

type Reconciler struct {
	Client client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	cfg := &v1alpha1.ELBConfig{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("ELBConfig"))
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
		For(&v1alpha1.ELBConfig{}).
		Watches(
			&v1alpha1.ELBConfig{},
			handler.EnqueueRequestsFromMapFunc(r.requestsForELBConfigChange),
		).
		Watches(
			&v1alpha1.KropathConfig{},
			handler.EnqueueRequestsFromMapFunc(r.requestsForKropathConfigChange),
		).
		Complete(r)
}

func (r *Reconciler) reconcile(ctx context.Context, cfg *v1alpha1.ELBConfig) (bool, ctrl.Result, error) {
	globalKropath, err := r.loadKropathConfig(ctx, kroSystemNamespace, cfg.Name)
	if err != nil {
		return false, ctrl.Result{}, err
	}
	localKropath, err := r.loadKropathConfig(ctx, cfg.Namespace, cfg.Name)
	if err != nil {
		return false, ctrl.Result{}, err
	}
	globalELB, err := r.loadELBConfig(ctx, kroSystemNamespace, cfg.Name)
	if err != nil {
		return false, ctrl.Result{}, err
	}

	cfg.Status.ObservedGeneration = cfg.Generation

	// Augment each ELBKropathSection with tier-level tags so they flow
	// through MergeELBCascade (same pattern as KMS/SQS tag augmentation).
	globalKropathMandatoryELB := globalKropath.Spec.Mandatory.ELB
	globalKropathMandatoryELB.Tags = globalKropath.Spec.Mandatory.Tags
	localKropathMandatoryELB := localKropath.Spec.Mandatory.ELB
	localKropathMandatoryELB.Tags = localKropath.Spec.Mandatory.Tags
	localKropathDefaultsELB := localKropath.Spec.Defaults.ELB
	localKropathDefaultsELB.Tags = localKropath.Spec.Defaults.Tags
	globalKropathDefaultsELB := globalKropath.Spec.Defaults.ELB
	globalKropathDefaultsELB.Tags = globalKropath.Spec.Defaults.Tags

	eff := cascade.MergeELBCascade(
		globalKropathMandatoryELB,
		localKropathMandatoryELB,
		globalELB.Spec.Mandatory,
		cfg.Spec.Mandatory,
		cfg.Spec.Defaults,
		globalELB.Spec.Defaults,
		localKropathDefaultsELB,
		globalKropathDefaultsELB,
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
	newEffConfig := v1alpha1.EffectiveELBConfig{
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

func (r *Reconciler) loadELBConfig(ctx context.Context, namespace, name string) (*v1alpha1.ELBConfig, error) {
	cfg := &v1alpha1.ELBConfig{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("ELBConfig"))
	if err := r.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, cfg); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return &v1alpha1.ELBConfig{}, nil
		}
		return nil, err
	}
	return cfg, nil
}

func (r *Reconciler) requestsForKropathConfigChange(ctx context.Context, obj client.Object) []ctrl.Request {
	cfg, ok := obj.(*v1alpha1.KropathConfig)
	if !ok {
		return nil
	}

	var list v1alpha1.ELBConfigList
	if cfg.Namespace == kroSystemNamespace {
		if err := r.Client.List(ctx, &list); err != nil {
			r.Log.Error(err, "unable to list ELB configs for global KropathConfig change")
			return nil
		}
	} else {
		if err := r.Client.List(ctx, &list, client.InNamespace(cfg.Namespace)); err != nil {
			r.Log.Error(err, "unable to list ELB configs for namespace KropathConfig change")
			return nil
		}
	}

	requests := make([]ctrl.Request, 0, len(list.Items))
	for _, item := range list.Items {
		if item.Namespace == kroSystemNamespace || item.Name != cfg.Name {
			continue
		}
		requests = append(requests, ctrl.Request{NamespacedName: types.NamespacedName{Namespace: item.Namespace, Name: item.Name}})
	}
	return requests
}

func (r *Reconciler) requestsForELBConfigChange(ctx context.Context, obj client.Object) []ctrl.Request {
	cfg, ok := obj.(*v1alpha1.ELBConfig)
	if !ok || cfg.Namespace != kroSystemNamespace {
		return nil
	}

	var list v1alpha1.ELBConfigList
	if err := r.Client.List(ctx, &list); err != nil {
		r.Log.Error(err, "unable to list ELB configs for global ELBConfig change")
		return nil
	}

	requests := make([]ctrl.Request, 0, len(list.Items))
	for _, item := range list.Items {
		if item.Namespace == kroSystemNamespace || item.Name != cfg.Name {
			continue
		}
		requests = append(requests, ctrl.Request{NamespacedName: types.NamespacedName{Namespace: item.Namespace, Name: item.Name}})
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
