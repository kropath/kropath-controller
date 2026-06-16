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

var s3Zero = cascade.S3Section{}

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

func TestMergeS3Cascade_MandatoryCascadeOrder(t *testing.T) {
	cases := []struct {
		name                   string
		globalKropathMandatory cascade.S3Section
		localKropathMandatory  cascade.S3Section
		globalS3Mandatory      cascade.S3Section
		localS3Mandatory       cascade.S3Section
		wantEncryption         string
		wantBlock              bool
		wantRetention          int64
	}{
		{
			name:                   "level1-wins",
			globalKropathMandatory: cascade.S3Section{EncryptionAlgorithm: "aws:kms", BlockPublicAccess: true, ObjectLockRetentionDays: 90},
			localKropathMandatory:  cascade.S3Section{EncryptionAlgorithm: "AES256", BlockPublicAccess: false, ObjectLockRetentionDays: 30},
			globalS3Mandatory:      cascade.S3Section{EncryptionAlgorithm: "aws:kms", BlockPublicAccess: false, ObjectLockRetentionDays: 14},
			localS3Mandatory:       cascade.S3Section{EncryptionAlgorithm: "AES256", BlockPublicAccess: true, ObjectLockRetentionDays: 7},
			wantEncryption:         "aws:kms",
			wantBlock:              true,
			wantRetention:          90,
		},
		{
			name:                   "level2-wins-when-1-absent",
			globalKropathMandatory: s3Zero,
			localKropathMandatory:  cascade.S3Section{EncryptionAlgorithm: "AES256", BlockPublicAccess: true, ObjectLockRetentionDays: 30},
			globalS3Mandatory:      cascade.S3Section{EncryptionAlgorithm: "aws:kms", BlockPublicAccess: false, ObjectLockRetentionDays: 14},
			localS3Mandatory:       cascade.S3Section{EncryptionAlgorithm: "AES256", BlockPublicAccess: false, ObjectLockRetentionDays: 7},
			wantEncryption:         "AES256",
			wantBlock:              true,
			wantRetention:          30,
		},
		{
			name:                   "level3-wins-when-1-2-absent",
			globalKropathMandatory: s3Zero,
			localKropathMandatory:  s3Zero,
			globalS3Mandatory:      cascade.S3Section{EncryptionAlgorithm: "aws:kms", BlockPublicAccess: true, ObjectLockRetentionDays: 14},
			localS3Mandatory:       cascade.S3Section{EncryptionAlgorithm: "AES256", BlockPublicAccess: false, ObjectLockRetentionDays: 7},
			wantEncryption:         "aws:kms",
			wantBlock:              true,
			wantRetention:          14,
		},
		{
			name:                   "level4-wins-when-1-2-3-absent",
			globalKropathMandatory: s3Zero,
			localKropathMandatory:  s3Zero,
			globalS3Mandatory:      s3Zero,
			localS3Mandatory:       cascade.S3Section{EncryptionAlgorithm: "AES256", BlockPublicAccess: true, ObjectLockRetentionDays: 7},
			wantEncryption:         "AES256",
			wantBlock:              true,
			wantRetention:          7,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeS3All(tc.globalKropathMandatory, tc.localKropathMandatory, tc.globalS3Mandatory, tc.localS3Mandatory, s3Zero, s3Zero, s3Zero, s3Zero)
			if got.Mandatory.EncryptionAlgorithm != tc.wantEncryption {
				t.Fatalf("mandatory.encryptionAlgorithm = %q, want %q", got.Mandatory.EncryptionAlgorithm, tc.wantEncryption)
			}
			if got.Mandatory.BlockPublicAccess != tc.wantBlock {
				t.Fatalf("mandatory.blockPublicAccess = %v, want %v", got.Mandatory.BlockPublicAccess, tc.wantBlock)
			}
			if got.Mandatory.ObjectLockRetentionDays != tc.wantRetention {
				t.Fatalf("mandatory.objectLockRetentionDays = %d, want %d", got.Mandatory.ObjectLockRetentionDays, tc.wantRetention)
			}
		})
	}
}

