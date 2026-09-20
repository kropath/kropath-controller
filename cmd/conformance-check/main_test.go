// Copyright 2026 The kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunRequiresACKNamespace(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{}, &stdout, &stderr)
	if code != 2 {
		t.Errorf("run() = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "--ack-namespace is required") {
		t.Errorf("stderr = %q, want it to mention --ack-namespace", stderr.String())
	}
}

func TestRunRejectsInvalidOutputFormat(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"--ack-namespace", "ack-system", "--output", "xml"}, &stdout, &stderr)
	if code != 2 {
		t.Errorf("run() = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), `--output must be "text" or "json"`) {
		t.Errorf("stderr = %q, want it to mention the allowed --output values", stderr.String())
	}
}

func TestRunRejectsUnknownFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"--not-a-real-flag"}, &stdout, &stderr)
	if code != 2 {
		t.Errorf("run() = %d, want 2", code)
	}
}

func TestRunFailsWithoutKubeconfig(t *testing.T) {
	// No in-cluster config and no kubeconfig on the test runner's $HOME (or an
	// explicitly broken one) makes ctrl.GetConfig() fail before any cluster
	// call — this exercises the error path once flag validation has passed.
	t.Setenv("KUBECONFIG", "/nonexistent/kubeconfig")
	t.Setenv("HOME", t.TempDir())

	var stdout, stderr bytes.Buffer
	code := run([]string{"--ack-namespace", "ack-system"}, &stdout, &stderr)
	if code != 2 {
		t.Errorf("run() = %d, want 2 (stderr: %s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "unable to load kubeconfig") {
		t.Errorf("stderr = %q, want it to mention the kubeconfig failure", stderr.String())
	}
}
