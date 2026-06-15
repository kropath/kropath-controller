// Copyright 2026 The kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import "testing"

func TestLeaderElectionNamespace(t *testing.T) {
	t.Run("defaults to default", func(t *testing.T) {
		t.Setenv("LEADER_ELECTION_NAMESPACE", "")
		t.Setenv("POD_NAMESPACE", "")

		if got := leaderElectionNamespace(); got != "default" {
			t.Fatalf("leaderElectionNamespace() = %q, want %q", got, "default")
		}
	})

	t.Run("uses pod namespace when set", func(t *testing.T) {
		t.Setenv("LEADER_ELECTION_NAMESPACE", "")
		t.Setenv("POD_NAMESPACE", "pod-ns")

		if got := leaderElectionNamespace(); got != "pod-ns" {
			t.Fatalf("leaderElectionNamespace() = %q, want %q", got, "pod-ns")
		}
	})

	t.Run("leader election env takes precedence", func(t *testing.T) {
		t.Setenv("LEADER_ELECTION_NAMESPACE", "lease-ns")
		t.Setenv("POD_NAMESPACE", "pod-ns")

		if got := leaderElectionNamespace(); got != "lease-ns" {
			t.Fatalf("leaderElectionNamespace() = %q, want %q", got, "lease-ns")
		}
	})
}