func TestMergeS3Cascade_DefaultsCascadeOrder(t *testing.T) {
	cases := []struct {
		name                  string
		localS3Defaults       cascade.S3Section
		globalS3Defaults      cascade.S3Section
		localKropathDefaults  cascade.S3Section
		globalKropathDefaults cascade.S3Section
		wantEncryption        string
		wantBlock             bool
		wantRetention         int64
	}{
		{
			name:                  "level6-wins",
			localS3Defaults:       cascade.S3Section{EncryptionAlgorithm: "AES256", BlockPublicAccess: false, ObjectLockRetentionDays: 30},
			globalS3Defaults:      cascade.S3Section{EncryptionAlgorithm: "aws:kms", BlockPublicAccess: true, ObjectLockRetentionDays: 14},
			localKropathDefaults:  cascade.S3Section{EncryptionAlgorithm: "AES256", BlockPublicAccess: true, ObjectLockRetentionDays: 7},
			globalKropathDefaults: cascade.S3Section{EncryptionAlgorithm: "aws:kms", BlockPublicAccess: false, ObjectLockRetentionDays: 3},
			wantEncryption:        "AES256",
			wantBlock:             true,
			wantRetention:         30,
		},
		{
			name:                  "level7-wins-when-6-absent",
			localS3Defaults:       s3Zero,
			globalS3Defaults:      cascade.S3Section{EncryptionAlgorithm: "aws:kms", BlockPublicAccess: true, ObjectLockRetentionDays: 14},
			localKropathDefaults:  cascade.S3Section{EncryptionAlgorithm: "AES256", BlockPublicAccess: true, ObjectLockRetentionDays: 7},
			globalKropathDefaults: cascade.S3Section{EncryptionAlgorithm: "aws:kms", BlockPublicAccess: false, ObjectLockRetentionDays: 3},
			wantEncryption:        "aws:kms",
			wantBlock:             true,
			wantRetention:         14,
		},
		{
			name:                  "level8-wins-when-6-7-absent",
			localS3Defaults:       s3Zero,
			globalS3Defaults:      s3Zero,
			localKropathDefaults:  cascade.S3Section{EncryptionAlgorithm: "AES256", BlockPublicAccess: true, ObjectLockRetentionDays: 7},
			globalKropathDefaults: cascade.S3Section{EncryptionAlgorithm: "aws:kms", BlockPublicAccess: false, ObjectLockRetentionDays: 3},
			wantEncryption:        "AES256",
			wantBlock:             true,
			wantRetention:         7,
		},
		{
			name:                  "level9-wins-when-6-7-8-absent",
			localS3Defaults:       s3Zero,
			globalS3Defaults:      s3Zero,
			localKropathDefaults:  s3Zero,
			globalKropathDefaults: cascade.S3Section{EncryptionAlgorithm: "aws:kms", BlockPublicAccess: true, ObjectLockRetentionDays: 3},
			wantEncryption:        "aws:kms",
			wantBlock:             true,
			wantRetention:         3,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeS3All(s3Zero, s3Zero, s3Zero, s3Zero, tc.localS3Defaults, tc.globalS3Defaults, tc.localKropathDefaults, tc.globalKropathDefaults)
			if got.Defaults.EncryptionAlgorithm != tc.wantEncryption {
				t.Fatalf("defaults.encryptionAlgorithm = %q, want %q", got.Defaults.EncryptionAlgorithm, tc.wantEncryption)
			}
			if got.Defaults.BlockPublicAccess != tc.wantBlock {
				t.Fatalf("defaults.blockPublicAccess = %v, want %v", got.Defaults.BlockPublicAccess, tc.wantBlock)
			}
			if got.Defaults.ObjectLockRetentionDays != tc.wantRetention {
				t.Fatalf("defaults.objectLockRetentionDays = %d, want %d", got.Defaults.ObjectLockRetentionDays, tc.wantRetention)
			}
		})
	}
}

