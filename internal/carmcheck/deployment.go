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

package carmcheck

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	enableCARMFlag     = "--enable-carm"
	watchNamespaceFlag = "--watch-namespace"
)

// ackIgnoredNamespaces mirrors ACK's fixed, non-configurable ignore list —
// verified against aws-controllers-k8s/runtime's
// pkg/runtime/service_controller.go, where BindControllerManager constructs
// its namespace cache with
// Ignored: []string{NamespaceKubeSystem, NamespaceKubePublic, NamespaceKubeNodeLease}.
// Unlike --enable-carm or --watch-namespace, this list is not read from a
// flag, env var, or ConfigMap, so a match here is a certain violation, never
// a SeverityUnknown guess.
var ackIgnoredNamespaces = []string{"kube-system", "kube-public", "kube-node-lease"}

// ackDeployments lists every Deployment in namespace. It does not try to
// filter for "the ACK ones" by label or image, because ACK controllers carry
// no standard discovery label across services — any Deployment in the
// dedicated ACK namespace is assumed relevant.
func ackDeployments(ctx context.Context, c client.Client, namespace string) ([]appsv1.Deployment, error) {
	var list appsv1.DeploymentList
	if err := c.List(ctx, &list, client.InNamespace(namespace)); err != nil {
		return nil, err
	}
	return list.Items, nil
}

// deploymentFlag returns the value of a "--flag value", "--flag=value", or
// bare boolean "--flag" command-line argument from any container in the
// Deployment's pod template. ok is false if no container declares the flag
// at all — the caller must treat that as "cannot verify", not "unset", since
// ACK also accepts this configuration via environment variable or Helm
// feature gate.
func deploymentFlag(d appsv1.Deployment, flag string) (value string, ok bool) {
	for _, container := range d.Spec.Template.Spec.Containers {
		for i, arg := range container.Args {
			switch {
			case arg == flag:
				if i+1 < len(container.Args) && !strings.HasPrefix(container.Args[i+1], "--") {
					return container.Args[i+1], true
				}
				return "true", true
			case strings.HasPrefix(arg, flag+"="):
				return strings.TrimPrefix(arg, flag+"="), true
			}
		}
	}
	return "", false
}

// checkEnableCARM implements ADR-015 §5.8.4 precondition 2: --enable-carm=true.
func checkEnableCARM(ctx context.Context, c client.Client, ackNamespace string) []Finding {
	deployments, err := ackDeployments(ctx, c, ackNamespace)
	if err != nil {
		return []Finding{{
			Check:    "enable-carm",
			Severity: SeverityUnknown,
			Message:  fmt.Sprintf("could not list Deployments in namespace %q: %v", ackNamespace, err),
		}}
	}
	if len(deployments) == 0 {
		return []Finding{{
			Check:    "enable-carm",
			Severity: SeverityUnknown,
			Message:  fmt.Sprintf("no Deployments found in namespace %q; point --ack-namespace at the namespace hosting the ACK controllers", ackNamespace),
		}}
	}

	var findings []Finding
	for _, d := range deployments {
		val, ok := deploymentFlag(d, enableCARMFlag)
		switch {
		case !ok:
			findings = append(findings, Finding{
				Check:    "enable-carm",
				Severity: SeverityUnknown,
				Message: fmt.Sprintf(
					"Deployment %s/%s: %s not found in container args; it may be set via an environment variable or a Helm feature gate this checker cannot observe (ADR-015 §5.8.4 precondition 2)",
					d.Namespace, d.Name, enableCARMFlag,
				),
			})
		case val == "false":
			findings = append(findings, Finding{
				Check:    "enable-carm",
				Severity: SeverityBlocking,
				Message: fmt.Sprintf(
					"Deployment %s/%s: %s=false; with CARM off, ACK reads the namespace's owner-account-id annotation but assumes no role and uses its own controller credentials instead (ADR-015 §5.8.4 precondition 2)",
					d.Namespace, d.Name, enableCARMFlag,
				),
			})
		case val == "true":
			findings = append(findings, Finding{
				Check:    "enable-carm",
				Severity: SeverityOK,
				Message:  fmt.Sprintf("Deployment %s/%s: %s=true", d.Namespace, d.Name, enableCARMFlag),
			})
		default:
			findings = append(findings, Finding{
				Check:    "enable-carm",
				Severity: SeverityUnknown,
				Message:  fmt.Sprintf("Deployment %s/%s: %s has unrecognized value %q", d.Namespace, d.Name, enableCARMFlag, val),
			})
		}
	}
	return findings
}

