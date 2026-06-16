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

type AWSS3Config struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSS3ConfigSpec   `json:"spec,omitempty"`
	Status AWSS3ConfigStatus `json:"status,omitempty"`
}

type AWSS3ConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []AWSS3Config `json:"items"`
}

type AWSS3ConfigSpec struct {
	Mandatory cascade.S3Section `json:"mandatory,omitempty"`
	Defaults  cascade.S3Section `json:"defaults,omitempty"`
}

type AWSS3ConfigStatus struct {
	EffectiveConfig    AWSEffectiveS3Config `json:"effectiveConfig,omitempty"`
	ObservedGeneration int64                `json:"observedGeneration,omitempty"`
	SyncedTimestamp    string               `json:"syncedTimestamp,omitempty"`
}

type AWSEffectiveS3Config struct {
	AWS       AWSProviderIdentity `json:"aws,omitempty"`
	Mandatory cascade.S3Section   `json:"mandatory,omitempty"`
	Defaults  cascade.S3Section   `json:"defaults,omitempty"`
}

func (in *AWSS3Config) DeepCopyInto(out *AWSS3Config) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

func (in *AWSS3Config) DeepCopy() *AWSS3Config {
	if in == nil {
		return nil
	}
	out := new(AWSS3Config)
	in.DeepCopyInto(out)
	return out
}

func (in *AWSS3Config) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *AWSS3ConfigList) DeepCopyInto(out *AWSS3ConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]AWSS3Config, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *AWSS3ConfigList) DeepCopy() *AWSS3ConfigList {
	if in == nil {
		return nil
	}
	out := new(AWSS3ConfigList)
	in.DeepCopyInto(out)
	return out
}

func (in *AWSS3ConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
