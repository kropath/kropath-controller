// Copyright 2026 kropath contributors.
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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type AWSPolicyDocument struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSPolicyDocumentSpec   `json:"spec,omitempty"`
	Status AWSPolicyDocumentStatus `json:"status,omitempty"`
}

type AWSPolicyDocumentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []AWSPolicyDocument `json:"items"`
}

type AWSPolicyDocumentSpec struct {
	Sources      []PolicySource    `json:"sources,omitempty"`
	Statements   []PolicyStatement `json:"statements,omitempty"`
	DocumentJSON string            `json:"documentJSON,omitempty"`
}

type AWSPolicyDocumentStatus struct {
	ResolvedDocumentJSON string             `json:"resolvedDocumentJSON,omitempty"`
	StatementCount       int                `json:"statementCount,omitempty"`
	SourceCount          int                `json:"sourceCount,omitempty"`
	Conditions           []metav1.Condition `json:"conditions,omitempty"`
	ObservedGeneration   int64              `json:"observedGeneration,omitempty"`
}

type PolicySource struct {
	Name string `json:"name"`
}

type PolicyStatement struct {
	Sid        string            `json:"sid,omitempty"`
	Effect     string            `json:"effect"`
	Principals []PolicyPrincipal `json:"principals,omitempty"`
	Actions    []string          `json:"actions,omitempty"`
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

type PolicyCondition struct {
	Operator string   `json:"operator"`
	Key      string   `json:"key"`
	Values   []string `json:"values,omitempty"`
}

type PolicyRef struct {
	Kind  string `json:"kind"`
	Name  string `json:"name"`
	Field string `json:"field,omitempty"`
}

func (in *AWSPolicyDocument) DeepCopyInto(out *AWSPolicyDocument) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

func (in *AWSPolicyDocument) DeepCopy() *AWSPolicyDocument {
	if in == nil {
		return nil
	}
	out := new(AWSPolicyDocument)
	in.DeepCopyInto(out)
	return out
}

func (in *AWSPolicyDocument) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *AWSPolicyDocumentList) DeepCopyInto(out *AWSPolicyDocumentList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]AWSPolicyDocument, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *AWSPolicyDocumentList) DeepCopy() *AWSPolicyDocumentList {
	if in == nil {
		return nil
	}
	out := new(AWSPolicyDocumentList)
	in.DeepCopyInto(out)
	return out
}

func (in *AWSPolicyDocumentList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *AWSPolicyDocumentSpec) DeepCopyInto(out *AWSPolicyDocumentSpec) {
	*out = *in
	if in.Sources != nil {
		out.Sources = make([]PolicySource, len(in.Sources))
		copy(out.Sources, in.Sources)
	}
	if in.Statements != nil {
		out.Statements = make([]PolicyStatement, len(in.Statements))
		for i := range in.Statements {
			in.Statements[i].DeepCopyInto(&out.Statements[i])
		}
	}
}

func (in *AWSPolicyDocumentStatus) DeepCopyInto(out *AWSPolicyDocumentStatus) {
	*out = *in
	if in.Conditions != nil {
		out.Conditions = make([]metav1.Condition, len(in.Conditions))
		copy(out.Conditions, in.Conditions)
	}
}

func (in *PolicyStatement) DeepCopyInto(out *PolicyStatement) {
	*out = *in
	if in.Principals != nil {
		out.Principals = make([]PolicyPrincipal, len(in.Principals))
		copy(out.Principals, in.Principals)
	}
	if in.Actions != nil {
		out.Actions = append([]string(nil), in.Actions...)
	}
	if in.Resources != nil {
		out.Resources = make([]PolicyResource, len(in.Resources))
		copy(out.Resources, in.Resources)
	}
	if in.Conditions != nil {
		out.Conditions = make([]PolicyCondition, len(in.Conditions))
		copy(out.Conditions, in.Conditions)
	}
}
