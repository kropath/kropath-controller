// Copyright 2026 The kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package policydocument

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func setCondition(conditions []metav1.Condition, condition metav1.Condition) []metav1.Condition {
	condition.LastTransitionTime = metav1.Now()
	for i := range conditions {
		if conditions[i].Type == condition.Type {
			if conditions[i].Status == condition.Status &&
				conditions[i].Reason == condition.Reason &&
				conditions[i].Message == condition.Message {
				condition.LastTransitionTime = conditions[i].LastTransitionTime
			}
			conditions[i] = condition
			return conditions
		}
	}
	return append(conditions, condition)
}

func readyCondition(status metav1.ConditionStatus, reason, message string, generation int64) metav1.Condition {
	return metav1.Condition{
		Type:               "Ready",
		Status:             status,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: generation,
	}
}

func sourceNotReadyCondition(status metav1.ConditionStatus, reason, message string, generation int64) metav1.Condition {
	return metav1.Condition{
		Type:               "SourceNotReady",
		Status:             status,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: generation,
	}
}

func sidConflictCondition(status metav1.ConditionStatus, reason, message string, generation int64) metav1.Condition {
	return metav1.Condition{
		Type:               "SidConflict",
		Status:             status,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: generation,
	}
}
