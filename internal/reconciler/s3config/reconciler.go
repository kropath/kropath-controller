// Copyright 2026 The kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package s3config

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
	cfg := &v1alpha1.AWSS3Config{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("AWSS3Config"))
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
		For(&v1alpha1.AWSS3Config{}).
		Watches(
			&v1alpha1.AWSS3Config{},
			handler.EnqueueRequestsFromMapFunc(r.requestsForS3ConfigChange),
		).
		Watches(
			&v1alpha1.AWSKropathConfig{},
			handler.EnqueueRequestsFromMapFunc(r.requestsForKropathConfigChange),
		).
		Complete(r)
}

func (r *Reconciler) reconcile(ctx context.Context, cfg *v1alpha1.AWSS3Config) (bool, ctrl.Result, error) {
	globalKropath, err := r.loadKropathConfig(ctx, kroSystemNamespace, cfg.Name)
	if err != nil {
		return false, ctrl.Result{}, err
	}
	localKropath, err := r.loadKropathConfig(ctx, cfg.Namespace, cfg.Name)
	if err != nil {
		return false, ctrl.Result{}, err
	}
	globalS3, err := r.loadS3Config(ctx, kroSystemNamespace, cfg.Name)
	if err != nil {
		return false, ctrl.Result{}, err
	}

	cfg.Status.ObservedGeneration = cfg.Generation
	eff := cascade.MergeS3Cascade(
		s3SectionFromKropathTier(globalKropath.Spec.Mandatory),
		s3SectionFromKropathTier(localKropath.Spec.Mandatory),
		globalS3.Spec.Mandatory,
		cfg.Spec.Mandatory,
		cfg.Spec.Defaults,
		globalS3.Spec.Defaults,
		s3SectionFromKropathTier(localKropath.Spec.Defaults),
		s3SectionFromKropathTier(globalKropath.Spec.Defaults),
	)
	cfg.Status.EffectiveConfig = v1alpha1.AWSEffectiveS3Config{
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

func (r *Reconciler) loadS3Config(ctx context.Context, namespace, name string) (*v1alpha1.AWSS3Config, error) {
	cfg := &v1alpha1.AWSS3Config{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("AWSS3Config"))
	if err := r.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, cfg); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return &v1alpha1.AWSS3Config{}, nil
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

	var list v1alpha1.AWSS3ConfigList
	if cfg.Namespace == kroSystemNamespace {
		if err := r.Client.List(ctx, &list); err != nil {
			r.Log.Error(err, "unable to list S3 configs for global KropathConfig change")
			return nil
		}
	} else {
		if err := r.Client.List(ctx, &list, client.InNamespace(cfg.Namespace)); err != nil {
			r.Log.Error(err, "unable to list S3 configs for namespace KropathConfig change")
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

func (r *Reconciler) requestsForS3ConfigChange(ctx context.Context, obj client.Object) []ctrl.Request {
	cfg, ok := obj.(*v1alpha1.AWSS3Config)
	if !ok || cfg.Namespace != kroSystemNamespace {
		return nil
	}

	var list v1alpha1.AWSS3ConfigList
	if err := r.Client.List(ctx, &list); err != nil {
		r.Log.Error(err, "unable to list S3 configs for global S3Config change")
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

func s3SectionFromKropathTier(tier v1alpha1.AWSKropathConfigTier) cascade.S3Section {
	out := tier.S3
	if tier.NamingTemplate != "" {
		out.NamingTemplate = tier.NamingTemplate
	}
	out.Tags = mergeStringMaps(out.Tags, tier.Tags)
	out.SyncedLabels = mergeStringMaps(out.SyncedLabels, tier.SyncedLabels)
	out.SyncedAnnotations = mergeStringMaps(out.SyncedAnnotations, tier.SyncedAnnotations)
	return out
}

func mergeStringMaps(maps ...map[string]string) map[string]string {
	out := make(map[string]string)
	for _, input := range maps {
		for key, value := range input {
			out[key] = value
		}
	}
	return out
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
