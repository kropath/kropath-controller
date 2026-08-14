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

package util_test

import (
	"context"
	"testing"

	"github.com/kropath/kropath-controller/internal/reconciler/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func testClient(t *testing.T, objs ...runtime.Object) *fake.ClientBuilder {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme: %v", err)
	}
	return fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objs...)
}

func TestResolveGlobalNamespaceFromAnnotation(t *testing.T) {
	ns := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "payments-prod",
			Annotations: map[string]string{util.GlobalConfigNamespaceAnnotation: "platform-config"},
		},
	}
	c := testClient(t, ns).Build()

	got := util.ResolveGlobalNamespace(context.Background(), c, "payments-prod")
	if got != "platform-config" {
		t.Errorf("ResolveGlobalNamespace = %q, want %q", got, "platform-config")
	}
}

func TestResolveGlobalNamespaceDefaultsToKroSystem(t *testing.T) {
	ns := &corev1.Namespace{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{Name: "payments-prod"},
	}
	c := testClient(t, ns).Build()

	got := util.ResolveGlobalNamespace(context.Background(), c, "payments-prod")
	if got != util.DefaultGlobalNamespace {
		t.Errorf("ResolveGlobalNamespace = %q, want %q", got, util.DefaultGlobalNamespace)
	}
}

func TestResolveGlobalNamespaceMissingNamespaceFallsBack(t *testing.T) {
	c := testClient(t).Build()

	got := util.ResolveGlobalNamespace(context.Background(), c, "nonexistent")
	if got != util.DefaultGlobalNamespace {
		t.Errorf("ResolveGlobalNamespace = %q, want %q (fallback)", got, util.DefaultGlobalNamespace)
	}
}

func TestResolveGlobalNamespaceEmptyAnnotationFallsBack(t *testing.T) {
	ns := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "payments-prod",
			Annotations: map[string]string{util.GlobalConfigNamespaceAnnotation: ""},
		},
	}
	c := testClient(t, ns).Build()

	got := util.ResolveGlobalNamespace(context.Background(), c, "payments-prod")
	if got != util.DefaultGlobalNamespace {
		t.Errorf("ResolveGlobalNamespace = %q, want %q (empty annotation falls back)", got, util.DefaultGlobalNamespace)
	}
}
