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

	"github.com/kropath/kropath-controller/api/v1alpha1"
	"github.com/kropath/kropath-controller/internal/cascade"
	"github.com/kropath/kropath-controller/internal/reconciler/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func profileTestClient(t *testing.T, objs ...runtime.Object) *fake.ClientBuilder {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme: %v", err)
	}
	return fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objs...)
}

func s3Config(namespace, name string, blockPublicAccess bool) *v1alpha1.S3Config {
	return &v1alpha1.S3Config{
		TypeMeta:   metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "S3Config"},
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Spec: v1alpha1.S3ConfigSpec{
			Mandatory: cascade.S3ConfigSection{BlockPublicAccess: blockPublicAccess},
		},
	}
}

func TestLoadConfigWithFallthrough_RequestedProfileExists(t *testing.T) {
	c := profileTestClient(t, s3Config("tenant", "pci", true)).Build()

	obj, found, viaFallthrough, err := util.LoadConfigWithFallthrough[v1alpha1.S3Config](
		context.Background(), c, v1alpha1.GroupVersion.WithKind("S3Config"), "tenant", "pci", util.DefaultConfigProfile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found || viaFallthrough {
		t.Fatalf("found=%v viaFallthrough=%v, want found=true viaFallthrough=false", found, viaFallthrough)
	}
	if !obj.Spec.Mandatory.BlockPublicAccess {
		t.Fatalf("expected the requested profile's own config, not the fallthrough target")
	}
}

func TestLoadConfigWithFallthrough_FallsThroughWhenProfileMissing(t *testing.T) {
	c := profileTestClient(t, s3Config("tenant", util.DefaultConfigProfile, true)).Build()

	obj, found, viaFallthrough, err := util.LoadConfigWithFallthrough[v1alpha1.S3Config](
		context.Background(), c, v1alpha1.GroupVersion.WithKind("S3Config"), "tenant", "pci", util.DefaultConfigProfile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found || !viaFallthrough {
		t.Fatalf("found=%v viaFallthrough=%v, want found=true viaFallthrough=true", found, viaFallthrough)
	}
	if !obj.Spec.Mandatory.BlockPublicAccess {
		t.Fatalf("expected the fallthrough target's config to be merged")
	}
}

func TestLoadConfigWithFallthrough_NeitherExists(t *testing.T) {
	c := profileTestClient(t).Build()

	obj, found, viaFallthrough, err := util.LoadConfigWithFallthrough[v1alpha1.S3Config](
		context.Background(), c, v1alpha1.GroupVersion.WithKind("S3Config"), "tenant", "pci", util.DefaultConfigProfile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found || viaFallthrough {
		t.Fatalf("found=%v viaFallthrough=%v, want found=false viaFallthrough=false", found, viaFallthrough)
	}
	if obj == nil {
		t.Fatalf("expected a non-nil zero-value object so callers can merge an empty tier")
	}
}

func TestLoadConfigWithFallthrough_RequestedProfileIsAlreadyDefault(t *testing.T) {
	// configRef == "general-policy" and it does not exist: no second Get should be attempted,
	// and the outcome must still be found=false (not a false fallthrough success).
	c := profileTestClient(t).Build()

	obj, found, viaFallthrough, err := util.LoadConfigWithFallthrough[v1alpha1.S3Config](
		context.Background(), c, v1alpha1.GroupVersion.WithKind("S3Config"), "tenant", util.DefaultConfigProfile, util.DefaultConfigProfile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found || viaFallthrough {
		t.Fatalf("found=%v viaFallthrough=%v, want found=false viaFallthrough=false", found, viaFallthrough)
	}
	if obj == nil {
		t.Fatalf("expected a non-nil zero-value object")
	}
}

func TestConfigProfileResolvedCondition(t *testing.T) {
	now := metav1.Now()

	cases := []struct {
		name           string
		found          bool
		viaFallthrough bool
		wantStatus     metav1.ConditionStatus
		wantReason     string
	}{
		{"direct hit", true, false, metav1.ConditionTrue, "ProfileFound"},
		{"fallthrough hit", true, true, metav1.ConditionTrue, "ProfileFallthrough"},
		{"unresolved", false, false, metav1.ConditionFalse, "ProfileUnresolved"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cond := util.ConfigProfileResolvedCondition("pci", tc.found, tc.viaFallthrough, 3, now)
			if cond.Type != util.ConfigProfileResolvedConditionType {
				t.Fatalf("Type = %q, want %q", cond.Type, util.ConfigProfileResolvedConditionType)
			}
			if cond.Status != tc.wantStatus {
				t.Fatalf("Status = %q, want %q", cond.Status, tc.wantStatus)
			}
			if cond.Reason != tc.wantReason {
				t.Fatalf("Reason = %q, want %q", cond.Reason, tc.wantReason)
			}
			if cond.ObservedGeneration != 3 {
				t.Fatalf("ObservedGeneration = %d, want 3", cond.ObservedGeneration)
			}
		})
	}
}
