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

package acmconfig

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
	cfg := &v1alpha1.ACMConfig{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("ACMConfig"))
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
		For(&v1alpha1.ACMConfig{}).
		Watches(
			&v1alpha1.ACMConfig{},
			handler.EnqueueRequestsFromMapFunc(r.requestsForACMConfigChange),
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

func (r *Reconciler) reconcile(ctx context.Context, cfg *v1alpha1.ACMConfig) (bool, ctrl.Result, error) {
	globalNS := util.ResolveGlobalNamespace(ctx, r.Client, cfg.Namespace)
	globalKropath, err := r.loadKropathConfig(ctx, globalNS, cfg.Name)
	if err != nil {
		return false, ctrl.Result{}, err
	}
	localKropath, err := r.loadKropathConfig(ctx, cfg.Namespace, "default")
	if err != nil {
		return false, ctrl.Result{}, err
	}
	globalACM, err := r.loadACMConfig(ctx, globalNS, cfg.Name)
	if err != nil {
		return false, ctrl.Result{}, err
	}

	cfg.Status.ObservedGeneration = cfg.Generation

	// Augment each ACMKropathSection with the tier-level tags so they flow through
	// MergeACMCascade. CertificateManager family section carries only keyAlgorithm
	// and certificateTransparencyLogging; tier-level tags are promoted here.
	globalKropathMandatoryACM := globalKropath.Spec.Mandatory.CertificateManager
	globalKropathMandatoryACM.Tags = globalKropath.Spec.Mandatory.Tags
	localKropathMandatoryACM := localKropath.Spec.Mandatory.CertificateManager
	localKropathMandatoryACM.Tags = localKropath.Spec.Mandatory.Tags
	localKropathDefaultsACM := localKropath.Spec.Defaults.CertificateManager
	localKropathDefaultsACM.Tags = localKropath.Spec.Defaults.Tags
	globalKropathDefaultsACM := globalKropath.Spec.Defaults.CertificateManager
	globalKropathDefaultsACM.Tags = globalKropath.Spec.Defaults.Tags

	eff := cascade.MergeACMCascade(
		globalKropathMandatoryACM,
		localKropathMandatoryACM,
		globalACM.Spec.Mandatory,
		cfg.Spec.Mandatory,
		cfg.Spec.Defaults,
		globalACM.Spec.Defaults,
		localKropathDefaultsACM,
		globalKropathDefaultsACM,
	)

	now := metav1.Now()

	newCond := metav1.Condition{
		Type:               "Valid",
		Status:             metav1.ConditionTrue,
		Reason:             "ValidationPassed",
		Message:            "",
		ObservedGeneration: cfg.Generation,
		LastTransitionTime: now,
	}
	newEffConfig := v1alpha1.EffectiveACMConfig{
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

func (r *Reconciler) loadACMConfig(ctx context.Context, namespace, name string) (*v1alpha1.ACMConfig, error) {
	cfg := &v1alpha1.ACMConfig{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("ACMConfig"))
	if err := r.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, cfg); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return &v1alpha1.ACMConfig{}, nil
		}
		return nil, err
	}
	return cfg, nil
}

func (r *Reconciler) requestsForKropathConfigChange(ctx context.Context, obj client.Object) []ctrl.Request {
	kpc, ok := obj.(*v1alpha1.KropathConfig)
	if !ok {
		return nil
	}

	var list v1alpha1.ACMConfigList
	if err := r.Client.List(ctx, &list); err != nil {
		r.Log.Error(err, "unable to list ACMConfig configs for KropathConfig change")
		return nil
	}

	requests := make([]ctrl.Request, 0, len(list.Items))
	for _, item := range list.Items {
		if kpc.Name == "default" && kpc.Namespace == item.Namespace {
			requests = append(requests, ctrl.Request{
				NamespacedName: types.NamespacedName{Namespace: item.Namespace, Name: item.Name},
			})
			continue
		}
		globalNS := util.ResolveGlobalNamespace(ctx, r.Client, item.Namespace)
		if kpc.Namespace == globalNS && kpc.Name == item.Name {
			requests = append(requests, ctrl.Request{
				NamespacedName: types.NamespacedName{Namespace: item.Namespace, Name: item.Name},
			})
		}
	}
	return requests
}

func (r *Reconciler) requestsForACMConfigChange(ctx context.Context, obj client.Object) []ctrl.Request {
	trigger, ok := obj.(*v1alpha1.ACMConfig)
	if !ok {
		return nil
	}

	var list v1alpha1.ACMConfigList
	if err := r.Client.List(ctx, &list); err != nil {
		r.Log.Error(err, "unable to list ACMConfig configs for ACMConfig change")
		return nil
	}

	requests := make([]ctrl.Request, 0, len(list.Items))
	for _, item := range list.Items {
		if item.Namespace == trigger.Namespace {
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

func (r *Reconciler) requestsForNamespaceChange(ctx context.Context, obj client.Object) []ctrl.Request {
	ns, ok := obj.(*corev1.Namespace)
	if !ok {
		return nil
	}

	var list v1alpha1.ACMConfigList
	if err := r.Client.List(ctx, &list, client.InNamespace(ns.Name)); err != nil {
		r.Log.Error(err, "unable to list ACMConfig configs for namespace change", "namespace", ns.Name)
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

func conditionNeedsUpdate(conditions []metav1.Condition, new metav1.Condition) bool {
	for _, c := range conditions {
		if c.Type == new.Type {
			return c.Status != new.Status || c.Reason != new.Reason || c.Message != new.Message
		}
	}
	return true
}

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
