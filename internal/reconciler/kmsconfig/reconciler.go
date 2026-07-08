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

package kmsconfig

import (
	"context"
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
	cfg := &v1alpha1.AWSKMSConfig{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("AWSKMSConfig"))
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
		For(&v1alpha1.AWSKMSConfig{}).
		Watches(
			&v1alpha1.AWSKMSConfig{},
			handler.EnqueueRequestsFromMapFunc(r.requestsForKMSConfigChange),
		).
		Watches(
			&v1alpha1.AWSKropathConfig{},
			handler.EnqueueRequestsFromMapFunc(r.requestsForKropathConfigChange),
		).
		Complete(r)
}

func (r *Reconciler) reconcile(ctx context.Context, cfg *v1alpha1.AWSKMSConfig) (bool, ctrl.Result, error) {
	globalKropath, err := r.loadKropathConfig(ctx, kroSystemNamespace, cfg.Name)
	if err != nil {
		return false, ctrl.Result{}, err
	}
	localKropath, err := r.loadKropathConfig(ctx, cfg.Namespace, cfg.Name)
	if err != nil {
		return false, ctrl.Result{}, err
	}
	globalKMS, err := r.loadKMSConfig(ctx, kroSystemNamespace, cfg.Name)
	if err != nil {
		return false, ctrl.Result{}, err
	}

	cfg.Status.ObservedGeneration = cfg.Generation

	// Augment each KMSKropathSection with the tier-level tags (AC-15):
	// KMSKropathSection carries cross-type KMS fields only; tier-level tags from
	// AWSKropathConfigTier.Tags are promoted here so they flow through MergeKMSCascade.
	globalKropathMandatoryKMS := globalKropath.Spec.Mandatory.KMS
	globalKropathMandatoryKMS.Tags = globalKropath.Spec.Mandatory.Tags
	localKropathMandatoryKMS := localKropath.Spec.Mandatory.KMS
	localKropathMandatoryKMS.Tags = localKropath.Spec.Mandatory.Tags
	localKropathDefaultsKMS := localKropath.Spec.Defaults.KMS
	localKropathDefaultsKMS.Tags = localKropath.Spec.Defaults.Tags
	globalKropathDefaultsKMS := globalKropath.Spec.Defaults.KMS
	globalKropathDefaultsKMS.Tags = globalKropath.Spec.Defaults.Tags

	eff := cascade.MergeKMSCascade(
		globalKropathMandatoryKMS,
		localKropathMandatoryKMS,
		globalKMS.Spec.Mandatory,
		cfg.Spec.Mandatory,
		cfg.Spec.Defaults,
		globalKMS.Spec.Defaults,
		localKropathDefaultsKMS,
		globalKropathDefaultsKMS,
	)

	now := metav1.Now()

	valid, reason, message := cascade.ValidateKMSKeySpec(eff.Mandatory)
	if !valid {
		cfg.Status.Conditions = setCondition(cfg.Status.Conditions, metav1.Condition{
			Type:               "Valid",
			Status:             metav1.ConditionFalse,
			Reason:             reason,
			Message:            message,
			ObservedGeneration: cfg.Generation,
			LastTransitionTime: now,
		})
		cfg.Status.SyncedTimestamp = now.UTC().Format(time.RFC3339)
		return true, ctrl.Result{}, nil
	}

	cfg.Status.Conditions = setCondition(cfg.Status.Conditions, metav1.Condition{
		Type:               "Valid",
		Status:             metav1.ConditionTrue,
		Reason:             "ValidationPassed",
		Message:            "",
		ObservedGeneration: cfg.Generation,
		LastTransitionTime: now,
	})
	cfg.Status.EffectiveConfig = v1alpha1.AWSEffectiveKMSConfig{
		AWS:       mergeAWSIdentity(localKropath.Spec.AWS, globalKropath.Spec.AWS),
		Mandatory: eff.Mandatory,
		Defaults:  eff.Defaults,
	}
	cfg.Status.SyncedTimestamp = now.UTC().Format(time.RFC3339)

	return true, ctrl.Result{}, nil
}

func (r *Reconciler) loadKropathConfig(ctx context.Context, namespace, name string) (*v1alpha1.AWSKropathConfig, error) {
	cfg := &v1alpha1.AWSKropathConfig{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("AWSKropathConfig"))
	if err := r.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, cfg); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return &v1alpha1.AWSKropathConfig{}, nil
		}
		return nil, err
	}
	return cfg, nil
}

func (r *Reconciler) loadKMSConfig(ctx context.Context, namespace, name string) (*v1alpha1.AWSKMSConfig, error) {
	cfg := &v1alpha1.AWSKMSConfig{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("AWSKMSConfig"))
	if err := r.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, cfg); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return &v1alpha1.AWSKMSConfig{}, nil
		}
		return nil, err
	}
	return cfg, nil
}

func (r *Reconciler) requestsForKropathConfigChange(ctx context.Context, obj client.Object) []ctrl.Request {
	cfg, ok := obj.(*v1alpha1.AWSKropathConfig)
	if !ok {
		return nil
	}

	var list v1alpha1.AWSKMSConfigList
	if cfg.Namespace == kroSystemNamespace {
		if err := r.Client.List(ctx, &list); err != nil {
			r.Log.Error(err, "unable to list KMS configs for global KropathConfig change")
			return nil
		}
	} else {
		if err := r.Client.List(ctx, &list, client.InNamespace(cfg.Namespace)); err != nil {
			r.Log.Error(err, "unable to list KMS configs for namespace KropathConfig change")
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

func (r *Reconciler) requestsForKMSConfigChange(ctx context.Context, obj client.Object) []ctrl.Request {
	cfg, ok := obj.(*v1alpha1.AWSKMSConfig)
	if !ok || cfg.Namespace != kroSystemNamespace {
		return nil
	}

	var list v1alpha1.AWSKMSConfigList
	if err := r.Client.List(ctx, &list); err != nil {
		r.Log.Error(err, "unable to list KMS configs for global KMSConfig change")
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

func mergeAWSIdentity(local, global v1alpha1.AWSProviderIdentity) v1alpha1.AWSProviderIdentity {
	out := global
	if local.AccountID != "" {
		out.AccountID = local.AccountID
	}
	if local.Region != "" {
		out.Region = local.Region
	}
	return out
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
