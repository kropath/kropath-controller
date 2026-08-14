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

// Package util provides shared helpers for reconcilers.
package util

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// GlobalConfigNamespaceAnnotation is the annotation key on a Namespace that
	// designates which namespace holds the global-tier KropathConfig for resources
	// in that namespace. If absent, DefaultGlobalNamespace is used.
	GlobalConfigNamespaceAnnotation = "aws.kropath.run/global-config-namespace"

	// DefaultGlobalNamespace is the fallback global config namespace when
	// the annotation is absent or the namespace object cannot be read.
	DefaultGlobalNamespace = "kro-system"
)

// ResolveGlobalNamespace returns the effective global-tier config namespace for
// resources in the given resourceNamespace. It reads the
// aws.kropath.run/global-config-namespace annotation from the namespace object.
// If the annotation is absent or the namespace cannot be fetched, it returns
// DefaultGlobalNamespace ("kro-system").
func ResolveGlobalNamespace(ctx context.Context, c client.Client, resourceNamespace string) string {
	var ns corev1.Namespace
	if err := c.Get(ctx, types.NamespacedName{Name: resourceNamespace}, &ns); err != nil {
		return DefaultGlobalNamespace
	}
	if ann := ns.Annotations[GlobalConfigNamespaceAnnotation]; ann != "" {
		return ann
	}
	return DefaultGlobalNamespace
}
