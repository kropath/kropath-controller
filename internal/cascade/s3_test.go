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
	globalS3Mandatory,
	localS3Mandatory,
	localS3Defaults,
	globalS3Defaults,
	localKropathDefaults,
	globalKropathDefaults cascade.S3Section,
) cascade.EffectiveS3Config {
	return cascade.MergeS3Cascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalS3Mandatory,
		localS3Mandatory,
		localS3Defaults,
		globalS3Defaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

func TestMergeS3Cascade_MandatoryLevel1Wins(t *testing.T) {
	got := mergeS3All(
		cascade.S3Section{EncryptionAlgorithm: "aws:kms"},
		cascade.S3Section{EncryptionAlgorithm: "AES256"},
		cascade.S3Section{EncryptionAlgorithm: "AES256"},
		cascade.S3Section{EncryptionAlgorithm: "aws:kms"},
		zeroS3, zeroS3, zeroS3, zeroS3,
	)

	if got.Mandatory.EncryptionAlgorithm != "aws:kms" {
		t.Fatalf("mandatory.encryptionAlgorithm = %q, want %q", got.Mandatory.EncryptionAlgorithm, "aws:kms")
	}
}

func TestMergeS3Cascade_MandatoryBooleanLevel1Wins(t *testing.T) {
	got := mergeS3All(
		cascade.S3Section{BlockPublicAccess: true, EnforceHttpsOnly: true, LoggingEnabled: true},
		cascade.S3Section{BlockPublicAccess: false, EnforceHttpsOnly: false, LoggingEnabled: false},
		cascade.S3Section{BlockPublicAccess: false, EnforceHttpsOnly: false, LoggingEnabled: false},
		cascade.S3Section{BlockPublicAccess: false, EnforceHttpsOnly: false, LoggingEnabled: false},
		zeroS3, zeroS3, zeroS3, zeroS3,
	)

	if !got.Mandatory.BlockPublicAccess {
		t.Fatal("mandatory.blockPublicAccess should be true")
	}
	if !got.Mandatory.EnforceHttpsOnly {
		t.Fatal("mandatory.enforceHttpsOnly should be true")
	}
	if !got.Mandatory.LoggingEnabled {
		t.Fatal("mandatory.loggingEnabled should be true")
	}
}

func TestMergeS3Cascade_DefaultsCascadeOrder(t *testing.T) {
	got := mergeS3All(
		zeroS3, zeroS3, zeroS3, zeroS3,
		cascade.S3Section{Versioning: "local-s3"},
		cascade.S3Section{Versioning: "global-s3"},
		cascade.S3Section{Versioning: "local-kpc"},
		cascade.S3Section{Versioning: "global-kpc"},
	)

	if got.Defaults.Versioning != "local-s3" {
		t.Fatalf("defaults.versioning = %q, want %q", got.Defaults.Versioning, "local-s3")
	}
}

func TestMergeS3Cascade_DefaultsGlobalKropathFallback(t *testing.T) {
	got := mergeS3All(
		zeroS3, zeroS3, zeroS3, zeroS3,
		zeroS3, zeroS3,
		zeroS3,
		cascade.S3Section{EncryptionAlgorithm: "aws:kms"},
	)

	if got.Defaults.EncryptionAlgorithm != "aws:kms" {
		t.Fatalf("defaults.encryptionAlgorithm = %q, want %q", got.Defaults.EncryptionAlgorithm, "aws:kms")
	}
	if got.Mandatory.EncryptionAlgorithm != "" {
		t.Fatalf("mandatory.encryptionAlgorithm = %q, want empty", got.Mandatory.EncryptionAlgorithm)
	}
}

func TestMergeS3Cascade_PassthroughZeroValues(t *testing.T) {
	got := mergeS3All(zeroS3, zeroS3, zeroS3, zeroS3, zeroS3, zeroS3, zeroS3, zeroS3)

	if got.Mandatory.ObjectLockMode != "" {
		t.Fatalf("mandatory.objectLockMode = %q, want empty", got.Mandatory.ObjectLockMode)
	}
	if got.Mandatory.ObjectLockRetentionDays != 0 {
		t.Fatalf("mandatory.objectLockRetentionDays = %d, want 0", got.Mandatory.ObjectLockRetentionDays)
	}
	if got.Defaults.ObjectLockMode != "" {
		t.Fatalf("defaults.objectLockMode = %q, want empty", got.Defaults.ObjectLockMode)
	}
	if got.Defaults.ObjectLockRetentionDays != 0 {
		t.Fatalf("defaults.objectLockRetentionDays = %d, want 0", got.Defaults.ObjectLockRetentionDays)
	}
}
