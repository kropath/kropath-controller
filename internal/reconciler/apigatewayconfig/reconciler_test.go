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

package apigatewayconfig

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kropath/kropath-controller/api/v1alpha1"
	"github.com/kropath/kropath-controller/internal/cascade"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	ctrl "sigs.k8s.io/controller-runtime"
)

// AC-1: globalKropathConfig.mandatory.apigateway.endpointType="REGIONAL" propagates (level 1 wins).
func TestReconcileAC1GlobalKropathEndpointTypeLevel1(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfig("general-policy", v1alpha1.KropathConfigTier{
			ApiGateway: cascade.ApiGatewayKropathSection{EndpointType: "REGIONAL"},
		}),
		localApiGatewayConfig("payments-prod", "general-policy", cascade.ApiGatewayConfigSection{}, cascade.ApiGatewayConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getApiGatewayConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Mandatory.EndpointType; got != "REGIONAL" {
		t.Fatalf("mandatory.endpointType = %q, want REGIONAL", got)
	}
}

// AC-2: globalApiGatewayConfig.mandatory.apiKeySource="HEADER" propagates (level 3 wins when L1-L2 absent).
func TestReconcileAC2GlobalApiGatewayConfigApiKeySourceLevel3(t *testing.T) {
	rec, _ := testReconciler(t,
		globalApiGatewayConfig("general-policy", cascade.ApiGatewayConfigSection{ApiKeySource: "HEADER"}, cascade.ApiGatewayConfigSection{}),
		localApiGatewayConfig("payments-prod", "general-policy", cascade.ApiGatewayConfigSection{}, cascade.ApiGatewayConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getApiGatewayConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Mandatory.ApiKeySource; got != "HEADER" {
		t.Fatalf("mandatory.apiKeySource = %q, want HEADER", got)
	}
}

// AC-3: localApiGatewayConfig.defaults.namingTemplate="{namespace}-{name}" propagates (level 6).
func TestReconcileAC3LocalApiGatewayConfigDefaultsNamingTemplate(t *testing.T) {
	rec, _ := testReconciler(t,
		localApiGatewayConfig("payments-prod", "general-policy",
			cascade.ApiGatewayConfigSection{},
			cascade.ApiGatewayConfigSection{NamingTemplate: "{namespace}-{name}"},
		),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getApiGatewayConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Defaults.NamingTemplate; got != "{namespace}-{name}" {
		t.Fatalf("defaults.namingTemplate = %q, want {namespace}-{name}", got)
	}
	if got := updated.Status.EffectiveConfig.Mandatory.NamingTemplate; got != "" {
		t.Fatalf("mandatory.namingTemplate = %q, want empty", got)
	}
}

// AC-4: globalKropathConfig.mandatory.tags augmented into KropathSection tags cascade.
func TestReconcileAC4GlobalKropathTagsAugmented(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfig("general-policy", v1alpha1.KropathConfigTier{
			Tags: map[string]string{"cost-centre": "infra"},
		}),
		localApiGatewayConfig("payments-prod", "general-policy",
			cascade.ApiGatewayConfigSection{Tags: map[string]string{"service": "api"}},
			cascade.ApiGatewayConfigSection{},
		),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getApiGatewayConfig(t, rec.Client, "payments-prod", "general-policy")
	tags := updated.Status.EffectiveConfig.Mandatory.Tags
	if tags["cost-centre"] != "infra" {
		t.Fatalf("tags[cost-centre] = %q, want infra (from global KropathConfig)", tags["cost-centre"])
	}
	if tags["service"] != "api" {
		t.Fatalf("tags[service] = %q, want api (from local ApiGatewayConfig mandatory)", tags["service"])
	}
}

// AC-5: Provider identity from globalKropathConfig propagates to effCfg.aws.*.
func TestReconcileAC5ProviderIdentityPropagates(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfigWithAWS("general-policy", v1alpha1.ProviderIdentity{AccountID: "123456789012", Region: "ap-southeast-2"}),
		localApiGatewayConfig("payments-prod", "general-policy", cascade.ApiGatewayConfigSection{}, cascade.ApiGatewayConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getApiGatewayConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.AWS.AccountID; got != "123456789012" {
		t.Fatalf("aws.accountId = %q, want 123456789012", got)
	}
	if got := updated.Status.EffectiveConfig.AWS.Region; got != "ap-southeast-2" {
		t.Fatalf("aws.region = %q, want ap-southeast-2", got)
	}
}

// AC-6: disableExecuteApiEndpoint enforced from global ApiGatewayConfig mandatory (level 3).
func TestReconcileAC6DisableExecuteApiEndpointLevel3(t *testing.T) {
	rec, _ := testReconciler(t,
		globalApiGatewayConfig("general-policy",
			cascade.ApiGatewayConfigSection{DisableExecuteApiEndpoint: true},
			cascade.ApiGatewayConfigSection{},
		),
		localApiGatewayConfig("payments-prod", "general-policy", cascade.ApiGatewayConfigSection{}, cascade.ApiGatewayConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getApiGatewayConfig(t, rec.Client, "payments-prod", "general-policy")
	if !updated.Status.EffectiveConfig.Mandatory.DisableExecuteApiEndpoint {
		t.Fatalf("mandatory.disableExecuteApiEndpoint = false, want true")
	}
}

func TestRequestsForKropathConfigChangeGlobal(t *testing.T) {
	rec, _ := testReconciler(t,
		localApiGatewayConfig("payments-prod", "general-policy", cascade.ApiGatewayConfigSection{}, cascade.ApiGatewayConfigSection{}),
		localApiGatewayConfig("sandbox", "general-policy", cascade.ApiGatewayConfigSection{}, cascade.ApiGatewayConfigSection{}),
		localApiGatewayConfig("payments-prod", "other-policy", cascade.ApiGatewayConfigSection{}, cascade.ApiGatewayConfigSection{}),
	)

	got := rec.requestsForKropathConfigChange(context.Background(), &v1alpha1.KropathConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "general-policy", Namespace: kroSystemNamespace},
	})

	// Global change should enqueue all non-kro-system configs with any name (list all).
	// Filter: only configs whose name matches cfg.Name.
	if len(got) != 2 {
		t.Fatalf("requests len = %d, want 2 (%#v)", len(got), got)
	}
}

func TestRequestsForApiGatewayConfigChangeNonGlobalIgnored(t *testing.T) {
	rec, _ := testReconciler(t,
		localApiGatewayConfig("payments-prod", "general-policy", cascade.ApiGatewayConfigSection{}, cascade.ApiGatewayConfigSection{}),
	)

	got := rec.requestsForApiGatewayConfigChange(context.Background(), &v1alpha1.ApiGatewayConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "general-policy", Namespace: "payments-prod"},
	})

	if len(got) != 0 {
		t.Fatalf("requests len = %d, want 0 (%#v)", len(got), got)
	}
}

// ─── Test fixtures ──────────────────────────────────────────────────────────

func testReconciler(t *testing.T, objs ...runtime.Object) (*Reconciler, *v1alpha1.ApiGatewayConfig) {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}
	builder := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&v1alpha1.ApiGatewayConfig{})
	for _, obj := range objs {
		builder = builder.WithRuntimeObjects(obj)
	}
	cl := builder.Build()
	cfg := &v1alpha1.ApiGatewayConfig{}
	if err := cl.Get(context.Background(), client.ObjectKey{Namespace: "payments-prod", Name: "general-policy"}, cfg); err != nil {
		t.Fatalf("seed local ApiGatewayConfig: %v", err)
	}
	return &Reconciler{Client: cl, Log: logr.Discard(), Scheme: scheme}, cfg
}

func globalKropathConfig(name string, tier v1alpha1.KropathConfigTier) *v1alpha1.KropathConfig {
	return &v1alpha1.KropathConfig{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "KropathConfig"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: kroSystemNamespace,
		},
		Spec: v1alpha1.KropathConfigSpec{
			Mandatory: tier,
		},
	}
}

