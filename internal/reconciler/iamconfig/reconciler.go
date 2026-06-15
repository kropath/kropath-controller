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

package iamconfig

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
	cfg := &v1alpha1.AWSIAMConfig{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("AWSIAMConfig"))
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
		For(&v1alpha1.AWSIAMConfig{}).
		Watches(
			&v1alpha1.AWSIAMConfig{},
			handler.EnqueueRequestsFromMapFunc(r.requestsForIAMConfigChange),
		).
		Watches(
			&v1alpha1.AWSKropathConfig{},
			handler.EnqueueRequestsFromMapFunc(r.requestsForKropathConfigChange),
		).
		Complete(r)
}

func (r *Reconciler) reconcile(ctx context.Context, cfg *v1alpha1.AWSIAMConfig) (bool, ctrl.Result, error) {
	globalKropath, err := r.loadKropathConfig(ctx, kroSystemNamespace, cfg.Name)
	if err != nil {
		return false, ctrl.Result{}, err
	}
	localKropath, err := r.loadKropathConfig(ctx, cfg.Namespace, cfg.Name)
	if err != nil {
		return false, ctrl.Result{}, err
	}
	globalIAM, err := r.loadIAMConfig(ctx, kroSystemNamespace, cfg.Name)
	if err != nil {
		return false, ctrl.Result{}, err
	}

	cfg.Status.ObservedGeneration = cfg.Generation
	eff := cascade.MergeIAMCascade(
		globalKropath.Spec.Mandatory.IAM,
		localKropath.Spec.Mandatory.IAM,
		globalIAM.Spec.Mandatory,
		cfg.Spec.Mandatory,
		cfg.Spec.Defaults,
		globalIAM.Spec.Defaults,
		localKropath.Spec.Defaults.IAM,
		globalKropath.Spec.Defaults.IAM,
	)
	cfg.Status.EffectiveConfig = v1alpha1.AWSEffectiveIAMConfig{
		AWS:       mergeAWSIdentity(localKropath.Spec.AWS, globalKropath.Spec.AWS),
		Mandatory: eff.Mandatory,
		Defaults:  eff.Defaults,
	}
	cfg.Status.SyncedTimestamp = metav1.Now().UTC().Format(time.RFC3339)

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

func (r *Reconciler) loadIAMConfig(ctx context.Context, namespace, name string) (*v1alpha1.AWSIAMConfig, error) {
	cfg := &v1alpha1.AWSIAMConfig{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("AWSIAMConfig"))
	if err := r.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, cfg); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return &v1alpha1.AWSIAMConfig{}, nil
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

	var list v1alpha1.AWSIAMConfigList
	if cfg.Namespace == kroSystemNamespace {
		if err := r.Client.List(ctx, &list); err != nil {
			r.Log.Error(err, "unable to list IAM configs for global KropathConfig change")
			return nil
		}
	} else {
		if err := r.Client.List(ctx, &list, client.InNamespace(cfg.Namespace)); err != nil {
			r.Log.Error(err, "unable to list IAM configs for namespace KropathConfig change")
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

func (r *Reconciler) requestsForIAMConfigChange(ctx context.Context, obj client.Object) []ctrl.Request {
	cfg, ok := obj.(*v1alpha1.AWSIAMConfig)
	if !ok || cfg.Namespace != kroSystemNamespace {
		return nil
	}

	var list v1alpha1.AWSIAMConfigList
	if err := r.Client.List(ctx, &list); err != nil {
		r.Log.Error(err, "unable to list IAM configs for global IAMConfig change")
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