func TestMergeS3Cascade_AC1GlobalMandatoryEncryption(t *testing.T) {
	got := mergeS3All(
		cascade.S3Section{EncryptionAlgorithm: "aws:kms"},
		s3Zero, s3Zero, s3Zero,
		s3Zero, s3Zero, s3Zero, s3Zero,
	)

	if got.Mandatory.EncryptionAlgorithm != "aws:kms" {
		t.Fatalf("mandatory.encryptionAlgorithm = %q, want aws:kms", got.Mandatory.EncryptionAlgorithm)
	}
	if got.Defaults.EncryptionAlgorithm != "" {
		t.Fatalf("defaults.encryptionAlgorithm = %q, want empty", got.Defaults.EncryptionAlgorithm)
	}
}

func TestMergeS3Cascade_AC4BooleanMandatoryWinsOverFalse(t *testing.T) {
	got := mergeS3All(
		cascade.S3Section{EnforceHttpsOnly: true},
		s3Zero,
		cascade.S3Section{EnforceHttpsOnly: false},
		s3Zero,
		s3Zero, s3Zero, s3Zero, s3Zero,
	)

	if !got.Mandatory.EnforceHttpsOnly {
		t.Fatal("mandatory.enforceHttpsOnly = false, want true")
	}
}

func TestMergeS3Cascade_AC5DefaultsOnly(t *testing.T) {
	got := mergeS3All(
		s3Zero, s3Zero, s3Zero, s3Zero,
		s3Zero,
		cascade.S3Section{EncryptionAlgorithm: "aws:kms"},
		s3Zero, s3Zero,
	)

	if got.Defaults.EncryptionAlgorithm != "aws:kms" {
		t.Fatalf("defaults.encryptionAlgorithm = %q, want aws:kms", got.Defaults.EncryptionAlgorithm)
	}
	if got.Mandatory.EncryptionAlgorithm != "" {
		t.Fatalf("mandatory.encryptionAlgorithm = %q, want empty", got.Mandatory.EncryptionAlgorithm)
	}
}

func TestMergeS3Cascade_AC6GlobalKmsKeyWins(t *testing.T) {
	got := mergeS3All(
		cascade.S3Section{KmsKeyArn: "arn:aws:kms:us-east-1:123:key/global"},
		cascade.S3Section{KmsKeyArn: "arn:aws:kms:us-east-1:123:key/local"},
		s3Zero, s3Zero,
		s3Zero, s3Zero, s3Zero, s3Zero,
	)

	if got.Mandatory.KmsKeyArn != "arn:aws:kms:us-east-1:123:key/global" {
		t.Fatalf("mandatory.kmsKeyArn = %q, want global", got.Mandatory.KmsKeyArn)
	}
}

func TestMergeS3Cascade_AC8LoggingPair(t *testing.T) {
	got := mergeS3All(
		cascade.S3Section{LoggingEnabled: true, LogDeliveryBucket: "org-access-logs"},
		s3Zero, s3Zero, s3Zero,
		s3Zero, s3Zero, s3Zero, s3Zero,
	)

	if !got.Mandatory.LoggingEnabled {
		t.Fatal("mandatory.loggingEnabled = false, want true")
	}
	if got.Mandatory.LogDeliveryBucket != "org-access-logs" {
		t.Fatalf("mandatory.logDeliveryBucket = %q, want org-access-logs", got.Mandatory.LogDeliveryBucket)
	}
}

func TestMergeS3Cascade_AC9ZeroValueObjectLockPassesThrough(t *testing.T) {
	got := mergeS3All(
		cascade.S3Section{ObjectLockMode: ""},
		s3Zero, s3Zero, s3Zero,
		s3Zero, s3Zero, s3Zero, s3Zero,
	)

	if got.Mandatory.ObjectLockMode != "" {
		t.Fatalf("mandatory.objectLockMode = %q, want empty", got.Mandatory.ObjectLockMode)
	}
	if got.Mandatory.ObjectLockRetentionDays != 0 {
		t.Fatalf("mandatory.objectLockRetentionDays = %d, want 0", got.Mandatory.ObjectLockRetentionDays)
	}
}
