// Copyright 2026 The kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	AWSPolicyDocumentKind = "AWSPolicyDocument"

	ConditionReady          = "Ready"
	ConditionSourceNotReady = "SourceNotReady"
	ConditionSidConflict    = "SidConflict"
)

type AWSPolicyDocumentSpec struct {
	Sources      []PolicyDocumentSource `json:"sources,omitempty"`
	Statements   []PolicyStatement      `json:"statements,omitempty"`
	DocumentJSON string                 `json:"documentJSON,omitempty"`
}

type PolicyDocumentSource struct {
	Name string `json:"name"`
}

type PolicyStatement struct {
	Sid        string            `json:"sid,omitempty"`
	Effect     string            `json:"effect"`
	Principals []PolicyPrincipal `json:"principals,omitempty"`
	Actions    []string          `json:"actions"`
	Resources  []PolicyResource  `json:"resources,omitempty"`
	Conditions []PolicyCondition `json:"conditions,omitempty"`
}

type PolicyPrincipal struct {
	Type string     `json:"type"`
	ARN  string     `json:"arn,omitempty"`
	Ref  *PolicyRef `json:"ref,omitempty"`
}

type PolicyResource struct {
	ARN string     `json:"arn,omitempty"`
	Ref *PolicyRef `json:"ref,omitempty"`
}

type PolicyRef struct {
	Kind  string `json:"kind"`
	Name  string `json:"name"`
	Field string `json:"field,omitempty"`
}

type PolicyCondition struct {
	Operator string   `json:"operator"`
	Key      string   `json:"key"`
	Values   []string `json:"values"`
}

type AWSPolicyDocumentStatus struct {
	ResolvedDocumentJSON string             `json:"resolvedDocumentJSON,omitempty"`
	StatementCount       int32              `json:"statementCount,omitempty"`
	SourceCount          int32              `json:"sourceCount,omitempty"`
	Conditions           []metav1.Condition `json:"conditions,omitempty"`
	ObservedGeneration   int64              `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type AWSPolicyDocument struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSPolicyDocumentSpec   `json:"spec,omitempty"`
	Status AWSPolicyDocumentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type AWSPolicyDocumentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSPolicyDocument `json:"items"`
}

func (in *AWSPolicyDocument) DeepCopyObject() runtime.Object {
	if in == nil {
		return nil
	}
	out := new(AWSPolicyDocument)
	copyThroughJSON(in, out)
	return out
}

func (in *AWSPolicyDocumentList) DeepCopyObject() runtime.Object {
	if in == nil {
		return nil
	}
	out := new(AWSPolicyDocumentList)
	copyThroughJSON(in, out)
	return out
}

func copyThroughJSON(src, dst any) {
	b, err := json.Marshal(src)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(b, dst); err != nil {
		panic(err)
	}
}
