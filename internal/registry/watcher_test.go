// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ---- crdServable tests -----------------------------------------------------

type crdServableCase struct {
	name      string
	group     string
	kind      string
	versions  []map[string]interface{}
	conditions []map[string]interface{}
	wantGVK   schema.GroupVersionKind
	wantOK    bool
}

func buildCRD(group, kind string, versions []map[string]interface{}, conditions []map[string]interface{}) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apiextensions.k8s.io",
		Version: "v1",
		Kind:    "CustomResourceDefinition",
	})
	u.Object["spec"] = map[string]interface{}{
		"group": group,
		"names": map[string]interface{}{"kind": kind},
		"versions": func() []interface{} {
			out := make([]interface{}, len(versions))
			for i, v := range versions {
				out[i] = v
			}
			return out
		}(),
	}
	condList := make([]interface{}, len(conditions))
	for i, c := range conditions {
		condList[i] = c
	}
	u.Object["status"] = map[string]interface{}{"conditions": condList}
	return u
}

func TestCRDServable(t *testing.T) {
	cases := []crdServableCase{
		{
			name:  "Established+v1alpha1 served",
			group: "aws.kropath.run", kind: "ELBConfig",
			versions:   []map[string]interface{}{{"name": "v1alpha1", "served": true, "storage": true}},
			conditions: []map[string]interface{}{{"type": "Established", "status": "True"}},
			wantGVK:    schema.GroupVersionKind{Group: "aws.kropath.run", Version: "v1alpha1", Kind: "ELBConfig"},
			wantOK:     true,
		},
		{
			name:  "Not Established",
			group: "aws.kropath.run", kind: "ELBConfig",
			versions:   []map[string]interface{}{{"name": "v1alpha1", "served": true}},
			conditions: []map[string]interface{}{{"type": "Established", "status": "False"}},
			wantOK:     false,
		},
		{
			name:  "No Established condition",
			group: "aws.kropath.run", kind: "ELBConfig",
			versions:   []map[string]interface{}{{"name": "v1alpha1", "served": true}},
			conditions: []map[string]interface{}{},
			wantOK:     false,
		},
		{
			name:  "v1alpha1 not served",
			group: "aws.kropath.run", kind: "ELBConfig",
			versions:   []map[string]interface{}{{"name": "v1alpha1", "served": false}},
			conditions: []map[string]interface{}{{"type": "Established", "status": "True"}},
			wantOK:     false,
		},
		{
			name:  "Different version name",
			group: "aws.kropath.run", kind: "ELBConfig",
			versions:   []map[string]interface{}{{"name": "v1beta1", "served": true}},
			conditions: []map[string]interface{}{{"type": "Established", "status": "True"}},
			wantOK:     false,
		},
		{
			name:  "Missing group",
			group: "", kind: "ELBConfig",
			versions:   []map[string]interface{}{{"name": "v1alpha1", "served": true}},
			conditions: []map[string]interface{}{{"type": "Established", "status": "True"}},
			wantOK:     false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			u := buildCRD(tc.group, tc.kind, tc.versions, tc.conditions)
			gotGVK, gotOK := crdServable(u)
			if gotOK != tc.wantOK {
				t.Errorf("crdServable ok: got %v, want %v", gotOK, tc.wantOK)
			}
			if gotOK && gotGVK != tc.wantGVK {
				t.Errorf("crdServable GVK: got %v, want %v", gotGVK, tc.wantGVK)
			}
		})
	}
}

// ---- stripCRDSchema tests --------------------------------------------------

func TestStripCRDSchema_RemovesSchema(t *testing.T) {
	u := buildCRD("aws.kropath.run", "ELBConfig",
		[]map[string]interface{}{
			{
				"name":   "v1alpha1",
				"served": true,
				"schema": map[string]interface{}{"openAPIV3Schema": map[string]interface{}{"type": "object"}},
			},
		},
		[]map[string]interface{}{{"type": "Established", "status": "True"}},
	)

	out, err := stripCRDSchema(u)
	if err != nil {
		t.Fatalf("stripCRDSchema error: %v", err)
	}
	stripped, ok := out.(*unstructured.Unstructured)
	if !ok {
		t.Fatalf("expected *unstructured.Unstructured, got %T", out)
	}
	versions, _, _ := unstructured.NestedSlice(stripped.Object, "spec", "versions")
	if len(versions) == 0 {
		t.Fatal("versions should not be empty")
	}
	vm := versions[0].(map[string]interface{})
	if _, hasSchema := vm["schema"]; hasSchema {
		t.Error("schema field should have been removed")
	}
	if name, _ := vm["name"].(string); name != "v1alpha1" {
		t.Errorf("name field should be preserved, got %q", name)
	}
}

func TestStripCRDSchema_KeepsStatusConditions(t *testing.T) {
	u := buildCRD("aws.kropath.run", "ELBConfig",
		[]map[string]interface{}{{"name": "v1alpha1", "served": true}},
		[]map[string]interface{}{{"type": "Established", "status": "True"}},
	)
	out, err := stripCRDSchema(u)
	if err != nil {
		t.Fatalf("stripCRDSchema error: %v", err)
	}
	stripped := out.(*unstructured.Unstructured)
	conditions, _, _ := unstructured.NestedSlice(stripped.Object, "status", "conditions")
	if len(conditions) == 0 {
		t.Error("status.conditions should be preserved")
	}
}

func TestStripCRDSchema_NonUnstructured_Passthrough(t *testing.T) {
	// A plain string (not *unstructured.Unstructured) should be returned unchanged.
	input := interface{}("not-an-unstructured")
	out, err := stripCRDSchema(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != input {
		t.Errorf("expected passthrough for non-Unstructured input, got %v", out)
	}
}
