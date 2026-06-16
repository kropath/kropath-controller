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
) cascade.EffectiveS3Cascade {
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

func TestMergeS3Cascade_AC1GlobalMandatoryEncryptionWins(t *testing.T) {
	got := mergeS3All(
		cascade.S3Section{EncryptionAlgorithm: "aws:kms"},
		s3Zero, s3Zero, s3Zero, s3Zero, s3Zero, s3Zero, s3Zero,
	)

	if got.Mandatory.EncryptionAlgorithm != "aws:kms" {
		t.Fatalf("mandatory.encryptionAlgorithm = %q, want aws:kms", got.Mandatory.EncryptionAlgorithm)
	}
	if got.Defaults.EncryptionAlgorithm != "" {
		t.Fatalf("defaults.encryptionAlgorithm = %q, want empty", got.Defaults.EncryptionAlgorithm)
	}
}

func TestMergeS3Cascade_AC2Level1WinsOverLevel3(t *testing.T) {
	got := mergeS3All(
		cascade.S3Section{EncryptionAlgorithm: "aws:kms"},
		s3Zero,
		cascade.S3Section{EncryptionAlgorithm: "AES256"},
		s3Zero, s3Zero, s3Zero, s3Zero, s3Zero,
	)

	if got.Mandatory.EncryptionAlgorithm != "aws:kms" {
		t.Fatalf("mandatory.encryptionAlgorithm = %q, want level-1 value", got.Mandatory.EncryptionAlgorithm)
	}
}

func TestMergeS3Cascade_AC3BlockPublicAccess(t *testing.T) {
	got := mergeS3All(
		cascade.S3Section{BlockPublicAccess: true},
		s3Zero, s3Zero, s3Zero, s3Zero, s3Zero, s3Zero, s3Zero,
	)

	if !got.Mandatory.BlockPublicAccess {
		t.Fatal("mandatory.blockPublicAccess = false, want true")
	}
}

func TestMergeS3Cascade_AC4EnforceHttpsOnly(t *testing.T) {
	got := mergeS3All(
		cascade.S3Section{EnforceHttpsOnly: true},
		s3Zero,
		cascade.S3Section{EnforceHttpsOnly: false},
		s3Zero, s3Zero, s3Zero, s3Zero, s3Zero,
	)

	if !got.Mandatory.EnforceHttpsOnly {
		t.Fatal("mandatory.enforceHttpsOnly = false, want true")
	}
}

func TestMergeS3Cascade_AC5DefaultsOnly(t *testing.T) {
	got := mergeS3All(
		s3Zero, s3Zero, s3Zero, s3Zero,
		cascade.S3Section{EncryptionAlgorithm: "aws:kms"},
		s3Zero, s3Zero, s3Zero,
	)

	if got.Defaults.EncryptionAlgorithm != "aws:kms" {
		t.Fatalf("defaults.encryptionAlgorithm = %q, want aws:kms", got.Defaults.EncryptionAlgorithm)
	}
	if got.Mandatory.EncryptionAlgorithm != "" {
		t.Fatalf("mandatory.encryptionAlgorithm = %q, want empty", got.Mandatory.EncryptionAlgorithm)
	}
}

func TestMergeS3Cascade_AC6GlobalMandatoryKmsKeyWins(t *testing.T) {
	got := mergeS3All(
		cascade.S3Section{KmsKeyArn: "arn:aws:kms:us-east-1:123:key/global"},
		cascade.S3Section{KmsKeyArn: "arn:aws:kms:us-east-1:123:key/local"},
		s3Zero, s3Zero, s3Zero, s3Zero, s3Zero, s3Zero,
	)

	if got.Mandatory.KmsKeyArn != "arn:aws:kms:us-east-1:123:key/global" {
		t.Fatalf("mandatory.kmsKeyArn = %q, want global value", got.Mandatory.KmsKeyArn)
	}
}

func TestMergeS3Cascade_AC8LoggingPair(t *testing.T) {
	got := mergeS3All(
		cascade.S3Section{LoggingEnabled: true, LogDeliveryBucket: "org-access-logs"},
		s3Zero, s3Zero, s3Zero, s3Zero, s3Zero, s3Zero, s3Zero,
	)

	if !got.Mandatory.LoggingEnabled {
		t.Fatal("mandatory.loggingEnabled = false, want true")
	}
	if got.Mandatory.LogDeliveryBucket != "org-access-logs" {
		t.Fatalf("mandatory.logDeliveryBucket = %q, want org-access-logs", got.Mandatory.LogDeliveryBucket)
	}
}

func TestMergeS3Cascade_AC9ObjectLockZeroValuePassesThrough(t *testing.T) {
	got := mergeS3All(
		cascade.S3Section{ObjectLockMode: ""},
		s3Zero, s3Zero, s3Zero, s3Zero, s3Zero, s3Zero, s3Zero,
	)

	if got.Mandatory.ObjectLockMode != "" {
		t.Fatalf("mandatory.objectLockMode = %q, want empty", got.Mandatory.ObjectLockMode)
	}
}

func TestMergeS3CascadeMapsAndNamingTemplate(t *testing.T) {
	got := mergeS3All(
		cascade.S3Section{
			NamingTemplate: "global-{namespace}",
			Tags:           map[string]string{"global": "one"},
		},
		cascade.S3Section{
			NamingTemplate: "local-{namespace}",
			Tags:           map[string]string{"local": "two", "shared": "local"},
		},
		s3Zero,
		cascade.S3Section{
			Tags:              map[string]string{"shared": "s3"},
			SyncedLabels:      map[string]string{"label": "s3"},
			SyncedAnnotations: map[string]string{"anno": "s3"},
		},
		cascade.S3Section{
			Tags:              map[string]string{"defaults": "local"},
			SyncedLabels:      map[string]string{"label": "local"},
			SyncedAnnotations: map[string]string{"anno": "local"},
		},
		cascade.S3Section{
			Tags:              map[string]string{"defaults": "global"},
			SyncedLabels:      map[string]string{"label": "global"},
			SyncedAnnotations: map[string]string{"anno": "global"},
		},
		cascade.S3Section{
			Tags:              map[string]string{"kpc": "local"},
			SyncedLabels:      map[string]string{"label": "kpc"},
			SyncedAnnotations: map[string]string{"anno": "kpc"},
		},
		cascade.S3Section{
			Tags:              map[string]string{"kpc": "global"},
			SyncedLabels:      map[string]string{"label": "kpc-global"},
			SyncedAnnotations: map[string]string{"anno": "kpc-global"},
		},
	)

	if got.Mandatory.NamingTemplate != "global-{namespace}" {
		t.Fatalf("mandatory.namingTemplate = %q, want global template", got.Mandatory.NamingTemplate)
	}
	if got.Mandatory.Tags["global"] != "one" {
		t.Fatalf("mandatory.tags[global] = %q, want one", got.Mandatory.Tags["global"])
	}
	if got.Mandatory.Tags["shared"] != "local" {
		t.Fatalf("mandatory.tags[shared] = %q, want local", got.Mandatory.Tags["shared"])
	}
	if got.Mandatory.SyncedLabels["label"] != "s3" {
		t.Fatalf("mandatory.syncedLabels[label] = %q, want s3", got.Mandatory.SyncedLabels["label"])
	}
	if got.Defaults.Tags["defaults"] != "local" {
		t.Fatalf("defaults.tags[defaults] = %q, want local", got.Defaults.Tags["defaults"])
	}
	if got.Defaults.SyncedAnnotations["anno"] != "local" {
		t.Fatalf("defaults.syncedAnnotations[anno] = %q, want local", got.Defaults.SyncedAnnotations["anno"])
	}
}
