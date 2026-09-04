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

type MWAAConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MWAAConfigSpec   `json:"spec,omitempty"`
	Status MWAAConfigStatus `json:"status,omitempty"`
}

type MWAAConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []MWAAConfig `json:"items"`
}

type MWAAConfigSpec struct {
	Mandatory cascade.MWAAConfigSection `json:"mandatory,omitempty"`
	Defaults  cascade.MWAAConfigSection `json:"defaults,omitempty"`
}

type MWAAConfigStatus struct {
	EffectiveConfig    EffectiveMWAAConfig `json:"effectiveConfig,omitempty"`
	Conditions         []metav1.Condition  `json:"conditions,omitempty"`
	ObservedGeneration int64               `json:"observedGeneration,omitempty"`
	SyncedTimestamp    string              `json:"syncedTimestamp,omitempty"`
}

type EffectiveMWAAConfig struct {
	AWS       ProviderIdentity             `json:"aws,omitempty"`
	Mandatory cascade.EffectiveMWAASection `json:"mandatory,omitempty"`
	Defaults  cascade.EffectiveMWAASection `json:"defaults,omitempty"`
}

func (in *MWAAConfig) DeepCopyInto(out *MWAAConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	// Spec
	out.Spec = in.Spec
	if in.Spec.Mandatory.AirflowConfigurationOptions != nil {
		out.Spec.Mandatory.AirflowConfigurationOptions = make(map[string]string, len(in.Spec.Mandatory.AirflowConfigurationOptions))
		for k, v := range in.Spec.Mandatory.AirflowConfigurationOptions {
			out.Spec.Mandatory.AirflowConfigurationOptions[k] = v
		}
	}
	if in.Spec.Mandatory.Tags != nil {
		out.Spec.Mandatory.Tags = make(map[string]string, len(in.Spec.Mandatory.Tags))
		for k, v := range in.Spec.Mandatory.Tags {
			out.Spec.Mandatory.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.SyncedLabels != nil {
		out.Spec.Mandatory.SyncedLabels = make(map[string]string, len(in.Spec.Mandatory.SyncedLabels))
		for k, v := range in.Spec.Mandatory.SyncedLabels {
			out.Spec.Mandatory.SyncedLabels[k] = v
		}
	}
	if in.Spec.Mandatory.SyncedAnnotations != nil {
		out.Spec.Mandatory.SyncedAnnotations = make(map[string]string, len(in.Spec.Mandatory.SyncedAnnotations))
		for k, v := range in.Spec.Mandatory.SyncedAnnotations {
			out.Spec.Mandatory.SyncedAnnotations[k] = v
		}
	}
	if in.Spec.Defaults.AirflowConfigurationOptions != nil {
		out.Spec.Defaults.AirflowConfigurationOptions = make(map[string]string, len(in.Spec.Defaults.AirflowConfigurationOptions))
		for k, v := range in.Spec.Defaults.AirflowConfigurationOptions {
			out.Spec.Defaults.AirflowConfigurationOptions[k] = v
		}
	}
	if in.Spec.Defaults.Tags != nil {
		out.Spec.Defaults.Tags = make(map[string]string, len(in.Spec.Defaults.Tags))
		for k, v := range in.Spec.Defaults.Tags {
			out.Spec.Defaults.Tags[k] = v
		}
	}
	if in.Spec.Defaults.SyncedLabels != nil {
		out.Spec.Defaults.SyncedLabels = make(map[string]string, len(in.Spec.Defaults.SyncedLabels))
		for k, v := range in.Spec.Defaults.SyncedLabels {
			out.Spec.Defaults.SyncedLabels[k] = v
		}
	}
	if in.Spec.Defaults.SyncedAnnotations != nil {
		out.Spec.Defaults.SyncedAnnotations = make(map[string]string, len(in.Spec.Defaults.SyncedAnnotations))
		for k, v := range in.Spec.Defaults.SyncedAnnotations {
			out.Spec.Defaults.SyncedAnnotations[k] = v
		}
	}
	// Status
	out.Status = in.Status
	if in.Status.Conditions != nil {
		out.Status.Conditions = make([]metav1.Condition, len(in.Status.Conditions))
		copy(out.Status.Conditions, in.Status.Conditions)
	}
	if in.Status.EffectiveConfig.Mandatory.AirflowConfigurationOptions != nil {
		out.Status.EffectiveConfig.Mandatory.AirflowConfigurationOptions = make(map[string]string, len(in.Status.EffectiveConfig.Mandatory.AirflowConfigurationOptions))
		for k, v := range in.Status.EffectiveConfig.Mandatory.AirflowConfigurationOptions {
			out.Status.EffectiveConfig.Mandatory.AirflowConfigurationOptions[k] = v
		}
	}
	if in.Status.EffectiveConfig.Mandatory.Tags != nil {
		out.Status.EffectiveConfig.Mandatory.Tags = make(map[string]string, len(in.Status.EffectiveConfig.Mandatory.Tags))
		for k, v := range in.Status.EffectiveConfig.Mandatory.Tags {
			out.Status.EffectiveConfig.Mandatory.Tags[k] = v
		}
	}
	if in.Status.EffectiveConfig.Mandatory.SyncedLabels != nil {
		out.Status.EffectiveConfig.Mandatory.SyncedLabels = make(map[string]string, len(in.Status.EffectiveConfig.Mandatory.SyncedLabels))
		for k, v := range in.Status.EffectiveConfig.Mandatory.SyncedLabels {
			out.Status.EffectiveConfig.Mandatory.SyncedLabels[k] = v
		}
	}
	if in.Status.EffectiveConfig.Mandatory.SyncedAnnotations != nil {
		out.Status.EffectiveConfig.Mandatory.SyncedAnnotations = make(map[string]string, len(in.Status.EffectiveConfig.Mandatory.SyncedAnnotations))
		for k, v := range in.Status.EffectiveConfig.Mandatory.SyncedAnnotations {
			out.Status.EffectiveConfig.Mandatory.SyncedAnnotations[k] = v
		}
	}
	if in.Status.EffectiveConfig.Defaults.AirflowConfigurationOptions != nil {
		out.Status.EffectiveConfig.Defaults.AirflowConfigurationOptions = make(map[string]string, len(in.Status.EffectiveConfig.Defaults.AirflowConfigurationOptions))
		for k, v := range in.Status.EffectiveConfig.Defaults.AirflowConfigurationOptions {
			out.Status.EffectiveConfig.Defaults.AirflowConfigurationOptions[k] = v
		}
	}
	if in.Status.EffectiveConfig.Defaults.Tags != nil {
		out.Status.EffectiveConfig.Defaults.Tags = make(map[string]string, len(in.Status.EffectiveConfig.Defaults.Tags))
		for k, v := range in.Status.EffectiveConfig.Defaults.Tags {
			out.Status.EffectiveConfig.Defaults.Tags[k] = v
		}
	}
	if in.Status.EffectiveConfig.Defaults.SyncedLabels != nil {
		out.Status.EffectiveConfig.Defaults.SyncedLabels = make(map[string]string, len(in.Status.EffectiveConfig.Defaults.SyncedLabels))
		for k, v := range in.Status.EffectiveConfig.Defaults.SyncedLabels {
			out.Status.EffectiveConfig.Defaults.SyncedLabels[k] = v
		}
	}
	if in.Status.EffectiveConfig.Defaults.SyncedAnnotations != nil {
		out.Status.EffectiveConfig.Defaults.SyncedAnnotations = make(map[string]string, len(in.Status.EffectiveConfig.Defaults.SyncedAnnotations))
		for k, v := range in.Status.EffectiveConfig.Defaults.SyncedAnnotations {
			out.Status.EffectiveConfig.Defaults.SyncedAnnotations[k] = v
		}
	}
}

func (in *MWAAConfig) DeepCopy() *MWAAConfig {
	if in == nil {
		return nil
	}
	out := new(MWAAConfig)
	in.DeepCopyInto(out)
	return out
}

func (in *MWAAConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *MWAAConfigList) DeepCopyInto(out *MWAAConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]MWAAConfig, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *MWAAConfigList) DeepCopy() *MWAAConfigList {
	if in == nil {
		return nil
	}
	out := new(MWAAConfigList)
	in.DeepCopyInto(out)
	return out
}

func (in *MWAAConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
