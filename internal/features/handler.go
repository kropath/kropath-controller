// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package features

import (
	"encoding/json"
	"fmt"
	"net/http"
)

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
// The full response body is computed once at call time from the supplied version
// info and reconciler list. Only GET and HEAD are accepted; any other method
// returns 405 with a JSON error body and an Allow header. A ?name=<pkg> query
// parameter filters the response to the single reconciler whose Package matches
// exactly; an unknown package returns 404 with a JSON error body.
func Handler(ver, gitCommit, buildDate, goVersion string, all []Reconciler) http.Handler {
	// Normalise nil so marshaling produces [] rather than null.
	if all == nil {
		all = []Reconciler{}
	}

	full, err := json.Marshal(Response{
		Version:   ver,
		GitCommit: gitCommit,
		BuildDate: buildDate,
		GoVersion: goVersion,
		Features:  all,
	})
	if err != nil {
		// All fields are strings and string slices — marshaling cannot fail.
		panic(fmt.Sprintf("features.Handler: marshal: %v", err))
	}

	// Pre-index by Package for O(1) lookup on ?name= queries.
	byPkg := make(map[string]Reconciler, len(all))
	for _, r := range all {
		byPkg[r.Package] = r
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
			body, _ := json.Marshal(Response{
				Version:   ver,
				GitCommit: gitCommit,
				BuildDate: buildDate,
				GoVersion: goVersion,
				Features:  []Reconciler{rec},
			})
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(body)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(full)
	})
}
