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

type AWSProviderIdentity struct {
	AccountID string `json:"accountId,omitempty"`
	Region    string `json:"region,omitempty"`
}

// S3Section holds the AWS S3 governance fields shared by
// AWSKropathConfig.spec.mandatory.s3 / .defaults.s3.
//
// Zero value of each field is the permissive sentinel (not enforced).
type S3Section struct {
	// EncryptionAlgorithm controls the bucket default encryption algorithm.
	// Empty string = no enforcement.
	EncryptionAlgorithm string `json:"encryptionAlgorithm,omitempty"`

	// KmsKeyArn is the ARN of the KMS key to use for SSE-KMS.
	// Empty string = no enforcement.
	KmsKeyArn string `json:"kmsKeyArn,omitempty"`

	// BlockPublicAccess blocks public access when true.
	// false (zero value) = not enforced.
	BlockPublicAccess bool `json:"blockPublicAccess,omitempty"`

	// Versioning controls the bucket versioning state.
	// Empty string = no enforcement.
	Versioning string `json:"versioning,omitempty"`

	// LoggingEnabled requires server access logging when true.
	// false (zero value) = not enforced.
	LoggingEnabled bool `json:"loggingEnabled,omitempty"`

	// LogDeliveryBucket is the target bucket for server access logs.
	// Empty string = not enforced.
	LogDeliveryBucket string `json:"logDeliveryBucket,omitempty"`

	// EnforceHttpsOnly denies non-TLS access when true.
	// false (zero value) = not enforced.
	EnforceHttpsOnly bool `json:"enforceHttpsOnly,omitempty"`

	// ObjectLockMode controls S3 Object Lock governance mode.
	// Empty string = not enforced.
	ObjectLockMode string `json:"objectLockMode,omitempty"`

	// ObjectLockRetentionDays controls S3 Object Lock retention.
	// 0 = not enforced.
	ObjectLockRetentionDays int64 `json:"objectLockRetentionDays,omitempty"`
}

type AWSKropathConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec AWSKropathConfigSpec `json:"spec,omitempty"`
}

type AWSKropathConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []AWSKropathConfig `json:"items"`
}

type AWSKropathConfigSpec struct {
	Mandatory AWSKropathConfigTier `json:"mandatory,omitempty"`
	Defaults  AWSKropathConfigTier `json:"defaults,omitempty"`
	AWS       AWSProviderIdentity  `json:"aws,omitempty"`
}

type AWSKropathConfigTier struct {
	IAM cascade.IAMSection `json:"iam,omitempty"`
	S3  cascade.S3Section  `json:"s3,omitempty"`
}

type AWSIAMConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSIAMConfigSpec   `json:"spec,omitempty"`
	Status AWSIAMConfigStatus `json:"status,omitempty"`
}

type AWSIAMConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []AWSIAMConfig `json:"items"`
}

type AWSIAMConfigSpec struct {
	Mandatory cascade.IAMSection `json:"mandatory,omitempty"`
	Defaults  cascade.IAMSection `json:"defaults,omitempty"`
}

type AWSIAMConfigStatus struct {
	EffectiveConfig    AWSEffectiveIAMConfig `json:"effectiveConfig,omitempty"`
	ObservedGeneration int64                 `json:"observedGeneration,omitempty"`
	SyncedTimestamp    string                `json:"syncedTimestamp,omitempty"`
}

type AWSEffectiveIAMConfig struct {
	AWS       AWSProviderIdentity `json:"aws,omitempty"`
	Mandatory cascade.IAMSection  `json:"mandatory,omitempty"`
	Defaults  cascade.IAMSection  `json:"defaults,omitempty"`
}

func (in *AWSKropathConfig) DeepCopyInto(out *AWSKropathConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
}

func (in *AWSKropathConfig) DeepCopy() *AWSKropathConfig {
	if in == nil {
		return nil
	}
	out := new(AWSKropathConfig)
	in.DeepCopyInto(out)
	return out
}

func (in *AWSKropathConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *AWSKropathConfigList) DeepCopyInto(out *AWSKropathConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]AWSKropathConfig, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *AWSKropathConfigList) DeepCopy() *AWSKropathConfigList {
	if in == nil {
		return nil
	}
	out := new(AWSKropathConfigList)
	in.DeepCopyInto(out)
	return out
}

func (in *AWSKropathConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *AWSIAMConfig) DeepCopyInto(out *AWSIAMConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

func (in *AWSIAMConfig) DeepCopy() *AWSIAMConfig {
	if in == nil {
		return nil
	}
	out := new(AWSIAMConfig)
	in.DeepCopyInto(out)
	return out
}

func (in *AWSIAMConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *AWSIAMConfigList) DeepCopyInto(out *AWSIAMConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]AWSIAMConfig, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *AWSIAMConfigList) DeepCopy() *AWSIAMConfigList {
	if in == nil {
		return nil
	}
	out := new(AWSIAMConfigList)
	in.DeepCopyInto(out)
	return out
}

func (in *AWSIAMConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