func globalKropathConfigWithAWS(name string, aws v1alpha1.ProviderIdentity) *v1alpha1.KropathConfig {
	cfg := globalKropathConfig(name, v1alpha1.KropathConfigTier{})
	cfg.Spec.AWS = aws
	return cfg
}

func globalApiGatewayConfig(name string, mandatory, defaults cascade.ApiGatewayConfigSection) *v1alpha1.ApiGatewayConfig {
	return &v1alpha1.ApiGatewayConfig{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "ApiGatewayConfig"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: kroSystemNamespace,
		},
		Spec: v1alpha1.ApiGatewayConfigSpec{
			Mandatory: mandatory,
			Defaults:  defaults,
		},
	}
}

func localApiGatewayConfig(namespace, name string, mandatory, defaults cascade.ApiGatewayConfigSection) *v1alpha1.ApiGatewayConfig {
	cfg := globalApiGatewayConfig(name, mandatory, defaults)
	cfg.Namespace = namespace
	return cfg
}

func getApiGatewayConfig(t *testing.T, c client.Client, namespace, name string) *v1alpha1.ApiGatewayConfig {
	t.Helper()
	cfg := &v1alpha1.ApiGatewayConfig{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("ApiGatewayConfig"))
	if err := c.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: name}, cfg); err != nil {
		t.Fatalf("get ApiGatewayConfig %s/%s: %v", namespace, name, err)
	}
	return cfg
}

func req(namespace, name string) ctrl.Request {
	return ctrl.Request{NamespacedName: client.ObjectKey{Namespace: namespace, Name: name}}
}
