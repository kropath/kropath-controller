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

package elasticacheconfig

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
	cfg := &v1alpha1.ElastiCacheConfig{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("ElastiCacheConfig"))
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
		For(&v1alpha1.ElastiCacheConfig{}).
		Watches(
			&v1alpha1.ElastiCacheConfig{},
			handler.EnqueueRequestsFromMapFunc(r.requestsForElastiCacheConfigChange),
		).
		Watches(
			&v1alpha1.KropathConfig{},
			handler.EnqueueRequestsFromMapFunc(r.requestsForKropathConfigChange),
		).
		Complete(r)
}

func (r *Reconciler) reconcile(ctx context.Context, cfg *v1alpha1.ElastiCacheConfig) (bool, ctrl.Result, error) {
	globalKropath, err := r.loadKropathConfig(ctx, kroSystemNamespace, cfg.Name)
	if err != nil {
		return false, ctrl.Result{}, err
	}
	localKropath, err := r.loadKropathConfig(ctx, cfg.Namespace, cfg.Name)
	if err != nil {
		return false, ctrl.Result{}, err
	}
	globalEC, err := r.loadElastiCacheConfig(ctx, kroSystemNamespace, cfg.Name)
	if err != nil {
		return false, ctrl.Result{}, err
	}

	cfg.Status.ObservedGeneration = cfg.Generation

	// Augment each ElastiCacheKropathSection with the tier-level tags so they
	// flow through MergeElastiCacheCascade (mirrors the SQS/KMS tag augmentation pattern).
	globalKropathMandatoryEC := globalKropath.Spec.Mandatory.ElastiCache
	globalKropathMandatoryEC.Tags = globalKropath.Spec.Mandatory.Tags
	localKropathMandatoryEC := localKropath.Spec.Mandatory.ElastiCache
	localKropathMandatoryEC.Tags = localKropath.Spec.Mandatory.Tags
	localKropathDefaultsEC := localKropath.Spec.Defaults.ElastiCache
	localKropathDefaultsEC.Tags = localKropath.Spec.Defaults.Tags
	globalKropathDefaultsEC := globalKropath.Spec.Defaults.ElastiCache
	globalKropathDefaultsEC.Tags = globalKropath.Spec.Defaults.Tags

	eff := cascade.MergeElastiCacheCascade(
		globalKropathMandatoryEC,
		localKropathMandatoryEC,
		globalEC.Spec.Mandatory,
		cfg.Spec.Mandatory,
		cfg.Spec.Defaults,
		globalEC.Spec.Defaults,
		localKropathDefaultsEC,
		globalKropathDefaultsEC,
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
	newEffConfig := v1alpha1.EffectiveElastiCacheConfig{
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

func (r *Reconciler) loadElastiCacheConfig(ctx context.Context, namespace, name string) (*v1alpha1.ElastiCacheConfig, error) {
	cfg := &v1alpha1.ElastiCacheConfig{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("ElastiCacheConfig"))
	if err := r.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, cfg); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return &v1alpha1.ElastiCacheConfig{}, nil
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

	var list v1alpha1.ElastiCacheConfigList
	if cfg.Namespace == kroSystemNamespace {
		if err := r.Client.List(ctx, &list); err != nil {
			r.Log.Error(err, "unable to list ElastiCache configs for global KropathConfig change")
			return nil
		}
	} else {
		if err := r.Client.List(ctx, &list, client.InNamespace(cfg.Namespace)); err != nil {
			r.Log.Error(err, "unable to list ElastiCache configs for namespace KropathConfig change")
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

func (r *Reconciler) requestsForElastiCacheConfigChange(ctx context.Context, obj client.Object) []ctrl.Request {
	cfg, ok := obj.(*v1alpha1.ElastiCacheConfig)
	if !ok || cfg.Namespace != kroSystemNamespace {
		return nil
	}

	var list v1alpha1.ElastiCacheConfigList
	if err := r.Client.List(ctx, &list); err != nil {
		r.Log.Error(err, "unable to list ElastiCache configs for global ElastiCacheConfig change")
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

// conditionNeedsUpdate returns true when no existing condition matches Type+Status+Reason+Message,
// meaning a status write is necessary to reflect the new state.
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
