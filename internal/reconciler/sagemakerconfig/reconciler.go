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

package sagemakerconfig

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
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

// Reconciler reconciles SageMakerConfig CRs and propagates effective config.
type Reconciler struct {
	Client client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	cfg := &v1alpha1.SageMakerConfig{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("SageMakerConfig"))
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
	_, err := r.BuildWithManager(mgr)
	return err
}

func (r *Reconciler) BuildWithManager(mgr ctrl.Manager) (controller.Controller, error) {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.SageMakerConfig{}).
		Watches(
			&v1alpha1.SageMakerConfig{},
			handler.EnqueueRequestsFromMapFunc(r.requestsForSageMakerConfigChange),
		).
		Watches(
			&v1alpha1.KropathConfig{},
			handler.EnqueueRequestsFromMapFunc(r.requestsForKropathConfigChange),
		).
		Watches(
			&corev1.Namespace{},
			handler.EnqueueRequestsFromMapFunc(r.requestsForNamespaceChange),
		).
		Build(r)
}

func (r *Reconciler) reconcile(ctx context.Context, cfg *v1alpha1.SageMakerConfig) (bool, ctrl.Result, error) {
	globalNS := util.ResolveGlobalNamespace(ctx, r.Client, cfg.Namespace)
	globalKropath, err := r.loadKropathConfig(ctx, globalNS, cfg.Name)
	if err != nil {
		return false, ctrl.Result{}, err
	}
	localKropath, err := r.loadKropathConfig(ctx, cfg.Namespace, "default")
	if err != nil {
		return false, ctrl.Result{}, err
	}
	globalSM, err := r.loadSageMakerConfig(ctx, globalNS, cfg.Name)
	if err != nil {
		return false, ctrl.Result{}, err
	}

	cfg.Status.ObservedGeneration = cfg.Generation

	// Augment each SageMakerKropathSection with tier-level tags so they flow
	// through MergeSageMakerCascade (mirrors the ECR/EKS tag augmentation pattern).
	globalKropathMandatorySM := globalKropath.Spec.Mandatory.SageMaker
	globalKropathMandatorySM.Tags = globalKropath.Spec.Mandatory.Tags
	localKropathMandatorySM := localKropath.Spec.Mandatory.SageMaker
	localKropathMandatorySM.Tags = localKropath.Spec.Mandatory.Tags
	localKropathDefaultsSM := localKropath.Spec.Defaults.SageMaker
	localKropathDefaultsSM.Tags = localKropath.Spec.Defaults.Tags
	globalKropathDefaultsSM := globalKropath.Spec.Defaults.SageMaker
	globalKropathDefaultsSM.Tags = globalKropath.Spec.Defaults.Tags

	eff := cascade.MergeSageMakerCascade(
		globalKropathMandatorySM,  // level 1
		localKropathMandatorySM,   // level 2
		globalSM.Spec.Mandatory,   // level 3
		cfg.Spec.Mandatory,        // level 4
		cfg.Spec.Defaults,         // level 6
		globalSM.Spec.Defaults,    // level 7
		localKropathDefaultsSM,    // level 8
		globalKropathDefaultsSM,   // level 9
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

	awsIdentity := mergeAWSIdentity(localKropath.Spec.AWS, globalKropath.Spec.AWS)
	newEffConfig := v1alpha1.EffectiveSageMakerConfigStatus{
		AWS:       v1alpha1.ProviderIdentity{AccountID: awsIdentity.AccountID, Region: awsIdentity.Region},
		Mandatory: eff.Mandatory,
		Defaults:  eff.Defaults,
	}

	condChanged := conditionNeedsUpdate(cfg.Status.Conditions, newCond)
	effChanged := !reflect.DeepEqual(cfg.Status.EffectiveConfig, newEffConfig)

	if !condChanged && !effChanged {
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

func (r *Reconciler) loadSageMakerConfig(ctx context.Context, namespace, name string) (*v1alpha1.SageMakerConfig, error) {
	cfg := &v1alpha1.SageMakerConfig{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("SageMakerConfig"))
	if err := r.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, cfg); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return &v1alpha1.SageMakerConfig{}, nil
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

	var list v1alpha1.SageMakerConfigList
	if err := r.Client.List(ctx, &list); err != nil {
		r.Log.Error(err, "unable to list SageMakerConfig for KropathConfig change")
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

func (r *Reconciler) requestsForSageMakerConfigChange(ctx context.Context, obj client.Object) []ctrl.Request {
	trigger, ok := obj.(*v1alpha1.SageMakerConfig)
	if !ok {
		return nil
	}

	var list v1alpha1.SageMakerConfigList
	if err := r.Client.List(ctx, &list); err != nil {
		r.Log.Error(err, "unable to list SageMakerConfig for SageMakerConfig change")
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

	var list v1alpha1.SageMakerConfigList
	if err := r.Client.List(ctx, &list, client.InNamespace(ns.Name)); err != nil {
		r.Log.Error(err, "unable to list SageMakerConfig for namespace change", "namespace", ns.Name)
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