// checkWatchNamespaceScope implements ADR-015 §5.8.4 precondition 1: every
// kropath-managed namespace must be inside ACK's --watch-namespace scope.
// approvedNamespace() in ACK's runtime gates the entire namespace annotation
// cache on this — an out-of-scope namespace's owner-account-id and
// default-region annotations are read by kropath and ignored by ACK.
func checkWatchNamespaceScope(ctx context.Context, c client.Client, ackNamespace string, managed []corev1.Namespace) []Finding {
	deployments, err := ackDeployments(ctx, c, ackNamespace)
	if err != nil || len(deployments) == 0 {
		return []Finding{{
			Check:    "watch-namespace-scope",
			Severity: SeverityUnknown,
			Message:  fmt.Sprintf("could not inspect %s: no ACK Deployments found in namespace %q", watchNamespaceFlag, ackNamespace),
		}}
	}

	var findings []Finding
	for _, d := range deployments {
		val, ok := deploymentFlag(d, watchNamespaceFlag)
		if !ok || strings.TrimSpace(val) == "" {
			findings = append(findings, Finding{
				Check:    "watch-namespace-scope",
				Severity: SeverityOK,
				Message: fmt.Sprintf(
					"Deployment %s/%s: %s is unset (cluster-wide) — every kropath-managed namespace is in scope",
					d.Namespace, d.Name, watchNamespaceFlag,
				),
			})
			continue
		}

		watched := make(map[string]bool)
		for _, ns := range strings.Split(val, ",") {
			if ns = strings.TrimSpace(ns); ns != "" {
				watched[ns] = true
			}
		}

		var missing []string
		for _, ns := range managed {
			if !watched[ns.Name] {
				missing = append(missing, ns.Name)
			}
		}

		if len(missing) > 0 {
			sort.Strings(missing)
			findings = append(findings, Finding{
				Check:    "watch-namespace-scope",
				Severity: SeverityBlocking,
				Message: fmt.Sprintf(
					"Deployment %s/%s: %s=%q does not include kropath-managed namespace(s) %s; ACK ignores their annotations entirely (ADR-015 §5.8.4 precondition 1)",
					d.Namespace, d.Name, watchNamespaceFlag, val, strings.Join(missing, ", "),
				),
			})
		} else {
			findings = append(findings, Finding{
				Check:    "watch-namespace-scope",
				Severity: SeverityOK,
				Message: fmt.Sprintf(
					"Deployment %s/%s: %s covers all %d kropath-managed namespace(s)",
					d.Namespace, d.Name, watchNamespaceFlag, len(managed),
				),
			})
		}
	}
	return findings
}

// checkNamespaceIgnoreList implements the second half of ADR-015 §5.8.4
// precondition 1: a kropath-managed namespace must not be one of ACK's fixed
// ignore-list namespaces. approvedNamespace() excludes these three
// unconditionally, so a match makes ACK ignore the namespace's annotations
// no matter how --watch-namespace is configured — this is why the check runs
// independently of checkWatchNamespaceScope and of any ACK Deployment lookup.
func checkNamespaceIgnoreList(managed []corev1.Namespace) []Finding {
	var findings []Finding
	for _, ns := range managed {
		if slices.Contains(ackIgnoredNamespaces, ns.Name) {
			findings = append(findings, Finding{
				Check:     "watch-namespace-scope",
				Severity:  SeverityBlocking,
				Namespace: ns.Name,
				Message: fmt.Sprintf(
					"namespace %q is in ACK's fixed ignore list (%s); approvedNamespace() excludes it regardless of %s (ADR-015 §5.8.4 precondition 1)",
					ns.Name, strings.Join(ackIgnoredNamespaces, ", "), watchNamespaceFlag,
				),
			})
		}
	}
	if len(findings) == 0 {
		findings = append(findings, Finding{
			Check:    "watch-namespace-scope",
			Severity: SeverityOK,
			Message:  fmt.Sprintf("no kropath-managed namespace matches ACK's fixed ignore list (%s)", strings.Join(ackIgnoredNamespaces, ", ")),
		})
	}
	return findings
}
