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

package cascade_test

import (
	"testing"

	"github.com/kropath/kropath-controller/internal/cascade"
)

var zeroS3 = cascade.S3Section{}

func mergeS3All(
	globalKropathMandatory,
	localKropathMandatory,
	globalS3CfgMandatory,
	localS3CfgMandatory,
	localS3CfgDefaults,
	globalS3CfgDefaults,
	localKropathDefaults,
	globalKropathDefaults cascade.S3Section,
) cascade.EffectiveS3Config {
	return cascade.MergeS3Cascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalS3CfgMandatory,
		localS3CfgMandatory,
		localS3CfgDefaults,
		globalS3CfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

func TestMergeS3Cascade_AC1(t *testing.T) {
	got := mergeS3All(
		cascade.S3Section{EncryptionAlgorithm: "aws:kms"},
		zeroS3, zeroS3, zeroS3,
		zeroS3, zeroS3, zeroS3, zeroS3,
	)
	if got.Mandatory.EncryptionAlgorithm != "aws:kms" {
		t.Fatalf("mandatory.encryptionAlgorithm = %q, want aws:kms", got.Mandatory.EncryptionAlgorithm)
	}
}

func TestMergeS3Cascade_AC2(t *testing.T) {
	got := mergeS3All(
		cascade.S3Section{EncryptionAlgorithm: "aws:kms"},
		zeroS3,
		cascade.S3Section{EncryptionAlgorithm: "AES256"},
		zeroS3,
		zeroS3, zeroS3, zeroS3, zeroS3,
	)
	if got.Mandatory.EncryptionAlgorithm != "aws:kms" {
		t.Fatalf("level-1 must win, got %q", got.Mandatory.EncryptionAlgorithm)
	}
}

func TestMergeS3Cascade_AC3(t *testing.T) {
	got := mergeS3All(
		cascade.S3Section{BlockPublicAccess: true},
		zeroS3, zeroS3, zeroS3,
		zeroS3, zeroS3, zeroS3, zeroS3,
	)
	if !got.Mandatory.BlockPublicAccess {
		t.Fatal("mandatory.blockPublicAccess = false, want true")
	}
}

func TestMergeS3Cascade_AC4(t *testing.T) {
	got := mergeS3All(
		cascade.S3Section{EnforceHttpsOnly: true},
		zeroS3,
		cascade.S3Section{EnforceHttpsOnly: false},
		zeroS3,
		zeroS3, zeroS3, zeroS3, zeroS3,
	)
	if !got.Mandatory.EnforceHttpsOnly {
		t.Fatal("mandatory.enforceHttpsOnly = false, want true")
	}
}

func TestMergeS3Cascade_AC5(t *testing.T) {
	got := mergeS3All(
		zeroS3, zeroS3, zeroS3, zeroS3,
		zeroS3,
		cascade.S3Section{EncryptionAlgorithm: "aws:kms"},
		zeroS3, zeroS3,
	)
	if got.Defaults.EncryptionAlgorithm != "aws:kms" {
		t.Fatalf("defaults.encryptionAlgorithm = %q, want aws:kms", got.Defaults.EncryptionAlgorithm)
	}
	if got.Mandatory.EncryptionAlgorithm != "" {
		t.Fatalf("mandatory.encryptionAlgorithm = %q, want empty", got.Mandatory.EncryptionAlgorithm)
	}
}

func TestMergeS3Cascade_AC6(t *testing.T) {
	got := mergeS3All(
		cascade.S3Section{KmsKeyArn: "arn:aws:kms:us-east-1:123:key/global"},
		zeroS3,
		cascade.S3Section{KmsKeyArn: "arn:aws:kms:us-east-1:123:key/ns"},
		zeroS3,
		zeroS3, zeroS3, zeroS3, zeroS3,
	)
	if got.Mandatory.KmsKeyArn != "arn:aws:kms:us-east-1:123:key/global" {
		t.Fatalf("level-1 must win, got %q", got.Mandatory.KmsKeyArn)
	}
}

func TestMergeS3Cascade_AC8(t *testing.T) {
	got := mergeS3All(
		cascade.S3Section{LoggingEnabled: true, LogDeliveryBucket: "org-access-logs"},
		zeroS3, zeroS3, zeroS3,
		zeroS3, zeroS3, zeroS3, zeroS3,
	)
	if !got.Mandatory.LoggingEnabled {
		t.Fatal("mandatory.loggingEnabled = false, want true")
	}
	if got.Mandatory.LogDeliveryBucket != "org-access-logs" {
		t.Fatalf("mandatory.logDeliveryBucket = %q, want org-access-logs", got.Mandatory.LogDeliveryBucket)
	}
}

func TestMergeS3Cascade_AC9(t *testing.T) {
	got := mergeS3All(
		cascade.S3Section{ObjectLockMode: "", ObjectLockRetentionDays: 0},
		zeroS3, zeroS3, zeroS3,
		zeroS3, zeroS3, zeroS3, zeroS3,
	)
	if got.Mandatory.ObjectLockMode != "" {
		t.Fatalf("mandatory.objectLockMode = %q, want empty", got.Mandatory.ObjectLockMode)
	}
	if got.Mandatory.ObjectLockRetentionDays != 0 {
		t.Fatalf("mandatory.objectLockRetentionDays = %d, want 0", got.Mandatory.ObjectLockRetentionDays)
	}
}

func TestMergeS3Cascade_DefaultsCascadeOrder(t *testing.T) {
	got := mergeS3All(
		zeroS3, zeroS3, zeroS3, zeroS3,
		cascade.S3Section{Versioning: "local-default"},
		cascade.S3Section{Versioning: "global-s3-default"},
		cascade.S3Section{Versioning: "local-kropath-default"},
		cascade.S3Section{Versioning: "global-kropath-default"},
	)
	if got.Defaults.Versioning != "local-default" {
		t.Fatalf("defaults.versioning = %q, want local-default", got.Defaults.Versioning)
	}
}

func TestMergeS3Cascade_MandatoryCascadeOrder(t *testing.T) {
	got := mergeS3All(
		cascade.S3Section{Versioning: "global-kropath"},
		cascade.S3Section{Versioning: "local-kropath"},
		cascade.S3Section{Versioning: "global-s3"},
		cascade.S3Section{Versioning: "local-s3"},
		zeroS3, zeroS3, zeroS3, zeroS3,
	)
	if got.Mandatory.Versioning != "global-kropath" {
		t.Fatalf("mandatory.versioning = %q, want global-kropath", got.Mandatory.Versioning)
	}
}
