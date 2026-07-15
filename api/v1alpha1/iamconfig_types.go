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

type ProviderIdentity struct {
	AccountID string `json:"accountId,omitempty"`
	Region    string `json:"region,omitempty"`
}

type KropathConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec KropathConfigSpec `json:"spec,omitempty"`
}

type KropathConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []KropathConfig `json:"items"`
}

type KropathConfigSpec struct {
	Mandatory KropathConfigTier `json:"mandatory,omitempty"`
	Defaults  KropathConfigTier `json:"defaults,omitempty"`
	AWS       ProviderIdentity  `json:"aws,omitempty"`
}

type KropathConfigTier struct {
	IAM  cascade.IAMSection        `json:"iam,omitempty"`
	S3   cascade.S3Section         `json:"s3,omitempty"`
	KMS  cascade.KMSKropathSection `json:"kms,omitempty"`
	Tags map[string]string         `json:"tags,omitempty"`
}

type IAMConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IAMConfigSpec   `json:"spec,omitempty"`
	Status IAMConfigStatus `json:"status,omitempty"`
}

type IAMConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []IAMConfig `json:"items"`
}

type IAMConfigSpec struct {
	Mandatory cascade.IAMSection `json:"mandatory,omitempty"`
	Defaults  cascade.IAMSection `json:"defaults,omitempty"`
}

type IAMConfigStatus struct {
	EffectiveConfig    EffectiveIAMConfig `json:"effectiveConfig,omitempty"`
	ObservedGeneration int64                 `json:"observedGeneration,omitempty"`
	SyncedTimestamp    string                `json:"syncedTimestamp,omitempty"`
}

type EffectiveIAMConfig struct {
	AWS       ProviderIdentity `json:"aws,omitempty"`
	Mandatory cascade.IAMSection  `json:"mandatory,omitempty"`
	Defaults  cascade.IAMSection  `json:"defaults,omitempty"`
}

func (in *KropathConfig) DeepCopyInto(out *KropathConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	if in.Spec.Mandatory.KMS.AllowedKeySpecs != nil {
		out.Spec.Mandatory.KMS.AllowedKeySpecs = make([]string, len(in.Spec.Mandatory.KMS.AllowedKeySpecs))
		copy(out.Spec.Mandatory.KMS.AllowedKeySpecs, in.Spec.Mandatory.KMS.AllowedKeySpecs)
	}
	if in.Spec.Mandatory.KMS.Tags != nil {
		out.Spec.Mandatory.KMS.Tags = make(map[string]string, len(in.Spec.Mandatory.KMS.Tags))
		for k, v := range in.Spec.Mandatory.KMS.Tags {
			out.Spec.Mandatory.KMS.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.Tags != nil {
		out.Spec.Mandatory.Tags = make(map[string]string, len(in.Spec.Mandatory.Tags))
		for k, v := range in.Spec.Mandatory.Tags {
			out.Spec.Mandatory.Tags[k] = v
		}
	}
	if in.Spec.Defaults.KMS.AllowedKeySpecs != nil {
		out.Spec.Defaults.KMS.AllowedKeySpecs = make([]string, len(in.Spec.Defaults.KMS.AllowedKeySpecs))
		copy(out.Spec.Defaults.KMS.AllowedKeySpecs, in.Spec.Defaults.KMS.AllowedKeySpecs)
	}
	if in.Spec.Defaults.KMS.Tags != nil {
		out.Spec.Defaults.KMS.Tags = make(map[string]string, len(in.Spec.Defaults.KMS.Tags))
		for k, v := range in.Spec.Defaults.KMS.Tags {
			out.Spec.Defaults.KMS.Tags[k] = v
		}
	}
	if in.Spec.Defaults.Tags != nil {
		out.Spec.Defaults.Tags = make(map[string]string, len(in.Spec.Defaults.Tags))
		for k, v := range in.Spec.Defaults.Tags {
			out.Spec.Defaults.Tags[k] = v
		}
	}
}

func (in *KropathConfig) DeepCopy() *KropathConfig {
	if in == nil {
		return nil
	}
	out := new(KropathConfig)
	in.DeepCopyInto(out)
	return out
}

func (in *KropathConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *KropathConfigList) DeepCopyInto(out *KropathConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]KropathConfig, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *KropathConfigList) DeepCopy() *KropathConfigList {
	if in == nil {
		return nil
	}
	out := new(KropathConfigList)
	in.DeepCopyInto(out)
	return out
}

func (in *KropathConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *IAMConfig) DeepCopyInto(out *IAMConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

func (in *IAMConfig) DeepCopy() *IAMConfig {
	if in == nil {
		return nil
	}
	out := new(IAMConfig)
	in.DeepCopyInto(out)
	return out
}

func (in *IAMConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *IAMConfigList) DeepCopyInto(out *IAMConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]IAMConfig, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *IAMConfigList) DeepCopy() *IAMConfigList {
	if in == nil {
		return nil
	}
	out := new(IAMConfigList)
	in.DeepCopyInto(out)
	return out
}

func (in *IAMConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
