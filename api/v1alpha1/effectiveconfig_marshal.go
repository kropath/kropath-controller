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
	"encoding/json"
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// marshalConfigStatus renders a <ResourceFamily>ConfigStatus, omitting the
// "effectiveConfig" key entirely when effectiveConfig equals its Go zero
// value. Every <ResourceFamily>ConfigStatus type in this package declares
// EffectiveConfig as a non-pointer struct (so reconcilers can keep assigning
// it by value), but encoding/json's `omitempty` is a no-op on non-pointer
// struct fields -- it only omits booleans, numbers, strings, and nil
// pointers/interfaces/maps/slices, never a zero-valued struct. Left to the
// struct tag alone, a withheld effectiveConfig still serializes as
// "effectiveConfig": {"aws":{},"defaults":{},"mandatory":{}} instead of being
// absent from the object.
//
// Per docs/specs/controller-account-region-placement.md §6.2/§6.4 and
// ADR-015 §5.5, an absent effectiveConfig is the deliberate contract: it is
// the backstop that fails a kro RGD outright (missing key access) rather
// than letting it silently build a cloud resource from an empty cascade.
// Each <ResourceFamily>ConfigStatus.MarshalJSON method in this package
// delegates here instead of duplicating this logic (KRO-1199).
func marshalConfigStatus(effectiveConfig interface{}, observedGeneration int64, syncedTimestamp string, conditions []metav1.Condition) ([]byte, error) {
	out := map[string]json.RawMessage{}

	zero := reflect.Zero(reflect.TypeOf(effectiveConfig)).Interface()
	if !reflect.DeepEqual(effectiveConfig, zero) {
		b, err := json.Marshal(effectiveConfig)
		if err != nil {
			return nil, err
		}
		out["effectiveConfig"] = b
	}
	if observedGeneration != 0 {
		b, err := json.Marshal(observedGeneration)
		if err != nil {
			return nil, err
		}
		out["observedGeneration"] = b
	}
	if syncedTimestamp != "" {
		b, err := json.Marshal(syncedTimestamp)
		if err != nil {
			return nil, err
		}
		out["syncedTimestamp"] = b
	}
	if len(conditions) > 0 {
		b, err := json.Marshal(conditions)
		if err != nil {
			return nil, err
		}
		out["conditions"] = b
	}
	return json.Marshal(out)
}
