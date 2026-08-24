// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package features

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// StateReader provides per-reconciler runtime state from the coordinator.
// Implemented by *registry.Coordinator; defined here to avoid an import cycle.
type StateReader interface {
	ReconcilerState(pkg string) (state string, missingKinds []string)
}

// Response is the JSON body served by GET /features and printed by "kropath-operator features".
// The Features field uses the package name as the query key for ?name= filtering.
type Response struct {
	Version   string       `json:"version"`
	GitCommit string       `json:"gitCommit"`
	BuildDate string       `json:"buildDate"`
	GoVersion string       `json:"goVersion"`
	Features  []Reconciler `json:"features"`
}

// Handler returns an http.Handler for the /features endpoint.
//
// When sr is non-nil, State and MissingKinds are populated per-request from the
// live coordinator. When sr is nil (e.g. the offline "features" subcommand), the
// fields are omitted. Only GET and HEAD are accepted; any other method returns 405
// with a JSON error body and an Allow header. A ?name=<pkg> query parameter filters
// the response to the single reconciler whose Package matches exactly; an unknown
// package returns 404 with a JSON error body.
func Handler(ver, gitCommit, buildDate, goVersion string, all []Reconciler, sr StateReader) http.Handler {
	// Normalise nil so marshaling produces [] rather than null.
	if all == nil {
		all = []Reconciler{}
	}

	// Pre-index by Package for O(1) lookup on ?name= queries.
	byPkg := make(map[string]Reconciler, len(all))
	for _, r := range all {
		byPkg[r.Package] = r
	}

	// buildFeatures returns a fresh slice with state populated when sr != nil.
	buildFeatures := func(base []Reconciler) []Reconciler {
		if sr == nil {
			return base
		}
		out := make([]Reconciler, len(base))
		copy(out, base)
		for i := range out {
			out[i].State, out[i].MissingKinds = sr.ReconcilerState(out[i].Package)
		}
		return out
	}

	// When there is no coordinator, pre-marshal the static response once.
	var staticFull []byte
	if sr == nil {
		var err error
		staticFull, err = json.Marshal(Response{
			Version:   ver,
			GitCommit: gitCommit,
			BuildDate: buildDate,
			GoVersion: goVersion,
			Features:  all,
		})
		if err != nil {
			panic(fmt.Sprintf("features.Handler: marshal: %v", err))
		}
	}

	marshalResponse := func(feats []Reconciler) []byte {
		b, _ := json.Marshal(Response{
			Version:   ver,
			GitCommit: gitCommit,
			BuildDate: buildDate,
			GoVersion: goVersion,
			Features:  feats,
		})
		return b
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead:
		default:
			w.Header().Set("Allow", "GET, HEAD")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = w.Write([]byte(`{"error":"method not allowed"}`))
			return
		}

		if name := r.URL.Query().Get("name"); name != "" {
			rec, ok := byPkg[name]
			if !ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				notFound, _ := json.Marshal(map[string]string{"error": "feature " + name + " not found"})
				_, _ = w.Write(notFound)
				return
			}
			if sr != nil {
				rec.State, rec.MissingKinds = sr.ReconcilerState(rec.Package)
			}
			body := marshalResponse([]Reconciler{rec})
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(body)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if sr == nil {
			_, _ = w.Write(staticFull)
			return
		}
		_, _ = w.Write(marshalResponse(buildFeatures(all)))
	})
}
