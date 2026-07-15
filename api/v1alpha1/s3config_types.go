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

package v1alpha1

import (
	"github.com/kropath/kropath-controller/internal/cascade"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type S3Config struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   S3ConfigSpec   `json:"spec,omitempty"`
	Status S3ConfigStatus `json:"status,omitempty"`
}

type S3ConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []S3Config `json:"items"`
}

type S3ConfigSpec struct {
	Mandatory cascade.S3Section `json:"mandatory,omitempty"`
	Defaults  cascade.S3Section `json:"defaults,omitempty"`
}

type S3ConfigStatus struct {
	EffectiveConfig    cascade.EffectiveS3Config `json:"effectiveConfig,omitempty"`
	ObservedGeneration int64                     `json:"observedGeneration,omitempty"`
	SyncedTimestamp    string                    `json:"syncedTimestamp,omitempty"`
}

func (in *S3Config) DeepCopyInto(out *S3Config) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

func (in *S3Config) DeepCopy() *S3Config {
	if in == nil {
		return nil
	}
	out := new(S3Config)
	in.DeepCopyInto(out)
	return out
}

func (in *S3Config) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *S3ConfigList) DeepCopyInto(out *S3ConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]S3Config, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *S3ConfigList) DeepCopy() *S3ConfigList {
	if in == nil {
		return nil
	}
	out := new(S3ConfigList)
	in.DeepCopyInto(out)
	return out
}

func (in *S3ConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
